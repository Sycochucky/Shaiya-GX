package db

import (
	"database/sql"
	"errors"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/password"
)

// LoginResult is what PS_UserData.dbo.usp_Try_GameLogin_Taiwan selects.
type LoginResult struct {
	Status            int32
	UserUID           int32
	CredentialVersion int64
}

// TryGameLogin translates the original procedure into one SQLite transaction,
// including account status, the login log and UserLoginStatus.
func TryGameLogin(userID, passwordValue string, sessionID int64, userIP string) (LoginResult, error) {
	reader, err := ReadPool()
	if err != nil {
		return LoginResult{}, err
	}
	c, cancel := ctx()
	defer cancel()
	// Password checks and failed attempts never take SQLite's writer lock.
	var stored string
	var uid, leave int32
	err = reader.QueryRowContext(c, `SELECT Pw,UserUID,Leave FROM PS_UserData_Users_Master WHERE UserID=?`, userID).Scan(&stored, &uid, &leave)
	if errors.Is(err, sql.ErrNoRows) {
		return LoginResult{Status: -3}, password.Dummy(c, passwordValue)
	}
	if err != nil {
		return LoginResult{}, err
	}
	valid, legacy, err := password.VerifyContext(c, stored, passwordValue)
	if err != nil {
		return LoginResult{}, err
	}
	if leave != 0 {
		return LoginResult{Status: -3, UserUID: uid}, nil
	}
	if !valid {
		return LoginResult{Status: -1, UserUID: uid}, nil
	}
	p, err := Pool(UserData)
	if err != nil {
		return LoginResult{}, err
	}
	var upgraded string
	if legacy {
		upgraded, err = password.HashContext(c, passwordValue)
		if err != nil {
			return LoginResult{}, err
		}
	}
	result := LoginResult{Status: -3}
	err = withTx(c, p, func(tx *sql.Tx) error {
		var uid int32
		var status int32
		var leave int
		var matches bool
		var version int64
		e := tx.QueryRowContext(c, `SELECT UserUID,Status,Leave,Pw=?2,COALESCE((SELECT Version FROM GX_AccountSecurity WHERE UserUID=PS_UserData_Users_Master.UserUID),1) FROM PS_UserData_Users_Master WHERE UserID=?1`, userID, stored).Scan(&uid, &status, &leave, &matches, &version)
		if errors.Is(e, sql.ErrNoRows) {
			return nil
		}
		if e != nil {
			return e
		}
		result = LoginResult{Status: status, UserUID: uid, CredentialVersion: version}
		if leave != 0 {
			result.Status = -3
			return nil
		}
		if !matches {
			result.Status = -1
			return nil
		}
		var banned bool
		e = tx.QueryRowContext(c, `SELECT EXISTS(SELECT 1 FROM GX_OperatorAccountBans WHERE UserUID=?1)
			OR EXISTS(SELECT 1 FROM GX_OperatorIPBans WHERE Address=?2)`, uid, userIP).Scan(&banned)
		if e != nil {
			return e
		}
		if banned {
			result.Status = 2
			return nil
		}
		if status >= 2 && status <= 6 {
			var expired bool
			e = tx.QueryRowContext(c, `SELECT datetime(BlockEndDate)<=CURRENT_TIMESTAMP FROM PS_UserData_Users_Block WHERE UserUID=?1`, uid).Scan(&expired)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return e
			}
			if expired {
				result.Status = 0
				if _, e = tx.ExecContext(c, `UPDATE PS_UserData_Users_Master SET Status=0 WHERE UserUID=?1`, uid); e != nil {
					return e
				}
			}
		}
		if !CanPlay(result.Status) {
			return nil
		}
		var active int
		if e = tx.QueryRowContext(c, `SELECT COUNT(*) FROM GX_ActiveSessions WHERE UserUID=?`, uid).Scan(&active); e != nil {
			return e
		}
		if active != 0 {
			result.Status = 1
			return nil
		}
		if upgraded != "" {
			if _, e = tx.ExecContext(c, `UPDATE PS_UserData_Users_Master SET Pw=? WHERE UserUID=?`, upgraded, uid); e != nil {
				return e
			}
		}
		_, e = tx.ExecContext(c, `INSERT INTO PS_UserData_UserLoginStatus(UserUID,UserID,LoginStatus,LoginTime,LogoutTime,LoginSession,LastPlayTime)
   VALUES(?1,?2,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?3,0)
   ON CONFLICT(UserUID) DO UPDATE SET LoginStatus=1,LoginTime=CURRENT_TIMESTAMP,LogoutTime=CURRENT_TIMESTAMP,LoginSession=excluded.LoginSession,LastPlayTime=0`, uid, userID, sessionID)
		if e != nil {
			return e
		}
		_, e = tx.ExecContext(c, `INSERT INTO Rapture_LoginLog(UserUID,SessionID,UserIP,LogType) VALUES(?1,?2,?3,0)`, uid, sessionID, userIP)
		return e
	})
	return result, err
}

