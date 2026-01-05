package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

// Variable partagée
var (
	localBest      Organism
	localBestMutex sync.RWMutex
)

func LaunchClient(serverAddr string) {
	nbCoeurs := runtime.NumCPU()
	fmt.Printf("=== SUPER-CLIENT (%d Cœurs) ===\n", nbCoeurs)

	// 1. Chargement
	fmt.Println("[1/4] Chargement image locale...")
	targetImg := LoadImage(InputFile)
	MaxX = targetImg.Bounds().Dx()
	MaxY = targetImg.Bounds().Dy()
	avgColor := ComputeAverageColor(targetImg)
	fmt.Printf("      Image chargée (%dx%d)\n", MaxX, MaxY)

	// 2. Connexion
	fmt.Printf("[2/4] Connexion au serveur %s...\n", serverAddr)
	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		fmt.Println("❌ ERREUR CONNEXION :", err)
		return
	}
	defer conn.Close()
	fmt.Println("✅ [3/4] CONNECTÉ AU SERVEUR !")

	sendChan := make(chan Organism, nbCoeurs)
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	// --- A. RÉCEPTIONNISTE ---
	go func() {
		fmt.Println("      (Réceptionniste prêt)")
		firstMsg := true
		for {
			var msg NetworkMessage
			err := decoder.Decode(&msg)
			if err != nil {
				fmt.Println("\n❌ Serveur déconnecté (Lecture).")
				os.Exit(1)
			}

			if firstMsg {
				fmt.Printf("✅ [4/4] CONFIG INITIALE REÇUE (Score à battre: %.0f)\n", msg.Organism.Score)
				fmt.Println("🔨 Les 20 ouvriers démarrent... (Patience !)")
				firstMsg = false
			}

			localBestMutex.Lock()
			// On accepte la mise à jour si c'est mieux ou si c'est la première fois
			if localBest.DNA == nil || msg.Organism.Score < localBest.Score {
				localBest = msg.Organism
			}
			localBestMutex.Unlock()
		}
	}()

	// --- B. EXPÉDITEUR ---
	go func() {
		fmt.Println("      (Expéditeur prêt)")
		for candidate := range sendChan {
			err := encoder.Encode(NetworkMessage{Organism: candidate})
			if err != nil {
				fmt.Println("\n❌ Serveur déconnecté (Ecriture).")
				os.Exit(1)
			}
			fmt.Printf("\n🚀 ENVOI AMÉLIORATION (Score: %.0f)", candidate.Score)
		}
	}()

	// --- C. OUVRIERS ---
	for w := 0; w < nbCoeurs; w++ {
		go func(id int) {
			myBuffer := image.NewRGBA(targetImg.Bounds())
			const BatchSize = 200 // Taille du lot

			for {
				// 1. Récupérer la référence
				localBestMutex.RLock()
				if localBest.DNA == nil {
					localBestMutex.RUnlock()
					time.Sleep(10 * time.Millisecond)
					continue
				}
				candidate := localBest
				// Copie Profonde du DNA (Indispensable)
				newDNA := make([]Shape, len(localBest.DNA))
				copy(newDNA, localBest.DNA)
				candidate.DNA = newDNA
				localBestMutex.RUnlock()

				// 2. Travailler (Batch)
				bestOfBatch := candidate
				improved := false

				for i := 0; i < BatchSize; i++ {
					// Copie temp pour mutation
					temp := bestOfBatch
					tempDNA := make([]Shape, len(bestOfBatch.DNA))
					copy(tempDNA, bestOfBatch.DNA)
					temp.DNA = tempDNA

					// Mutate
					progress := float64(len(temp.DNA)) / TargetComplexity
					if progress > 1.0 { progress = 1.0 }
					Mutate(&temp, targetImg, progress)

					// Render
					RenderToBuffer(temp.DNA, myBuffer, avgColor)
					temp.Score = DiffEuclidienne(myBuffer, targetImg)

					if temp.Score < bestOfBatch.Score {
						bestOfBatch = temp
						improved = true
					}
				}

				// 3. Feedback Visuel (Juste un point pour montrer la vie)
				// Le worker 0 affiche des points pour tout le monde
				if id == 0 {
					fmt.Print(".")
				}

				// 4. Envoyer si mieux
				if improved {
					localBestMutex.RLock()
					isStillBetter := bestOfBatch.Score < localBest.Score
					localBestMutex.RUnlock()

					if isStillBetter {
						sendChan <- bestOfBatch
					}
				}
			}
		}(w)
	}

	// Garder le main en vie
	select {}
}
