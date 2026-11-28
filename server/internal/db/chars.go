package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

// Character mirrors PS_GameData.dbo.Chars. CurrentHP/MP/SP are base values
// (see world.CharacterStats); Chars.HP/MP/SP store them.
type Character struct {
	ID, AccountID                     uint32
	Slot                              uint8
	Name                              string
	Level                             uint16
	MapID                             uint16
	PosX, PosY, PosZ                  float32
	Dir                               int16
	Race, Mode                        uint8 // Family, Grow
	Hair, Face, Height                uint8
	Class, Gender                     uint8 // Job, Sex
	Str, Rec, Int, Wis, Dex, Luc      uint16
	StatPoint, SkillPoint             uint16
	MaxHP, MaxMP, MaxSP               uint32
	CurrentHP, CurrentMP, CurrentSP   uint32
	Exp, Money                        uint32
	Kills, Deaths, Victories, Defeats uint32
	KillLevel, DeadLevel              uint8
	Hg, Vg                            int16
	Cg, Og, Ig                        uint8
	PendingDelete, RenamePending      bool
}

// NewCharacter is a character to create: the client's request plus the
// starting stats, pools and position the world server derives for it.
type NewCharacter struct {
	Slot, Race, Mode, Hair, Face, Height, Class, Gender uint8
	Name                                                string
	MapID                                               uint16
	PosX, PosY, PosZ                                    float32
	Str, Dex, Rec, Int, Wis, Luc                        uint16
	HP, MP, SP                                          uint16
}

// DeleteWindow is how long a deleted character stays restorable.
const DeleteWindow = 2 * time.Hour

// MoneyCap keeps gold within the signed Money column's bit pattern.
const MoneyCap = 4_000_000_000

// BaseStatCap is the configured character base-stat limit, within the signed
// smallint character columns and SaveCharacter parameters.
const BaseStatCap = 32000

func u16(v int64) uint16 { return clampU16(v) }

