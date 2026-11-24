// Package control manages the detached Login and World processes for a single GX installation.
package control

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/controlplane"
	"github.com/Spelunkern/Shaiya-GX/server/internal/filelock"
)

type Service string

const (
	Login Service = "gx_login"
	World Service = "gx_world"
)

func Services() []Service { return []Service{Login, World} }

func (s Service) Label() string {
	if s == World {
		return "World"
	}
	return "Login"
}

type State string

const (
	Stopped  State = "Stopped"
	Starting State = "Starting"
	Ready    State = "Ready"
	Stopping State = "Stopping"
	Failed   State = "Failed"
	Unknown  State = "Unknown"
)

type Status struct {
	State    State
	PID      int
	Endpoint string
	Players  int
	Detail   string
}

type Action string

const (
	Start   Action = "Start"
	Stop    Action = "Stop"
	Restart Action = "Restart"
)

type Manager struct {
	root    string
	mu      sync.Mutex
	failure map[Service]string
}

func New(root string) *Manager {
	return &Manager{root: root, failure: make(map[Service]string)}
}

// Installation opens the portable root containing the controller's build directory.
func Installation() (*Manager, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	return New(filepath.Dir(filepath.Dir(exe))), nil
}

func (m *Manager) Root() string { return m.root }

func exeName(service Service) string {
	name := string(service)
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func (m *Manager) binary(service Service) string {
	return filepath.Join(m.root, "build", exeName(service))
}

func (m *Manager) lock(ctx context.Context) (func(), error) {
	directory := filepath.Join(m.root, "runtime", "control")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return filelock.Acquire(ctx, filepath.Join(directory, "operations.lock"))
}

func (m *Manager) current(ctx context.Context, service Service, found []int) Status {
	switch len(found) {
	case 0:
		m.mu.Lock()
		detail := m.failure[service]
		m.mu.Unlock()
		if detail != "" {
			return Status{State: Failed, Detail: detail}
		}
		return Status{State: Stopped}
	case 1:
		pid := found[0]
		status, err := controlplane.ReadStatus(ctx, m.root, string(service), pid, m.binary(service))
		if err != nil || status.Endpoint == "" {
			return Status{State: Unknown, PID: pid, Detail: "Process is running. Readiness is unavailable."}
		}
		return Status{State: Ready, PID: pid, Endpoint: status.Endpoint, Players: status.Players}
	default:
		return Status{State: Unknown, Detail: "Multiple processes match this installation."}
	}
}

func (m *Manager) Snapshot(ctx context.Context) map[Service]Status {
	result := make(map[Service]Status, 2)
	configured, configErr := config.ReadRuntimeEnv(m.root)
	processes, processErr := m.processes()
	for _, service := range Services() {
		status := Status{State: Unknown}
		if processErr == nil {
			status = m.current(ctx, service, processes[service])
		} else {
			status.Detail = processErr.Error()
		}
		if status.Endpoint == "" && configErr == nil {
			key, fallback := "RAPTURE_LOGIN_LISTEN", "127.0.0.1:30800"
			if service == World {
				key, fallback = "RAPTURE_WORLD_LISTEN", "127.0.0.1:30810"
			}
			status.Endpoint = config.RuntimeValue(configured, key)
			if status.Endpoint == "" {
				status.Endpoint = fallback
			}
		}
		result[service] = status
	}
	return result
}

// Apply serializes GUI and CLI changes to this installation. Restart starts
// only services whose old processes were verified stopped.
func (m *Manager) Apply(ctx context.Context, action Action, targets []Service, progress func(Service, Status)) error {
	if action != Start && action != Stop && action != Restart {
		return fmt.Errorf("unknown action %q", action)
	}
	selected := make(map[Service]bool, len(targets))
	for _, service := range targets {
		if service != Login && service != World {
			return fmt.Errorf("unknown GX service %q", service)
		}
		selected[service] = true
	}
	release, err := m.lock(ctx)
	if err != nil {
		return fmt.Errorf("acquire installation control: %w", err)
	}
	defer release()
	var failures []error
	if action != Start {
		for _, service := range []Service{World, Login} {
			if !selected[service] {
				continue
			}
			if progress != nil {
				progress(service, Status{State: Stopping})
			}
			if err := m.stop(ctx, service); err != nil {
				failures = append(failures, fmt.Errorf("stop %s: %w", service.Label(), err))
				delete(selected, service)
			}
			if progress != nil {
				progress(service, m.Snapshot(ctx)[service])
			}
		}
	}
	if action != Stop {
		for _, service := range Services() {
			if !selected[service] {
				continue
			}
			if progress != nil {
				progress(service, Status{State: Starting})
			}
			err := m.start(ctx, service)
			m.mu.Lock()
			delete(m.failure, service)
			if err != nil {
				m.failure[service] = err.Error()
				failures = append(failures, fmt.Errorf("start %s: %w", service.Label(), err))
			}
			m.mu.Unlock()
			if progress != nil {
				progress(service, m.Snapshot(ctx)[service])
			}
		}
	}
	return errors.Join(failures...)
}

func (m *Manager) start(ctx context.Context, service Service) error {
	processes, err := m.processes()
	if err != nil {
		return err
	}
	if found := processes[service]; len(found) > 0 {
		status := m.current(ctx, service, found)
		if status.State == Ready {
			return nil
		}
		return errors.New(status.Detail)
	}
	if _, err := config.ReadRuntimeEnv(m.root); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	path := m.binary(service)
	errorPath := filepath.Join(m.root, "runtime", "control", string(service)+".stderr")
	stderr, err := os.OpenFile(errorPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	cmd := exec.Command(path)
	cmd.Dir, cmd.Stderr = m.root, stderr
	detach(cmd)
	err = cmd.Start()
	stderr.Close()
	if err != nil {
		return err
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	deadline, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case err := <-exited:
			if line := lastStderr(errorPath); line != "" {
				return errors.New(line)
			}
			if err != nil {
				return err
			}
			return errors.New("process exited before reporting readiness")
		case <-deadline.Done():
			return fmt.Errorf("service did not report ready: %w", deadline.Err())
		case <-tick.C:
			if status, err := controlplane.ReadStatus(deadline, m.root, string(service), cmd.Process.Pid, path); err == nil && status.Endpoint != "" {
				return nil
			}
		}
	}
}

func (m *Manager) stop(ctx context.Context, service Service) error {
	processes, err := m.processes()
	if err != nil {
		return err
	}
	found := processes[service]
	if len(found) == 0 {
		return nil
	}
	if len(found) != 1 {
		return errors.New("multiple matching processes")
	}
	pid := found[0]
	if err := controlplane.RequestStop(ctx, m.root, string(service), pid, m.binary(service)); err != nil {
		return err
	}
	deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-deadline.Done():
			return fmt.Errorf("could not verify process exit: %w", deadline.Err())
		case <-tick.C:
			processes, err := m.processes()
			if err != nil {
				return err
			}
			if len(processes[service]) == 0 {
				return nil
			}
		}
	}
}

func lastStderr(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return ""
	}
	data := make([]byte, min(info.Size(), 4096))
	n, _ := f.ReadAt(data, info.Size()-int64(len(data)))
	last := strings.TrimSpace(string(data[:n]))
	if index := strings.LastIndexByte(last, '\n'); index >= 0 {
		last = last[index+1:]
	}
	return strings.TrimSpace(last)
}
