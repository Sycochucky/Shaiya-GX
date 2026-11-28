package config

import "time"

// DailyWindow is a server-local time range. An end before the start wraps midnight.
type DailyWindow struct {
	StartMinute, EndMinute uint16
}

func (w DailyWindow) Valid() bool {
	return w.StartMinute < 24*60 && w.EndMinute < 24*60 && w.StartMinute != w.EndMinute
}

func (w DailyWindow) Contains(now time.Time) bool {
	minute := uint16(now.Hour()*60 + now.Minute())
	if w.StartMinute < w.EndMinute {
		return minute >= w.StartMinute && minute < w.EndMinute
	}
	return minute >= w.StartMinute || minute < w.EndMinute
}

func (w DailyWindow) Next(ready time.Time) time.Time {
	if w.Contains(ready) {
		return ready
	}
	year, month, day := ready.Date()
	open := time.Date(year, month, day, int(w.StartMinute)/60, int(w.StartMinute)%60, 0, 0, ready.Location())
	if !open.After(ready) {
		open = open.AddDate(0, 0, 1)
	}
	return open
}
