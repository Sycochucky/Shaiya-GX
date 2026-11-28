package world

import (
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
	"strings"
	"sync"
)

// Guilds follow ps_game.exe CGuildMgr / CGuild / CGuildCreate and
// CUser::PacketGuild 0x004764a0. The original game server delegated every
// write to the DB agent (0x0Bxx); Rapture calls the same procedures.
//
// Ranks are the guild levels: 1 master, 2 sub-masters (with the master the
// "admins", at most 7), 3-9 members (new members join at 9); 0 is a
// pending applicant.

const (
	guildRankMaster  = 1
	guildRankAdmin   = 2
	guildRankMember  = 9
	guildMaxAdmins   = 7
	guildListBatch   = 14 // 0x0D2F entries per packet
	guildMemberBatch = 40 // 0x0D12/0x0D13/0x0D16 entries per packet

	// 0x0D0C status modes; a level byte (1-80) is a character level change
	// and 200 + rank a rank change.
	guildStatusDisband = 0x65
	guildStatusLeft    = 0x66
	guildStatusExpel   = 0x67
	guildStatusOnline  = 0x68
	guildStatusOffline = 0x69
	guildStatusRank    = 200
)

type guildMember struct {
	CharID uint32
	Name   string
	Job    uint8
	Level  uint16
	Rank   uint8
	player *ConnectedPlayer // nil when offline
}

type guild struct {
	ID         uint32
	Name       string
	Master     string
	Remark     string
	Country    uint8
	Points     uint32
	Rank       uint8
	Etin       uint32
	BuyHouse   uint8
	UseHouse   uint8
	EtinReturn uint32
	KeepEtin   uint32
	// gmHouse: founded by a GM account. A test guild whose house is always
	// active, ranked and paid, with no GRB or etin upkeep.
	gmHouse bool
	// gmGRB: a GM owner is inside the guild's ranking instance outside the
	// normal schedule; points are credited when that GM leaves.
	gmGRB bool
	// Monster points earned in the current guild ranking battle.
	grbPoints  uint32
	gd         guildDuel // guarded by duelMu
	members    map[uint32]*guildMember
	applicants map[uint32]*guildMember
	bank       [guildBankSlots]*db.Item
	npcs       [guildNpcSlots]guildNpcSlot // CGuild +0x47c, guarded by guildMu
}

// guildCreate is CGuildCreate: a creation (disband false) or disband vote.
type guildCreate struct {
	disband  bool
	name     string
	remark   string
	masterID uint32
	master   *ConnectedPlayer
	pending  map[uint32]bool // ids that have not accepted yet
	players  []*ConnectedPlayer
	accepted int
	guild    *guild // the guild being disbanded
}

// Per-player guild state, guarded by guildMu.
type playerGuildState struct {
	guild   *guild
	create  *guildCreate
	applied *guild
	userID  string
}

var (
	guildMu      sync.Mutex
	guildsByID   = map[uint32]*guild{}
	guildOfChar  = map[uint32]*guild{} // members and applicants
	nextGuildID  uint32
	guildsLoaded bool
)

// LoadGuilds is CGuildMgr::InitAddGuild / InitAddUser from the database.
func LoadGuilds(log *logx.Logger) error {
	rows, members, err := db.LoadGuilds()
	if err != nil {
		return err
	}
	next, err := db.NextGuildID()
	if err != nil {
		return err
	}
	guildMu.Lock()
	defer guildMu.Unlock()
	guildsByID, guildOfChar = map[uint32]*guild{}, map[uint32]*guild{}
	for _, r := range rows {
		guildsByID[r.ID] = &guild{ID: r.ID, Name: r.Name, Master: r.MasterName, Remark: r.Remark, Country: r.Country,
			Points: r.Points, Rank: r.Rank, Etin: r.Etin, BuyHouse: r.BuyHouse, UseHouse: r.UseHouse,
			EtinReturn: r.EtinReturn, KeepEtin: r.KeepEtin, gmHouse: r.MasterAdmin, members: map[uint32]*guildMember{}, applicants: map[uint32]*guildMember{}}
		guildsByID[r.ID].forceGMHouse()
		next = max(next, r.ID+1)
	}
	for _, m := range members {
		g := guildsByID[m.GuildID]
		if g == nil {
			continue
		}
		e := &guildMember{CharID: m.CharID, Name: m.Name, Job: m.Job, Level: m.Level, Rank: m.Rank}
		// InitAddUser: the guild's master name always holds rank 1.
		if strings.EqualFold(m.Name, g.Master) {
			e.Rank = guildRankMaster
		}
		if e.Rank == 0 {
			g.applicants[e.CharID] = e
		} else {
			g.members[e.CharID] = e
		}
		guildOfChar[e.CharID] = g
	}
	loadGuildHouseIni()
	npcRows, err := db.LoadGuildNpcLevels()
	if err != nil {
		return err
	}
	for _, r := range npcRows {
		// DBGuild::LoadNpcLvList: a valid slot overwrites only its row.
		if g := guildsByID[r.GuildID]; g != nil && int(r.Slot) < guildNpcSlots {
			g.npcs[r.Slot] = guildNpcSlot{Type: r.Type, Level: uint16(r.Level)}
		}
	}
	for _, g := range guildsByID {
		if err := loadGuildBank(g); err != nil {
			log.Printf("GX Server World failed to load guild bank %d: %v", g.ID, err)
		}
	}
	nextGuildID, guildsLoaded = next, true
	log.Printf("GX Server World loaded guilds=%d members=%d next_id=%d", len(rows), len(members), next)
	return nil
}

