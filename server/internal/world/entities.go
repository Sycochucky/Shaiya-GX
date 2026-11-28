package world

import (
	"math"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/mapnav"
	"github.com/Spelunkern/Shaiya-GX/internal/npcquest"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/ai"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// Visibility and combat radii, matching the real server's cell view model.
const (
	VisibilityCellSize = 40
	ViewportRadius     = 120.0
	CombatRadius       = BroadcastRadius // CZone::PSendView
	MobAttackRange     = 3.0
	MobAttackInterval  = 2500 * time.Millisecond

	mobMotionRun  = 1
	mobMotionWalk = 0
)

type NpcPatrolPoint struct {
	X, Y, Z float32
	Angle   uint16
}

type MapNpc struct {
	Summoned         bool
	ID               uint32
	Type             uint8
	TypeID           uint16
	PosX, PosY, PosZ float32
	Angle            uint16
	patrol           []NpcPatrolPoint
	patrolIndex      int
	cell             cell
	hasCell          bool
	backward         bool      // walking the patrol points back
	nextMoveAt       time.Time // CNpc +0x4c
	guardReady       time.Time
	guardTarget      guardTarget
	guardReturning   bool
}

type MobArea struct{ LowerX, LowerZ, UpperX, UpperZ float32 }

func (a MobArea) normalized() MobArea {
	return MobArea{min(a.LowerX, a.UpperX), min(a.LowerZ, a.UpperZ), max(a.LowerX, a.UpperX), max(a.LowerZ, a.UpperZ)}
}
func (a MobArea) clampX(x float32) float32 { return min(max(x, a.LowerX), a.UpperX) }
func (a MobArea) clampZ(z float32) float32 { return min(max(z, a.LowerZ), a.UpperZ) }
func (a MobArea) contains(x, z float32) bool {
	return x >= a.LowerX && x <= a.UpperX && z >= a.LowerZ && z <= a.UpperZ
}

type MapMob struct {
	aiMapID          uint16
	aiSummoned       bool
	aiAttackType     uint8
	lua              mobAIState
	event            *eventActor
	eventGeneration  uint64
	eventFresh       bool
	eventDamage      []eventDamageEntry
	lootDamage       []eventDamageEntry // damage per attacker: who earns the loot
	elementRestored  bool               // RemApplySkillDebuff68 restores base even with another68 active.
	ID               uint32
	MobID            uint16
	Level            uint16
	PosX, PosY, PosZ float32
	nav              *mobNav
	navPath          []mapnav.Point
	navGoal          mapnav.Point
	navRefresh       time.Time
	// Pos is the destination while a movement segment is in progress.
	moveFrom             mapnav.Point
	moveStart            time.Time
	moveDuration         time.Duration
	moveMotion           uint8
	moveSpeedTier        uint8
	stoppedOnDeath       bool
	spawnMinY, spawnMaxY float32
	spawnHeightKnown     bool
	homeY                float32
	homeX, homeZ         float32
	CurrentHP, MaxHP     uint32
	MoneyMin, MoneyMax   uint32
	Exp                  uint32
	Area                 *MobArea
	respawnDelay         time.Duration // 0: no respawn (ungrouped) / immediate (grouped)
	group                *mobGroup     // stMobRespawn: the area entry it belongs to
	respawnAt            time.Time
	roamStep             uint32
	Attack               uint32
	AI                   uint8
	ChaseRange           float32
	AggroTarget          uint32 // 0: none
	lastAttackAt         time.Time
	nextAttackAt         time.Time        // CMob m_dwNextAttackTime
	noTargetUntil        time.Time        // CMob::Init: no target search for 5 s
	aggro                map[uint32]int64 // CDamage ledger: damage and skill aggro
	slotReadyAt          [3]time.Time     // per attack slot (AttackAI)
	returningHome        bool
	def                  *db.MobDef
	nextMoveAt           time.Time   // idle wander
	nextChaseAt          time.Time   // chase / return step
	Effects              []MobEffect // applied skills (TypeEffect 3)
	cell                 cell        // view cell, with the move hysteresis
	hasCell              bool
	diedAt               time.Time
	corpseGone           bool // the corpse left the view (5 s after death)
}

type MapPortal struct {
	Index                     uint8
	PortalID                  int32
	FactionOrPortalID         int32
	MinLevel, MaxLevel        uint16
	TargetMapID               uint16
	PosX, PosY, PosZ          float32
	TargetX, TargetY, TargetZ float32
	volume                    *portalVolume
}

func (p *MapPortal) Contains(x, y, z float32) bool {
	if p.volume != nil {
		// The original tests the actor origin, .9 above the player's feet.
		return p.volume.contains(x, y+.9, z, 0)
	}
	return abs32(p.PosX-x) <= 5 && abs32(p.PosY-y) <= 5 && abs32(p.PosZ-z) <= 5
}
func (p *MapPortal) IsOpen() bool { return p.FactionOrPortalID <= 2 }
func (p *MapPortal) AllowsFaction(f uint8) bool {
	switch p.FactionOrPortalID {
	case 1:
		return f == 0
	case 2:
		return f == 1
	}
	return true
}
func (p *MapPortal) AllowsLevel(level uint16) bool { return level >= p.MinLevel && level <= p.MaxLevel }

// MapSpawnArea is one stMapBind of the svmap: a rebirth group of boxes.
// Faction is its type: 1 light, 2 fury, 0 and 3 both.
type MapSpawnArea struct {
	Faction int32
	Boxes   []SpawnBox
}

type SpawnBox struct {
	LowerX, LowerY, LowerZ float32
	UpperX, UpperY, UpperZ float32
}

// distanceSq is measured, as CMap::GetBindPos does, to the lower corner of
// the group's first box.
func (a *MapSpawnArea) distanceSq(x, z float32) float32 {
	b := a.Boxes[0]
	return (b.LowerX-x)*(b.LowerX-x) + (b.LowerZ-z)*(b.LowerZ-z)
}
func (a *MapSpawnArea) allowsFaction(f uint8) bool {
	switch a.Faction {
	case 0, 3:
		return true
	case 1:
		return f == 0
	case 2:
		return f == 1
	}
	return false
}

// RandomPoint is GetBindPos's pick: a random box of the group and a random
// whole-unit offset inside it, 0.9 above its floor.
func (a *MapSpawnArea) RandomPoint() (x, y, z float32) {
	b := a.Boxes[rand.IntN(len(a.Boxes))]
	offset := func(lo, hi float32) float32 {
		if span := int(math.Round(float64(hi - lo))); span > 0 {
			return float32(rand.IntN(span)) + lo
		}
		return lo
	}
	return offset(b.LowerX, b.UpperX), b.LowerY + 0.9, offset(b.LowerZ, b.UpperZ)
}

// MapEntities is one map's NPCs, mobs, portals and spawn areas.
type MapEntities struct {
	NPCs       []MapNpc
	Mobs       []MapMob
	Portals    []MapPortal
	SpawnAreas []MapSpawnArea
}

type mapShard struct {
	guardClaims      map[guardTarget]time.Time
	items            map[uint32]*mapItemEntry
	lua              *ai.VM
	aiCommands       []mobAICommand
	aiCommandCost    int
	aiPendingSummons int
	npcRevision      atomic.Uint64
	questNPCs        questNPCCatalogCache
	questMobSpawns   []proto.QuestMobSpawn
	questEventSpawns map[uint32]proto.QuestMobSpawn
	questMobRevision atomic.Uint64
	questMobCatalog  [][]byte
	mu               sync.Mutex
	entities         MapEntities
}

type mapGrid struct {
	size uint32
	mask []byte
}

func shard(mapID uint16) *mapShard {
	realm.shardsMu.Lock()
	defer realm.shardsMu.Unlock()
	s, ok := realm.shards[mapID]
	if !ok {
		s = &mapShard{}
		realm.shards[mapID] = s
	}
	return s
}

func shardIDs() []uint16 {
	realm.shardsMu.Lock()
	defer realm.shardsMu.Unlock()
	ids := make([]uint16, 0, len(realm.shards))
	for id := range realm.shards {
		ids = append(ids, id)
	}
	return ids
}

func cloneMob(m *MapMob) MapMob {
	c := *m
	c.lua.actor = nil
	c.lua.rng = nil
	c.Effects = append([]MobEffect(nil), m.Effects...)
	c.eventDamage = append([]eventDamageEntry(nil), m.eventDamage...)
	c.lootDamage = append([]eventDamageEntry(nil), m.lootDamage...)
	if m.Area != nil {
		a := *m.Area
		c.Area = &a
	}
	return c
}

func (m *MapMob) snapshotAt(now time.Time) MapMob {
	c := cloneMob(m)
	c.setPoint(m.positionAt(now))
	c.moveDuration = 0
	return c
}

// EntitiesOnMap returns a copy of a map's entities.
func EntitiesOnMap(mapID uint16) MapEntities {
	s := shard(mapID)
	s.mu.Lock()
	defer s.mu.Unlock()
	out := MapEntities{NPCs: append([]MapNpc(nil), s.entities.NPCs...),
		Portals:    append([]MapPortal(nil), s.entities.Portals...),
		SpawnAreas: append([]MapSpawnArea(nil), s.entities.SpawnAreas...)}
	for i := range s.entities.Mobs {
		out.Mobs = append(out.Mobs, cloneMob(&s.entities.Mobs[i]))
	}
	return out
}

func RegisterMapGrid(mapID uint16, size uint32, mask []byte) {
	realm.gridsMu.Lock()
	realm.grids[mapID] = mapGrid{size, mask}
	realm.gridsMu.Unlock()
}

// mapSize is the svmap's side length.
func mapSize(mapID uint16) (float32, bool) {
	realm.gridsMu.RLock()
	defer realm.gridsMu.RUnlock()
	g, ok := realm.grids[mapID]
	return float32(g.size), ok
}

func hasGrid(mapID uint16) bool {
	realm.gridsMu.RLock()
	defer realm.gridsMu.RUnlock()
	_, ok := realm.grids[mapID]
	return ok
}

func abs32(v float32) float32 { return float32(math.Abs(float64(v))) }

func DistanceXZ(ax, az, bx, bz float32) float32 {
	dx, dz := ax-bx, az-bz
	return float32(math.Sqrt(float64(dx*dx + dz*dz)))
}

type cell struct{ x, z int32 }

func coordCell(v float32, size float32) int32 {
	if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) || v <= 0 {
		return 0
	}
	return int32(math.Floor(math.Round(float64(v)) / float64(size)))
}

