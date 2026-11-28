package world

import (
	"context"
	"github.com/Spelunkern/Shaiya-GX/server/internal/scheduler"
	"github.com/Spelunkern/Shaiya-GX/server/internal/session"
	"sync"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
)

// StartSweeps launches the background world ticks: map item despawn, mob
// roam/respawn/combat, NPC patrol and mob DoTs.
type Runtime struct {
	state  *realmState
	fatal  context.CancelFunc
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

// Close stops producers before storage closes and waits for the current job.
func (r *Runtime) Close() {
	r.once.Do(r.cancel)
	<-r.done
	r.state.runtimeMu.Lock()
	if r.state.runtime == r {
		r.state.runtime = nil
	}
	r.state.runtimeMu.Unlock()
}

func StartSweeps(ctx context.Context, log *logx.Logger, onFatal ...context.CancelFunc) *Runtime {
	ctx, cancel := context.WithCancel(ctx)
	r := &Runtime{state: realm, cancel: cancel, done: make(chan struct{})}
	if len(onFatal) > 0 {
		r.fatal = onFatal[0]
	}
	realm.runtimeMu.Lock()
	if realm.runtime != nil {
		realm.runtimeMu.Unlock()
		cancel()
		panic("world runtime already running")
	}
	realm.runtime = r
	realm.runtimeMu.Unlock()
	jobs := []scheduler.Job{{Name: "storage", Interval: 5 * time.Second, Run: func(time.Time) { db.Observe(); db.RetryAccountReleases() }}}
	add := func(name string, d time.Duration, fn func(time.Time)) {
		jobs = append(jobs, scheduler.Job{Name: name, Interval: d, Run: fn})
	}
	var announcedRewardEvent *config.TimedRewardEvent
	announcedPvP := map[uint16]bool{}
	add("reward_events", time.Second, func(now time.Time) {
		pollRewardEvent(&announcedRewardEvent, now, log)
		pollPvPWindows(announcedPvP, now, log)
	})
	add("items", 2*time.Second, func(now time.Time) {
		for _, e := range sweepExpiredMapItems() {
			log.Printf("GX Server World despawned map_item_id=%d on map=%d", e.Item.ID, e.MapID)
			dispatchEntityLeave(e.MapID, viewItem, e.Item.ID, "0x0402 map item despawn")
		}
	})
	add("respawns", time.Second, func(now time.Time) {
		for _, c := range TickMobCorpses() {
			dispatchEntityLeave(c.MapID, viewMob, c.ID, "0x0602 mob corpse removed")
		}
		dispatchMoves(TickMobRespawns(viewersByMap(), 128))
	})
	add("patrol", 250*time.Millisecond, func(now time.Time) {
		tick := tickNpcs(now, snapshotCombatPlayers(AllPlayers()), 64)
		dispatchMoves(tick.moves)
		for _, hit := range tick.strikes {
			dispatchNpcStrike(hit, log)
		}
	})
	add("effects", time.Second, func(now time.Time) { runDotSweep(log) })
	add("guilds", 5*time.Second, func(now time.Time) { TickGuildRanking(now, log) })
	add("killcount", time.Second, func(now time.Time) { TickKillCount(now, log) })
	add("relics", 250*time.Millisecond, func(now time.Time) { TickRelics(now, log) })
	add("instances", instanceSweepTick, func(now time.Time) { TickInstances(now, log) })
	add("combat", mobNavigationTick, func(now time.Time) {
		players := AllPlayers()
		if len(players) == 0 {
			tick := MobCombatTick{}
			runMobAICommands(&tick)
			dispatchMoves(tick.Moves)
			return
		}
		snaps := snapshotCombatPlayers(players)
		tick := TickMobCombat(snaps, 256)
		for _, a := range tick.Attacks {
			queueMobAttack(a)
		}
		dispatchMoves(tick.Moves)
		for _, r := range tick.Restores {
			sendView(r.MapID, r.X, r.Z, r.Packet, "0x0610 mob hp restore")
		}
	})
	started := time.Now()

	notices := scheduleNotices(noticeCatalog, started)
	if len(notices) > 0 {
		add("notices", time.Second, func(now time.Time) {
			for i := range notices {
				if notices[i].due(now) {
					broadcastNotice(notices[i].message)
				}
			}
		})
	}
	engine := scheduler.New(started, jobs)
	go func() {
		defer close(r.done)
		defer func() { worldBots.Lock(); defer worldBots.Unlock(); clearWorldBots() }()
		var workers sync.WaitGroup
		workers.Go(func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for ctx.Err() == nil {
				if err := db.ExpireAuctions(); err != nil {
					log.Printf("auction expiry: %v", err)
				}
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
		})
		workers.Go(func() {
			botScheduler := scheduler.New(started, []scheduler.Job{{Name: "bots", Interval: botTick, Run: tickWorldBots}})
			botScheduler.Run(ctx)
		})
		workers.Go(func() {
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()
			for ctx.Err() == nil {
				if err := session.PurgeExpired(); err != nil {
					log.Printf("session handoff expiry: %v", err)
				}
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
		})
		engine.Run(ctx)
		workers.Wait()
	}()
	return r
}

func snapshotCombatPlayers(players []*ConnectedPlayer) []PlayerCombatSnapshot {
	snaps := make([]PlayerCombatSnapshot, 0, len(players))
	for _, p := range players {
		snapshot := p.Snapshot()
		x, z := snapshot.PosX, snapshot.PosZ
		c, _ := p.viewCell()
		snaps = append(snaps, PlayerCombatSnapshot{CharacterID: p.CharacterID, MapID: p.MapID(), PosX: x, PosY: snapshot.PosY, PosZ: z, Alive: p.IsAlive(), Faction: p.Faction(), Cell: c, HP: snapshot.CurrentHP, Invisible: p.noMobTarget.Load() || p.immune.Load(), GMUndetected: p.immune.Load() || !p.GMVisible()})
	}
	return snaps
}

// queueMobAttack hands a mob's attack to its target, or for an area mob
// skill (TargetType 6 around the mob, 7 around the target) to every living
// player within ApplyRange.
func queueMobAttack(a MobAttackAction) {
	target := FindPlayerByID(a.TargetCharacterID)
	if target == nil || target.noMobTarget.Load() || target.immune.Load() {
		return
	}
	first := pendingDamage{MobGlobalID: a.MobGlobalID, MobID: a.MobID, Attack: a.Attack, Slot: a.Slot, First: true, SkillID: a.SkillID}
	def := db.MustDefs().Mobs[a.MobID]
	skillID := a.SkillID
	if skillID == 0 && def != nil {
		skillID = def.Attacks[a.Slot].Type
	}
	if def == nil || skillID == 0 {
		target.QueueDamage(first)
		return
	}
	sk := db.MustDefs().NPCSkill(skillID)
	if sk == nil || (sk.TargetType != targetTypeEnemiesNearCaster && sk.TargetType != targetTypeEnemiesNearTarget) {
		target.QueueDamage(first)
		return
	}
	x, z := a.MobX, a.MobZ
	if sk.TargetType == targetTypeEnemiesNearTarget {
		x, z = target.PositionXZ()
	}
	recipients := PlayersNear(target.MapID(), x, z, float32(sk.ApplyRange))
	var primary *ConnectedPlayer
	for _, p := range recipients {
		if p.IsAlive() && !p.immune.Load() && (primary == nil || p == target) {
			primary = p
		}
	}
	for _, p := range recipients {
		if p.IsAlive() && !p.immune.Load() {
			d := first
			d.First, d.StartsCast = false, p == primary
			p.QueueDamage(d)
		}
	}
}

func viewersByMap() map[uint16][][2]float32 {
	out := map[uint16][][2]float32{}
	for _, p := range AllPlayers() {
		x, z := p.PositionXZ()
		out[p.MapID()] = append(out[p.MapID()], [2]float32{x, z})
	}
	return out
}

func dispatchMoves(list []entityMove) {
	for _, e := range list {
		dispatchEntityMove(e)
	}
}

func runDotSweep(log *logx.Logger) {
	for _, o := range TickMobEffects() {
		m := &o.Mob
		if o.Speed != nil {
			sendNear(o.MapID, m.PosX, m.PosZ, BroadcastRadius, o.Speed, "0x0609 mob speed")
		}
		if !o.Damaged {
			continue
		}
		sendNear(o.MapID, m.PosX, m.PosZ, CombatRadius,
			buildMobPeriodicDamage(m.ID, o.SkillID, o.SkillLevel, uint16(min(max(o.Damage[0], 0), 65535)),
				uint16(min(max(o.Damage[1], 0), 65535)), uint16(min(max(o.Damage[2], 0), 65535))), "0x0607 mob skill damage")
		sendNear(o.MapID, m.PosX, m.PosZ, CombatRadius, buildMobCurrentHP(m.ID, m.CurrentHP), "0x0610 mob current hp")
		if !o.Killed {
			continue
		}
		log.Printf("GX Server World skill effect kill character_id=%d mob_global_id=%d mob_id=%d", o.CasterID, m.ID, m.MobID)
		sendMobDeath(o.MapID, m, mobDeathCredit(m, o.CasterID), "0x0606 mob death (skill effect)")
		rewardKill(FindPlayerByID(o.CasterID), o.CasterID, o.MapID, m, log)
	}
}

func storageFailure(err error) {
	if !db.StorageUnavailable(err) {
		return
	}
	realm.runtimeMu.Lock()
	r := realm.runtime
	realm.runtimeMu.Unlock()
	if r != nil {
		if r.fatal != nil {
			r.fatal()
		} else {
			r.cancel()
		}
	}
}
