package world

import (
	"errors"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/admin"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// Operate serializes live edits with selection, gameplay, and logout.
func Operate(req operatorwire.Request) (operatorwire.Response, error) {
	switch req.Action {
	case "ban-ip", "unban-ip":
		return operatorwire.Response{}, changeIPBan(req)
	case "accounts", "create", "ip-bans":
		return admin.Operate(req)
	case "password", "account", "stats", "save-stats", "details", "delete", "restore", "ban-account", "unban-account":
	default:
		return operatorwire.Response{}, errors.New("unsupported live operator action")
	}
	realm.activeAccountsMu.Lock()
	s := realm.activeAccounts[req.AccountID]
	if s == nil {
		// Keep a newly connecting account from loading the character mid-edit.
		defer realm.activeAccountsMu.Unlock()
		return operateOffline(req)
	}
	realm.activeAccountsMu.Unlock()

	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	// Logout can release this account while the request waits for stateMu.
	realm.activeAccountsMu.Lock()
	current := realm.activeAccounts[req.AccountID]
	realm.activeAccountsMu.Unlock()
	if current != s {
		return operatorwire.Response{}, errors.New("account session changed; refresh and retry")
	}
	if s.char != nil && s.char.ID == req.CharacterID {
		switch req.Action {
		case "stats", "save-stats":
			return s.operateCharacter(req)
		case "delete", "restore":
			return operatorwire.Response{}, errors.New("character is online; return to character selection or sign out first")
		}
	}
	res, err := operateOffline(req)
	if err != nil {
		return res, err
	}
	switch {
	case req.Action == "details" && s.char != nil:
		for i := range res.Characters {
			res.Characters[i].Online = res.Characters[i].ID == s.char.ID
		}
	case (req.Action == "delete" || req.Action == "restore") && s.char == nil:
		s.sendCharacterSelect()
	case req.Action == "ban-account", req.Action == "password", req.Action == "account" && req.ExpectedAccount.GM != req.GM:
		// SQLite revoked the handoffs before closing just this account's socket.
		s.authorizationAt = time.Time{}
		netx.Abort(s.conn)
	case req.Action == "account" && s.char != nil:
		s.points = req.Points
		s.send(buildAccountPoints(s.points), "operator point balance")
		s.send(buildCashPoints(s.points), "operator cash points")
	}
	return res, nil
}

func changeIPBan(req operatorwire.Request) error {
	realm.activeAccountsMu.Lock()
	defer realm.activeAccountsMu.Unlock()
	address, err := admin.ChangeIPBan(req)
	if err != nil || req.Action == "unban-ip" {
		return err
	}
	for _, s := range realm.activeAccounts {
		if s.conn != nil && netx.PeerIP(s.conn) == address {
			netx.Abort(s.conn)
		}
	}
	return nil
}

// Deletion also changes the guild roster, which has its own World lock.
func operateOffline(req operatorwire.Request) (operatorwire.Response, error) {
	if req.Action != "delete" {
		return admin.Operate(req)
	}
	guildMu.Lock()
	defer guildMu.Unlock()
	res, err := admin.Operate(req)
	if err == nil {
		guildCharacterDeleted(req.CharacterID)
	}
	return res, err
}

func (s *Session) operateCharacter(req operatorwire.Request) (operatorwire.Response, error) {
	if s.persistenceFailed {
		return operatorwire.Response{}, errors.New("character storage is unavailable")
	}
	next := *s.char
	res, err := db.OperatorCharacterSnapshot(next)
	if err != nil {
		return res, err
	}
	res.Online = true
	if req.Action == "stats" {
		return res, nil
	}
	if req.Stats == nil || req.ExpectedStats == nil {
		return res, errors.New("missing character values")
	}
	old, values := *req.ExpectedStats, *req.Stats
	if err := db.ValidateOperatorStats(res, old, values); err != nil {
		return res, err
	}
	if old == values {
		return res, nil
	}
	next.Str, next.Dex, next.Rec, next.Int, next.Wis, next.Luc = values.Strength, values.Dexterity, values.Rec, values.Intelligence, values.Wisdom, values.Luck
	next.StatPoint, next.SkillPoint = values.StatPoints, values.SkillPoints
	next.Money, next.Exp, next.Kills = values.Gold, values.EXP, values.Kills
	normalize(&next)
	stats := StatsFor(&next, s.equipment())
	next.CurrentHP, next.CurrentMP, next.CurrentSP = stats.CurrentHP, stats.CurrentMP, stats.CurrentSP
	if err := db.SaveCharacter(&next); err != nil {
		return res, err
	}
	*s.char = next
	s.syncPlayer()

	// Use the native refresh packets also used by the in-game GM commands.
	if old.EXP != values.EXP || old.SkillPoints != values.SkillPoints {
		s.send(buildCharacterDetails(&next, &stats), "operator character details")
	}
	if old.StatPoints != values.StatPoints || old.SkillPoints != values.SkillPoints {
		s.send(buildLevelUp(packet.LevelUpSelf, &next), "operator unspent points")
	}
	if old.Gold != values.Gold {
		s.send(buildSetMoney(next.Money), "operator gold balance")
	}
	if old.Kills != values.Kills {
		s.send(buildKillCountUpdate(0, next.Kills), "operator rank kills")
		w := packet.New(0x0522)
		w.U32(next.ID)
		w.U32(next.Kills)
		frame := w.MustFinish()
		s.send(frame, "operator rank")
		s.sendToViewers(frame, "operator rank (view)")
	}
	if old.Strength != values.Strength || old.Dexterity != values.Dexterity || old.Rec != values.Rec ||
		old.Intelligence != values.Intelligence || old.Wisdom != values.Wisdom || old.Luck != values.Luck || old.StatPoints != values.StatPoints {
		s.send(buildStatsReset(next.StatPoint, next.Str, next.Rec, next.Int, next.Wis, next.Dex, next.Luc), "operator base stats")
		s.send(buildPvPRewardStatus(&next), "operator unchanged PvP rewards")
		s.sendStatPackets()
		s.partyPools()
	}
	res.Stats = &values
	return res, nil
}
