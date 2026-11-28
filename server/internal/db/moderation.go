package db

import (
	"database/sql"
	"net/netip"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

// SetOperatorAccountBan preserves the original account status. Blocking an
// account invalidates tickets from every Login connection for that account.
func SetOperatorAccountBan(uid uint32, ban bool, reason string) error {
	pool, err := Pool(UserData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	return withTx(c, pool, func(tx *sql.Tx) error {
		if !ban {
			_, err := tx.ExecContext(c, `DELETE FROM GX_OperatorAccountBans WHERE UserUID=?`, uid)
			return err
		}
		if _, err := tx.ExecContext(c, `INSERT INTO GX_OperatorAccountBans(UserUID,Reason) VALUES(?,?)
			ON CONFLICT(UserUID) DO UPDATE SET Reason=excluded.Reason`, uid, reason); err != nil {
			return err
		}
		return revokeCredentialTx(c, tx, uid)
	})
}

func SetOperatorIPBan(address netip.Addr, ban bool, reason string) error {
	pool, err := Pool(UserData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	if !ban {
		_, err := pool.ExecContext(c, `DELETE FROM GX_OperatorIPBans WHERE Address=?`, address.String())
		return err
	}
	_, err = pool.ExecContext(c, `INSERT INTO GX_OperatorIPBans(Address,Reason) VALUES(?,?)
		ON CONFLICT(Address) DO UPDATE SET Reason=excluded.Reason`, address.String(), reason)
	return err
}

func OperatorIPBanned(address netip.Addr) (bool, error) {
	pool, err := ReadPool()
	if err != nil {
		return false, err
	}
	c, cancel := ctx()
	defer cancel()
	var banned bool
	err = pool.QueryRowContext(c, `SELECT EXISTS(SELECT 1 FROM GX_OperatorIPBans WHERE Address=?)`, address.String()).Scan(&banned)
	return banned, err
}

func OperatorIPBans(query string) ([]operatorwire.IPBan, bool, error) {
	pool, err := ReadPool()
	if err != nil {
		return nil, false, err
	}
	c, cancel := ctx()
	defer cancel()
	rows, err := pool.QueryContext(c, `SELECT Address,Reason,CreatedAt FROM GX_OperatorIPBans
		WHERE instr(Address,lower(?))>0 ORDER BY CreatedAt DESC,Address LIMIT ?`, query, operatorAccountLimit+1)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	var bans []operatorwire.IPBan
	for rows.Next() {
		var ban operatorwire.IPBan
		var created int64
		if err := rows.Scan(&ban.Address, &ban.Reason, &created); err != nil {
			return nil, false, err
		}
		ban.CreatedAt = time.Unix(created, 0).UTC()
		bans = append(bans, ban)
	}
	return bans[:min(len(bans), operatorAccountLimit)], len(bans) > operatorAccountLimit, rows.Err()
}
