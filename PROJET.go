package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"   // Enregistre le décodeur JPG
	_ "image/jpeg" // Enregistre le décodeur JPEG
	"image/png"
	"math/rand"
	"os"
	"runtime"
	"time"
)

// ==========================================
// 1. STRUCTURES DE DONNÉES
// ==========================================

// Config contient les paramètres globaux
const (
	OutputFile    = "evolution.png" // L'image qu'on va regarder évoluer
	SaveFrequency = 100
	//MaxX          = 200 // Largeur max
	//MaxY          = 200 // Hauteur max
	MinRadius        = 5       // Taille min d'un cercle
	MaxRadius        = 30      // Taille max    // Sauvegarder l'image toutes les X générations
	TargetComplexity = 25000.0 // Nombre approximatif de formes dans l'image cible
)

var ( // Variables globales qui vont s'adapter à la résolution de l'image cible
	MaxX int
	MaxY int
)

type Shape struct {
	X, Y   int        // Centre du cercle
	Radius int        // Rayon
	Color  color.RGBA // Rouge, Vert, Bleu, Alpha (Transparence)
}

// Organism représente une solution complète (une image candidate)
type Organism struct {
	DNA   []Shape     // La liste des instructions de dessin
	Score float64     // La distance Euclidienne (plus petit = mieux)
	Image *image.RGBA // Le rendu visuel (cache pour éviter de redessiner)
}

// Job est ce que le Main envoie aux Workers
type Job struct {
	BestOrganism Organism // Le meilleur dessin actuel
}

// Result est ce que le Worker renvoie au Main
type Result struct {
	Organism Organism // Le candidat modifié
	IsBetter bool     // Indique si le score est meilleur
}

// ==========================================
// 2. FONCTIONS MÉTIERS (A IMPLEMENTER)
// ==========================================

func NewRandomShape(target *image.RGBA) Shape {
	x := rand.Intn(MaxX)
	y := rand.Intn(MaxY)

	// On récupère la couleur du vrai pixel à cet endroit, optimisation avec heuristique
	// Optimisation : accès direct mémoire (comme en C)
	offset := (y * target.Stride) + (x * 4)
	r := target.Pix[offset]
	g := target.Pix[offset+1]
	b := target.Pix[offset+2]

	return Shape{
		X:      x,
		Y:      y,
		Radius: rand.Intn(MaxRadius-MinRadius) + MinRadius,
		Color: color.RGBA{
			R: r, // On prend la vraie couleur !
			G: g,
			B: b,
			A: uint8(rand.Intn(200) + 30), // Alpha aléatoire (30-230)
		},
	}
}