func cellFor(x, z float32) cell {
	return cell{coordCell(x, VisibilityCellSize), coordCell(z, VisibilityCellSize)}
}

// isNeighborCell is the 5x5 view window.
func isNeighborCell(a, b cell) bool {
	dx, dz := a.x-b.x, a.z-b.z
	if dx < 0 {
		dx = -dx
	}
	if dz < 0 {
		dz = -dz
	}
	return dx <= 2 && dz <= 2
}

// EntitiesNear returns the entities in the 5x5 cell window around (x, z).
func EntitiesNear(mapID uint16, x, z float32) MapEntities {
	return EntitiesInView(mapID, cellFor(x, z))
}

// EntitiesInView returns the NPCs and the visible mobs whose cell is in the
// 5x5 window around the viewer's cell.
func EntitiesInView(mapID uint16, viewer cell) MapEntities {
	s := shard(mapID)
	s.mu.Lock()
	defer s.mu.Unlock()
	var out MapEntities
	for i := range s.entities.NPCs {
		n := &s.entities.NPCs[i]
		c := n.cell
		if !n.hasCell {
			c = cellFor(n.PosX, n.PosZ)
		}
		if isNeighborCell(viewer, c) {
			out.NPCs = append(out.NPCs, *n)
		}
	}
	for i := range s.entities.Mobs {
		m := &s.entities.Mobs[i]
		c := m.cell
		if !m.hasCell {
			c = cellFor(m.PosX, m.PosZ)
		}
		if m.visible() && isNeighborCell(viewer, c) {
			out.Mobs = append(out.Mobs, cloneMob(m))
		}
	}
	return out
}

func IsMapWalkable(mapID uint16, x, z float32) bool {
	realm.gridsMu.RLock()
	g, ok := realm.grids[mapID]
	realm.gridsMu.RUnlock()
	if !ok {
		return true
	}
	return walkableInMask(g.size, g.mask, x, z)
}

func walkableInMask(size uint32, mask []byte, x, z float32) bool {
	if size == 0 || len(mask) == 0 || math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) ||
		math.IsNaN(float64(z)) || math.IsInf(float64(z), 0) {
		return true
	}
	if x < 0 || z < 0 {
		return false
	}
	ix, iz := uint32(math.Floor(float64(x))), uint32(math.Floor(float64(z)))
	if ix >= size || iz >= size {
		return false
	}
	bit := uint64(iz)*uint64(size) + uint64(ix)
	if bit/8 >= uint64(len(mask)) {
		return false
	}
	return (mask[bit/8]>>(bit%8))&1 == 0
}

func MobByID(mapID uint16, id uint32) (MapMob, bool) {
	s := shard(mapID)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.entities.Mobs {
		if s.entities.Mobs[i].ID == id {
			return s.entities.Mobs[i].snapshotAt(time.Now()), true
		}
	}
	return MapMob{}, false
}

type MobDamageResult struct {
	Mob            MapMob
	Speed          []byte // 0x0609 when the hit woke the mob
	DamageHP       uint16
	AppliedHP      uint32
	Killed         bool
	RespawnSeconds uint64
}

