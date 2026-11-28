package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
)

func gamePathFile(root string) string {
	return filepath.Join(root, ".local", "game-executable")
}

// gamePath recognizes the client that belongs to this installation.
func gamePath(root string) string {
	if stored, err := os.ReadFile(gamePathFile(root)); err == nil {
		if path, err := gameExecutable(strings.TrimSpace(string(stored))); err == nil {
			return path
		}
	}
	name := "Shaiya GX"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	locations := []string{filepath.Join(root, "Client"), filepath.Join(filepath.Dir(root), "Client"), root}
	for current := root; ; {
		// A controller staged under server/out can find this checkout's normal client build.
		if info, err := os.Stat(filepath.Join(current, "cmd", "shaiya", "main.go")); err == nil && info.Mode().IsRegular() {
			locations = append(locations, filepath.Join(current, "build"))
			break
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	for _, dir := range locations {
		if path, err := gameExecutable(filepath.Join(dir, name)); err == nil {
			return path
		}
	}
	return ""
}

func (d *dashboard) runGame() {
	if d.gameBusy {
		return
	}
	if d.status[control.Login].State != control.Ready || d.status[control.World].State != control.Ready {
		d.showGameStatus("Start Login and World before running the game.")
		return
	}
	path := gamePath(d.manager.Root())
	if path == "" {
		d.pickGame(d.launchGame)
		return
	}
	d.launchGame(path)
}

func (d *dashboard) pickGame(launch func(string)) {
	root := d.manager.Root()
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Full path to the GX client")
	entry.SetText(gamePath(root))
	status := widget.NewLabel("Paste the full path to your Shaiya GX executable.")
	status.Wrapping = fyne.TextWrapWord
	use := widget.NewButton("Use executable", nil)
	use.Importance = widget.HighImportance
	picker := dialog.NewCustom("Choose Game", "Cancel", container.NewVBox(entry, status, use), d.window)
	picker.Resize(fyne.NewSize(630, 180))
	use.OnTapped = func() {
		path, err := gameExecutable(strings.Trim(strings.TrimSpace(entry.Text), "\\\""))
		if err == nil {
			pathFile := gamePathFile(root)
			if err = os.MkdirAll(filepath.Dir(pathFile), 0700); err == nil {
				err = os.WriteFile(pathFile, []byte(path), 0600)
			}
		}
		if err != nil {
			status.SetText(err.Error())
			return
		}
		picker.Hide()
		d.showGameStatus("Game client selected.")
		if launch != nil {
			launch(path)
		}
	}
	entry.OnSubmitted = func(string) { use.OnTapped() }
	picker.Show()
}

func (d *dashboard) showGameStatus(message string) {
	d.refreshSelectedGame()
	d.gameStatus.SetText(message)
	d.gameStatus.Show()
}

func (d *dashboard) launchGame(path string) {
	if d.gameBusy {
		return
	}
	d.gameBusy = true
	d.game.Disable()
	d.showGameStatus("Starting " + filepath.Base(path) + "…")
	update := func(message string, active bool) {
		fyne.Do(func() {
			if d.ctx.Err() != nil {
				return
			}
			d.gameBusy = active
			setEnabled(d.game, !d.busy && !d.gameBusy)
			d.showGameStatus(message)
		})
	}
	go func() {
		cmd, err := gameCommand(path, d.manager.Snapshot(d.ctx))
		if d.ctx.Err() != nil {
			return
		}
		if err == nil {
			err = startGame(cmd, update)
		}
		if err != nil {
			update("Game could not start: "+err.Error(), false)
		}
	}()
}

// The released client writes diagnostics to both its own log and stderr.
// Watch this process's output so a previous launch cannot signal readiness.
type gameStartupLog struct {
	ready chan struct{}
	tail  string
	seen  bool
}

func (l *gameStartupLog) Write(p []byte) (int, error) {
	const marker = "Client ready framebuffer="
	if !l.seen {
		output := l.tail + string(p)
		if strings.Contains(output, marker) {
			l.seen = true
			l.tail = ""
			close(l.ready)
		} else {
			l.tail = output[max(0, len(output)-len(marker)+1):]
		}
	}
	return len(p), nil
}

func startGame(cmd *exec.Cmd, update func(string, bool)) error {
	ready := make(chan struct{})
	startup := &gameStartupLog{ready: ready}
	cmd.Stderr = startup
	if err := cmd.Start(); err != nil {
		return err
	}
	logPath := filepath.Join(cmd.Dir, "logs", "client.log")
	update("Game started. Loading the login screen…", true)
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	timer := time.NewTimer(15 * time.Second)
	defer timer.Stop()
	slow := timer.C
	for {
		select {
		case <-ready:
			ready, slow = nil, nil
			update(fmt.Sprintf("Login screen ready. Login %s, World port %s.", cmd.Args[2], cmd.Args[4]), true)
		case <-slow:
			slow = nil
			update("The game is still loading. Startup log: "+logPath, true)
		case err := <-done:
			if err != nil {
				update(fmt.Sprintf("Game exited: %v. Startup log: %s", err, logPath), false)
			} else if !startup.seen {
				update("Game closed during startup. Startup log: "+logPath, false)
			} else {
				update("Game closed.", false)
			}
			return nil
		}
	}
}

func gameExecutable(path string) (string, error) {
	path = filepath.FromSlash(path)
	if !filepath.IsAbs(path) {
		return "", errors.New("choose the full path to the GX client")
	}
	if runtime.GOOS == "windows" && !strings.EqualFold(filepath.Ext(path), ".exe") {
		return "", errors.New("choose the GX client .exe")
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("choose a GX client executable file")
	}
	return path, nil
}

func gameCommand(path string, status map[control.Service]control.Status) (*exec.Cmd, error) {
	if status[control.Login].State != control.Ready || status[control.World].State != control.Ready {
		return nil, errors.New("Login and World must be ready to launch the GX client")
	}
	path, err := gameExecutable(path)
	if err != nil {
		return nil, err
	}
	host, port, err := net.SplitHostPort(status[control.Login].Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid Login endpoint: %w", err)
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		host = "127.0.0.1"
		if ip.To4() == nil {
			host = "::1"
		}
	}
	_, worldPort, err := net.SplitHostPort(status[control.World].Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid World endpoint: %w", err)
	}
	portNumber, err := strconv.ParseUint(worldPort, 10, 16)
	if err != nil || portNumber == 0 {
		return nil, errors.New("invalid World port")
	}
	cmd := exec.Command(path, "-login", net.JoinHostPort(host, port), "-world-port", worldPort)
	cmd.Dir = filepath.Dir(path)
	setGameProcessAttributes(cmd)
	return cmd, nil
}
