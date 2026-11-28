package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Spelunkern/Shaiya-GX/internal/authwire"
	"github.com/Spelunkern/Shaiya-GX/server/internal/password"
	"strings"
)

// CreateAccount never stores plaintext and never replaces an existing account.
func CreateAccount(name, secret string, admin bool) (int32, error) {
	if err := authwire.Validate(name, secret); err != nil {
		return 0, err
	}
	hash, err := password.Hash(secret)
	if err != nil {
		return 0, err
	}
	pool, err := Pool(UserData)
	if err != nil {
		return 0, err
	}
	c, cancel := ctx()
	defer cancel()
	var uid int32
	err = withTx(c, pool, func(tx *sql.Tx) error {
		var exists int
		if err := tx.QueryRowContext(c, `SELECT COUNT(*) FROM PS_UserData_Users_Master WHERE UserID=?`, name).Scan(&exists); err != nil {
			return err
		}
		if exists != 0 {
			return fmt.Errorf("account already exists")
		}
		var next int64
		if err := tx.QueryRowContext(c, `SELECT COALESCE(MAX(UserUID),0)+1 FROM PS_UserData_Users_Master`).Scan(&next); err != nil {
			return err
		}
		if next > 2147483647 {
			return fmt.Errorf("account identity space exhausted")
		}
		uid = int32(next)
		_, err := tx.ExecContext(c, `INSERT INTO PS_UserData_Users_Master(UserUID,UserID,Pw,Admin,AdminLevel,UseQueue,Status,Leave,UserType,Point) VALUES(?,?,?,?,?,0,0,0,'N',0)`, uid, name, hash, admin, 0)
		return err
	})
	return uid, err
}

// SetAccountPassword revokes pending tickets and active sessions through a
// monotonically increasing credential version. World checks it every 30 seconds.
func SetAccountPassword(name, secret string) error {
	if err := authwire.Validate(name, secret); err != nil {
		return err
	}
	hash, err := password.Hash(secret)
	if err != nil {
		return err
	}
	pool, err := Pool(UserData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	return withTx(c, pool, func(tx *sql.Tx) error {
		var uid int32
		if err := tx.QueryRowContext(c, `SELECT UserUID FROM PS_UserData_Users_Master WHERE UserID=?`, name).Scan(&uid); err != nil {
			return err
		}
		if _, err := tx.ExecContext(c, `UPDATE PS_UserData_Users_Master SET Pw=? WHERE UserUID=?`, hash, uid); err != nil {
			return err
		}
		return revokeCredentialTx(c, tx, uint32(uid))
	})
}

func revokeCredentialTx(c context.Context, tx *sql.Tx, uid uint32) error {
	if _, err := tx.ExecContext(c, `INSERT INTO GX_AccountSecurity(UserUID,Version) VALUES(?,2) ON CONFLICT(UserUID) DO UPDATE SET Version=Version+1`, uid); err != nil {
		return err
	}
	_, err := tx.ExecContext(c, `DELETE FROM GX_SessionHandoff WHERE UserUID=?`, uid)
	return err
}

// UpgradeLegacyPasswords hashes passwords still stored in the legacy format.
// Each update compares the original row so a concurrent password reset wins.
// Invalid legacy credentials are counted and can be reset with the account CLI.
func UpgradeLegacyPasswords() (upgraded, needsReset int, err error) {
	reader, err := ReadPool()
	if err != nil {
		return 0, 0, err
	}
	writer, err := Pool(UserData)
	if err != nil {
		return 0, 0, err
	}
	var cursor int64
	for {
		batch, err := legacyPasswords(reader, cursor)
		if err != nil || len(batch) == 0 {
			return upgraded, needsReset, err
		}
		for _, record := range batch {
			cursor = record.uid
			if len(record.secret) == 0 || len(record.secret) > 128 || strings.ContainsRune(record.secret, 0) {
				needsReset++
				continue
			}
			hash, err := password.Hash(record.secret)
			if err != nil {
				return upgraded, needsReset, err
			}
			c, cancel := ctx()
			result, err := writer.ExecContext(c, `UPDATE PS_UserData_Users_Master SET Pw=? WHERE UserUID=? AND Pw=?`, hash, record.uid, record.secret)
			cancel()
			if err != nil {
				return upgraded, needsReset, err
			}
			n, err := result.RowsAffected()
			if err != nil {
				return upgraded, needsReset, err
			}
			upgraded += int(n)
		}
	}
}

type legacyPassword struct {
	uid    int64
	secret string
}

func legacyPasswords(reader *sql.DB, after int64) ([]legacyPassword, error) {
	c, cancel := ctx()
	defer cancel()
	rows, err := reader.QueryContext(c, `SELECT UserUID,COALESCE(Pw,'') FROM PS_UserData_Users_Master
		WHERE UserUID>? AND (Pw IS NULL OR substr(Pw,1,1)<>'$') ORDER BY UserUID LIMIT 64`, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []legacyPassword
	for rows.Next() {
		var record legacyPassword
		if err := rows.Scan(&record.uid, &record.secret); err != nil {
			return nil, err
		}
		entries = append(entries, record)
	}
	return entries, rows.Err()
}
