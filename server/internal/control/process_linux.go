package control

import (
	"os"
	"path/filepath"
	"strconv"
)

func (m *Manager) processes() (map[Service][]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	paths := map[string]Service{}
	for _, service := range Services() {
		path := m.binary(service)
		paths[path], paths[path+" (deleted)"] = service, service
	}
	found := make(map[Service][]int)
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		if path, err := os.Readlink(filepath.Join("/proc", entry.Name(), "exe")); err == nil {
			if service, ok := paths[path]; ok {
				found[service] = append(found[service], pid)
			}
		}
	}
	return found, nil
}
