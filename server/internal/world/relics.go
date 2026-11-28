package world

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// Original reference: ps_game.exe SHA256 0fb7c9f...12bc547.
// The controller always locks before a shard. Combat calls it only after
// releasing the shard; ordinary AI never locks this controller.
const (
	relicListOpcode   packet.Opcode = 0x2101
	relicOwnerOpcode  packet.Opcode = 0x2102
	relicNoticeOpcode packet.Opcode = 0x2103
)

type eventActor struct {
	mapID      uint16
	slot       int
	relic      *relic
	boss       *linkedBoss
	spawn      *eventSpawn
	principal  bool
	handled    bool
	generation uint64
}
type relicCapture struct {
	owner, old uint8
	mob        uint16
	score      bool
}
type eventPublication struct {
	mapID  uint16
	global bool
	frame  []byte
}

var relics = struct {
	sync.Mutex
	loaded       bool
	config       relicConfig
	byMap        map[uint16][]*relic
	bosses       map[uint16][]*linkedBoss
	portals      []*relicPortal
	score        [2]uint32
	resetPending bool
}{byMap: map[uint16][]*relic{}, bosses: map[uint16][]*linkedBoss{}}

func runtimeRelicCountry(owner uint8) uint8 { return (owner + 2) % 3 }
func wireRelicCountry(country uint8) uint8  { return (country + 1) % 3 }
func relicKeep(r *relic) db.Keep {
	return db.Keep{ID: r.key, Country: (runtimeRelicCountry(r.last)+1)<<4 | runtimeRelicCountry(r.owner), ResetTime: r.changeTime}
}
func restoreRelic(r *relic, k db.Keep) error {
	if k.Country&15 > 2 {
		return fmt.Errorf("keep %#x has invalid country", k.ID)
	}
	r.owner = wireRelicCountry(k.Country & 15)
	last := int(k.Country>>4) - 1
	if last < 0 || last > 2 || r.owner != 0 {
		r.last = r.owner
	} else {
		r.last = wireRelicCountry(uint8(last))
	}
	r.changeTime = k.ResetTime
	return nil
}
func buildRelicOwner(id uint32, owner uint8) []byte {
	w := packet.New(relicOwnerOpcode)
	w.U32(id)
	w.U8(owner)
	return w.MustFinish()
}
func buildRelicNotice(mob uint16, old, next uint8) []byte {
	w := packet.New(relicNoticeOpcode)
	w.U16(mob)
	w.U8(old)
	w.U8(next)
	return w.MustFinish()
}
func publishEvents(list []eventPublication) {
	for _, p := range list {
		if p.global {
			broadcast(AllPlayers(), p.frame, "obelisk event", nil)
		} else {
			broadcast(PlayersOnMap(p.mapID), p.frame, "obelisk map state", nil)
		}
	}
}

