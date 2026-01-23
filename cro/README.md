# projetELPG19
Passage projet V1 à V2:
ajout d'un buffer afin de ne pas avoir à recalculer une nouvelle image (comme un malloc) dans la fonction "Render". Au lieu de créer une nouvelle image pour chaque worker (on jette image -> garbage collector la supprime // on prend la nouvelle image) , on réutilise la même zone mémoire où l'on efface l'ancienne image pour la remplacer par la version améliorée (on "efface" l'ancienne image -> on met la nouvelle à la place).


il nous reste la mesure de performance pour tant de worker sur tant de generation; en gros faire en sorte que l'on nous demande le nombre de worker que l'on veut faire fonctionner et le nombre de generation que l'on veut faire au maximum.
