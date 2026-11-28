package control

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const logReadLimit int64 = 256 << 10

const operatorLog Service = "gx_operator"

type LogEntry struct {
	Time                   time.Time
	Level                  string
	Service                Service
	Message                string
	Activity               string
	MapID                  *uint16
	AccountID, CharacterID uint32
	Character              string
}

type logPosition struct {
	file   os.FileInfo
	offset int64
}

// LogReader tracks complete JSON records without holding files open between
// polls, so the Windows services can rotate their logs.
type LogReader struct {
	root      string
	positions map[Service]logPosition
}

func NewLogReader(root string) *LogReader {
	return &LogReader{root: root, positions: make(map[Service]logPosition)}
}

func (r *LogReader) Poll() ([]LogEntry, error) {
	var entries []LogEntry
	var errs []error
	for _, service := range append(Services(), operatorLog) {
		path := filepath.Join(r.root, "logs", string(service)+".log")
		next, records, err := readLog(path, service, r.positions[service])
		if err != nil {
			errs = append(errs, fmt.Errorf("%s logs: %w", service.Label(), err))
			continue
		}
		r.positions[service] = next
		entries = append(entries, records...)
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Time.Before(entries[j].Time) })
	return entries, errors.Join(errs...)
}

func readLog(path string, service Service, pos logPosition) (logPosition, []LogEntry, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return pos, nil, nil
	}
	if err != nil {
		return pos, nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return pos, nil, err
	}
	var entries []LogEntry
	offset := pos.offset
	if pos.file == nil || !os.SameFile(pos.file, info) {
		if previous, err := os.Open(path + ".1"); err == nil {
			if old, err := previous.Stat(); err == nil && !os.SameFile(info, old) &&
				(pos.file == nil || os.SameFile(pos.file, old)) {
				_, entries, _ = readLogRecords(previous, service, old.Size(), offset)
			}
			previous.Close()
		}
		offset = 0
	} else if info.Size() < offset {
		offset = 0
	}
	offset, records, err := readLogRecords(f, service, info.Size(), offset)
	if err != nil {
		return pos, nil, err
	}
	return logPosition{file: info, offset: offset}, append(entries, records...), nil
}

func readLogRecords(f *os.File, service Service, size, offset int64) (int64, []LogEntry, error) {
	if offset > size {
		offset = 0
	}
	start := max(offset, size-logReadLimit)
	if start == size {
		return start, nil, nil
	}
	data := make([]byte, size-start)
	n, err := f.ReadAt(data, start)
	if err != nil && !errors.Is(err, io.EOF) {
		return offset, nil, err
	}
	data = data[:n]
	var entries []LogEntry
	first := true
	for {
		line, rest, complete := bytes.Cut(data, []byte{'\n'})
		if !complete {
			return start, entries, nil
		}
		var wire struct {
			Time        time.Time `json:"time"`
			Level       string    `json:"level"`
			Activity    string    `json:"activity"`
			MapID       *uint16   `json:"map_id"`
			AccountID   uint32    `json:"account_id"`
			CharacterID uint32    `json:"character_id"`
			Character   string    `json:"character"`
			Message     string    `json:"message"`
		}
		if (!first || start == offset) && json.Unmarshal(line, &wire) == nil && wire.Message != "" {
			entries = append(entries, LogEntry{Time: wire.Time, Level: wire.Level, Service: service, Message: wire.Message, Activity: wire.Activity, MapID: wire.MapID, AccountID: wire.AccountID, CharacterID: wire.CharacterID, Character: wire.Character})
		}
		first = false
		start += int64(len(line) + 1)
		data = rest
	}
}
