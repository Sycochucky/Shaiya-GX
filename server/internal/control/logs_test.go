package control

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLiveLogsFollowPartialRecordsAndRotation(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "logs")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "gx_login.log")
	reader := NewLogReader(root)
	poll := func(want ...string) {
		t.Helper()
		rows, err := reader.Poll()
		if err != nil || len(rows) != len(want) {
			t.Fatalf("read %d rows, expected %d: %v", len(rows), len(want), err)
		}
		for i, text := range want {
			if rows[i].Service != Login || rows[i].Message != text {
				t.Fatalf("unexpected row: %+v", rows[i])
			}
		}
	}
	appendLog := func(text string) {
		t.Helper()
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.WriteString(text); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	line := func(message string) string {
		return `{"time":"2026-11-25T00:00:00Z","level":"info","message":"` + message + "\"}\n"
	}
	poll()
	appendLog(line("one") + strings.TrimSuffix(line("two"), "\n"))
	poll("one")
	poll()
	appendLog("\n" + line("three"))
	if err := os.Rename(path, path+".1"); err != nil {
		t.Fatal(err)
	}
	appendLog(line("four"))
	poll("two", "three", "four")
	poll()
}
