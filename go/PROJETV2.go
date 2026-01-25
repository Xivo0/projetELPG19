package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg" // Support JPEGOn n'utilise pas directement les fonctions de jpeg dans le code, mais on a besoin que la librairie s'initialise (elle appelle sa fonction init()) pour s'enregistrer comme format décodable.
	"image/png"    // Support PNG
	"math/rand"
	"os"
	"runtime"
	"time"
)

// ==========================================
// 1. CONFIGURATION & STRUCTURES
// ==========================================

const (
	InputFile        = "target.png"    // l'image source
	OutputFile       = "evolution.png" // Le résultat
	SaveFrequency    = 100             // Sauvegarde toutes les X générations
	MinRadius        = 3               // Taille min forme
	MaxRadius        = 40              // Taille max forme
	TargetComplexity = 1000.0          // Complexité visée pour l'ajustement auto
)

// Types de formes
const (
	ShapeTypeCircle = 0
	ShapeTypeRect   = 1
)

var (
	MaxX int
	MaxY int
)

// structure pour une forme
type Shape struct {
	Type   int        // 0: Cercle, 1: Rectangle
	X, Y   int        // Centre
	Radius int        // Rayon ou "Demi-largeur"
	Color  color.RGBA // Couleur
}

type Organism struct {
	DNA   []Shape     // La liste des instructions de dessin
	Score float64     // La différence (plus bas = mieux)
	Image *image.RGBA // Image rendue (optionnel, seulement si nécessaire)
}

type Job struct {
	BestOrganism Organism
}

type Result struct {
	Organism Organism
	IsBetter bool
}

// ==========================================
// 2. FONCTIONS MÉTIERS OPTIMISÉES
// ==========================================

// ComputeAverageColor parcourt tous les pixels de l'image cible (boucle X et Y), additionne le Rouge, Vert, Bleu, et divise par le nombre total de pixels.
// Cela permet de commencer avec un fond qui n'est pas noir, mais "moyen".
func ComputeAverageColor(img *image.RGBA) color.RGBA {
	var r, g, b, count uint64
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			off := img.PixOffset(x, y)
			r += uint64(img.Pix[off+0])
			g += uint64(img.Pix[off+1])
			b += uint64(img.Pix[off+2])
			count++
		}
	}
	return color.RGBA{
		R: uint8(r / count),
		G: uint8(g / count),
		B: uint8(b / count),
		A: 255,
	}
}

func NewRandomShape(target *image.RGBA) Shape {
	x := rand.Intn(MaxX)
	y := rand.Intn(MaxY)

	// accès direct mémoire pour piquer la couleur au lieu de prendre une couleur 100% aléatoire,
	//on regarde l'image cible à la position (X, Y) où on va poser la forme, et on prend cette couleur.
	offset := (y * target.Stride) + (x * 4)
	r := target.Pix[offset]
	g := target.Pix[offset+1]
	b := target.Pix[offset+2]

	return Shape{
		Type:   rand.Intn(2), // 50% Cercle, 50% Rectangle
		X:      x,
		Y:      y,
		Radius: rand.Intn(MaxRadius-MinRadius) + MinRadius,
		Color: color.RGBA{
			R: r,
			G: g,
			B: b,
			A: uint8(rand.Intn(200) + 30), // Alpha min 30 pour éviter les formes invisibles
		},
	}
}

// RenderToBuffer : Dessine l'ADN sur une image existante (Pas d'allocation mémoire !)
func RenderToBuffer(dna []Shape, img *image.RGBA, bg color.RGBA) {
	// 1. Reset rapide du fond (memset style)
	bgR, bgG, bgB := bg.R, bg.G, bg.B// On stocke les composantes de la couleur de fond dans des variables locales
    // pour éviter d'accéder à la structure `bg` à chaque tour de boucle
	for i := 0; i < len(img.Pix); i += 4 {// On parcourt tout le tableau de pixels de l'image
		img.Pix[i+0] = bgR
		img.Pix[i+1] = bgG
		img.Pix[i+2] = bgB
		img.Pix[i+3] = 255 // Opaque
	}

	// 2. Dessin des formes
	for _, shape := range dna {
		drawShape(img, shape)
	}
}