// CharactersForAccount lists live characters (usp_Read_Chars_R) plus those
// deleted within DeleteWindow, flagged PendingDelete.
func CharactersForAccount(userUID uint32) ([]Character, error) {
	pool, err := ReadPool()
	if err != nil {
		return nil, err
	}
	c, cancel := ctx()
	defer cancel()
	rows, err := pool.QueryContext(c, `SELECT CharID, CharName, Slot,
Family, Grow, Hair, Face, [Size], Job, Sex, [Level], [Str], Dex, Rec, [Int], Luc, Wis,
HP, MP, SP, Map, RenameCnt, RemainTime,
CAST(strftime('%Y',RegDate) AS INTEGER) AS MakeTime_YYYY,
CAST(strftime('%m',RegDate) AS INTEGER) AS MakeTime_MM,
CAST(strftime('%d',RegDate) AS INTEGER) AS MakeTime_DD,
CAST(strftime('%H',RegDate) AS INTEGER) AS MakeTime_HH,
CAST(strftime('%M',RegDate) AS INTEGER) AS MakeTime_MI,
CAST(strftime('%S',RegDate) AS INTEGER) AS MakeTime_SS

FROM PS_GameData_Chars

WHERE ServerID=@serverid AND UserUID=@useruid AND Del=@del`,
		sql.Named("serverid", config.ServerID), sql.Named("useruid", int32(userUID)), sql.Named("del", 0))
	if err != nil {
		return nil, fmt.Errorf("usp_Read_Chars_R: %w", err)
	}
	var chars []Character
	for rows.Next() {
		var ch Character
		var level, str, dex, rec, intel, luc, wis, hp, mp, sp, mapID int64
		var renameCnt uint8
		var remain int32
		var mk [6]int32
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Slot, &ch.Race, &ch.Mode, &ch.Hair, &ch.Face, &ch.Height,
			&ch.Class, &ch.Gender, &level, &str, &dex, &rec, &intel, &luc, &wis, &hp, &mp, &sp, &mapID,
			&renameCnt, &remain, &mk[0], &mk[1], &mk[2], &mk[3], &mk[4], &mk[5]); err != nil {
			rows.Close()
			return nil, err
		}
		ch.AccountID = userUID
		ch.Name = strings.TrimRight(ch.Name, " ")
		ch.Level, ch.MapID = max(u16(level), 1), u16(mapID)
		ch.Str, ch.Dex, ch.Rec, ch.Int, ch.Luc, ch.Wis = u16(str), u16(dex), u16(rec), u16(intel), u16(luc), u16(wis)
		ch.MaxHP, ch.MaxMP, ch.MaxSP = uint32(max(hp, 0)), uint32(max(mp, 0)), uint32(max(sp, 0))
		ch.CurrentHP, ch.CurrentMP, ch.CurrentSP = ch.MaxHP, ch.MaxMP, ch.MaxSP
		ch.RenamePending = renameCnt > 0
		chars = append(chars, ch)
	}
	rows.Close()
	// No original procedure lists recently deleted characters with their
	// DeleteDate, so the restore window is read directly.
	pending, err := pool.QueryContext(c, `SELECT CharID, CharName, Slot, Family, Grow, Hair, Face, Size, Job, Sex,
		Level, Str, Dex, Rec, [Int], Luc, Wis, HP, MP, SP, Map, RenameCnt FROM PS_GameData_Chars
		WHERE ServerID = ?1 AND UserUID = ?2 AND Del = 1 AND DeleteDate >= datetime('now','-2 hours')
		ORDER BY DeleteDate DESC, CharID DESC`,
		config.ServerID, int32(userUID))
	if err != nil {
		return nil, err
	}
	defer pending.Close()
	for pending.Next() {
		var ch Character
		var level, str, dex, rec, intel, luc, wis, hp, mp, sp, mapID int64
		var renameCnt uint8
		if err := pending.Scan(&ch.ID, &ch.Name, &ch.Slot, &ch.Race, &ch.Mode, &ch.Hair, &ch.Face, &ch.Height,
			&ch.Class, &ch.Gender, &level, &str, &dex, &rec, &intel, &luc, &wis, &hp, &mp, &sp, &mapID, &renameCnt); err != nil {
			return nil, err
		}
		ch.AccountID = userUID
		ch.Name = strings.TrimRight(ch.Name, " ")
		ch.Level, ch.MapID = max(u16(level), 1), u16(mapID)
		ch.Str, ch.Dex, ch.Rec, ch.Int, ch.Luc, ch.Wis = u16(str), u16(dex), u16(rec), u16(intel), u16(luc), u16(wis)
		ch.MaxHP, ch.MaxMP, ch.MaxSP = uint32(max(hp, 0)), uint32(max(mp, 0)), uint32(max(sp, 0))
		ch.PendingDelete, ch.RenamePending = true, renameCnt > 0
		chars = append(chars, ch)
	}
	return chars, nil
}

// CharacterDetail loads a playable character with
// usp_Read_Chars_Detail2_R (which also stamps JoinDate). It returns
// ErrNotFound unless the character belongs to userUID and is not deleted.
func CharacterDetail(userUID, charID uint32) (Character, error) {
	ch, err := readCharacterDetail(userUID, charID)
	if err != nil {
		return Character{}, err
	}
	pool, err := Pool(GameData)
	if err != nil {
		return Character{}, err
	}
	c, cancel := ctx()
	defer cancel()
	result, err := pool.ExecContext(c, `UPDATE PS_GameData_Chars SET JoinDate=CURRENT_TIMESTAMP WHERE CharID=? AND UserUID=? AND Del=0`, charID, userUID)
	if err != nil {
		return Character{}, err
	}
	if err := requireOne(result); err != nil {
		return Character{}, err
	}
	return ch, nil
}

