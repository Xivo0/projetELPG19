// Dans LaunchServer...

    // 1. Charger l'image cible
    targetImg := LoadImage("PAIN.png")
    avgColor := ComputeAverageColor(targetImg) // Calculer la couleur moyenne

    // 2. Initialiser le meilleur organisme
    // Au lieu de partir de vide, on calcule le score d'une image vide (couleur de fond)
    emptyImg := image.NewRGBA(targetImg.Bounds())
    RenderToBuffer([]Shape{}, emptyImg, avgColor) // Rempli avec la couleur moyenne
    initialScore := DiffEuclidienne(emptyImg, targetImg)

    bestOrganism = Organism{DNA: []Shape{}, Score: initialScore}
    
// ... Le reste est identique, sauf pour la sauvegarde :
// Dans handleClient, quand on sauvegarde :
    if *generation%SaveFrequency == 0 {
        // Il faut recréer une image temporaire car le serveur n'a pas de buffer permanent
        tempImg := image.NewRGBA(image.Rect(0,0,MaxX, MaxY))
        RenderToBuffer(bestOrganism.DNA, tempImg, avgColor) // Utiliser RenderToBuffer
        saveFile, _ := os.Create(OutputFile)
        png.Encode(saveFile, tempImg)
        saveFile.Close()
    }
