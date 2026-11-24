package control

import (
	"errors"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func (m *Manager) processes() (map[Service][]int, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)
	names := map[string]Service{exeName(Login): Login, exeName(World): World}
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	found := make(map[Service][]int)
	var path [32768]uint16
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		service, ok := names[strings.ToLower(windows.UTF16ToString(entry.ExeFile[:]))]
		if !ok {
			continue
		}
		process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, entry.ProcessID)
		if err != nil {
			continue
		}
		size := uint32(len(path))
		err = windows.QueryFullProcessImageName(process, 0, &path[0], &size)
		windows.CloseHandle(process)
		if err == nil && strings.EqualFold(filepath.Clean(windows.UTF16ToString(path[:size])), m.binary(service)) {
			found[service] = append(found[service], int(entry.ProcessID))
		}
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return nil, err
	}
	return found, nil
}
