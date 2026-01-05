package main

import (
	"image"
	"image/color"
	"math/rand"
)

// --- CONFIGURATION ---
const (
	OutputFile       = "evolution.png"
	SaveFrequency    = 100
	MinRadius        = 3
	MaxRadius        = 40
	TargetComplexity = 1000.0
)

// Types de formes
const (
	ShapeTypeCircle = 0
	ShapeTypeRect   = 1
)

var (
	MaxX int
	MaxY int
)

// --- NOUVELLE STRUCTURE SHAPE ---
type Shape struct {
	Type   int // 0: Cercle, 1: Rectangle
	X, Y   int
	Radius int
	Color  color.RGBA
}

type Organism struct {
	DNA   []Shape
	Score float64
	// On retire le champ Image ici pour ne pas l'envoyer sur le réseau (trop lourd)
}

// Enveloppe réseau
type NetworkMessage struct {
	Organism Organism
}

// --- FONCTIONS OPTIMISÉES ---

// RenderToBuffer remplace Render. Il dessine sur une image existante.
func RenderToBuffer(dna []Shape, img *image.RGBA, bg color.RGBA) {
	// 1. Reset du fond (Rapide)
	bgR, bgG, bgB := bg.R, bg.G, bg.B
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = bgR
		img.Pix[i+1] = bgG
		img.Pix[i+2] = bgB
		img.Pix[i+3] = 255
	}
	// 2. Dessin
	for _, shape := range dna {
		drawShape(img, shape)
	}
}

// La fonction drawShape de ta V2 (Copiée telle quelle)
func drawShape(img *image.RGBA, s Shape) {
    // ... Copie ici tout le contenu de ta fonction drawShape de la V2 ...
    // ... C'est parfait tel quel ...
    // (Je ne remets pas le code pour gagner de la place, mais utilise ta V2)
}

// La fonction DiffEuclidienne de ta V2
func DiffEuclidienne(img1, img2 *image.RGBA) float64 {
    // ... Copie ta fonction V2 ici ...
    var totalDiff float64 = 0.0
	for i := 0; i < len(img1.Pix); i += 4 {
		d1 := int(img1.Pix[i]) - int(img2.Pix[i])
		d2 := int(img1.Pix[i+1]) - int(img2.Pix[i+1])
		d3 := int(img1.Pix[i+2]) - int(img2.Pix[i+2])
		totalDiff += float64(d1*d1 + d2*d2 + d3*d3)
	}
	return totalDiff
}

// La fonction Mutate de ta V2
func Mutate(o *Organism, target *image.RGBA, progress float64) {
    // ... Copie ta fonction V2 ici ...
    // N'oublie pas d'utiliser NewRandomShape V2 aussi
}

func NewRandomShape(target *image.RGBA) Shape {
    // ... Copie ta fonction V2 ici ...
     return Shape{
        Type:   rand.Intn(2),
        X:      rand.Intn(MaxX),
        Y:      rand.Intn(MaxY),
        // ... etc ...
    }
}

// Ajouter ComputeAverageColor ici aussi pour que le Serveur l'utilise
func ComputeAverageColor(img *image.RGBA) color.RGBA {
	var r, g, b, count uint64
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			off := img.PixOffset(x, y)
			r += uint64(img.Pix[off+0])
			g += uint64(img.Pix[off+1])
			b += uint64(img.Pix[off+2])
			count++
		}
	}
	return color.RGBA{R: uint8(r / count), G: uint8(g / count), B: uint8(b / count), A: 255}
}
