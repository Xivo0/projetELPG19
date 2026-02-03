- **fichiers** : Verifier d'avoir tout les fichiers (clientv2.go;commonv2.go;mainv2.go;serverv2.go;target.png) dans le meme dossier.
- **Lancement**: Ouvrir 1 terminal pour le server et un ou plusieurs autre pour le/les clients.
                 - Terminal 1 (Serveur), Lancez la commande suivante: go run . -mode server
                 - Terminal 2 (Client),  Lancez la commande suivante: go run . -mode client
                 - Note pour machine distante : Si vous lancez le client sur une machine extérieure (sur le même réseau),
                                                spécifiez l'adresse IP du serveur : go run . -mode client -addr IP_DU_SERVEUR:8080