func drawShape(img *image.RGBA, s Shape) { //on dessine une forme sur l'image
	// Bounding Box : au lieu de parcourir toute l'image (et perdre du temps de calcul), on se limite au carré minimal qui contient la forme pour la trouver
	minX, maxX := s.X-s.Radius, s.X+s.Radius
	minY, maxY := s.Y-s.Radius, s.Y+s.Radius

	bounds := img.Bounds() // Récupération du carré minimal pour ne pas sortir de l'image
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > bounds.Max.X { // Si la forme dépasse, on clamp : ie on ne sort pas de l'image (eviter les index out of bounds)
		maxX = bounds.Max.X
	}
	if maxY > bounds.Max.Y {
		maxY = bounds.Max.Y
	}

	// Pré-calculs 
	radiusSq := s.Radius * s.Radius // Pour le cercle
	srcR, srcG, srcB, alpha := int(s.Color.R), int(s.Color.G), int(s.Color.B), int(s.Color.A) // Couleur source et alpha
	invAlpha := 255 - alpha

	//on parcourt chaque pixel de la bounding box pour dessiner la forme
	for y := minY; y < maxY; y++ {
		lineOffset := y * img.Stride //pour trouver le début de la ligne
		
		// Optimisation du cercle
		dy := y - s.Y // Distance verticale au centre ie (y - centreY)
		dy2 := dy * dy // Carré de la distance verticale

		for x := minX; x < maxX; x++ {
			// Logique Géométrique pour vérifier si on est dans la forme (la distance du pixel au centre doit être inférieure au rayon du cercle)
			if s.Type == ShapeTypeCircle {
				dx := x - s.X //
				if dx*dx+dy2 > radiusSq {
					continue // Hors du cercle
				}
			}
			// Si Rectangle, on dessine tout ce qui est dans la boucle

			// Logique de Blending (Mélange) : on veux du "alpha blending" pour gérer la transparence (faire une transition plus douche d'une couleur vers une autre)
			offset := lineOffset + (x * 4) // Position du  (x,y) dans Pix[]

			// Formule: (Src * A + Dst * (255-A)) / 255 ou Src et Dst sont les couleurs source et destination
			r := (srcR*alpha + int(img.Pix[offset+0])*invAlpha) / 255 
			g := (srcG*alpha + int(img.Pix[offset+1])*invAlpha) / 255
			b := (srcB*alpha + int(img.Pix[offset+2])*invAlpha) / 255

			img.Pix[offset+0] = uint8(r) //On modifie la couleur du pixel (x,y) avec celles obtenues (un pixel prends 4 octets de couleurs + transparence)
			img.Pix[offset+1] = uint8(g)
			img.Pix[offset+2] = uint8(b)
			// Alpha reste 255 (image finale opaque)
		}
	}
}

func DiffEuclidienne(img1, img2 *image.RGBA) float64 { //Calcul de la différence entre deux images
	var totalDiff float64 = 0.0

	// On suppose que les images ont la même taille (garanti par le code)
	for i := 0; i < len(img1.Pix); i += 4 {
		// Diff R
		d1 := int(img1.Pix[i]) - int(img2.Pix[i])
		// Diff G
		d2 := int(img1.Pix[i+1]) - int(img2.Pix[i+1])
		// Diff B
		d3 := int(img1.Pix[i+2]) - int(img2.Pix[i+2])

		totalDiff += float64(d1*d1 + d2*d2 + d3*d3)
	}
	return totalDiff
}

func Mutate(o *Organism, target *image.RGBA, progress float64) {
	// Adapter le rayon max au progrès (plus on avance, plus les formes sont petites)
	currentMaxRadius := int(float64(MaxRadius) * (1.1 - progress))
	if currentMaxRadius < MinRadius {
		currentMaxRadius = MinRadius
	}

	roulette := rand.Float64() // nombre réel entre 0 et 1

	// 1. Ajout d'une forme (Plus fréquent et utilisé au début)
	if len(o.DNA) == 0 || roulette < 0.1 {
		o.DNA = append(o.DNA, NewRandomShape(target))
		return
	}

	// 2. Suppression d'une forme
	if roulette < 0.15 {
		index := rand.Intn(len(o.DNA))
		o.DNA = append(o.DNA[:index], o.DNA[index+1:]...)
		return
	}

	// 3. Echange de la position de 2 formes
	if roulette < 0.20 {
		i1 := rand.Intn(len(o.DNA))
		i2 := rand.Intn(len(o.DNA))
		o.DNA[i1], o.DNA[i2] = o.DNA[i2], o.DNA[i1]
		return
	}

	// 4. Modficiation d'une forme aléatoire
	index := rand.Intn(len(o.DNA))
	s := &o.DNA[index]

	switch rand.Intn(4) { // 4 types de modifs
	case 0: // Position, déplacement de la forme de +/- 10 pixels (horizontal et vertical)
		s.X += rand.Intn(21) - 10 
		s.Y += rand.Intn(21) - 10
		// Clamp, si on sort du cadre
		if s.X < 0 { s.X = 0 }
		if s.X > MaxX { s.X = MaxX }
		if s.Y < 0 { s.Y = 0 }
		if s.Y > MaxY { s.Y = MaxY }

	case 1: // Taille de la forme, on la modifie (toujours entre min et max radius)
		s.Radius += rand.Intn(11) - 5
		if s.Radius < MinRadius { s.Radius = MinRadius }
		if s.Radius > currentMaxRadius { s.Radius = currentMaxRadius }

	case 2: // Couleur (changement aléatoire et souvent distinct)
		switch rand.Intn(4) {
		case 0: s.Color.R = uint8(rand.Intn(256))
		case 1: s.Color.G = uint8(rand.Intn(256))
		case 2: s.Color.B = uint8(rand.Intn(256))
		case 3: 
			newA := int(s.Color.A) + rand.Intn(30) - 15
			if newA < 10 { newA = 10 } // Pas trop transparent
			if newA > 255 { newA = 255 }
			s.Color.A = uint8(newA)
		}
	
	case 3: // Changement de TYPE (Cercle <-> Rect)
		s.Type = 1 - s.Type
	}
}

