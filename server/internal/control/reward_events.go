package control

import (
	"context"
	"slices"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func (m *Manager) RewardEvents() (config.RewardEvents, error) {
	return config.ReadRewardEvents(m.root)
}

func (m *Manager) SaveRewardEvent(ctx context.Context, edit config.RewardEventEdit) error {
	release, err := m.lock(ctx)
	if err != nil {
		return err
	}
	defer release()
	processes, err := m.processes()
	if err != nil {
		return err
	}
	if len(processes[World]) > 0 {
		_, err := m.worldOperator(ctx, processes[World], operatorwire.Request{Action: "reward-event", RewardEvent: &edit})
		return err
	}
	ids, err := m.RewardMaps()
	if err != nil {
		return err
	}
	_, err = config.SaveRewardEvent(m.root, edit, time.Now(), func(id uint16) bool { return slices.Contains(ids, id) })
	return err
}
