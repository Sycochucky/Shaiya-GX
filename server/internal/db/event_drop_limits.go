package db

import "time"

// EnsureEventDropLimits creates storage in the mutable runtime database.
func EnsureEventDropLimits() error {
	pool, err := Pool(GameData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	_, err = pool.ExecContext(c, `CREATE TABLE IF NOT EXISTS GX_EventDropLimits (
		Day TEXT NOT NULL, RuleID INTEGER NOT NULL, UserUID INTEGER NOT NULL,
		DropCount INTEGER NOT NULL, PRIMARY KEY (RuleID, UserUID)
	) WITHOUT ROWID`)
	return err
}

// ClaimEventDrop atomically reserves one of the account's daily awards.
func ClaimEventDrop(ruleID, accountID uint32, limit uint16, now time.Time) (bool, error) {
	pool, err := Pool(GameData)
	if err != nil {
		return false, err
	}
	c, cancel := ctx()
	defer cancel()
	result, err := pool.ExecContext(c, `INSERT INTO GX_EventDropLimits (Day, RuleID, UserUID, DropCount)
		VALUES (?, ?, ?, 1) ON CONFLICT(RuleID, UserUID) DO UPDATE
		SET Day = excluded.Day,
		    DropCount = CASE WHEN Day = excluded.Day THEN DropCount + 1 ELSE 1 END
		WHERE Day <> excluded.Day OR DropCount < ?`, now.UTC().Format(time.DateOnly), ruleID, accountID, limit)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count == 1, err
}
