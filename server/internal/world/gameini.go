package world

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/catalogfiles"
	"github.com/Spelunkern/Shaiya-GX/internal/maprules"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func gameDataDir() string { return filepath.Join(config.DataDir(), "game") }

type iniFile = config.GameINI

func readIni(path string) (iniFile, error) { return config.ReadGameINI(path) }

// numbers extracts every number in s ("(1, 577.98, 78.56, 1762.12)").
var numberRE = regexp.MustCompile(`-?\d+(\.\d+)?`)

func numbers(s string) []float64 {
	var out []float64
	for _, m := range numberRE.FindAllString(s, -1) {
		if v, err := strconv.ParseFloat(m, 64); err == nil {
			out = append(out, v)
		}
	}
	return out
}

// MapPos is a map and position from the original configs.
type MapPos struct {
	MapID   uint16
	X, Y, Z float32
}

// ZoneConfig holds the runtime rules for one map.
type ZoneConfig struct {
	Dungeon bool // MapType D (F is a field)
	// Rebirth destinations are configured independently for each faction.
	RebirthLight, RebirthFury *MapPos
	MinLevel, MaxLevel        uint16
	WeatherType               uint8
	WeatherFrequency          time.Duration
	MapCurses                 [3]uint16
	WarType                   string
	CreateType                string // instanced maps (R, H, P)
	// CreateTime (day of week, start hour, hours) entries, up to 7 as
	// CMap::LoadConfig stores them (the ranking battle reads two).
	Schedule []grbSlot
	// CreateMinUser (party members to enter a P map), EnterMaxUser (an
	// instance's capacity, 30 by default) and ExpireTime (minutes; an empty
	// instance goes ExpireTime x 60 s + 30 s after it first empties).
	CreateMinUser, EnterMaxUser int
	Expire                      time.Duration
}

// MapConfig holds the map definitions and race starts loaded at boot.
type MapConfig struct {
	// Start is the new-character position per race (Chars.Family): human,
	// elf, vail, deatheater.
	Start [4]MapPos
	Zones map[uint16]ZoneConfig
}

var mapConfig = MapConfig{
	Start: [4]MapPos{{1, 577.98, 78.56, 1762.12}, {1, 1678.92, 38.92, 1776.91},
		{2, 165.25, 118.50, 397.89}, {2, 1838.93, 129.33, 443.43}},
	Zones: map[uint16]ZoneConfig{},
}

func loadMapProperties(log *logx.Logger) error {
	c, err := catalogfiles.Read(filepath.Join(gameDataDir(), catalogfiles.Map), catalogfiles.ParseMap)
	if err != nil {
		return err
	}
	next, err := buildMapProperties(c)
	if err != nil {
		return err
	}
	mapConfig = next
	weatherStarted = time.Now()
	log.Printf("GX Server World loaded Map.ini zones=%d", len(next.Zones))
	return nil
}

func buildMapProperties(c maprules.Catalog) (MapConfig, error) {
	config := MapConfig{Zones: make(map[uint16]ZoneConfig, len(c.Maps))}
	point := func(r maprules.Row) *MapPos {
		if r[0] == nil {
			return nil
		}
		return &MapPos{uint16(r[0].(int64)), float32(r[1].(float64)), float32(r[2].(float64)), float32(r[3].(float64))}
	}
	for _, row := range c.Starts {
		config.Start[row[0].(int64)] = *point(row[1:])
	}
	for _, row := range c.Maps {
		minute := func(i int) time.Duration { return time.Duration(row[i].(int64)) * time.Minute }
		z := ZoneConfig{Dungeon: row[1] == "D", WarType: row[2].(string), CreateType: row[3].(string),
			MinLevel: uint16(row[4].(int64)), MaxLevel: uint16(row[5].(int64)),
			CreateMinUser: int(row[6].(int64)), EnterMaxUser: int(row[7].(int64)), Expire: minute(8) + 30*time.Second,
			WeatherType: uint8(row[9].(int64)), WeatherFrequency: minute(10),
			RebirthLight: point(row[17:21]), RebirthFury: point(row[21:25])}
		for i := range z.MapCurses {
			z.MapCurses[i] = uint16(row[11+i].(int64))
			if id := z.MapCurses[i]; id != 0 && mapCurseDef(db.MustDefs(), id) == nil {
				return MapConfig{}, fmt.Errorf("map %v curse %d must reference a level 1 applied skill", row[0], id)
			}
		}
		for i := 25; i < len(row); i += 3 {
			if row[i] != nil {
				z.Schedule = append(z.Schedule, grbSlot{Day: int(row[i].(int64)), Hour: int(row[i+1].(int64)), Hours: int(row[i+2].(int64))})
			}
		}
		config.Zones[uint16(row[0].(int64))] = z
	}
	return config, nil
}

