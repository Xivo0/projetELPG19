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
	InputFile        = "target.png" //image de base que l'on veut
	OutputFile       = "evolution.png" //image générée par le serveur
	SaveFrequency    = 1 //tout les n générations, on met à jour evolution.png
	MinRadius        = 3 //rayons min et max des cercles
	MaxRadius        = 40
	TargetComplexity = 1000.0 //estimation du nb de formes que contient l'image de fin
)

const (
	ShapeTypeCircle = 0 //formes géométriques
	ShapeTypeRect   = 1
)

var (
	MaxX int
	MaxY int
)

// --- STRUCTURES ---

type Shape struct {  //Structure pour les formes
	Type   int // 0: Cercle, 1: Rectangle
	X, Y   int
	Radius int
	Color  color.RGBA
}

type Organism struct { //Structure qui contient les formes à appliquer
	DNA   []Shape
	Score float64
}

// NetworkMessage : Ce qui voyage entre le Serveur et le Client
type NetworkMessage struct {
	Organism Organism
}

// --- FONCTIONS MÉTIERS (V2 Optimisée) --- //à expliquer

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

func RenderToBuffer(dna []Shape, img *image.RGBA, bg color.RGBA) { //Lorsque nous gardons notre image, on veut en fait "l'effacer" à chaque fois plutot que d'en reprendre une
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
	if minX < 0 { minX = 0 }//fenetrage pour ne pas dépasser le cadre de l'image target
	if minY < 0 { minY = 0 }
	if maxX > bounds.Max.X { maxX = bounds.Max.X }
	if maxY > bounds.Max.Y { maxY = bounds.Max.Y }

	radiusSq := s.Radius * s.Radius
	srcR, srcG, srcB, alpha := int(s.Color.R), int(s.Color.G), int(s.Color.B), int(s.Color.A)
	invAlpha := 255 - alpha

	for y := minY; y < maxY; y++ {
		lineOffset := y * img.Stride
		dy := y - s.Y //à expliquer
		dy2 := dy * dy

		for x := minX; x < maxX; x++ {
			if s.Type == ShapeTypeCircle {
				dx := x - s.X
				if dx*dx+dy2 > radiusSq { continue }
			}
			// Blending optimisé (idem à expliquer)
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

func DiffEuclidienne(img1, img2 *image.RGBA) float64 { //calcule la diff euclidienne entre deux images
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
//Principe d'une roulette de probabilité, 10% de chance de créer une forme, 5%
	if len(o.DNA) == 0 || roulette < 0.1 {
		o.DNA = append(o.DNA, NewRandomShape(target))
		return
	}
	if roulette < 0.15 {//On prend une forme dans le dna aléatoirement, on ajoute une nouvelle forme (à epxliquer)
		index := rand.Intn(len(o.DNA))
		o.DNA = append(o.DNA[:index], o.DNA[index+1:]...)
		return
	}
	if roulette < 0.20 { // On swappe des formes dans le tableau
		i1, i2 := rand.Intn(len(o.DNA)), rand.Intn(len(o.DNA))
		o.DNA[i1], o.DNA[i2] = o.DNA[i2], o.DNA[i1]
		return
	}

	// Modification d'une forme aléatoire
	index := rand.Intn(len(o.DNA))
	s := &o.DNA[index]
	switch rand.Intn(4) {
	case 0: // Position de la forme
		s.X += rand.Intn(21) - 10
		s.Y += rand.Intn(21) - 10
		if s.X < 0 { s.X = 0 }
		if s.X > MaxX { s.X = MaxX }
		if s.Y < 0 { s.Y = 0 }
		if s.Y > MaxY { s.Y = MaxY }
	case 1: // Taille de la forme
		s.Radius += rand.Intn(11) - 5
		if s.Radius < MinRadius { s.Radius = MinRadius }
		if s.Radius > currentMaxRadius { s.Radius = currentMaxRadius }
	case 2: // Couleur de la forme
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
	case 3: // Type de la forme
		s.Type = 1 - s.Type
	}
}

func NewRandomShape(target *image.RGBA) Shape { //Crée une nouvelle forme de taille aléatoire avec une opacité minimale
	x, y := rand.Intn(MaxX), rand.Intn(MaxY)
	offset := (y * target.Stride) + (x * 4)
	return Shape{
		Type:   rand.Intn(2),//choix random de la forme
		X:      x, Y: y,//attribue la taille x et y
		Radius: rand.Intn(MaxRadius-MinRadius) + MinRadius, //rayon aléatoire
		Color: color.RGBA{
			R: target.Pix[offset],
			G: target.Pix[offset+1],
			B: target.Pix[offset+2],
			A: uint8(rand.Intn(200) + 30),
		},
	}
}

func LoadImage(path string) *image.RGBA { //charge l'image 
	f, err := os.Open(path)
	if err != nil {
		// STOP TOUT ! Affiche l'erreur critique
		fmt.Println("ERREUR CRITIQUE : Impossible de trouver l'image :", path)
		fmt.Println("Vérifie que le fichier 'target.png' est bien dans ce dossier :")
		dir, _ := os.Getwd()
		fmt.Println(dir)
		panic(err) // Arrête le programme brutalement
	}
	defer f.Close()

	src, _, err := image.Decode(f) //decode l'image et vérifie qu'elle respecte le format
	if err != nil {
		fmt.Println("ERREUR : L'image n'est pas un PNG valide !")
		panic(err)
	}

	b := src.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, src, b.Min, draw.Src) //à expliquer
	return rgba
}