// LoadRelics restores the existing Keeps table, then populates primary maps.
// No schema/seed rewrite and no synthetic character/quest/kill credit occurs.
func LoadRelics(log *logx.Logger) error {
	raw, err := os.ReadFile(filepath.Join(gameDataDir(), "Obelisk.ini"))
	if err != nil {
		return err
	}
	cfg, err := parseRelicConfig(raw)
	if err != nil {
		return err
	}
	saved, err := db.ReadKeeps()
	if err != nil {
		return err
	}
	relics.Lock()
	defer relics.Unlock()
	if relics.loaded {
		return fmt.Errorf("obelisk controller already loaded")
	}
	var defaults []db.Keep
	id := uint32(0)
	for _, r := range cfg.relics {
		if !hasGrid(r.mapID) {
			continue
		}
		size, _ := mapSize(r.mapID)
		if r.pos[0] < 0 || r.pos[2] < 0 || r.pos[0] >= size || r.pos[2] >= size {
			continue
		}
		r.id = id
		id++
		if k, ok := saved[r.key]; ok {
			if err := restoreRelic(r, k); err != nil {
				return err
			}
		} else {
			defaults = append(defaults, relicKeep(r))
		}
		relics.byMap[r.mapID] = append(relics.byMap[r.mapID], r)
	}
	relics.config = cfg
	if err := bindRelicPortals(cfg.portals, saved, &id, &defaults); err != nil {
		return err
	}
	if err := db.SaveKeeps(defaults); err != nil {
		return err
	}
	total := 0
	for _, list := range relics.byMap {
		total += len(list)
	}
	if total > 0 {
		kcMu.Lock()
		score := kcMaxKill()/uint32(total) + 1
		relics.score = [2]uint32{score, score}
		kcMu.Unlock()
	}
	for mapID, list := range relics.byMap {
		s := shard(mapID)
		s.mu.Lock()
		for _, r := range list {
			r.principal = allocateEventActor(s, mapID, r, nil, nil, true)
			for _, g := range r.guards {
				for range g.count {
					g.actors = append(g.actors, allocateEventActor(s, mapID, r, nil, g, false))
				}
			}
			spawnRelicLocked(s, r, time.Now())
		}
		s.mu.Unlock()
	}
	relics.bosses = map[uint16][]*linkedBoss{}
	for _, b := range cfg.bosses {
		if hasGrid(b.mapID) {
			initializeBossLocked(b.mapID, b, time.Now())
		}
	}
	relics.loaded = true
	log.Printf("GX Server World obelisks=%d linked_bosses=%d portal_links=%d restored_keeps=%d", total, len(cfg.bosses), len(relics.portals), len(saved))
	return nil
}
func allocateEventActor(s *mapShard, mapID uint16, r *relic, b *linkedBoss, g *eventSpawn, principal bool) *eventActor {
	a := &eventActor{mapID: mapID, slot: len(s.entities.Mobs), relic: r, boss: b, spawn: g, principal: principal}
	m := NewStaticMob(nil, 0, 0, 0, nil, 0)
	m.CurrentHP = 0
	m.corpseGone = true
	m.event = a
	s.entities.Mobs = append(s.entities.Mobs, m)
	return a
}
func spawnEventActor(s *mapShard, a *eventActor, id uint16, pos [3]float32, now time.Time) bool {
	def := db.MustDefs().Mobs[id]
	if id == 0 || def == nil {
		return false
	}
	old := s.entities.Mobs[a.slot]
	m := NewStaticMob(def, id, pos[0], pos[2], nil, 0)
	m.ID = old.ID
	m.PosY, m.homeY = pos[1], pos[1]
	m.event = a
	m.eventFresh = old.MobID != 0
	m.spawnMinY, m.spawnMaxY, m.spawnHeightKnown = pos[1], pos[1], true
	// Event mobs (obelisk bosses, their companions and chests) follow the same
	// placement rules as every other mob: they are grounded on the navigation
	// floor chosen by their authored height, so they neither float above open
	// terrain nor sink to the floor of a dungeon. Where no floor accepts the
	// spot, a stationary actor keeps its authored position instead of vanishing.
	nav, err := loadMobNavigation(a.mapID)
	if err != nil {
		return false
	}
	placed := nav == nil || m.placeOnNavigation(nav, a.mapID)
	if !placed {
		if !m.stationary() {
			return false
		}
		m.nav = nil
		m.PosX, m.PosY, m.PosZ = pos[0], pos[1], pos[2]
		m.homeY = pos[1]
	}
	if !m.stationary() {
		// Mobile companions use the ordinary idle wander/chase path around that home point.
		m.Area = &MobArea{LowerX: m.PosX - mobHomeRadius, LowerZ: m.PosZ - mobHomeRadius, UpperX: m.PosX + mobHomeRadius, UpperZ: m.PosZ + mobHomeRadius}
	}
	a.generation++
	m.eventGeneration = a.generation
	a.handled = false
	m.noTargetUntil = now.Add(5 * time.Second)
	s.entities.Mobs[a.slot] = m
	s.setQuestEventSpawn(&m, true)
	return true
}
func removeEventActor(s *mapShard, a *eventActor) {
	m := &s.entities.Mobs[a.slot]
	s.setQuestEventSpawn(m, false)
	m.CurrentHP = 0
	m.Effects = nil
	m.aggro = nil
	m.AggroTarget = 0
	m.corpseGone = true
	m.respawnAt = time.Time{}
}
func eventSpawnPosition(mapID uint16, g *eventSpawn) ([3]float32, bool) {
	p := g.pos
	span := max(int(g.radius)*2, 1)
	p[0] += float32(rand.IntN(span)) - g.radius
	p[2] += float32(rand.IntN(span)) - g.radius
	realm.gridsMu.RLock()
	grid, ok := realm.grids[mapID]
	realm.gridsMu.RUnlock()
	x, z := int(p[0]), int(p[2])
	if !ok || x < 0 || z < 0 || x >= int(grid.size) || z >= int(grid.size) {
		return p, false
	}
	at := int(grid.size)*z + x
	if at/8 < len(grid.mask) && grid.mask[at/8]&(1<<uint(at%8)) != 0 {
		return p, false
	}
	return p, true
}
func populateEventSpawn(s *mapShard, g *eventSpawn, stage int, now time.Time) {
	if stage >= len(g.mobs) || g.mobs[stage] == 0 {
		return
	}
	g.next = time.Time{}
	for _, a := range g.actors {
		// Random offsets may land on a blocked edge. Try another authored
		// spawn point before leaving this slot for the controller's retry.
		for attempt := 0; attempt < 32; attempt++ {
			if pos, ok := eventSpawnPosition(a.mapID, g); ok && spawnEventActor(s, a, g.mobs[stage], pos, now) {
				break
			}
		}
	}
}
func spawnRelicLocked(s *mapShard, r *relic, now time.Time) {
	r.standing = spawnEventActor(s, r.principal, r.mobs[r.owner], r.pos, now)
	if !r.standing {
		return
	}
	r.next = time.Time{}
	for _, g := range r.guards {
		populateEventSpawn(s, g, int(r.owner), now)
	}
}
func tickEventSpawn(s *mapShard, g *eventSpawn, stage int, now time.Time) {
	if stage >= len(g.mobs) || g.mobs[stage] == 0 || now.Before(g.next) {
		return
	}
	for _, a := range g.actors {
		dead := &s.entities.Mobs[a.slot]
		if dead.CurrentHP > 0 || !dead.diedAt.IsZero() && (!a.handled || !dead.corpseGone) {
			continue
		}
		if p, ok := eventSpawnPosition(a.mapID, g); ok && spawnEventActor(s, a, g.mobs[stage], p, now) {
			g.next = now.Add(time.Duration(db.MustDefs().Mobs[g.mobs[stage]].RespawnSeconds) * time.Second)
			return
		}
	}
}

