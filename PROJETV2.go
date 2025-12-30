// 1. Nouvelle Structure Shape
type Shape struct {
	Type   int // 0 = Cercle, 1 = Rectangle
	X, Y   int
	Radius int // Sert de "demi-largeur" pour le rectangle
	Color  color.RGBA
}

// 2. Optimisation DrawShape (Supporte Rectangles)
func drawShape(img *image.RGBA, s Shape) {
	minX, maxX := s.X-s.Radius, s.X+s.Radius
	minY, maxY := s.Y-s.Radius, s.Y+s.Radius

	bounds := img.Bounds()
	if minX < 0 { minX = 0 }
	if minY < 0 { minY = 0 }
	if maxX > bounds.Max.X { maxX = bounds.Max.X }
	if maxY > bounds.Max.Y { maxY = bounds.Max.Y }

	radiusSq := s.Radius * s.Radius
	
	// Pré-calcul couleurs pour éviter de le faire dans la boucle
	srcR, srcG, srcB, a := int(s.Color.R), int(s.Color.G), int(s.Color.B), int(s.Color.A)
	invA := 255 - a

	for y := minY; y < maxY; y++ {
		// Optimisation: calculer dy2 une fois par ligne (pour les cercles)
		dy := y - s.Y
		dy2 := dy * dy
		
		// Optimisation: pointeur direct sur le début de la ligne
		lineOffset := y * img.Stride

		for x := minX; x < maxX; x++ {
			// Si Cercle, test de distance
			if s.Type == 0 {
				dx := x - s.X
				if dx*dx+dy2 > radiusSq {
					continue
				}
			}
			// Si Rectangle (Type 1), on dessine tout le carré défini par Radius

			offset := lineOffset + (x * 4)
			
			// Mélange rapide sans division flottante
			r := (srcR*a + int(img.Pix[offset+0])*invA) / 255
			g := (srcG*a + int(img.Pix[offset+1])*invA) / 255
			b := (srcB*a + int(img.Pix[offset+2])*invA) / 255

			img.Pix[offset+0] = uint8(r)
			img.Pix[offset+1] = uint8(g)
			img.Pix[offset+2] = uint8(b)
			img.Pix[offset+3] = 255
		}
	}
}

// 3. Render optimisé (Ne crée pas d'image, remplit un buffer existant)
func RenderTo(dna []Shape, img *image.RGBA, bgColor color.RGBA) {
	// Reset de l'image avec la couleur de fond
	// Une boucle simple est très rapide en Go (optimisée par le compilo en memset/memclr)
	bgR, bgG, bgB := bgColor.R, bgColor.G, bgColor.B
	
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = bgR
		img.Pix[i+1] = bgG
		img.Pix[i+2] = bgB
		img.Pix[i+3] = 255
	}

	for _, shape := range dna {
		drawShape(img, shape)
	}
}

// 4. Worker modifié pour la gestion mémoire
func Worker(id int, jobs <-chan Job, results chan<- Result, targetImg *image.RGBA) {
	// Buffer réutilisable unique par worker !
	// Évite de créer et détruire 100 images par seconde.
	myBuffer := image.NewRGBA(targetImg.Bounds())
	
	// Couleur de fond initiale (Moyenne ou Noir)
	bgColor := color.RGBA{0, 0, 0, 255} 

	for job := range jobs {
		candidate := job.BestOrganism.Copy()
		
		// ... Logique de mutation (penser à update NewRandomShape pour gérer le Type 0 ou 1) ...
        // Note: Dans Mutate, ajoute un case pour changer le TYPE de forme aussi.
        // case 3: // CHANGER FORME
        //      s.Type = 1 - s.Type // Bascule 0 <-> 1

		nbFormes := len(candidate.DNA)
		progress := float64(nbFormes) / TargetComplexity
		if progress > 1.0 { progress = 1.0 }
		
		Mutate(&candidate, targetImg, progress)

		// Rendu dans notre buffer persistant
		RenderTo(candidate.DNA, myBuffer, bgColor)
		
		// Pour le retour, on doit cloner l'image car myBuffer sera écrasé au prochain tour
		// MAIS: On ne clone que si c'est mieux (optimisation différée)
		// Astuce: On calcule le score sur myBuffer
		candidate.Score = DiffEuclidienne(myBuffer, targetImg)

		isBetter := candidate.Score < job.BestOrganism.Score
		
		if isBetter {
			// SEULEMENT si c'est mieux, on sauvegarde l'image pour l'envoyer au Main
			savedImg := image.NewRGBA(myBuffer.Bounds())
			copy(savedImg.Pix, myBuffer.Pix)
			candidate.Image = savedImg
		} else {
			candidate.Image = nil // Pas besoin d'envoyer l'image si c'est un échec
		}

		results <- Result{Organism: candidate, IsBetter: isBetter}
	}
}
