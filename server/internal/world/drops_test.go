package world

import (
	"math"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/collision"
	"github.com/Spelunkern/Shaiya-GX/internal/mapgeometry"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
)

func TestMonsterDropChanceScaling(t *testing.T) {
	for _, c := range []struct {
		name           string
		drop           db.MobDrop
		percent        uint32
		span, boundary uint64
		want           bool
	}{
		{"normal", db.MobDrop{ItemOrder: 1, DropRate: 20}, 100, 10_000, 2000, false},
		{"double", db.MobDrop{ItemOrder: 1, DropRate: 20}, 200, 10_000, 3999, true},
		{"capped", db.MobDrop{ItemOrder: 6, DropRate: 60}, 200, 10_000, 9999, true},
		{"fractional rare", db.MobDrop{ItemOrder: 7, DropRate: 1}, 50, 100_000_000, 49, true},
		{"rare threshold", db.MobDrop{ItemOrder: 9, DropRate: 1}, 50, 100_000_000, 50, false},
		{"off", db.MobDrop{ItemOrder: 9, DropRate: 1_000_000}, 0, 100_000_000, 0, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := rollMonsterDrop(c.drop, c.percent, func(span uint64) uint64 {
				if span != c.span {
					t.Fatalf("roll range = %d, want %d", span, c.span)
				}
				return c.boundary
			})
			if got != c.want {
				t.Fatalf("drop result = %t, want %t", got, c.want)
			}
		})
	}
}

func TestMonsterLootUsesCurrentRate(t *testing.T) {
	s, _ := forgeRuleSession(t)
	mapID := s.char.MapID
	t.Cleanup(func() { clearMapItems(mapID) })
	oldRates := rewardRates.Load()
	t.Cleanup(func() { rewardRates.Store(oldRates) })
	db.MustDefs().MobDrops[1] = []db.MobDrop{
		{ItemOrder: 1, Grade: 1, DropRate: 50},
		{ItemOrder: 7, Grade: 1, DropRate: 500_000},
	}
	buildDropCatalog()
	mob := &MapMob{ID: 1, MobID: 1}
	for _, c := range []struct {
		percent uint32
		want    int
	}{{0, 0}, {200, 2}} {
		rates := config.DefaultRewardRates()
		rates.MobDropPercent = c.percent
		rewardRates.Store(&rates)
		if drops := rollMobDrops(mapID, mob, s.char.ID, s.char.ID, nil, currentRewardRates(), s.log); len(drops) != c.want {
			t.Fatalf("loot at %d%% = %d items, want %d", c.percent, len(drops), c.want)
		}
	}
}

func TestMonsterGoldScalesPayoutAndKeepsChance(t *testing.T) {
	for _, c := range []struct {
		name                   string
		minimum, maxBonus      uint32
		percent, chance, bonus uint32
		want                   uint32
	}{
		{"normal", 40, 10, 100, 29, 8, 48},
		{"scaled and rounded", 40, 10, 150, 29, 7, 70},
		{"miss", 40, 10, 200, 30, 8, 0},
		{"off", 40, 0, 0, 0, 0, 0},
		{"saturates", math.MaxUint32, 10, 10_000, 0, 9, math.MaxUint32},
	} {
		t.Run(c.name, func(t *testing.T) {
			mob := &MapMob{MoneyMin: c.minimum, MoneyMax: c.maxBonus}
			got := rollMonsterGold(mob, c.percent, 30, func(limit uint32) uint32 {
				if limit == 100 {
					return c.chance
				}
				if c.chance >= 30 || limit != c.maxBonus {
					t.Fatalf("unexpected bonus roll with limit %d", limit)
				}
				return c.bonus
			})
			if got != c.want {
				t.Fatalf("gold = %d, want %d", got, c.want)
			}
		})
	}
}

func TestMobDropStaysOnKillFloor(t *testing.T) {
	g := &mapgeometry.Geometry{}
	g.Reindex()
	g.Add(collision.Triangle{{4, 16, 4}, {4, 16, 6}, {6, 17, 4}})
	g.Add(collision.Triangle{{6, 17, 6}, {6, 17, 4}, {4, 16, 6}})
	g.Add(collision.Triangle{{0, 2, 0}, {0, 2, 20}, {20, 2, 0}})
	g.Add(collision.Triangle{{0, 20, 0}, {0, 20, 20}, {20, 20, 0}})

	origin := mapgeometry.Position{5, 16.5, 5}
	candidate := mapgeometry.Position{5.4, 16.5, 5.3}
	if got := floorMobDrop(g, origin, candidate); got[0] != candidate[0] || got[2] != candidate[2] || math.Abs(float64(got[1]-16.7)) > .001 {
		t.Fatalf("expected loot on the sloped middle floor, got %v", got)
	}
	ledge := mapgeometry.Position{5.875, 16.9375, 5}
	if got := floorMobDrop(g, ledge, mapgeometry.Position{6.25, ledge[1], 5}); got != ledge {
		t.Fatalf("loot scattered over a ledge: %v", got)
	}
	for _, tri := range collision.BlockerBox(collision.Vec{5.25, 16, 4}, collision.Vec{5.3, 18, 6}, 0) {
		g.Add(tri)
	}
	if got := floorMobDrop(g, origin, candidate); got != origin {
		t.Fatalf("loot scattered through a wall: %v", got)
	}
}

func TestMobDropQuestAndGroupOverflowKeepKillHeight(t *testing.T) {
	s, _ := forgeRuleSession(t)
	mapID := s.char.MapID
	t.Cleanup(func() { clearMapItems(mapID) })
	defs := db.MustDefs()
	quest := &db.ItemDef{Type: 27, TypeID: 1, Quality: 100}
	defs.Items[db.ItemKey{Type: quest.Type, TypeID: quest.TypeID}] = quest
	oldLinks := questItemLinks
	t.Cleanup(func() { questItemLinks = oldLinks })
	questItemLinks = map[[2]uint8]questItemLink{{quest.Type, quest.TypeID}: {Start: []uint16{1}}}
	defs.MobDrops[1] = []db.MobDrop{{ItemOrder: 1, Grade: 1, DropRate: 100}}
	buildDropCatalog()

	mob := &MapMob{ID: 1, MobID: 1, PosY: -12.5, def: &db.MobDef{QuestItemID: 1}}
	rewardKill(nil, s.char.ID, mapID, mob, s.log)
	ground := itemsOnMap(mapID)
	if len(ground) != 2 {
		t.Fatalf("expected loot and quest item, got %v", ground)
	}
	for _, drop := range ground {
		if drop.PosY != mob.PosY {
			t.Fatalf("item %d/%d lost kill height: %g", drop.Item.Type, drop.Item.TypeID, drop.PosY)
		}
	}
	clearMapItems(mapID)

	p := &ConnectedPlayer{CharacterID: s.char.ID, party: &party{}}
	if drops := rollMobDrops(mapID, mob, s.char.ID, s.char.ID, []*ConnectedPlayer{p}, currentRewardRates(), s.log); len(drops) != 0 {
		t.Fatalf("group item was not queued: %v", drops)
	}
	for _, slot := range freeInventorySlots(s.items) {
		s.items = append(s.items, db.Item{Bag: slot[0], Slot: slot[1], Type: 16, TypeID: 1, Count: 1})
	}
	s.applyLoot(p)
	ground = itemsOnMap(mapID)
	if len(ground) != 1 || ground[0].PosY != mob.PosY {
		t.Fatalf("group overflow did not fall on the kill floor: %v", ground)
	}
}
