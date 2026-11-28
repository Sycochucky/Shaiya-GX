// Package controlplane owns loopback diagnostics and authenticated local operations.
package controlplane

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
	"github.com/Spelunkern/Shaiya-GX/server/internal/telemetry"
)

type descriptor struct {
	PID        int    `json:"pid"`
	Executable string `json:"executable"`
	Address    string `json:"address"`
	Token      string `json:"token"`
}
type Server struct {
	http       *http.Server
	listener   net.Listener
	done       chan struct{}
	path       string
	descriptor descriptor
	once       sync.Once
}

// ServiceStatus contains the operational data exposed to the local controller.
type ServiceStatus struct {
	Endpoint string            `json:"endpoint"`
	Players  int               `json:"players"`
	Shield   netx.ShieldStatus `json:"shield"`
}

func loopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	ip := net.ParseIP(host)
	return err == nil && ip != nil && ip.IsLoopback()
}
func Start(ctx context.Context, service, addr string, cancel context.CancelFunc, status func() ServiceStatus, shield *netx.Shield, observe ...func()) (*Server, error) {
	return start(ctx, service, addr, cancel, status, shield, nil, observe)
}

// StartWorld adds the authenticated World operator endpoint.
func StartWorld(ctx context.Context, addr string, cancel context.CancelFunc, status func() ServiceStatus, shield *netx.Shield, operator func(operatorwire.Request) (operatorwire.Response, error), observe ...func()) (*Server, error) {
	return start(ctx, "gx_world", addr, cancel, status, shield, operator, observe)
}

func start(ctx context.Context, service, addr string, cancel context.CancelFunc, status func() ServiceStatus, shield *netx.Shield, operator func(operatorwire.Request) (operatorwire.Response, error), observe []func()) (*Server, error) {
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	if !loopback(addr) {
		return nil, fmt.Errorf("diagnostics must bind a numeric loopback address")
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	var secret [32]byte
	if _, err = rand.Read(secret[:]); err != nil {
		listener.Close()
		return nil, err
	}
	exe, err := os.Executable()
	if err != nil {
		listener.Close()
		return nil, err
	}
	if resolved, e := filepath.EvalSymlinks(exe); e == nil {
		exe = resolved
	}
	s := &Server{listener: listener, done: make(chan struct{}), path: filepath.Join("runtime", "control", service+".json"), descriptor: descriptor{PID: os.Getpid(), Executable: exe, Address: listener.Addr().String(), Token: hex.EncodeToString(secret[:])}}
	mux := http.NewServeMux()
	authorize := func(r *http.Request) bool {
		return subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+s.descriptor.Token)) == 1
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if ctx.Err() != nil {
			http.Error(w, "stopping", 503)
			return
		}
		w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		if shield != nil {
			shield.Status()
		}
		for _, sample := range observe {
			sample()
		}
		telemetry.WritePrometheus(w)
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if ctx.Err() != nil {
			http.Error(w, "stopping", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		current := status()
		if shield != nil {
			current.Shield = shield.Status()
		}
		json.NewEncoder(w).Encode(current)
	})
	mux.HandleFunc("POST /shield", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if ctx.Err() != nil || shield == nil {
			http.Error(w, "shield unavailable", http.StatusServiceUnavailable)
			return
		}
		minutes, err := strconv.Atoi(r.URL.Query().Get("minutes"))
		if err != nil || minutes < 0 || minutes > 60 {
			http.Error(w, "minutes must be 0..60", http.StatusBadRequest)
			return
		}
		if err := shield.SetAttack(time.Duration(minutes) * time.Minute); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(shield.Status())
	})
	mux.HandleFunc("POST /shutdown", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(r) {
			http.Error(w, "unauthorized", 401)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.(http.Flusher).Flush()
		cancel()
	})
	if operator != nil {
		mux.HandleFunc("POST /operator", func(w http.ResponseWriter, r *http.Request) {
			if !authorize(r) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if ctx.Err() != nil {
				http.Error(w, "World is stopping", http.StatusServiceUnavailable)
				return
			}
			data := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
			data.DisallowUnknownFields()
			var req operatorwire.Request
			if err := data.Decode(&req); err != nil || data.Decode(new(any)) != io.EOF {
				http.Error(w, "invalid operator request", http.StatusBadRequest)
				return
			}
			res, err := operator(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(res)
		})
	}
	if os.Getenv("RAPTURE_PPROF") == "1" {
		protect := func(handler http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if !authorize(r) {
					http.Error(w, "unauthorized", 401)
					return
				}
				if seconds, _ := strconv.Atoi(r.URL.Query().Get("seconds")); seconds > 30 {
					http.Error(w, "profile duration exceeds 30 seconds", 400)
					return
				}
				handler(w, r)
			}
		}
		mux.HandleFunc("GET /debug/pprof/", protect(pprof.Index))
		mux.HandleFunc("GET /debug/pprof/profile", protect(pprof.Profile))
		mux.HandleFunc("GET /debug/pprof/trace", protect(pprof.Trace))
		mux.HandleFunc("GET /debug/pprof/cmdline", protect(pprof.Cmdline))
		mux.HandleFunc("GET /debug/pprof/symbol", protect(pprof.Symbol))
	}
	s.http = &http.Server{Handler: mux, ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 35 * time.Second, IdleTimeout: 10 * time.Second, MaxHeaderBytes: 8192}
	if err = os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		listener.Close()
		return nil, err
	}
	encoded, _ := json.Marshal(s.descriptor)
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".control-*")
	if err != nil {
		listener.Close()
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if err = protectDescriptor(tmp.Name()); err != nil {
		tmp.Close()
		listener.Close()
		return nil, fmt.Errorf("protect local control identity: %w", err)
	}
	if _, err = tmp.Write(encoded); err == nil {
		err = tmp.Close()
	} else {
		tmp.Close()
	}
	if err == nil {
		err = os.Rename(tmp.Name(), s.path)
	}
	if err != nil {
		listener.Close()
		return nil, err
	}
	go func() { defer close(s.done); s.http.Serve(listener) }()
	return s, nil
}
func (s *Server) Address() string { return s.listener.Addr().String() }
func (s *Server) Close() {
	s.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.http.Shutdown(ctx)
		s.http.Close()
		<-s.done
		// A newer process may have replaced the descriptor while this one stopped.
		if data, err := os.ReadFile(s.path); err == nil {
			var current descriptor
			if json.Unmarshal(data, &current) == nil && current.Token == s.descriptor.Token {
				os.Remove(s.path)
			}
		}
	})
}

