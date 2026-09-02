package fractals

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	CanvasWidth  = 800
	CanvasHeight = 600
)

type FractalKind int

const (
	Mandelbrot FractalKind = 1
	Julia      FractalKind = 2
	Leaf       FractalKind = 3
)

func ParseFractalKind(val int) (FractalKind, error) {
	switch val {
	case 1:
		return Mandelbrot, nil
	case 2:
		return Julia, nil
	case 3:
		return Leaf, nil
	default:
		return 0, fmt.Errorf("tipo de fractal inválido: %d", val)
	}
}

type FractalPoint struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Intensity int     `json:"intensity"`
}

type Bounds struct {
	XMin float64 `json:"xMin"`
	XMax float64 `json:"xMax"`
	YMin float64 `json:"yMin"`
	YMax float64 `json:"yMax"`
}

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) GetFractal(kind FractalKind, bounds Bounds, maxIterations int) []FractalPoint {
	switch kind {
	case Mandelbrot:
		return e.GenerateMandelbrot(bounds, maxIterations)
	case Julia:
		return e.GenerateJulia(bounds, maxIterations)
	case Leaf:
		return e.GenerateLeaf()
	default:
		return []FractalPoint{}
	}
}

func encodeIntensity(iter, maxIterations int) int {
	if iter == maxIterations {
		return 0
	}
	return (iter * 255) / maxIterations
}

func (e *Engine) GenerateMandelbrot(bounds Bounds, maxIterations int) []FractalPoint {
	points := make([]FractalPoint, 0, CanvasWidth*CanvasHeight)
	xRange := bounds.XMax - bounds.XMin
	yRange := bounds.YMax - bounds.YMin

	if xRange <= 0 || yRange <= 0 {
		return []FractalPoint{}
	}

	for screenY := 0; screenY < CanvasHeight; screenY++ {
		for screenX := 0; screenX < CanvasWidth; screenX++ {
			cRe := bounds.XMin + (float64(screenX) * xRange / float64(CanvasWidth))
			cIm := bounds.YMin + (float64(screenY) * yRange / float64(CanvasHeight))

			zRe, zIm := 0.0, 0.0
			iter := 0

			for zRe*zRe+zIm*zIm <= 4.0 && iter < maxIterations {
				nextRe := zRe*zRe - zIm*zIm + cRe
				nextIm := 2.0*zRe*zIm + cIm
				zRe, zIm = nextRe, nextIm
				iter++
			}

			points = append(points, FractalPoint{
				X:         float64(screenX),
				Y:         float64(screenY),
				Intensity: encodeIntensity(iter, maxIterations),
			})
		}
	}
	return points
}

func (e *Engine) GenerateJulia(bounds Bounds, maxIterations int) []FractalPoint {
	points := make([]FractalPoint, 0, CanvasWidth*CanvasHeight)
	xRange := bounds.XMax - bounds.XMin
	yRange := bounds.YMax - bounds.YMin

	cRe := -0.400
	cIm := 0.600

	for screenY := 0; screenY < CanvasHeight; screenY++ {
		for screenX := 0; screenX < CanvasWidth; screenX++ {
			zRe := bounds.XMin + (float64(screenX) * xRange / float64(CanvasWidth))
			zIm := bounds.YMin + (float64(screenY) * yRange / float64(CanvasHeight))
			iter := 0

			for zRe*zRe+zIm*zIm <= 4.0 && iter < maxIterations {
				nextRe := zRe*zRe - zIm*zIm + cRe
				nextIm := 2.0*zRe*zIm + cIm
				zRe, zIm = nextRe, nextIm
				iter++
			}

			points = append(points, FractalPoint{
				X:         float64(screenX),
				Y:         float64(screenY),
				Intensity: encodeIntensity(iter, maxIterations),
			})
		}
	}
	return points
}

func (e *Engine) GenerateLeaf() []FractalPoint {
	var pixelGrid [CanvasWidth][CanvasHeight]int
	x, y := 0.0, 0.0
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	totalPoints := 150000

	for i := 0; i < totalPoints; i++ {
		var nextX, nextY float64
		r := rng.Intn(100)

		switch {
		case r < 1:
			nextX = 0.0
			nextY = 0.16 * y
		case r < 86:
			nextX = 0.85*x + 0.04*y
			nextY = -0.04*x + 0.85*y + 1.6
		case r < 93:
			nextX = 0.20*x - 0.26*y
			nextY = 0.23*x + 0.22*y + 1.6
		default:
			nextX = -0.15*x + 0.28*y
			nextY = 0.26*x + 0.24*y + 0.44
		}

		x, y = nextX, nextY

		screenX := int(math.Round((x + 2.182) * float64(CanvasWidth-1) / (2.655 + 2.182)))
		screenY := int(math.Round((9.96 - y) * float64(CanvasHeight-1) / 9.96))

		if screenX >= 0 && screenX < CanvasWidth && screenY >= 0 && screenY < CanvasHeight {
			pixelGrid[screenX][screenY] = 200
		}
	}

	var points []FractalPoint
	for px := 0; px < CanvasWidth; px++ {
		for py := 0; py < CanvasHeight; py++ {
			if pixelGrid[px][py] > 0 {
				points = append(points, FractalPoint{
					X:         float64(px),
					Y:         float64(py),
					Intensity: pixelGrid[px][py],
				})
			}
		}
	}
	return points
}