// eventMobDestroyed receives the already resolved single mob death, including
// DoTs. It removes dependent populations immediately; controller ticks never
// depend on nearby players and cannot resurrect a stale guard generation.
func eventMobDestroyed(killer *ConnectedPlayer, mapID uint16, m *MapMob, log logPrinter) {
	a := m.event
	if a == nil {
		return
	}
	now := time.Now()
	relics.Lock()
	s := shard(mapID)
	s.mu.Lock()
	if a.handled || a.generation != m.eventGeneration || a.slot >= len(s.entities.Mobs) || s.entities.Mobs[a.slot].ID != m.ID {
		s.mu.Unlock()
		relics.Unlock()
		return
	}
	a.handled = true
	var publications []eventPublication
	var score *relicCapture
	if !a.principal {
		if a.spawn.next.IsZero() {
			a.spawn.next = now.Add(time.Duration(m.def.RespawnSeconds) * time.Second)
		}
	} else if r := a.relic; r != nil && r.standing {
		r.standing = false
		r.next = now.Add(mobCorpseDelay + time.Duration(m.def.RespawnSeconds)*time.Second)
		for _, g := range r.guards {
			for _, member := range g.actors {
				removeEventActor(s, member)
			}
		}
		owner := uint8(0)
		if killer == nil && r.owner != 0 {
			owner = 3 - r.owner
		}
		if killer != nil {
			owner = killer.Faction() + 1
		}
		if owner == 0 && r.initial != 0 {
			owner = r.initial
		}
		r.pending = &relicCapture{old: r.owner, owner: owner, mob: m.MobID, score: true}
		score = commitRelicCapture(r, &publications, log)
	} else if a.boss != nil {
		bossDestroyedLocked(s, a.boss, now, &publications)
	}
	s.mu.Unlock()
	if a.boss != nil && a.principal {
		syncBossPortal(a.boss)
	}
	relics.Unlock()
	// View ownership is reconciled now, including immediate dependent removals.
	syncEventVisibility(mapID)
	publishEvents(publications)
	if score != nil {
		scoreRelicCapture(*score, log)
	}
}
func commitRelicCapture(r *relic, out *[]eventPublication, log logPrinter) *relicCapture {
	if r.pending == nil {
		return nil
	}
	capture := *r.pending
	next := *r
	next.owner = capture.owner
	if next.owner != 0 {
		next.last = next.owner
	}
	if capture.old != next.owner {
		if err := db.SaveKeeps([]db.Keep{relicKeep(&next)}); err != nil {
			if log != nil {
				log.Printf("GX Server World keep save failed key=%#x: %v", r.key, err)
			}
			return nil
		}
	}
	r.owner, r.last, r.pending = next.owner, next.last, nil
	*out = append(*out, eventPublication{mapID: r.mapID, frame: buildRelicOwner(r.id, r.owner)})
	if capture.score {
		*out = append(*out, eventPublication{global: true, frame: buildRelicNotice(capture.mob, capture.old, capture.owner)})
	}
	return &capture
}
func scoreRelicCapture(c relicCapture, log logPrinter) {
	if !c.score || c.owner == 0 {
		return
	}
	relics.Lock()
	amounts := relics.score
	relics.Unlock()
	AddKillCount(c.owner-1, amounts[c.owner-1], log)
	if c.old != 0 {
		SubKillCount(c.old-1, amounts[c.old-1], log)
	}
}
func syncEventVisibility(mapID uint16) {
	viewers := len(PlayersOnMap(mapID)) > 0
	s := shard(mapID)
	s.mu.Lock()
	var moves []entityMove
	var leaves []uint32
	now := time.Now()
	for i := range s.entities.Mobs {
		m := &s.entities.Mobs[i]
		if m.event == nil {
			continue
		}
		if m.eventFresh {
			if viewers {
				leaves = append(leaves, m.ID)
			}
			m.eventFresh = false
		}
		if !viewers {
			continue
		}
		if !m.visible() {
			leaves = append(leaves, m.ID)
		} else {
			enter, move := m.appearance(now, true)
			moves = append(moves, entityMove{MapID: mapID, Kind: viewMob, ID: m.ID, To: m.cell, Fresh: true, Enter: enter, Move: move})
		}
	}
	s.mu.Unlock()
	for _, id := range leaves {
		dispatchEntityLeave(mapID, viewMob, id, "obelisk dependent removed")
	}
	dispatchMoves(moves)
}