// Mutate prend un organisme et le modifie légèrement au hasard
func Mutate(o *Organism, target *image.RGBA, progress float64) {
	// TODO:
	// 1. Choisir un nombre au hasard
	// 2. Soit modifier une couleur d'une forme existante
	// 3. Soit bouger un point d'une forme
	// 4. Soit ajouter/supprimer une forme

	//Calcul du rayon max e
	currentMaxRadius := int(float64(MaxRadius) * (1.0 - progress))
	if currentMaxRadius < MinRadius {
		currentMaxRadius = MinRadius
	}
	// Générateur de probabilité (entre 0.0 et 1.0)
	roulette := rand.Float64()

	// ---------------------------------------------------------
	// CAS 1 : CROISSANCE (Ajouter une forme)
	// Probabilité : 10% (ou 100% si l'ADN est vide au début)
	// ---------------------------------------------------------
	if len(o.DNA) == 0 || roulette < 0.1 {
		nouvelleForme := NewRandomShape(target)
		// append est l'équivalent d'un realloc + push_back, à expliquer.
		o.DNA = append(o.DNA, nouvelleForme) //ie on rajoute une nouvelle shape au tableau de shapes.
		return
	}

	// ---------------------------------------------------------
	// CAS 2 : ÉLAGAGE (Supprimer une forme)
	// Probabilité : 10% (seulement si on a des formes)
	// ---------------------------------------------------------
	if roulette < 0.2 {
		// On choisit un index au hasard à supprimer
		indexKill := rand.Intn(len(o.DNA))

		// LA syntaxe Go pour supprimer un élément d'un tableau dynamique :
		// On concatène [début ... i] avec [i+1 ... fin]
		o.DNA = append(o.DNA[:indexKill], o.DNA[indexKill+1:]...)
		return
	}

	// ---------------------------------------------------------
	// CAS 3 : MUTATION FINE (Modifier une forme existante)
	// Probabilité : 80% (Le reste du temps)
	// ---------------------------------------------------------

	// A. On cible une forme au hasard dans l'ADN
	indexModif := rand.Intn(len(o.DNA))

	// B. On prend un POINTEUR (&) vers cette forme.
	// ⚠️ CRUCIAL : Si tu fais 's := o.DNA[i]', tu copies la structure
	// et tu modifies la copie. L'ADN original ne changera pas !
	s := &o.DNA[indexModif]

	// C. On choisit quelle propriété modifier
	switch rand.Intn(3) { // 0, 1 ou 2
	case 0: // BOUGER (Position)
		// On décale de -10 à +10 pixels
		s.X += rand.Intn(21) - 10
		s.Y += rand.Intn(21) - 10
		// Clamp (garde-fou) pour ne pas sortir de l'image
		if s.X < 0 {
			s.X = 0
		}
		if s.Y < 0 {
			s.Y = 0
		}
		if s.X > MaxX {
			s.X = MaxX
		}
		if s.Y > MaxY {
			s.Y = MaxY
		}

	case 1: // REDIMENSIONNER (Taille)
		change := rand.Intn(5) - 2
		s.Radius += change
		if s.Radius > currentMaxRadius {
			s.Radius = currentMaxRadius
		}

	case 2: // RECOLORER (Couleur)
		// On change une seule composante (R, G, B ou A) pour affiner
		switch rand.Intn(4) {
		case 0:
			s.Color.R = uint8(rand.Intn(256))
		case 1:
			s.Color.G = uint8(rand.Intn(256))
		case 2:
			s.Color.B = uint8(rand.Intn(256))
		case 3:
			s.Color.A = uint8(rand.Intn(256)) // Changer la transparence change beaucoup le rendu !
		}
	}
}

// Render transforme l'ADN (les formes) en pixels sur une image

func Render(dna []Shape, width, height int) *image.RGBA {
	// 1. Allouer l'image (malloc géant caché)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 2. Remplir le fond en Noir Opaque (R=0, G=0, B=0, A=255)
	// Par défaut Go met tout à 0 (donc transparent). Pour voir quelque chose, il faut de l'opacité.
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255 // Canal Alpha à fond
	}

	// 3. Dessiner chaque forme une par une
	for _, shape := range dna {
		drawCircle(img, shape)
	}

	return img
}

func drawCircle(img *image.RGBA, s Shape) {
	// A. Optimisation : Bounding Box (Boîte englobante)
	// On ne parcourt que le carré autour du cercle, pas toute l'image
	minX := s.X - s.Radius
	maxX := s.X + s.Radius
	minY := s.Y - s.Radius
	maxY := s.Y + s.Radius

	// Clamp (On s'assure qu'on ne sort pas de l'image)
	bounds := img.Bounds()
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > bounds.Max.X {
		maxX = bounds.Max.X
	}
	if maxY > bounds.Max.Y {
		maxY = bounds.Max.Y
	}

	// Pré-calcul du rayon au carré pour éviter des sqrt() lents dans la boucle
	radiusSq := s.Radius * s.Radius

	// B. La double boucle sur les pixels (Comme en C)
	for y := minY; y < maxY; y++ {
		// Petite optimisation : on calcule dy une seule fois par ligne
		dy := y - s.Y
		dy2 := dy * dy

		for x := minX; x < maxX; x++ {
			dx := x - s.X

			// C. Test géométrique : Est-on dans le cercle ?
			if dx*dx+dy2 <= radiusSq {

				// D. Accès mémoire direct au pixel (x, y)
				offset := (y * img.Stride) + (x * 4)

				// E. Alpha Blending (Mélange des couleurs)
				// Formule : Out = Source * Alpha + Dest * (1 - Alpha)

				// Couleur du fond actuel (Destination)
				dstR := int(img.Pix[offset+0])
				dstG := int(img.Pix[offset+1])
				dstB := int(img.Pix[offset+2])

				// Couleur de la forme (Source)
				srcR := int(s.Color.R)
				srcG := int(s.Color.G)
				srcB := int(s.Color.B)
				alpha := int(s.Color.A) // 0 à 255

				// Calcul du mélange (C'est des maths de base pour la transparence)
				// On divise par 256 (via bitshift >> 8) pour normaliser
				// R
				r := (srcR*alpha + dstR*(255-alpha)) / 255
				// G
				g := (srcG*alpha + dstG*(255-alpha)) / 255
				// B
				b := (srcB*alpha + dstB*(255-alpha)) / 255

				// Écriture du résultat
				img.Pix[offset+0] = uint8(r)
				img.Pix[offset+1] = uint8(g)
				img.Pix[offset+2] = uint8(b)
				// On laisse l'Alpha du fond à 255 (Opaque)
				img.Pix[offset+3] = 255
			}
		}
	}
}

