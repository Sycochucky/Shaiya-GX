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
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
)

type Service string

const (
	Login Service = "gx_login"
	World Service = "gx_world"
)

func Services() []Service { return []Service{Login, World} }

func (s Service) Label() string {
	if s == operatorLog {
		return "Operator"
	}
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
	State     State
	PID       int
	Endpoint  string
	Players   int
	Shield    netx.ShieldStatus
	Detail    string
	Processes int
}

type Action string

const (
	Start     Action = "Start"
	Stop      Action = "Stop"
	Restart   Action = "Restart"
	ForceStop Action = "Force stop"
)

type Manager struct {
	root         string
	mu           sync.Mutex
	failure      map[Service]string
	startTimeout time.Duration
}

func New(root string) *Manager {
	return &Manager{root: root, failure: make(map[Service]string), startTimeout: 40 * time.Second}
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

// CheckInstallation verifies the executables required by the full dashboard.
func (m *Manager) CheckInstallation() error {
	return m.checkExecutables([]Service{Login, World, "gx_console"})
}

func (m *Manager) checkExecutables(targets []Service) error {
	var missing []string
	for _, service := range targets {
		info, err := os.Stat(m.binary(service))
		if err != nil || !info.Mode().IsRegular() {
			missing = append(missing, exeName(service))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("incomplete GX installation: missing %s", strings.Join(missing, ", "))
	}
	return nil
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
			return Status{State: Unknown, PID: pid, Processes: 1, Detail: "Process is running. Readiness is unavailable."}
		}
		return Status{State: Ready, PID: pid, Processes: 1, Endpoint: status.Endpoint, Players: status.Players, Shield: status.Shield}
	default:
		return Status{State: Unknown, Processes: len(found), Detail: "Multiple processes match this installation."}
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
			status.Endpoint = configuredEndpoint(configured, service)
		}
		result[service] = status
	}
	return result
}

// Apply serializes GUI and CLI changes to this installation. Restart starts
// only services whose old processes were verified stopped.
func (m *Manager) Apply(ctx context.Context, action Action, targets []Service, progress func(Service, Status)) error {
	if action != Start && action != Stop && action != Restart && action != ForceStop {
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
	if action == Start {
		owner, err := m.conflictingInstallation(ctx, selected)
		if err != nil {
			return err
		}
		if owner != nil {
			return fmt.Errorf("a running GX installation already owns these ports: %s", owner.Root())
		}
	}
	if action == Start || action == Restart {
		var starting []Service
		for _, service := range Services() {
			if selected[service] {
				starting = append(starting, service)
			}
		}
		if err := m.checkExecutables(starting); err != nil {
			return err
		}
	}
	var failures []error
	if action != Start {
		for _, service := range []Service{World, Login} {
			if !selected[service] {
				continue
			}
			if progress != nil {
				progress(service, Status{State: Stopping})
			}
			if err := m.stop(ctx, service, action == ForceStop); err != nil {
				failures = append(failures, fmt.Errorf("stop %s: %w", service.Label(), err))
				delete(selected, service)
			} else {
				m.mu.Lock()
				delete(m.failure, service)
				m.mu.Unlock()
			}
			if progress != nil {
				progress(service, m.Snapshot(ctx)[service])
			}
		}
	}
	if action == Start || action == Restart {
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
	timeout := m.startTimeout
	if service == World {
		// World must finish loading map navigation and spawn data before it
		// can advertise readiness. Cold Windows disks can take over 40s.
		timeout = max(timeout, 120*time.Second)
	}
	deadline, cancel := context.WithTimeout(ctx, timeout)
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
			cause := fmt.Errorf("service did not report ready: %w", deadline.Err())
			if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				return errors.Join(cause, fmt.Errorf("could not stop the launched process: %w", err))
			}
			select {
			case <-exited:
				return cause
			case <-time.After(5 * time.Second):
				return errors.Join(cause, errors.New("launched process exit could not be verified"))
			}
		case <-tick.C:
			if status, err := controlplane.ReadStatus(deadline, m.root, string(service), cmd.Process.Pid, path); err == nil && status.Endpoint != "" {
				return nil
			}
		}
	}
}

func (m *Manager) stop(ctx context.Context, service Service, force bool) error {
	processes, err := m.processes()
	if err != nil {
		return err
	}
	found := processes[service]
	if len(found) == 0 {
		return nil
	}
	if force {
		for _, pid := range found {
			if err := terminateOwned(pid, m.binary(service)); err != nil {
				return err
			}
		}
	} else {
		if len(found) != 1 {
			return errors.New("multiple matching processes")
		}
		if err := controlplane.RequestStop(ctx, m.root, string(service), found[0], m.binary(service)); err != nil {
			return err
		}
	}
	deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()
	var observationErr error
	for {
		select {
		case <-deadline.Done():
			return fmt.Errorf("could not verify process exit: %w", errors.Join(deadline.Err(), observationErr))
		case <-tick.C:
			processes, err := m.processes()
			if err != nil {
				observationErr = err
				continue
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
