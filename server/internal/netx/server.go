package netx

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/Spelunkern/Shaiya-GX/server/internal/telemetry"
	"net"
	"net/netip"
	"sync"
	"time"
)

var admitted = telemetry.NumberMetric("gx_connections_accepted_total")
var rejected = telemetry.NumberMetric("gx_connections_rejected_total")
var connections = telemetry.NumberMetric("gx_connections_active")

// PeerIP returns a canonical peer address from the transport, including when
// the connection is wrapped in TLS. A scope is local to the network interface.
func PeerIP(conn net.Conn) netip.Addr {
	address, err := netip.ParseAddrPort(conn.RemoteAddr().String())
	if err != nil {
		return netip.Addr{}
	}
	return address.Addr().Unmap().WithZone("")
}

// Serve bounds connections globally and by source. Cancellation closes sockets
// and waits for handlers to finish persisting before the caller closes storage.
func Serve(ctx context.Context, listener net.Listener, capacity, perIP int, handle func(net.Conn), admission *RateLimiter, shield *Shield) error {
	starts := NewRateLimiter(4096, 120, time.Minute)
	if admission != nil {
		starts = admission
	}
	var mu sync.Mutex
	active := map[net.Conn]string{}
	sources := map[string]int{}
	var workers sync.WaitGroup
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			listener.Close()
			mu.Lock()
			for conn := range active {
				Abort(conn)
			}
			mu.Unlock()
		case <-done:
		}
	}()
	defer func() {
		listener.Close()
		mu.Lock()
		for conn := range active {
			Abort(conn)
		}
		mu.Unlock()
		workers.Wait()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		source, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			Abort(conn)
			continue
		}
		if shield != nil && !shield.Admit(source) {
			rejected.Add(1)
			Abort(conn)
			continue
		}
		if !starts.Allow(source) {
			rejected.Add(1)
			Abort(conn)
			continue
		}
		mu.Lock()
		sourceLimit := perIP
		if shield != nil {
			sourceLimit = shield.PerIP(perIP)
		}
		if ctx.Err() != nil || len(active) >= capacity || sources[source] >= sourceLimit {
			if shield != nil && sourceLimit < perIP && sources[source] >= sourceLimit && sources[source] < perIP {
				shield.RecordReject()
			}
			if shield != nil && len(active) >= capacity {
				shield.ObserveCapacity(source)
			}
			rejected.Add(1)
			mu.Unlock()
			Abort(conn)
			continue
		}
		admitted.Add(1)
		connections.Add(1)
		active[conn] = source
		sources[source]++
		mu.Unlock()
		workers.Go(func() {
			defer func() {
				_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
				_ = conn.Close()
				mu.Lock()
				delete(active, conn)
				connections.Add(-1)
				sources[source]--
				if sources[source] == 0 {
					delete(sources, source)
				}
				mu.Unlock()
			}()
			Tune(conn)
			if secure, ok := conn.(*tls.Conn); ok {
				handshakeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				err := secure.HandshakeContext(handshakeCtx)
				cancel()
				if err != nil {
					return
				}
			}
			handle(conn)
		})
	}
}

// RateLimiter has bounded memory and fails closed when its source budget fills.
// Idle entries are reclaimed opportunistically, without background goroutines.
type RateLimiter struct {
	mu       sync.Mutex
	entries  map[string]bucket
	capacity int
	burst    float64
	interval time.Duration
	sweep    time.Time
}
type bucket struct {
	tokens float64
	at     time.Time
}

func NewRateLimiter(capacity, burst int, interval time.Duration) *RateLimiter {
	return &RateLimiter{entries: map[string]bucket{}, capacity: capacity, burst: float64(burst), interval: interval}
}
func (l *RateLimiter) reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make(map[string]bucket)
	l.sweep = time.Time{}
}
func (l *RateLimiter) Allow(key string) bool                { return l.allow(key, time.Now()) }
func (l *RateLimiter) AllowN(key string, units int) bool    { return l.allowN(key, units, time.Now()) }
func (l *RateLimiter) allow(key string, now time.Time) bool { return l.allowN(key, 1, now) }
func (l *RateLimiter) allowN(key string, units int, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.Sub(l.sweep) >= l.interval {
		for k, b := range l.entries {
			if now.Sub(b.at) >= l.interval {
				delete(l.entries, k)
			}
		}
		l.sweep = now
	}
	b, ok := l.entries[key]
	if !ok {
		if len(l.entries) >= l.capacity {
			return false
		}
		b = bucket{l.burst, now}
	}
	b.tokens = min(l.burst, b.tokens+now.Sub(b.at).Seconds()*l.burst/l.interval.Seconds())
	b.at = now
	allowed := units > 0 && b.tokens >= float64(units)
	if allowed {
		b.tokens -= float64(units)
	}
	l.entries[key] = b
	return allowed
}