// DamageMob applies damage unless the mob is dead or resetting home.
func DamageMob(mapID uint16, id uint32, damage uint32) (MobDamageResult, bool) {
	return DamageMobFrom(mapID, id, damage, 0)
}
func DamageMobFrom(mapID uint16, id uint32, damage, attacker uint32) (MobDamageResult, bool) {
	s := shard(mapID)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.entities.Mobs {
		m := &s.entities.Mobs[i]
		if m.ID != id {
			continue
		}
		h := &mobAIHost{mob: m, shard: s, mapID: mapID, now: time.Now(), tick: &MobCombatTick{}}
		if !m.ensureAI(h) {
			return MobDamageResult{}, false
		}
		if m.CurrentHP == 0 || m.returningHome {
			return MobDamageResult{}, false
		}
		m.lua.attacked, m.lua.recentHit = true, true
		m.lua.attackedCount++
		if m.lua.counting {
			m.lua.hits++
		}
		h.event("attacked", attacker)
		applied := min(damage, m.CurrentHP)
		if m.lua.invulnerable {
			applied = 0
			damage = 0
		}
		m.recordEventDamage(attacker, applied)
		m.recordLootDamage(attacker, applied)
		m.CurrentHP -= applied
		r := MobDamageResult{DamageHP: uint16(min(damage, 65535)), AppliedHP: applied, Killed: m.CurrentHP == 0}
		now := time.Now()
		if r.Killed {
			m.died(now)
			r.RespawnSeconds = uint64(m.configuredRespawnDelay(mapID) / time.Second)
		} else if applied > 0 {
			r.Speed = m.removeSleep()
		}
		r.Mob = m.snapshotAt(now)
		return r, true
	}
	return MobDamageResult{}, false
}

type PlayerCombatSnapshot struct {
	CharacterID      uint32
	MapID            uint16
	PosX, PosY, PosZ float32
	Alive            bool
	Faction          uint8
	Cell             cell
	HP               uint32
	Invisible        bool
	// GMUndetected is a GM in /char on or /attack on mode. Unlike skill stealth,
	// which guards see through, the original admin flags hide the user from
	// every NPC (bVisible and bAttackable).
	GMUndetected bool
}

type MobAttackAction struct {
	MobGlobalID       uint32
	MobID             uint16
	Attack            uint32
	TargetCharacterID uint32
	Slot              uint8
	MobX, MobZ        float32
	SkillID           uint16
}

// mapPacket is a packet to broadcast near (x, z) on a map.
type mapPacket struct {
	MapID  uint16
	X, Z   float32
	Packet []byte
}

type MobCombatTick struct {
	Attacks  []MobAttackAction
	Moves    []entityMove
	Restores []mapPacket
}

type viewerBuckets map[cell][][2]float32

func bucketViewers(viewers [][2]float32) viewerBuckets {
	b := viewerBuckets{}
	for _, v := range viewers {
		c := cellFor(v[0], v[1])
		b[c] = append(b[c], v)
	}
	return b
}

func (b viewerBuckets) nearAny(x, z float32, radius float32) bool {
	own := cellFor(x, z)
	r2 := radius * radius
	for c, vs := range b {
		if !isNeighborCell(own, c) {
			continue
		}
		for _, v := range vs {
			dx, dz := x-v[0], z-v[1]
			if dx*dx+dz*dz <= r2 {
				return true
			}
		}
	}
	return false
}

// TickMobRespawns revives due mobs at a fresh position in their area.
func TickMobRespawns(viewersByMap map[uint16][][2]float32, maxPerMap int) []entityMove {
	var out []entityMove
	now := time.Now()
	for mapID, viewers := range viewersByMap {
		bossWindows := tuningForMap(mapID).BossWindows
		s := shard(mapID)
		s.mu.Lock()
		buckets := bucketViewers(viewers)
		respawned := 0
		for i := range s.entities.Mobs {
			if respawned >= maxPerMap {
				break
			}
			m := &s.entities.Mobs[i]
			if m.CurrentHP > 0 || m.event != nil {
				continue
			}
			if window, scheduled := bossWindows[m.MobID]; scheduled && !m.aiSummoned && !window.Contains(now) {
				continue
			}
			var x, z float32
			if g := m.group; g != nil {
				if !m.corpseGone || g.alive >= g.count || now.Before(g.respawnAt) {
					continue
				}
				var ok bool
				if x, z, ok = m.respawnSpot(mapID); !ok {
					continue
				}
				if m.nav != nil {
					copy := *m
					copy.PosX, copy.PosZ = x, z
					if !copy.placeRespawn(mapID) {
						continue
					}
					m.PosY, m.homeY = copy.PosY, copy.homeY
					x, z = copy.PosX, copy.PosZ
					m.navPath = nil
				}
				g.alive++
				g.respawnAt = now.Add(m.configuredRespawnDelay(mapID))
			} else {
				if m.respawnAt.IsZero() || m.respawnAt.After(now) {
					continue
				}
				x, z = m.pickRespawnPosition(mapID)
				if m.nav != nil {
					copy := *m
					copy.PosX, copy.PosZ = x, z
					copy.PosY = m.homeY
					if !copy.placeRespawn(mapID) {
						continue
					}
					m.PosY, m.homeY = copy.PosY, copy.homeY
					x, z = copy.PosX, copy.PosZ
					m.navPath = nil
				}
			}
			m.PosX, m.PosZ, m.homeX, m.homeZ = x, z, x, z
			m.CurrentHP = max(m.MaxHP, 1)
			m.lootDamage = nil
			if m.def != nil {
			}
			m.resetAI()
			m.respawnAt = time.Time{}
			m.AggroTarget, m.lastAttackAt, m.nextAttackAt, m.returningHome = 0, time.Time{}, time.Time{}, false
			respawned++
			m.cell, m.hasCell = cellFor(x, z), true
			m.diedAt, m.corpseGone, m.stoppedOnDeath = time.Time{}, false, false
			m.moveDuration = 0
			m.noTargetUntil, m.aggro = now.Add(5*time.Second), nil
			out = append(out, entityMove{MapID: mapID, Kind: viewMob, ID: m.ID, To: m.cell,
				Fresh: true, Changed: true, Enter: buildMobEnter(m, true)})
			_ = buckets
		}
		s.mu.Unlock()
	}
	return out
}

// CNpc::UpdateMove pauses at each waypoint and reverses at route endpoints.
func (n *MapNpc) stepPatrol(now time.Time, mapID uint16) bool {
	def := findNpcDef(n.Type, n.TypeID)
	if def == nil || def.MoveDistance <= 0 {
		n.nextMoveAt = now.Add(time.Minute)
		return false
	}
	interval := time.Duration(max(def.MoveSpeed, 0)) * time.Millisecond
	p := n.patrol[n.patrolIndex]
	dx, dz := p.X-n.PosX, p.Z-n.PosZ
	dist := float32(math.Hypot(float64(dx), float64(dz)))
	step := float32(def.MoveDistance)
	if dist <= step {
		if !IsMapWalkable(mapID, p.X, p.Z) {
			n.nextMoveAt = now.Add(interval + 10*time.Second)
			return false
		}
		n.PosX, n.PosY, n.PosZ, n.Angle = p.X, p.Y, p.Z, p.Angle
		if n.backward {
			n.patrolIndex--
			if n.patrolIndex < 0 {
				n.patrolIndex, n.backward = 1, false
			}
		} else {
			n.patrolIndex++
			if n.patrolIndex >= len(n.patrol) {
				n.patrolIndex, n.backward = len(n.patrol)-2, true
			}
		}
		n.nextMoveAt = now.Add(time.Duration(rand.IntN(10000))*time.Millisecond + interval + 10*time.Second)
		return true
	}
	n.nextMoveAt = now.Add(interval)
	x, z := n.PosX+dx/dist*step, n.PosZ+dz/dist*step
	if !IsMapWalkable(mapID, x, z) {
		return false
	}
	n.Angle = npcDirection(dx, dz)
	n.PosX, n.PosZ = x, z
	n.PosY += (p.Y - n.PosY) / dist * step
	return true
}

