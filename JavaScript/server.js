const net = require('net');

// --- DONNÉES ET FONCTIONS DU JEU (Inchangées) ---
function creerCarte(type, valeur, nom, effect = null) { return { type, valeur, nom, effect }; }

function genererPaquet() {
    let paquet = [];
    paquet.push(creerCarte('nombre', 0, '0')); 
    paquet.push(creerCarte('nombre', 1, '1')); 
    for (let i = 2; i <= 12; i++) {
        for (let j = 0; j < i; j++) paquet.push(creerCarte('nombre', i, i.toString()));
    }
    paquet.push(creerCarte('modifier', 2, '+2'));
    paquet.push(creerCarte('modifier', 4, '+4'));
    paquet.push(creerCarte('modifier', 6, '+6'));
    paquet.push(creerCarte('modifier', 8, '+8'));
    paquet.push(creerCarte('modifier', 10, '+10'));
    paquet.push(creerCarte('multiplicateur', 2, 'x2'));
    for (let i = 0; i < 3; i++) paquet.push(creerCarte('action', 0, 'FREEZE (Gel)', 'freeze'));
    for (let i = 0; i < 3; i++) paquet.push(creerCarte('action', 0, 'FLIP 3 (Piocher 3)', 'flip3'));
    for (let i = 0; i < 3; i++) paquet.push(creerCarte('action', 0, 'SECONDE CHANCE', 'second_chance'));
    return paquet;
}

function melangerPaquet(paquet) {
    for (let i = paquet.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [paquet[i], paquet[j]] = [paquet[j], paquet[i]];
    }
    return paquet;
}

function calculerScore(main, aFaitFlip7) {
    let sommeNombres = 0, sommeBonus = 0, aMultiplicateur = false;
    for (const carte of main) {
        if (carte.type === 'nombre') sommeNombres += carte.valeur;
        else if (carte.type === 'modifier') sommeBonus += carte.valeur;
        else if (carte.type === 'multiplicateur') aMultiplicateur = true;
    }
    let scoreTotal = sommeNombres;
    if (aMultiplicateur) scoreTotal *= 2;
    scoreTotal += sommeBonus;
    if (aFaitFlip7) scoreTotal += 15;
    return scoreTotal;
}

// --- GESTION RÉSEAU ---

function broadcast(joueurs, message) {
    joueurs.forEach(j => {
        if(!j.socket.destroyed) j.socket.write(message + '\n');
    });
    console.log(message);
}

function tell(joueur, message) {
    if(!joueur.socket.destroyed) joueur.socket.write(message + '\n');
}

function demanderAuJoueur(joueur, question) {
    return new Promise((resolve) => {
        tell(joueur, question);
        // On écoute une seule fois ('once') la réponse
        joueur.socket.once('data', (data) => {
            resolve(data.toString().trim().toLowerCase());
        });
    });
}

// --- LOGIQUE DU JEU ---

async function jouerTour(joueur, tousLesJoueurs, paquet) {
    broadcast(tousLesJoueurs, `\n--- Tour de ${joueur.nom} ---`);
    
    let main = [];
    let nombresUniques = [];
    let aSecondeChance = false;
    let cartesAPiocherForcees = 0;

    while (true) {
        if (cartesAPiocherForcees === 0) {
            tell(joueur, `Votre main: ${calculerScore(main, false)} pts.`);
            let reponse = await demanderAuJoueur(joueur, "Piocher ? (o/n) : ");
            
            if (reponse !== 'o') {
                broadcast(tousLesJoueurs, `${joueur.nom} décide de s'arrêter.`);
                break; 
            }
        } else {
            broadcast(tousLesJoueurs, `> ${joueur.nom} doit piocher (Reste: ${cartesAPiocherForcees})`);
            cartesAPiocherForcees--;
        }

        if (paquet.length === 0) { broadcast(tousLesJoueurs, "Paquet vide !"); break; }
        
        const carte = paquet.shift();
        main.push(carte);
        broadcast(tousLesJoueurs, `   -> Carte tirée : [ ${carte.nom} ]`);

        if (carte.effect === 'freeze') {
            broadcast(tousLesJoueurs, "   ❄️  FREEZE ! Perdu !");
            return 0;
        }
        if (carte.effect === 'flip3') cartesAPiocherForcees += 3;
        if (carte.effect === 'second_chance') aSecondeChance = true;

        if (carte.type === 'nombre') {
            if (nombresUniques.includes(carte.valeur)) {
                if (aSecondeChance) {
                    broadcast(tousLesJoueurs, `   ❤️  Doublon ${carte.valeur} sauvé par Seconde Chance !`);
                    
                    // CORRECTION : On applique la règle de défausse 
                    
                    // 1. On retire le doublon qu'on vient d'ajouter (c'est la dernière carte de 'main')
                    main.pop(); 

                    // 2. On cherche et on retire la carte "SECONDE CHANCE" de la main
                    const indexSC = main.findIndex(c => c.effect === 'second_chance');
                    if (indexSC !== -1) {
                        main.splice(indexSC, 1); // On la supprime
                    }

                    // 3. On désactive le bouclier
                    aSecondeChance = false;

                    // Note : Le score ne doit PAS augmenter.
                } else {
                    broadcast(tousLesJoueurs, `   💥 Doublon (${carte.valeur}) ! Tour perdu.`);
                    return 0;
                }
            } else {
                nombresUniques.push(carte.valeur);
            }
        }

        if (nombresUniques.length >= 7) {
            broadcast(tousLesJoueurs, "   🎉 FLIP 7 ! Bonus 15 points !");
            return calculerScore(main, true);
        }
    }
    return calculerScore(main, false);
}

