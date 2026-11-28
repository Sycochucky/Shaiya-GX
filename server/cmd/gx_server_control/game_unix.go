//go:build !windows

package main

import "os/exec"

func setGameProcessAttributes(_ *exec.Cmd) {}
