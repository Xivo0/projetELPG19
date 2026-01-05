package main

import (
	"encoding/gob"
	"fmt"
	"image/png"
	"net"
	"os"
	"sync"
	"time"
)

var (
	bestOrganism Organism
	bestMutex    sync.Mutex
)

func LaunchServer() {
	fmt.Println("=== MODE SERVEUR ===")
	
	// 1. Charger l'image cible
	targetImg := LoadImage("PAIN.png")
	MaxX = targetImg.Bounds().Dx()
	MaxY = targetImg.Bounds().Dy()
	fmt.Printf("Cible chargée: %dx%d\n", MaxX, MaxY)

	// 2. Initialiser le meilleur organisme (vide)
	bestOrganism = Organism{DNA: []Shape{}, Score: 1e15}

	[cite_start]// 3. Ouvrir le port TCP [cite: 223]
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	fmt.Println("Serveur écoute sur le port 8080...")

	generation := 0

	for {
		// 4. Accepter connexion
		[cite_start]conn, err := ln.Accept() // [cite: 225]
		if err != nil {
			continue
		}
		
		// 5. Déléguer à une goroutine
		[cite_start]go handleClient(conn, &generation) // [cite: 234]
	}
}

func handleClient(conn net.Conn, generation *int) {
	defer conn.Close()
	fmt.Println("Nouveau client connecté:", conn.RemoteAddr())

	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	for {
		// A. Envoyer le meilleur organisme actuel au client
		bestMutex.Lock()
		msgToSend := NetworkMessage{Organism: bestOrganism} // Copie implicite
		bestMutex.Unlock()

		err := encoder.Encode(msgToSend)
		if err != nil {
			fmt.Println("Client déconnecté (envoi)")
			return
		}

		// B. Recevoir la proposition du client
		var receivedMsg NetworkMessage
		[cite_start]err = decoder.Decode(&receivedMsg) // Bloquant [cite: 239]
		if err != nil {
			fmt.Println("Client déconnecté (réception)")
			return
		}

		// C. Vérifier si c'est mieux
		bestMutex.Lock()
		if receivedMsg.Organism.Score < bestOrganism.Score {
			*generation++
			bestOrganism = receivedMsg.Organism
			fmt.Printf("[Gen %d] Record battu ! Score: %.0f (par %s)\n", 
				*generation, bestOrganism.Score, conn.RemoteAddr())

			// Sauvegarde régulière
			if *generation%SaveFrequency == 0 {
				// On recrée l'image pour la sauvegarder
				img := Render(bestOrganism.DNA, MaxX, MaxY)
				saveFile, _ := os.Create(OutputFile)
				png.Encode(saveFile, img)
				saveFile.Close()
			}
		}
		bestMutex.Unlock()
	}
}