// TickRelics runs independently of map viewers. The mutex also serializes
// captures/reset/portal changes; packet transmission follows the lock release.
func TickRelics(now time.Time, log logPrinter) {
	relics.Lock()
	if !relics.loaded {
		relics.Unlock()
		return
	}
	var publications []eventPublication
	var scores []relicCapture
	var maps []uint16
	for mapID, list := range relics.byMap {
		s := shard(mapID)
		s.mu.Lock()
		for _, r := range list {
			if r.pending != nil {
				if c := commitRelicCapture(r, &publications, log); c != nil {
					scores = append(scores, *c)
				}
			}
			if !r.standing && r.pending == nil && !r.next.IsZero() && !now.Before(r.next) {
				spawnRelicLocked(s, r, now)
			}
			if r.standing {
				for _, g := range r.guards {
					tickEventSpawn(s, g, int(r.owner), now)
				}
			}
		}
		s.mu.Unlock()
		maps = append(maps, mapID)
	}
	for mapID, list := range relics.bosses {
		s := shard(mapID)
		s.mu.Lock()
		for _, b := range list {
			tickBossLocked(s, b, now, &publications)
		}
		s.mu.Unlock()
		for _, b := range list {
			syncBossPortal(b)
		}
		maps = append(maps, mapID)
	}
	updateRelicPortalsLocked(&publications, log)
	reset := relics.resetPending
	relics.resetPending = false
	relics.Unlock()
	for _, mapID := range maps {
		syncEventVisibility(mapID)
	}
	publishEvents(publications)
	for _, c := range scores {
		scoreRelicCapture(c, log)
	}
	if reset {
		ResetRelics(now, log)
	}
}