// mobGroup is stMobRespawn: an svmap area entry of Count mobs. The first
// corpse removal from a full group starts its timer (RespawnTime, +-60 min for
// Attackspecial3 >= 9, CZone::UpdateMobDeath); CZone::UpdateMobRespawn then
// brings back one mob each time the timer runs out, at a random spot of
// the area, restarting the timer.
type mobGroup struct {
	count, alive int
	respawnAt    time.Time
}

// died books a death (shard lock held).
func (m *MapMob) died(now time.Time) {
	m.stoppedOnDeath = m.movingAt(now)
	m.setPoint(m.positionAt(now))
	m.cell = cellFor(m.PosX, m.PosZ)
	m.moveDuration = 0
	m.Effects, m.diedAt, m.corpseGone = nil, now, false
	if m.lua.actor != nil {
		h := &mobAIHost{mob: m, shard: shard(m.aiMapID), mapID: m.aiMapID, now: now, tick: &MobCombatTick{}}
		h.event("death", m.lua.attackedCount)
	}
	if m.group == nil && m.respawnDelay > 0 {
		m.respawnAt = now.Add(m.configuredRespawnDelay(m.aiMapID))
	}
}

// removeCorpse books the native UpdateMobDeath removal (shard lock held).
// Group counts include corpses until this point; later deaths cannot start
// another timer while the group is still being replenished.
func (m *MapMob) removeCorpse(now time.Time) {
	if m.corpseGone {
		return
	}
	m.corpseGone = true
	g := m.group
	if g == nil || m.event != nil {
		return
	}
	if g.alive == g.count {
		delay := m.configuredRespawnDelay(m.aiMapID)
		if m.def != nil && m.def.Attacks[2].Special >= 9 && delay > 0 {
			delay += time.Duration(rand.IntN(120)-60) * time.Minute
		}
		g.respawnAt = now.Add(max(delay, 0))
	}
	g.alive--
}

// respawnSpot is UpdateMobRespawn's position: lower + rand(width) on each
// axis; ok false when it is not walkable (the next update tries again).
func (m *MapMob) respawnSpot(mapID uint16) (float32, float32, bool) {
	if m.Area == nil {
		return m.homeX, m.homeZ, true
	}
	a := m.Area.normalized()
	w, d := int(a.UpperX-a.LowerX), int(a.UpperZ-a.LowerZ)
	x, z := a.LowerX, a.LowerZ
	if w > 0 {
		x += float32(rand.IntN(w))
	}
	if d > 0 {
		z += float32(rand.IntN(d))
	}
	return x, z, IsMapWalkable(mapID, x, z)
}

func (n *MapNpc) relocate(mapID uint16, motion uint8) entityMove {
	from, fresh := n.cell, !n.hasCell
	n.cell, n.hasCell = cellAfterMove(n.cell, n.hasCell, n.PosX, n.PosZ), true
	return entityMove{MapID: mapID, Kind: viewNpc, ID: n.ID, From: from, To: n.cell, Fresh: fresh,
		Changed: fresh || from != n.cell, Move: buildNpcMove(n, motion), Enter: buildNpcEnter(n)}
}

// mobCorpseDelay is CMob::Damage's death tick: the corpse stays 5 s.
const mobCorpseDelay = 5 * time.Second

// visibleMob: alive, or dead with its corpse still shown.
func (m *MapMob) visible() bool { return m.CurrentHP > 0 || (!m.diedAt.IsZero() && !m.corpseGone) }

// TickMobCorpses is CZone::UpdateMobDeath: corpses older than 5 s leave
// the view (0x0602).
func TickMobCorpses() []entityMove {
	now := time.Now()
	var out []entityMove
	for _, mapID := range shardIDs() {
		s := shard(mapID)
		s.mu.Lock()
		for i := range s.entities.Mobs {
			m := &s.entities.Mobs[i]
			if m.CurrentHP == 0 && !m.diedAt.IsZero() && !m.corpseGone && now.Sub(m.diedAt) >= mobCorpseDelay {
				m.removeCorpse(now)
				out = append(out, entityMove{MapID: mapID, Kind: viewMob, ID: m.ID})
			}
		}
		s.mu.Unlock()
	}
	return out
}

// SetMobAggro is CMob::SetTarget for a player's hit: an idle mob takes the
// attacker as its target. The ledger grows through AddMobAggro, with the
// original amount (hitAggro): the attacker's REC times ten plus the skill's
// Arrow, whatever the damage. CMob::AttackAI later turns to the top of the
// ledger (ReTargetByAggro).
func SetMobAggro(mapID uint16, id, characterID uint32) { AddMobAggro(mapID, id, characterID, 0) }

func AddMobAggro(mapID uint16, id, characterID uint32, amount int64) {
	s := shard(mapID)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.entities.Mobs {
		if m := &s.entities.Mobs[i]; m.ID == id && m.CurrentHP > 0 {
			if m.aggro == nil {
				m.aggro = map[uint32]int64{}
			}
			m.aggro[characterID] += amount
			if m.AggroTarget == 0 {
				m.AggroTarget, m.returningHome = characterID, false
			}
		}
	}
}

// ForceMobTarget is CMob::SetTarget(user, true): the mob turns to the user at
// once; the next ReTargetByAggro may still pick whoever has more aggro.
func ForceMobTarget(mapID uint16, id, characterID uint32) {
	s := shard(mapID)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.entities.Mobs {
		if m := &s.entities.Mobs[i]; m.ID == id && m.CurrentHP > 0 {
			m.AggroTarget, m.returningHome = characterID, false
		}
	}
}

// ClearMobAggro is CMob::Attack after killing its target: back to idle
// (status 0) with the dead user's aggro gone; UpdateIdle walks it home
// when it is more than 10 units away.
func ClearMobAggro(mapID uint16, id uint32) {
	s := shard(mapID)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.entities.Mobs {
		if m := &s.entities.Mobs[i]; m.ID == id {
			delete(m.aggro, m.AggroTarget)
			m.dropTarget(false)
		}
	}
}

// Mob combat as CMob::UpdateIdle / UpdateAttack / Chase / ReturnHome run it.
const (
	mobLeash    = 45.0 // target farther than this from home or from the mob
	mobHomeDone = 10.0 // ReturnHome completes within 10 units of home
)

