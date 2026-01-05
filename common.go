package main

import (
	"fmt"
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
	OutputFile       = "evolution.png"
	SaveFrequency    = 100
	MinRadius        = 5
	MaxRadius        = 30
	TargetComplexity = 25000.0
)

var (
	MaxX int
	MaxY int
)

// --- STRUCTURES DONNÉES ---

type Shape struct {
	X, Y   int
	Radius int
	Color  color.RGBA
}

type Organism struct {
	DNA   []Shape
	Score float64
	// Image n'est PAS envoyée sur le réseau (trop lourd), on la recrée si besoin
}

// NetworkMessage est l'enveloppe pour échanger via TCP
type NetworkMessage struct {
	Organism Organism
}

// --- FONCTIONS MÉTIER (Copiées/Collées de ta V1) ---

func NewRandomShape(target *image.RGBA) Shape {
	x := rand.Intn(MaxX)
	y := rand.Intn(MaxY)
	offset := (y * target.Stride) + (x * 4)
	return Shape{
		X:      x,
		Y:      y,
		Radius: rand.Intn(MaxRadius-MinRadius) + MinRadius,
		Color: color.RGBA{
			R: target.Pix[offset],
			G: target.Pix[offset+1],
			B: target.Pix[offset+2],
			A: uint8(rand.Intn(200) + 30),
		},
	}
}

func Mutate(o *Organism, target *image.RGBA, progress float64) {
	currentMaxRadius := int(float64(MaxRadius) * (1.0 - progress))
	if currentMaxRadius < MinRadius {
		currentMaxRadius = MinRadius
	}
	roulette := rand.Float64()

	if len(o.DNA) == 0 || roulette < 0.1 {
		o.DNA = append(o.DNA, NewRandomShape(target))
		return
	}
	if roulette < 0.2 {
		indexKill := rand.Intn(len(o.DNA))
		o.DNA = append(o.DNA[:indexKill], o.DNA[indexKill+1:]...)
		return
	}

	indexModif := rand.Intn(len(o.DNA))
	s := &o.DNA[indexModif]

	switch rand.Intn(3) {
	case 0: // Position
		s.X += rand.Intn(21) - 10
		s.Y += rand.Intn(21) - 10
		if s.X < 0 { s.X = 0 }
		if s.Y < 0 { s.Y = 0 }
		if s.X > MaxX { s.X = MaxX }
		if s.Y > MaxY { s.Y = MaxY }
	case 1: // Taille
		s.Radius += rand.Intn(5) - 2
		if s.Radius > currentMaxRadius { s.Radius = currentMaxRadius }
	case 2: // Couleur
		switch rand.Intn(4) {
		case 0: s.Color.R = uint8(rand.Intn(256))
		case 1: s.Color.G = uint8(rand.Intn(256))
		case 2: s.Color.B = uint8(rand.Intn(256))
		case 3: s.Color.A = uint8(rand.Intn(256))
		}
	}
}

func Render(dna []Shape, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fond Noir Opaque
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
	}
	for _, shape := range dna {
		drawCircle(img, shape)
	}
	return img
}

func drawCircle(img *image.RGBA, s Shape) {
	minX, maxX := s.X-s.Radius, s.X+s.Radius
	minY, maxY := s.Y-s.Radius, s.Y+s.Radius
	bounds := img.Bounds()
	if minX < 0 { minX = 0 }
	if minY < 0 { minY = 0 }
	if maxX > bounds.Max.X { maxX = bounds.Max.X }
	if maxY > bounds.Max.Y { maxY = bounds.Max.Y }

	radiusSq := s.Radius * s.Radius
	for y := minY; y < maxY; y++ {
		dy := y - s.Y
		dy2 := dy * dy
		for x := minX; x < maxX; x++ {
			dx := x - s.X
			if dx*dx+dy2 <= radiusSq {
				offset := (y * img.Stride) + (x * 4)
				srcR, srcG, srcB, alpha := int(s.Color.R), int(s.Color.G), int(s.Color.B), int(s.Color.A)
				dstR, dstG, dstB := int(img.Pix[offset]), int(img.Pix[offset+1]), int(img.Pix[offset+2])
				
				img.Pix[offset+0] = uint8((srcR*alpha + dstR*(255-alpha)) / 255)
				img.Pix[offset+1] = uint8((srcG*alpha + dstG*(255-alpha)) / 255)
				img.Pix[offset+2] = uint8((srcB*alpha + dstB*(255-alpha)) / 255)
				img.Pix[offset+3] = 255
			}
		}
	}
}

func DiffEuclidienne(img1, img2 *image.RGBA) float64 {
	var totalDiff float64 = 0.0
	for i := 0; i < len(img1.Pix); i += 4 {
		r1, r2 := int(img1.Pix[i]), int(img2.Pix[i])
		g1, g2 := int(img1.Pix[i+1]), int(img2.Pix[i+1])
		b1, b2 := int(img1.Pix[i+2]), int(img2.Pix[i+2])
		
		totalDiff += float64((r1-r2)*(r1-r2) + (g1-g2)*(g1-g2) + (b1-b2)*(b1-b2))
	}
	return totalDiff
}

func LoadImage(path string) *image.RGBA {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("Erreur load:", err)
		return image.NewRGBA(image.Rect(0, 0, 200, 200))
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil { return image.NewRGBA(image.Rect(0, 0, 200, 200)) }
	
	b := src.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, src, b.Min, draw.Src)
	return rgba
}
