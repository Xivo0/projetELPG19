package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	width := 800
	height := 600

	// Équivalent du malloc : on alloue une struct image
	// NewRGBA alloue le tableau de pixels pour toi.
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Couleur du rectangle (Rouge opaque)
	// En C ce serait juste 4 uint8_t. Ici on a une petite struct helper.
	col := color.RGBA{R: 255, G: 0, B: 255, A: 255}

	// Coordonnées du rectangle
	rectX, rectY := 100, 100
	rectW, rectH := 200, 150

	// La double boucle, classique du C
	for y := rectY; y < rectY+rectH; y++ {
		for x := rectX; x < rectX+rectW; x++ {

			// LE COEUR DU SUJET : Accès mémoire direct
			// On calcule la position dans le tableau 1D
			offset := (y * img.Stride) + (x * 4)

			img.Pix[offset+0] = col.R
			img.Pix[offset+1] = col.G
			img.Pix[offset+2] = col.B
			img.Pix[offset+3] = col.A
		}
	}

	// Sauvegarde (Là, Go est plus sympa que C, pas besoin de coder le header PNG à la main)
	f, _ := os.Create("test.png")
	defer f.Close() // "defer" c'est pour dire "n'oublie pas le fclose à la fin de la fonction"
	png.Encode(f, img)
}
