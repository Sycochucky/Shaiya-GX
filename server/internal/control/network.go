package control

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

type NetworkSettings struct {
	LoginListen string
	WorldListen string
	PublicIP    string
}

func (m *Manager) NetworkSettings() (NetworkSettings, map[string]bool, error) {
	values, err := config.ReadRuntimeEnv(m.root)
	if err != nil {
		return NetworkSettings{}, nil, err
	}
	effective := func(key, fallback string) string {
		if value := config.RuntimeValue(values, key); value != "" {
			return value
		}
		return fallback
	}
	settings := NetworkSettings{
		LoginListen: effective(config.LoginListenKey, "127.0.0.1:30800"),
		WorldListen: effective(config.WorldListenKey, "127.0.0.1:30810"),
		PublicIP:    effective(config.PublicIPKey, "127.0.0.1"),
	}
	overridden := make(map[string]bool, 3)
	for _, key := range []string{config.LoginListenKey, config.WorldListenKey, config.PublicIPKey} {
		_, overridden[key] = os.LookupEnv(key)
	}
	return settings, overridden, nil
}

func parseListenEndpoint(name, endpoint string) (net.IP, int, error) {
	host, portText, err := net.SplitHostPort(strings.TrimSpace(endpoint))
	if err != nil {
		return nil, 0, fmt.Errorf("%s must be an IPv4 address and port, for example 0.0.0.0:30800", name)
	}
	ip := net.ParseIP(host).To4()
	port, portErr := strconv.Atoi(portText)
	if ip == nil || portErr != nil || port < 1 || port > 65535 {
		return nil, 0, fmt.Errorf("%s must be an IPv4 address with a port from 1 to 65535", name)
	}
	return ip, port, nil
}

func validateNetworkSettings(settings NetworkSettings) error {
	loginIP, loginPort, err := parseListenEndpoint("Login listen", settings.LoginListen)
	if err != nil {
		return err
	}
	worldIP, worldPort, err := parseListenEndpoint("World listen", settings.WorldListen)
	if err != nil {
		return err
	}
	if loginPort == worldPort && (loginIP.IsUnspecified() || worldIP.IsUnspecified() || loginIP.Equal(worldIP)) {
		return errors.New("Login and World must use different ports when they bind the same address")
	}
	publicIP := net.ParseIP(strings.TrimSpace(settings.PublicIP)).To4()
	if publicIP == nil || publicIP.IsUnspecified() || publicIP.IsMulticast() {
		return errors.New("advertised address must be a reachable numeric IPv4 address")
	}
	publicBind := loginIP.IsUnspecified() || worldIP.IsUnspecified() || !loginIP.IsLoopback() || !worldIP.IsLoopback()
	if publicBind && publicIP.IsLoopback() {
		return errors.New("advertised address cannot be loopback while Login or World accepts remote connections")
	}
	return nil
}

func endpointAvailable(endpoint string) bool {
	listener, err := net.Listen("tcp4", endpoint)
	if err != nil {
		return false
	}
	return listener.Close() == nil
}

func (m *Manager) SaveNetworkSettings(ctx context.Context, settings NetworkSettings) error {
	settings.LoginListen = strings.TrimSpace(settings.LoginListen)
	settings.WorldListen = strings.TrimSpace(settings.WorldListen)
	settings.PublicIP = strings.TrimSpace(settings.PublicIP)
	if err := validateNetworkSettings(settings); err != nil {
		return err
	}
	release, err := m.lock(ctx)
	if err != nil {
		return err
	}
	defer release()
	processes, err := m.processes()
	if err != nil {
		return err
	}
	if len(processes[Login])+len(processes[World]) != 0 {
		return errors.New("stop Login and World before changing network settings")
	}
	if !endpointAvailable(settings.LoginListen) {
		return fmt.Errorf("Login endpoint %s is unavailable", settings.LoginListen)
	}
	if !endpointAvailable(settings.WorldListen) {
		return fmt.Errorf("World endpoint %s is unavailable", settings.WorldListen)
	}
	inherited := make(map[string]string, 3)
	for _, key := range []string{config.LoginListenKey, config.WorldListenKey, config.PublicIPKey} {
		if value, ok := os.LookupEnv(key); ok {
			inherited[key] = value
		}
	}
	return config.SaveRuntimeSettings(m.root, []config.RuntimeSetting{
		{Key: config.LoginListenKey, Value: settings.LoginListen},
		{Key: config.WorldListenKey, Value: settings.WorldListen},
		{Key: config.PublicIPKey, Value: settings.PublicIP},
	}, inherited)
}

func suggestPorts(available func(int) bool) (int, int, error) {
	find := func(start, avoid int) int {
		for port := start; port <= 65535 && port < start+4096; port++ {
			if port != avoid && available(port) {
				return port
			}
		}
		return 0
	}
	login := find(30800, 0)
	world := find(30810, login)
	if login == 0 || world == 0 {
		return 0, 0, errors.New("no free GX port pair found in the smart-port range")
	}
	return login, world, nil
}

func SuggestPorts() (int, int, error) {
	return suggestPorts(func(port int) bool {
		return endpointAvailable(net.JoinHostPort("0.0.0.0", strconv.Itoa(port)))
	})
}

func LocalIPv4s() []string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return []string{"127.0.0.1"}
	}
	seen := map[string]bool{"127.0.0.1": true}
	result := []string{"127.0.0.1"}
	for _, address := range addresses {
		ip, _, err := net.ParseCIDR(address.String())
		if err != nil || ip.To4() == nil || ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
			continue
		}
		value := ip.To4().String()
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result[1:])
	return result
}