// DiffEuclidienne compare deux images pixel par pixel
// Retourne la "distance" totale. 0 = images identiques.
func DiffEuclidienne(img1, img2 *image.RGBA) float64 {
	var totalDiff float64 = 0.0

	// Petit garde-fou : on s'assure que les tableaux ont la même taille
	// Sinon on risque un "panic: index out of range"
	if len(img1.Pix) != len(img2.Pix) {
		fmt.Println("Erreur : Images de tailles différentes !")
		return 1e15 // Retourne un score énorme (infini) pour dire "c'est nul"
	}

	// BOUCLE PRINCIPALE (Optimisée "Style C")
	// On avance de 4 en 4 car un pixel = 4 octets (R, G, B, A)
	for i := 0; i < len(img1.Pix); i += 4 {

		// ------------------------------------------------
		// Canal ROUGE (Offset i)
		// ------------------------------------------------
		// ⚠️ CAST CRUCIAL : uint8 (0-255) -> int
		// Si tu fais (20 - 200) en uint8, ça underflow et donne un grand nombre !
		// En int, ça donne -180, ce qui est correct pour le calcul.
		r1 := int(img1.Pix[i])
		r2 := int(img2.Pix[i])
		diffR := r1 - r2
		totalDiff += float64(diffR * diffR) // Carré de la différence

		// ------------------------------------------------
		// Canal VERT (Offset i+1)
		// ------------------------------------------------
		g1 := int(img1.Pix[i+1])
		g2 := int(img2.Pix[i+1])
		diffG := g1 - g2
		totalDiff += float64(diffG * diffG)

		// ------------------------------------------------
		// Canal BLEU (Offset i+2)
		// ------------------------------------------------
		b1 := int(img1.Pix[i+2])
		b2 := int(img2.Pix[i+2])
		diffB := b1 - b2
		totalDiff += float64(diffB * diffB)

		// Note : On ignore souvent l'Alpha (i+3) pour la comparaison
		// car on compare l'apparence visuelle finale.
	}

	return totalDiff
}

// Copy crée une copie profonde de l'organisme (CRUCIAL pour la concurrence)
func (o Organism) Copy() Organism {
	newOrg := Organism{
		Score: o.Score,
		DNA:   make([]Shape, len(o.DNA)),
	}
	copy(newOrg.DNA, o.DNA)
	// Note: On ne copie pas forcément o.Image ici, on le recréera au Render
	return newOrg
}

// ==========================================
// 3. LOGIQUE CONCURRENTE (WORKERS)
// ==========================================

func Worker(id int, jobs <-chan Job, results chan<- Result, targetImg *image.RGBA) {
	fmt.Printf("Worker %d démarré\n", id)

	for job := range jobs {
		// 1. On travaille sur une copie pour ne pas casser l'original du Main
		candidate := job.BestOrganism.Copy()

		//Calcul du progrès basé sur la complexité cible
		nbFormes := len(candidate.DNA)
		progress := float64(nbFormes) / TargetComplexity
		if progress > 1.0 {
			progress = 1.0
		}

		// 2. On applique une modification aléatoire
		Mutate(&candidate, targetImg, progress) // Progression basée sur la taille de l'ADN

		// 3. On génère l'image résultante
		candidate.Image = Render(candidate.DNA, targetImg.Bounds().Dx(), targetImg.Bounds().Dy())

		// 4. On calcule le score (Distance Euclidienne)
		candidate.Score = DiffEuclidienne(candidate.Image, targetImg)

		// 5. On compare avec le score du parent
		// Si la distance est plus petite, on a amélioré le dessin !
		isBetter := candidate.Score < job.BestOrganism.Score

		// 6. On renvoie le résultat
		results <- Result{Organism: candidate, IsBetter: isBetter}
	}
}

