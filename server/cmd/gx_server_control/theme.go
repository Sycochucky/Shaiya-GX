package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
)

type controlTheme struct{ fyne.Theme }

var (
	base         = rgb(19, 23, 27)
	surface      = rgb(27, 33, 39)
	border       = rgb(48, 57, 67)
	ink          = rgb(233, 238, 242)
	quiet        = rgb(169, 181, 192)
	teal         = rgb(114, 220, 202)
	healthy      = rgb(117, 226, 148)
	warning      = rgb(255, 200, 113)
	fault        = rgb(255, 144, 157)
	faultSurface = rgb(57, 38, 46)
)

func rgb(r, g, b uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: 255} }

var colors = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:          base,
	theme.ColorNameOverlayBackground:   surface,
	theme.ColorNameMenuBackground:      surface,
	theme.ColorNameHeaderBackground:    surface,
	theme.ColorNameButton:              rgb(40, 49, 57),
	theme.ColorNameInputBackground:     base,
	theme.ColorNameSeparator:           border,
	theme.ColorNameForeground:          ink,
	theme.ColorNamePrimary:             teal,
	theme.ColorNameForegroundOnPrimary: base,
	theme.ColorNameFocus:               teal,
	theme.ColorNameSelection:           rgb(38, 70, 69),
	theme.ColorNameHover:               rgb(53, 66, 75),
	theme.ColorNameError:               fault,
}

func (t controlTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if value, ok := colors[name]; ok {
		return value
	}
	return t.Theme.Color(name, theme.VariantDark)
}

func statusColor(state control.State) color.Color {
	switch state {
	case control.Ready:
		return healthy
	case control.Starting, control.Stopping:
		return warning
	case control.Failed:
		return fault
	default:
		return quiet
	}
}

func panel(body fyne.CanvasObject) fyne.CanvasObject {
	background := canvas.NewRectangle(surface)
	background.CornerRadius = 6
	background.StrokeColor = border
	background.StrokeWidth = 1
	return container.NewStack(background, container.NewPadded(body))
}