// TickMobCombat advances targeting, attacks, chasing and returning home.
func TickMobCombat(players []PlayerCombatSnapshot, maxPerMap int) MobCombatTick {
	var tick MobCombatTick
	now := time.Now()
	byMap := map[uint16][]PlayerCombatSnapshot{}
	for _, p := range players {
		byMap[p.MapID] = append(byMap[p.MapID], p)
	}
	for mapID, mapPlayers := range byMap {
		s := shard(mapID)
		s.mu.Lock()
		actions := 0
		for i := range s.entities.Mobs {
			if actions >= maxPerMap {
				break
			}
			m := &s.entities.Mobs[i]
			if m.CurrentHP == 0 {
				m.AggroTarget, m.returningHome, m.aggro = 0, false, nil
				continue
			}
			h := &mobAIHost{mob: m, shard: s, mapID: mapID, now: now, players: mapPlayers, tick: &tick}
			if !m.ensureAI(h) {
				continue
			}
			h.notifyTarget()
			wasReturning := m.returningHome
			before := len(tick.Attacks) + len(tick.Moves)
			if !h.event("idle") {
				continue
			}
			h.notifyTarget()
			if !wasReturning && m.returningHome {
				h.event("return_home", m.lua.attackedCount)
			}
			if m.lua.fixed && m.lua.arrived {
				m.lua.fixed = false
				h.event("move_end")
			}
			if m.lua.needReset {
				m.lua.needReset = false
				m.lua.lastTarget = 0
				m.lua.attacked = false
				m.lua.mask = 15
				h.event("reset")
			}
			h.event("combat", int(int64(m.CurrentHP)*100/int64(max(m.MaxHP, 1))), m.lua.attackedCount)
			m.refreshMoveSpeed(mapID, now, &tick)
			if len(tick.Attacks)+len(tick.Moves) > before {
				actions++
			}
		}
		s.mu.Unlock()
	}
	runMobAICommands(&tick)
	return tick
}

func (m *MapMob) liveTarget(players []PlayerCombatSnapshot) *PlayerCombatSnapshot {
	if m.AggroTarget == 0 {
		return nil
	}
	for k := range players {
		if players[k].CharacterID == m.AggroTarget && players[k].Alive && !players[k].Invisible {
			return &players[k]
		}
	}
	return nil
}

// dropTarget: status 0 (idle) or 3 (return home).
func (m *MapMob) dropTarget(home bool) {
	m.AggroTarget, m.lastAttackAt, m.nextAttackAt = 0, time.Time{}, time.Time{}
	m.returningHome = home
}

// findTarget is CMob::FindTarget for AI 1 and 3: the first living user of
// another country within ChaseRange, cell by cell over the 3x3 cells
// around the mob.
func (m *MapMob) findTarget(now time.Time, players []PlayerCombatSnapshot, extended bool) *PlayerCombatSnapshot {
	if m.def == nil || m.ChaseRange <= 0 || now.Before(m.noTargetUntil) {
		return nil
	}
	radius := int32(1)
	if extended {
		radius = 2
	}
	pos := m.positionAt(now)
	for dz := -radius; dz <= radius; dz++ {
		for dx := -radius; dx <= radius; dx++ {
			c := cell{m.cell.x + dx, m.cell.z + dz}
			for k := range players {
				p := &players[k]
				if p.Cell != c || !p.Alive || p.Invisible || (m.def.Country < 2 && p.Faction == m.def.Country) {
					continue
				}
				if DistanceXZ(p.PosX, p.PosZ, pos[0], pos[2]) <= m.ChaseRange {
					m.AggroTarget = p.CharacterID
					return p
				}
			}
		}
	}
	return nil
}

// targetWeakest is CMob::FindAttackUserMinHP over the damage ledger.
func (m *MapMob) targetWeakest(players []PlayerCombatSnapshot) *PlayerCombatSnapshot {
	var best *PlayerCombatSnapshot
	for k := range players {
		p := &players[k]
		if _, hit := m.aggro[p.CharacterID]; !hit || !p.Alive || p.Invisible || (m.def.Country < 2 && p.Faction == m.def.Country) {
			continue
		}
		if best == nil || p.HP < best.HP {
			best = p
		}
	}
	return best
}

// retargetByAggro is CMob::ReTargetByAggro: the living attacker on the map
// with the most aggro.
func (m *MapMob) retargetByAggro(players []PlayerCombatSnapshot) {
	var best uint32
	var top int64 = -1
	for k := range players {
		p := &players[k]
		if v, ok := m.aggro[p.CharacterID]; ok && p.Alive && !p.Invisible && v > top {
			best, top = p.CharacterID, v
		}
	}
	if best != 0 {
		m.AggroTarget = best
	}
}

// stepChase advances toward the target, stopping just inside attack reach.
func (m *MapMob) stepChase(now time.Time, mapID uint16, t *PlayerCombatSnapshot, d float32, mods *mobMods, tick *MobCombatTick) bool {
	if m.movingAt(now) || now.Before(m.nextChaseAt) || !mods.canMove() || m.def == nil || m.stationary() {
		return false
	}
	step := m.chaseStep(mods)
	if m.canReachTarget(t) {
		if d <= m.attackReach() {
			return false
		}
		step = min(step, d-(m.attackReach()-.1))
	}
	if m.nav != nil {
		next, ok := m.navigate(now, mapID, mapnav.Point{t.PosX, t.PosY, t.PosZ}, step)
		if !ok {
			m.nextChaseAt = now.Add(time.Second)
			m.dropTarget(true)
			m.noTargetUntil = now.Add(mobAttackIdle)
			return false
		}
		tick.Moves = append(tick.Moves, m.moveTo(mapID, now, next, mobMotionRun))
		return true
	}
	nx := m.PosX + (t.PosX-m.PosX)/d*step
	nz := m.PosZ + (t.PosZ-m.PosZ)/d*step
	if !IsMapWalkable(mapID, nx, nz) {
		// No straight line (the original then paths around walls; not
		// modeled): give up and return home.
		m.dropTarget(true)
		return false
	}
	tick.Moves = append(tick.Moves, m.moveTo(mapID, now, mapnav.Point{nx, m.PosY, nz}, mobMotionRun))
	return true
}

