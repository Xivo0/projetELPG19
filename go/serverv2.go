package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net"
	"os"
	"sync"
)

var (
	bestOrganism   Organism
	bestMutex      sync.Mutex
	serverAvgColor color.RGBA
)

func LaunchServer() {
	fmt.Println("=== SERVEUR V2 ===")
	targetImg := LoadImage(InputFile)
	MaxX, MaxY = targetImg.Bounds().Dx(), targetImg.Bounds().Dy()
	serverAvgColor = ComputeAverageColor(targetImg)

	// Init Organisme vide avec fond couleur moyenne
	startImg := image.NewRGBA(targetImg.Bounds())
	RenderToBuffer([]Shape{}, startImg, serverAvgColor)
	bestOrganism = Organism{DNA: []Shape{}, Score: DiffEuclidienne(startImg, targetImg)}

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	fmt.Println("Serveur écoute sur 8080...")

	gen := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn, &gen)
	}
}

func handleClient(conn net.Conn, gen *int) {
	defer conn.Close()
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	// Buffer pour sauvegarde
	saveBuffer := image.NewRGBA(image.Rect(0, 0, MaxX, MaxY))

	for {
		// A. Envoyer le Best actuel au client
		bestMutex.Lock()
		err := enc.Encode(NetworkMessage{Organism: bestOrganism})
		bestMutex.Unlock()
		if err != nil {
			return
		}

		// B. Attendre le retour du client
		var msg NetworkMessage
		err = dec.Decode(&msg)
		if err != nil {
			return
		}

		// C. Mettre à jour si mieux
		bestMutex.Lock()
		if msg.Organism.Score < bestOrganism.Score {
			*gen++
			bestOrganism = msg.Organism
			fmt.Printf("[Gen %d] Nouveau Record: %.0f (Client %s)\n", *gen, bestOrganism.Score, conn.RemoteAddr())

			// Sauvegarde
			if *gen%20 == 0 { // Sauvegarde plus fréquente car serveur centralisé
				RenderToBuffer(bestOrganism.DNA, saveBuffer, serverAvgColor)
				f, _ := os.Create(OutputFile)
				png.Encode(f, saveBuffer)
				f.Close()
			}
		}
		bestMutex.Unlock()
	}
}
