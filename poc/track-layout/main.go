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
	"github.com/llgcode/draw2d/draw2dkit"
	"github.com/llgcode/draw2d/draw2dsvg"
)

const (
	scaleFactor = 0.5
	margin      = 60
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
	jsonFile, _ := os.Open(fmt.Sprintf("./track.%s.json", track))
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
		bytes, err := io.ReadAll(carJsonFile)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(bytes, &carAiw)
		if err != nil {
			panic(err)
		}
	}

	maxX := math.Inf(-1)
	maxY := math.Inf(-1)
	maxZ := math.Inf(-1)
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

	offsetX := minX - margin
	fMinX := minX - offsetX + (margin / 2)
	fMaxX := maxX - offsetX + (margin / 2)

	offsetZ := minZ - margin
	fMinZ := minZ - offsetZ + (margin / 2)
	fMaxZ := maxZ - offsetZ + (margin / 2)

	fmt.Printf("Fixed X: (%f, %f, off: %f)\n", fMinX, fMaxX, -offsetX)
	fmt.Printf("Fixed Z: (%f, %f, off: %f)\n", fMinZ, fMaxZ, -offsetZ)

	drawImage(trackAiw, carAiw, fMinX, fMaxX, -offsetX, fMinZ, fMaxZ, -offsetZ, maxType)
}

func drawImage(trackAiw, carAiw AIW, minX, maxX, offsetX, minZ, maxZ, offsetZ float64, maxType int) {
	// Initialize the graphic context on an RGBA image
	// dest := image.NewRGBA(image.Rect(0, 0, 1297, 1210.0))
	// dest := image.NewRGBA(image.Rect(0, 0, int(minX+maxX), int(minZ+maxZ)))

	maxX = maxX * (1.0 - scaleFactor)
	maxZ = maxZ * (1.0 - scaleFactor)
	minX = minX * (1.0 - scaleFactor)
	minZ = minZ * (1.0 - scaleFactor)
	offsetX = offsetX * (1.0 - scaleFactor)
	offsetZ = offsetZ * (1.0 - scaleFactor)
	width := maxX
	height := maxZ
	rotate := false
	if width < height {
		// fmt.Println("Rotating")
		rotate = true
		height = maxX
		width = maxZ
	}

	// height = height
	// width = width

	rect := image.Rect(0, 0, int(width), int(height))

	fmt.Printf("Width: %f\nHeight: %f\n", width, height)

	// dest := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	// gc := draw2dimg.NewGraphicContext(dest)
	dest := draw2dsvg.NewSvg()
	dest.Width = fmt.Sprintf("%d", int(width))
	dest.Height = fmt.Sprintf("%d", int(height))
	gc := draw2dsvg.NewGraphicContext(dest)

	// gc.SetFillColor(image.White)
	// draw2dkit.RoundedRectangle(gc, 0, 0, width, height, 0, 0)
	// gc.FillStroke()

	// Set some properties
	// gc.SetFillColor(color.RGBA{0x44, 0xff, 0x44, 0xff})
	// gc.SetStrokeColor(color.RGBA{0xff, 0xff, 0xff, 0xff})

	// for i := maxType; i >= 100; i-- {
	// 	aiwFiltered := AIW{}
	// 	for _, data := range trackAiw {
	// 		if data.Type == i {
	// 			aiwFiltered = append(aiwFiltered, data)
	// 		}
	// 	}

	// 	drawType(gc, aiwFiltered, minX, maxX, offsetX, minZ, maxZ, offsetZ, i, rotate, width, height, rect, scaleFactor)
	// }

	// for i := 99; i >= 3; i-- {
	// 	aiwFiltered := AIW{}
	// 	for _, data := range trackAiw {
	// 		if data.Type == i {
	// 			aiwFiltered = append(aiwFiltered, data)
	// 		}
	// 	}

	// 	drawType(gc, aiwFiltered, minX, maxX, offsetX, minZ, maxZ, offsetZ, i, rotate, width, height, rect, scaleFactor)
	// }

	// for i := 160; i >= 0; i-- {
	for i := 0; i <= 0; i++ {
		aiwFiltered := AIW{}
		for _, data := range trackAiw {
			if data.Type == i {
				aiwFiltered = append(aiwFiltered, data)
			}
		}
		drawType(gc, aiwFiltered, minX, maxX, offsetX, minZ, maxZ, offsetZ, i, rotate, width, height, rect, scaleFactor)
	}

	// drawType(gc, carAiw, minX, maxX, minY, maxY, minZ, maxZ, -1, rotate, width, height, rect, scaleFactor)

	// Save to file
	// draw2dimg.SaveToPngFile("hello.png", dest)
	draw2dsvg.SaveToSvgFile("hello.svg", dest)
}

