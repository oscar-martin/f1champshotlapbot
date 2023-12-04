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
	"github.com/llgcode/draw2d/draw2dkit"
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
	carJsonFile, err := os.Open(fmt.Sprintf("./car.%s.json", track))
	carEnabled := true
	if err != nil {
		// panic(err)
		carEnabled = false
	}
	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	var trackAiw AIW
	err = json.Unmarshal(bytes, &trackAiw)
	if err != nil {
		panic(err)
	}

	var carAiw AIW
	if carEnabled {
		err := json.Unmarshal(bytes, &carAiw)
		if err != nil {
			panic(err)
		}
		bytes, err := io.ReadAll(carJsonFile)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(bytes, &carAiw)
		if err != nil {
			panic(err)
		}
	}

	maxX := 0.0
	maxY := 0.0
	maxZ := 0.0
	minX := math.Inf(1)
	minY := math.Inf(1)
	minZ := math.Inf(1)
	minType := 10000
	maxType := 0
	for _, data := range trackAiw {
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
	}
	fmt.Println(len(trackAiw))
	fmt.Printf("X: (%f, %f)\n", minX, maxX)
	fmt.Printf("Y: (%f, %f)\n", minY, maxY)
	fmt.Printf("Z: (%f, %f)\n", minZ, maxZ)
	drawImage(trackAiw, carAiw, math.Abs(minX), maxX, math.Abs(minY), maxY, math.Abs(minZ), maxZ, maxType)
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

func drawImage(trackAiw, carAiw AIW, minX, maxX, minY, maxY, minZ, maxZ float64, maxType int) {
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

	gc.SetFillColor(image.White)
	draw2dkit.RoundedRectangle(gc, 0, 0, width, height, 0, 0)
	gc.FillStroke()

	// Set some properties
	// gc.SetFillColor(color.RGBA{0x44, 0xff, 0x44, 0xff})
	// gc.SetStrokeColor(color.RGBA{0xff, 0xff, 0xff, 0xff})

	// Draw a closed shape
	// for i := 100; i >= 2; i-- {
	// 	aiwFiltered := AIW{}
	// 	for _, data := range aiw {
	// 		if data.Type == i {
	// 			aiwFiltered = append(aiwFiltered, data)
	// 		}
	// 	}

	// 	drawType(gc, aiwFiltered, minX, maxX, minY, maxY, minZ, maxZ, i, rotate, width, height, dest.Rect)
	// }

	for i := maxType; i >= 100; i-- {
		aiwFiltered := AIW{}
		for _, data := range trackAiw {
			if data.Type == i {
				aiwFiltered = append(aiwFiltered, data)
			}
		}

		drawType(gc, aiwFiltered, minX, maxX, minY, maxY, minZ, maxZ, i, rotate, width, height, dest.Rect)
	}

	for i := 1; i >= 0; i-- {
		aiwFiltered := AIW{}
		for _, data := range trackAiw {
			if data.Type == i {
				aiwFiltered = append(aiwFiltered, data)
			}
		}

		drawType(gc, aiwFiltered, minX, maxX, minY, maxY, minZ, maxZ, i, rotate, width, height, dest.Rect)
	}

	drawType(gc, carAiw, minX, maxX, minY, maxY, minZ, maxZ, -1, rotate, width, height, dest.Rect)

	// Save to file
	draw2dimg.SaveToPngFile("hello.png", dest)
}

func drawType(gc draw2d.GraphicContext, aiw AIW, minX, maxX, minY, maxY, minZ, maxZ float64, t int, rotate bool, width, height float64, rect image.Rectangle) {
	gc.Save()

	if t == 0 {
		gc.SetStrokeColor(image.Black)
		gc.SetLineWidth(20)
	} else if t == -1 {
		gc.SetStrokeColor(image.White)
		gc.SetLineWidth(3)
	} else {
		gc.SetStrokeColor(color.RGBA{0x88, 0x88, 0x88, 0xff})
		gc.SetLineWidth(12)
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
