package control

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/controlplane"
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
)

// SetShield changes the two services as one operator action. An error after the
// first update attempts to restore its previous mode and reports any failure.
func (m *Manager) SetShield(ctx context.Context, minutes int) error {
	if minutes < 0 || minutes > 60 {
		return errors.New("shield minutes must be 0..60")
	}
	release, err := m.lock(ctx)
	if err != nil {
		return fmt.Errorf("acquire installation control: %w", err)
	}
	defer release()
	processes, err := m.processes()
	if err != nil {
		return err
	}
	previous := make(map[Service]netx.ShieldStatus, 2)
	for _, service := range Services() {
		if len(processes[service]) != 1 {
			return fmt.Errorf("%s must have exactly one running process", service.Label())
		}
		status := m.current(ctx, service, processes[service])
		if status.State != Ready || status.Shield.Mode == "" {
			return fmt.Errorf("%s shield status unavailable", service.Label())
		}
		previous[service] = status.Shield
	}
	var changed []Service
	for _, service := range Services() {
		if _, err := controlplane.RequestShield(ctx, m.root, string(service), processes[service][0], m.binary(service), minutes); err != nil {
			var rollback []error
			for _, prior := range changed {
				original := previous[prior]
				remaining := 0
				if original.Mode == netx.ShieldUnderAttack {
					remaining = int((time.Until(original.Until) + time.Minute - 1) / time.Minute)
					remaining = max(1, min(remaining, 60))
				}
				rollbackCtx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
				_, restoreErr := controlplane.RequestShield(rollbackCtx, m.root, string(prior), processes[prior][0], m.binary(prior), remaining)
				cancel()
				if restoreErr != nil {
					rollback = append(rollback, fmt.Errorf("restore %s: %w", prior.Label(), restoreErr))
				}
			}
			return errors.Join(append([]error{fmt.Errorf("change %s shield: %w", service.Label(), err)}, rollback...)...)
		}
		changed = append(changed, service)
	}
	return nil
}
