// Package logx writes structured JSON logs with bounded disk retention.
package logx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	mu              sync.Mutex
	file            *os.File
	name, path      string
	debug           bool
	quiet           bool
	bytes, maxBytes int64
	backups         int
}

func New(name string) (*Logger, error) {
	return openLogger(".", name, false)
}

// NewLocal writes an installation-owned log without sending records to the
// controller's stdout. The caller holds the installation's operation lock.
func NewLocal(root, name string) (*Logger, error) { return openLogger(root, name, true) }

func openLogger(root, name string, quiet bool) (*Logger, error) {
	if filepath.Base(name) != name {
		return nil, fmt.Errorf("invalid logger name")
	}
	directory := filepath.Join(root, "logs")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, name+".log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	mb := int64(16)
	if raw := os.Getenv("RAPTURE_LOG_MAX_MB"); raw != "" {
		mb, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || mb < 1 || mb > 1024 {
			file.Close()
			return nil, fmt.Errorf("RAPTURE_LOG_MAX_MB must be 1..1024")
		}
	}
	return &Logger{file: file, name: name, path: path, quiet: quiet, debug: os.Getenv("RAPTURE_DEBUG") == "1", bytes: info.Size(), maxBytes: mb << 20, backups: 3}, nil
}
func (l *Logger) rotate() error {
	if err := l.file.Close(); err != nil {
		return err
	}
	l.file = nil
	for n := l.backups; n >= 1; n-- {
		source := l.path
		if n > 1 {
			source = fmt.Sprintf("%s.%d", l.path, n-1)
		}
		target := fmt.Sprintf("%s.%d", l.path, n)
		if n == l.backups {
			if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
		if err := os.Rename(source, target); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	l.file = f
	l.bytes = 0
	return nil
}
func (l *Logger) record(level, message, session string, attrs ...any) {
	entry := map[string]any{"time": time.Now().UTC().Format(time.RFC3339Nano), "level": level, "component": l.name, "message": message}
	if session != "" {
		entry["session"] = session
	}
	for i := 0; i+1 < len(attrs); i += 2 {
		if key, ok := attrs[i].(string); ok {
			if _, reserved := entry[key]; !reserved {
				entry[key] = attrs[i+1]
			}
		}
	}
	encoded, err := json.Marshal(entry)
	if err != nil {
		encoded, _ = json.Marshal(map[string]any{"level": "error", "message": "invalid log attributes", "component": l.name})
	}
	encoded = append(encoded, '\n')
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.quiet {
		os.Stdout.Write(encoded)
	}
	if l.file == nil {
		return
	}
	if l.bytes+int64(len(encoded)) > l.maxBytes {
		if err := l.rotate(); err != nil {
			fmt.Fprintln(os.Stderr, "GX log rotation failed:", err)
			return
		}
	}
	n, err := l.file.Write(encoded)
	l.bytes += int64(n)
	if err != nil {
		fmt.Fprintln(os.Stderr, "GX log write failed:", err)
	}
}
func (l *Logger) Printf(format string, args ...any) {
	l.record("info", fmt.Sprintf(format, args...), "")
}
func (l *Logger) Event(message string, attrs ...any) { l.record("info", message, "", attrs...) }
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}
func (l *Logger) Tagged(tag string) *Tag { t := &Tag{logger: l}; t.SetTag(tag); return t }

type Tag struct {
	logger *Logger
	tag    atomic.Pointer[string]
}

func (t *Tag) SetTag(tag string) { t.tag.Store(&tag) }
func (t *Tag) Tag() string {
	if tag := t.tag.Load(); tag != nil {
		return *tag
	}
	return ""
}
func (t *Tag) Printf(format string, args ...any) {
	t.logger.record("info", fmt.Sprintf(format, args...), t.Tag())
}
func (t *Tag) Event(message string, attrs ...any) {
	t.logger.record("info", message, t.Tag(), attrs...)
}
func (t *Tag) Debugf(format string, args ...any) {
	if t.logger.debug {
		t.logger.record("debug", fmt.Sprintf(format, args...), t.Tag())
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	l.record("error", fmt.Sprintf(format, args...), "")
}
func (t *Tag) Errorf(format string, args ...any) {
	t.logger.record("error", fmt.Sprintf(format, args...), t.Tag())
}
