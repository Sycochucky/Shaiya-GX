package world

import (
	"fmt"
	"math"
	"math/rand/v2"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/catalogfiles"
	"github.com/Spelunkern/Shaiya-GX/internal/mapgeometry"
	"github.com/Spelunkern/Shaiya-GX/internal/worlddrops"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
)

// Mob loot as the original server rolls it (CMob loot path, recovered by
// Eternity's mob_drops): MobItems orders 1-6 use a per-cent rate and 7-9 a
// per-million rate (order 0 is unused); each hit picks a random item of
// that grade, subject to the item's Drop time limit. World drops use a
// separate roll per eligible rule. Gold drops 30% of the time as
// Money1 + rand(Money2).

func chanceSucceeds(percent uint32) bool {
	switch {
	case percent == 0:
		return false
	case percent >= 100:
		return true
	}
	return rand.Uint32N(100) < percent
}

// dropLimitIntervals are the original Drop selectors (ms per window).
var dropLimitIntervals = [7]uint32{0, 1440000, 14400000, 28800000, 43200000, 86400000, 259200000}

type gradeEntry struct {
	def *db.ItemDef
	// Drop limit: at most one drop per window, the window start
	// randomized within its last tenth.
	limited          bool
	delay, randDelay time.Duration
	nextAt           time.Time
}

var dropCatalog struct {
	mu         sync.Mutex
	grades     map[uint16][]*gradeEntry
	worldRules []worlddrops.Rule
}

func isQuestItemType(t uint8) bool { return t == 27 || t == 28 || t == 29 }

// buildDropCatalog groups items by grade, unrestricted items first, then
// those with a Server or Drop restriction, as CGameData's init does.
func buildDropCatalog() {
	defs := db.MustDefs()
	var all []*db.ItemDef
	for _, d := range defs.Items {
		if d.Grade > 0 && d.Grade <= 1024 && d.Type != 0 && d.TypeID != 0 && !isQuestItemType(d.Type) {
			all = append(all, d)
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Type != all[j].Type {
			return all[i].Type < all[j].Type
		}
		return all[i].TypeID < all[j].TypeID
	})
	grades := map[uint16][]*gradeEntry{}
	for pass := 0; pass < 2; pass++ {
		for _, d := range all {
			restricted := (d.Server != 0 && d.Server < 10) || d.Drop != 0
			if restricted != (pass == 1) {
				continue
			}
			e := &gradeEntry{def: d}
			if d.Drop != 0 && int(d.Drop) < len(dropLimitIntervals) {
				interval := time.Duration(dropLimitIntervals[d.Drop]) * time.Millisecond
				e.limited, e.randDelay = true, interval/10
				e.delay = interval - e.randDelay
			}
			grades[d.Grade] = append(grades[d.Grade], e)
		}
	}
	dropCatalog.mu.Lock()
	dropCatalog.grades = grades
	dropCatalog.mu.Unlock()
}

func loadWorldDrops(log *logx.Logger) error {
	rules, err := catalogfiles.Read(filepath.Join(gameDataDir(), catalogfiles.WorldDrops), catalogfiles.ParseWorldDrops)
	if err != nil {
		return err
	}
	dropCatalog.mu.Lock()
	dropCatalog.worldRules = rules
	dropCatalog.mu.Unlock()
	log.Printf("GX Server World loaded WorldDrops.ini rules=%d", len(rules))
	return nil
}

// pickGradeItem is MobDropCatalog::select: a random item of the grade,
// dropped only if its time limit allows it.
func pickGradeItem(grade uint16) *db.ItemDef {
	dropCatalog.mu.Lock()
	defer dropCatalog.mu.Unlock()
	list := dropCatalog.grades[grade]
	if len(list) == 0 {
		return nil
	}
	return takeCatalogDrop(list[rand.IntN(len(list))])
}

func pickExactDrop(def *db.ItemDef) *db.ItemDef {
	dropCatalog.mu.Lock()
	defer dropCatalog.mu.Unlock()
	for _, e := range dropCatalog.grades[def.Grade] {
		if e.def.Type == def.Type && e.def.TypeID == def.TypeID {
			return takeCatalogDrop(e)
		}
	}
	return nil
}

func takeCatalogDrop(e *gradeEntry) *db.ItemDef {
	if e.limited {
		now := time.Now()
		if now.Before(e.nextAt) {
			return nil
		}
		jitter := time.Duration(0)
		if e.randDelay > 0 {
			jitter = time.Duration(rand.Int64N(int64(e.randDelay)))
		}
		e.nextAt = now.Add(e.delay + jitter)
	}
	return e.def
}

func rollWorldDrops(level uint16, percent uint32, roll func(uint64) uint64, drop func(worlddrops.Rule)) {
	dropCatalog.mu.Lock()
	rules := dropCatalog.worldRules
	dropCatalog.mu.Unlock()
	for _, rule := range rules {
		if level >= rule.MinLevel && level <= rule.MaxLevel && rule.DropChance > 0 &&
			rollScaledDrop(rule.DropChance, worlddrops.RollRange, percent, roll) {
			drop(rule)
		}
	}
}

// dropPosition scatters a drop within half a unit along each horizontal axis.
func dropPosition(x, z float32) (float32, float32) {
	return x + float32(rand.IntN(1000))/1000 - 0.5, z + float32(rand.IntN(1000))/1000 - 0.5
}

// mobDropPosition keeps CMob::GetDropPosRand's scatter around the monster and
// uses its authoritative height to select the floor.
func mobDropPosition(mob *MapMob) (x, y, z float32) {
	origin := mapgeometry.Position{mob.PosX, mob.PosY, mob.PosZ}
	candidate := origin
	candidate[0], candidate[2] = dropPosition(mob.PosX, mob.PosZ)
	if mob.nav != nil {
		nav, err := loadMobNavigation(mob.nav.id)
		if err != nil || nav == nil {
			return origin[0], origin[1], origin[2]
		}
		if g, ok := nav.PlayerGeometry(); ok {
			candidate = floorMobDrop(g, origin, candidate)
		}
	}
	return candidate[0], candidate[1], candidate[2]
}