// ResetRelics is the blessing-activation reset: configured base ownership,
// guards removed, no capture scoring/notices, then a ten-minute spawn delay.
func ResetRelics(now time.Time, log logPrinter) {
	relics.Lock()
	if !relics.loaded {
		relics.Unlock()
		return
	}
	var keeps []db.Keep
	for _, list := range relics.byMap {
		for _, r := range list {
			copy := *r
			copy.owner = r.initial
			if copy.owner != 0 {
				copy.last = copy.owner
			}
			keeps = append(keeps, relicKeep(&copy))
		}
	}
	if err := db.SaveKeeps(keeps); err != nil {
		relics.resetPending = true
		relics.Unlock()
		if log != nil {
			log.Printf("GX Server World obelisk reset save failed: %v", err)
		}
		return
	}
	var publications []eventPublication
	var maps []uint16
	for mapID, list := range relics.byMap {
		s := shard(mapID)
		s.mu.Lock()
		for _, r := range list {
			removeEventActor(s, r.principal)
			for _, g := range r.guards {
				for _, a := range g.actors {
					removeEventActor(s, a)
				}
			}
			r.owner = r.initial
			if r.owner != 0 {
				r.last = r.owner
			}
			r.standing = false
			r.pending = nil
			r.next = now.Add(kcHoldTime)
			publications = append(publications, eventPublication{mapID: mapID, frame: buildRelicOwner(r.id, r.owner)})
		}
		s.mu.Unlock()
		maps = append(maps, mapID)
	}
	updateRelicPortalsLocked(&publications, log)
	relics.Unlock()
	for _, mapID := range maps {
		syncEventVisibility(mapID)
	}
	publishEvents(publications)
}
func relicListFor(mapID uint16) [][]byte {
	relics.Lock()
	defer relics.Unlock()
	type row struct {
		id    uint32
		owner uint8
		x, z  float32
	}
	var rows []row
	for _, r := range relics.byMap[mapID] {
		rows = append(rows, row{r.id, r.owner, r.pos[0], r.pos[2]})
	}
	for _, p := range relics.portals {
		if p.mapID == mapID {
			rows = append(rows, row{p.id, portalMarker(p.country), p.pos[0], p.pos[2]})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].id < rows[j].id })
	var out [][]byte
	for len(rows) > 0 {
		n := min(100, len(rows))
		w := packet.New(relicListOpcode)
		w.U8(uint8(n))
		for _, r := range rows[:n] {
			w.U32(r.id)
			w.U8(r.owner)
			w.F32(r.x)
			w.F32(r.z)
		}
		out = append(out, w.MustFinish())
		rows = rows[n:]
	}
	return out
}
func (s *Session) sendRelicList() {
	for _, p := range relicListFor(s.char.MapID) {
		s.send(p, "0x2101 obelisks/portals")
	}
}
func (m *MapMob) stationary() bool {
	return m.def != nil && m.def.ChaseStep == 0 && m.def.NormalStep == 0
}
func mobFriendly(def *db.MobDef, faction uint8) bool {
	return def != nil && def.Country < 2 && def.Country == faction
}

// Managed principals always use the original fair-combat damage owner, even
// when another player landed the final hit. Aggro cannot substitute for damage.
type eventDamageEntry struct {
	ID     uint32
	Amount uint64
}

func (m *MapMob) recordEventDamage(id, damage uint32) {
	if id == 0 || damage == 0 || m.event == nil || !m.event.principal {
		return
	}
	for i := range m.eventDamage {
		if m.eventDamage[i].ID == id {
			m.eventDamage[i].Amount += uint64(damage)
			return
		}
	}
	m.eventDamage = append(m.eventDamage, eventDamageEntry{id, uint64(damage)})
}
func eventDamageWinner(m *MapMob) uint32 {
	var best eventDamageEntry
	for _, d := range m.eventDamage {
		if d.Amount > best.Amount {
			best = d
		}
	}
	return best.ID
}
func eventCreditedPlayer(last *ConnectedPlayer, mapID uint16, m *MapMob) *ConnectedPlayer {
	id := eventDamageWinner(m)
	if id == 0 {
		return last
	}
	if last != nil && last.CharacterID == id {
		return last
	}
	p := FindPlayerByID(id)
	if p != nil && p.MapID() == mapID {
		return p
	}
	return nil
}

func mobDeathCredit(m *MapMob, last uint32) uint32 {
	if m.event != nil && m.event.principal {
		if id := eventDamageWinner(m); id != 0 {
			return id
		}
	}
	return last
}