func readCharacterDetail(userUID, charID uint32) (Character, error) {
	pool, err := ReadPool()
	if err != nil {
		return Character{}, err
	}
	c, cancel := ctx()
	defer cancel()
	var ch Character
	var owner int32
	var level, statPoint, skillPoint, str, dex, rec, intel, luc, wis, hp, mp, sp, mapID, dir int64
	var exp, money, k1, k2, k3, k4 int32
	var guildID sql.NullInt64
	var storedMoney sql.NullInt64
	var hg, vg int16
	var cg, og, ig, del uint8
	var guildLevel, killLevel, deadLevel sql.NullInt64
	var event1, event2, renameCnt sql.NullInt64
	var leave [6]sql.NullInt32
	var posX, posY, posZ float32
	err = pool.QueryRowContext(c, `SELECT C.UserUID,C.CharID,C.CharName,C.Slot,C.Family,C.Grow,C.Hair,C.Face,C.Size,C.Job,C.Sex,C.Level,C.StatPoint,C.SkillPoint,C.Str,C.Dex,C.Rec,C.[Int],C.Luc,C.Wis,C.HP,C.MP,C.SP,C.Map,C.Dir,C.Exp,C.Money,C.PosX,C.PosY,C.PosZ,C.Hg,C.Vg,C.Cg,C.Og,C.Ig,C.Del,C.K1,C.K2,C.K3,C.K4,COALESCE(M.Money,0),COALESCE(G.GuildID,0),COALESCE(G.GuildLevel,0),C.KillLevel,C.DeadLevel,E.Event1,E.Event2,C.RenameCnt,CAST(strftime('%Y',C.LeaveDate) AS INTEGER),CAST(strftime('%m',C.LeaveDate) AS INTEGER),CAST(strftime('%d',C.LeaveDate) AS INTEGER),CAST(strftime('%H',C.LeaveDate) AS INTEGER),CAST(strftime('%M',C.LeaveDate) AS INTEGER),CAST(strftime('%S',C.LeaveDate) AS INTEGER) FROM PS_GameData_Chars C LEFT JOIN PS_GameData_UserStoredMoney M ON M.UserUID=C.UserUID AND M.ServerID=C.ServerID LEFT JOIN PS_GameData_GuildChars G ON G.CharID=C.CharID AND G.Del=0 LEFT JOIN PS_GameData_CharEvents E ON E.CharID=C.CharID WHERE C.CharID=@charid`, sql.Named("charid", int32(charID))).Scan(
		&owner, &ch.ID, &ch.Name, &ch.Slot, &ch.Race, &ch.Mode, &ch.Hair, &ch.Face, &ch.Height, &ch.Class,
		&ch.Gender, &level, &statPoint, &skillPoint, &str, &dex, &rec, &intel, &luc, &wis, &hp, &mp, &sp,
		&mapID, &dir, &exp, &money, &posX, &posY, &posZ, &hg, &vg, &cg, &og, &ig, &del, &k1, &k2, &k3, &k4,
		&storedMoney, &guildID, &guildLevel, &killLevel, &deadLevel, &event1, &event2, &renameCnt,
		&leave[0], &leave[1], &leave[2], &leave[3], &leave[4], &leave[5])
	if errors.Is(err, sql.ErrNoRows) || (err == nil && (uint32(owner) != userUID || del != 0)) {
		return Character{}, ErrNotFound
	}
	if err != nil {
		return Character{}, fmt.Errorf("usp_Read_Chars_Detail2_R: %w", err)
	}
	ch.AccountID = uint32(owner)
	ch.Name = strings.TrimRight(ch.Name, " ")
	ch.Level, ch.MapID, ch.Dir = max(u16(level), 1), u16(mapID), int16(dir)
	ch.StatPoint, ch.SkillPoint = u16(statPoint), u16(skillPoint)
	ch.Str, ch.Dex, ch.Rec, ch.Int, ch.Luc, ch.Wis = u16(str), u16(dex), u16(rec), u16(intel), u16(luc), u16(wis)
	ch.MaxHP, ch.MaxMP, ch.MaxSP = uint32(max(hp, 0)), uint32(max(mp, 0)), uint32(max(sp, 0))
	ch.CurrentHP, ch.CurrentMP, ch.CurrentSP = ch.MaxHP, ch.MaxMP, ch.MaxSP
	ch.Exp, ch.Money = uint32(max(exp, 0)), uint32(money) // Money is an unsigned bit pattern
	ch.PosX, ch.PosY, ch.PosZ = posX, posY, posZ
	ch.Hg, ch.Vg, ch.Cg, ch.Og, ch.Ig = hg, vg, cg, og, ig
	ch.Kills, ch.Deaths, ch.Victories, ch.Defeats = uint32(max(k1, 0)), uint32(max(k2, 0)), uint32(max(k3, 0)), uint32(max(k4, 0))
	ch.KillLevel, ch.DeadLevel = uint8(killLevel.Int64), uint8(deadLevel.Int64)
	ch.RenamePending = renameCnt.Int64 > 0
	return ch, nil
}

