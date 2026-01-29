const net = require('net');
const readline = require('readline');

// CONFIGURATION
const PORT = 26000;
// Remplace 'localhost' par l'IP du serveur si tu joues sur deux PC différents
// Exemple : const HOST = '192.168.1.15'; 
const HOST = '10.56.112.92'; 

const client = new net.Socket();

// Interface pour lire le clavier du joueur
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

client.connect(PORT, HOST, () => {
    console.log('Connecté au serveur Flip 7 !');
});

// Quand le serveur nous parle
client.on('data', (data) => {
    console.log(data.toString());
});

// Quand le serveur coupe la connexion
client.on('close', () => {
    console.log('Connexion fermée');
    process.exit(0);
});

// Quand le joueur tape au clavier, on envoie au serveur
rl.on('line', (line) => {
    client.write(line.trim());
});