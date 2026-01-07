package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/png"
	"net"
	"os"
	"sync"
)

var (
	bestOrganism Organism
	bestMutex    sync.Mutex //Sert pour le Lock/Delock pour empecher une double execution
	avgColor     color.RGBA // Couleur de fond
)

func LaunchServer() {
	fmt.Println("=== MODE SERVEUR (V2) ===")

	// 1. Charger l'image cible
	targetImg := LoadImage(InputFile)
	MaxX = targetImg.Bounds().Dx()
	MaxY = targetImg.Bounds().Dy()
	fmt.Printf("Cible: %dx%d\n", MaxX, MaxY)

	// 2. Calculer la couleur moyenne (Fond) Pour démarrer les calculs (car on a un score de base réduit comparé à une image noire)
	avgColor = ComputeAverageColor(targetImg)

	// 3. Initialiser le premier organisme (Vide mais scoré correctement)
	// On crée une image vide remplie de la couleur moyenne
	startImg := image.NewRGBA(targetImg.Bounds())
	RenderToBuffer([]Shape{}, startImg, avgColor)
	startScore := DiffEuclidienne(startImg, targetImg)

	bestOrganism = Organism{DNA: []Shape{}, Score: startScore}
	fmt.Printf("Score initial (Fond uni) : %.0f\n", startScore)

	// 4. Réseau
	ln, err := net.Listen("tcp", ":8080")//à expliquer les fonctions réseau
	if err != nil { panic(err) }
	fmt.Println("En attente de clients sur le port 8080...")

	generation := 0
	for {
		conn, err := ln.Accept()
		if err != nil { continue }
		go handleClient(conn, &generation)
	}
}

func handleClient(conn net.Conn, generation *int) {
	defer conn.Close()
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	//A EXPLIQUER TOUT LE RESTE
	for {
		// A. Envoyer
		bestMutex.Lock()
		err := encoder.Encode(NetworkMessage{Organism: bestOrganism})
		bestMutex.Unlock()
		if err != nil { return }

		// B. Recevoir
		var msg NetworkMessage
		err = decoder.Decode(&msg)
		if err != nil { return }

		// C. Mettre à jour
		bestMutex.Lock()
		if msg.Organism.Score < bestOrganism.Score {
			*generation++
			bestOrganism = msg.Organism
			fmt.Printf("[Gen %d] Record: %.0f (Client %s)\n", *generation, bestOrganism.Score, conn.RemoteAddr())

			if *generation%SaveFrequency == 0 {
				// Sauvegarde V2 : On utilise RenderToBuffer sur une image temporaire
				tempImg := image.NewRGBA(image.Rect(0, 0, MaxX, MaxY))
				RenderToBuffer(bestOrganism.DNA, tempImg, avgColor)
				f, _ := os.Create(OutputFile)
				png.Encode(f, tempImg)
				f.Close()
			}
		}
		bestMutex.Unlock()
	}
}
