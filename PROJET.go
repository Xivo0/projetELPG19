package main

import (
	"fmt"
	"image"
	"image/png"
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
	SaveFrequency = 100             // Sauvegarder l'image toutes les X générations
)

// Shape représente une forme (ex: Triangle, Cercle, Polygone)
type Shape struct {
	// TODO: Ajouter les propriétés (Points X/Y, Couleur R/G/B/A, Type)
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

// Mutate prend un organisme et le modifie légèrement au hasard
func Mutate(o *Organism) {
	// TODO:
	// 1. Choisir un nombre au hasard
	// 2. Soit modifier une couleur d'une forme existante
	// 3. Soit bouger un point d'une forme
	// 4. Soit ajouter/supprimer une forme
}

// Render transforme l'ADN (les formes) en pixels sur une image
func Render(dna []Shape, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// TODO:
	// 1. Remplir le fond (noir ou blanc)
	// 2. Boucler sur les Shapes et les dessiner (image/draw)
	// Astuce : Gérer la transparence (Alpha) est crucial pour le rendu artistique
	return img
}

// DiffEuclidienne compare deux images pixel par pixel
// Retourne la "distance" totale. 0 = images identiques.
func DiffEuclidienne(img1, img2 *image.RGBA) float64 {
	var totalDiff float64 = 0.0

	// TODO: Optimisation critique ici (accès direct via Pix)
	// Pour chaque pixel i (R, G, B) :
	// diffR = (R1 - R2)
	// diffG = (G1 - G2)
	// diffB = (B1 - B2)
	// totalDiff += sqrt(diffR^2 + diffG^2 + diffB^2) ou juste la somme des carrés

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

		// 2. On applique une modification aléatoire
		Mutate(&candidate)

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
	targetImg := image.NewRGBA(image.Rect(0, 0, 200, 200))

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
		Score: 9999999999.0, // Score infini au début
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