async function lancerPartie(joueursConnectes) {
    jeuEnCours = true; // Verrouille le lobby
    broadcast(joueursConnectes, "=== LA PARTIE COMMENCE ! ===");
    
    // Nettoyage des écouteurs 'data' du lobby pour éviter les conflits
    joueursConnectes.forEach(j => j.socket.removeAllListeners('data'));
    
    let scores = joueursConnectes.map(() => 0);
    let partieTerminee = false;
    let round = 1;

    while (!partieTerminee) {
        broadcast(joueursConnectes, `\n=== MANCHE ${round} ===`);
        let paquet = melangerPaquet(genererPaquet());

        // La boucle s'adapte automatiquement à N joueurs
        for (let i = 0; i < joueursConnectes.length; i++) {
            // Vérification si le joueur est toujours connecté
            if (joueursConnectes[i].socket.destroyed) {
                broadcast(joueursConnectes, `${joueursConnectes[i].nom} s'est déconnecté.`);
                continue;
            }

            let points = await jouerTour(joueursConnectes[i], joueursConnectes, paquet);
            scores[i] += points;
            broadcast(joueursConnectes, `>>> Total ${joueursConnectes[i].nom} : ${scores[i]}`);
        }

        if (scores.some(s => s >= 200)) partieTerminee = true;
        else round++;
    }

    // Fin de partie
    let maxScore = Math.max(...scores);
    scores.forEach((s, i) => {
        if (s === maxScore) broadcast(joueursConnectes, `🏆 ${joueursConnectes[i].nom} GAGNE avec ${s} points !`);
    });
    
    broadcast(joueursConnectes, "Le serveur va s'arrêter. Merci d'avoir joué !");
    process.exit(0);
}

// --- GESTION DU LOBBY (SALLE D'ATTENTE) ---

const PORT = 26000;
let joueursEnAttente = [];
let jeuEnCours = false;

const server = net.createServer((socket) => {
    
    // Si la partie a déjà commencé, on refuse les nouveaux
    if (jeuEnCours) {
        socket.write("Désolé, la partie est déjà en cours !\n");
        socket.destroy();
        return;
    }

    const nomJoueur = `Joueur ${joueursEnAttente.length + 1}`;
    const joueur = { nom: nomJoueur, socket: socket };
    joueursEnAttente.push(joueur);

    console.log(`${nomJoueur} connecté.`);
    
    // Message de bienvenue différent selon si c'est l'hôte (J1) ou les autres
    if (joueursEnAttente.length === 1) {
        socket.write(`Bienvenue ${nomJoueur} ! Vous êtes l'Hôte.\n`);
        socket.write(`Attendez que tout le monde soit là, puis tapez START pour lancer.\n`);
    } else {
        socket.write(`Bienvenue ${nomJoueur} ! En attente de l'Hôte pour lancer la partie...\n`);
        // On prévient tout le monde qu'un nouveau est arrivé
        broadcast(joueursEnAttente, `>> ${nomJoueur} a rejoint la partie (${joueursEnAttente.length} joueurs).`);
    }

    // ÉCOUTEUR DU LOBBY (Pour détecter "START")
    socket.on('data', (data) => {
        // On ignore les messages si le jeu a commencé (car c'est lancerPartie qui gère)
        if (jeuEnCours) return;

        const message = data.toString().trim();
        
        // Seul le Joueur 1 peut lancer
        if (joueur === joueursEnAttente[0] && message.toUpperCase() === 'START') {
            if (joueursEnAttente.length < 1) {
                socket.write("Il faut au moins 1 joueur pour commencer !\n");
            } else {
                console.log("Lancement de la partie !");
                lancerPartie(joueursEnAttente);
            }
        }
    });
    
    // Gestion déconnexion dans le lobby
    socket.on('close', () => {
        if (!jeuEnCours) {
            joueursEnAttente = joueursEnAttente.filter(j => j !== joueur);
            broadcast(joueursEnAttente, `<< ${nomJoueur} a quitté le lobby.`);
        }
    });

    socket.on('error', (err) => {
        console.log(`Erreur avec ${nomJoueur}: ${err.message}`);
    });
});

// IMPORTANT : '0.0.0.0' permet d'accepter les connexions de TOUS les PC du réseau
server.listen(PORT, '0.0.0.0', () => {
    console.log(`Serveur démarré sur le port ${PORT}.`);
    console.log(`En attente des joueurs... (Le Joueur 1 devra taper START)`);
});