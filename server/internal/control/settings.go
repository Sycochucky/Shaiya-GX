package control

import (
	"context"
	"os"
	"strconv"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

// RewardSettings returns the editable local values and tracks inherited overrides.
func (m *Manager) RewardSettings() (map[string]string, map[string]bool, error) {
	values, err := config.ReadRuntimeEnv(m.root)
	if err != nil {
		return nil, nil, err
	}
	defaults := config.DefaultRewardRates()
	fields := []struct {
		key      string
		fallback uint32
	}{
		{config.MobEXPKey, defaults.MobEXPPercent},
		{config.QuestEXPKey, defaults.QuestEXPPercent},
		{config.PvPKillKey, defaults.PvPKills},
		{config.MobDropKey, defaults.MobDropPercent},
		{config.MobGoldKey, defaults.MobGoldPercent},
		{config.WorldDropKey, defaults.WorldDropPercent},
	}
	effective := make(map[string]string, len(fields))
	override := make(map[string]bool, len(fields))
	for _, field := range fields {
		key := field.key
		_, override[key] = os.LookupEnv(key)
		effective[key] = config.RuntimeValue(values, key)
		if effective[key] == "" {
			effective[key] = strconv.FormatUint(uint64(field.fallback), 10)
		}
	}
	return effective, override, nil
}

func (m *Manager) SaveRewardSettings(ctx context.Context, values map[string]string) error {
	_, err := config.ParseRewardRates(values)
	if err != nil {
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
	if len(processes[World]) > 0 {
		drops, gold, eventDrops := values[config.MobDropKey], values[config.MobGoldKey], values[config.WorldDropKey]
		_, err := m.worldOperator(ctx, processes[World], operatorwire.Request{
			Action: "reward-rates", MobRate: values[config.MobEXPKey], QuestRate: values[config.QuestEXPKey],
			KillRate: values[config.PvPKillKey], DropRate: &drops, GoldRate: &gold, WorldDropRate: &eventDrops,
		})
		return err
	}
	_, err = config.SaveRewardRates(m.root, values, config.InheritedRewardRates())
	return err
}
