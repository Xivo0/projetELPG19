const net = require('net');
const readline = require('readline');

const PORT = 26000;

const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});


rl.question('Entrez l\'adresse IP du serveur (laissez vide pour localhost) : ', (ip) => {
    const HOST = ip.trim() || 'localhost'; // localhost par défaut si vide

    const client = new net.Socket();

    client.connect(PORT, HOST, () => {
        console.log(`Connecté au serveur Flip 7 sur ${HOST} !`);
    });

    // Gestion des données reçues
    client.on('data', (data) => {
        process.stdout.write(data.toString()); 
    });

    client.on('close', () => {
        console.log('\nConnexion fermée par le serveur.');
        process.exit(0);
    });

    client.on('error', (err) => {
        console.error(`Erreur de connexion : ${err.message}`);
        process.exit(1);
    });

    // Envoi des commandes joueur au serveur
    rl.on('line', (line) => {
        client.write(line.trim());
    });
});
