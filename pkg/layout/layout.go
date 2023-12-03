package layout

import (
	"image/color"
	"math"

	"image"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

type Data struct {
	Type int     `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Z    float64 `json:"z"`
}

type AIW []Data

func BuildLayoutPNG(track string, aiw AIW) error {
	maxX := 0.0
	maxY := 0.0
	maxZ := 0.0
	minX := math.Inf(1)
	minY := math.Inf(1)
	minZ := math.Inf(1)
	minType := 10000
	maxType := 0
	aiwFiltered := AIW{}
	for _, data := range aiw {
		if data.Type != 0 {
			continue
		}
		if data.X > maxX {
			maxX = data.X
		}
		if data.Y > maxY {
			maxY = data.Y
		}
		if data.Z > maxZ {
			maxZ = data.Z
		}
		if data.Type > maxType {
			maxType = data.Type
		}
		if data.X < minX {
			minX = data.X
		}
		if data.Y < minY {
			minY = data.Y
		}
		if data.Z < minZ {
			minZ = data.Z
		}
		if data.Type < minType {
			minType = data.Type
		}
		aiwFiltered = append(aiwFiltered, data)
	}
	// fmt.Println(len(aiw))
	// fmt.Printf("X: (%f, %f)\n", minX, maxX)
	// fmt.Printf("Y: (%f, %f)\n", minY, maxY)
	// fmt.Printf("Z: (%f, %f)\n", minZ, maxZ)
	return drawImage(track, aiwFiltered, math.Abs(minX), maxX, math.Abs(minY), maxY, math.Abs(minZ), maxZ)
}

// Flips the image around the Y axis.
func invertY(gc draw2d.GraphicContext, rect image.Rectangle, factor float64) {
	height := rect.Max.Y
	gc.Translate(0, float64(height))
	gc.Scale(1.0-factor, -1.0+factor)

	x := (float64(rect.Max.X) * factor) / 2
	y := (float64(rect.Max.Y) * factor) / 2
	gc.Translate(x, y)
}

func drawImage(filepath string, aiw AIW, minX, maxX, minY, maxY, minZ, maxZ float64) error {
	// Initialize the graphic context on an RGBA image
	width := minX + maxX
	height := minZ + maxZ
	rotate := false
	if width < height {
		// fmt.Println("Rotating")
		rotate = true
		height = minX + maxX
		width = minZ + maxZ
	}

	dest := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	gc := draw2dimg.NewGraphicContext(dest)

	// Set some properties
	// gc.SetFillColor(color.RGBA{0x44, 0xff, 0x44, 0xff})
	gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
	gc.SetLineWidth(30)

	// Draw a closed shape
	gc.BeginPath() // Initialize a new path
	initX, initZ := 0.0, 0.0
	for _, data := range aiw {
		x := data.X + minX
		z := data.Z + minZ
		if initX == 0.0 {
			gc.MoveTo(x, z) // Move to a position to start the new path
			initX, initZ = x, z
		} else {
			gc.LineTo(x, z)
		}
	}
	gc.LineTo(initX, initZ)
	gc.Close()
	invertY(gc, dest.Rect, 0.1)

	if rotate {
		gc.Rotate(math.Pi / 2)
		f := width / height
		gc.Translate(0, -f*float64(dest.Rect.Max.Y))
	}

	gc.Stroke()

	// Save to file
	return draw2dimg.SaveToPngFile(filepath, dest)
}