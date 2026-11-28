package session

import (
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/netx"
)

// World is the game-server connection. Packets delivers every decoded
// payload in order; it is closed when the connection ends, after which Err
// reports why.
type World struct {
	conn         *netx.Conn
	Packets      <-chan []byte
	err          error
	done         chan struct{}
	stop         chan struct{}
	stopOnce     sync.Once
	sendMu       sync.Mutex
	outbound     chan []byte
	writerDone   chan struct{}
	pendingBytes int
	writeErr     error
}

// NewWorld wraps an established world connection and starts its reader and writer.
func NewWorld(c *netx.Conn) *World { return startWorld(c) }

func startWorld(c *netx.Conn) *World {
	ch := make(chan []byte, 256)
	w := &World{conn: c, Packets: ch, done: make(chan struct{}), stop: make(chan struct{}), outbound: make(chan []byte, 128), writerDone: make(chan struct{})}
	go w.writeWorld(w.outbound)
	go func() {
		defer close(ch)
		defer w.Close()
		defer close(w.done)
		for {
			p, err := c.Read()
			if err != nil {
				w.err = err
				return
			}
			select {
			case ch <- p:
			case <-w.stop:
				return
			}
		}
	}()
	return w
}

// Send queues a payload in protocol order without waiting for the socket.
// Overflow closes the connection so gameplay commands cannot disappear silently.
func (w *World) Send(p []byte) error {
	if p == nil {
		return nil
	}
	if len(p) < 2 || len(p) > netx.MaxFrame-2 {
		return fmt.Errorf("invalid world payload length %d", len(p))
	}
	w.sendMu.Lock()
	select {
	case <-w.stop:
		w.sendMu.Unlock()
		return net.ErrClosed
	default:
	}
	if w.outbound == nil {
		w.sendMu.Unlock()
		return net.ErrClosed
	}
	if len(w.outbound) < cap(w.outbound) && w.pendingBytes+len(p) <= 1<<20 {
		w.outbound <- slices.Clone(p)
		w.pendingBytes += len(p)
		w.sendMu.Unlock()
		return nil
	}
	w.sendMu.Unlock()
	w.Close()
	return errors.New("outbound world queue is full")
}

func (w *World) writeWorld(outbound <-chan []byte) {
	defer close(w.writerDone)
	for {
		select {
		case <-w.stop:
			return
		case p, ok := <-outbound:
			if !ok {
				w.writeErr = w.conn.CloseWrite()
				return
			}
			err := w.conn.Write(p)
			w.sendMu.Lock()
			w.pendingBytes -= len(p)
			w.sendMu.Unlock()
			if err != nil {
				w.writeErr = err
				w.Close()
				return
			}
		}
	}
}

func (w *World) Close() error {
	if w.conn == nil {
		return nil
	}
	var err error
	w.stopOnce.Do(func() {
		w.sendMu.Lock()
		close(w.stop)
		if w.outbound != nil {
			close(w.outbound)
			w.outbound = nil
		}
		w.sendMu.Unlock()
		err = w.conn.Abort()
	})
	return err
}

// Shutdown waits for the server to persist and release the account before
// closing its socket (Rapture Session.close, then the connection's close).
// Drain packets so the reader cannot remain blocked on its output channel.
func (w *World) Shutdown(timeout time.Duration) error {
	defer w.Close()
	if err := w.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	w.sendMu.Lock()
	if w.outbound != nil {
		close(w.outbound)
		w.outbound = nil
	}
	w.sendMu.Unlock()
	writerDone := w.writerDone
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case _, ok := <-w.Packets:
			if !ok {
				if err := w.Err(); err != nil && !errors.Is(err, io.EOF) {
					return err
				}
				return nil
			}
		case <-writerDone:
			if w.writeErr != nil {
				return w.writeErr
			}
			writerDone = nil
		case <-timer.C:
			return fmt.Errorf("world shutdown timed out after %s", timeout)
		}
	}
}

// Err is valid after Packets is closed.
func (w *World) Err() error { <-w.done; return w.err }
