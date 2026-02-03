package main

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math/rand"
	"os"
)


// CONFIGURATION GLOBALE

const (
	InputFile        = "target.png"    // Doit être présent chez tout le monde
	OutputFile       = "evolution.png" // Créé par le serveur
	MinRadius        = 3
	MaxRadius        = 50
	TargetComplexity = 5000.0 // Estimation du nombres de formes de l'image finale pour appliquer les heuristiques
	ShapeAlphaMin    = 30
	ShapeAlphaMax    = 200
)

var (
	MaxX int // Val max des formes
	MaxY int
)

// STRUCTURES

type Shape struct {
	Type   int // 0: Cercle, 1: Rectangle
	X, Y   int
	Radius int // Sert de "Demi-largeur" pour le carré
	Color  color.RGBA
}

type Organism struct {
	DNA   []Shape
	Score float64
}

// Structure pour le réseau (Gob)
type NetworkMessage struct {
	Organism Organism
}

// FONCTIONS MÉTIERS

func ComputeAverageColor(img *image.RGBA) color.RGBA { //Renvoie la couleur moyenne de l'image pour commencer avec un fond coloré et gagner du temps d'éxecution
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
	return color.RGBA{uint8(r / count), uint8(g / count), uint8(b / count), 255}
}

func RenderToBuffer(dna []Shape, img *image.RGBA, bg color.RGBA) { 
	bgR, bgG, bgB := bg.R, bg.G, bg.B
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = bgR
		img.Pix[i+1] = bgG
		img.Pix[i+2] = bgB
		img.Pix[i+3] = 255
	}
	for _, s := range dna {
		drawShapeOptimized(img, s)
	}
}

func drawShapeOptimized(img *image.RGBA, s Shape) {
	// Calcul de la zone (Bounding Box)
	minX, maxX := s.X-s.Radius, s.X+s.Radius
	minY, maxY := s.Y-s.Radius, s.Y+s.Radius

	bounds := img.Bounds()
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > bounds.Max.X {
		maxX = bounds.Max.X
	}
	if maxY > bounds.Max.Y {
		maxY = bounds.Max.Y
	}

	// Pré-calculs
	radiusSq := s.Radius * s.Radius
	srcR, srcG, srcB, alpha := int(s.Color.R), int(s.Color.G), int(s.Color.B), int(s.Color.A)
	invAlpha := 255 - alpha

	for y := minY; y < maxY; y++ {
		lineOffset := y * img.Stride

		// Optimisation Cercle (calculé seulement si besoin)
		dy := y - s.Y
		dy2 := dy * dy

		for x := minX; x < maxX; x++ {

			// LOGIQUE DE FORME :
			// Si c'est un CERCLE (Type 0), on vérifie la distance.
			// Si c'est un RECTANGLE (Type 1), on dessine tout ce qui est dans la boucle.
			if s.Type == 0 {
				dx := x - s.X
				if dx*dx+dy2 > radiusSq {
					continue // On est dans le coin du carré, mais hors du cercle
				}
			}

			// Blending (Identique pour les deux)
			off := lineOffset + (x * 4)
			r := (srcR*alpha + int(img.Pix[off+0])*invAlpha) / 255
			g := (srcG*alpha + int(img.Pix[off+1])*invAlpha) / 255
			b := (srcB*alpha + int(img.Pix[off+2])*invAlpha) / 255

			img.Pix[off+0] = uint8(r)
			img.Pix[off+1] = uint8(g)
			img.Pix[off+2] = uint8(b)
		}
	}
}

func DiffEuclidienne(img1, img2 *image.RGBA) float64 {
	var total float64
	for i := 0; i < len(img1.Pix); i += 4 {
		d1 := int(img1.Pix[i]) - int(img2.Pix[i])
		d2 := int(img1.Pix[i+1]) - int(img2.Pix[i+1])
		d3 := int(img1.Pix[i+2]) - int(img2.Pix[i+2])
		total += float64(d1*d1*d1*3 + d2*d2*4 + d3*d3*2)
	}
	return total
}

func Mutate(o *Organism, target *image.RGBA) {
	roulette := rand.Float64()
	progress := float64(len(o.DNA)) / TargetComplexity
	if progress > 1.0 {
		progress = 1.0
	}

	currentMaxRadius := int(float64(MaxRadius) * (1.1 - progress))
	if currentMaxRadius < MinRadius {
		currentMaxRadius = MinRadius
	}

	// Dans Mutate, remplace le bloc de création par ça,
	// ou mets à jour ta fonction NewRandomShape si tu l'as extraite :

	if len(o.DNA) == 0 || roulette < 0.3 {
		x, y := rand.Intn(MaxX), rand.Intn(MaxY)
		off := (y * target.Stride) + (x * 4)

		newS := Shape{
			Type: rand.Intn(2), // 0 ou 1 (Cercle ou Rectangle)
			X:    x, Y: y,
			Radius: rand.Intn(MaxRadius-MinRadius) + MinRadius,
			Color:  color.RGBA{target.Pix[off], target.Pix[off+1], target.Pix[off+2], uint8(rand.Intn(ShapeAlphaMax-ShapeAlphaMin) + ShapeAlphaMin)},
		}

		if newS.Radius > currentMaxRadius {
			newS.Radius = currentMaxRadius
		}
		o.DNA = append(o.DNA, newS)
		return
	}
	if roulette < 0.35 && len(o.DNA) > 0 {
		idx := rand.Intn(len(o.DNA))
		o.DNA = append(o.DNA[:idx], o.DNA[idx+1:]...)
		return
	}
	if roulette < 0.40 && len(o.DNA) > 1 {
		i1, i2 := rand.Intn(len(o.DNA)), rand.Intn(len(o.DNA))
		o.DNA[i1], o.DNA[i2] = o.DNA[i2], o.DNA[i1]
		return
	}
	if len(o.DNA) > 0 {
		idx := rand.Intn(len(o.DNA))
		s := &o.DNA[idx]
		switch rand.Intn(4) {
		case 0:
			move := int(20.0 * (1.1 - progress))
			if move < 2 {
				move = 2
			}
			s.X += rand.Intn(move*2) - move
			s.Y += rand.Intn(move*2) - move
			if s.X < 0 {
				s.X = 0
			} else if s.X > MaxX {
				s.X = MaxX
			}
			if s.Y < 0 {
				s.Y = 0
			} else if s.Y > MaxY {
				s.Y = MaxY
			}
		case 1:
			s.Radius += rand.Intn(5) - 2
			if s.Radius < MinRadius {
				s.Radius = MinRadius
			}
			if s.Radius > currentMaxRadius {
				s.Radius = currentMaxRadius
			}
		case 2:
			nA := int(s.Color.A) + rand.Intn(20) - 10
			if nA < ShapeAlphaMin {
				nA = ShapeAlphaMin
			} else if nA > ShapeAlphaMax {
				nA = ShapeAlphaMax
			}
			s.Color.A = uint8(nA)
		case 3: // Changement de Type
			s.Type = 1 - s.Type // Bascule 0->1 ou 1->0
		}

	}
}

func (o Organism) Copy() Organism {
	newOrg := Organism{Score: o.Score, DNA: make([]Shape, len(o.DNA))}
	copy(newOrg.DNA, o.DNA)
	return newOrg
}

func LoadImage(path string) *image.RGBA {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	b := src.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, src, b.Min, draw.Src)
	return rgba
}