var rewardRates atomic.Pointer[config.RewardRates]
var rewardSettingsMu sync.Mutex

func currentRewardRates() config.RewardRates {
	if rates := rewardRates.Load(); rates != nil {
		return *rates
	}
	return config.DefaultRewardRates()
}

// SaveRewardRates commits the file before publishing the new settings.
func SaveRewardRates(req operatorwire.Request, inherited map[string]string, log *logx.Logger) error {
	rewardSettingsMu.Lock()
	defer rewardSettingsMu.Unlock()
	active := currentRewardRates()
	fields := map[string]string{
		config.MobEXPKey: req.MobRate, config.QuestEXPKey: req.QuestRate, config.PvPKillKey: req.KillRate,
	}
	// Older controllers omit settings they cannot edit.
	for _, setting := range []struct {
		key   string
		input *string
		value uint32
	}{
		{config.MobDropKey, req.DropRate, active.MobDropPercent},
		{config.MobGoldKey, req.GoldRate, active.MobGoldPercent},
		{config.WorldDropKey, req.WorldDropRate, active.WorldDropPercent},
	} {
		if setting.input != nil {
			fields[setting.key] = *setting.input
		} else {
			fields[setting.key] = strconv.FormatUint(uint64(setting.value), 10)
		}
	}
	rates, err := config.SaveRewardRates(".", fields, inherited)
	if err != nil {
		return err
	}
	rewardRates.Store(&rates)
	log.Printf("GX Server World applied live reward rates: monster EXP %d%%, quest EXP %d%%, PvP rank kills %dx, monster loot %d%%, monster gold %d%%, global event loot %d%%",
		rates.MobEXPPercent, rates.QuestEXPPercent, rates.PvPKills, rates.MobDropPercent, rates.MobGoldPercent, rates.WorldDropPercent)
	return nil
}

