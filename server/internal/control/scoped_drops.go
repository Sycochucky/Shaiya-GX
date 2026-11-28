package control

import (
	"context"
	"errors"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func (m *Manager) ScopedDropRules() ([]config.ScopedDropRule, error) {
	return config.ReadScopedDrops(m.root)
}

func (m *Manager) SaveScopedDrop(ctx context.Context, edit config.ScopedDropEdit) error {
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
		return errors.New("start World before editing event drop rules")
	}
	_, err = m.worldOperator(ctx, processes[World], operatorwire.Request{Action: "scoped-drop", ScopedDrop: &edit})
	return err
}
