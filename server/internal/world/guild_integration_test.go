//go:build integration

package world

import (
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// TestGuildLifecycle runs the CUser::PacketGuild flows with seven
// synthetic users: solo and grouped creation with the configured fee, invitation, application and acceptance, rank change, remark,
// expel, leave and disband by vote, checking the realm.registry and the
// GuildChars rows. Its rows are removed afterwards.
func TestGuildLifecycle(t *testing.T) {
	if _, err := os.Stat("data/game"); err != nil {
		t.Chdir("../..")
	}
	config.LoadRuntimeEnv()
	if err := db.Open(db.UserData, db.GameData, db.GameDefs); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ReloadDefsForTest(); err != nil { // unit tests may have replaced them
		t.Fatal(err)
	}
	logger, err := logx.New("gx_world_test")
	if err != nil {
		t.Fatal(err)
	}
	if err := LoadGuilds(logger); err != nil {
		t.Fatal(err)
	}
	const mapID = 60030
	session := func(_ uint32, name string) *Session {
		a, b := net.Pipe()
		go io.Copy(io.Discard, b)
		t.Cleanup(func() { a.Close(); b.Close() })
		account := testAccount(t)
		id, err := db.CreateCharacter(account, db.NewCharacter{Slot: 0, Name: name, HP: 100, MP: 100, SP: 100})
		if err != nil {
			t.Fatal(err)
		}
		row, err := db.CharacterByID(id)
		if err != nil {
			t.Fatal(err)
		}
		ch := &row
		ch.Level = 40
		ch.Mode = growNormal
		ch.Class = 1
		ch.Money = guildHouse.CreateMoney + 5000
		ch.MapID = mapID
		ch.PosX = 100
		ch.PosZ = 100
		if err := db.SaveCharacter(ch); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			pool, _ := db.Pool(db.GameData)
			pool.Exec(`DELETE FROM PS_GameData_CharItems WHERE CharID=?`, id)
			pool.Exec(`DELETE FROM PS_GameData_GuildChars WHERE CharID=?`, id)
			pool.Exec(`DELETE FROM PS_GameData_Chars WHERE CharID=?`, id)
		})
		s := &Session{char: ch, log: logger.Tagged(name), logger: logger, faction: FactionLight,
			skillCooldowns: map[uint16]time.Time{}, out: &connWriter{conn: a}}
		s.account = account
		s.accountID = uint32(account.UserUID)
		s.player = RegisterPlayer(*ch, nil, false, FactionLight, s.out, logger, s)
		t.Cleanup(s.player.Unregister)
		return s
	}
	var ss []*Session
	pt := &party{id: 990001, itemDivision: 1}
	for i := range 7 {
		s := session(uint32(990101+i), "GTest"+string(rune('A'+i)))
		ss = append(ss, s)
		pt.members = append(pt.members, s.player)
	}
	partyMu.Lock()
	for _, s := range ss {
		s.player.party = pt
	}
	partyMu.Unlock()
	outsider := session(990120, "GTestOut")
	applicant := session(990121, "GTestApp")
	var created uint32
	defer func() {
		pool, _ := db.Pool(db.GameData)
		if created != 0 {
			pool.Exec(`DELETE FROM PS_GameData_GuildStoredItems WHERE GuildID = ?1; DELETE FROM PS_GameData_GuildNpcLv WHERE GuildID = ?1; DELETE FROM PS_GameData_GuildChars WHERE GuildID = ?1; DELETE FROM PS_GameData_GuildDetails WHERE GuildID = ?1;
				DELETE FROM PS_GameData_Guilds WHERE GuildID = ?1;
				DELETE FROM PS_UserData_CreatedGuilds WHERE ServerID = 1 AND GuildID = ?1`, int32(created))
		}
		pool.Exec(`DELETE FROM PS_GameData_GuildChars WHERE CharID BETWEEN 990100 AND 990199;
			DELETE FROM PS_GameData_CharItems WHERE CharID BETWEEN 990100 AND 990199`)
	}()
	guildOf := func(s *Session) *guild {
		guildMu.Lock()
		defer guildMu.Unlock()
		return s.player.guild.guild
	}
	rows := func(id uint32) (members, applicants int) {
		pool, _ := db.Pool(db.GameData)
		pool.QueryRow(`SELECT COUNT(CASE WHEN GuildLevel > 0 THEN 1 END), COUNT(CASE WHEN GuildLevel = 0 THEN 1 END)
			FROM PS_GameData_GuildChars WHERE GuildID = ?1 AND Del = 0`, int32(id)).Scan(&members, &applicants)
		return
	}

	request := func(name, remark string) []byte {
		p := make([]byte, 2+25+65)
		p[0], p[1] = 0x21, 0x0d
		copy(p[2:27], name)
		copy(p[27:], remark)
		return p
	}
	for _, s := range []*Session{outsider, applicant} {
		s.char.Level, s.char.Mode = 1, growEasy
		s.syncPlayer()
	}
	trackGuild := func(g *guild) {
		t.Cleanup(func() {
			pool, _ := db.Pool(db.GameData)
			pool.Exec(`DELETE FROM PS_GameData_GuildChars WHERE GuildID=?1;
				DELETE FROM PS_GameData_GuildDetails WHERE GuildID=?1;
				DELETE FROM PS_GameData_Guilds WHERE GuildID=?1;
				DELETE FROM PS_UserData_CreatedGuilds WHERE ServerID=1 AND GuildID=?1`, g.ID)
		})
	}
	originalFee := guildHouse.CreateMoney
	t.Cleanup(func() { guildHouse.CreateMoney = originalFee })
	guildHouse.CreateMoney = outsider.char.Money + 1
	outsider.handleGuild(packet.GuildCreate, request("GXTestSolo", "Solo notice"))
	if guildOf(outsider) != nil {
		t.Fatal("solo founder created a guild without covering the fee")
	}
	guildHouse.CreateMoney = 2500
	beforeGold := outsider.char.Money
	outsider.handleGuild(packet.GuildCreate, request("GXTestSolo", "Solo notice"))
	solo := guildOf(outsider)
	if solo == nil || len(solo.members) != 1 || solo.Remark != "Solo notice" || outsider.char.Money != beforeGold-2500 {
		t.Fatalf("solo founding failed: guild=%+v gold=%d", solo, outsider.char.Money)
	}
	trackGuild(solo)
	if row, err := db.CharacterByID(outsider.char.ID); err != nil || row.Money != outsider.char.Money {
		t.Fatalf("solo fee not persisted: gold=%d err=%v", row.Money, err)
	}
	outsider.guildLeave()
	pair := &party{id: 990002, members: []*ConnectedPlayer{outsider.player, applicant.player}}
	partyMu.Lock()
	outsider.player.party, applicant.player.party = pair, pair
	partyMu.Unlock()
	applicant.handleGuild(packet.GuildCreate, request("GXTestPair", "Pair notice"))
	applicant.guildCreateAnswer(true)
	applicant.guildCreateAnswer(true)
	if guildOf(applicant) != nil {
		t.Fatal("pair guild created before both members accepted")
	}
	outsider.guildCreateAnswer(true)
	gPair := guildOf(applicant)
	if gPair == nil || len(gPair.members) != 2 || guildOf(outsider) != gPair || gPair.Master != applicant.char.Name {
		t.Fatalf("pair founding failed: %+v", gPair)
	}
	trackGuild(gPair)
	outsider.guildLeave()
	outsider.guildApply(gPair.ID)
	applicant.guildAnswerApplicant(true, outsider.char.ID)
	if guildOf(outsider) != gPair {
		t.Fatal("immediate reentry to the same guild failed")
	}
	outsider.guildLeave()
	applicant.guildLeave()
	partyMu.Lock()
	outsider.player.party, applicant.player.party = nil, nil
	partyMu.Unlock()
	guildHouse.CreateMoney = originalFee

	ss[0].handleGuild(packet.GuildCreate, request("RaptureTestGuild", "hello"))
	for i, s := range ss {
		if guildOf(s) != nil {
			t.Fatalf("guild created before member %d accepted", i)
		}
		s.handleGuild(packet.GuildCreateAgree, []byte{0x22, 0x0d, 1})
	}
	g := guildOf(ss[0])
	if g == nil || g.Name != "RaptureTestGuild" || g.Remark != "hello" || len(g.members) != 7 || g.members[ss[0].char.ID].Rank != guildRankMaster {
		t.Fatalf("guild not created: %+v", g)
	}
	created = g.ID
	ss[0].completePendingGuild()
	if ss[0].char.Money != 5000 {
		t.Fatalf("master money %d, want 5000", ss[0].char.Money)
	}
	if m, _ := rows(g.ID); m != 7 {
		t.Fatalf("GuildChars members %d, want 7", m)
	}

	// Guild bank (bag 0xFF): closed while unranked, rank 9 cannot deposit,
	// the master deposits, moves and withdraws (5% of Sell x count).
	var item *db.ItemDef
	for _, d := range db.MustDefs().Items {
		if d.Sell >= 100 && d.Sell <= 20000 && realType(d.Type) != 27 && d.ReqOg == 0 && d.Count == 1 {
			item = d
			break
		}
	}
	if item == nil {
		t.Fatal("no storable item definition")
	}
	give := func(s *Session) {
		it := db.Item{Bag: 1, Slot: 0, Type: item.Type, TypeID: item.TypeID, Quality: item.Quality, Count: 1}
		if err := db.AddCharacterItem(s.char.ID, &it); err != nil {
			t.Fatal(err)
		}
		s.items = append(s.items, it)
	}
	bankRows := func() (n int) {
		pool, _ := db.Pool(db.GameData)
		pool.QueryRow(`SELECT COUNT(*) FROM PS_GameData_GuildStoredItems WHERE GuildID = ?1 AND Del = 0`, int32(g.ID)).Scan(&n)
		return
	}
	give(ss[0])
	give(ss[5])
	// Ranking no longer matters: an unranked guild uses its bank.
	ss[5].handleGuildBankMove(1, 0, guildBankBag, 4)  // rank 9
	ss[0].handleGuildBankMove(1, 0, guildBankBag, 45) // beyond the 40 permitted slots
	if g.bank[4] != nil || g.bank[45] != nil {
		t.Fatal("refused deposit stored")
	}
	ss[0].handleGuildBankMove(1, 0, guildBankBag, 3)
	if g.bank[3] == nil || findItem(ss[0].items, 1, 0) >= 0 || bankRows() != 1 {
		t.Fatalf("deposit failed: bank %v rows %d", g.bank[3], bankRows())
	}
	ss[0].handleGuildBankMove(guildBankBag, 3, guildBankBag, 7)
	if g.bank[3] != nil || g.bank[7] == nil || g.bank[7].Slot != 7 {
		t.Fatal("bank rearrange failed")
	}
	ss[5].handleGuildBankMove(guildBankBag, 7, 2, 0) // members cannot withdraw
	if g.bank[7] == nil {
		t.Fatal("rank 9 withdrew")
	}
	money := ss[0].char.Money
	ss[0].handleGuildBankMove(guildBankBag, 7, 2, 5)
	fee := uint32(float32(uint32(item.Sell)) * float32(0.05))
	if g.bank[7] != nil || findItem(ss[0].items, 2, 5) < 0 || ss[0].char.Money != money-fee || bankRows() != 0 {
		t.Fatalf("withdraw: bank %v money %d want %d rows %d", g.bank[7], ss[0].char.Money, money-fee, bankRows())
	}

	// Guild house (0x0D34, 0x0D25, 0x0D35, settlement): only the master buys,
	// upgrades need an active house and etin above the upkeep, donations
	// consume the etin items of the bags.
	houseRow := func() (buy, use, etin, keep, ret int) {
		pool, _ := db.Pool(db.GameData)
		pool.QueryRow(`SELECT BuyHouse,UseHouse,Etin,KeepEtin,EtinReturnCnt FROM PS_GameData_GuildDetails WHERE GuildID=?1`, int32(g.ID)).Scan(&buy, &use, &etin, &keep, &ret)
		return
	}
	ss[5].handleGuild(packet.GuildHouseBuy, []byte{0x34, 0x0d, 0, 0, 0, 0})
	ss[0].handleGuild(packet.GuildHouseBuy, []byte{0x34, 0x0d, 0, 0, 0, 0}) // 4000 gold only
	if g.BuyHouse != 0 {
		t.Fatal("house bought without the right or the money")
	}
	ss[0].char.Money = guildHouse.BuyMoney + 10
	if err := db.SaveCharacter(ss[0].char); err != nil {
		t.Fatal(err)
	}
	ss[0].handleGuild(packet.GuildHouseBuy, []byte{0x34, 0x0d, 0, 0, 0, 0})
	if buy, _, _, keep, _ := houseRow(); g.BuyHouse != 1 || ss[0].char.Money != 10 || buy != 1 || keep != int(guildHouse.DefaultKeep) {
		t.Fatalf("house purchase: guild %d money %d row buy %d keep %d", g.BuyHouse, ss[0].char.Money, buy, keep)
	}
	upgrade := func(s *Session, npcType uint16, level uint8) {
		s.handleGuild(packet.GuildNpcUpgrade, []byte{0x25, 0x0d, byte(npcType), byte(npcType >> 8), level})
	}
	upgrade(ss[0], guildNpcGate, 1) // UseHouse is still 0
	if g.npcLevel(guildNpcGate) != 0 {
		t.Fatal("NPC upgraded in an inactive house")
	}
	guildMu.Lock()
	g.UseHouse, g.Etin = 1, 100
	guildMu.Unlock()
	upgrade(ss[0], guildNpcGate, 1) // etin below the upkeep
	if g.npcLevel(guildNpcGate) != 0 {
		t.Fatal("NPC upgraded without etin")
	}
	guildMu.Lock()
	g.Etin = 5000
	guildMu.Unlock()
	upgrade(ss[5], guildNpcGate, 1) // rank 9
	upgrade(ss[0], guildNpcGate, 2) // a new NPC starts at level 1
	if g.npcLevel(guildNpcGate) != 0 {
		t.Fatal("NPC upgrade accepted from a member or at level 2")
	}
	upgrade(ss[0], guildNpcGate, 1)
	if g.npcs[0] != (guildNpcSlot{Type: guildNpcGate, Level: 1}) || g.Etin != 4000 || g.KeepEtin != guildHouse.DefaultKeep+uint32(guildHouse.Services[0].Keep[1]) {
		t.Fatalf("gate upgrade: slots %v etin %d keep %d", g.npcs, g.Etin, g.KeepEtin)
	}
	upgrade(ss[0], guildNpcGate, 1) // repeated
	upgrade(ss[0], guildNpcBank, 1)
	if g.npcLevel(guildNpcGate) != 1 || g.npcs[1] != (guildNpcSlot{Type: guildNpcBank, Level: 1}) || g.bankSlots() != int(guildHouse.BankSlots[1]) {
		t.Fatalf("NPC slots %v bank slots %d", g.npcs, g.bankSlots())
	}
	if list := g.npcListPacket(); len(list) != 4+1+8 || list[4] != 2 {
		t.Fatalf("NPC list %v", list)
	}
	npcPool, _ := db.Pool(db.GameData)
	var rowsNpc int
	npcPool.QueryRow(`SELECT COUNT(*) FROM PS_GameData_GuildNpcLv WHERE GuildID=?1`, int32(g.ID)).Scan(&rowsNpc)
	if rowsNpc != 2 {
		t.Fatalf("GuildNpcLv rows %d, want 2", rowsNpc)
	}
	etinItem := func(typeID uint8, count uint16, slot uint8) {
		it := db.Item{Bag: 1, Slot: slot, Type: guildEtinType, TypeID: typeID, Count: count}
		if err := db.AddCharacterItem(ss[0].char.ID, &it); err != nil {
			t.Fatal(err)
		}
		ss[0].items = append(ss[0].items, it)
	}
	etinItem(3, 2, 10) // [100]Etin x2
	etinItem(1, 7, 11) // [1]Etin x7
	etinBefore := g.Etin
	ss[0].dispatch(packet.GuildDonate, []byte{0x35, 0x0d, 0, 0, 0, 0})
	if g.Etin != etinBefore+207 || g.EtinReturn != 1 || findItem(ss[0].items, 1, 10) >= 0 || findItem(ss[0].items, 1, 11) >= 0 {
		t.Fatalf("donation: etin %d (want %d) returns %d", g.Etin, etinBefore+207, g.EtinReturn)
	}
	if _, _, etin, _, ret := houseRow(); etin != int(g.Etin) || ret != 1 {
		t.Fatalf("donation row: etin %d returns %d", etin, ret)
	}
	guildMu.Lock()
	keep, etin := g.KeepEtin, g.Etin
	g.settleGuildHouse()
	paid := g.Etin == etin-keep && g.UseHouse == 1
	g.Etin = keep - 1
	g.settleGuildHouse()
	suspended := g.UseHouse == 2
	g.Etin, g.UseHouse = etin, 1
	guildMu.Unlock()
	if !paid || !suspended {
		t.Fatalf("settlement: paid %v suspended %v", paid, suspended)
	}

	// The player initiates the application and the master approves it.
	outsider.guildApply(g.ID)
	ss[0].guildAnswerApplicant(true, outsider.char.ID)
	if guildOf(outsider) != g || g.members[outsider.char.ID].Rank != guildRankMember {
		t.Fatal("approved applicant did not join")
	}
	// Application, accepted by the master.
	applicant.guildApply(g.ID)
	if _, a := rows(g.ID); a != 1 || g.applicants[applicant.char.ID] == nil {
		t.Fatal("application not stored")
	}
	ss[0].guildAnswerApplicant(true, applicant.char.ID)
	if guildOf(applicant) != g {
		t.Fatal("applicant not accepted")
	}
	if m, a := rows(g.ID); m != 9 || a != 0 {
		t.Fatalf("GuildChars members %d applicants %d, want 9/0", m, a)
	}
	// Rank: promote member 1 to sub-master (rank 9 -> 8 per step).
	for range 7 {
		ss[0].guildChangeRank(0, ss[1].char.ID)
	}
	if r := g.members[ss[1].char.ID].Rank; r != guildRankAdmin {
		t.Fatalf("promoted rank %d, want 2", r)
	}
	// A sub-master cannot promote to its own rank.
	ss[1].guildChangeRank(0, ss[2].char.ID)
	ss[1].guildChangeRank(0, ss[2].char.ID)
	before := g.members[ss[2].char.ID].Rank
	for range 7 {
		ss[1].guildChangeRank(0, ss[2].char.ID)
	}
	if r := g.members[ss[2].char.ID].Rank; r != 3 || before != 7 {
		t.Fatalf("sub-master promotion rank %d (before %d), want 3", r, before)
	}
	// Remark, expel, leave.
	ss[0].guildRemark(append([]byte{0x20, 0x0d}, []byte("new notice\x00")...))
	if g.Remark != "new notice" {
		t.Fatalf("remark %q", g.Remark)
	}
	ss[1].guildExpel(outsider.char.ID)
	applicant.guildLeave()
	ss[0].guildLeave() // the master cannot leave a populated guild
	if guildOf(outsider) != nil || guildOf(applicant) != nil || guildOf(ss[0]) != g {
		t.Fatal("expel/leave state wrong")
	}
	if m, _ := rows(g.ID); m != 7 {
		t.Fatalf("GuildChars members %d after expel+leave, want 7", m)
	}
	// Disband by vote: the master and two party admins.
	ss[0].guildChangeRank(0, ss[3].char.ID)
	for range 7 {
		ss[0].guildChangeRank(0, ss[3].char.ID)
	}
	ss[0].guildDisbandRequest()
	ss[0].guildDisbandAnswer(true)
	ss[1].guildDisbandAnswer(true)
	if guildOf(ss[0]) == nil {
		t.Fatal("disbanded with two acceptances")
	}
	ss[3].guildDisbandAnswer(true)
	if guildOf(ss[0]) != nil || guildOf(ss[4]) != nil {
		t.Fatal("guild not disbanded")
	}
	var del int
	pool, _ := db.Pool(db.GameData)
	pool.QueryRow(`SELECT Del FROM PS_GameData_Guilds WHERE GuildID = ?1`, int32(g.ID)).Scan(&del)
	if del != 1 {
		t.Fatalf("Guilds.Del %d, want 1", del)
	}
}
