package world

import (
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

type activityTotals struct {
	kills, pvp, drops, pickups, gold atomic.Uint64
}

var mapActivityCounters sync.Map

func activityForMap(id uint16) *activityTotals {
	if existing, ok := mapActivityCounters.Load(id); ok {
		return existing.(*activityTotals)
	}
	actual, _ := mapActivityCounters.LoadOrStore(id, &activityTotals{})
	return actual.(*activityTotals)
}

func MapActivity(id uint16) (operatorwire.MapActivity, error) {
	if !hasGrid(id) {
		return operatorwire.MapActivity{}, fmt.Errorf("map %d is unavailable", id)
	}
	totals := activityForMap(id)
	out := operatorwire.MapActivity{
		Instances:    len(instanceZones(id)),
		MonsterKills: totals.kills.Load(), PvPKills: totals.pvp.Load(), Drops: totals.drops.Load(),
		ItemsCollected: totals.pickups.Load(), GoldGained: totals.gold.Load(),
	}
	out.BossRespawns = mapBossRespawns(id)
	for _, p := range AllPlayers() {
		if realMap(p.MapID()) == id {
			out.Players++
		}
	}
	return out, nil
}

func mapBossRespawns(id uint16) []operatorwire.BossRespawn {
	windows := tuningForMap(id).BossWindows
	if len(windows) == 0 {
		return nil
	}
	current := make(map[uint16]operatorwire.BossRespawn, len(windows))
	for mobID := range windows {
		current[mobID] = operatorwire.BossRespawn{MonsterID: mobID}
	}
	now := time.Now()
	for _, zone := range append(instanceZones(id), id) {
		s := shard(zone)
		s.mu.Lock()
		for i := range s.entities.Mobs {
			mob := &s.entities.Mobs[i]
			window, ok := windows[mob.MobID]
			if !ok || mob.aiSummoned || mob.event != nil {
				continue
			}
			state := current[mob.MobID]
			state.Alive = state.Alive || mob.CurrentHP > 0
			if mob.CurrentHP == 0 {
				at := mob.respawnAt
				if mob.group != nil && mob.group.alive < mob.group.count {
					at = mob.group.respawnAt
				}
				if !at.IsZero() {
					if at.Before(now) {
						at = now
					}
					at = window.Next(at.Local())
					if state.Next.IsZero() || at.Before(state.Next) {
						state.Next = at
					}
				}
			}
			current[mob.MobID] = state
		}
		s.mu.Unlock()
	}
	result := make([]operatorwire.BossRespawn, 0, len(current))
	for _, state := range current {
		result = append(result, state)
	}
	slices.SortFunc(result, func(a, b operatorwire.BossRespawn) int { return int(a.MonsterID) - int(b.MonsterID) })
	return result
}
