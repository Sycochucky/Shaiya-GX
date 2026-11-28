package control

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func imagePath(handle windows.Handle) (string, error) {
	var path [32768]uint16
	size := uint32(len(path))
	err := windows.QueryFullProcessImageName(handle, 0, &path[0], &size)
	return filepath.Clean(windows.UTF16ToString(path[:size])), err
}

func processExited(handle windows.Handle) bool {
	var code uint32
	return windows.GetExitCodeProcess(handle, &code) == nil && code != 259 // STILL_ACTIVE
}

func terminateOwned(pid int, path string) error {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_TERMINATE|windows.SYNCHRONIZE, false, uint32(pid))
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		return nil
	}
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	if processExited(handle) {
		return nil
	}
	actual, err := imagePath(handle)
	if err != nil {
		return err
	}
	if !strings.EqualFold(actual, path) {
		return errors.New("process image changed before termination")
	}
	if err := windows.TerminateProcess(handle, 1); err != nil && !processExited(handle) {
		return err
	}
	status, err := windows.WaitForSingleObject(handle, 5_000)
	if err != nil {
		return err
	}
	if status != windows.WAIT_OBJECT_0 {
		return errors.New("force-stopped process did not exit")
	}
	return nil
}

func scanGXProcesses() ([]gxProcess, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)
	names := map[string]Service{exeName(Login): Login, exeName(World): World}
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var found []gxProcess
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		service, ok := names[strings.ToLower(windows.UTF16ToString(entry.ExeFile[:]))]
		if !ok {
			continue
		}
		process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, entry.ProcessID)
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("cannot identify %s PID %d: %w", service.Label(), entry.ProcessID, err)
		}
		path, err := imagePath(process)
		exited := err != nil && processExited(process)
		windows.CloseHandle(process)
		if exited {
			continue
		}
		if err != nil && !errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			return nil, fmt.Errorf("cannot identify %s PID %d: %w", service.Label(), entry.ProcessID, err)
		}
		if err == nil {
			found = append(found, gxProcess{service: service, pid: int(entry.ProcessID), executable: path})
		}
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return nil, err
	}
	return found, nil
}
