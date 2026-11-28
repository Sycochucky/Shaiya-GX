// Package config resolves runtime settings from the environment, falling
// back to .local/runtime.env for local overrides. SQLite needs no credentials.
package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const ServerID = 1

// LoadRuntimeEnv reads KEY=VALUE lines from .local/runtime.env into the
// process environment without overriding variables that are already set.
func LoadRuntimeEnv() error {
	// Explicit data/database overrides keep the caller's working directory.
	// Otherwise standalone binaries find their own package even when launched
	// from another directory (including a double-click inside build/).
	if os.Getenv("RAPTURE_DATA_DIR") == "" && os.Getenv("RAPTURE_DB_PATH") == "" {
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		if resolved, err := filepath.EvalSymlinks(executable); err == nil {
			executable = resolved
		}
		directory := filepath.Dir(executable)
		for _, root := range []string{directory, filepath.Dir(directory)} {
			if info, err := os.Stat(filepath.Join(root, "data", "rapture.sqlite")); err == nil && !info.IsDir() {
				if err = os.Chdir(root); err != nil {
					return fmt.Errorf("enter server directory: %w", err)
				}
				break
			}
		}
	}
	return LoadRuntimeEnvFrom(".")
}

// ReadRuntimeEnv validates a runtime.env without changing the process environment.
func ReadRuntimeEnv(root string) (map[string]string, error) {
	file, err := os.Open(RuntimeEnvPath(root))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return parseRuntimeEnv(bufio.NewScanner(file))
}

func RuntimeEnvPath(root string) string { return filepath.Join(root, ".local", "runtime.env") }

func parseRuntimeEnv(scanner *bufio.Scanner) (map[string]string, error) {
	assignments := map[string]string{}
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		key, value, ok := strings.Cut(text, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		valid := ok && key != ""
		for i, c := range key {
			if !(c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || i > 0 && c >= '0' && c <= '9') {
				valid = false
			}
		}
		if !valid {
			return nil, fmt.Errorf("invalid runtime.env assignment at line %d", line)
		}
		assignments[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return assignments, nil
}

// LoadRuntimeEnvFrom is LoadRuntimeEnv for the checkout at root.
func LoadRuntimeEnvFrom(root string) error {
	assignments, err := ReadRuntimeEnv(root)
	if err != nil {
		return err
	}
	for key, value := range assignments {
		if _, set := os.LookupEnv(key); !set {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	resolvePaths(root)
	return nil
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// DatabasePath is the mutable runtime copy, never the versioned seed.
func DatabasePath() string { return env("RAPTURE_DB_PATH", filepath.Join("runtime", "rapture.sqlite")) }
func DatabaseSeed() string { return env("RAPTURE_DB_SEED", filepath.Join(DataDir(), "rapture.sqlite")) }

// DSN applies connection settings to every connection, including both servers.
func DSN(_ string) (string, error) {
	return sqliteDSN(false)
}

// ReadDSN opens a WAL reader that cannot mutate gameplay data.
func ReadDSN() (string, error) { return sqliteDSN(true) }

func sqliteDSN(readOnly bool) (string, error) {
	path, err := filepath.Abs(DatabasePath())
	if err != nil {
		return "", err
	}
	u := url.URL{Scheme: "file", Path: sqliteURIPath(filepath.ToSlash(path))}
	q := url.Values{}
	q.Add("_pragma", "busy_timeout(10000)")
	q.Add("_pragma", "foreign_keys(ON)")
	q.Add("_pragma", "trusted_schema(OFF)")
	if readOnly {
		q.Set("mode", "ro")
		q.Add("_pragma", "query_only(ON)")
	} else {
		q.Set("mode", "rw")
		q.Add("_pragma", "journal_mode(WAL)")
		q.Add("_pragma", "synchronous(FULL)")
		q.Set("_txlock", "immediate")
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// DataDir holds map data (data/maps/*.svmap).
func DataDir() string { return env("RAPTURE_DATA_DIR", "data") }

const (
	PublicIPKey    = "RAPTURE_PUBLIC_IP"
	LoginListenKey = "RAPTURE_LOGIN_LISTEN"
	WorldListenKey = "RAPTURE_WORLD_LISTEN"
)

// PublicIP is the world IPv4 address sent to clients after server selection.
func PublicIP() string { return env(PublicIPKey, "127.0.0.1") }

func LoginListen() string { return env(LoginListenKey, "127.0.0.1:30800") }
func WorldListen() string { return env(WorldListenKey, "127.0.0.1:30810") }

const (
	DefaultCombatDEXFirst  = 500
	DefaultCombatDEXSecond = 800
)

// CombatDEXThresholds reads the two points where DEX gains change rate.
func CombatDEXThresholds() (uint32, uint32, error) {
	const firstKey = "RAPTURE_COMBAT_DEX_FIRST_THRESHOLD"
	const secondKey = "RAPTURE_COMBAT_DEX_SECOND_THRESHOLD"
	first, err := strconv.ParseUint(env(firstKey, strconv.Itoa(DefaultCombatDEXFirst)), 10, 32)
	if err != nil || first == 0 {
		return 0, 0, fmt.Errorf("%s must be a positive 32-bit integer", firstKey)
	}
	second, err := strconv.ParseUint(env(secondKey, strconv.Itoa(DefaultCombatDEXSecond)), 10, 32)
	if err != nil || second <= first {
		return 0, 0, fmt.Errorf("%s must be a 32-bit integer greater than %s", secondKey, firstKey)
	}
	return uint32(first), uint32(second), nil
}

// Launchers and direct binaries resolve storage/data against their checkout.
func resolvePaths(root string) {
	for key, value := range map[string]string{"RAPTURE_DATA_DIR": DataDir(), "RAPTURE_DB_PATH": DatabasePath(), "RAPTURE_DB_SEED": DatabaseSeed()} {
		if !filepath.IsAbs(value) {
			if absolute, err := filepath.Abs(filepath.Join(root, value)); err == nil {
				os.Setenv(key, absolute)
			}
		}
	}
}

// A Windows drive is part of the URI path, never its authority. Without the
// leading slash net/url emits file://C:/..., which SQLite rejects.
func sqliteURIPath(path string) string {
	if len(path) >= 2 && path[1] == ':' {
		return "/" + path
	}
	return path
}

// FallDamageEnabled defaults to the GX landing rule; an explicit false opts out.
func FallDamageEnabled() bool {
	value := os.Getenv("RAPTURE_FALL_DAMAGE")
	if value == "" {
		return true
	}
	enabled, _ := strconv.ParseBool(value)
	return enabled
}