// stepHome runs toward the spawn and clears effects and aggro on arrival.
// CMob::ReturnHome sends movement mode 1 with the chase step and time.
func (m *MapMob) stepHome(now time.Time, mapID uint16, mods *mobMods, tick *MobCombatTick) bool {
	if m.movingAt(now) || now.Before(m.nextChaseAt) {
		return false
	}
	if m.nav != nil {
		goal := mapnav.Point{m.homeX, m.homeY, m.homeZ}
		if mapnav.Distance(m.point(), goal) < .1 {
			m.returningHome, m.lastAttackAt, m.Effects, m.aggro = false, time.Time{}, nil, nil
			return false
		}
		if !mods.canMove() {
			return false
		}
		step := m.chaseStep(mods)
		next, ok := m.navigate(now, mapID, goal, step)
		if !ok {
			m.nextChaseAt = now.Add(time.Second)
			return false
		}
		tick.Moves = append(tick.Moves, m.moveTo(mapID, now, next, mobMotionRun))
		return true
	}
	d := DistanceXZ(m.PosX, m.PosZ, m.homeX, m.homeZ)
	if d < mobHomeDone {
		m.returningHome, m.lastAttackAt, m.Effects, m.aggro = false, time.Time{}, nil, nil
		return false
	}
	if !mods.canMove() {
		return false
	}
	step := m.chaseStep(mods)
	next := mapnav.Point{m.homeX, m.homeY, m.homeZ}
	if d-(float32(m.def.Size)+1) >= step {
		nx := m.PosX + (m.homeX-m.PosX)/d*step
		nz := m.PosZ + (m.homeZ-m.PosZ)/d*step
		if IsMapWalkable(mapID, nx, nz) {
			next = mapnav.Point{nx, m.PosY, nz}
		}
	}
	tick.Moves = append(tick.Moves, m.moveTo(mapID, now, next, mobMotionRun))
	return true
}

