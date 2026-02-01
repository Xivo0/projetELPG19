package main

import (
	"flag"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	mode := flag.String("mode", "server", "Mode: 'server' ou 'client'")
	addr := flag.String("addr", "localhost:8080", "IP du serveur (pour le client)")
	flag.Parse()

	if *mode == "server" {
		LaunchServer()
	} else {
		LaunchClient(*addr)
	}
}
