package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"

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

func main() {
	track := "barna"
	if len(os.Args) >= 2 {
		track = os.Args[1]
	}
	jsonFile, err := os.Open(fmt.Sprintf("./track.%s.json", track))
	if err != nil {
		panic(err)
	}
	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	var aiw AIW
	err = json.Unmarshal(bytes, &aiw)
	if err != nil {
		panic(err)
	}
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
	fmt.Println(len(aiw))
	fmt.Printf("X: (%f, %f)\n", minX, maxX)
	fmt.Printf("Y: (%f, %f)\n", minY, maxY)
	fmt.Printf("Z: (%f, %f)\n", minZ, maxZ)
	drawImage(aiwFiltered, math.Abs(minX), maxX, math.Abs(minY), maxY, math.Abs(minZ), maxZ, maxType)
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

func drawImage(aiw AIW, minX, maxX, minY, maxY, minZ, maxZ float64, maxType int) {
	// Initialize the graphic context on an RGBA image
	// dest := image.NewRGBA(image.Rect(0, 0, 1297, 1210.0))
	// dest := image.NewRGBA(image.Rect(0, 0, int(minX+maxX), int(minZ+maxZ)))

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
	// gc.SetStrokeColor(color.RGBA{0xff, 0xff, 0xff, 0xff})

	// Draw a closed shape
	for i := 1; i >= 0; i-- {
		aiwFiltered := AIW{}
		for _, data := range aiw {
			if data.Type == i {
				aiwFiltered = append(aiwFiltered, data)
			}
		}

		drawType(gc, aiwFiltered, minX, maxX, minY, maxY, minZ, maxZ, i, rotate, width, height, dest.Rect)
	}

	// invertY(gc, dest.Rect, 0.05)

	// if rotate {
	// 	gc.Rotate(math.Pi / 2)
	// 	f := width / height
	// 	gc.Translate(0, -f*float64(dest.Rect.Max.Y))
	// }

	// gc.Stroke()
	// Save to file
	draw2dimg.SaveToPngFile("hello.png", dest)
}

func drawType(gc draw2d.GraphicContext, aiw AIW, minX, maxX, minY, maxY, minZ, maxZ float64, t int, rotate bool, width, height float64, rect image.Rectangle) {
	// gc.BeginPath() // Initialize a new path
	gc.Save()
	if t == 0 {
		gc.SetStrokeColor(image.White)
		gc.SetLineWidth(15)
	} else {
		gc.SetStrokeColor(color.RGBA{0xaa, 0xaa, 0xaa, 0xff})
		gc.SetLineWidth(10)
	}
	initX, initZ := 0.0, 0.0
	// size := len(aiw)
	for _, data := range aiw {
		if data.Type != t {
			continue
		}
		x := data.X + minX
		z := data.Z + minZ
		if initX == 0.0 && initZ == 0.0 {
			gc.MoveTo(x, z) // Move to a position to start the new path
			initX, initZ = x, z
		} else {
			gc.LineTo(x, z)
		}
	}
	if t == 0 {
		gc.LineTo(initX, initZ)
	}
	invertY(gc, rect, 0.05)

	if rotate {
		gc.Rotate(math.Pi / 2)
		f := width / height
		gc.Translate(0, -f*float64(rect.Max.Y))
	}

	gc.Stroke()
	gc.Restore()
}