func AddEntities(mapID uint16, e MapEntities) {
	s := shard(mapID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entities.NPCs = append(s.entities.NPCs, e.NPCs...)
	if len(e.NPCs) > 0 {
		s.npcRevision.Add(1)
	}
	for _, mob := range e.Mobs {
		if mob.MobID != 0 {
			s.questMobSpawns = append(s.questMobSpawns, proto.QuestMobSpawn{MobID: mob.MobID, X: mob.PosX, Z: mob.PosZ})
		}
	}
	if len(e.Mobs) > 0 {
		s.invalidateQuestMobCatalog()
	}
	s.entities.Mobs = append(s.entities.Mobs, e.Mobs...)
	s.entities.Portals = append(s.entities.Portals, e.Portals...)
	s.entities.SpawnAreas = append(s.entities.SpawnAreas, e.SpawnAreas...)
}

// FindNearestSpawnArea is CMap::GetBindPos's group choice: the nearest
// non-empty group open to the faction; when none is, the first group (the
// original's index stays 0), and ok is false only if that group is empty.
func FindNearestSpawnArea(mapID uint16, faction uint8, x, z float32) (MapSpawnArea, bool) {
	areas := EntitiesOnMap(mapID).SpawnAreas
	best, bestDist := -1, float32(math.Inf(1))
	for i, a := range areas {
		if len(a.Boxes) == 0 || !a.allowsFaction(faction) {
			continue
		}
		if d := float32(math.Sqrt(float64(a.distanceSq(x, z)))); d < bestDist {
			best, bestDist = i, d
		}
	}
	if best < 0 {
		return MapSpawnArea{}, false
	}
	return areas[best], true
}

func FindPortalByIndex(mapID uint16, index uint8) (MapPortal, bool) {
	for _, p := range EntitiesOnMap(mapID).Portals {
		if p.Index == index {
			return p, true
		}
	}
	return MapPortal{}, false
}

func FindPortalAt(mapID uint16, x, y, z float32) (MapPortal, bool) {
	var legacy MapPortal
	foundLegacy := false
	for _, p := range EntitiesOnMap(mapID).Portals {
		if !p.Contains(x, y, z) {
			continue
		}
		if p.volume != nil {
			return p, true
		}
		if !foundLegacy {
			legacy, foundLegacy = p, true
		}
	}
	return legacy, foundLegacy
}

func NewPatrolNpc(npcType uint8, typeID uint16, points []NpcPatrolPoint) (MapNpc, bool) {
	if len(points) == 0 || npcquest.RetiredType(npcType) {
		return MapNpc{}, false
	}
	p := points[0]
	return MapNpc{ID: realm.nextNpcID.Add(1) - 1, Type: npcType, TypeID: typeID, PosX: p.X, PosY: p.Y, PosZ: p.Z,
		Angle: p.Angle, patrol: points, cell: cellFor(p.X, p.Z), hasCell: true, patrolIndex: min(1, len(points)-1),
		nextMoveAt: time.Now().Add(time.Duration(10000+rand.IntN(10000)) * time.Millisecond)}, true
}

// NewStaticMob builds a mob from its definition; respawn 0 never respawns.
func NewStaticMob(def *db.MobDef, mobID uint16, x, z float32, area *MobArea, respawnSeconds uint64) MapMob {
	m := MapMob{ID: realm.nextMobID.Add(1) - 1, MobID: mobID, Level: 1, PosX: x, PosZ: z, homeX: x, homeZ: z,
		MaxHP: 100, respawnDelay: time.Duration(respawnSeconds) * time.Second}
	if def != nil {
		m.Level, m.MaxHP, m.MoneyMin, m.MoneyMax = def.Level, max(def.MaxHP, 1), def.MoneyMin, def.MoneyMax
		m.Exp, m.Attack, m.AI, m.ChaseRange = def.Exp, def.Attack, def.AI, def.ChaseRange
		m.def = def
	}
	// CMob::Init: the first wander comes within a minute.
	m.nextMoveAt = time.Now().Add(time.Duration(rand.IntN(60000)) * time.Millisecond)
	m.CurrentHP = m.MaxHP
	if m.def != nil {
	}
	m.resetAI()
	m.cell, m.hasCell = cellFor(x, z), true
	m.noTargetUntil = time.Now().Add(5 * time.Second)
	if area != nil {
		a := area.normalized()
		m.Area = &a
	}
	return m
}

func (m *MapMob) pickRespawnPosition(mapID uint16) (float32, float32) {
	if m.Area == nil {
		return m.homeX, m.homeZ
	}
	a := m.Area.normalized()
	m.roamStep++
	for attempt := uint32(0); attempt < 64; attempt++ {
		seed := m.ID + m.roamStep*97 + attempt
		x := a.LowerX + (a.UpperX-a.LowerX)*float32(seed%997)/996
		z := a.LowerZ + (a.UpperZ-a.LowerZ)*float32((seed/997)%991)/990
		if roamWalkable(mapID, x, z, a) {
			return x, z
		}
	}
	x, z := a.clampX(m.homeX), a.clampZ(m.homeZ)
	if roamWalkable(mapID, x, z, a) {
		return x, z
	}
	return m.homeX, m.homeZ
}

// mobSpeedOffsets are the original move-speed step offsets (0x00542538);
// mobs start at speed 2 (no offset).
var mobSpeedOffsets = [5]float32{-4, -2, 0, 2, 4}

const mobHomeRadius = 10.0

// advanceRoam is the idle state's wander (update_mob_ai_idle): every
// 10 s + NormalTime + up to 30 s, a step of NormalStep + 0..4 units in a
// random direction, staying within 10 units of home.
func (m *MapMob) advanceRoam(now time.Time, mapID uint16, tick *MobCombatTick) {
	if m.def == nil || m.movingAt(now) || now.Before(m.nextMoveAt) {
		return
	}
	if DistanceXZ(m.PosX, m.PosZ, m.homeX, m.homeZ) > mobHomeRadius {
		m.returningHome = true // CMob::UpdateIdle: status 3
		return
	}
	pause := 10*time.Second + time.Duration(rand.IntN(30000))*time.Millisecond
	if m.def.NormalStep == 0 {
		m.nextMoveAt = now.Add(pause)
		return
	}
	m.nextMoveAt = now.Add(pause + time.Duration(m.def.NormalTimeMS)*time.Millisecond)
	mods := m.mods()
	length := float32(m.def.NormalStep) + mobSpeedOffsets[min(max(2+mods.MoveSpeed, 0), 4)] + float32(rand.IntN(500)/100)
	angle := float64(rand.IntN(10000)) * math.Pi / 180
	nx := m.PosX - float32(math.Sin(angle))*length
	nz := m.PosZ + float32(math.Cos(angle))*length
	if DistanceXZ(nx, nz, m.homeX, m.homeZ) >= mobHomeRadius || !IsMapWalkable(mapID, nx, nz) {
		return
	}
	next := mapnav.Point{nx, m.PosY, nz}
	if m.nav != nil {
		var ok bool
		next, ok = m.navigate(now, mapID, next, length)
		if !ok {
			return
		}
	}
	tick.Moves = append(tick.Moves, m.moveTo(mapID, now, next, mobMotionWalk))
}

// chaseStep is the maximum path segment for a chase or return home.
func (m *MapMob) chaseStep(mods *mobMods) float32 {
	if m.def == nil {
		return 1
	}
	idx := min(max(2+mods.MoveSpeed, 0), 4)
	return max(float32(m.def.ChaseStep)+mobSpeedOffsets[idx], 1)
}

// attackReach is vertical_mobs_attack_range: Size + Attackrange + 0.3.
func (m *MapMob) attackReach() float32 {
	if m.def == nil {
		return MobAttackRange
	}
	return float32(m.def.Size) + float32(m.def.Attacks[0].Range) + 0.3
}

// mobAttackAnimation caps how long a mob holds still while an attack plays.
const mobAttackAnimation = time.Second

func later(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}

// mobAttackIdle is AttackAI's wait when no slot is ready (2500 ms plus
// g_ConstMobAttackSpeed, 0 at normal speed).
const mobAttackIdle = 2500 * time.Millisecond

// pickAttackSlot keeps CMob::AttackAI's slot order and eligibility checks.
// GX draws once per decision for routine spells. A declined spell remains
// ready and the mob can still choose a basic or conditional attack.
func (m *MapMob) pickAttackSlot(now time.Time, mods *mobMods, inReach func(int, float32) bool, rollSpell func() bool) (slotIndex uint8, found, declinedSpell bool) {
	if m.def == nil {
		return 0, true, false
	}
	maxHP := float64(float32(m.MaxHP))
	threshold := func(f float32) uint32 { return uint32(int32(maxHP * float64(f))) }
	for slot := 2; slot >= 0; slot-- {
		a := &m.def.Attacks[slot]
		if a.OK == 0 || now.Before(m.slotReadyAt[slot]) {
			continue
		}
		switch a.OK {
		case 3:
			if m.CurrentHP > threshold(0.4) || m.CurrentHP < threshold(0.2) {
				continue
			}
		case 4:
			if m.CurrentHP > threshold(0.05) {
				continue
			}
		case 5:
			if m.CurrentHP >= threshold(0.2) {
				continue
			}
		}
		reach := float32(m.def.Size) + float32(a.Range) + 0.3
		routineSpell := false
		if a.Type == 0 && mods.Disabled[0] > 0 {
			continue // CMob::Attack: m_pAttack[0].nDisable
		}
		if a.Type != 0 {
			if sk := db.MustDefs().NPCSkill(a.Type); sk != nil {
				if mods.Disabled[sk.Kind] > 0 {
					continue
				}
				routineSpell = a.OK < 3
				if sk.TypeEffect == 2 {
					reach = float32(math.Inf(1)) // Self-heal has no target range.
				} else {
					reach = float32(m.def.Size) + float32(sk.AttackRange) + 0.3
				}
			}
		}
		if !inReach(slot, reach) {
			continue
		}
		if routineSpell && (declinedSpell || !rollSpell()) {
			declinedSpell = true
			continue
		}
		return uint8(slot), true, declinedSpell
	}
	return 0, false, declinedSpell
}

// useSlot starts a slot's cooldown and the next attack time. A mob heal
// skill (TypeEffect 2) applies at once and returns its HP packet.
func (m *MapMob) useSlot(slot uint8, now time.Time) []byte {
	a := &m.def.Attacks[slot]
	if a.Type != 0 {
		if sk := db.MustDefs().NPCSkill(a.Type); sk != nil {
			m.slotReadyAt[slot] = now.Add(time.Duration(sk.ResetTimeSeconds) * time.Second)
			m.nextAttackAt = now.Add(mobAttackIdle + m.mods().attackDelayOffset())
			if sk.TypeEffect == 2 {
				m.CurrentHP = min(satAdd(m.CurrentHP, uint32(max(sk.HealHP, 0))), m.MaxHP)
				return buildMobCurrentHP(m.ID, m.CurrentHP)
			}
			return nil
		}
	}
	// CMob::Attack: the slot waits its own AttackTime, the mob slot 1's.
	offset := m.mods().attackDelayOffset()
	m.slotReadyAt[slot] = now.Add(time.Duration(float64(time.Duration(a.TimeMS)*time.Millisecond+offset) / float64(m.attackSpeedFactor())))
	m.nextAttackAt = now.Add(time.Duration(float64(m.attackInterval()+offset) / float64(m.attackSpeedFactor())))
	return nil
}

func (m *MapMob) attackSpeedFactor() float32 {
	if m.lua.speed > 0 {
		return m.lua.speed
	}
	return 1
}

// movable: not stunned, asleep or stopped.
func (m *MapMob) movable() bool {
	o := m.mods()
	return o.canMove()
}

func (m *MapMob) attackInterval() time.Duration {
	if m.def == nil || m.def.Attacks[0].TimeMS == 0 {
		return MobAttackInterval
	}
	return time.Duration(m.def.Attacks[0].TimeMS) * time.Millisecond
}

func roamWalkable(mapID uint16, x, z float32, a MobArea) bool {
	return a.contains(x, z) && IsMapWalkable(mapID, x, z)
}

// Packet builders.

func buildNpcEnter(n *MapNpc) []byte {
	w := packet.New(packet.MapNpcEnter)
	w.U32(n.ID)
	w.U8(n.Type)
	w.U16(n.TypeID)
	w.F32(n.PosX)
	w.F32(n.PosY)
	w.F32(n.PosZ)
	w.U16(n.Angle)
	if n.Summoned {
		w.Bool(true)
	}
	return w.MustFinish()
}

func buildNpcLeave(id uint32) []byte {
	w := packet.New(packet.MapNpcLeave)
	w.U32(id)
	return w.MustFinish()
}

func buildNpcMove(n *MapNpc, motion uint8) []byte {
	w := packet.New(packet.MapNpcMove)
	w.U32(n.ID)
	w.U8(motion)
	w.F32(n.PosX)
	w.F32(n.PosY)
	w.F32(n.PosZ)
	return w.MustFinish()
}

func (m *MapMob) wireHeightKnown() bool {
	return m.nav != nil || m.event != nil && m.spawnHeightKnown
}

func buildMobEnter(m *MapMob, isNew bool) []byte {
	op := packet.MobEnter
	heightKnown := m.wireHeightKnown()
	if heightKnown || m.def != nil && m.def.Country < 2 {
		op = packet.Opcode(proto.GXMobEnter)
	}
	w := packet.New(op)
	w.U32(m.ID)
	w.Bool(isNew)
	w.U16(m.MobID)
	w.F32(m.PosX)
	if op == packet.Opcode(proto.GXMobEnter) {
		w.F32(m.PosY)
	}
	w.F32(m.PosZ)
	if m.def != nil && m.def.Country < 2 {
		w.U8(m.def.Country)
		w.Bool(heightKnown)
	}
	return w.MustFinish()
}

func buildMobMove(m *MapMob, motion uint8) []byte {
	op := packet.MobMove
	heightKnown := m.wireHeightKnown()
	if heightKnown {
		op = packet.Opcode(proto.GXMobMove)
	}
	w := packet.New(op)
	w.U32(m.ID)
	w.U8(motion)
	w.F32(m.PosX)
	if heightKnown {
		w.F32(m.PosY)
	}
	w.F32(m.PosZ)
	if heightKnown {
		w.U32(uint32(m.moveDuration / time.Millisecond))
	}
	return w.MustFinish()
}

func buildMobLeave(id uint32) []byte {
	w := packet.New(packet.MobLeave)
	w.U32(id)
	return w.MustFinish()
}

// buildMobTargetHP (0x0305): mob id, HP (client reader 0x0058f2d0).
func buildMobTargetHP(m *MapMob) []byte {
	w := packet.New(packet.TargetMobHpUpdate)
	w.U32(m.ID)
	w.I32(int32(min(m.CurrentHP, math.MaxInt32)))
	return w.MustFinish()
}

// buildMobState (0x0304): mob id, HP, attack speed, move speed.
func buildMobState(m *MapMob) []byte {
	w := packet.New(packet.TargetMobGetState)
	w.U32(m.ID)
	w.I32(int32(min(m.CurrentHP, math.MaxInt32)))
	mods := m.mods()
	w.U8(mods.attackSpeedIndex())
	w.U8(mods.moveSpeedIndex())
	return w.MustFinish()
}

func buildCharacterMobAttack(characterID, mobID uint32, result uint8, damage uint16) []byte {
	w := packet.New(packet.CharacterMobAutoAttack)
	w.U8(result)
	w.U32(characterID)
	w.U32(mobID)
	w.U16(damage)
	w.U16(0)
	w.U16(0)
	return w.MustFinish()
}

func buildMobAttackCharacter(result uint8, mobID, characterID uint32, damage uint16) []byte {
	w := packet.New(packet.MobAttackCharacter)
	w.U8(result)
	w.U32(mobID)
	w.U32(characterID)
	w.U16(damage)
	w.U16(0)
	w.U16(0)
	return w.MustFinish()
}

func buildCharacterDeath(characterID uint32, deathType uint8, killerID uint32) []byte {
	w := packet.New(packet.CharacterDeath)
	w.U32(characterID)
	w.U8(deathType)
	w.U32(killerID)
	return w.MustFinish()
}

func buildCharacterDeathState(characterID uint32) []byte {
	w := packet.New(packet.CharacterDeathState)
	w.U32(characterID)
	w.U8(0x07)
	w.U8(0x02)
	return w.MustFinish()
}

func buildCharacterReviveInit(characterID uint32) []byte {
	w := packet.New(packet.CharacterReviveInit)
	w.U32(characterID)
	return w.MustFinish()
}

// buildCharacterResurrect (0x0551) is CUser::Rebirth's packet: id, kind
// (1 in place, 4 bind point, 5 party leader), current exp, position.
func buildCharacterResurrect(characterID uint32, kind uint8, exp uint32, x, y, z float32) []byte {
	w := packet.New(packet.CharacterResurrect)
	w.U32(characterID)
	w.U8(kind)
	w.U32(exp)
	w.F32(x)
	w.F32(y)
	w.F32(z)
	return w.MustFinish()
}

func buildCharacterReviveHitpoints(characterID, hp, mp, sp uint32) []byte {
	w := packet.New(packet.CharacterReviveHitpoints)
	w.U32(characterID)
	w.U32(hp)
	w.U32(mp)
	w.U32(sp)
	return w.MustFinish()
}

func buildMobDeath(mobID, killerID uint32) []byte {
	return buildMobDeathBy(mobID, killerID, 1)
}

func (m *MapMob) deathPositionPacket() []byte {
	if m.stoppedOnDeath && m.wireHeightKnown() {
		return buildMobMove(m, mobMotionWalk)
	}
	return nil
}

func buildMobDeathBy(mobID, killerID uint32, deathType uint8) []byte {
	w := packet.New(packet.MobDeath)
	w.U32(mobID)
	w.U8(deathType)
	w.U32(killerID)
	return w.MustFinish()
}

func buildMobPeriodicDamage(mobID uint32, skillID uint16, level uint8, hp, sp, mp uint16) []byte {
	w := packet.New(packet.MobPeriodicDamage)
	w.U32(mobID)
	w.U16(skillID)
	w.U8(level)
	w.U16(hp)
	w.U16(sp)
	w.U16(mp)
	return w.MustFinish()
}

func buildMobCurrentHP(mobID, hp uint32) []byte {
	w := packet.New(packet.MobCurrentHitpoints)
	w.U32(mobID)
	w.U32(hp)
	return w.MustFinish()
}

// HealMob restores HP to a living mob (a drain's heal); returns the mob.
func HealMob(mapID uint16, id, amount uint32) (MapMob, uint32, bool) {
	s := shard(mapID)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.entities.Mobs {
		m := &s.entities.Mobs[i]
		if m.ID == id && m.CurrentHP > 0 {
			before := m.CurrentHP
			m.CurrentHP = min(satAdd(m.CurrentHP, amount), m.MaxHP)
			return m.snapshotAt(time.Now()), m.CurrentHP - before, true
		}
	}
	return MapMob{}, 0, false
}

// recordLootDamage books damage against the attacker for the loot owner.
func (m *MapMob) recordLootDamage(id, damage uint32) {
	if id == 0 || damage == 0 {
		return
	}
	for i := range m.lootDamage {
		if m.lootDamage[i].ID == id {
			m.lootDamage[i].Amount += uint64(damage)
			return
		}
	}
	m.lootDamage = append(m.lootDamage, eventDamageEntry{id, uint64(damage)})
}
