const { genererPaquet, melangerPaquet, calculerScore } = require('./deck');
const { style, icones, formatCarte, afficherHUD, broadcast, tell, demanderAuJoueur, sleep } = require('./utils');

async function jouerTour(joueur, tousLesJoueurs, paquet, defausse) {
    if (joueur.elimine || joueur.aFini) return;

    let doitPiocher = false;

    if (joueur.cartesAPiocherForcees > 0) {
        doitPiocher = true;
    } else {
        tell(joueur, afficherHUD(joueur, calculerScore));
        broadcast(tousLesJoueurs.filter(j => j !== joueur), `${style.dim}>> C'est au tour de ${joueur.nom}...${style.reset}`);
        await sleep(800);

        let reponse = await demanderAuJoueur(joueur, `\n${style.yellow}Action > ${style.reset}Piocher une carte ? (${style.green}o${style.reset}/${style.red}n${style.reset}) : `);
        
        if (reponse === 'o') {
            doitPiocher = true;
        } else {
            broadcast(tousLesJoueurs, `${style.red}🛑 ${joueur.nom} s'arrête.${style.reset} Score final : ${style.bright}${calculerScore(joueur.main, false)}${style.reset}`);
            joueur.aFini = true;
            await sleep(1000);
            return;
        }
    }

    while (doitPiocher && !joueur.elimine && !joueur.aFini) {
        
        if (paquet.length === 0) {
            if (defausse.length > 0) {
                broadcast(tousLesJoueurs, `\n${style.yellow}🔄 La pioche est vide, remélange de la défausse...${style.reset}`);
                paquet.push(...melangerPaquet(defausse));
                defausse.length = 0; // On vide la défausse après transfert
                await sleep(1500);
            } else {
                broadcast(tousLesJoueurs, "⚠️ Plus aucune carte dans le jeu !");
                joueur.aFini = true;
                break;
            }
        }

        if (joueur.cartesAPiocherForcees > 0) {
            tell(joueur, afficherHUD(joueur, calculerScore));
            broadcast(tousLesJoueurs, `${style.magenta}⚠️ ${joueur.nom} DOIT piocher (Reste: ${joueur.cartesAPiocherForcees})${style.reset}`);
            joueur.cartesAPiocherForcees--;
            await sleep(1000);
        }

        const carte = paquet.shift();
        joueur.main.push(carte);
        broadcast(tousLesJoueurs, `${icones.carte} ${joueur.nom} a tiré : ${formatCarte(carte)}`);
        await sleep(1500);

        if (carte.effect === 'freeze') {
            broadcast(tousLesJoueurs, `${style.bgBlue}${style.white} ${icones.freeze} FREEZE ! ${joueur.nom} est gelé ! ${style.reset}`);
            joueur.elimine = true;
            joueur.scoreManche = 0;
            return;
        }
        if (carte.effect === 'flip3') {
            broadcast(tousLesJoueurs, `${style.magenta}${icones.flip3} FLIP 3 ! +3 cartes immédiates !${style.reset}`);
            joueur.cartesAPiocherForcees += 3;
            await sleep(1000);
        }
        if (carte.effect === 'second_chance') {
            broadcast(tousLesJoueurs, `${style.green}${icones.second_chance} SECONDE CHANCE acquise !${style.reset}`);
            joueur.aSecondeChance = true;
            await sleep(1000);
        }

        if (carte.type === 'nombre') {
            if (joueur.nombresUniques.includes(carte.valeur)) {
                if (joueur.aSecondeChance) {
                    await sleep(1000);
                    broadcast(tousLesJoueurs, `${style.green}❤️ Doublon ${carte.valeur} sauvé par SECONDE CHANCE !${style.reset}`);
                    joueur.main.pop(); 
                    const indexSC = joueur.main.findIndex(c => c.effect === 'second_chance');
                    if (indexSC !== -1) joueur.main.splice(indexSC, 1);
                    joueur.aSecondeChance = false;
                    await sleep(1500);
                } else {
                    broadcast(tousLesJoueurs, `${style.bgRed}${style.white} ${icones.mort} BOOM ! Doublon (${carte.valeur}) ! ${joueur.nom} a perdu. ${style.reset}`);
                    joueur.elimine = true;
                    joueur.scoreManche = 0;
                    return;
                }
            } else {
                joueur.nombresUniques.push(carte.valeur);
            }
        }

        if (joueur.nombresUniques.length >= 7) {
            broadcast(tousLesJoueurs, `${style.bgGreen}${style.black} ${icones.win} FLIP 7 ! ${joueur.nom} RÉUSSIT L'IMPOSSIBLE ! ${style.reset}`);
            joueur.aFini = true;
            joueur.aFaitFlip7 = true;
        }

        if (joueur.cartesAPiocherForcees === 0) {
            doitPiocher = false;
        }
    }
}

