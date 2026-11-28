package app

import (
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

func TestReflectionCaptureRejectsChangedInputs(t *testing.T) {
	base := reflectionInput{target: &render.WaterTarget{}, view: render.Identity()}
	capture := reflectionCapture{input: base}
	for _, change := range []struct {
		name string
		edit func(*reflectionInput)
	}{
		{"camera", func(in *reflectionInput) { in.view[12] = 1 }},
		{"projection", func(in *reflectionInput) { in.projection[0] = 1 }},
		{"jitter", func(in *reflectionInput) { in.jitter[0] = .001 }},
		{"resize", func(in *reflectionInput) { in.target = &render.WaterTarget{} }},
		{"terrain streaming", func(in *reflectionInput) { in.revision++ }},
		{"fog", func(in *reflectionInput) { in.fog[0] = 1 }},
		{"settings", func(in *reflectionInput) { in.settings.Shadows = true }},
	} {
		in := base
		change.edit(&in)
		if capture.reuse(in, 1.0/120) {
			t.Errorf("reused a capture after changing %s", change.name)
		}
	}
	for _, dt := range []float64{0, -1, .1} {
		if capture.reuse(base, dt) {
			t.Errorf("reused a capture with dt=%f", dt)
		}
	}
	for i := range 4 {
		if got := capture.reuse(base, 1.0/120); got != (i < 3) {
			t.Fatalf("unexpected reflection reuse at 120 Hz on step %d: %v", i+1, got)
		}
	}
}
