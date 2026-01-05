package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"net"
)

func LaunchClient(serverAddr string) {
	fmt.Println("=== MODE CLIENT (V2 Optimisé) ===")

	// Charger la cible locale pour les calculs
	targetImg := LoadImage(InputFile)
	MaxX = targetImg.Bounds().Dx()
	MaxY = targetImg.Bounds().Dy()
	
	// Calculer la couleur moyenne pour effacer le buffer
	avgColor := ComputeAverageColor(targetImg)

	// --- OPTIMISATION V2 : Buffer unique ---
	workBuffer := image.NewRGBA(targetImg.Bounds())

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil { panic(err) }
	defer conn.Close()
	fmt.Println("Connecté au serveur. Calcul en cours...")

	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	for {
		// 1. Recevoir
		var msg NetworkMessage
		err := decoder.Decode(&msg)
		if err != nil { return }

		candidate := msg.Organism

		// 2. Muter
		progress := float64(len(candidate.DNA)) / TargetComplexity
		if progress > 1.0 { progress = 1.0 }
		Mutate(&candidate, targetImg, progress)

		// 3. Rendu V2 (Sur le buffer existant)
		RenderToBuffer(candidate.DNA, workBuffer, avgColor)
		candidate.Score = DiffEuclidienne(workBuffer, targetImg)

		// 4. Renvoyer (Seulement si mieux, ou tout le temps selon stratégie)
		// Ici on renvoie tout pour simplifier la boucle synchrone
		// (On pourrait optimiser en ne renvoyant que si Score < msg.Score)
		if candidate.Score < msg.Organism.Score {
			// On a trouvé mieux !
			encoder.Encode(NetworkMessage{Organism: candidate})
		} else {
			// Pas mieux, on renvoie l'original pour dire "j'ai fini ce tour"
			// (Dans une version avancée, on pourrait boucler localement plus longtemps)
			encoder.Encode(NetworkMessage{Organism: candidate})
		}
	}
}
