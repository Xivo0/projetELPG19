# projetELPG19
Passage projet V1 à V2:
ajout d'un buffer afin de ne pas avoir à recalculer une nouvelle image (comme un malloc) dans la fonction "Render". Au lieu de créer une nouvelle image pour chaque worker (on jette image -> garbage collector la supprime // on prend la nouvelle image) , on réutilise la même zone mémoire où l'on efface l'ancienne image pour la remplacer par la version améliorée (on "efface" l'ancienne image -> on met la nouvelle à la place).
