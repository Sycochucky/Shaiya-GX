package main

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func setGameProcessAttributes(cmd *exec.Cmd) {
	// The client executable uses the console subsystem and opens its own GUI.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NO_WINDOW}
}