// Logout and the original two-minute online-point credit commit together.
func TryGameLogout(userUID int32, sessionID int64) error {
	p, err := Pool(UserData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	return withTx(c, p, func(tx *sql.Tx) error {
		var loginSession int64
		var minutes int
		var loginTime string
		var online int
		e := tx.QueryRowContext(c, `SELECT LoginSession,MAX(0,CAST((julianday('now')-julianday(LoginTime))*1440 AS INTEGER)),LoginTime,LoginStatus FROM PS_UserData_UserLoginStatus WHERE UserUID=?1`, userUID).Scan(&loginSession, &minutes, &loginTime, &online)
		if errors.Is(e, sql.ErrNoRows) {
			return nil
		}
		if e != nil {
			return e
		}
		// A delayed disconnect must never alter a newer login or credit twice.
		if sessionID != loginSession || online == 0 {
			return nil
		}
		if _, e = tx.ExecContext(c, `INSERT INTO PS_UserData_UserOnlineTimeLog(UserUID,OldPoint,AddPoint,LoginTime,LogoutTime,SessionID)
    SELECT UserUID,COALESCE(Point,0),?2,?3,CURRENT_TIMESTAMP,?4 FROM PS_UserData_Users_Master WHERE UserUID=?1`, userUID, minutes, loginTime, sessionID); e != nil {
			return e
		}
		if _, e = tx.ExecContext(c, `UPDATE PS_UserData_Users_Master SET Point=COALESCE(Point,0)+?2 WHERE UserUID=?1`, userUID, minutes/2); e != nil {
			return e
		}
		if _, e = tx.ExecContext(c, `UPDATE PS_UserData_UserLoginStatus SET LoginStatus=0,LogoutTime=CURRENT_TIMESTAMP,LogoutSession=?2,LastPlayTime=?3 WHERE UserUID=?1`, userUID, sessionID, minutes); e != nil {
			return e
		}
		_, e = tx.ExecContext(c, `INSERT INTO Rapture_LoginLog(UserUID,SessionID,LogType) VALUES(?1,?2,1)`, userUID, sessionID)
		return e
	})
}

// LoginDecision maps the procedure's Status to the client's login result
// byte and permission (auth status), as recovered from the original
// ps_login (CUser login handling; see Eternity src/auth/login_protocol.cpp).
// Permission 11 is a regular account; 1..5 are the admin tiers.
func LoginDecision(login LoginResult) (result uint8, permission uint8) {
	permission = 11
	if login.UserUID == 0 {
		return 1, permission
	}
	switch login.Status {
	case -3:
		result = 1
	case -2:
		result = 2
	case -1:
		result = 3
	case 1:
		result = 9
	case 2:
		result = 4
	case 6:
		result = 5
	case 0x11:
		result = 10
	case 0x21:
		result = 11
	case 0x31:
		result = 12
	case 0x41:
		result = 13
	case 0x51:
		result = 14
	case 0x80:
		result = 6
	case 0x100:
		result = 7
	case 0x200:
		result = 8
	default:
		if !CanPlay(login.Status) {
			return 1, permission
		}
		switch {
		case login.Status&2 != 0:
			result = 10
		case login.Status&1 != 0:
			result = 11
		case (uint32(login.Status)>>7)&7 != 0:
			result = 6
		default:
			permission = uint8((login.Status >> 4) & 7)
			if permission == 0 {
				permission = 11
			}
		}
	}
	return result, permission
}

// Account is the account state the world server needs.
// CanPlay rejects unrecognized authorization states rather than granting
// permissions from arbitrary legacy status bits.
func CanPlay(status int32) bool {
	return status == 0 || status == 16 || status == 32 || status == 48 || status == 64 || status == 80
}

type Account struct {
	UserUID           int32
	CredentialVersion int64
	UserID            string
	Status            int16
	Admin             bool
	AdminLevel        uint8
	Point             int32
	Country           uint8 // 0 light, 1 dark, 2 none chosen yet
}

var ErrNotFound = errors.New("not found")

// AccountByUID reads Users_Master plus the realm's UserMaxGrow row
// (PS_GameData.dbo.usp_Read_User_MaxGrow_R).
func AccountByUID(userUID int32) (Account, error) {
	pool, err := ReadPool()
	if err != nil {
		return Account{}, err
	}
	c, cancel := ctx()
	defer cancel()
	var a Account
	var country int16
	err = pool.QueryRowContext(c, `SELECT U.UserUID,U.UserID,CASE WHEN B.UserUID IS NULL THEN U.Status ELSE 2 END,U.Admin,U.AdminLevel,U.Point,
		COALESCE(S.Version,1),COALESCE(G.Country,2)
		FROM PS_UserData_Users_Master U LEFT JOIN GX_AccountSecurity S ON S.UserUID=U.UserUID
		LEFT JOIN GX_OperatorAccountBans B ON B.UserUID=U.UserUID
		LEFT JOIN PS_GameData_UserMaxGrow G ON G.ServerID=?2 AND G.UserUID=U.UserUID
		WHERE U.UserUID=?1 AND U.Leave=0`, userUID, config.ServerID).
		Scan(&a.UserUID, &a.UserID, &a.Status, &a.Admin, &a.AdminLevel, &a.Point, &a.CredentialVersion, &country)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, err
	}
	a.Country = uint8(country)
	return a, nil
}