func localRequest(ctx context.Context, root, service string, pid int, executable, method, path string, body []byte) (*http.Response, error) {
	data, err := os.ReadFile(filepath.Join(root, "runtime", "control", service+".json"))
	if err != nil {
		return nil, err
	}
	var d descriptor
	if err = json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	expected, _ := filepath.Abs(executable)
	if resolved, e := filepath.EvalSymlinks(expected); e == nil {
		expected = resolved
	}
	sameExecutable := filepath.Clean(expected) == filepath.Clean(d.Executable)
	if runtime.GOOS == "windows" {
		sameExecutable = strings.EqualFold(filepath.Clean(expected), filepath.Clean(d.Executable))
	}
	if d.PID != pid || !sameExecutable || !loopback(d.Address) || len(d.Token) != 64 {
		return nil, errors.New("local control identity mismatch")
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://"+d.Address+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+d.Token)
	if path == "/operator" {
		req.Header.Set("Content-Type", "application/json")
		return operatorClient.Do(req)
	}
	return localClient.Do(req)
}

var localClient = &http.Client{
	Timeout:       2 * time.Second,
	Transport:     &http.Transport{Proxy: nil, DisableKeepAlives: true},
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

// ReadStatus validates the descriptor against the running installation.
func ReadStatus(ctx context.Context, root, service string, pid int, executable string) (ServiceStatus, error) {
	resp, err := localRequest(ctx, root, service, pid, executable, http.MethodGet, "/status", nil)
	if err != nil {
		return ServiceStatus{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ServiceStatus{}, fmt.Errorf("status unavailable: HTTP %d", resp.StatusCode)
	}
	var status ServiceStatus
	err = json.NewDecoder(resp.Body).Decode(&status)
	return status, err
}

// RequestStop validates the PID and executable before sending the private token.
func RequestStop(ctx context.Context, root, service string, pid int, executable string) error {
	resp, err := localRequest(ctx, root, service, pid, executable, http.MethodPost, "/shutdown", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("shutdown refused: %d", resp.StatusCode)
	}
	return nil
}

// RequestShield changes only the verified local service's temporary admission mode.
// Zero minutes returns it to the ordinary always-on network policy.
func RequestShield(ctx context.Context, root, service string, pid int, executable string, minutes int) (netx.ShieldStatus, error) {
	if minutes < 0 || minutes > 60 {
		return netx.ShieldStatus{}, errors.New("shield minutes must be 0..60")
	}
	resp, err := localRequest(ctx, root, service, pid, executable, http.MethodPost, fmt.Sprintf("/shield?minutes=%d", minutes), nil)
	if err != nil {
		return netx.ShieldStatus{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return netx.ShieldStatus{}, fmt.Errorf("shield change refused: HTTP %d", resp.StatusCode)
	}
	var state netx.ShieldStatus
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return netx.ShieldStatus{}, err
	}
	return state, nil
}

var operatorClient = &http.Client{
	Timeout:       33 * time.Second,
	Transport:     localClient.Transport,
	CheckRedirect: localClient.CheckRedirect,
}

// RunOperator binds the request to this installation's exact World process.
func RunOperator(ctx context.Context, root string, pid int, executable string, req operatorwire.Request) (operatorwire.Response, error) {
	var result operatorwire.Response
	payload, err := json.Marshal(req)
	if err != nil {
		return result, err
	}
	defer clear(payload)
	resp, err := localRequest(ctx, root, "gx_world", pid, executable, http.MethodPost, "/operator", payload)
	if err != nil {
		return result, fmt.Errorf("World operator request failed. Refresh the current values before retrying: %w", err)
	}
	defer resp.Body.Close()
	data := io.LimitReader(resp.Body, 64<<10)
	if resp.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(data)
		return result, fmt.Errorf("World operator HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}
	return result, json.NewDecoder(data).Decode(&result)
}
