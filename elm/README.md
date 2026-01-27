Un jeu de devinette interactif développé en **Elm**. Le programme charge une liste de mots depuis un fichier local, en choisit un au hasard, et récupère ses définitions via une API externe pour aider le joueur.

- **Chargement dynamique** : Récupère une liste de mots depuis un fichier `words.txt`.
- **Intégration API** : Utilise l'API [Free Dictionary](https://dictionaryapi.dev/) pour afficher les définitions en temps réel.
- **Mode Triche** : Possibilité d'afficher/cacher le mot secret pour les joueurs en difficulté.
- **Compiler** : Bien penser à heberger un serveur en local sur le port 8000 (via python par exemple.)
- **lancer**: -elm make propro.elm --output= index2.html

             -Rajouter la ligne : <link rel="stylesheet" href="style.css"> dans le </head> du index2.html

             -heberger un serveur local sur le port 8000
  
