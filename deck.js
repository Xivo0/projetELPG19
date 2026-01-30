function creerCarte(type, valeur, nom, effect = null) {
    return { type, valeur, nom, effect };
}

function genererPaquet() {
    let paquet = [];
    paquet.push(creerCarte('nombre', 0, '0'));
    paquet.push(creerCarte('nombre', 1, '1'));
    for (let i = 2; i <= 12; i++) {
        for (let j = 0; j < i; j++) 
            paquet.push(creerCarte('nombre', i, i.toString()));
    }
    paquet.push(creerCarte('modifier', 2, '+2'));
    paquet.push(creerCarte('modifier', 4, '+4'));
    paquet.push(creerCarte('modifier', 6, '+6'));
    paquet.push(creerCarte('modifier', 8, '+8'));
    paquet.push(creerCarte('modifier', 10, '+10'));
    paquet.push(creerCarte('multiplicateur', 2, 'x2'));
    for (let i = 0; i < 3; i++) 
        paquet.push(creerCarte('action', 0, 'FREEZE (Gel)', 'freeze'));
    for (let i = 0; i < 3; i++) 
        paquet.push(creerCarte('action', 0, 'FLIP 3 (Piocher 3)', 'flip3'));
    for (let i = 0; i < 3; i++) 
        paquet.push(creerCarte('action', 0, 'SECONDE CHANCE', 'second_chance'));
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
    let sommeNombres = 0;
    let sommeBonus = 0;
    let aMultiplicateur = false;
    for (const carte of main) {
        if (carte.type === 'nombre') 
            sommeNombres += carte.valeur;
        else if (carte.type === 'modifier') 
            sommeBonus += carte.valeur;
        else if (carte.type === 'multiplicateur') 
            aMultiplicateur = true;
    }
    let scoreTotal = sommeNombres;
    if (aMultiplicateur) 
        scoreTotal *= 2;
    scoreTotal += sommeBonus;
    if (aFaitFlip7) 
        scoreTotal += 15;
    return scoreTotal;
}
module.exports = { genererPaquet, melangerPaquet, calculerScore };