package config

import (
	"testing"
	"time"
)

func TestDailyWindowCrossesMidnight(t *testing.T) {
	window := DailyWindow{StartMinute: 22 * 60, EndMinute: 2 * 60}
	loc := time.FixedZone("server", 11*3600)
	afterMidnight := time.Date(2026, 11, 26, 1, 0, 0, 0, loc)
	if !window.Valid() || !window.Contains(afterMidnight) {
		t.Fatal("overnight window rejected an eligible time")
	}
	closeTime := afterMidnight.Add(time.Hour)
	if window.Contains(closeTime) {
		t.Fatal("the end of the window must be exclusive")
	}
	if next := window.Next(closeTime); next.Hour() != 22 || next.Day() != closeTime.Day() {
		t.Fatalf("wrong next window: %s", next)
	}
}
