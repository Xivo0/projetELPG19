// server.js
const net = require('net');
const { lancerPartie } = require('./game');
const { broadcast } = require('./utils');

const PORT = 26000;
let joueursEnAttente = [];
let jeuEnCours = false;

const server = net.createServer((socket) => {
    if (jeuEnCours) {
        socket.write("Désolé, la partie est déjà en cours !\n");
        socket.destroy();
        return;
    }

    const nomJoueur = `Joueur ${joueursEnAttente.length + 1}`;
    const joueur = { nom: nomJoueur, socket: socket };
    joueursEnAttente.push(joueur);

    console.log(`${nomJoueur} connecté.`);

    if (joueursEnAttente.length === 1) {
        socket.write(`\nBienvenue ${nomJoueur} ! Vous êtes l'Hôte.\n`);
        socket.write(`Attendez les autres joueurs, puis tapez START pour lancer.\n`);
    } else {
        socket.write(`\nBienvenue ${nomJoueur} ! En attente du lancement...\n`);
        broadcast(joueursEnAttente, `>> ${nomJoueur} a rejoint la partie (${joueursEnAttente.length} joueurs).`);
    }

    socket.on('data', (data) => {
        if (jeuEnCours) return;
        const message = data.toString().trim();
        // Commande START (uniquement l'hôte)
        if (joueur === joueursEnAttente[0] && message.toUpperCase() === 'START') {
            if (joueursEnAttente.length < 1) { // Mettre < 2 pour forcer le multijoueur
                socket.write("Il faut au moins 1 joueur pour commencer !\n");
            } else {
                console.log("Lancement de la partie !");
                lancerPartie(joueursEnAttente);
            }
        }
    });

    socket.on('close', () => {
        if (!jeuEnCours) {
            joueursEnAttente = joueursEnAttente.filter(j => j !== joueur);
            broadcast(joueursEnAttente, `<< ${nomJoueur} a quitté le lobby.`);
        }
    });

    socket.on('error', (err) => console.log(`Erreur avec ${nomJoueur}: ${err.message}`));
});

server.listen(PORT, '0.0.0.0', () => {
    console.log(`Serveur démarré sur le port ${PORT}.`);
});