// LoadGameConfig loads the original PS_Game ini files from data/game/.
func LoadGameConfig(log *logx.Logger) error {
	rates, err := config.ReadRewardRates(nil)
	if err != nil {
		return err
	}
	first, second, err := config.CombatDEXThresholds()
	if err != nil {
		return err
	}
	if err := loadMobAIStrings(); err != nil {
		return err
	}
	if err := loadMobAI(); err != nil {
		return err
	}
	log.Printf("GX Server World loaded authoritative Lua 5.5 mob AI catalog")
	if err := loadMapProperties(log); err != nil {
		return fmt.Errorf("map properties: %w", err)
	}
	if err := loadSafeZones(log); err != nil {
		return fmt.Errorf("safe zones: %w", err)
	}
	if err := loadEnchantRules(log); err != nil {
		return fmt.Errorf("enchant rules: %w", err)
	}
	if err := loadWorldDrops(log); err != nil {
		return fmt.Errorf("world drops: %w", err)
	}
	buildDropCatalog()
	dropRules, err := config.ReadScopedDrops(".")
	if err != nil {
		return err
	}
	cappedDrops := false
	for _, rule := range dropRules {
		if err := validateScopedDrop(rule); err != nil {
			return fmt.Errorf("event drop rule %d: %w", rule.ID, err)
		}
		cappedDrops = cappedDrops || rule.DailyLimit > 0
	}
	if cappedDrops {
		if err := db.EnsureEventDropLimits(); err != nil {
			return err
		}
	}
	scopedDropRules.Store(&dropRules)
	if err := loadItemBoxes(log); err != nil {
		return fmt.Errorf("item boxes: %w", err)
	}
	if err := loadNpcQuest(log); err != nil {
		return fmt.Errorf("NPC and quest definitions: %w", err)
	}
	if err := loadTeleportRunes(); err != nil {
		return fmt.Errorf("teleport runes: %w", err)
	}
	if err := loadServerNotices(log); err != nil {
		return fmt.Errorf("server notices: %w", err)
	}
	if err := loadKillStatus(log); err != nil {
		return fmt.Errorf("goddess blessing: %w", err)
	}
	overrides, err := config.ReadMapRewards(".")
	if err != nil {
		return err
	}
	events, err := config.ReadRewardEvents(".")
	if err != nil {
		return err
	}
	tuning, err := config.ReadMapTuning(".")
	if err != nil {
		return err
	}
	for _, policy := range tuning {
		if policy.ReentryMinutes != nil {
			if err := db.EnsureDungeonReentry(); err != nil {
				return err
			}
			dungeonReentryPrepared.Store(true)
			break
		}
	}
	economy, err := config.ReadEconomy(".")
	if err != nil {
		return err
	}
	economySettings.Store(&economyState{Economy: economy, revision: 1})
	mapTuning.Store(&tuning)
	currentRewardEvent.Store(events.Scheduled)
	mapRewardOverrides.Store(&overrides)
	combatDEXFirst, combatDEXSecond = first, second
	rewardRates.Store(&rates)
	log.Printf("GX Server World reward rates: monster EXP %d%%, quest EXP %d%%, PvP rank kills %dx, monster loot %d%%, monster gold %d%%, global event loot %d%%",
		rates.MobEXPPercent, rates.QuestEXPPercent, rates.PvPKills, rates.MobDropPercent, rates.MobGoldPercent, rates.WorldDropPercent)
	log.Printf("GX Server World combat DEX thresholds: %d, %d", first, second)
	return nil
}

// rebirthPosition is where a character of faction revives on mapID, per
// CMap::LoadConfig: Light and Fury have independent destinations.
func rebirthPosition(mapID uint16, faction uint8) (MapPos, bool) {
	z, found := mapConfig.Zones[realMap(mapID)]
	if !found {
		return MapPos{}, false
	}
	switch {
	case faction == FactionLight && z.RebirthLight != nil:
		return *z.RebirthLight, true
	case faction == FactionDark && z.RebirthFury != nil:
		return *z.RebirthFury, true
	}
	return MapPos{}, false
}

// bindPosition is CMap::GetBindPos 0x004295a0: the configured faction rebirth point
// when its map exists and the point lies inside it (0.9 above), else a
// random point of the nearest svmap rebirth group of this map, else the
// current position (zeroed if outside the map).
func bindPosition(mapID uint16, faction uint8, x, y, z float32) (uint16, float32, float32, float32) {
	m, x, y, z, _ := bindPositionFound(mapID, faction, x, y, z)
	return m, x, y, z
}

// bindPositionFound is bindPosition with GetBindPos's result: false when
// neither a configured point nor a rebirth group applies.
func bindPositionFound(mapID uint16, faction uint8, x, y, z float32) (uint16, float32, float32, float32, bool) {
	if p, ok := rebirthPosition(mapID, faction); ok {
		if size, ok := mapSize(p.MapID); ok && p.X >= 0 && p.X < size && p.Z >= 0 && p.Z < size {
			return p.MapID, p.X, p.Y + 0.9, p.Z, true
		}
	}
	if area, ok := FindNearestSpawnArea(mapID, faction, x, z); ok {
		x, y, z := area.RandomPoint()
		return mapID, x, y, z, true
	}
	if size, ok := mapSize(mapID); ok && (x < 0 || x >= size || z < 0 || z >= size) {
		x, z = 0, 0
	}
	return mapID, x, y, z, false
}