// CharacterIDInSlot runs usp_Read_CharSlots_R.
func CharacterIDInSlot(userUID uint32, slot uint8) (uint32, bool, error) {
	pool, err := ReadPool()
	if err != nil {
		return 0, false, err
	}
	c, cancel := ctx()
	defer cancel()
	var id int32
	err = pool.QueryRowContext(c, `SELECT CharID FROM PS_GameData_Chars WHERE ServerID=@serverid AND UserUID=@useruid AND Slot=@slot AND Del=0`, sql.Named("serverid", config.ServerID),
		sql.Named("useruid", int32(userUID)), sql.Named("slot", slot)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return uint32(id), err == nil, err
}

// NameTaken reports whether a live character, or one deleted within the
// restore window, already uses name.
func NameTaken(name string) (bool, error) {
	pool, err := ReadPool()
	if err != nil {
		return false, err
	}
	c, cancel := ctx()
	defer cancel()
	var n int
	err = pool.QueryRowContext(c, `SELECT COUNT(*) FROM PS_GameData_Chars WHERE CharName = ?1
		AND (Del = 0 OR (Del = 1 AND DeleteDate >= datetime('now','-2 hours')))`, name).Scan(&n)
	return n > 0, err
}

// CreateCharacter runs usp_Create_Char_R. The procedure itself rejects
// duplicate names (return -2) and fixes SkillPoint at 5.
// The parameters mirror the original CUser::ProcessCharacterCreate: level
// 1, no stat or skill points, Vg 200.
func CreateCharacter(account Account, nc NewCharacter) (uint32, error) {
	return CreateCharacterWithItems(account, nc, nil)
}

// CreateCharacterWithItems commits the character and initial inventory together.
// Slot/name validation is repeated under the writer lock, not trusted to callers.
func CreateCharacterWithItems(account Account, nc NewCharacter, items []Item) (uint32, error) {
	pool, err := Pool(GameData)
	if err != nil {
		return 0, err
	}
	c, cancel := ctx()
	defer cancel()
	var id int64
	err = withTx(c, pool, func(tx *sql.Tx) error {
		var count int
		if e := tx.QueryRowContext(c, `SELECT COUNT(*) FROM PS_GameData_Chars WHERE ServerID=? AND UserUID=? AND Slot=? AND Del=0`, config.ServerID, account.UserUID, nc.Slot).Scan(&count); e != nil {
			return e
		}
		if count != 0 {
			return fmt.Errorf("character slot is occupied")
		}
		if e := tx.QueryRowContext(c, `SELECT COUNT(*) FROM PS_GameData_Chars WHERE CharName=?1 AND (Del=0 OR (Del=1 AND Level>10 AND datetime(DeleteDate)>datetime('now','-7 days')))`, strings.TrimSpace(nc.Name)).Scan(&count); e != nil {
			return e
		}
		if count != 0 {
			return fmt.Errorf("character name is already in use")
		}
		res, e := tx.ExecContext(c, `INSERT INTO PS_GameData_Chars([ServerID],[UserID],[UserUID],[CharName],[Slot],[Family],[Grow],[Hair],[Face],[Size],[Job],[Sex],[Level],[StatPoint],[SkillPoint],[Str],[Dex],[Rec],[Int],[Luc],[Wis],[HP],[MP],[SP],[Map],[Dir],[Exp],[Money],[PosX],[PosY],[PosZ],[Hg],[Vg],[Cg],[Og],[Ig],RenameCnt,RemainTime) VALUES(@serverid,@userid,@useruid,@charname,@slot,@family,@grow,@hair,@face,@size,@job,@sex,@level,@statpoint,@skillpoint,@str,@dex,@rec,@int,@luc,@wis,@hp,@mp,@sp,@map,@dir,@exp,@money,@posx,@posy,@posz,@hg,@vg,@cg,@og,@ig,0,0)`,
			sql.Named("serverid", config.ServerID), sql.Named("userid", account.UserID),
			sql.Named("useruid", account.UserUID), sql.Named("charname", nc.Name), sql.Named("slot", nc.Slot),
			sql.Named("family", nc.Race), sql.Named("grow", nc.Mode), sql.Named("hair", nc.Hair),
			sql.Named("face", nc.Face), sql.Named("size", nc.Height), sql.Named("job", nc.Class),
			sql.Named("sex", nc.Gender), sql.Named("level", int16(1)), sql.Named("statpoint", int16(0)),
			sql.Named("skillpoint", int16(5)), sql.Named("str", int16(nc.Str)), sql.Named("dex", int16(nc.Dex)),
			sql.Named("rec", int16(nc.Rec)), sql.Named("int", int16(nc.Int)), sql.Named("luc", int16(nc.Luc)),
			sql.Named("wis", int16(nc.Wis)), sql.Named("hp", int16(nc.HP)), sql.Named("mp", int16(nc.MP)), sql.Named("sp", int16(nc.SP)),
			sql.Named("map", int16(nc.MapID)), sql.Named("dir", int16(0)), sql.Named("exp", int32(0)),
			sql.Named("money", int32(0)), sql.Named("posx", nc.PosX), sql.Named("posy", nc.PosY),
			sql.Named("posz", nc.PosZ), sql.Named("hg", int16(0)), sql.Named("vg", int16(200)),
			sql.Named("cg", uint8(0)), sql.Named("og", uint8(0)), sql.Named("ig", uint8(0)))
		if e != nil {
			return e
		}
		id, e = res.LastInsertId()
		if e != nil {
			return e
		}
		for _, it := range items {
			if e := insertCharacterItem(c, tx, uint32(id), &it); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return uint32(id), nil
}

// DeleteCharacter soft-deletes a character and removes its guild membership in
// one transaction. A guild master must resolve the guild's leadership first.
func DeleteCharacter(userUID, charID uint32) (bool, error) {
	pool, err := Pool(GameData)
	if err != nil {
		return false, err
	}
	c, cancel := ctx()
	defer cancel()
	deleted := false
	err = withTx(c, pool, func(tx *sql.Tx) error {
		var guildMaster bool
		err := tx.QueryRowContext(c, `SELECT EXISTS (
			SELECT 1 FROM PS_GameData_Guilds WHERE MasterCharID=?1 AND Del=0)
			FROM PS_GameData_Chars WHERE CharID=?1 AND UserUID=?2 AND ServerID=?3 AND Del=0`,
			charID, userUID, config.ServerID).Scan(&guildMaster)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if guildMaster {
			return errors.New("character leads an active guild; transfer leadership or disband the guild first")
		}
		if _, err := tx.ExecContext(c, `UPDATE PS_GameData_Guilds SET TotalCount=MAX(0,TotalCount-1)
			WHERE GuildID IN (SELECT GuildID FROM PS_GameData_GuildChars WHERE CharID=? AND Del=0)`, charID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(c, `UPDATE PS_UserData_CreatedGuilds AS mirror SET TotalCount=(
			SELECT TotalCount FROM PS_GameData_Guilds WHERE GuildID=mirror.GuildID), UpdateDate=CURRENT_TIMESTAMP
			WHERE ServerID=? AND GuildID IN (SELECT GuildID FROM PS_GameData_GuildChars WHERE CharID=? AND Del=0)
				AND EXISTS (SELECT 1 FROM PS_GameData_Guilds WHERE GuildID=mirror.GuildID)`, config.ServerID, charID); err != nil {
			return err
		}
		for _, statement := range []string{
			`UPDATE PS_GameData_GuildChars SET Del=1,LeaveDate=CURRENT_TIMESTAMP WHERE CharID=? AND Del=0`,
			`UPDATE PS_GameData_Chars SET Del=1,DeleteDate=CURRENT_TIMESTAMP WHERE CharID=?`,
		} {
			if _, err := tx.ExecContext(c, statement, charID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(c, `INSERT INTO PS_GameData__DeletedChars(ServerID,CharID,DeleteDate) VALUES(?,?,CURRENT_TIMESTAMP)`,
			config.ServerID, charID); err != nil {
			return err
		}
		deleted = true
		return nil
	})
	return deleted && err == nil, err
}

// RestoreCharacter undoes a deletion within DeleteWindow. The EP4.5
// procedures have no restore; this is the inverse of usp_Delete_Char_E's
// Chars update.
func RestoreCharacter(userUID, charID uint32) (bool, error) {
	pool, err := Pool(GameData)
	if err != nil {
		return false, err
	}
	c, cancel := ctx()
	defer cancel()
	res, err := pool.ExecContext(c, `UPDATE PS_GameData_Chars SET Del = 0, DeleteDate = NULL
		WHERE CharID = ?1 AND UserUID = ?2 AND ServerID = ?3 AND Del = 1 AND DeleteDate >= datetime('now','-2 hours')
        AND NOT EXISTS (SELECT 1 FROM PS_GameData_Chars AS occupied
        WHERE occupied.UserUID=PS_GameData_Chars.UserUID
        AND occupied.ServerID=PS_GameData_Chars.ServerID
        AND occupied.Slot=PS_GameData_Chars.Slot AND occupied.Del=0)`,
		charID, userUID, config.ServerID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// EnableRename runs Usp_Save_Char_EnableRename_E (name-change stone).
func EnableRename(charID uint32) error {
	return execSQL(GameData, `UPDATE PS_GameData_Chars SET RenameCnt = RenameCnt+1  WHERE CharID = @charid`, sql.Named("charid", int32(charID)))
}

// RenameCharacter commits a pending rename with usp_Update_Char_Rename and
// consumes one RenameCnt.
func RenameCharacter(userUID, charID uint32, oldName, newName string) (bool, error) {
	pool, err := Pool(GameData)
	if err != nil {
		return false, err
	}
	c, cancel := ctx()
	defer cancel()
	var n int
	if err := pool.QueryRowContext(c, `SELECT COUNT(*) FROM PS_GameData_Chars WHERE CharID=?1 AND UserUID=?2
		AND Del=0 AND RenameCnt >= 1`, int32(charID), int32(userUID)).Scan(&n); err != nil || n == 0 {
		return false, err
	}
	renamed := false
	err = withTx(c, pool, func(tx *sql.Tx) error {
		var count int
		if e := tx.QueryRowContext(c, `SELECT COUNT(*) FROM PS_GameData_Chars WHERE CharName=?1 AND Del=0`, newName).Scan(&count); e != nil {
			return e
		}
		if count > 0 {
			return nil
		}
		res, e := tx.ExecContext(c, `UPDATE PS_GameData_Chars SET OldCharName=?3,CharName=?4,RenameCnt=RenameCnt-1 WHERE CharID=?1 AND UserUID=?2 AND Del=0 AND RenameCnt>0`, int32(charID), int32(userUID), oldName, newName)
		if e != nil {
			return e
		}
		n, e := res.RowsAffected()
		renamed = n == 1
		if e != nil {
			return e
		}
		if renamed {
			for _, q := range []string{`UPDATE PS_GameData_Guilds SET MasterName=?2 WHERE MasterCharID=?1`, `UPDATE PS_GameData_FriendChars SET FriendName=?2 WHERE FriendID=?1`, `UPDATE PS_GameData_BanChars SET BanName=?2 WHERE BanID=?1`, `INSERT INTO PS_GameData_CharRenameLog(ServerID,CharID,CharName) VALUES(1,?1,?2)`} {
				if _, e = tx.ExecContext(c, q, int32(charID), newName); e != nil {
					return e
				}
			}
		}
		return e
	})
	return renamed, err
}

// MapForSave maps a world zone to the map stored in Chars.Map (the world
// replaces it so instanced zones save their real map).
var MapForSave = func(zone uint16) uint16 { return zone }

// SaveCharacter persists level, points, stats, HP/MP/SP, position, exp,
// money and kill counters with usp_Save_Char_Info_E.
func SaveCharacter(ch *Character) error {
	pool, err := Pool(GameData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	return groupWrite(c, pool, func(c context.Context, tx *sql.Tx) error {
		_, err := writeCharacter(c, tx, ch)
		return err
	})
}

func writeCharacter(c context.Context, executor sqlExecutor, ch *Character) (sql.Result, error) {
	clamp16 := func(v uint32) int16 { return int16(min(v, 32767)) }
	return executor.ExecContext(c, `UPDATE PS_GameData_Chars
SET [Level] = @level, StatPoint = @statpoint, SkillPoint = @skillpoint,
[Str] = @str, dex = @dex, Rec = @rec, [int] = @int, Wis = @wis, Luc = @luc,
HP = @hp, Mp = @mp, Sp = @sp,
Map = @map, dir = @dir, [exp] = @exp, [money] = @money,
PosX = @posx, PosY = @posy, PosZ = @posz, hg = @hg, vg = @vg, cg = @cg, og = @og, ig = @ig,
KillLevel=@killlevel, DeadLevel=@deadlevel,LeaveDate=CURRENT_TIMESTAMP, K1=@k1,K2=@k2,K3=@k3,K4=@k4
WHERE CharID = @charid AND Del=0`,
		sql.Named("charid", int32(ch.ID)), sql.Named("level", int32(ch.Level)),
		sql.Named("statpoint", int16(ch.StatPoint)), sql.Named("skillpoint", int16(ch.SkillPoint)),
		sql.Named("str", int16(ch.Str)), sql.Named("dex", int16(ch.Dex)), sql.Named("rec", int16(ch.Rec)),
		sql.Named("int", int16(ch.Int)), sql.Named("wis", int16(ch.Wis)), sql.Named("luc", int16(ch.Luc)),
		sql.Named("hp", clamp16(ch.CurrentHP)), sql.Named("mp", clamp16(ch.CurrentMP)), sql.Named("sp", clamp16(ch.CurrentSP)),
		sql.Named("map", int16(MapForSave(ch.MapID))), sql.Named("dir", ch.Dir), sql.Named("exp", int32(ch.Exp)),
		sql.Named("money", int32(ch.Money)), sql.Named("posx", ch.PosX), sql.Named("posy", ch.PosY),
		sql.Named("posz", ch.PosZ), sql.Named("hg", int32(ch.Hg)), sql.Named("vg", int32(ch.Vg)),
		sql.Named("cg", int32(ch.Cg)), sql.Named("og", int32(ch.Og)), sql.Named("ig", int32(ch.Ig)),
		sql.Named("k1", int32(ch.Kills)), sql.Named("k2", int32(ch.Deaths)), sql.Named("k3", int32(ch.Victories)),
		sql.Named("k4", int32(ch.Defeats)), sql.Named("killlevel", ch.KillLevel), sql.Named("deadlevel", ch.DeadLevel))
}

// SaveAppearance runs usp_Save_Char_BodyChg_E.
func SaveAppearance(charID uint32, hair, face, size, sex uint8) error {
	return execSQL(GameData, `UPDATE PS_GameData_Chars
SET Hair=@hair, Face=@face, [Size]=@size,  Sex=@sex
WHERE CharID = @charid`, sql.Named("charid", int32(charID)),
		sql.Named("hair", hair), sql.Named("face", face), sql.Named("size", size), sql.Named("sex", sex))
}

// CharacterByName finds a live character's account and id by name.
func CharacterByName(name string) (charID, userUID uint32, found bool, err error) {
	pool, err := ReadPool()
	if err != nil {
		return 0, 0, false, err
	}
	c, cancel := ctx()
	defer cancel()
	var id, uid int32
	err = pool.QueryRowContext(c, `SELECT CharID, UserUID FROM PS_GameData_Chars WHERE CharName = ?1 AND Del = 0 LIMIT 1`, name).Scan(&id, &uid)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	return uint32(id), uint32(uid), err == nil, err
}

// CharacterByID loads any live character regardless of account (used to
// credit an offline caster).
func CharacterByID(charID uint32) (Character, error) {
	pool, err := ReadPool()
	if err != nil {
		return Character{}, err
	}
	c, cancel := ctx()
	defer cancel()
	var uid int32
	err = pool.QueryRowContext(c, `SELECT UserUID FROM PS_GameData_Chars WHERE CharID = ?1 AND Del = 0`, int32(charID)).Scan(&uid)
	if errors.Is(err, sql.ErrNoRows) {
		return Character{}, ErrNotFound
	}
	if err != nil {
		return Character{}, err
	}
	return readCharacterDetail(uint32(uid), charID)
}

func execSQL(database, name string, args ...any) error {
	pool, err := Pool(database)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	if _, err := pool.ExecContext(c, name, args...); err != nil {
		reportStorageFailure(err)
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
