package control

import (
	"context"
	"errors"
	"net"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/controlplane"
)

type gxProcess struct {
	service    Service
	pid        int
	executable string
}

func samePath(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// SameInstallation compares portable roots using the platform's path rules.
func SameInstallation(a, b string) bool { return samePath(a, b) }

func (m *Manager) processes() (map[Service][]int, error) {
	running, err := scanGXProcesses()
	if err != nil {
		return nil, err
	}
	found := make(map[Service][]int)
	for _, proc := range running {
		if samePath(proc.executable, m.binary(proc.service)) {
			found[proc.service] = append(found[proc.service], proc.pid)
		}
	}
	return found, nil
}

func configuredEndpoint(settings map[string]string, service Service) string {
	key, fallback := "RAPTURE_LOGIN_LISTEN", "127.0.0.1:30800"
	if service == World {
		key, fallback = "RAPTURE_WORLD_LISTEN", "127.0.0.1:30810"
	}
	if endpoint := config.RuntimeValue(settings, key); endpoint != "" {
		return endpoint
	}
	return fallback
}

func overlappingEndpoints(a, b string) bool {
	hostA, portA, errA := net.SplitHostPort(a)
	hostB, portB, errB := net.SplitHostPort(b)
	if errA != nil || errB != nil || portA == "0" || portA != portB {
		return false
	}
	if strings.EqualFold(hostA, hostB) || hostA == "" || hostB == "" {
		return true
	}
	ipA, ipB := net.ParseIP(hostA), net.ParseIP(hostB)
	return ipA.IsUnspecified() || ipB.IsUnspecified() || ipA != nil && ipA.Equal(ipB) ||
		strings.EqualFold(hostA, "localhost") && ipB.IsLoopback() ||
		strings.EqualFold(hostB, "localhost") && ipA.IsLoopback()
}

// ConflictingInstallation verifies ownership of the configured ports. It
// reads live control endpoints without starting, stopping or adopting a service.
func (m *Manager) ConflictingInstallation(ctx context.Context) (*Manager, error) {
	return m.conflictingInstallation(ctx, map[Service]bool{Login: true, World: true})
}

func (m *Manager) conflictingInstallation(ctx context.Context, selected map[Service]bool) (*Manager, error) {
	settings, err := config.ReadRuntimeEnv(m.root)
	if err != nil {
		return nil, err
	}
	running, err := scanGXProcesses()
	if err != nil {
		return nil, err
	}
	owned := make(map[Service]bool)
	for _, proc := range running {
		if samePath(proc.executable, m.binary(proc.service)) {
			owned[proc.service] = true
		}
	}
	var found *Manager
	for _, proc := range running {
		owner, report := m.verifyOtherInstallation(ctx, proc)
		if owner == nil {
			continue
		}
		conflicts := false
		for service, starting := range selected {
			if starting && !owned[service] && overlappingEndpoints(configuredEndpoint(settings, service), report.Endpoint) {
				conflicts = true
				break
			}
		}
		if !conflicts {
			continue
		}
		if found != nil && !samePath(found.Root(), owner.Root()) {
			return nil, errors.New("different GX installations are using the configured ports")
		}
		found = owner
	}
	return found, nil
}

func (m *Manager) verifyOtherInstallation(ctx context.Context, proc gxProcess) (*Manager, controlplane.ServiceStatus) {
	root := filepath.Dir(filepath.Dir(proc.executable))
	owner := New(root)
	if samePath(root, m.Root()) || !samePath(proc.executable, owner.binary(proc.service)) {
		return nil, controlplane.ServiceStatus{}
	}
	report, err := controlplane.ReadStatus(ctx, root, string(proc.service), proc.pid, proc.executable)
	if err != nil || report.Endpoint == "" {
		return nil, controlplane.ServiceStatus{}
	}
	return owner, report
}

// RunningInstallations lists verified local owners regardless of port. A
// standalone dashboard can use this list without guessing another root.
func (m *Manager) RunningInstallations(ctx context.Context) ([]string, error) {
	processes, err := scanGXProcesses()
	if err != nil {
		return nil, err
	}
	var roots []string
	for _, proc := range processes {
		owner, _ := m.verifyOtherInstallation(ctx, proc)
		if owner == nil || owner.CheckInstallation() != nil {
			continue
		}
		seen := false
		for _, root := range roots {
			if samePath(root, owner.Root()) {
				seen = true
				break
			}
		}
		if !seen {
			roots = append(roots, owner.Root())
		}
	}
	sort.Strings(roots)
	return roots, nil
}
