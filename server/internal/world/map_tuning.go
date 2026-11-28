package world

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

var mapTuning atomic.Pointer[map[uint16]config.MapTuning]
var dungeonReentryPrepared atomic.Bool

func tuningForMap(zone uint16) config.MapTuning {
	if current := mapTuning.Load(); current != nil && len(*current) > 0 {
		return (*current)[realMap(zone)]
	}
	return config.MapTuning{}
}

func SaveMapTuning(req operatorwire.Request, log *logx.Logger) error {
	if req.MapTuning == nil {
		return fmt.Errorf("missing controls for map %d", req.MapID)
	}
	base, ok := mapConfig.Zones[req.MapID]
	if !ok || !hasGrid(req.MapID) {
		return fmt.Errorf("map %d is unavailable in World", req.MapID)
	}
	tuning := *req.MapTuning
	if err := tuning.Validate(); err != nil {
		return err
	}
	if (tuning.Drain || tuning.InstanceMinutes != nil || tuning.ReentryMinutes != nil) && base.CreateType != partyCreateType {
		return fmt.Errorf("map %d is not a party instance", req.MapID)
	}
	if tuning.AdmissionOverride() {
		for _, start := range mapConfig.Start {
			if start.MapID == req.MapID {
				return fmt.Errorf("map %d is a starting map and must remain accessible", req.MapID)
			}
		}
	}
	minLevel, maxLevel := base.MinLevel, base.MaxLevel
	if tuning.MinLevel != nil {
		minLevel = max(minLevel, *tuning.MinLevel)
	}
	if tuning.MaxLevel != nil {
		maxLevel = mapMaxLevel(maxLevel, *tuning.MaxLevel)
	}
	if maxLevel != 0 && minLevel > maxLevel {
		return fmt.Errorf("map %d effective minimum level is above its maximum", req.MapID)
	}
	for mob := range tuning.BossWindows {
		if db.MustDefs().Mobs[mob] == nil {
			return fmt.Errorf("unknown monster %d", mob)
		}
	}
	for mob := range tuning.MobRespawnSeconds {
		if db.MustDefs().Mobs[mob] == nil {
			return fmt.Errorf("unknown monster %d", mob)
		}
	}
	rewardSettingsMu.Lock()
	defer rewardSettingsMu.Unlock()
	if tuning.ReentryMinutes != nil && !dungeonReentryPrepared.Load() {
		if err := db.EnsureDungeonReentry(); err != nil {
			return err
		}
		dungeonReentryPrepared.Store(true)
	}
	all, err := config.SaveMapTuning(".", req.MapID, tuning)
	if err != nil {
		return err
	}
	mapTuning.Store(&all)
	log.Event("Applied map controls", "activity", "staff.settings", "map_id", req.MapID)
	return nil
}

func (s *Session) mapAdmissionAllowed(mapID, level uint16) bool {
	if s.isAdmin {
		return true
	}
	tuning := tuningForMap(mapID)
	if tuning.Closed || tuning.Faction != nil && *tuning.Faction != s.faction {
		return false
	}
	actual := mapConfig.Zones[realMap(mapID)]
	if tuning.MinLevel != nil {
		actual.MinLevel = max(actual.MinLevel, *tuning.MinLevel)
	}
	if tuning.MaxLevel != nil {
		actual.MaxLevel = mapMaxLevel(actual.MaxLevel, *tuning.MaxLevel)
	}
	if (actual.MinLevel != 0 && level < actual.MinLevel) || (actual.MaxLevel != 0 && level > actual.MaxLevel) {
		return false
	}
	if tuning.MinimumParty != nil && *tuning.MinimumParty > 0 {
		_, members := partyOfPlayer(s.player)
		return members >= int(*tuning.MinimumParty)
	}
	return true
}

// Configured times affect subsequently started timers. Summoned mobs and
// existing non-respawning mobs retain their authored behavior.
func (m *MapMob) configuredRespawnDelay(zone uint16) time.Duration {
	base := m.respawnDelay
	if m.aiSummoned || m.event != nil || (m.group == nil && base == 0) {
		return base
	}
	tuning := tuningForMap(zone)
	if seconds, ok := tuning.MobRespawnSeconds[m.MobID]; ok {
		return time.Duration(seconds) * time.Second
	}
	if tuning.RespawnPercent == nil {
		return base
	}
	percent := time.Duration(*tuning.RespawnPercent)
	if base >= (7*24*time.Hour)*100/percent {
		return 7 * 24 * time.Hour
	}
	return base * percent / 100
}

func mapMaxLevel(native, custom uint16) uint16 {
	if native == 0 {
		return custom
	}
	if custom == 0 {
		return native
	}
	return min(native, custom)
}

// PvP windows govern new map combat. Native faction eligibility, safe
// zones and voluntary player or guild duels keep their existing checks.
func (s *Session) pvpWindowAllows(target *ConnectedPlayer) bool {
	window := tuningForMap(s.char.MapID).PvPWindow
	return window == nil || window.Contains(time.Now()) ||
		(s.player != nil && (duelFoes(s.player, target) || guildDuelFoes(s.player, target)))
}

func pollPvPWindows(previous map[uint16]bool, now time.Time, log *logx.Logger) {
	current := mapTuning.Load()
	if current == nil {
		return
	}
	for id, tuning := range *current {
		if tuning.PvPWindow == nil || !tuning.PvPAnnouncements {
			delete(previous, id)
			continue
		}
		open := tuning.PvPWindow.Contains(now)
		old, known := previous[id]
		previous[id] = open
		if open == old || !known {
			continue
		}
		message := "PvP battle window closed."
		if open {
			message = "PvP battle window is open."
		}
		log.Event(message, "activity", "staff.settings", "map_id", id)
		sendRewardNotice(config.RewardPreset{Maps: []uint16{id}}, message)
	}
	for id := range previous {
		if _, ok := (*current)[id]; !ok {
			delete(previous, id)
		}
	}
}