async function lancerPartie(joueursConnectes) {
    let jeuEnCours = true;
    
    let paquet = melangerPaquet(genererPaquet());
    let defausse = [];

    const banniere = `
${style.yellow}
    ███████╗██╗     ██╗██████╗     ███████╗
    ██╔════╝██║     ██║██╔══██╗    ╚════██║
    █████╗  ██║     ██║██████╔╝        ██╔╝
    ██╔══╝  ██║     ██║██╔═══╝        ██╔╝ 
    ██║     ███████╗██║██║            ██║  
    ╚═╝     ╚══════╝╚═╝╚═╝            ╚═╝  
${style.reset}
    `;
    
    broadcast(joueursConnectes, banniere);
    await sleep(800)
    broadcast(joueursConnectes, `${style.bgBlue}   LA PARTIE COMMENCE !   ${style.reset}\n`);

    let scores = joueursConnectes.map(() => 0);
    let partieTerminee = false;
    let round = 1;

    while (!partieTerminee) {
        broadcast(joueursConnectes, `\n--- MANCHE ${round} ---`);
        await sleep(1000);

        joueursConnectes.forEach(j => {
            j.main = [];
            j.nombresUniques = [];
            j.elimine = false;
            j.aFini = false;
            j.cartesAPiocherForcees = 0;
            j.aSecondeChance = false;
            j.scoreManche = 0;
            j.aFaitFlip7 = false;
        });

        let mancheEnCours = true;
        while (mancheEnCours) {
            let joueursActifs = 0;

            for (let i = 0; i < joueursConnectes.length; i++) {
                let j = joueursConnectes[i];
                if (!j.elimine && !j.aFini && !j.socket.destroyed) {
                    joueursActifs++;
                    await jouerTour(j, joueursConnectes, paquet, defausse);
                }
            }

            if (joueursActifs === 0) mancheEnCours = false;
        }

        broadcast(joueursConnectes, "\n--- Fin de la manche ---");
        await sleep(1000);
        for (let i = 0; i < joueursConnectes.length; i++) {
            let j = joueursConnectes[i];
            let pointsManche = 0;

            if (!j.elimine) {
                pointsManche = calculerScore(j.main, j.aFaitFlip7);
            }
            scores[i] += pointsManche;
            broadcast(joueursConnectes, `${j.nom} : +${pointsManche} (Total: ${scores[i]})`);

            // Envoi de toutes les cartes de la main vers la défausse 
            defausse.push(...j.main);

            await sleep(500);
        }

        if (scores.some(s => s >= 200)) partieTerminee = true;
        else round++;
    }

    let classement = joueursConnectes.map((joueur, index) => {
        return { nom: joueur.nom, score: scores[index] };
    });
    classement.sort((a, b) => b.score - a.score);

    broadcast(joueursConnectes, "\n🏆 ====== CLASSEMENT FINAL ====== 🏆");
    classement.forEach((c, index) => {
        broadcast(joueursConnectes, `${index + 1}. ${c.nom} : ${c.score} points`);
    });
    
    broadcast(joueursConnectes, "Le serveur va s'arrêter.");
    process.exit(0);
}

module.exports = { lancerPartie };