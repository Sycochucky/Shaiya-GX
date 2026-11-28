package world

import (
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
)

type dungeonRun struct {
	deadline time.Time
	reentry  time.Duration
}

func configuredDungeonRun(mapID uint16, now time.Time) dungeonRun {
	var run dungeonRun
	if mapConfig.Zones[mapID].CreateType == partyCreateType {
		rules := tuningForMap(mapID)
		if rules.InstanceMinutes != nil {
			run.deadline = now.Add(time.Duration(*rules.InstanceMinutes) * time.Minute)
		}
		if rules.ReentryMinutes != nil {
			run.reentry = time.Duration(*rules.ReentryMinutes) * time.Minute
		}
	}
	return run
}

func dungeonTiming(zone uint16) (dungeonRun, uint16) {
	if !isInstanceZone(zone) {
		return dungeonRun{}, zone
	}
	realm.instanceMu.Lock()
	defer realm.instanceMu.Unlock()
	return realm.instanceTiming[zone], realm.instanceMaps[zone]
}

func (run dungeonRun) expired(now time.Time) bool {
	return !run.deadline.IsZero() && !now.Before(run.deadline)
}

func (s *Session) dungeonAdmission(mapID uint16, now time.Time) bool {
	if s.isAdmin || s.char == nil || realMap(s.char.MapID) == mapID {
		return true
	}
	if !dungeonReentryPrepared.Load() {
		return true
	}
	until, err := db.DungeonReentry(s.char.ID, mapID)
	if err != nil {
		s.logErr("read dungeon re-entry", err)
		return false
	}
	if now.Before(until) {
		s.log.Printf("GX Server World re-entry cooldown map=%d character=%d", mapID, s.char.ID)
		return false
	}
	return true
}

func (s *Session) noteDungeonExit(previousZone uint16) {
	run, mapID := dungeonTiming(previousZone)
	if run.reentry == 0 {
		return
	}
	if err := db.SetDungeonReentry(s.char.ID, mapID, time.Now().Add(run.reentry)); err != nil {
		s.logErr("save dungeon re-entry", err)
	}
}

func (s *Session) finishExpiredDungeon() {
	if s.player == nil || !isInstanceZone(s.char.MapID) {
		return
	}
	run, _ := dungeonTiming(s.char.MapID)
	if !run.expired(time.Now()) {
		return
	}
	p := birthPosition(s.char)
	if s.applyTeleport(p.MapID, p.X, p.Y, p.Z, "dungeon time expired") {
		s.activity("player.map", s.char.Name+" returned from an expired dungeon")
	}
}
