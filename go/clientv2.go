package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"net"
	"runtime"
	"sync"
)

// BATCH_SIZE : Chaque worker fait 50 essais avant de se synchroniser.
// Avec 20 workers, ça fait 1000 essais locaux entre chaque communication réseau.
const LocalBatchSize = 50

func LaunchClient(serverAddr string) {
	fmt.Println("=== CLIENT V2 (MULTI-CORE) ===")
	targetImg := LoadImage(InputFile)
	MaxX, MaxY = targetImg.Bounds().Dx(), targetImg.Bounds().Dy()
	avgColor := ComputeAverageColor(targetImg)

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	nbWorkers := runtime.NumCPU()
	fmt.Printf("CPU Locaux: %d. Démarrage du pool local...\n", nbWorkers)

	for {
		// 1. Recevoir le "Chef d'oeuvre" du serveur
		var msg NetworkMessage
		err := decoder.Decode(&msg)
		if err != nil {
			return
		}

		serverOrganism := msg.Organism

		// 2. Lancer la compétition locale (Map-Reduce)
		// On utilise un WaitGroup pour attendre que tous les cœurs aient fini leur batch
		var wg sync.WaitGroup
		localResults := make(chan Organism, nbWorkers)

		for w := 0; w < nbWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Chaque worker a son propre buffer (Comme PROJETV2)
				workerBuffer := image.NewRGBA(targetImg.Bounds())

				// On part de la version du serveur
				bestLocal := serverOrganism.Copy()

				// Boucle locale rapide (Batch)
				for i := 0; i < LocalBatchSize; i++ {
					candidate := bestLocal.Copy()
					Mutate(&candidate, targetImg)
					RenderToBuffer(candidate.DNA, workerBuffer, avgColor)
					candidate.Score = DiffEuclidienne(workerBuffer, targetImg)

					if candidate.Score < bestLocal.Score {
						bestLocal = candidate
					}
				}
				// On renvoie notre champion local
				localResults <- bestLocal
			}()
		}

		// 3. Attendre tout le monde
		wg.Wait()
		close(localResults)

		// 4. Trouver le meilleur parmi tous les workers locaux
		bestOfBatch := serverOrganism
		improved := false

		for res := range localResults {
			if res.Score < bestOfBatch.Score {
				bestOfBatch = res
				improved = true
			}
		}

		// 5. Renvoyer au serveur (seulement si on a trouvé mieux)
		if improved {
			// fmt.Printf("Amélioration locale trouvée (%.0f)\n", bestOfBatch.Score)
			encoder.Encode(NetworkMessage{Organism: bestOfBatch})
		} else {
			// Sinon on renvoie l'original pour dire "J'ai fini, donne la suite"
			encoder.Encode(NetworkMessage{Organism: serverOrganism})
		}
	}
}