func drawType(gc draw2d.GraphicContext, aiw AIW, minX, maxX, offsetX, minZ, maxZ, offsetZ float64, t int, rotate bool, width, height float64, rect image.Rectangle, factor float64) {
	gc.Save()

	if t == 0 {
		gc.SetStrokeColor(image.Black)
		gc.SetLineWidth(20 * (1.0 - scaleFactor))
	} else if t == -1 {
		gc.SetStrokeColor(image.White)
		gc.SetLineWidth(3)
	} else if t < 2 {
		gc.SetStrokeColor(color.RGBA{0x88, 0x88, 0x88, 0xff})
		gc.SetLineWidth(12 * (1.0 - scaleFactor))
	} else {
		gc.SetStrokeColor(color.RGBA{0xFF, 0x00, 0x00, 0xff})
		gc.SetLineWidth(2 * (1.0 - scaleFactor))
	}
	initX, initZ := 0.0, 0.0
	// size := len(aiw)
	if t != 2 && t != 3 {
		for _, data := range aiw {
			if data.Type != t {
				continue
			}
			x := data.X*(1.0-scaleFactor) + offsetX
			z := data.Z*(1.0-scaleFactor) + offsetZ
			if initX == 0.0 && initZ == 0.0 {
				gc.MoveTo(x, z) // Move to a position to start the new path
				initX, initZ = x, z
			} else {
				gc.LineTo(x, z)
			}
		}
	} else {
		var finishLine1, finishLine2 Data
		for _, data := range aiw {
			if data.Type != t && data.Type != t+1 {
				continue
			}
			x := data.X*(1.0-scaleFactor) + offsetX
			z := data.Z*(1.0-scaleFactor) + offsetZ
			var finishLine Data
			finishLine.X = x
			finishLine.Z = z
			if data.Type == t {
				finishLine1 = finishLine
			} else {
				finishLine2 = finishLine
				break
			}
			continue
		}
		draw2dkit.Rectangle(gc, finishLine1.X, finishLine1.Z, finishLine2.X, finishLine2.Z)
	}
	// if t == 0 {
	// 	gc.LineTo(initX, initZ)
	// }
	invertY(gc, rect, factor)

	if rotate {
		gc.Rotate(math.Pi / 2)
		f := width / height
		gc.Translate(0, -f*float64(rect.Max.Y))
	}
	// points := gc.GetPath().Points
	// m := gc.GetMatrixTransform()
	// m.Transform(points)
	// for i := 0; i < len(points); i = i + 2 {
	// 	fmt.Printf(`{"x": %f, "y": %f, "z": %f, "type": %d},`, points[i], 0.0, points[i+1], t)
	// }

	gc.Stroke()
	gc.Restore()
}

// Flips the image around the Y axis.
func invertY(gc draw2d.GraphicContext, rect image.Rectangle, factor float64) {
	height := rect.Max.Y
	gc.Translate(0, float64(height))
	// gc.Scale(1.0-factor, -1.0+factor)
	gc.Scale(1.0, -1.0)

	// x := (float64(rect.Max.X) * (1.0 - factor)) / 16
	// y := (float64(rect.Max.Y) * (1.0 - factor)) / 16
	// x := (float64(rect.Max.X) * factor) / 2
	// y := (float64(rect.Max.Y) * factor) / 2
	// gc.Translate(x, y)
}
