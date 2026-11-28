package ui

import (
	"fmt"

	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

const StatusWidth, StatusHeight = 277, 486

const (
	statusAllocationWidth = 132
	statusAllocationTop   = 131
	statusAttributeTop    = 172
	statusAttributeStep   = 20
	statusAttributeHeight = 15
	statusLeftValue       = 73
	statusRightValue      = 188
	statusValueWidth      = 68
)

// Status uses the original stats art and controls (0x502df0). The delivery
// layout starts at the right edge; numeric allocation replaces the graph.
type Status struct {
	Position             WindowPosition
	frames               [2]*render.Texture
	resourceBars, expBar *render.Texture
	close, expand        Button
	apply, cancel        Button
	steps                [12]Button
	heldStep             int // one-based index of the button that owns the press
	repeatIn             float64
	repeated             bool
	Expanded             bool
	Draft                proto.StatPoints
}

type StatusState struct {
	Title                  string
	Class                  string
	Element                string
	AttackSpeed, MoveSpeed string
	Base                   proto.StatPoints
	Points                 uint16
	Details                proto.Details
	HP, MP, SP             uint32
	Stats                  proto.AdditionalStats
	Known, Busy            bool
}

// Display order STR/REC/INT/WIS/DEX/LUC differs from the allocation wire order.
var statOrder = [...]int{0, 2, 3, 4, 1, 5}

// displayedAttributes returns the main grid in STR/REC/INT/WIS/DEX/LUC order.
// The base allocation values are shown until an additional-stats snapshot arrives.
func (s StatusState) displayedAttributes() [6]uint32 {
	if s.Known {
		d := s.Stats
		return [6]uint32{d.Str, d.Rec, d.Int, d.Wis, d.Dex, d.Luc}
	}
	var base [6]uint32
	for i, index := range statOrder {
		base[i] = uint32(s.Base[index])
	}
	return base
}

func StatusOrigin(w, h int, expanded bool) (int, int) {
	width := StatusWidth
	if expanded {
		width += statusAllocationWidth // Native expanded bounds, 0x502ab0.
	}
	return max(0, w-width), min(114, max(0, h-StatusHeight))
}

func NewStatus(t Textures) (*Status, error) {
	ld := &loader{t: t}
	v := &Status{frames: [2]*render.Texture{ld.get("stats_off.png"), ld.get("stats_on.png")}}
	v.resourceBars, v.expBar = ld.get("main_stats_bar.png"), ld.get("main_bottom_bar.png")
	v.close = nativeClose(ld.get("talk1_close_button.png"), 250, 8)
	v.expand = Button{Tex: ld.get("stats_button.png"), OffX: 0, OffY: 156, DrawW: 32, DrawH: 128, HitW: 27, HitH: 102}
	for i := range v.expand.UV {
		v.expand.UV[i] = UV{0, 0, 1, 1}
	}
	tex := ld.get("stats_on_select_button.png")
	for i := range v.steps {
		b := &v.steps[i]
		// Native hit box 0x16 x 0x16 (0x502df0).
		*b = Button{Tex: tex, OffX: 85 + i%2*26, OffY: 128 + i/2*20, DrawW: 32, DrawH: 32, HitW: 22, HitH: 22}
		for state, col := range []int{0, 1, 2, 2, 3} {
			b.UV[state] = UV{float32(col) / 4, float32(i%2) / 2, float32(col+1) / 4, float32(i%2+1) / 2}
		}
	}
	tex = ld.get("stats_on_button.png")
	for i, b := range []*Button{&v.apply, &v.cancel} {
		*b = Button{Tex: tex, OffX: 15 + i*65, OffY: 269, DrawW: 64, DrawH: 32, HitW: 46, HitH: 20}
		for state, col := range []int{0, 1, 2, 2, 3} {
			b.UV[state] = UV{float32(col) / 4, float32(i) / 2, float32(col+1) / 4, float32(i+1) / 2}
		}
	}
	return v, ld.err
}

func (v *Status) Clear() {
	v.Draft, v.Expanded = proto.StatPoints{}, false
	v.heldStep = 0
	for _, b := range v.buttons() {
		b.held = false
	}
}

func (v *Status) controls(w, h int, state StatusState) {
	x, y := v.Origin(w, h)
	mainX := x
	if v.Expanded {
		mainX += statusAllocationWidth
	}
	v.close.OriginX, v.close.OriginY = mainX, y
	v.expand.OriginX, v.expand.OriginY = mainX, y
	for i := range v.steps {
		b := &v.steps[i]
		index := statOrder[i/2]
		b.OriginX, b.OriginY = x, y
		b.Disabled = state.Busy
		if i%2 == 0 {
			b.Disabled = b.Disabled || v.Draft.Total() >= uint32(state.Points) || uint32(state.Base[index])+uint32(v.Draft[index]) >= 65535
		} else {
			b.Disabled = b.Disabled || v.Draft[index] == 0
		}
	}
	for _, b := range []*Button{&v.apply, &v.cancel} {
		b.OriginX, b.OriginY = x, y
		b.Disabled = state.Busy
	}
	// Apply never greys out in the original: with nothing allocated it only folds the sheet.
	v.apply.Disabled = v.apply.Disabled || v.Draft.Total() > uint32(state.Points)
}

// FreeAt reports whether a point of the sheet is empty background, where a
// press starts a window drag. Controls keep their clicks, and the expanded
// graph panel (x..x+0x9b, y+0x4e..y+0x12a in 0x5037f0) swallows the press.
func (v *Status) FreeAt(mx, my, w, h int, state StatusState) bool {
	v.controls(w, h, state)
	x, y := v.Origin(w, h)
	if v.Expanded && mx >= x && mx <= x+0x9b && my >= y+0x4e && my <= y+0x12a {
		return false
	}
	active := []*Button{&v.close}
	if v.Expanded {
		active = append(active, &v.apply, &v.cancel)
		for i := range v.steps {
			active = append(active, &v.steps[i])
		}
	} else {
		active = append(active, &v.expand)
	}
	probe := platform.Input{MouseX: mx, MouseY: my}
	for _, b := range active {
		if b.hit(&probe) {
			return false
		}
	}
	return true
}

// Update returns Apply and Close; Draft is never written into confirmed state.
func (v *Status) Update(in *platform.Input, w, h int, state StatusState, dt float64) (bool, bool) {
	v.controls(w, h, state)
	if in.Pressed {
		v.heldStep = 0
		for _, b := range v.buttons() {
			b.held = false
		}
	}
	if v.close.Update(in) {
		return false, true
	}
	// The side tab only opens the sheet (0x5037f0); Apply and Cancel fold it.
	if !v.Expanded {
		if v.expand.Update(in) {
			v.Expanded = true
		}
		return false, false
	}
	// The 22px hit boxes overlap on the 20px row pitch; like 0x5037f0 the last
	// button under the pointer wins.
	for i := len(v.steps) - 1; i >= 0; i-- {
		b := &v.steps[i]
		pointer := *in
		pointer.Pressed = in.Pressed && v.heldStep == 0 && !b.Disabled && b.hit(in)
		if pointer.Pressed {
			// Native repeat buttons (mode 5, 0x527450) act on the press itself and
			// then every 100 ms while held: no initial delay.
			v.heldStep, v.repeatIn = i+1, statRepeatInterval
			v.stepStat(i, state, 1)
		}
		b.Update(&pointer) // pressed/hover art only; the action already ran
		if v.heldStep != i+1 {
			continue
		}
		switch {
		case in.Clicked || !in.MouseDown || !b.hit(in) || b.Disabled:
			v.heldStep, b.held = 0, false
		case !pointer.Pressed:
			v.repeatStatPoint(i, state, dt)
		}
	}
	if v.cancel.Update(in) {
		v.Clear()
		return false, false
	}
	return v.apply.Update(in), false
}

const statRepeatInterval = .1

// stepStat moves one allocation row by n points: even buttons add, odd ones
// take back, both bounded by the free points, the 16-bit cap and the draft.
func (v *Status) stepStat(i int, state StatusState, n int) {
	index := statOrder[i/2]
	if i%2 == 0 {
		available := int(state.Points) - int(v.Draft.Total())
		room := 65535 - int(state.Base[index]) - int(v.Draft[index])
		v.Draft[index] += uint16(max(0, min(n, available, room)))
		return
	}
	v.Draft[index] -= uint16(max(0, min(n, int(v.Draft[index]))))
}

func (v *Status) repeatStatPoint(i int, state StatusState, dt float64) {
	v.repeatIn -= max(0, dt)
	if v.repeatIn > 1e-9 {
		return
	}
	repeats := 1 + int((-v.repeatIn+1e-9)/statRepeatInterval)
	v.repeatIn += float64(repeats) * statRepeatInterval
	v.stepStat(i, state, repeats)
	v.repeated = true
}

func (v *Status) buttons() []*Button {
	b := []*Button{&v.close, &v.expand, &v.apply, &v.cancel}
	for i := range v.steps {
		b = append(b, &v.steps[i])
	}
	return b
}

// The original field insets are 15 pixels tall on both status frames.
func drawStatusValue(s *render.Sprites, f *Font, value string, x, y, width int, tint ...[4]float32) {
	if value == "" {
		return
	}
	const height = 15
	textWidth := float32(f.Width(value))
	textHeight := float32(f.Height)
	scale := min(float32(1), float32(width-4)/max(1, textWidth), height/max(1, textHeight))
	left := x + int((float32(width)-textWidth*scale)/2)
	top := y + int((height-textHeight*scale)/2)
	color := white
	if len(tint) > 0 {
		color = tint[0]
	}
	f.DrawScaled(s, value, float32(left), float32(top), scale, color, 0)
}

func (v *Status) Draw(s *render.Sprites, f *Font, w, h int, state StatusState) {
	v.controls(w, h, state)
	x, y := v.Origin(w, h)
	variant, width, mainX := 0, StatusWidth, x
	if v.Expanded {
		variant, width, mainX = 1, width+statusAllocationWidth, x+statusAllocationWidth
	}
	s.Draw(v.frames[variant], float32(x), float32(y), float32(width), StatusHeight, 0, 0, float32(width)/512, StatusHeight/512.0, white)
	// Cover the unused field baked into both native frames with the panel background.
	modeX := 150 + mainX - x
	s.Draw(v.frames[variant], float32(mainX+150), float32(y+78), 107, 17, float32(modeX)/512, 115.0/512, float32(modeX+107)/512, 120.0/512, white)
	// Phoenix stats_hud.gd: EXP (73,132) 183x15; pools (73,265+20i) 68x15.
	xp := progress(state.Details.Exp, state.Details.PrevExp, state.Details.NextExp)
	s.Draw(v.expBar, float32(mainX+73), float32(y+132), 183*xp, 15, 0, 0, 428.0/512*xp, 6.0/16, white)
	for i, pool := range []Resource{{state.HP, state.Details.MaxHP}, {state.MP, state.Details.MaxMP}, {state.SP, state.Details.MaxSP}} {
		ratio := progress(pool.Current, 0, pool.Max)
		s.Draw(v.resourceBars, float32(mainX+73), float32(y+265+20*i), 68*ratio, 15, 0, float32(i)/4, ratio, float32(i)/4+12.0/64, white)
	}
	// stats_off/stats_on use the same cells, offset 132px in the expanded frame.
	const left, right, columnWidth = statusLeftValue, statusRightValue, statusValueWidth
	text := func(value string, dx, dy, width int) {
		drawStatusValue(s, f, value, mainX+dx, y+dy, width)
	}
	text(state.Class, left, 79, columnWidth)
	text(state.Title, left, 99, 183)
	text(fmt.Sprintf("%d / %d", state.Details.Exp, state.Details.NextExp), left, 132, 183)
	for i, value := range state.displayedAttributes() {
		text(fmt.Sprint(value), left+i/3*(right-left), statusAttributeTop+i%3*statusAttributeStep, columnWidth)
	}
	text(fmt.Sprint(state.Points), right, 232, columnWidth)
	for i, pool := range []Resource{{state.HP, state.Details.MaxHP}, {state.MP, state.Details.MaxMP}, {state.SP, state.Details.MaxSP}} {
		text(fmt.Sprintf("%d/%d", pool.Current, pool.Max), left, 265+i*20, columnWidth)
	}
	text(state.Element, left, 325, columnWidth)
	derived := []string{"-", "-", "-", "-"}
	if state.Known {
		d := state.Stats
		derived = []string{fmt.Sprintf("%d-%d", d.MinAttack, d.MaxAttack), fmt.Sprintf("%d-%d", d.MinMagic, d.MaxMagic), fmt.Sprint(d.Defense), fmt.Sprint(d.Resistance)}
	}
	for i, value := range derived {
		text(value, right, 265+i*20, columnWidth)
	}
	text(state.AttackSpeed, 92, 345, 96)
	text(state.MoveSpeed, 92, 365, 96)
	for i, value := range []uint32{state.Details.Kills, state.Details.Deaths, state.Details.Wins, state.Details.Losses} {
		text(fmt.Sprint(value), left+i%2*(right-left), 405+i/2*20, columnWidth)
	}
	v.close.Draw(s)
	if !v.Expanded {
		v.expand.Draw(s)
	} else {
		drawStatusValue(s, f, fmt.Sprint(uint32(state.Points)-min(uint32(state.Points), v.Draft.Total())), x+78, y+102, 68)
		// 0x503ee0: each row shows base + draft; yellow once points are placed.
		for i, index := range statOrder {
			tint := white
			if v.Draft[index] != 0 {
				tint = [4]float32{1, 1, 0, 1}
			}
			drawStatusValue(s, f, fmt.Sprint(uint32(state.Base[index])+uint32(v.Draft[index])), x+51, y+statusAllocationTop+i*statusAttributeStep, 30, tint)
		}
		for i := range v.steps {
			v.steps[i].Draw(s)
		}
		v.apply.Draw(s)
		v.cancel.Draw(s)
	}
}

// AttributeMessageAt returns the SysMsg for the hovered attribute.
func (v *Status) AttributeMessageAt(mx, my, w, h int) int {
	x, y := v.Origin(w, h)
	mx, my = mx-x, my-y
	// EP5.4 0x507670 checks the allocation rows, then 0x502c40
	// looks up SysMsg 5283–5288 in STR/REC/INT/WIS/DEX/LUC order.
	if v.Expanded {
		if dy := my - statusAllocationTop; mx >= 20 && mx <= 80 && dy >= 0 &&
			dy < len(statOrder)*statusAttributeStep && dy%statusAttributeStep < statusAttributeHeight {
			return 5283 + dy/statusAttributeStep
		}
		mx -= statusAllocationWidth
	}
	dy := my - statusAttributeTop
	if dy < 0 || dy >= 3*statusAttributeStep || dy%statusAttributeStep >= statusAttributeHeight {
		return 0
	}
	row := dy / statusAttributeStep
	switch {
	case mx >= 20 && mx < statusLeftValue+statusValueWidth:
		return 5283 + row
	case mx >= 148 && mx < statusRightValue+statusValueWidth:
		return 5286 + row
	}
	return 0
}

func (v *Status) Origin(w, h int) (int, int) {
	expanded := v != nil && v.Expanded
	x, y := StatusOrigin(w, h, expanded)
	width := StatusWidth
	if expanded {
		width += statusAllocationWidth
	}
	if v == nil {
		return x, y
	}
	return v.Position.Origin(w, h, width, StatusHeight, x, y)
}
