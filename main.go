package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	// Initialisation de l'aléatoire
	rand.Seed(time.Now().UnixNano())

	// Lecture des arguments (ex: -mode=client)
	mode := flag.String("mode", "server", "Mode: 'server' ou 'client'")
	addr := flag.String("addr", "localhost:8080", "Adresse du serveur (pour le client)")
	
	flag.Parse()

	if *mode == "server" {
		LaunchServer() // Défini dans server.go
	} else if *mode == "client" {
		LaunchClient(*addr) // Défini dans client.go
	} else {
		fmt.Println("Mode inconnu. Utilisez -mode=server ou -mode=client")
	}
}
