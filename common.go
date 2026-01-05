package main

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math/rand"
	"os"
)

// --- CONFIGURATION ---
const (
	InputFile        = "target.png" // Assure-toi que l'image s'appelle bien comme ça !
	OutputFile       = "evolution.png"
	SaveFrequency    = 100
	MinRadius        = 3
	MaxRadius        = 40
	TargetComplexity = 1000.0
)

const (
	ShapeTypeCircle = 0
	ShapeTypeRect   = 1
)

var (
	MaxX int
	MaxY int
)

// --- STRUCTURES ---

type Shape struct {
	Type   int // 0: Cercle, 1: Rectangle
	X, Y   int
	Radius int
	Color  color.RGBA
}

type Organism struct {
	DNA   []Shape
	Score float64
}

// NetworkMessage : Ce qui voyage entre le Serveur et le Client
type NetworkMessage struct {
	Organism Organism
}

// --- FONCTIONS MÉTIERS (V2 Optimisée) ---

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

func RenderToBuffer(dna []Shape, img *image.RGBA, bg color.RGBA) {
	// 1. Reset rapide du fond
	bgR, bgG, bgB := bg.R, bg.G, bg.B
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = bgR
		img.Pix[i+1] = bgG
		img.Pix[i+2] = bgB
		img.Pix[i+3] = 255
	}
	// 2. Dessin des formes
	for _, shape := range dna {
		drawShape(img, shape)
	}
}

func drawShape(img *image.RGBA, s Shape) {
	minX, maxX := s.X-s.Radius, s.X+s.Radius
	minY, maxY := s.Y-s.Radius, s.Y+s.Radius
	bounds := img.Bounds()
	if minX < 0 { minX = 0 }
	if minY < 0 { minY = 0 }
	if maxX > bounds.Max.X { maxX = bounds.Max.X }
	if maxY > bounds.Max.Y { maxY = bounds.Max.Y }

	radiusSq := s.Radius * s.Radius
	srcR, srcG, srcB, alpha := int(s.Color.R), int(s.Color.G), int(s.Color.B), int(s.Color.A)
	invAlpha := 255 - alpha

	for y := minY; y < maxY; y++ {
		lineOffset := y * img.Stride
		dy := y - s.Y
		dy2 := dy * dy

		for x := minX; x < maxX; x++ {
			if s.Type == ShapeTypeCircle {
				dx := x - s.X
				if dx*dx+dy2 > radiusSq { continue }
			}
			// Blending optimisé
			offset := lineOffset + (x * 4)
			r := (srcR*alpha + int(img.Pix[offset+0])*invAlpha) / 255
			g := (srcG*alpha + int(img.Pix[offset+1])*invAlpha) / 255
			b := (srcB*alpha + int(img.Pix[offset+2])*invAlpha) / 255

			img.Pix[offset+0] = uint8(r)
			img.Pix[offset+1] = uint8(g)
			img.Pix[offset+2] = uint8(b)
		}
	}
}

func DiffEuclidienne(img1, img2 *image.RGBA) float64 {
	var totalDiff float64 = 0.0
	for i := 0; i < len(img1.Pix); i += 4 {
		d1 := int(img1.Pix[i]) - int(img2.Pix[i])
		d2 := int(img1.Pix[i+1]) - int(img2.Pix[i+1])
		d3 := int(img1.Pix[i+2]) - int(img2.Pix[i+2])
		totalDiff += float64(d1*d1 + d2*d2 + d3*d3)
	}
	return totalDiff
}

func Mutate(o *Organism, target *image.RGBA, progress float64) {
	currentMaxRadius := int(float64(MaxRadius) * (1.1 - progress))
	if currentMaxRadius < MinRadius { currentMaxRadius = MinRadius }
	roulette := rand.Float64()

	if len(o.DNA) == 0 || roulette < 0.1 {
		o.DNA = append(o.DNA, NewRandomShape(target))
		return
	}
	if roulette < 0.15 {
		index := rand.Intn(len(o.DNA))
		o.DNA = append(o.DNA[:index], o.DNA[index+1:]...)
		return
	}
	if roulette < 0.20 { // Z-Index swap
		i1, i2 := rand.Intn(len(o.DNA)), rand.Intn(len(o.DNA))
		o.DNA[i1], o.DNA[i2] = o.DNA[i2], o.DNA[i1]
		return
	}

	// Modification
	index := rand.Intn(len(o.DNA))
	s := &o.DNA[index]
	switch rand.Intn(4) {
	case 0: // Pos
		s.X += rand.Intn(21) - 10
		s.Y += rand.Intn(21) - 10
		if s.X < 0 { s.X = 0 }
		if s.X > MaxX { s.X = MaxX }
		if s.Y < 0 { s.Y = 0 }
		if s.Y > MaxY { s.Y = MaxY }
	case 1: // Taille
		s.Radius += rand.Intn(11) - 5
		if s.Radius < MinRadius { s.Radius = MinRadius }
		if s.Radius > currentMaxRadius { s.Radius = currentMaxRadius }
	case 2: // Couleur
		switch rand.Intn(4) {
		case 0: s.Color.R = uint8(rand.Intn(256))
		case 1: s.Color.G = uint8(rand.Intn(256))
		case 2: s.Color.B = uint8(rand.Intn(256))
		case 3:
			newA := int(s.Color.A) + rand.Intn(30) - 15
			if newA < 10 { newA = 10 }
			if newA > 255 { newA = 255 }
			s.Color.A = uint8(newA)
		}
	case 3: // Type
		s.Type = 1 - s.Type
	}
}

func NewRandomShape(target *image.RGBA) Shape {
	x, y := rand.Intn(MaxX), rand.Intn(MaxY)
	offset := (y * target.Stride) + (x * 4)
	return Shape{
		Type:   rand.Intn(2),
		X:      x, Y: y,
		Radius: rand.Intn(MaxRadius-MinRadius) + MinRadius,
		Color: color.RGBA{
			R: target.Pix[offset],
			G: target.Pix[offset+1],
			B: target.Pix[offset+2],
			A: uint8(rand.Intn(200) + 30),
		},
	}
}

func LoadImage(path string) *image.RGBA {
	f, err := os.Open(path)
	if err != nil {
		// Image vide par défaut pour éviter le crash si pas de fichier
		return image.NewRGBA(image.Rect(0, 0, 200, 200))
	}
	defer f.Close()
	src, _, _ := image.Decode(f)
	b := src.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, src, b.Min, draw.Src)
	return rgba
}
