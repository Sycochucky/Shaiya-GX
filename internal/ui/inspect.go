package ui

import (
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

const InspectWidth, InspectHeight = 256, 287

type Inspect struct {
	Position WindowPosition
	frame    *render.Texture
	close    Button
}

func NewInspect(t Textures, art *render.Texture) (*Inspect, error) {
	ld := &loader{t: t}
	return &Inspect{frame: art, close: nativeClose(ld.get("talk1_close_button.png"), 220, 6)}, ld.err
}

func (v *Inspect) Origin(w, h int) (int, int) {
	return v.Position.Origin(w, h, InspectWidth, InspectHeight, (w-InspectWidth)/2, (h-InspectHeight)/2)
}

func (v *Inspect) Update(in *platform.Input, w, h int) bool {
	v.close.OriginX, v.close.OriginY = v.Origin(w, h)
	return v.close.Update(in) || in.HasKey(platform.KeyEscape)
}

func (v *Inspect) Draw(s *render.Sprites, f *Font, w, h int, name string) {
	x, y := v.Origin(w, h)
	s.Draw(v.frame, float32(x), float32(y), InspectWidth, InspectHeight, 0, 0, 1, InspectHeight/512.0, white)
	f.Draw(s, name, x+max(24, (InspectWidth-f.Width(name))/2), y+32, white, 208)
	v.close.OriginX, v.close.OriginY = x, y
	v.close.Draw(s)
}

// InspectCell centers the native slot grid in the supplied inspect.png frames.
func InspectCell(slot uint8) (x, y, size int) {
	x, y, size, _ = InventoryCell(0, slot)
	return x, y - 1, size
}
