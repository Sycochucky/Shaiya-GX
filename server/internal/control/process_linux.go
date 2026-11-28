package control

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

func terminateOwned(pid int, expected string) error {
	handle, err := unix.PidfdOpen(pid, 0)
	if errors.Is(err, unix.ESRCH) {
		return nil
	}
	if err != nil {
		return err
	}
	defer unix.Close(handle)
	actual, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(pid), "exe"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if strings.TrimSuffix(actual, " (deleted)") != expected {
		return errors.New("process image changed before termination")
	}
	err = unix.PidfdSendSignal(handle, unix.SIGKILL, nil, 0)
	if errors.Is(err, unix.ESRCH) {
		return nil
	}
	return err
}

func scanGXProcesses() ([]gxProcess, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	names := map[string]Service{exeName(Login): Login, exeName(World): World}
	var found []gxProcess
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		if path, err := os.Readlink(filepath.Join("/proc", entry.Name(), "exe")); err == nil {
			path = strings.TrimSuffix(path, " (deleted)")
			if service, ok := names[filepath.Base(path)]; ok {
				found = append(found, gxProcess{service: service, pid: pid, executable: path})
			}
		}
	}
	return found, nil
}
