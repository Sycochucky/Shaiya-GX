package ui

import (
	"fmt"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/platform"
)

func TestQuestTrackerNativeLayout(t *testing.T) {
	q := new(QuestTracker)
	q.Anchor(600, 100, 1024)
	if x, y := q.Origin(1024, 768); x != 88 || y != 233 {
		t.Fatalf("journal anchor = %d,%d, want 88,233 (0x4ffb00)", x, y)
	}
	q.Anchor(300, 100, 1024)
	if x, _ := q.Origin(1024, 768); x != 1024-256 {
		t.Fatalf("negative anchor = %d, want screen width - 256", x)
	}
	var entries []QuestTrackerEntry
	for id := uint16(1); id <= 6; id++ {
		entries = append(entries, QuestTrackerEntry{ID: id, Title: fmt.Sprint("Quest ", id)})
	}
	entries[0].Expanded, entries[0].Lines = true, []string{"a", "b"}
	q.SetEntries(entries)
	if len(q.rows) != 8 {
		t.Fatalf("rows = %d", len(q.rows))
	}
	if q.TitleAt(15, 43) != 1 || q.TitleAt(15, 63) != 0 || q.TitleAt(215, 103) != 2 || q.TitleAt(216, 103) != 0 || q.TitleAt(20, 123) != 0 {
		t.Fatal("title hit rows do not follow 15..215 x 43+20n")
	}
	x, y := q.Origin(1024, 768)
	q.Update(&platform.Input{MouseX: x + 50, MouseY: y + 50, Wheel: -1}, 1024, 768)
	if q.offset != 1 {
		t.Fatal("wheel did not scroll")
	}
	if id, _, _ := q.Update(&platform.Input{MouseX: x + 20, MouseY: y + 43, Pressed: true, MouseDown: true}, 1024, 768); id != 0 {
		t.Fatalf("objective row toggled quest %d", id)
	}
	if id, _, _ := q.Update(&platform.Input{MouseX: x + 20, MouseY: y + 83, Pressed: true, MouseDown: true}, 1024, 768); id != 2 {
		t.Fatalf("title press toggled %d, want 2", id)
	}
	q.Reset()
	if q.offset != 0 || len(q.rows) != 0 {
		t.Fatal("reset left stale rows")
	}
}
