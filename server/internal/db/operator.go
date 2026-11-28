package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/Spelunkern/Shaiya-GX/internal/pvprank"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

const operatorAccountLimit = 100

type rowQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func OperatorAccounts(query string) ([]operatorwire.Account, bool, error) {
	p, err := ReadPool()
	if err != nil {
		return nil, false, err
	}
	c, cancel := ctx()
	defer cancel()
	rows, err := p.QueryContext(c, `SELECT U.UserUID,U.UserID,COALESCE(U.Admin,0)<>0,MAX(0,COALESCE(U.Point,0)),B.UserUID IS NOT NULL,COALESCE(B.Reason,'')
FROM PS_UserData_Users_Master U LEFT JOIN GX_OperatorAccountBans B ON B.UserUID=U.UserUID
WHERE U.Leave=0 AND instr(lower(U.UserID),lower(?))>0 ORDER BY U.UserID COLLATE NOCASE,U.UserUID LIMIT ?`, query, operatorAccountLimit+1)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	var out []operatorwire.Account
	for rows.Next() {
		var a operatorwire.Account
		if err := rows.Scan(&a.ID, &a.Name, &a.GM, &a.Points, &a.Banned, &a.BanReason); err != nil {
			return nil, false, err
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	return out[:min(len(out), operatorAccountLimit)], len(out) > operatorAccountLimit, nil
}

func operatorAccount(c context.Context, q rowQueryer, uid uint32) (operatorwire.Account, error) {
	var a operatorwire.Account
	err := q.QueryRowContext(c, `SELECT U.UserUID,U.UserID,COALESCE(U.Admin,0)<>0,MAX(COALESCE(U.Point,0),0),B.UserUID IS NOT NULL,COALESCE(B.Reason,'')
FROM PS_UserData_Users_Master U LEFT JOIN GX_OperatorAccountBans B ON B.UserUID=U.UserUID WHERE U.UserUID=? AND U.Leave=0`,
		uid).Scan(&a.ID, &a.Name, &a.GM, &a.Points, &a.Banned, &a.BanReason)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNotFound
	}
	return a, err
}

func OperatorAccount(uid uint32) (operatorwire.Account, error) {
	p, err := ReadPool()
	if err != nil {
		return operatorwire.Account{}, err
	}
	c, cancel := ctx()
	defer cancel()
	return operatorAccount(c, p, uid)
}

func SetOperatorAccount(expected operatorwire.Account, gm bool, points uint32) error {
	if points > math.MaxInt32 {
		return fmt.Errorf("Item Mall points must be at most %d", math.MaxInt32)
	}
	p, err := Pool(UserData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	return withTx(c, p, func(tx *sql.Tx) error {
		current, err := operatorAccount(c, tx, expected.ID)
		if err != nil {
			return err
		}
		if current != expected {
			return errors.New("account changed since it was opened; refresh before saving")
		}
		if _, err := tx.ExecContext(c, `UPDATE PS_UserData_Users_Master SET Admin=?,Point=? WHERE UserUID=?`, gm, points, current.ID); err != nil {
			return err
		}
		if gm != current.GM {
			return revokeCredentialTx(c, tx, current.ID)
		}
		return nil
	})
}

// OperatorCharacterSnapshot exposes the editable fields and experience bounds.
func OperatorCharacterSnapshot(ch Character) (operatorwire.Response, error) {
	minEXP, maxEXP, err := operatorEXPBounds(ch.Level)
	if err != nil {
		return operatorwire.Response{}, err
	}
	return operatorwire.Response{
		Stats: &operatorwire.Stats{
			Strength: ch.Str, Dexterity: ch.Dex, Rec: ch.Rec, Intelligence: ch.Int,
			Wisdom: ch.Wis, Luck: ch.Luc, StatPoints: ch.StatPoint, SkillPoints: ch.SkillPoint,
			Gold: ch.Money, EXP: ch.Exp, Kills: ch.Kills,
		},
		MinEXP: minEXP, MaxEXP: maxEXP,
	}, nil
}

func operatorEXPBounds(level uint16) (uint32, uint32, error) {
	defs, err := Definitions()
	if err != nil {
		return 0, 0, err
	}
	if level == 0 || level > defs.LevelCap {
		return 0, 0, errors.New("character level is outside the GX experience table")
	}
	floor := defs.Exp[level-1]
	if level == defs.LevelCap {
		return floor, floor, nil
	}
	return floor, max(floor+1, defs.Exp[level]) - 1, nil
}

func OperatorStats(uid, charID uint32) (operatorwire.Response, error) {
	p, err := ReadPool()
	if err != nil {
		return operatorwire.Response{}, err
	}
	c, cancel := ctx()
	defer cancel()
	var ch Character
	var money, exp int64
	err = p.QueryRowContext(c, `SELECT COALESCE(Level,1),COALESCE(Str,0),COALESCE(Dex,0),COALESCE(Rec,0),
COALESCE(Int,0),COALESCE(Wis,0),COALESCE(Luc,0),COALESCE(StatPoint,0),COALESCE(SkillPoint,0),
COALESCE(Money,0),COALESCE(Exp,0),COALESCE(K1,0)
FROM PS_GameData_Chars WHERE ServerID=? AND UserUID=? AND CharID=? AND Del=0`, config.ServerID, uid, charID).
		Scan(&ch.Level, &ch.Str, &ch.Dex, &ch.Rec, &ch.Int, &ch.Wis, &ch.Luc, &ch.StatPoint, &ch.SkillPoint, &money, &exp, &ch.Kills)
	if errors.Is(err, sql.ErrNoRows) {
		return operatorwire.Response{}, ErrNotFound
	}
	if err != nil {
		return operatorwire.Response{}, err
	}
	ch.Money, ch.Exp = uint32(money), uint32(exp)
	return OperatorCharacterSnapshot(ch)
}

// ValidateOperatorStats rejects stale forms and checks the native field bounds.
func ValidateOperatorStats(snapshot operatorwire.Response, expected, next operatorwire.Stats) error {
	if *snapshot.Stats != expected {
		return errors.New("character changed since it was opened; reopen stats before saving")
	}
	if max(next.Strength, next.Dexterity, next.Rec, next.Intelligence, next.Wisdom, next.Luck) > BaseStatCap ||
		max(next.StatPoints, next.SkillPoints) > math.MaxInt16 || next.Gold > MoneyCap || next.Kills > pvprank.CounterCap {
		return errors.New("character values exceed GX stat, point, gold or rank kill limits")
	}
	if next.EXP != expected.EXP && (next.EXP < snapshot.MinEXP || next.EXP > snapshot.MaxEXP) {
		return fmt.Errorf("total EXP for the current level must be between %d and %d", snapshot.MinEXP, snapshot.MaxEXP)
	}
	return nil
}

func SetOperatorStats(uid, charID uint32, expected, next operatorwire.Stats) error {
	// The caller holds World ownership or the account session lock.
	snapshot, err := OperatorStats(uid, charID)
	if err != nil {
		return err
	}
	if err := ValidateOperatorStats(snapshot, expected, next); err != nil {
		return err
	}
	p, err := Pool(GameData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	result, err := p.ExecContext(c, `UPDATE PS_GameData_Chars SET Str=?,Dex=?,Rec=?,Int=?,Wis=?,Luc=?,StatPoint=?,SkillPoint=?,Money=?,Exp=?,K1=?
WHERE ServerID=? AND UserUID=? AND CharID=? AND Del=0`,
		next.Strength, next.Dexterity, next.Rec, next.Intelligence, next.Wisdom, next.Luck,
		next.StatPoints, next.SkillPoints, int32(next.Gold), int32(next.EXP), next.Kills, config.ServerID, uid, charID)
	if err != nil {
		return err
	}
	return requireOne(result)
}
