package ui

import (
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

// EP5.4 QUESTMINI (vtable 0x6c337c, 0x1878 bytes). Geometry from the
// constructor 0x4dc0f0, setup 0x4dd090, draw 0x4dc9b0 and input 0x4dc540.
const (
	QuestTrackerWidth  = 0xf4 // 244
	QuestTrackerHeight = 0x83 // 131
	trackerTextX       = 15   // 0xf; titles hit 15..0xd7
	trackerTitleHitW   = 0xd7 - trackerTextX
	trackerTitleW      = 0xb4 // title clip width
	trackerTimeX       = trackerTextX + 0xa0
	trackerListY       = 0x2b // 43
	trackerRowHeight   = 0x14 // 20
	trackerVisibleRows = 4    // rows stop once 0x4f is passed
	trackerCloseX      = 218
	trackerCloseY      = 1
	trackerScrollX     = 224
	trackerScrollY     = 38
	trackerScrollH     = 77
)

var (
	trackerTitleInk = rgb8(0x93, 0xc5, 0x21)
	trackerReadyInk = rgb8(200, 200, 0)
	trackerTimeInk  = rgb8(0xf2, 0x9f, 0xa3)
)

func rgb8(r, g, b uint8) [4]float32 {
	return [4]float32{float32(r) / 255, float32(g) / 255, float32(b) / 255, 1}
}

// QuestTrackerEntry is one active quest in journal order. Time is the
// original "%5d%s" minutes column, empty when the quest has no limit.
type QuestTrackerEntry struct {
	ID              uint16
	Title, Time     string
	Lines           []string
	Ready, Expanded bool
}

type trackerRow struct {
	id           uint16 // Only title rows are clickable.
	text, time   string
	title, ready bool
}

// QuestTracker owns layout and pointer state. Active quests belong to the game state.
type QuestTracker struct {
	Position   WindowPosition
	art        *render.Texture
	close      Button
	scroll     panelScroll
	rows       []trackerRow
	offset     int
	defX, defY int
	hasDefault bool
}

func NewQuestTracker(t Textures, art *render.Texture) (*QuestTracker, error) {
	ld := &loader{t: t}
	return &QuestTracker{
		art:    art,
		close:  nativeClose(ld.get("talk1_close_button.png"), trackerCloseX, trackerCloseY),
		scroll: newPanelScroll(ld.get("scrollbar.png"), trackerScrollX, trackerScrollY, trackerScrollH),
	}, ld.err
}

// Anchor sets the opening position used by 0x4ffb00: 512 px left of the
// journal and 133 px below its top, or screen width - 256 when that is negative.
func (q *QuestTracker) Anchor(journalX, journalY, w int) {
	q.defX, q.defY, q.hasDefault = journalX-0x200, journalY+0x85, true
	if q.defX < 0 {
		q.defX = w - 0x100
	}
}

func (q *QuestTracker) Origin(w, h int) (int, int) {
	x, y := q.defX, q.defY
	if !q.hasDefault {
		x, y = w-0x100, 0x85
	}
	return q.Position.Origin(w, h, QuestTrackerWidth, QuestTrackerHeight, x, y)
}

func (q *QuestTracker) SetEntries(entries []QuestTrackerEntry) {
	q.rows = q.rows[:0]
	for _, e := range entries {
		q.rows = append(q.rows, trackerRow{id: e.ID, text: e.Title, time: e.Time, title: true, ready: e.Ready})
		if e.Expanded {
			for _, line := range e.Lines {
				q.rows = append(q.rows, trackerRow{text: line})
			}
		}
	}
	q.offset = min(q.offset, max(0, len(q.rows)-trackerVisibleRows))
}

func (q *QuestTracker) Reset() {
	q.rows = q.rows[:0]
	q.offset = 0
	q.close.held, q.close.state = false, StateNormal
	q.scroll.dragging = false
	for i := range q.scroll.arrows {
		q.scroll.arrows[i].held = false
	}
}

// TitleAt returns the quest whose title row is under a window-relative point.
func (q *QuestTracker) TitleAt(rx, ry int) uint16 {
	if rx < trackerTextX || rx > trackerTextX+trackerTitleHitW || ry < trackerListY {
		return 0
	}
	i := (ry - trackerListY) / trackerRowHeight
	if i >= trackerVisibleRows || (ry-trackerListY)%trackerRowHeight > 0x13 || q.offset+i >= len(q.rows) {
		return 0
	}
	if r := q.rows[q.offset+i]; r.title {
		return r.id
	}
	return 0
}

// Update returns the quest whose expansion toggles on this press (0x4dc540
// acts on the button press) and whether the close button was clicked.
func (q *QuestTracker) Update(in *platform.Input, w, h int) (toggle uint16, close, captured bool) {
	x, y := q.Origin(w, h)
	q.close.OriginX, q.close.OriginY = x, y
	close = q.close.Update(in)
	over := in.MouseX >= x && in.MouseX < x+QuestTrackerWidth && in.MouseY >= y && in.MouseY < y+QuestTrackerHeight
	if over && in.Wheel != 0 {
		if in.Wheel > 0 {
			q.offset--
		} else {
			q.offset++
		}
	}
	captured = q.scroll.dragging
	q.offset = q.scroll.Update(in, x, y, len(q.rows), trackerVisibleRows, q.offset)
	captured = captured || q.scroll.dragging
	if in.Pressed && !close {
		toggle = q.TitleAt(in.MouseX-x, in.MouseY-y)
	}
	return toggle, close, captured
}

func (q *QuestTracker) Draw(s *render.Sprites, f *Font, w, h int) {
	x, y := q.Origin(w, h)
	if q.art != nil {
		// 0x549a70 draws the whole 256x256 quest_mini page; the title is baked in.
		s.Draw(q.art, float32(x), float32(y), 256, 256, 0, 0, 1, 1, white)
	}
	end := min(len(q.rows), q.offset+trackerVisibleRows)
	for i, row := range q.rows[q.offset:end] {
		top := y + trackerListY + i*trackerRowHeight
		if !row.title {
			f.Draw(s, row.text, x+trackerTextX, top, white, QuestTrackerWidth-trackerTextX)
			continue
		}
		ink := trackerTitleInk
		if row.ready {
			ink = trackerReadyInk
		}
		f.Draw(s, row.text, x+trackerTextX, top, ink, trackerTitleW)
		if row.time != "" {
			f.Draw(s, row.time, x+trackerTimeX, top, trackerTimeInk, QuestTrackerWidth-trackerTimeX)
		}
	}
	q.scroll.Draw(s, x, y, len(q.rows), trackerVisibleRows, q.offset)
	q.close.OriginX, q.close.OriginY = x, y
	q.close.Draw(s)
}
