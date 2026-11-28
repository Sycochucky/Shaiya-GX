package netx

import (
	"fmt"
	"sync"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/telemetry"
)

const (
	ShieldStandard    = "standard"
	ShieldUnderAttack = "under_attack"
	attackDuration    = 10 * time.Minute
	attackPerIP       = 16
)

var shieldActive = telemetry.NumberMetric("gx_shield_under_attack")
var shieldRejected = telemetry.NumberMetric("gx_shield_rejected_total")
var shieldAutomatic = telemetry.NumberMetric("gx_shield_automatic_activations_total")

// ShieldStatus is the current, time-bounded admission state of one service.
type ShieldStatus struct {
	Mode                 string    `json:"mode"`
	Until                time.Time `json:"until,omitempty"`
	Automatic            bool      `json:"automatic,omitempty"`
	RejectedTotal        int64     `json:"rejected_total"`
	AutomaticActivations int64     `json:"automatic_activations"`
}

// Shield adds a stricter source budget during an attack. It never replaces the
// ordinary global and per-source limits, and never terminates admitted players.
type Shield struct {
	mu              sync.Mutex
	until           time.Time
	automatic       bool
	windowStart     time.Time
	overloads       int
	overloadSources map[string]struct{}
	attackStarts    *RateLimiter
}

func NewShield() *Shield {
	return &Shield{attackStarts: NewRateLimiter(4096, 30, time.Minute)}
}

func (s *Shield) status(now time.Time) ShieldStatus {
	if !s.until.IsZero() && !now.Before(s.until) {
		s.until = time.Time{}
		s.automatic = false
		shieldActive.Set(0)
	}
	if s.until.IsZero() {
		return ShieldStatus{Mode: ShieldStandard, RejectedTotal: shieldRejected.Value(), AutomaticActivations: shieldAutomatic.Value()}
	}
	return ShieldStatus{Mode: ShieldUnderAttack, Until: s.until, Automatic: s.automatic, RejectedTotal: shieldRejected.Value(), AutomaticActivations: shieldAutomatic.Value()}
}

func (s *Shield) Status() ShieldStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status(time.Now())
}

// SetAttack accepts only short-lived operator requests; zero returns to the
// ordinary policy. The caller must authenticate before reaching this method.
func (s *Shield) SetAttack(duration time.Duration) error {
	if duration != 0 && (duration < time.Minute || duration > time.Hour) {
		return fmt.Errorf("shield duration must be 1..60 minutes")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.until = time.Time{}
	s.automatic = false
	s.windowStart = time.Time{}
	s.overloads = 0
	s.overloadSources = nil
	s.attackStarts.reset()
	if duration != 0 {
		s.until = time.Now().Add(duration)
		shieldActive.Set(1)
	} else {
		shieldActive.Set(0)
	}
	return nil
}

// Admit is called before TLS and game handshakes. The stricter budget is
// charged only while Under Attack is active.
func (s *Shield) Admit(source string) bool {
	if s.Status().Mode != ShieldUnderAttack {
		return true
	}
	if !s.attackStarts.Allow(source) {
		shieldRejected.Add(1)
		return false
	}
	return true
}

func (s *Shield) PerIP(ordinary int) int {
	if s.Status().Mode == ShieldUnderAttack {
		return min(ordinary, attackPerIP)
	}
	return ordinary
}

func (s *Shield) RecordReject() { shieldRejected.Add(1) }

// ObserveCapacity activates a temporary stricter mode only when a service is
// already full and rejects sustained new arrivals. It cannot trigger from a
// single burst below the capacity limit.
func (s *Shield) ObserveCapacity(source string) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status(now).Mode == ShieldUnderAttack {
		return
	}
	if now.Sub(s.windowStart) >= 10*time.Second {
		s.windowStart = now
		s.overloads = 0
		s.overloadSources = make(map[string]struct{}, 8)
	}
	s.overloads++
	if len(s.overloadSources) < 8 {
		s.overloadSources[source] = struct{}{}
	}
	if s.overloads >= 64 && len(s.overloadSources) >= 4 {
		s.until = now.Add(attackDuration)
		s.automatic = true
		s.overloads = 0
		s.overloadSources = nil
		s.attackStarts.reset()
		shieldActive.Set(1)
		shieldAutomatic.Add(1)
	}
}