func (o Organism) Copy() Organism {
	newOrg := Organism{
		Score: o.Score,
		DNA:   make([]Shape, len(o.DNA)), //on doit allouer espace tableau
	}
	copy(newOrg.DNA, o.DNA)
	return newOrg
}

// ==========================================
// 3. WORKER OPTIMISÉ (BUFFER REUSE)
// ==========================================

func Worker(id int, jobs <-chan Job, results chan<- Result, targetImg *image.RGBA, bgCol color.RGBA) {
	// ! CRUCIAL : Allocation unique du buffer de travail par Worker !
	// On réutilise cette image pour dessiner, au lieu d'en créer une nouvelle à chaque fois.
	myBuffer := image.NewRGBA(targetImg.Bounds()) // On alloue (résolution_en_pixels * 4 octets, 4 octets pour RGBA)

	// Utilisation d'une source aléatoire locale si on voulait (ici global rand est ok pour simplicité)
	
	for job := range jobs {
		candidate := job.BestOrganism.Copy()

		// Mutation adaptative
		progress := float64(len(candidate.DNA)) / TargetComplexity // TargetComplexity = nombre de formes max/objectif
		//progress commence de 0 environ et se rapproche de 1 (0% à 100% de complétion grosso-modo)
		if progress > 1.0 { progress = 1.0 }
		
		Mutate(&candidate, targetImg, progress)

		// Rendu sur notre buffer "effaçable", le programme lit les vecteurs du DNA
		// il écrit les pixels correspondants directement dans la zone mémoire de myBuffer
		// (il écrase les pixels de la boucle précédente, on a du recyclage)
		RenderToBuffer(candidate.DNA, myBuffer, bgCol) 

		// Calcul du score sur le buffer
		candidate.Score = DiffEuclidienne(myBuffer, targetImg)

		isBetter := candidate.Score < job.BestOrganism.Score // plus on est proche de 0, mieux c'est

		if isBetter {
			// Si c'est mieux, on doit cloner l'image pour l'envoyer au Main
			// (Car myBuffer va être effacé au prochain tour)
			// Petite optimisation : on ne clone que si nécessaire
			finalImg := image.NewRGBA(myBuffer.Bounds()) // On alloue encore pour sauvegarder (résolution_en_pixels * 4 octets, 4 octets pour RGBA)
			copy(finalImg.Pix, myBuffer.Pix)
			candidate.Image = finalImg
		} else {
			candidate.Image = nil // Pas d'image si échec
		}

		results <- Result{Organism: candidate, IsBetter: isBetter}
	}
}

// ==========================================
// 4. MAIN
// ==========================================

