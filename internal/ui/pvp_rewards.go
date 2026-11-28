package ui

import (
	"strings"

	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

const PvPRewardWidth, PvPRewardHeight = 509, 414

// PvPRewardPanel uses native 0x00518ec0/0x00519ba0 combat art and field positions.
type PvPRewardPanel struct {
	Position          WindowPosition
	frame, guide, bar *render.Texture
	close             Button
	claims            [2]Button
	scroll            panelScroll
	offset, rows      int
}

type PvPRewardState struct {
	Name, Description                       string
	Portrait                                *render.Texture
	Reward, Kills, Deaths, DeathLevel, Gold string
	KillStatus, DeathStatus                 string
	KillProgress, DeathProgress             float32
	CanKill, CanDeath, Busy                 bool
}

func NewPvPRewardPanel(t Textures) (*PvPRewardPanel, error) {
	ld := &loader{t: t}
	v := &PvPRewardPanel{frame: ld.get("combat.png"), guide: ld.get("grade3.png"), bar: ld.get("stats_on_graph.png")}
	v.close = nativeClose(ld.get("talk1_close_button.png"), 481, 8)
	tex := ld.get("combat_button.png")
	for i := range v.claims {
		v.claims[i] = Button{Tex: tex, OffX: 335, OffY: 151 + i*194, DrawW: 256, DrawH: 32, HitW: 134, HitH: 27}
		for state, row := range []int{0, 1, 2, 2, 3} {
			v.claims[i].UV[state] = UV{float32(1-i) / 2, float32(row) / 4, float32(2-i) / 2, float32(row+1) / 4}
		}
	}
	v.scroll = newPanelScroll(ld.get("scrollbar.png"), 240, 82, 244)
	return v, ld.err
}

func (v *PvPRewardPanel) Origin(w, h int) (int, int) {
	x, y := MarketOrigin(w, h)
	return v.Position.Origin(w, h, PvPRewardWidth, PvPRewardHeight, x, y)
}

func (v *PvPRewardPanel) controls(w, h int, state PvPRewardState) {
	x, y := v.Origin(w, h)
	v.close.OriginX, v.close.OriginY = x, y
	for i := range v.claims {
		v.claims[i].OriginX, v.claims[i].OriginY = x, y
		v.claims[i].Disabled = state.Busy || (i == 0 && !state.CanKill) || (i == 1 && !state.CanDeath)
	}
}

func (v *PvPRewardPanel) Reset() { v.offset = 0 }

func (v *PvPRewardPanel) Update(in *platform.Input, w, h int, state PvPRewardState) (claim int, close bool) {
	v.controls(w, h, state)
	if v.close.Update(in) || in.HasKey(platform.KeyEscape) {
		return 0, true
	}
	for i := range v.claims {
		if v.claims[i].Update(in) {
			return i + 1, false
		}
	}
	x, y := v.Origin(w, h)
	if in.MouseX >= x+15 && in.MouseX < x+255 && in.MouseY >= y+82 && in.MouseY < y+326 {
		v.offset = max(0, min(max(0, v.rows-18), v.offset-int(in.Wheel)*3))
	}
	v.offset = v.scroll.Update(in, x, y, v.rows, 18, v.offset)
	return 0, false
}

func (v *PvPRewardPanel) Draw(s *render.Sprites, f *Font, w, h int, state PvPRewardState) {
	v.controls(w, h, state)
	x, y := v.Origin(w, h)
	s.Draw(v.frame, float32(x), float32(y), 509, 414, 0, 0, 509.0/512, 414.0/512, white)
	if state.Portrait != nil {
		s.Draw(state.Portrait, float32(x+7), float32(y+7), 64, 64, 0, 0, 1, 1, white)
	}
	f.Draw(s, state.Name, x+65+max(0, (155-f.Width(state.Name))/2), y+35, white, 155)
	// Insert the native rank guide at the SysMsg marker.
	parts := strings.SplitN(state.Description, "{fn0}", 2)
	before, _ := wrapMarkupWidth(f, parts[0], white, true, 226)
	var after [][]chatRun
	if len(parts) > 1 {
		after, _ = wrapMarkupWidth(f, parts[1], white, true, 226)
	}
	guideRows := 0
	if len(parts) > 1 {
		guideRows = 14
	}
	v.rows = len(before) + guideRows + len(after)
	v.offset = min(v.offset, max(0, v.rows-18))
	restore := s.ClipRect(float32(x+15), float32(y+82), 226, 244)
	drawRows := func(rows [][]chatRun, start int) {
		for i, row := range rows {
			for _, run := range row {
				f.Draw(s, run.text, x+15+run.x, y+82+(start+i-v.offset)*13, run.color, 226-run.x)
			}
		}
	}
	drawRows(before, 0)
	if guideRows > 0 {
		s.Draw(v.guide, float32(x+15), float32(y+82+(len(before)-v.offset)*13), 200, 176, 0, 0, 200.0/256, 176.0/256, white)
	}
	drawRows(after, len(before)+guideRows)
	restore()
	v.scroll.Draw(s, x, y, v.rows, 18, v.offset)
	for _, row := range []struct {
		value       float32
		x, y, width int
		v0, v1      float32
	}{{state.KillProgress, 335, 114, 153, .25, .5}, {state.DeathProgress, 288, 289, 200, 0, .25}} {
		ratio := max(float32(0), min(float32(1), row.value))
		// Native quad arguments are u0,u1,v0,v1: both U values are zero.
		s.Draw(v.bar, float32(x+row.x), float32(y+row.y), float32(row.width)*ratio, 16, 0, row.v0, 0, row.v1, white)
	}
	center := func(value string, dx, dy, width int) {
		f.Draw(s, value, x+dx+max(0, (width-f.Width(value))/2), y+dy, white, width)
	}
	center(state.Reward, 335, 91, 153)
	center(state.Kills, 335, 116, 153)
	center(state.Deaths, 335, 243, 153)
	center(state.DeathLevel, 288, 291, 200)
	center(state.Gold, 335, 319, 153)

	v.close.Draw(s)
	for i := range v.claims {
		v.claims[i].Draw(s)
	}
	center(state.KillStatus, 288, 185, 200)
	center(state.DeathStatus, 288, 379, 200)
}
