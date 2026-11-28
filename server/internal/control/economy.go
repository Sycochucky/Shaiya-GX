package control

import (
	"context"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func (m *Manager) Economy() (config.Economy, error) { return config.ReadEconomy(m.root) }

func (m *Manager) SaveEconomy(ctx context.Context, e config.Economy) error {
	if err := e.Validate(); err != nil {
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
	if len(processes[World]) == 0 {
		return config.SaveEconomy(m.root, e)
	}
	_, err = m.worldOperator(ctx, processes[World], operatorwire.Request{Action: "economy-settings", Economy: &e})
	return err
}
