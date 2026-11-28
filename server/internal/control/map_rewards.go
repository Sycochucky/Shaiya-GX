package control

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Spelunkern/Shaiya-GX/internal/catalogfiles"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func (m *Manager) RewardMaps() ([]uint16, error) {
	values, err := config.ReadRuntimeEnv(m.root)
	if err != nil {
		return nil, err
	}
	data := config.RuntimeValue(values, "RAPTURE_DATA_DIR")
	if data == "" {
		data = "data"
	}
	if !filepath.IsAbs(data) {
		data = filepath.Join(m.root, data)
	}
	catalog, err := catalogfiles.Read(filepath.Join(data, "game", catalogfiles.Map), catalogfiles.ParseMap)
	if err != nil {
		return nil, err
	}
	ids := make([]uint16, 0, len(catalog.Maps))
	for _, row := range catalog.Maps {
		id := uint16(row[0].(int64))
		if info, err := os.Stat(filepath.Join(data, "maps", fmt.Sprintf("%d.svmap", id))); err == nil && !info.IsDir() {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (m *Manager) MapRewards() (config.MapRewards, error) {
	return config.ReadMapRewards(m.root)
}

func (m *Manager) SaveMapReward(ctx context.Context, id uint16, override config.MapRewardOverrides) error {
	if err := override.Validate(); err != nil {
		return err
	}
	release, err := m.lock(ctx)
	if err != nil {
		return err
	}
	defer release()
	ids, err := m.RewardMaps()
	if err != nil {
		return err
	}
	known := false
	for _, mapID := range ids {
		if mapID == id {
			known = true
			break
		}
	}
	if !known {
		return fmt.Errorf("map %d is unavailable in this installation", id)
	}
	processes, err := m.processes()
	if err != nil {
		return err
	}
	if len(processes[World]) > 0 {
		_, err := m.worldOperator(ctx, processes[World], operatorwire.Request{
			Action: "map-reward", MapID: id, MapReward: &override,
		})
		return err
	}
	_, err = config.SaveMapReward(m.root, id, override)
	return err
}
