//go:build integration

package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	gxsession "github.com/Spelunkern/Shaiya-GX/internal/session"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
	"github.com/Spelunkern/Shaiya-GX/server/internal/controlplane"
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func TestNativeLiveOperator(t *testing.T) {
	root := os.Getenv("GX_NATIVE_LIVE_FIXTURE")
	if root == "" {
		t.Skip("requires native binaries and a disposable server package")
	}
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GX_TLS", "0")
	// Override TestMain's shared integration environment so the native server
	// uses this package's seed and a runtime retained for the post-run audit.
	t.Setenv("RAPTURE_DATA_DIR", filepath.Join(root, "data"))
	t.Setenv("RAPTURE_DB_SEED", filepath.Join(root, "data", "rapture.sqlite"))
	t.Setenv("RAPTURE_DB_PATH", filepath.Join(root, "runtime", "rapture.sqlite"))
	a, err := net.Listen("tcp", "127.0.0.1:0")
	check(err)
	b, err := net.Listen("tcp", "127.0.0.1:0")
	check(err)
	loginAddress, worldAddress := a.Addr().String(), b.Addr().String()
	worldPort := b.Addr().(*net.TCPAddr).Port
	a.Close()
	b.Close()
	env := config.RuntimeEnvPath(root)
	check(os.MkdirAll(filepath.Dir(env), 0700))
	check(os.WriteFile(env, []byte("RAPTURE_LOGIN_LISTEN="+loginAddress+"\nRAPTURE_WORLD_LISTEN="+worldAddress+
		"\nRAPTURE_MOB_EXP_PERCENT=125\nRAPTURE_QUEST_EXP_PERCENT=125\nRAPTURE_MOB_DROP_PERCENT=0\n"), 0600))

	m := control.New(root)
	ctx := t.Context()
	t.Cleanup(func() {
		if err := m.Apply(context.WithoutCancel(ctx), control.ForceStop, control.Services(), nil); err != nil {
			t.Error(err)
		}
	})
	check(m.Apply(ctx, control.Start, control.Services(), nil))
	original := m.Snapshot(ctx)
	ensurePIDs := func() {
		t.Helper()
		snapshot := control.New(root).Snapshot(ctx)
		for _, service := range control.Services() {
			if got := snapshot[service]; got.State != control.Ready || got.Processes != 1 || got.PID != original[service].PID {
				t.Fatalf("%s changed during live edit: %+v", service, got)
			}
		}
	}
	ensurePIDs()
	operate := func(req operatorwire.Request) operatorwire.Response {
		t.Helper()
		res, err := m.Operator(ctx, req)
		check(err)
		return res
	}
	expectError := func(req operatorwire.Request, detail string) {
		t.Helper()
		if _, err := m.Operator(ctx, req); err == nil || !strings.Contains(err.Error(), detail) {
			t.Fatalf("%s should reject with %q, got %v", req.Action, detail, err)
		}
	}
	var descriptor struct{ Address string }
	wire, err := os.ReadFile(filepath.Join(root, "runtime", "control", "gx_world.json"))
	check(err)
	check(json.Unmarshal(wire, &descriptor))
	clear(wire)
	unauth, err := (&http.Client{Timeout: 3 * time.Second}).Post("http://"+descriptor.Address+"/operator", "application/json", strings.NewReader("{}"))
	check(err)
	unauth.Body.Close()
	if unauth.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated operator returned HTTP %d", unauth.StatusCode)
	}
	const oldPassword, newPassword = "temporary-apple", "temporary-lemon"
	suffix := strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
	name, bystander := "GXLive"+suffix, "GXWatch"+suffix
	operate(operatorwire.Request{Action: "create", Name: name, Password: oldPassword})
	operate(operatorwire.Request{Action: "create", Name: bystander, Password: oldPassword})
	ticketUser := "GXTicket" + suffix
	operate(operatorwire.Request{Action: "create", Name: ticketUser, Password: oldPassword})
	account := func(user string) operatorwire.Account {
		t.Helper()
		found := operate(operatorwire.Request{Action: "accounts", Query: user}).Accounts
		if len(found) != 1 {
			t.Fatalf("expected one throwaway account, got %d", len(found))
		}
		return found[0]
	}
	target := account(name)
	waitPacket := func(w *gxsession.World, opcode proto.Opcode, match func([]byte) bool) []byte {
		t.Helper()
		timer := time.NewTimer(9 * time.Second)
		defer timer.Stop()
		for {
			select {
			case p, ok := <-w.Packets:
				if !ok {
					t.Fatalf("World closed while waiting for 0x%04x: %v", opcode, w.Err())
				}
				if proto.Op(p) == opcode && (match == nil || match(p)) {
					return p
				}
			case <-timer.C:
				t.Fatalf("no matching 0x%04x from World", opcode)
			}
		}
	}
	closed := func(w *gxsession.World) {
		t.Helper()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		for {
			select {
			case _, ok := <-w.Packets:
				if !ok {
					return
				}
			case <-timer.C:
				t.Fatal("affected session was not disconnected")
			}
		}
	}
	dial := func(user, password string) *gxsession.World {
		t.Helper()
		var l *gxsession.Login
		deadline := time.Now().Add(5 * time.Second)
		for {
			l, err = gxsession.Dial(loginAddress, user, password)
			if err == nil || time.Now().After(deadline) {
				break
			}
			time.Sleep(150 * time.Millisecond)
		}
		check(err)
		w := connectNativeWorld(t, l, worldPort)
		waitPacket(w, proto.AccountFaction, nil)
		return w
	}
	observer := dial(bystander, oldPassword)
	player := dial(name, oldPassword)
	create := func(w *gxsession.World, uid uint32, slot uint8, charName string) uint32 {
		t.Helper()
		check(w.Send(proto.FactionPacket(0)))
		waitPacket(w, proto.AccountFaction, nil)
		check(w.Send(proto.CreateCharacterPacket(proto.NewCharacter{Slot: slot, Name: charName, Hair: 1, Face: 1, Height: 1})))
		if reply := waitPacket(w, proto.CreateCharacter, nil); len(reply) < 3 || reply[2] != 0 {
			t.Fatal("throwaway character creation rejected")
		}
		chars := operate(operatorwire.Request{Action: "details", AccountID: uid}).Characters
		for _, ch := range chars {
			if ch.Name == charName && !ch.Deleted {
				return ch.ID
			}
		}
		t.Fatalf("created character %q missing from roster", charName)
		return 0
	}
	selectCharacter := func(w *gxsession.World, id uint32) {
		t.Helper()
		check(w.Send(proto.SelectCharacterPacket(id)))
		if reply := waitPacket(w, proto.SelectCharacter, nil); len(reply) < 3 || reply[2] != 0 {
			t.Fatal("throwaway character selection rejected")
		}
		waitPacket(w, proto.CharacterSkillBar, nil)
		check(w.Send(proto.MapLoadedPacket()))
		waitPacket(w, proto.CharacterMove, nil)
	}
	mainName, altName := "GXTest"+suffix, "GXAlt"+suffix
	id := create(player, target.ID, 0, mainName)
	alt := create(player, target.ID, 1, altName)
	observerID := create(observer, account(bystander).ID, 0, "GXView"+suffix)
	selectCharacter(player, id)
	selectCharacter(observer, observerID)
	ping := func(w *gxsession.World) {
		t.Helper()
		check(w.Send(proto.NewWriter(proto.Ping).Bytes()))
		waitPacket(w, proto.Ping, nil)
	}
	read := operatorwire.Request{Action: "stats", AccountID: target.ID, CharacterID: id}
	before := operate(read)
	if !before.Online || before.Stats == nil {
		t.Fatal("missing live actor snapshot")
	}
	next := *before.Stats
	next.Strength += 3
	next.StatPoints += 1
	next.SkillPoints += 1
	next.Gold = 12345
	next.EXP = min(before.MinEXP+19, before.MaxEXP)
	if next.EXP == before.Stats.EXP {
		t.Fatal("fixture has no EXP room")
	}
	next.Kills = 87
	save := operatorwire.Request{Action: "save-stats", AccountID: target.ID, CharacterID: id, Stats: &next, ExpectedStats: before.Stats}
	invalid := next
	invalid.EXP = before.MaxEXP + 1
	bad := save
	bad.Stats = &invalid
	expectError(bad, "current level")
	operate(save)
	waitPacket(player, proto.CharacterDetails, func(p []byte) bool { d, e := proto.DecodeDetails(p); return e == nil && d.Exp == next.EXP })
	waitPacket(player, proto.LevelUpSelf, nil)
	waitPacket(player, proto.SetMoney, func(p []byte) bool { return len(p) >= 6 && binary.LittleEndian.Uint32(p[2:]) == next.Gold })
	waitPacket(player, proto.UserKillCountUpdate, func(p []byte) bool {
		v, e := proto.DecodePvPCountUpdate(p)
		return e == nil && v.Index == 0 && v.Value == next.Kills
	})
	waitPacket(player, proto.StatsReset, nil)
	if after := operate(read); !after.Online || *after.Stats != next {
		t.Fatalf("live character readback %+v", after)
	}
	expectError(save, "character changed")
	operateDelete := operatorwire.Request{Action: "delete", AccountID: target.ID, CharacterID: id, ConfirmName: mainName}
	expectError(operateDelete, "character is online")
	operateDelete.Action = "restore"
	expectError(operateDelete, "character is online")

	details := operatorwire.Request{Action: "details", AccountID: target.ID}
	isDeleted := func(id uint32, deleted, online bool) {
		t.Helper()
		for _, ch := range operate(details).Characters {
			if ch.ID == id && ch.Deleted == deleted && ch.Online == online {
				return
			}
		}
		t.Fatalf("character %d does not have deleted=%t online=%t", id, deleted, online)
	}
	isDeleted(id, false, true)
	isDeleted(alt, false, false)
	removeAlt := operatorwire.Request{Action: "delete", AccountID: target.ID, CharacterID: alt, ConfirmName: altName}
	restoreAlt := operatorwire.Request{Action: "restore", AccountID: target.ID, CharacterID: alt}
	wrongName := removeAlt
	wrongName.ConfirmName = "wrong name"
	expectError(wrongName, "exact character name")
	wrongAccount := removeAlt
	wrongAccount.AccountID = account(bystander).ID
	expectError(wrongAccount, "unavailable for this account")
	operate(removeAlt)
	isDeleted(alt, true, false)
	expectError(removeAlt, "unavailable")
	ping(player)
	ping(observer)
	operate(restoreAlt)
	isDeleted(alt, false, false)
	expectError(restoreAlt, "unavailable")
	operate(removeAlt)
	isDeleted(alt, true, false)

	check(player.Send(proto.LogoutPacket()))
	waitPacket(player, proto.Logout, nil)
	waitPacket(player, proto.AccountFaction, nil)
	check(player.Send(proto.SelectCharacterPacket(alt)))
	packet := waitPacket(player, proto.SelectCharacter, nil)
	if got, err := proto.DecodeIDResult(packet); err != nil || got.Result == 0 {
		t.Fatal("World accepted selection of a deleted character")
	}
	replacementName := "GXSlot" + suffix
	replacement := create(player, target.ID, 1, replacementName)
	expectError(restoreAlt, "check the restore window or slot")
	removeReplacement := operatorwire.Request{Action: "delete", AccountID: target.ID, CharacterID: replacement, ConfirmName: replacementName}
	operate(removeReplacement)
	operate(restoreAlt)
	waitPacket(player, proto.CharacterList, func(p []byte) bool {
		slot, err := proto.DecodeCharacterSlot(p)
		return err == nil && slot.ID == alt && !slot.PendingDelete
	})
	selectCharacter(player, alt)
	expectError(removeAlt, "character is online")
	check(player.Send(proto.LogoutPacket()))
	waitPacket(player, proto.Logout, nil)
	waitPacket(player, proto.AccountFaction, nil)
	selectCharacter(player, id)
	ping(observer)
	t.Log("Live delete/restore enforced selected-character safety, slot conflicts and typed ownership; the character-select roster refreshed.")
	ensurePIDs()

	points := operatorwire.Request{Action: "account", AccountID: target.ID, Points: 1550, ExpectedAccount: &target}
	operate(points)
	waitPacket(player, proto.CashPoint, func(p []byte) bool { return len(p) >= 6 && binary.LittleEndian.Uint32(p[2:]) == points.Points })
	ping(player)
	expectError(points, "account changed")
	current := account(name)
	if current.Points != points.Points || current.GM {
		t.Fatal("unexpected points or GM balance")
	}
	t.Log("Online character stats, EXP, rank kills and wallet updated on the existing client; stale and invalid edits rejected.")

	gmReply := func(w *gxsession.World, enabled bool) {
		t.Helper()
		check(w.Send(proto.ChatPacket(0, false, "/set money 12345")))
		p := waitPacket(w, proto.GXGMState, func(p []byte) bool {
			v, e := proto.DecodeGMState(p)
			return e == nil && (strings.Contains(v.Message, "set to") || strings.Contains(v.Message, "access denied"))
		})
		v, err := proto.DecodeGMState(p)
		check(err)
		if v.GM != enabled || strings.Contains(v.Message, "access denied") == enabled {
			t.Fatalf("wrong GM state %+v", v)
		}
	}
	gmReply(player, false)
	operate(operatorwire.Request{Action: "account", AccountID: target.ID, Points: current.Points, GM: true, ExpectedAccount: &current})
	closed(player)
	ping(observer)
	player = dial(name, oldPassword)
	selectCharacter(player, id)
	gmReply(player, true)
	if after := operate(read); !after.Online || *after.Stats != next {
		t.Fatalf("character edit did not survive logout: %+v", after)
	}
	operate(operatorwire.Request{Action: "password", AccountID: target.ID, Password: newPassword})
	closed(player)
	ping(observer)
	if l, e := gxsession.Dial(loginAddress, name, oldPassword); e == nil {
		l.Close()
		t.Fatal("old password accepted")
	} else {
		var failure gxsession.LoginError
		if !errors.As(e, &failure) {
			t.Fatalf("old password did not get a real login rejection: %v", e)
		}
	}
	player = dial(name, newPassword)
	selectCharacter(player, id)
	gmReply(player, true)
	if os.Getenv("GX_VERIFY_ACTIVITY_COMBAT") == "1" {
		rule := config.ScopedDropRule{ID: 77, Name: "Isolated copper test", Maps: []uint16{1},
			MinLevel: 1, MaxLevel: 1, Chance: 10000000000, ItemType: 16, ItemID: 1, DailyLimit: 1}
		check(m.SaveScopedDrop(ctx, config.ScopedDropEdit{Rule: &rule}))
		gold, goldChance := uint32(250), uint32(100)
		check(m.SaveMapReward(ctx, 1, config.MapRewardOverrides{MonsterGold: &gold, GoldChance: &goldChance}))
		for range 2 {
			check(player.Send(proto.ChatPacket(proto.ChannelNormal, true, "/mmake 2008 1")))
			spawn := waitPacket(player, proto.GXMobEnter, func(p []byte) bool {
				mob, err := proto.DecodeMobEnter(p)
				return err == nil && mob.MobID == 2008
			})
			mob, err := proto.DecodeMobEnter(spawn)
			check(err)
			check(player.Send(proto.ChatPacket(proto.ChannelNormal, true, "/kill #"+strconv.FormatUint(uint64(mob.ID), 10))))
			waitPacket(player, proto.MobDeath, func(p []byte) bool {
				return len(p) >= 6 && binary.LittleEndian.Uint32(p[2:6]) == mob.ID
			})
			balance := operate(read).Stats.Gold
			bag := waitPacket(player, proto.MapAddItem, func(p []byte) bool {
				item, err := proto.DecodeGroundItem(p)
				return err == nil && item.Type == 26 && item.TypeID == 0
			})
			groundGold, err := proto.DecodeGroundItem(bag)
			check(err)
			check(player.Send(proto.PickupPacket(groundGold.ID)))
			waitPacket(player, proto.SetMoney, func(p []byte) bool {
				return len(p) >= 6 && binary.LittleEndian.Uint32(p[2:]) == balance+250
			})
			ping(player)
		}
		drops, kills, goldGains := 0, 0, 0
		activity, err := control.NewLogReader(root).Poll()
		check(err)
		for _, entry := range activity {
			if entry.Activity == "loot.gold" && strings.Contains(entry.Message, "gained 250 gold") {
				goldGains++
			}
			if entry.Activity == "loot.drop" && strings.Contains(entry.Message, "Isolated copper test") {
				drops++
			}
			if entry.Activity == "kill.monster" && strings.Contains(entry.Message, "Small Thief Monkey") && entry.MapID != nil && *entry.MapID == 1 {
				kills++
			}
		}
		if drops != 1 || kills != 2 || goldGains != 2 {
			t.Fatalf("wanted two kills, two gold gains and one capped drop, got kills=%d drops=%d gold=%d", kills, drops, goldGains)
		}
		check(m.SaveScopedDrop(ctx, config.ScopedDropEdit{DeleteID: rule.ID}))
		t.Log("Two GM-assisted mob kills reached the kill feed. Two 250-gold pickups reached the feed. The account cap allowed exactly one scoped event drop.")
	} else {
		check(player.Send(proto.ChatPacket(proto.ChannelNormal, true, "/mmake 2008 1")))
		gmReply(player, true)
	}
	current = account(name)
	operate(operatorwire.Request{Action: "account", AccountID: target.ID, Points: current.Points, ExpectedAccount: &current})
	closed(player)
	player = dial(name, newPassword)
	selectCharacter(player, id)
	gmReply(player, false)
	ping(observer)
	ensurePIDs()
	pending, err := gxsession.Dial(loginAddress, ticketUser, oldPassword)
	check(err)
	defer pending.Close()
	_, err = pending.SelectServer(pending.Servers[0].ID, proto.ClientVersion)
	check(err)
	operate(operatorwire.Request{Action: "password", AccountID: account(ticketUser).ID, Password: newPassword})
	if w, err := pending.DialWorld(net.IPv4(127, 0, 0, 1), worldPort); err == nil {
		w.Close()
		t.Fatal("World accepted an unconsumed handoff issued before the password change")
	}
	t.Log("Password rotation and GM grant/revoke disconnected only the target. New credentials and permissions applied after reconnect.")

	requested := map[string]string{
		config.MobEXPKey: "200", config.QuestEXPKey: "175", config.PvPKillKey: "3",
		config.MobDropKey: "200", config.MobGoldKey: "250", config.WorldDropKey: "175",
	}
	check(m.SaveRewardSettings(ctx, requested))
	for _, key := range []string{config.MobGoldKey, config.WorldDropKey} {
		if err := m.SaveRewardSettings(ctx, map[string]string{key: "10001"}); err == nil {
			t.Fatalf("invalid rate accepted for %s", key)
		}
	}
	worldImage := filepath.Join(root, "build", "gx_world")
	if runtime.GOOS == "windows" {
		worldImage += ".exe"
	}
	drops := requested[config.MobDropKey]
	_, err = controlplane.RunOperator(ctx, root, original[control.World].PID, worldImage, operatorwire.Request{
		Action: "reward-rates", MobRate: "200", QuestRate: "175", KillRate: "3", DropRate: &drops,
	})
	check(err)
	rates, _, err := m.RewardSettings()
	check(err)
	for key, want := range requested {
		if rates[key] != want {
			t.Fatalf("%s = %q, want %q", key, rates[key], want)
		}
	}
	gold, event := uint32(400), uint32(0)
	check(m.SaveMapReward(ctx, 1, config.MapRewardOverrides{MonsterGold: &gold, EventLoot: &event}))
	verifyMapRates := func(pid int) {
		t.Helper()
		for _, mapID := range []uint16{1, 2} {
			res, err := controlplane.RunOperator(ctx, root, pid, worldImage, operatorwire.Request{
				Action: "map-rates-effective", MapID: mapID,
			})
			check(err)
			if res.EffectiveRates == nil {
				t.Fatalf("missing rates for map %d", mapID)
			}
			wantGold, wantEvent := uint32(250), uint32(175)
			if mapID == 1 {
				wantGold, wantEvent = 400, 0
			}
			if got := res.EffectiveRates; got.MobGoldPercent != wantGold || got.WorldDropPercent != wantEvent || got.MobEXPPercent != 200 {
				t.Fatalf("map %d rates %+v, want gold %d and event %d", mapID, got, wantGold, wantEvent)
			}
		}
	}
	verifyMapRates(original[control.World].PID)
	if err := m.SaveMapTuning(ctx, 1, config.MapTuning{Closed: true}); err == nil {
		t.Fatal("starting map accepted a close-entry override")
	}
	respawn := uint32(150)
	check(m.SaveMapTuning(ctx, 1, config.MapTuning{RespawnPercent: &respawn}))
	tuned, err := m.MapTuning()
	check(err)
	if tuned[1].RespawnPercent == nil || *tuned[1].RespawnPercent != respawn {
		t.Fatal("map respawn override did not persist")
	}
	mapActivity, err := m.MapActivity(ctx, 1)
	check(err)
	if mapActivity.Players < 1 {
		t.Fatalf("map activity did not include the online player: %+v", mapActivity)
	}
	presetGold := uint32(650)
	preset := config.RewardPreset{Name: "Map test event", Maps: []uint16{2}, Rates: config.MapRewardOverrides{MonsterGold: &presetGold}}
	check(m.SaveRewardEvent(ctx, config.RewardEventEdit{Preset: &preset}))
	check(m.SaveRewardEvent(ctx, config.RewardEventEdit{StartPreset: preset.Name, Minutes: 2}))
	checkGold := func(id uint16, want uint32) {
		t.Helper()
		res, err := controlplane.RunOperator(ctx, root, original[control.World].PID, worldImage, operatorwire.Request{Action: "map-rates-effective", MapID: id})
		check(err)
		if res.EffectiveRates == nil || res.EffectiveRates.MobGoldPercent != want {
			t.Fatalf("map %d event rate = %+v, want %d", id, res.EffectiveRates, want)
		}
	}
	checkGold(1, 400)
	checkGold(2, 650)
	check(m.SaveRewardEvent(ctx, config.RewardEventEdit{End: true}))
	verifyMapRates(original[control.World].PID)
	t.Log("A timed preset affected Map 2 only. Ending it restored the rates underneath it.")
	economy := config.DefaultEconomy()
	economy.VendorBuyPercent, economy.VendorSellPercent, economy.RerollGold = 150, 200, 75
	check(m.SaveEconomy(ctx, economy))
	readQuote := func(w *gxsession.World) proto.EconomySettings {
		t.Helper()
		p := waitPacket(w, proto.GXEconomySettings, func(p []byte) bool {
			q, err := proto.DecodeEconomySettings(p)
			return err == nil && q.BuyPercent == 150 && q.SellPercent == 200 && q.RerollGold == 75 && !q.Rejected
		})
		q, err := proto.DecodeEconomySettings(p)
		check(err)
		return q
	}
	quote := readQuote(player)
	if other := readQuote(observer); quote != other {
		t.Fatalf("clients received different economy quotes: %+v vs %+v", quote, other)
	}
	oldBalance := operate(read).Stats.Gold
	request := proto.ComposePacket(1, 0, 1, 1)
	for _, item := range [][]byte{request, proto.EconomyActionPacket(quote.Revision-1, quote.RerollGold, request)} {
		check(player.Send(item))
		waitPacket(player, proto.GXEconomySettings, func(p []byte) bool {
			q, err := proto.DecodeEconomySettings(p)
			return err == nil && q.Revision == quote.Revision && q.Rejected
		})
	}
	if got := operate(read).Stats.Gold; got != oldBalance {
		t.Fatalf("refused quote changed the balance from %d to %d", oldBalance, got)
	}
	check(m.SaveEconomy(ctx, config.DefaultEconomy()))
	t.Log("Both connected clients received the live quote. Legacy and stale reroll requests were refused without debiting gold.")
	for _, minutes := range []int{1, 0} {
		check(m.SetShield(ctx, minutes))
		want := netx.ShieldStandard
		if minutes > 0 {
			want = netx.ShieldUnderAttack
		}
		for _, service := range control.Services() {
			if got := m.Snapshot(ctx)[service].Shield.Mode; got != want {
				t.Fatalf("%s shield = %q, want %q", service, got, want)
			}
		}
	}
	t.Log("Map 1 override affected only Map 1. Shield switched to Under Attack and back while both sessions stayed connected.")
	const rateSummary = "monster EXP 200%, quest EXP 175%, PvP rank kills 3x, monster loot 200%, monster gold 250%, global event loot 175%"
	entries, err := control.NewLogReader(root).Poll()
	check(err)
	applied := false
	activity := map[string]bool{}
	for _, e := range entries {
		activity[e.Activity] = true
		if e.Service == control.World && strings.Contains(e.Message, "applied live reward rates: "+rateSummary) {
			applied = true
		}
		if strings.Contains(e.Message, oldPassword) || strings.Contains(e.Message, newPassword) {
			t.Fatal("credential appeared in server log")
		}
	}
	if !applied {
		t.Fatal("running World did not log applied rates")
	}
	for _, kind := range []string{"player.login", "player.enter", "player.leave", "staff.settings", "staff.command", "staff.operator"} {
		if !activity[kind] {
			t.Fatalf("missing structured activity %q", kind)
		}
	}
	ensurePIDs()
	ping(player)
	ping(observer)
	t.Log("Gold at 250% and event loot at 175% applied with both client sessions and service PIDs retained. Older save preserved the new rates.")
	const moderationReason = "private disposable moderation check"
	check(player.Close())
	player = dial(name, newPassword)
	selectCharacter(player, id)
	pending, err = gxsession.Dial(loginAddress, ticketUser, newPassword)
	check(err)
	_, err = pending.SelectServer(pending.Servers[0].ID, proto.ClientVersion)
	check(err)
	operate(operatorwire.Request{Action: "ban-account", AccountID: account(ticketUser).ID})
	if w, err := pending.DialWorld(net.IPv4(127, 0, 0, 1), worldPort); err == nil {
		w.Close()
		t.Fatal("pre-ban handoff reached World")
	}
	operate(operatorwire.Request{Action: "ban-account", AccountID: target.ID, BanReason: moderationReason})
	closed(player)
	ping(observer)
	ensurePIDs()
	if got := account(name); !got.Banned || got.BanReason != moderationReason {
		t.Fatal("account ban missing from operator readback")
	}
	rejectAccount := func(user, secret string) {
		t.Helper()
		l, err := gxsession.Dial(loginAddress, user, secret)
		if err == nil {
			l.Close()
			t.Fatal("banned account signed in")
		}
		var denied gxsession.LoginError
		if !errors.As(err, &denied) || denied.Code != 4 {
			t.Fatalf("expected the native account block response, got %v", err)
		}
	}
	rejectAccount(name, newPassword)
	operate(operatorwire.Request{Action: "unban-account", AccountID: target.ID})
	player = dial(name, newPassword)
	selectCharacter(player, id)
	ipTicketUser := "GXPeer" + suffix
	operate(operatorwire.Request{Action: "create", Name: ipTicketUser, Password: newPassword})
	ipTicket, err := gxsession.Dial(loginAddress, ipTicketUser, newPassword)
	check(err)
	_, err = ipTicket.SelectServer(ipTicket.Servers[0].ID, proto.ClientVersion)
	check(err)
	expectError(operatorwire.Request{Action: "ban-ip", IPAddress: "127.0.0.0/8"}, "one valid IPv4 or IPv6")
	operate(operatorwire.Request{Action: "ban-ip", IPAddress: "::ffff:127.0.0.1", BanReason: moderationReason})
	closed(player)
	closed(observer)
	ensurePIDs()
	requireLocalIPBan := func() {
		t.Helper()
		list := operate(operatorwire.Request{Action: "ip-bans"}).IPBans
		if len(list) != 1 || list[0].Address != "127.0.0.1" || list[0].Reason != moderationReason {
			t.Fatal("canonical IP ban missing from operator readback")
		}
		l, err := gxsession.Dial(loginAddress, name, newPassword)
		if err == nil {
			l.Close()
			t.Fatal("banned IP reached Login")
		}
		if w, err := ipTicket.DialWorld(net.IPv4(127, 0, 0, 1), worldPort); err == nil {
			w.Close()
			t.Fatal("banned IP reached World")
		}
	}
	requireLocalIPBan()
	t.Log("Live moderation revoked a pending handoff, kicked the targeted account, preserved the observer until the shared IP ban, and kept both PIDs.")
	check(m.Apply(ctx, control.Restart, []control.Service{control.World}, nil))
	restarted := m.Snapshot(ctx)
	verifyMapRates(restarted[control.World].PID)
	if restarted[control.Login].PID != original[control.Login].PID || restarted[control.World].PID == original[control.World].PID {
		t.Fatal("explicit World restart did not leave Login in place")
	}
	requireLocalIPBan()
	operate(operatorwire.Request{Action: "unban-ip", IPAddress: "127.0.0.1"})
	w, err := ipTicket.DialWorld(net.IPv4(127, 0, 0, 1), worldPort)
	check(err)
	waitPacket(w, proto.AccountFaction, nil)
	check(w.Close())
	rejectAccount(ticketUser, newPassword)
	operate(operatorwire.Request{Action: "unban-account", AccountID: account(ticketUser).ID})
	player = dial(name, newPassword)
	ping(player)
	check(player.Close())
	entries, err = control.NewLogReader(root).Poll()
	check(err)
	persisted := false
	for _, e := range entries {
		if strings.Contains(e.Message, moderationReason) {
			t.Fatal("private ban reason leaked into staff activity")
		}
		if e.Service == control.World && strings.Contains(e.Message, "World reward rates: "+rateSummary) {
			persisted = true
		}
	}
	if !persisted {
		t.Fatal("World did not reload saved rates on its next startup")
	}
	check(m.Apply(ctx, control.Stop, control.Services(), nil))
	stopped := m.Snapshot(ctx)
	for _, service := range control.Services() {
		if stopped[service].State != control.Stopped || stopped[service].Processes != 0 {
			t.Fatalf("%s still running: %+v", service, stopped[service])
		}
	}
	for _, address := range []string{loginAddress, worldAddress} {
		if conn, e := net.DialTimeout("tcp", address, 250*time.Millisecond); e == nil {
			conn.Close()
			t.Fatalf("listener remains at %s", address)
		}
	}
	t.Log("Live rate apply, persistence, unchanged service PIDs and clean shutdown passed.")
}

func connectNativeWorld(t *testing.T, login *gxsession.Login, port int) *gxsession.World {
	t.Helper()
	defer login.Close()
	if len(login.Servers) == 0 {
		t.Fatal("Login advertised no worlds")
	}
	if _, err := login.SelectServer(login.Servers[0].ID, proto.ClientVersion); err != nil {
		t.Fatal(err)
	}
	w, err := login.DialWorld(net.IPv4(127, 0, 0, 1), port)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { w.Close() })
	return w
}
