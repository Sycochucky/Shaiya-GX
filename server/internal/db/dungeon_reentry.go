package db

import (
	"database/sql"
	"errors"
	"time"
)

func EnsureDungeonReentry() error {
	pool, err := Pool(GameData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	_, err = pool.ExecContext(c, `CREATE TABLE IF NOT EXISTS GX_DungeonReentry (
		CharID INTEGER NOT NULL, MapID INTEGER NOT NULL, Until INTEGER NOT NULL,
		PRIMARY KEY(CharID, MapID)
	) WITHOUT ROWID`)
	return err
}

func SetDungeonReentry(character uint32, mapID uint16, until time.Time) error {
	pool, err := Pool(GameData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	_, err = pool.ExecContext(c, `INSERT INTO GX_DungeonReentry(CharID, MapID, Until) VALUES (?,?,?)
		ON CONFLICT(CharID, MapID) DO UPDATE SET Until=MAX(Until,excluded.Until)`, character, mapID, until.Unix())
	return err
}

func DungeonReentry(character uint32, mapID uint16) (time.Time, error) {
	pool, err := Pool(GameData)
	if err != nil {
		return time.Time{}, err
	}
	c, cancel := ctx()
	defer cancel()
	var expires int64
	err = pool.QueryRowContext(c, "SELECT Until FROM GX_DungeonReentry WHERE CharID=? AND MapID=?", character, mapID).Scan(&expires)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	return time.Unix(expires, 0), err
}
