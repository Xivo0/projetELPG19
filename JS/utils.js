const style = {
    reset: "\x1b[0m", bright: "\x1b[1m", dim: "\x1b[2m",
    red: "\x1b[31m", green: "\x1b[32m", yellow: "\x1b[33m",
    blue: "\x1b[34m", magenta: "\x1b[35m", cyan: "\x1b[36m", white: "\x1b[37m",
    bgRed: "\x1b[41m", bgGreen: "\x1b[42m", bgBlue: "\x1b[44m"
};

const icones = {
    freeze: "❄️", flip3: "🚀", second_chance: "🍀",
    modifier: "🔧", multiplicateur: "✖️", mort: "💀", win: "🏆", carte: "🃏"
};

function formatCarte(carte) {
    let s = style.reset + style.bright;
    let contenu = "";
    if (carte.type === 'nombre') { s += style.cyan; contenu = ` ${carte.valeur} `; }
    else if (carte.type === 'modifier') { s += style.green; contenu = `+${carte.valeur}`; }
    else if (carte.type === 'multiplicateur') { s += style.magenta; contenu = `x2`; }
    else if (carte.effect) {
        s += style.yellow;
        if (carte.effect === 'freeze') contenu = " FREEZE ";
        if (carte.effect === 'flip3') contenu = " FLIP 3 ";
        if (carte.effect === 'second_chance') contenu = " 2ND CHANCE ";
    }
    return `${style.white}[${s}${contenu}${style.white}]${style.reset}`;
}

function afficherHUD(joueur, calculerScoreFn) {
    const score = calculerScoreFn(joueur.main, false);
    const mainVisuelle = joueur.main.map(c => formatCarte(c)).join(' ');
    let info = `\n${style.blue}════════════════════════════════════════════════════${style.reset}\n`;
    info += `${style.bright} JOUEUR : ${joueur.nom} ${style.reset}\n`;
    info += ` SCORE  : ${style.green}${score} pts${style.reset}\n`;
    info += ` MAIN   : ${mainVisuelle}\n`;
    info += `${style.blue}════════════════════════════════════════════════════${style.reset}`;
    return info;
}


function broadcast(joueurs, message) {
    joueurs.forEach(j => {
        if (!j.socket.destroyed) 
            j.socket.write(message + '\n');
    });
    console.log(message);
}

function tell(joueur, message) {
    if (!joueur.socket.destroyed) 
        joueur.socket.write(message + '\n');
}

function demanderAuJoueur(joueur, question) {
    return new Promise((resolve) => {
        tell(joueur, question);
        joueur.socket.once('data', (data) => {
            resolve(data.toString().trim().toLowerCase());
        });
    });
}

const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

module.exports = { style, icones, formatCarte, afficherHUD, broadcast, tell, demanderAuJoueur, sleep };