func (g *guild) admins() int {
	n := 0
	for _, m := range g.members {
		if m.Rank <= guildRankAdmin {
			n++
		}
	}
	return n
}

// send is CGuild::Send: every online member.
func (g *guild) send(frame []byte, label string, except *ConnectedPlayer) {
	for _, m := range g.members {
		if m.player != nil && m.player != except {
			m.player.Send(frame, label)
		}
	}
}

// sendAdmin is CGuild::SendAdmin: the online master and sub-masters.
func (g *guild) sendAdmin(frame []byte, label string) {
	for _, m := range g.members {
		if m.player != nil && m.Rank <= guildRankAdmin {
			m.player.Send(frame, label)
		}
	}
}

func guildNameFor(charID uint32) string {
	guildMu.Lock()
	defer guildMu.Unlock()
	if g := guildOfChar[charID]; g != nil && g.members[charID] != nil {
		return g.Name
	}
	return ""
}

func guildIDFor(charID uint32) uint32 {
	guildMu.Lock()
	defer guildMu.Unlock()
	if g := guildOfChar[charID]; g != nil && g.members[charID] != nil {
		return g.ID
	}
	return 0
}

// deleteCharacter synchronizes self-service deletion with the guild roster.
func deleteCharacter(accountID, charID uint32) (bool, error) {
	guildMu.Lock()
	defer guildMu.Unlock()
	ok, err := db.DeleteCharacter(accountID, charID)
	if ok && err == nil {
		guildCharacterDeleted(charID)
	}
	return ok, err
}

// guildCharacterDeleted publishes the durable deletion while guildMu is held.
func guildCharacterDeleted(charID uint32) {
	g := guildOfChar[charID]
	if g == nil {
		return
	}
	delete(guildOfChar, charID)
	if g.applicants[charID] != nil {
		delete(g.applicants, charID)
		g.sendAdmin(guildID32(packet.GuildApplicantRemoved, charID), "0x0D18 guild applicant removed")
	}
	if g.members[charID] != nil {
		delete(g.members, charID)
		g.send(guildStatus(guildStatusLeft, charID), "0x0D0C guild member left", nil)
	}
}

// guildConnect is CUser::InitGuild + CGuildMgr::ConnectUser +
// SendCharacterGuild at world entry, then CGuildMgr::SendGuildList.
func (s *Session) guildConnect() {
	p, ch := s.player, s.char
	if p == nil || ch == nil {
		return
	}
	guildMu.Lock()
	p.guild.userID = s.account.UserID
	var lists [][]byte
	if g := guildOfChar[ch.ID]; g != nil {
		if m := g.members[ch.ID]; m != nil {
			m.player, m.Level, m.Job, m.Name = p, ch.Level, ch.Class, ch.Name
			p.guild.guild = g
			g.send(guildStatus(guildStatusOnline, ch.ID), "0x0D0C guild member online", p)
			lists = append(lists, g.memberListPackets()...)
			lists = append(lists, g.bankListPackets()...)
			if f := g.npcListPacket(); f != nil {
				lists = append(lists, f)
			}
			if f := g.houseStatusPacket(); f != nil {
				lists = append(lists, f)
			}
			if m.Rank <= guildRankAdmin {
				lists = append(lists, g.applicantListPackets()...)
			}
		} else if g.applicants[ch.ID] != nil {
			p.guild.applied = g
		}
	}
	lists = append(lists, guildListPackets(s.faction)...)
	guildMu.Unlock()
	for _, f := range lists {
		s.send(f, "guild list")
	}
}

