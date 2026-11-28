package ui

import (
	"fmt"
	"math"

	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

const MinimapWidth, MinimapHeight = 196, 197
const MinimapImageSize = 136

// Native 0x4c7860 coordinates, relative to the frame art (drawn at the window
// origin +4): the player arrow is a 32px quad centred on the anchor, and the
// 136px circle starts at MinimapCircleEdge on both axes.
const (
	MinimapAnchorX, MinimapAnchorY = 70, 78
	MinimapCircleEdge              = 12
	minimapPartyRadius             = 60 // 0x6be974
	minimapMarkerRadius            = 55 // 0x6c29bc
)

// Projection follows Phoenix tools/minimap_data.py: raw X grows right, raw
// Z grows north (up). Image and map sizes supply their actual calibration.
func MapPosition(x, z, mapSize float32) (u, v float32, ok bool) {
	if mapSize <= 0 {
		return
	}
	u, v = x/mapSize, 1-z/mapSize
	return u, v, u >= 0 && u <= 1 && v >= 0 && v <= 1
}

func MapWindow(u, v, span float32) UV {
	return MapWindowXY(u, v, span, span)
}

func MapPoint(u, v float32, view UV) (float32, float32, bool) {
	if view.U1 <= view.U0 || view.V1 <= view.V0 {
		return 0, 0, false
	}
	x, y := (u-view.U0)/(view.U1-view.U0), (v-view.V0)/(view.V1-view.V0)
	return x, y, x >= 0 && x <= 1 && y >= 0 && y <= 1
}

// Anchoring is responsive delivery placement; native artwork is unscaled here.
func MinimapOrigin(w int) (int, int) { return max(0, w-210), 14 }

// Native full-map artwork (0x4c0720); Phoenix build_unified_ui_frames.py
// recovers the same 16px edges, 43px top and 24px bottom around 512px imagery.
func FullMapRect(w, h int) (x, y int, scale float32) {
	scale = min(float32(1), float32(max(1, w-20))/544, float32(max(1, h-20))/579)
	return (w - int(544*scale)) / 2, (h - int(579*scale)) / 2, scale
}

type MapAction struct {
	Toggle, Close, World bool
	Zoom                 int
}

type Minimap struct {
	WorldAvailable, WorldReturn                                                   bool
	worldButton, mapButton                                                        Button
	Position, FullPosition                                                        WindowPosition
	Caption, MapName                                                              string
	plate                                                                         *render.Texture
	frame, arrow, party, questStart, questEnd, questMob, top, bottom, left, right *render.Texture
	relics                                                                        [3]*render.Texture
	relicPortal                                                                   *render.Texture
	alarm                                                                         *render.Texture
	// Alarm is the native pulse timer (0x4c6550): 0..2 seconds-halves, 0 when idle.
	Alarm   float32
	buttons [3]Button
	close   Button
}

func NewMinimap(t Textures) (*Minimap, error) {
	ld := &loader{t: t}
	m := &Minimap{frame: ld.get("main_map.png"), arrow: ld.get("arrow.png"), party: ld.get("radar_party.png"),
		questStart: ld.get("radar_queststart.png"), questEnd: ld.get("radar_questend.png"),
		top: ld.get("map_up.png"), bottom: ld.get("map_bottom.png"),
		left: ld.get("map_left.png"), right: ld.get("map_right.png")}
	m.questMob = m.party
	if optional, ok := t.(interface{ OptionalInterface(string) *render.Texture }); ok {
		if tex := optional.OptionalInterface("radar_part.png"); tex != nil {
			m.questMob = tex
		}
	}
	// Original minimap/map initializers 0x4c5bf0/0x4c0720.
	if tex, err := t.Interface("main_map_alarm.png"); err == nil {
		m.alarm = tex
	}
	for i := range m.relics {
		if tex, err := t.Interface(fmt.Sprintf("obelisk_%d.png", i)); err == nil {
			m.relics[i] = tex
		}
	}
	m.relicPortal = ld.get("portal.png")
	m.plate = ld.get("main_servertime.png")
	tex := ld.get("main_map_button.png")
	for i, angle := range []float64{-45, -27, 40} {
		pos := [2]int{int(math.Round(68 + 84*math.Cos(angle*math.Pi/180))), int(math.Round(68 + 84*math.Sin(angle*math.Pi/180)))}
		b := &m.buttons[i]
		*b = Button{Tex: tex, OffX: pos[0], OffY: pos[1], DrawW: 32, DrawH: 32, HitW: 28, HitH: 28}
		for state, col := range []int{0, 1, 2, 2, 3} {
			b.UV[state] = UV{float32(col) / 4, float32(i) / 4, float32(col+1) / 4, float32(i+1) / 4}
		}
	}
	m.close = nativeClose(ld.get("talk1_close_button.png"), 0, 0)
	m.worldButton = nativeClose(ld.get("wm_button_w.png"), 0, 0)
	m.mapButton = nativeClose(ld.get("wm_button_m.png"), 0, 0)
	// 0x4c3990: 32px artwork cells with independent 22px hit boxes.
	for _, b := range []*Button{&m.worldButton, &m.mapButton} {
		b.DrawW, b.DrawH = 32, 32
		for i, state := range []int{0, 1, 2, 2, 0} {
			b.UV[i] = UV{float32(state) / 4, 0, float32(state+1) / 4, 1}
		}
	}
	return m, ld.err
}

func (m *Minimap) Update(in *platform.Input, w, h int, full bool) MapAction {
	var action MapAction
	if full {
		x, y, scale := m.FullRect(w, h)
		m.close.OriginX, m.close.OriginY = x+int(515*scale), y+int(scale)
		m.close.DrawW, m.close.DrawH = int(22*scale), int(22*scale)
		m.close.HitW, m.close.HitH = int(22*scale), int(22*scale)
		action.Close = m.close.Update(in)
		if m.WorldAvailable {
			b := &m.worldButton
			if m.WorldReturn {
				b = &m.mapButton
			}
			b.OriginX, b.OriginY = x+int(490*scale), y+int(scale)
			b.DrawW, b.DrawH, b.HitW, b.HitH = int(32*scale), int(32*scale), int(22*scale), int(22*scale)
			action.World = b.Update(in)
		}
		return action
	}
	x, y := m.Origin(w, h)
	for i := range m.buttons {
		m.buttons[i].OriginX, m.buttons[i].OriginY = x, y
		clicked := m.buttons[i].Update(in)
		// 0x4c6550: the zoom buttons act every frame while they are held.
		if m.buttons[i].state == StatePressed {
			switch i {
			case 0:
				action.Zoom = 1
			case 1:
				action.Zoom = -1
			}
		}
		if clicked && i == 2 {
			action.Toggle = true
		}
	}
	return action
}

type MapMarkerKind uint8

const (
	MapParty MapMarkerKind = iota
	MapQuestStart
	MapQuestEnd
	MapQuestObjective
	MapRelicNeutral // relic owners follow the 0x2101 owner byte
	MapRelicLight
	MapRelicFury
	MapRelicPortal
)

// RelicMarkerKind maps a relic owner (0 neutral, 1 light, 2 fury).
func RelicMarkerKind(owner uint8) (MapMarkerKind, bool) {
	if owner > 2 {
		return 0, false
	}
	return MapRelicNeutral + MapMarkerKind(owner), true
}

type MapMarker struct {
	U, V float32
	Kind MapMarkerKind
}

// drawMarkers draws either the allies (party radar dots) or every other mark;
// the allies go over the player's arrow, everything else under it.
func (m *Minimap) drawMarkers(s *render.Sprites, markers []MapMarker, view UV, x, y, side, size float32, circular, allies bool) {
	base := size
	for _, mark := range markers {
		if (mark.Kind == MapParty) != allies {
			continue
		}
		size := base
		texture := m.party
		switch mark.Kind {
		case MapQuestStart:
			texture = m.questStart
		case MapQuestEnd:
			texture = m.questEnd
		case MapQuestObjective:
			texture = m.questMob
		case MapRelicPortal:
			texture, size = m.relicPortal, 32
		case MapRelicNeutral, MapRelicLight, MapRelicFury:
			// Native authored relic and portal cells are 32 pixels.
			texture, size = m.relics[mark.Kind-MapRelicNeutral], 32
			if texture == nil {
				continue
			}
		}
		px, py, ok := MapPoint(mark.U, mark.V, view)
		if !ok || px*side < size/2 || py*side < size/2 || (1-px)*side < size/2 || (1-py)*side < size/2 {
			continue
		}
		dx, dy := px-.5, py-.5
		if circular && dx*dx+dy*dy > (.5-size/side/2)*(.5-size/side/2) {
			continue
		}
		s.Draw(texture, x+px*side-size/2, y+py*side-size/2, size, size, 0, 0, 1, 1, white)
	}
}

// drawSmallMarkers follows the native radar loops of 0x4c7860: a mark is placed
// by its pixel offset from the player, shown only within 55px (60px for party
// dots) of him, and its unscaled artwork is anchored 16px up and left of the
// spot (so 32px cells centre on it and 16px dots sit 8px off).
func (m *Minimap) drawSmallMarkers(s *render.Sprites, markers []MapMarker, view UV, x, y float32, allies bool) {
	if view.U1 <= view.U0 || view.V1 <= view.V0 {
		return
	}
	for _, mark := range markers {
		if (mark.Kind == MapParty) != allies {
			continue
		}
		texture, size, radius := m.party, float32(16), float32(minimapPartyRadius)
		if !allies {
			radius = minimapMarkerRadius
		}
		switch mark.Kind {
		case MapQuestStart:
			texture = m.questStart
		case MapQuestEnd:
			texture = m.questEnd
		case MapQuestObjective:
			texture = m.questMob
		case MapRelicPortal:
			texture, size = m.relicPortal, 32
		case MapRelicNeutral, MapRelicLight, MapRelicFury:
			texture, size = m.relics[mark.Kind-MapRelicNeutral], 32
		}
		if texture == nil {
			continue
		}
		px := (mark.U - view.U0) / (view.U1 - view.U0) * MinimapImageSize
		py := (mark.V - view.V0) / (view.V1 - view.V0) * MinimapImageSize
		dx, dy := px-(MinimapAnchorX-MinimapCircleEdge), py-(MinimapAnchorY-MinimapCircleEdge)
		if float32(math.Hypot(float64(dx), float64(dy))) >= radius {
			continue
		}
		s.Draw(texture, x+px-16, y+py-16, size, size, 0, 0, 1, 1, white)
	}
}

func (m *Minimap) Draw(s *render.Sprites, f *Font, image, solid *render.Texture, w, h int, view UV, u, v, heading float32, mapID uint16, rawX, rawZ float32, markers []MapMarker) {
	x, y := m.Origin(w, h)
	// The backing is only visible when no map image is loaded.
	s.DrawCircle(solid, float32(x+16), float32(y+16), MinimapImageSize, 0, 0, 1, 1, [4]float32{.03, .04, .05, 1})
	if image != nil {
		// 0x4c58c0 draws one 256px quad whose UVs may leave [0,1]. The map texture
		// wraps (see loadMapImage), so past an edge the opposite side shows.
		s.DrawCircle(image, float32(x+16), float32(y+16), MinimapImageSize, view.U0, view.V0, view.U1, view.V1, white)
		m.drawSmallMarkers(s, markers, view, float32(x+16), float32(y+16), false)
		// 0x4c7860: the 32px arrow always sits on the anchor; the map moves.
		s.DrawRotated(m.arrow, float32(x+4+MinimapAnchorX), float32(y+4+MinimapAnchorY), 32, 32, heading, 0, 0, 1, 1, white)
		m.drawSmallMarkers(s, markers, view, float32(x+16), float32(y+16), true)
	}
	s.DrawImage(m.frame, float32(x+4), float32(y+4))
	s.Draw(m.plate, float32(x+9), float32(y+157), 150, 40, 0, 0, 150.0/256, 40.0/64, white)
	for i := range m.buttons {
		m.buttons[i].OriginX, m.buttons[i].OriginY = x, y
		m.buttons[i].Draw(s)
	}
	if m.alarm != nil && m.Alarm > 0 {
		// 0x4c8462: triangle pulse, alpha = 255 * (t > 1 ? |t - 2| : t).
		t := m.Alarm
		if t > 1 {
			t = float32(math.Abs(float64(t - 2)))
		}
		s.Draw(m.alarm, float32(x+4), float32(y+4), 256, 256, 0, 0, 1, 1, [4]float32{1, 1, 1, t})
	}
	label := m.Caption
	if label == "" {
		label = "--:--"
	}
	// 0x4c84fb..0x4c85b8: the clock is plain left-aligned text at the map
	// origin +(46,167), unscaled and without a shadow. The map frame is drawn
	// at the same origin, which this window places at (x+4, y+4).
	// Vertically it sits in the middle of the visible plate (rows 3..33 of the
	// 40px strip drawn at y+157), on whole pixels.
	f.Draw(s, label, x+4+46, y+157+18-f.Height/2, white, 0)
}

func (m *Minimap) DrawFull(s *render.Sprites, f *Font, image, solid *render.Texture, w, h int, view UV, u, v, heading float32, mapID uint16, markers []MapMarker) {
	x, y, scale := m.FullRect(w, h)
	mapX, mapY, side := float32(x)+16*scale, float32(y)+43*scale, 512*scale
	s.Draw(solid, mapX, mapY, side, side, 0, 0, 1, 1, [4]float32{.03, .04, .05, .98})
	if image != nil {
		s.Draw(image, mapX, mapY, side, side, view.U0, view.V0, view.U1, view.V1, white)
		m.drawMarkers(s, markers, view, mapX, mapY, side, 12*scale, false, false)
		if px, py, ok := MapPoint(u, v, view); ok {
			s.DrawRotated(m.arrow, mapX+px*side, mapY+py*side, 20*scale, 20*scale, heading, 0, 0, 1, 1, white)
		}
		m.drawMarkers(s, markers, view, mapX, mapY, side, 12*scale, false, true)
	} else {
		f.Draw(s, Tr(12371), int(mapX+12), int(mapY+12), white, 0)
	}
	s.Draw(m.top, float32(x), float32(y), 544*scale, 43*scale, 0, 0, 544.0/1024, 43.0/64, white)
	s.Draw(m.bottom, float32(x), mapY+side, 544*scale, 24*scale, 0, 0, 544.0/1024, 24.0/32, white)
	s.Draw(m.left, float32(x), mapY, 16*scale, side, 0, 0, 1, 1, white)
	s.Draw(m.right, mapX+side, mapY, 16*scale, side, 0, 0, 1, 1, white)
	m.close.OriginX, m.close.OriginY = x+int(515*scale), y+int(scale)
	m.close.DrawW, m.close.DrawH = int(22*scale), int(22*scale)

	m.close.Draw(s)
	if m.WorldAvailable {
		b := &m.worldButton
		if m.WorldReturn {
			b = &m.mapButton
		}
		b.OriginX, b.OriginY = x+int(490*scale), y+int(scale)
		b.DrawW, b.DrawH = int(32*scale), int(32*scale)
		b.Draw(s)
	}
	label := m.MapName
	if label == "" {
		label = Trf(12369, mapID)
	}
	f.Draw(s, label, x+int(24*scale), y+int(14*scale), white, max(1, int(450*scale)))
}

func (m *Minimap) Origin(w, h int) (int, int) {
	x, y := MinimapOrigin(w)
	return m.Position.Origin(w, h, MinimapWidth, MinimapHeight, x, y)
}
func (m *Minimap) FullRect(w, h int) (int, int, float32) {
	x, y, scale := FullMapRect(w, h)
	x, y = m.FullPosition.Origin(w, h, int(544*scale), int(579*scale), x, y)
	return x, y, scale
}

func (m *Minimap) DragHandleAt(mx, my, w, h int) bool {
	x, y := m.Origin(w, h)
	for _, b := range m.buttons {
		if mx >= x+b.OffX && mx < x+b.OffX+b.HitW && my >= y+b.OffY && my < y+b.OffY+b.HitH {
			return false
		}
	}
	dx, dy := mx-x-84, my-y-84
	return dx*dx+dy*dy <= 76*76 || mx >= x+9 && mx < x+159 && my >= y+157 && my < y+197 || mx >= x+35 && mx < x+110 && my >= y && my < y+20
}

func (m *Minimap) DrawWorldMarker(s *render.Sprites, x, y, scale, alpha float32) {
	s.Draw(m.party, x-8*scale, y-8*scale, 16*scale, 16*scale, 0, 0, 1, 1, [4]float32{1, 1, 1, alpha})
}

// MapCoordinates inverts the visible map crop, including zoom and panning.
func MapCoordinates(view UV, fx, fy, mapSize float32) (x, z float32, ok bool) {
	if mapSize <= 0 || fx < 0 || fx > 1 || fy < 0 || fy > 1 || view.U1 <= view.U0 || view.V1 <= view.V0 {
		return
	}
	return (view.U0 + fx*(view.U1-view.U0)) * mapSize, (1 - view.V0 - fy*(view.V1-view.V0)) * mapSize, true
}

// CharacterCoordinates always describes the actor, never the map cursor.
func CharacterCoordinates(x, y, z float32) string {
	return fmt.Sprintf("X: %.0f  Y: %.2f  Z: %.0f", x, y, z)
}
