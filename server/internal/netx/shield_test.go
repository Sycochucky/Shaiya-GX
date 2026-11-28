package netx

import (
	"testing"
	"time"
)

func TestShieldTemporaryModeAndAdmission(t *testing.T) {
	s := NewShield()
	if s.Status().Mode != ShieldStandard || !s.Admit("192.0.2.1") {
		t.Fatal("standard admission")
	}
	if err := s.SetAttack(30 * time.Second); err == nil {
		t.Fatal("accepted short attack mode")
	}
	if err := s.SetAttack(30 * time.Minute); err != nil {
		t.Fatal(err)
	}
	if s.PerIP(32) != 16 || s.PerIP(8) != 8 || s.Status().Automatic {
		t.Fatal("attack policy")
	}
	for i := 0; i < 30; i++ {
		if !s.Admit("192.0.2.1") {
			t.Fatal("source denied before attack budget")
		}
	}
	if s.Admit("192.0.2.1") || !s.Admit("192.0.2.2") {
		t.Fatal("attack budget did not isolate source")
	}
	if err := s.SetAttack(0); err != nil {
		t.Fatal(err)
	}
	if err := s.SetAttack(time.Minute); err != nil || !s.Admit("192.0.2.1") {
		t.Fatalf("new attack window retained the previous source budget: %v", err)
	}
	s.mu.Lock()
	s.until = time.Now().Add(-time.Second)
	s.mu.Unlock()
	if s.Status().Mode != ShieldStandard || s.PerIP(32) != 32 {
		t.Fatal("attack mode did not expire")
	}
}

func TestShieldAutomaticOverloadNeedsSustainedCapacity(t *testing.T) {
	s := NewShield()
	for i := 0; i < 64; i++ {
		s.ObserveCapacity("192.0.2.1")
	}
	if s.Status().Mode != ShieldStandard {
		t.Fatal("one source activated global attack mode")
	}
	for _, source := range []string{"192.0.2.2", "192.0.2.3", "192.0.2.4"} {
		s.ObserveCapacity(source)
	}
	if status := s.Status(); status.Mode != ShieldUnderAttack || !status.Automatic {
		t.Fatalf("automatic status = %+v", status)
	}
	if err := s.SetAttack(0); err != nil || s.Status().Mode != ShieldStandard {
		t.Fatalf("manual reset: %v", err)
	}
}
