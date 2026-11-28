package world

import (
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

func (s *Session) guildApply(guildID uint32) {
	me, ch := s.player, s.char
	guildMu.Lock()
	defer guildMu.Unlock()
	g := guildsByID[guildID]
	if me.guild.guild != nil || me.guild.create != nil || g == nil ||
		g.Country != s.faction || g.applicants[ch.ID] != nil {
		s.send(guildResult(packet.GuildApply, 0), "0x0D07 guild apply refused")
		return
	}
	if old := me.guild.applied; old != nil {
		old.cancelApplication(me)
	}
	m := &guildMember{CharID: ch.ID, Name: ch.Name, Job: ch.Class, Level: ch.Level}
	if err := db.AddGuildChar(g.ID, ch.ID, 0); err != nil {
		s.log.Errorf("GX Server World guild application insert failed: %v", err)
		return
	}
	g.applicants[ch.ID] = m
	guildOfChar[ch.ID] = g
	me.guild.applied = g
	w := packet.New(packet.GuildApplicant)
	writeGuildApplicant(w, m)
	g.sendAdmin(w.MustFinish(), "0x0D17 guild applicant")
	s.send(guildResult(packet.GuildApply, 1), "0x0D07 guild applied")
}

// cancelApplication is CGuild::JoinRequestCancle.
func (g *guild) cancelApplication(p *ConnectedPlayer) {
	if g.applicants[p.CharacterID] == nil {
		return
	}
	if err := db.DeleteGuildChar(g.ID, p.CharacterID); err != nil {
		storageFailure(err)
		return
	}
	delete(g.applicants, p.CharacterID)
	delete(guildOfChar, p.CharacterID)
	p.guild.applied = nil
	g.sendAdmin(guildID32(packet.GuildApplicantRemoved, p.CharacterID), "0x0D18 guild applicant removed")
}

// guildAnswerApplicant (0x0D08): accept byte, char id; JoinOK / JoinNO.
func (s *Session) guildAnswerApplicant(accept bool, charID uint32) {
	guildMu.Lock()
	defer guildMu.Unlock()
	g := s.player.guild.guild
	if g == nil || g.members[s.char.ID] == nil || g.members[s.char.ID].Rank > guildRankAdmin {
		return
	}
	a := g.applicants[charID]
	if a == nil {
		return
	}
	p := FindPlayerByID(charID)
	if !accept {
		if err := db.DeleteGuildChar(g.ID, charID); err != nil {
			s.logErr("guild decline", err)
			return
		}
		delete(g.applicants, charID)
		delete(guildOfChar, charID)
		if p != nil {
			p.guild.applied = nil
			p.Send(buildGuildJoined(0, g, guildRankMember), "0x0D08 guild application declined")
		}
		g.sendAdmin(guildID32(packet.GuildApplicantRemoved, charID), "0x0D18 guild applicant removed")
		return
	}
	if err := db.SaveGuildCharRank(charID, guildRankMember); err != nil {
		s.logErr("guild accept", err)
		return
	}
	delete(g.applicants, charID)
	a.Rank, a.player = guildRankMember, p
	g.members[charID] = a

	g.sendAdmin(guildID32(packet.GuildApplicantRemoved, charID), "0x0D18 guild applicant removed")
	g.send(buildGuildMemberAdded(a), "0x0D14 guild member added", nil)
	if p != nil {
		p.guild.guild, p.guild.applied = g, nil
		p.Send(buildGuildJoined(1, g, a.Rank), "0x0D08 guild application accepted")
		for _, f := range g.memberListPackets() {
			p.Send(f, "0x0D12 guild members")
		}
		for _, f := range g.bankListPackets() {
			p.Send(f, "0x0D29 guild bank")
		}
		if f := g.npcListPacket(); f != nil {
			p.Send(f, "0x0D24 guild NPC levels")
		}
		if f := g.houseStatusPacket(); f != nil {
			p.Send(f, "0x0D33 guild house owned")
		}
	}
}

// guildLeave (0x0D09): CGuild::JoinLeave; the master leaves only alone,
// and an emptied guild is disbanded.
func (s *Session) guildLeave() {
	me := s.player
	var left *guild
	defer func() { guildDuelLeave(me, left, 1) }() // PvPChkLeaveByGuildLeave, after guildMu
	guildMu.Lock()
	defer guildMu.Unlock()
	g := me.guild.guild
	if g == nil || me.guild.create != nil {
		return
	}
	m := g.members[s.char.ID]
	if m == nil || (m.Rank == guildRankMaster && len(g.members) != 1) {
		s.send(guildResult(packet.GuildLeave, 0), "0x0D09 guild leave refused")
		return
	}
	var err error
	if len(g.members) == 1 {
		err = db.DeleteLastGuildChar(g.ID, s.char.ID)
	} else {
		err = db.DeleteGuildChar(g.ID, s.char.ID)
	}
	if err != nil {
		s.logErr("guild leave", err)
		return
	}

	delete(g.members, s.char.ID)
	delete(guildOfChar, s.char.ID)
	me.guild.guild = nil
	left = g
	s.send(guildResult(packet.GuildLeave, 1), "0x0D09 guild left")
	g.send(guildStatus(guildStatusLeft, s.char.ID), "0x0D0C guild member left", nil)
	if len(g.members) == 0 {
		g.fire(s.logger, true)
	}
}

// guildExpel (0x0D0A): CGuild::JoinExpel, only lower ranks.
func (s *Session) guildExpel(charID uint32) {
	var expelled *ConnectedPlayer
	var from *guild
	defer func() {
		if expelled != nil {
			guildDuelLeave(expelled, from, 1)
		}
	}()
	guildMu.Lock()
	defer guildMu.Unlock()
	g := s.player.guild.guild
	if g == nil || s.player.guild.create != nil {
		return
	}
	me := g.members[s.char.ID]
	t := g.members[charID]
	if me == nil || me.Rank > guildRankAdmin {
		return
	}
	if t == nil || t.Rank <= me.Rank {
		s.send(guildResultID(packet.GuildExpel, 0, charID), "0x0D0A guild expel refused")
		return
	}
	if err := db.DeleteGuildChar(g.ID, charID); err != nil {
		s.logErr("guild expulsion", err)
		return
	}
	delete(g.members, charID)
	delete(guildOfChar, charID)
	if t.player != nil {
		t.player.Send(guildResultID(packet.GuildExpel, 1, charID), "0x0D0A guild expelled")
		t.player.guild.guild = nil
		expelled, from = t.player, g
	}
	s.send(guildResultID(packet.GuildExpel, 1, charID), "0x0D0A guild expel done")
	g.send(guildStatus(guildStatusExpel, charID), "0x0D0C guild member expelled", nil)
}

// guildChangeRank (0x0D0C): direction (0 promote, 1 demote), char id;
// CGuild::ChangeGuildUserLevel.
func (s *Session) guildChangeRank(direction uint8, charID uint32) {
	if charID == s.char.ID || direction > 1 {
		return
	}
	guildMu.Lock()
	defer guildMu.Unlock()
	g := s.player.guild.guild
	if g == nil {
		return
	}
	me, t := g.members[s.char.ID], g.members[charID]
	if me == nil || t == nil {
		return
	}
	next, ok := guildRankChange(me.Rank, t.Rank, direction, g.admins())
	if !ok {
		return
	}
	if err := db.SaveGuildCharRank(charID, next); err != nil {
		s.logErr("guild rank", err)
		return
	}
	t.Rank = next
	g.send(guildStatus(guildStatusRank+t.Rank, charID), "0x0D0C guild rank", nil)
}

// guildRemark (0x0D20): remark[65]; master only, no quote characters.

func guildRankChange(actor, target, direction uint8, admins int) (uint8, bool) {
	if actor > guildRankAdmin || target <= actor || direction > 1 {
		return 0, false
	}
	next := int(target) + 1
	if direction == 0 {
		next = int(target) - 1
	}
	if next < guildRankAdmin || next > guildRankMember {
		return 0, false
	}
	if next == guildRankAdmin && (actor == guildRankAdmin || admins >= guildMaxAdmins) {
		return 0, false
	}
	return uint8(next), true
}
