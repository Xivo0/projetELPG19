package main

import (
	"flag"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Gestion des arguments ligne de commande
	mode := flag.String("mode", "server", "Mode de lancement: 'server' ou 'client'")
	addr := flag.String("addr", "localhost:8080", "Adresse du serveur (pour le client)")
	
	flag.Parse()

	if *mode == "server" {
		LaunchServer()
	} else {
		LaunchClient(*addr)
	}
}