// guildListPackets is CGuildMgr::SendGuildList: 0x0D0D, the country's
// guilds in 0x0D2F batches of 14, 0x0D0E.
func (s *Session) guildDisconnect() {
	p := s.player
	if p == nil || s.char == nil {
		return
	}
	guildMu.Lock()
	defer guildMu.Unlock()
	if c := p.guild.create; c != nil {
		c.fail(7)
	}
	if g := p.guild.guild; g != nil {
		if m := g.members[s.char.ID]; m != nil {
			m.player = nil
		}
		g.send(guildStatus(guildStatusOffline, s.char.ID), "0x0D0C guild member offline", p)
	}
	p.guild = playerGuildState{}
}

// guildLevelChanged is CGuild::ChangeUserLevel on a level-up.
func (s *Session) guildLevelChanged() {
	guildMu.Lock()
	defer guildMu.Unlock()
	if g := s.player.guild.guild; g != nil {
		if m := g.members[s.char.ID]; m != nil {
			m.Level = s.char.Level
			g.send(guildStatus(uint8(s.char.Level), s.char.ID), "0x0D0C guild member level", nil)
		}
	}
}

// guildChat (0x1104) is CGuild::SendAllGuildChat: sender name, text.
func (s *Session) guildChat(message string, items proto.ChatItems) {
	guildMu.Lock()
	defer guildMu.Unlock()
	g := s.player.guild.guild
	if g == nil {
		return
	}
	w := packet.New(packet.ChatGuild)
	w.FixedString(s.char.Name, 21)
	writeChatText(w, message)
	g.send(chatWithItems(w.MustFinish(), items), "0x1104 guild chat", nil)
}

// handleGuild dispatches the 0x0Dxx requests of CUser::PacketGuild.
func (s *Session) handleGuild(op packet.Opcode, p []byte) {
	if s.char == nil || s.player == nil {
		return
	}
	switch op {
	case packet.GuildCreate:
		s.guildCreateRequest(p)
	case packet.GuildCreateAgree:
		if len(p) >= 3 {
			s.guildCreateAnswer(p[2] == 1)
		}
	case packet.GuildDisband:
		s.guildDisbandRequest()
	case packet.GuildDisbandAgree:
		if len(p) >= 3 {
			s.guildDisbandAnswer(p[2] == 1)
		}
	case packet.GuildApply:
		if len(p) >= 6 {
			id, _ := packet.U32(p[2:])
			s.guildApply(id)
		}
	case packet.GuildJoinResult:
		if len(p) >= 7 {
			id, _ := packet.U32(p[3:])
			s.guildAnswerApplicant(p[2] == 1, id)
		}
	case packet.GuildLeave:
		s.guildLeave()
	case packet.GuildExpel:
		if len(p) >= 6 {
			id, _ := packet.U32(p[2:])
			s.guildExpel(id)
		}
	case packet.GuildStatus:
		if len(p) >= 7 {
			id, _ := packet.U32(p[3:])
			s.guildChangeRank(p[2], id)
		}
	case packet.GuildRemark:
		s.guildRemark(p)
	case packet.GuildEtin:
		guildMu.Lock()
		g := s.player.guild.guild
		var etin uint32
		if g != nil {
			etin = g.Etin
		}
		guildMu.Unlock()
		if g != nil {
			s.send(guildID32(packet.GuildEtin, etin), "0x0D23 guild etin")
		}
	case packet.GuildNpcUpgrade:
		if len(p) < 5 {
			return
		}
		guildMu.Lock()
		defer guildMu.Unlock()
		if g := s.player.guild.guild; g != nil && g.members[s.char.ID] != nil && g.members[s.char.ID].Rank <= guildRankAdmin {
			s.guildNpcUpgrade(g, uint16(p[2])|uint16(p[3])<<8, p[4])
		}
	case packet.GuildHouseBuy:
		guildMu.Lock()
		defer guildMu.Unlock()
		g := s.player.guild.guild
		rank := uint8(0)
		if g != nil && g.members[s.char.ID] != nil {
			rank = g.members[s.char.ID].Rank
		} else {
			g = nil
		}
		s.guildHouseBuy(g, rank)
	case packet.GuildDonate:
		guildMu.Lock()
		defer guildMu.Unlock()
		if g := s.player.guild.guild; g != nil && g.members[s.char.ID] != nil {
			s.guildDonate(g)
		}
	}
}

