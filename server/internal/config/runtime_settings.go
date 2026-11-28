package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// RuntimeSetting is one ordered assignment for .local/runtime.env.
type RuntimeSetting struct {
	Key   string
	Value string
}

// SaveRuntimeSettings preserves comments and unrelated assignments. Values
// inherited by the controller process remain authoritative and are not saved.
func SaveRuntimeSettings(root string, settings []RuntimeSetting, inherited map[string]string) error {
	if _, err := ReadRuntimeEnv(root); err != nil {
		return err
	}
	path := RuntimeEnvPath(root)
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	content := string(raw)
	eol := "\n"
	if strings.Contains(content, "\r\n") {
		eol = "\r\n"
	}
	for _, setting := range settings {
		if _, overridden := inherited[setting.Key]; overridden {
			continue
		}
		content = replaceSetting(content, setting.Key, setting.Value, eol)
	}
	if content == string(raw) {
		return nil
	}
	return writeRuntimeFile(path, content)
}

func writeRuntimeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".runtime-env-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.WriteString(content)
	if writeErr == nil {
		writeErr = file.Sync()
	}
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	return replaceFile(file.Name(), path)
}