func floorMobDrop(g *mapgeometry.Geometry, origin, candidate mapgeometry.Position) mapgeometry.Position {
	y, ok := g.Floor(candidate, mapgeometry.MaxGroundStep, false)
	if !ok {
		return origin
	}
	candidate[1] = y
	// Test slightly above the floor so an offset cannot cross a wall.
	from, to := origin, candidate
	from[1] += mapgeometry.FootSkin
	to[1] += mapgeometry.FootSkin
	if !g.SegmentVisible(from, to) {
		return origin
	}
	return candidate
}

// rollMobDrops rolls one kill's loot, registers it on the map and returns
// the new map items.
//
// With two or more party members in range, the gold is split evenly among
// them instead of dropped (at least 1 each).
func rollMobDrops(mapID uint16, mob *MapMob, ownerID, killerID uint32, party []*ConnectedPlayer, rates config.RewardRates, log logPrinter) []MapItem {
	defs := db.MustDefs()
	var packets []MapItem
	place := func(def *db.ItemDef, money uint32) {
		x, y, z := mobDropPosition(mob)
		if def != nil {
			activityForMap(realMap(mapID)).drops.Add(1)
		}
		if def != nil && len(party) >= 1 {
			// Universal loot: a party or raid's item goes straight into one member's bag.
			item := newDroppedItem(def)
			if to := distributeLoot(party, &item, time.Now()); to != nil {
				to.QueueLoot(lootGrant{item: item, mapID: mapID, x: x, y: y, z: z})
				return
			}
		}
		mi := MapItem{ID: nextMapItemID(), OwnerID: ownerID, FromMob: true, DroppedAt: time.Now(), PosX: x, PosY: y, PosZ: z}
		if def != nil {
			mi.Item = newDroppedItem(def)
		} else {
			mi.Item, mi.Money = db.Item{Type: moneyItemType, Count: 1}, money
		}
		addMapItem(mapID, mi)
		packets = append(packets, mi)
	}
	for _, d := range defs.MobDrops[mob.MobID] {
		if d.ItemOrder < 1 || d.ItemOrder > 9 {
			continue
		}
		if !rollMonsterDrop(d, rates.MobDropPercent, rand.Uint64N) {
			continue
		}
		if def := pickGradeItem(d.Grade); def != nil {
			place(def, 0)
			logActivity(log, "loot.drop", fmt.Sprintf("%s dropped (monster %d)", def.Name, mob.MobID),
				"map_id", realMap(mapID), "character_id", ownerID, "mob_id", mob.MobID, "item_type", def.Type, "item_id", def.TypeID, "source", "monster")
			log.Printf("GX Server World mob drop mob_id=%d order=%d grade=%d item=%d/%d", mob.MobID, d.ItemOrder, d.Grade, def.Type, def.TypeID)
		}
	}
	rollWorldDrops(mob.Level, rates.WorldDropPercent, rand.Uint64N, func(rule worlddrops.Rule) {
		if def := pickGradeItem(rule.Grade); def != nil {
			place(def, 0)
			logActivity(log, "loot.drop", fmt.Sprintf("%s dropped (event rule %d)", def.Name, rule.Index),
				"map_id", realMap(mapID), "character_id", ownerID, "mob_id", mob.MobID, "item_type", def.Type, "item_id", def.TypeID, "source", "event", "rule", rule.Index)
			log.Printf("GX Server World world drop mob_id=%d rule=%d grade=%d item=%d/%d", mob.MobID, rule.Index, rule.Grade, def.Type, def.TypeID)
		}
	})
	rollScopedDrops(mapID, mob, ownerID, rates.WorldDropPercent, place, log)
	if grbKill(mapID, mob, killerID) {
		return packets
	}
	if amount := rollMonsterGold(mob, rates.MobGoldPercent, rates.MobGoldChance, rand.Uint32N); amount > 0 {
		if len(party) > 1 {
			share := max(amount/uint32(len(party)), 1)
			for _, m := range party {
				m.QueueGold(share)
			}
		} else {
			place(nil, amount)
		}
	}
	return packets
}

func rollMonsterDrop(drop db.MobDrop, percent uint32, roll func(uint64) uint64) bool {
	var span uint64 = 100
	if drop.ItemOrder >= 7 {
		span = 1_000_000
	}
	return rollScaledDrop(uint64(drop.DropRate), span, percent, roll)
}

func rollScaledDrop(chance, span uint64, percent uint32, roll func(uint64) uint64) bool {
	// Expand the native range so fractional probabilities survive scaling.
	span *= 100
	return roll(span) < min(chance*uint64(percent), span)
}

func rollMonsterGold(mob *MapMob, percent, chance uint32, roll func(uint32) uint32) uint32 {
	if roll(100) >= chance {
		return 0
	}
	amount := uint64(mob.MoneyMin)
	if mob.MoneyMax > 0 {
		amount += uint64(roll(mob.MoneyMax))
	}
	return uint32(min(amount*uint64(percent)/100, math.MaxUint32))
}

func newDroppedItem(def *db.ItemDef) db.Item {
	it := db.Item{Type: def.Type, TypeID: def.TypeID, Quality: def.Quality, Count: 1}
	if def.Server > 0 && def.ReqWis > 0 {
		it.CraftName = rollCraftName(min(def.Server, 9), uint8(min(def.ReqWis, 99)), rollsStatsOnly(def.Type), "")
	}
	return it
}