func fixedText(p []byte, at, n int) string {
	if len(p) < at {
		return ""
	}
	b := p[at:min(len(p), at+n)]
	for i, c := range b {
		if c == 0 {
			b = b[:i]
			break
		}
	}
	return packet.FixedText(b)
}

// guildDisbandRequest (0x0D03) starts the disband vote.
func (s *Session) guildDisbandRequest() {
	me := s.player
	partyMu.Lock()
	var party []*ConnectedPlayer
	if pt := partyOf(me); pt != nil {
		party = append(party, pt.members...)
	}
	partyMu.Unlock()
	guildMu.Lock()
	defer guildMu.Unlock()
	g := me.guild.guild
	m := (*guildMember)(nil)
	if g != nil {
		m = g.members[s.char.ID]
	}
	if g == nil || me.guild.create != nil || m == nil || m.Rank != guildRankMaster {
		s.send(guildResult(packet.GuildDisband, 0), "0x0D03 guild disband refused")
		return
	}
	if len(g.members) <= 2 {
		g.fire(s.logger)
		return
	}
	c := &guildCreate{disband: true, masterID: s.char.ID, master: me, guild: g, pending: map[uint32]bool{}}
	c.players = append(c.players, me)
	c.pending[s.char.ID] = true
	for _, p := range party {
		if p == me || p.MapID() != s.char.MapID || p.guild.create != nil {
			continue
		}
		if pm := g.members[p.CharacterID]; pm != nil && pm.Rank <= guildRankAdmin {
			c.players = append(c.players, p)
			c.pending[p.CharacterID] = true
		}
	}
	if len(c.players) < 3 {
		s.send(guildResult(packet.GuildDisband, 2), "0x0D03 guild disband refused")
		return
	}
	for _, p := range c.players {
		p.guild.create = c
	}
	c.sendAll(guildID32(packet.GuildDisbandAgree, s.char.ID), "0x0D04 guild disband agree")
}

// guildDisbandAnswer (0x0D04): more than 2 acceptances disband it.
func (s *Session) guildDisbandAnswer(accept bool) {
	guildMu.Lock()
	defer guildMu.Unlock()
	c := s.player.guild.create
	if c == nil || !c.disband || s.player.guild.guild == nil {
		return
	}
	if !accept {
		c.fail(3)
		return
	}
	if !c.pending[s.char.ID] {
		return
	}
	delete(c.pending, s.char.ID)
	if c.accepted++; c.accepted > 2 {
		c.reset()
		c.guild.fire(s.logger)
	}
}

// fire is CGuildMgr::FireGuild / CGuild::Fire.
func (g *guild) fire(log *logx.Logger, alreadyDeleted ...bool) {
	if len(alreadyDeleted) == 0 || !alreadyDeleted[0] {
		if err := db.DeleteGuild(g.ID); err != nil {
			log.Errorf("GX Server World guild delete failed: %v", err)
			return
		}
	}

	g.send(guildStatus(guildStatusDisband, 0), "0x0D0C guild disbanded", nil)
	for id, m := range g.members {
		if m.player != nil {
			m.player.guild.guild = nil
		}
		delete(guildOfChar, id)
	}
	for id, m := range g.applicants {
		if p := FindPlayerByID(m.CharID); p != nil {
			p.guild.applied = nil
		}
		delete(guildOfChar, id)
	}
	delete(guildsByID, g.ID)

	guildSendCountry(g.Country, guildID32(packet.GuildDisbanded, g.ID), "0x0D11 guild disbanded")
	log.Printf("GX Server World guild disbanded id=%d name=%q", g.ID, g.Name)
}

// guildRemark (0x0D20) updates the notice for the entire guild.
func (s *Session) guildRemark(p []byte) {
	remark := fixedText(p, 2, 64)
	reply := func(result uint8) {
		w := packet.New(packet.GuildRemark)
		w.U8(result)
		w.FixedString(remark, 65)
		s.send(w.MustFinish(), "0x0D20 guild remark")
	}
	guildMu.Lock()
	defer guildMu.Unlock()
	g := s.player.guild.guild
	if g == nil {
		return
	}
	if m := g.members[s.char.ID]; m == nil || m.Rank != guildRankMaster || strings.ContainsRune(remark, '\'') {
		reply(0)
		return
	}
	if err := db.SaveGuildRemark(g.ID, remark); err != nil {
		s.logErr("guild remark", err)
		reply(0)
		return
	}
	g.Remark = remark
	w := packet.New(packet.GuildRemark)
	w.U8(1)
	w.FixedString(remark, 65)
	g.send(w.MustFinish(), "0x0D20 guild remark", nil)

}
