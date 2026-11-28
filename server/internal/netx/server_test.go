package netx

import (
	"context"
	"errors"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimiterBoundsAndRefill(t *testing.T) {
	l := NewRateLimiter(2, 2, time.Minute)
	now := time.Now()
	if !l.allow("a", now) || !l.allow("a", now) || l.allow("a", now) {
		t.Fatal("burst")
	}
	if !l.allow("b", now) || l.allow("c", now) {
		t.Fatal("memory bound")
	}
	if !l.allow("a", now.Add(30*time.Second)) {
		t.Fatal("refill")
	}
	if !l.allow("c", now.Add(2*time.Minute)) {
		t.Fatal("idle reclamation")
	}
}

func TestServeShieldChangesAdmissionWithoutDroppingPlayers(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	shield := NewShield()
	shield.attackStarts = NewRateLimiter(1, 2, time.Hour)
	var first atomic.Bool
	firstReady := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		finished <- Serve(ctx, listener, 64, 64, func(conn net.Conn) {
			if first.CompareAndSwap(false, true) {
				close(firstReady)
				var b [1]byte
				if _, err := io.ReadFull(conn, b[:]); err == nil {
					conn.Write(b[:])
				}
				return
			}
			conn.Write([]byte{1})
		}, nil, shield)
	}()
	t.Cleanup(func() {
		shield.SetAttack(0)
		cancel()
		listener.Close()
		select {
		case err := <-finished:
			if err != nil {
				t.Errorf("shield server shutdown: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("shield server did not stop")
		}
	})
	existing, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer existing.Close()
	select {
	case <-firstReady:
	case <-time.After(3 * time.Second):
		t.Fatal("first connection not admitted")
	}
	if err := shield.SetAttack(30 * time.Minute); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		var b [1]byte
		_, err = conn.Read(b[:])
		conn.Close()
		if i < 2 && (err != nil || b[0] != 1) {
			t.Fatalf("admission %d: %v, %v", i, b, err)
		}
		if i == 2 && (err == nil || isTimeout(err)) {
			t.Fatalf("third connection was not rejected promptly: %v", err)
		}
	}
	existing.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := existing.Write([]byte{7}); err != nil {
		t.Fatal(err)
	}
	var echoed [1]byte
	if _, err := io.ReadFull(existing, echoed[:]); err != nil || echoed[0] != 7 {
		t.Fatalf("existing player disconnected: %v, %v", echoed, err)
	}
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
func TestServeCancellationDrainsHandler(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		finished <- Serve(ctx, listener, 2, 1, func(conn net.Conn) { close(entered); var b [1]byte; conn.Read(b[:]) }, nil, nil)
	}()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("handler not entered")
	}
	second, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	second.SetReadDeadline(time.Now().Add(time.Second))
	var b [1]byte
	if _, err := second.Read(b[:]); err == nil {
		t.Fatal("source cap not enforced")
	} else if e, ok := err.(net.Error); ok && e.Timeout() {
		t.Fatal("connection was not rejected")
	}
	cancel()
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown did not drain")
	}
}
func FuzzFrames(f *testing.F) {
	f.Add([]byte{4, 0, 1, 0})
	f.Add([]byte{255, 255, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		var frames Frames
		for len(data) > 0 {
			n := min(len(data), 37)
			list, ok := frames.Push(data[:n])
			data = data[n:]
			for _, frame := range list {
				if len(frame) < 4 || len(frame) > 65535 {
					t.Fatal("invalid frame")
				}
			}
			if !ok {
				return
			}
		}
	})
}

func TestRateLimiterChargesUnits(t *testing.T) {
	l := NewRateLimiter(1, 100, time.Second)
	now := time.Now()
	if !l.allowN("", 80, now) || l.allowN("", 21, now) || !l.allowN("", 20, now) {
		t.Fatal("unit budget")
	}
	if l.allowN("", 0, now) || l.allowN("", -1, now) {
		t.Fatal("invalid charge accepted")
	}
	if !l.allowN("", 50, now.Add(500*time.Millisecond)) {
		t.Fatal("unit refill")
	}
}