// ==========================================
// 4. MAIN (ORCHESTRATEUR)
// ==========================================

func main() {
	// A. Chargement de l'image cible (à remplacer par votre fichier)
	// targetImg := LoadImage("joconde.png")
	// Pour l'exemple, on crée une image cible noire de 200x200
	targetImg := LoadImage("PAIN.png") //image.NewRGBA(image.Rect(0, 0, 200, 200))

	// On adapte les variables globales MaxX et MaxY à l'image chargée
	bounds := targetImg.Bounds()
	MaxX = bounds.Dx() // Largeur (Width)
	MaxY = bounds.Dy() // Hauteur (Height)

	fmt.Printf("Image chargée : %dx%d pixels\n", MaxX, MaxY)

	// B. Initialisation
	// On détecte la puissance du PC
	nbCoeurs := runtime.NumCPU()
	fmt.Printf("CPU: %d cœurs détectés. Lancement de %d Workers.\n", nbCoeurs, nbCoeurs)

	// Création des canaux de communication
	jobs := make(chan Job, nbCoeurs)       // Le Main donne du travail
	results := make(chan Result, nbCoeurs) // Les Workers répondent

	// Lancement des Workers
	for w := 1; w <= nbCoeurs; w++ {
		go Worker(w, jobs, results, targetImg)
	}

	// Création du premier organisme (vide ou aléatoire)
	currentBest := Organism{
		DNA:   []Shape{},
		Score: 1e20, // Score infini au début
	}

	// C. Boucle d'évolution (Infinie)
	generation := 0
	startTime := time.Now()

	for {
		generation++

		// 1. On envoie du travail à tous nos workers
		// Chaque worker va essayer une mutation différente de "currentBest"
		for i := 0; i < nbCoeurs; i++ {
			jobs <- Job{BestOrganism: currentBest}
		}

		// 2. On attend les réponses et on garde le meilleur des candidats
		bestCandidateOfRound := currentBest
		improved := false

		for i := 0; i < nbCoeurs; i++ {
			res := <-results // Bloquant : on attend le retour du worker

			if res.IsBetter {
				// Est-ce que ce worker a fait mieux que le meilleur de ce tour ?
				if res.Organism.Score < bestCandidateOfRound.Score {
					bestCandidateOfRound = res.Organism
					improved = true
				}
			}
		}

		// 3. Mise à jour de l'élite
		if improved {
			currentBest = bestCandidateOfRound
			fmt.Printf("[Gen %d] Amélioration ! Score: %.2f (Temps: %s)\n",
				generation, currentBest.Score, time.Since(startTime))
		}

		// 4. Feedback Visuel
		// Toutes les X générations, on sauvegarde l'image sur le disque
		if generation%SaveFrequency == 0 {
			SaveOutput(currentBest.Image)
		}
	}
}

func LoadImage(path string) *image.RGBA {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("Erreur ouverture:", err, "- Utilisation image blanche 200x200")
		// Fallback manuel si pas d'image
		return image.NewRGBA(image.Rect(0, 0, 200, 200))
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		fmt.Println("Erreur décodage:", err, "- Utilisation image blanche 200x200")
		return image.NewRGBA(image.Rect(0, 0, 200, 200))
	}

	// Conversion propre en RGBA
	b := src.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, src, b.Min, draw.Src)
	return rgba
}

// Helper pour sauvegarder l'image
func SaveOutput(img *image.RGBA) {
	if img == nil {
		return
	}
	f, err := os.Create(OutputFile)
	if err != nil {
		fmt.Println("Erreur sauvegarde:", err)
		return
	}
	defer f.Close()
	png.Encode(f, img)
	// fmt.Println("Image sauvegardée :", OutputFile)
}
