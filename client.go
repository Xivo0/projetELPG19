package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"net"
)

func LaunchClient(serverAddr string) {
	fmt.Println("=== CLIENT OPTIMISÉ ===")

	targetImg := LoadImage("PAIN.png") // Fonction LoadImage définie dans server.go ou common.go
	MaxX = targetImg.Bounds().Dx()
	MaxY = targetImg.Bounds().Dy()
    
    // Calcul de la couleur moyenne (nécessaire pour effacer le buffer correctement)
    avgColor := ComputeAverageColor(targetImg)

	// --- OPTIMISATION V2 : Buffer réutilisable ---
	// On alloue l'image de travail UNE SEULE FOIS ici
	workBuffer := image.NewRGBA(targetImg.Bounds())

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil { panic(err) }
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	for {
		// 1. Recevoir le travail
		var msg NetworkMessage
		err := decoder.Decode(&msg)
		if err != nil {
			fmt.Println("Serveur perdu !")
			return
		}

		candidate := msg.Organism
		attempts++

		// 2. Muter
		progress := float64(len(candidate.DNA)) / TargetComplexity
		if progress > 1.0 { progress = 1.0 }
		Mutate(&candidate, targetImg, progress)

		// 3. Rendu & Score
		RenderToBuffer(candidate.DNA, workBuffer, avgColor)
		candidate.Score = DiffEuclidienne(workBuffer, targetImg)

		// 4. Renvoyer et LOGGUER
		if candidate.Score < msg.Organism.Score {
			// SUCCÈS : On affiche un message vert (ou juste du texte)
			fmt.Printf("[Test %d] ✨ J'ai trouvé mieux ! Score: %.0f\n", attempts, candidate.Score)
			encoder.Encode(NetworkMessage{Organism: candidate})
		} else {
			// ÉCHEC : On affiche juste un petit point pour dire "je suis en vie"
			// (Astuce : on n'affiche un point que tous les 100 échecs pour pas spammer)
			if attempts%100 == 0 {
				fmt.Print(".") 
			}
			encoder.Encode(NetworkMessage{Organism: candidate})
		}
	}
}
