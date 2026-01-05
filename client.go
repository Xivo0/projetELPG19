package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"net"
)

func LaunchClient(serverAddr string) {
	fmt.Println("=== MODE CLIENT (V2 + BATCHING) ===")

	// ... (Chargement image et initialisation identiques) ...
	targetImg := LoadImage(InputFile)
	MaxX = targetImg.Bounds().Dx()
	MaxY = targetImg.Bounds().Dy()
	avgColor := ComputeAverageColor(targetImg)
	workBuffer := image.NewRGBA(targetImg.Bounds())

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil { panic(err) }
	defer conn.Close()
	
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	// CONFIGURATION DU BATCH
	// Le client va faire 500 essais avant de parler au serveur
	const BatchSize = 500 

	for {
		// 1. Recevoir le point de départ du serveur
		var msg NetworkMessage
		err := decoder.Decode(&msg)
		if err != nil { return }

		// On garde le meilleur localement pour ce lot
		bestLocalCandidate := msg.Organism
		
		// 2. BOUCLE DE TRAVAIL LOCAL (Le Batch)
		for i := 0; i < BatchSize; i++ {
			// On part toujours du meilleur candidat connu (local ou serveur)
			candidate := bestLocalCandidate // Copie simple car Organism ne contient pas de pointeurs complexes (sauf slice DNA)
			
			// Attention : Il faut cloner le DNA pour ne pas modifier l'original par erreur
			// Go est tricky avec les slices :
			newDNA := make([]Shape, len(bestLocalCandidate.DNA))
			copy(newDNA, bestLocalCandidate.DNA)
			candidate.DNA = newDNA

			// Mutation
			progress := float64(len(candidate.DNA)) / TargetComplexity
			if progress > 1.0 { progress = 1.0 }
			Mutate(&candidate, targetImg, progress)

			// Rendu
			RenderToBuffer(candidate.DNA, workBuffer, avgColor)
			candidate.Score = DiffEuclidienne(workBuffer, targetImg)

			// Si c'est mieux que notre meilleur local, on le garde
			if candidate.Score < bestLocalCandidate.Score {
				bestLocalCandidate = candidate
			}
		}

		// 3. Fin du Batch : On envoie le résultat au serveur
		// On renvoie le meilleur qu'on a trouvé sur 500 essais
		if bestLocalCandidate.Score < msg.Organism.Score {
			// On a trouvé une pépite !
			encoder.Encode(NetworkMessage{Organism: bestLocalCandidate})
			fmt.Printf("✨ Envoi amélioration (Score %.0f)\n", bestLocalCandidate.Score)
		} else {
			// Rien trouvé d'intéressant dans ce lot, on renvoie l'original
			// pour dire au serveur qu'on est prêt pour la suite
			encoder.Encode(NetworkMessage{Organism: msg.Organism})
		}
	}
}