func main() {
	// 1. Initialisation
	rand.Seed(time.Now().UnixNano())// Initialise le générateur pseudo-aléatoire global : faire en sorte qu'il n'y ai pas a chaque fois EXACTEMENT la même évolution

	targetImg := LoadImage(InputFile)// charge l'image (image RGBA)
	bounds := targetImg.Bounds()//Récupère le rectangle de l’image (contient Min et Max de X et Y)
	MaxX, MaxY = bounds.Dx(), bounds.Dy()// largeur et hauteur pour génération aléatoire, rendu rapide...

	fmt.Printf("Cible: %dx%d | CPU: %d\n", MaxX, MaxY, runtime.NumCPU())// informatif: taille de l'image/ nombre de coeur dispo

	// 2. Calcul de la couleur moyenne pour le fond
	avgColor := ComputeAverageColor(targetImg)// Calcule la couleur moyenne de l’image cible
	fmt.Printf("Couleur de fond moyenne calculée: R%d G%d B%d\n", avgColor.R, avgColor.G, avgColor.B)// informe l'utilisateur de la cohérence entre couleur moyenne du fond et de l'image 

	// 3. Canaux et Workers
	nbWorkers := runtime.NumCPU()// 1 workers par coeur CPU
	jobs := make(chan Job, nbWorkers)// Job envoie des mutation  
	results := make(chan Result, nbWorkers)// les scores 

	for w := 1; w <= nbWorkers; w++ {
		go Worker(w, jobs, results, targetImg, avgColor)// ici on lance les go routine (worker avec buffer, image privé, pas d'alloc répétée et indépendant) 
	}

	// 4. Premier Organisme (Vide mais avec le bon score initial)
	// On crée une image vide avec la couleur de fond pour calculer le score de départ
	startImg := image.NewRGBA(bounds)// creation image vide meme taille que l'image cible 
	emptyDNA := []Shape{}// ADN vide que le fond 
	RenderToBuffer(emptyDNA, startImg, avgColor) // Juste le fond moyen pas de forme 
	startScore := DiffEuclidienne(startImg, targetImg)// calcule score initiale 

	currentBest := Organism{
		DNA:   emptyDNA,
		Score: startScore,
		Image: startImg,
	}
	
	SaveOutput(currentBest.Image) // Sauvegarde de l'état initial (fond uni)

	// 5. Boucle Principale
	generation := 0 // compteur du nombre de générations
	startTime := time.Now()//sert à calculer les FPS (générations par seconde)
	lastPrint := time.Now()//permet de limiter l’affichage console à 1 fois par seconde

	for {
		generation++

		// Envoi des jobs
		for i := 0; i < nbWorkers; i++ {
			jobs <- Job{BestOrganism: currentBest}//Tu envoies nbWorkers jobs Chaque job contient une copie logique du meilleur organisme actuel
		}

		// Réception des résultats
		bestOfRound := currentBest//meilleur organisme trouvé dans cette génération
		improved := false//indique si au moins une amélioration a été trouvée
		// comparaison des resultats des différents workers et changement du bestofround et de l'atat de improved 
		for i := 0; i < nbWorkers; i++ {
			res := <-results
			if res.IsBetter {
				if res.Organism.Score < bestOfRound.Score {
					bestOfRound = res.Organism
					improved = true
				}
			}
		}

		// Mise à jour si amélioration
		if improved {
			currentBest = bestOfRound
		}

		// Feedback console (toutes les secondes)
		if time.Since(lastPrint) > 1*time.Second { // empeche le spam console et verifie si 1sec est passé 
			fmt.Printf("[Gen %d] Formes: %d | Score: %.0f | FPS: %.0f\n", 
				generation, len(currentBest.DNA), currentBest.Score, float64(generation)/time.Since(startTime).Seconds())
			lastPrint = time.Now()// Réinitialise le timer d’affichage
		}//Gen : numéro de génération Formes : complexité actuelle (nombre de formes) Score : distance euclidienne à l’image cible FPS : générations par seconde

		// Sauvegarde Image
		if improved && generation%SaveFrequency == 0 {
			SaveOutput(currentBest.Image)
		}
	}
}

// Helpers...
func LoadImage(path string) *image.RGBA {// charge une image depuis le disque et retourne une image *RGBA
	f, err := os.Open(path)// ouvre le fichier image 
	if err != nil {
		fmt.Println("Erreur chargement:", err, "-> Création image test")
		return image.NewRGBA(image.Rect(0, 0, 200, 200))// au cas ou si image n'existe pas en crée une vide 200x200
	}
	defer f.Close() // ferme fichier a la fin de la func
	src, _, err := image.Decode(f)// decode les png JPEG...
	if err != nil {
		return image.NewRGBA(image.Rect(0, 0, 200, 200))// meme securiter si fail decodage 
	}
	b := src.Bounds()
	rgba := image.NewRGBA(b)//Crée une image RGBA exactement de la même taille
	draw.Draw(rgba, b, src, b.Min, draw.Src)//Copie les pixels de l’image source vers le buffer RGBA Conversion implicite vers RGBA
	return rgba
}

func SaveOutput(img *image.RGBA) {//sauvegarde l'image final du disque 
	if img == nil { return }
	f, err := os.Create(OutputFile)//Crée (ou écrase) evolution.png
	if err != nil { return }
	defer f.Close()//garantit la fermeture du fichier même en cas de return même si le code évolue
	png.Encode(f, img)//Encode l’image RGBA en PNG, Écrit les pixels sur disque
}
