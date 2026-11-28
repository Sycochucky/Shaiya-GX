package main

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

var bannerSize = fyne.NewSize(240, 135)

const bannerScale = 5.0 / 6.0

// drawBanner renders the built-in dashboard artwork.
func drawBanner() fyne.CanvasObject {
	background := canvas.NewRaster(shaiyaGXBackground)
	background.Resize(bannerSize)

	silver := color.NRGBA{R: 211, G: 221, B: 228, A: 225}
	gold := color.NRGBA{R: 214, G: 169, B: 83, A: 230}
	objects := []fyne.CanvasObject{background}
	objects = append(objects,
		bannerLine(fyne.NewPos(8, 151), fyne.NewPos(120, 20), silver, 2),
		bannerLine(fyne.NewPos(17, 162), fyne.NewPos(133, 27), silver, 1),
		bannerLine(fyne.NewPos(178, 145), fyne.NewPos(280, 23), gold, 2),
		bannerLine(fyne.NewPos(190, 154), fyne.NewPos(288, 39), gold, 1),
	)

	shadow := bannerText("SHAIYA", 31, color.NRGBA{A: 175}, fyne.NewPos(35, 59))
	title := bannerText("SHAIYA", 31, color.NRGBA{R: 226, G: 233, B: 238, A: 255}, fyne.NewPos(33, 57))
	gxShadow := bannerText("GX", 43, color.NRGBA{A: 190}, fyne.NewPos(181, 48))
	gx := bannerText("GX", 43, color.NRGBA{R: 224, G: 180, B: 92, A: 255}, fyne.NewPos(179, 46))
	subtitle := bannerText("GENESIS X", 11, color.NRGBA{R: 174, G: 192, B: 204, A: 235}, fyne.NewPos(106, 113))
	objects = append(objects, shadow, title, gxShadow, gx, subtitle)

	frame := canvas.NewRectangle(color.Transparent)
	frame.StrokeColor = border
	frame.StrokeWidth = 1
	frame.Resize(bannerSize)
	objects = append(objects, frame)

	return container.NewGridWrap(bannerSize, container.NewWithoutLayout(objects...))
}

func bannerText(value string, size float32, fill color.Color, position fyne.Position) *canvas.Text {
	text := canvas.NewText(value, fill)
	text.TextSize = size * bannerScale
	text.TextStyle = fyne.TextStyle{Bold: true}
	text.Move(fyne.NewPos(position.X*bannerScale, position.Y*bannerScale))
	return text
}

func bannerLine(start, end fyne.Position, fill color.Color, width float32) *canvas.Line {
	line := canvas.NewLine(fill)
	line.Position1 = fyne.NewPos(start.X*bannerScale, start.Y*bannerScale)
	line.Position2 = fyne.NewPos(end.X*bannerScale, end.Y*bannerScale)
	line.StrokeWidth = width
	return line
}

func shaiyaGXBackground(width, height int) image.Image {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	result := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			xf := float64(x) / float64(width)
			yf := float64(y) / float64(height)
			glow := math.Max(0, 1-math.Hypot(xf-0.34, (yf-0.43)*1.25)/0.72)
			grain := 4 * math.Sin(float64(x+y)*0.12)
			r := uint8(clamp(8+14*glow+grain, 0, 255))
			g := uint8(clamp(18+30*glow+grain, 0, 255))
			b := uint8(clamp(27+45*glow+grain, 0, 255))
			result.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return result
}

func clamp(value, minimum, maximum float64) float64 {
	return math.Max(minimum, math.Min(maximum, value))
}
