package layout

import (
	"bytes"
	"encoding/json"
	"image/color"
	"math"
	"os"

	"image"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dsvg"
)

type Data struct {
	Type int     `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Z    float64 `json:"z"`
}

type AIW []Data

const (
	ScaleSVG = 0.4
)

func getTrackSize(aiw AIW) (float64, float64, float64, float64, float64, float64, int, bool, image.Rectangle) {
	maxX := 0.0
	maxZ := 0.0
	minX := math.Inf(1)
	minZ := math.Inf(1)
	minType := 10000
	maxType := 0
	for _, data := range aiw {
		if data.X > maxX {
			maxX = data.X
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
		if data.Z < minZ {
			minZ = data.Z
		}
		if data.Type < minType {
			minType = data.Type
		}
	}

	// minX = math.Abs(minX)
	// minZ = math.Abs(minZ)
	// maxX = math.Abs(maxX)
	// maxZ = math.Abs(maxZ)

	offsetX := minX
	minX = minX - offsetX
	maxX = maxX - offsetX

	offsetZ := minZ
	minZ = minZ - offsetZ
	maxZ = maxZ - offsetZ

	// Initialize the graphic context on an RGBA image
	width := maxX
	height := maxZ
	rotate := false
	if width < height {
		rotate = true
		height = maxX
		width = maxZ
	}

	rect := image.Rect(0, 0, int(width), int(height))
	return minX, maxX, -offsetX, minZ, maxZ, -offsetZ, maxType, rotate, rect
}

func BuildLayoutPNG(track string, aiw AIW) error {
	minX, maxX, offsetX, minZ, maxZ, offsetZ, maxType, rotate, rect := getTrackSize(aiw)
	width := float64(rect.Max.X)
	height := float64(rect.Max.Y)

	dest := image.NewRGBA(rect)
	gc := draw2dimg.NewGraphicContext(dest)

	drawImage(gc, aiw, minX, maxX, offsetX, minZ, maxZ, offsetZ, maxType, rotate, width, height, rect, 0.1)
	return draw2dimg.SaveToPngFile(track, dest)
}

type SvgMetadata struct {
	MinX    float64         `json:"minX"`
	MaxX    float64         `json:"maxX"`
	OffsetX float64         `json:"offsetX"`
	MinZ    float64         `json:"minZ"`
	MaxZ    float64         `json:"maxZ"`
	OffsetZ float64         `json:"offsetZ"`
	Rotate  bool            `json:"rotate"`
	Width   float64         `json:"width"`
	Height  float64         `json:"height"`
	Rect    image.Rectangle `json:"-"`
}

func BuildLayoutSVG(track string, aiw AIW) error {
	minX, maxX, offsetX, minZ, maxZ, offsetZ, _, rotate, rect := getTrackSize(aiw)
	width := float64(rect.Max.X)
	height := float64(rect.Max.Y)

	dest := draw2dsvg.NewSvg()
	gc := draw2dsvg.NewGraphicContext(dest)

	drawImage(gc, aiw, minX, maxX, offsetX, minZ, maxZ, offsetZ, 1, rotate, width, height, rect, ScaleSVG)
	err := draw2dsvg.SaveToSvgFile(track, dest)
	if err != nil {
		return err
	}

	metadata := SvgMetadata{
		MinX:    minX,
		MaxX:    maxX,
		OffsetX: offsetX,
		MinZ:    minZ,
		MaxZ:    maxZ,
		OffsetZ: offsetZ,
		Rotate:  rotate,
		Width:   width,
		Height:  height,
	}

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	err = json.Compact(buffer, jsonBytes)
	if err != nil {
		return err
	}

	// append metadata to svg file as comments in the xml
	f, err := os.OpenFile(track, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, _ = f.Write([]byte("\n<!--\n"))
	_, _ = f.Write(buffer.Bytes())
	_, err = f.Write([]byte("\n-->"))

	return err
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

func drawImage(gc draw2d.GraphicContext, aiw AIW, minX, maxX, offsetX, minZ, maxZ, offsetZ float64, maxType int, rotate bool, width, height float64, rect image.Rectangle, scale float64) {
	// Draw shapes boxes
	for i := maxType; i >= 100; i-- {
		aiwFiltered := AIW{}
		for _, data := range aiw {
			if data.Type == i {
				aiwFiltered = append(aiwFiltered, data)
			}
		}
		drawType(gc, aiwFiltered, minX, maxX, offsetX, minZ, maxZ, offsetZ, i, rotate, width, height, rect, scale)
	}

	// Draw pitlane and main track
	for i := 1; i >= 0; i-- {
		aiwFiltered := AIW{}
		for _, data := range aiw {
			if data.Type == i {
				aiwFiltered = append(aiwFiltered, data)
			}
		}
		drawType(gc, aiwFiltered, minX, maxX, offsetX, minZ, maxZ, offsetZ, i, rotate, width, height, rect, scale)
	}
}

func drawType(gc draw2d.GraphicContext, aiw AIW, minX, maxX, offsetX, minZ, maxZ, offsetZ float64, t int, rotate bool, width, height float64, rect image.Rectangle, scale float64) {
	gc.Save()
	if t == 0 {
		gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
		gc.SetLineWidth(20)
	} else {
		gc.SetStrokeColor(color.RGBA{0x88, 0x88, 0x88, 0xff})
		gc.SetLineWidth(12)
	}
	initX, initZ := 0.0, 0.0

	for _, data := range aiw {
		if data.Type != t {
			continue
		}
		x := data.X + offsetX
		z := data.Z + offsetZ
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
	invertY(gc, rect, scale)

	if rotate {
		gc.Rotate(math.Pi / 2)
		f := width / height
		gc.Translate(0, -f*float64(rect.Max.Y))
	}

	gc.Stroke()
	gc.Restore()
}
