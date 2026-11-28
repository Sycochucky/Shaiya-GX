package control

import (
	"context"
	"errors"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func (m *Manager) MapTuning() (map[uint16]config.MapTuning, error) {
	return config.ReadMapTuning(m.root)
}

func (m *Manager) SaveMapTuning(ctx context.Context, id uint16, value config.MapTuning) error {
	if err := value.Validate(); err != nil {
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
		return errors.New("start World before saving map controls")
	}
	_, err = m.worldOperator(ctx, processes[World], operatorwire.Request{Action: "map-tuning", MapID: id, MapTuning: &value})
	return err
}

func (m *Manager) MapActivity(ctx context.Context, id uint16) (operatorwire.MapActivity, error) {
	processes, err := m.processes()
	if err != nil {
		return operatorwire.MapActivity{}, err
	}
	if len(processes[World]) == 0 {
		return operatorwire.MapActivity{}, errors.New("World is stopped")
	}
	result, err := m.worldOperator(ctx, processes[World], operatorwire.Request{Action: "map-activity", MapID: id})
	if err != nil {
		return operatorwire.MapActivity{}, err
	}
	if result.MapActivity == nil {
		return operatorwire.MapActivity{}, errors.New("map activity is unavailable")
	}
	return *result.MapActivity, nil
}
