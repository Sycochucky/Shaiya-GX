package world

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

var mapRewardOverrides atomic.Pointer[config.MapRewards]

func rewardRatesForMap(zone uint16) config.RewardRates {
	base := currentRewardRates()
	event := currentRewardEvent.Load()
	eventActive := event.Running(time.Now())
	current := mapRewardOverrides.Load()
	if (current == nil || len(*current) == 0) && !eventActive {
		return base
	}
	id := realMap(zone)
	if current != nil {
		base = (*current)[id].Apply(base)
	}
	if eventActive && event.Preset.Affects(id) {
		base = event.Preset.Rates.Apply(base)
	}
	return base
}

func MapRewardRates(id uint16) (config.RewardRates, error) {
	if _, ok := mapConfig.Zones[id]; !ok {
		return config.RewardRates{}, fmt.Errorf("map %d is unavailable in World", id)
	}
	return rewardRatesForMap(id), nil
}

func SaveMapReward(req operatorwire.Request, log *logx.Logger) error {
	if req.MapReward == nil {
		return fmt.Errorf("missing reward settings for map %d", req.MapID)
	}
	if _, ok := mapConfig.Zones[req.MapID]; !ok {
		return fmt.Errorf("map %d is unavailable in World", req.MapID)
	}
	rewardSettingsMu.Lock()
	defer rewardSettingsMu.Unlock()
	override := *req.MapReward
	rates, err := config.SaveMapReward(".", req.MapID, override)
	if err != nil {
		return err
	}
	mapRewardOverrides.Store(&rates)
	log.Event("Applied map reward rates", "activity", "staff.settings", "map_id", req.MapID)
	return nil
}
