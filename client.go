package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"net"
)

func LaunchClient(serverAddr string) {
	fmt.Println("=== CLIENT OPTIMISÉ ===")

	targetImg := LoadImage("target.png") // Fonction LoadImage définie dans server.go ou common.go
	MaxX = targetImg.Bounds().Dx()
	MaxY = targetImg.Bounds().Dy()

	// Calcul de la couleur moyenne (nécessaire pour effacer le buffer correctement)
	avgColor := ComputeAverageColor(targetImg)

	// --- OPTIMISATION V2 : Buffer réutilisable ---
	// On alloue l'image de travail UNE SEULE FOIS ici
	workBuffer := image.NewRGBA(targetImg.Bounds())

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	for {
		var msg NetworkMessage
		err := decoder.Decode(&msg)
		if err != nil {
			return
		}

		candidate := msg.Organism

		// Logique V2
		progress := float64(len(candidate.DNA)) / TargetComplexity
		if progress > 1.0 {
			progress = 1.0
		}

		Mutate(&candidate, targetImg, progress)

		// --- OPTIMISATION V2 : On dessine sur le buffer existant ---
		RenderToBuffer(candidate.DNA, workBuffer, avgColor)

		// Calcul du score
		candidate.Score = DiffEuclidienne(workBuffer, targetImg)

		// Envoi si mieux (ou toujours, selon ta stratégie)
		if candidate.Score < msg.Organism.Score {
			err = encoder.Encode(NetworkMessage{Organism: candidate})
			if err != nil {
				return
			}
		} else {
			// Important: Renvoyer quelque chose pour que le serveur ne bloque pas
			// On renvoie l'original non modifié pour dire "échec"
			err = encoder.Encode(NetworkMessage{Organism: candidate})
			if err != nil {
				return
			}
		}
	}
}
