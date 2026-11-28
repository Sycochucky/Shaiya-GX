package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/netip"
	"strings"

	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func runOperator(input io.Reader, output io.Writer) error {
	data := json.NewDecoder(io.LimitReader(input, 16<<10))
	data.DisallowUnknownFields()
	var req operatorwire.Request
	if err := data.Decode(&req); err != nil {
		return errors.New("invalid operator request")
	}
	var extra any
	if data.Decode(&extra) != io.EOF {
		return errors.New("invalid operator request")
	}
	switch req.Action {
	case "delete", "restore", "save-stats", "account", "password", "ban-account", "unban-account", "ban-ip", "unban-ip":
		release, err := db.AcquireWorldAdministration(context.Background())
		if err != nil {
			return fmt.Errorf("runtime is in use by another World or installation: %w", err)
		}
		defer release()
	}
	defer db.Close()
	res, err := Operate(req)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(res)
}

// Operate executes local account tasks and character tasks when their owner is serialized by the caller.
func Operate(req operatorwire.Request) (res operatorwire.Response, err error) {
	switch req.Action {
	case "accounts":
		if len(req.Query) > 128 {
			return res, errors.New("account search is too long")
		}
		res.Accounts, res.More, err = db.OperatorAccounts(strings.TrimSpace(req.Query))
	case "ip-bans":
		if len(req.Query) > 64 {
			return res, errors.New("IP search is too long")
		}
		res.IPBans, res.More, err = db.OperatorIPBans(strings.TrimSpace(req.Query))
	case "ban-ip", "unban-ip":
		_, err = ChangeIPBan(req)
	case "create":
		_, err = db.CreateAccount(req.Name, req.Password, req.GM)
	default:
		if req.AccountID == 0 || req.AccountID > math.MaxInt32 || req.CharacterID > math.MaxInt32 {
			return res, errors.New("invalid account or character identity")
		}
		account, err := db.OperatorAccount(req.AccountID)
		if err != nil {
			return res, err
		}
		return operateAccount(req, account)
	}
	return res, err
}

func operateAccount(req operatorwire.Request, account operatorwire.Account) (res operatorwire.Response, err error) {
	switch req.Action {
	case "details":
		res.Account = &account
		rows, err := db.CharactersForAccount(req.AccountID)
		if err != nil {
			return res, err
		}
		for _, char := range rows {
			res.Characters = append(res.Characters, operatorwire.Character{ID: char.ID, Name: char.Name, Level: char.Level, Deleted: char.PendingDelete})
		}
	case "ban-account", "unban-account":
		reason, err := moderationReason(req.BanReason)
		if err != nil {
			return res, err
		}
		return res, db.SetOperatorAccountBan(account.ID, req.Action == "ban-account", reason)
	case "password":
		err = db.SetAccountPassword(account.Name, req.Password)
	case "account":
		if req.ExpectedAccount == nil || req.ExpectedAccount.ID != req.AccountID {
			return res, errors.New("missing matching account snapshot")
		}
		err = db.SetOperatorAccount(*req.ExpectedAccount, req.GM, req.Points)
	case "stats":
		return db.OperatorStats(req.AccountID, req.CharacterID)
	case "save-stats":
		if req.Stats == nil || req.ExpectedStats == nil {
			return res, errors.New("missing character values")
		}
		err = db.SetOperatorStats(req.AccountID, req.CharacterID, *req.ExpectedStats, *req.Stats)
	case "delete", "restore":
		return res, changeCharacter(req)
	default:
		return res, errors.New("unknown operator action")
	}
	return res, err
}

// ChangeIPBan returns the exact peer address used by the database and World.
func ChangeIPBan(req operatorwire.Request) (netip.Addr, error) {
	address, err := netip.ParseAddr(strings.TrimSpace(req.IPAddress))
	if err != nil || address.Zone() != "" || address.Unmap().IsUnspecified() || address.IsMulticast() {
		return netip.Addr{}, errors.New("enter one valid IPv4 or IPv6 address")
	}
	reason, err := moderationReason(req.BanReason)
	if err != nil {
		return netip.Addr{}, err
	}
	address = address.Unmap()
	return address, db.SetOperatorIPBan(address, req.Action == "ban-ip", reason)
}

func moderationReason(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) > 240 {
		return "", errors.New("ban reason must be 240 UTF-8 bytes or less")
	}
	return value, nil
}

func changeCharacter(req operatorwire.Request) error {
	rows, err := db.CharactersForAccount(req.AccountID)
	if err != nil {
		return err
	}
	for _, char := range rows {
		if char.ID != req.CharacterID || char.PendingDelete != (req.Action == "restore") {
			continue
		}
		var ok bool
		if char.PendingDelete {
			ok, err = db.RestoreCharacter(req.AccountID, char.ID)
		} else {
			if req.ConfirmName != char.Name {
				return errors.New("type the exact character name to confirm deletion")
			}
			ok, err = db.DeleteCharacter(req.AccountID, char.ID)
		}
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("character could not be changed; check the restore window or slot")
		}
		return nil
	}
	return errors.New("character is unavailable for this account")
}
