package main

import (
	"strings"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

func TestLoginHandshakeBudgetRunsBeforeRSA(t *testing.T) {
	previous := handshakeByIP
	handshakeByIP = netx.NewRateLimiter(1, 1, time.Minute)
	t.Cleanup(func() { handshakeByIP = previous })
	remote := "192.0.2.1"
	if !handshakeByIP.Allow(remote) {
		t.Fatal("could not consume handshake budget")
	}
	w := packet.New(packet.LoginHandshake)
	w.U8(1)
	w.U8(0)
	c := &client{}
	if err := c.frame(w.MustFinish(), remote); err == nil || !strings.Contains(err.Error(), "handshake rate exceeded") {
		t.Fatalf("unexpected handshake result: %v", err)
	}
	if c.recv != nil {
		t.Fatal("RSA exchange ran after budget denial")
	}
}

func TestLoginHandshakeWorkIsBounded(t *testing.T) {
	previous := handshakeWork
	handshakeWork = netx.NewLimiter(1)
	t.Cleanup(func() { handshakeWork = previous })
	if !handshakeWork.Acquire() {
		t.Fatal("could not reserve handshake worker")
	}
	defer handshakeWork.Release()
	if _, err := deriveLoginMaterial([]byte{0}); err == nil || !strings.Contains(err.Error(), "capacity exceeded") {
		t.Fatalf("unbounded RSA work: %v", err)
	}
}
