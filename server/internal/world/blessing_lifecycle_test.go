package world

import (
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"testing"
	"time"
)

// Exercise the actual PvP credit hook and relic capture against the same
// country state, including its database record and authoritative entry packet.
func TestBlessingPvPRelicDatabaseAndResetLifecycle(t *testing.T) {
	r := relicFixture(t)
	oldLight, oldDark, err := db.ReadGodBless()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.SaveGodBless(oldLight, oldDark) })
	kcStore = func(l, d uint32) error { return db.SaveGodBless(l, d) }
	account := testAccount(t)
	id, err := db.CreateCharacter(account, db.NewCharacter{Slot: 4, Name: "BlessLifecycle", HP: 100, MapID: r.mapID})
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.Pool(db.GameData)
	t.Cleanup(func() { pool.Exec(`DELETE FROM PS_GameData_Chars WHERE CharID=?`, id) })
	ch, err := db.CharacterByID(id)
	if err != nil {
		t.Fatal(err)
	}
	ch.Kills = 0
	s, _ := gmWireSession(t, ch, false)
	killRelicFixture(t, r, 0)
	if kcWorld[0].count != 1 {
		t.Fatal("altar contribution")
	}
	s.killPCTo()
	if kcWorld[0].count != 2 || s.char.Kills != 1 || kcStatus(0)[1] != 10 {
		t.Fatal("PvP and altar did not share the country tier")
	}
	s.killPCTo()
	s.killPCTo()
	if kcWorld[0].count != 4 || kcWorld[0].resetAt.IsZero() || r.standing || r.owner != r.initial {
		t.Fatal("maximum did not activate/reset altar")
	}
	l, d, err := db.ReadGodBless()
	if err != nil || l != 4 || d != 0 {
		t.Fatal("maximum persistence", l, d, err)
	}
	p := buildKCStatus(0)
	v, err := proto.DecodeBlessing(p[2:])
	if err != nil || v.Count != 4 || v.RemainingMS < 590000 || v.RemainingMS > 600000 {
		t.Fatal("maximum entry snapshot", v, err)
	}
	// A new world restores saved maximum with its native fresh ten-minute hold.
	InitKillCount(s.logger)
	if kcWorld[0].count != 4 || kcWorld[0].resetAt.IsZero() {
		t.Fatal("DB restore lost blessing")
	}
	reset := kcWorld[0].resetAt
	AddKillCount(1, 10, nopLog{})
	if kcWorld[1].count != 0 {
		t.Fatal("opposing PvP counted during maximum")
	}
	TickKillCount(reset.Add(time.Second), nopLog{})
	l, d, err = db.ReadGodBless()
	if err != nil || l != 0 || d != 0 || kcStatus(0)[1] != 0 {
		t.Fatal("expiry did not clear DB/effects", l, d, err)
	}
	// Native KillPCTo gates country credit at the personal kill cap.
	s.char.Kills = pvpKillCap
	s.killPCTo()
	if kcWorld[0].count != 0 {
		t.Fatal("kill cap credited country")
	}
	oldRates := rewardRates
	t.Cleanup(func() { rewardRates = oldRates })
	rewardRates.PvPKills = 4
	s.char.Kills = 0
	s.killPCToCount(2)
	if s.char.Kills != 8 || kcWorld[0].count != 2 {
		t.Fatalf("rank multiplier changed actual faction kills: rank %d, faction %d", s.char.Kills, kcWorld[0].count)
	}
	s.char.Kills = pvpKillCap - 1
	s.killPCTo()
	if s.char.Kills != pvpKillCap || kcWorld[0].count != 3 {
		t.Fatalf("rank multiplier escaped the cap: rank %d, faction %d", s.char.Kills, kcWorld[0].count)
	}
}
