package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/controlplane"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

// Operator uses the authoritative World for live writes and snapshots. The
// private console handles offline edits and account lookup.
func (m *Manager) Operator(ctx context.Context, req operatorwire.Request) (res operatorwire.Response, err error) {
	release, err := m.lock(ctx)
	if err != nil {
		return res, err
	}
	defer release()
	switch req.Action {
	case "create", "password", "account", "save-stats", "delete", "restore", "ban-account", "unban-account", "ban-ip", "unban-ip":
		log, openErr := logx.NewLocal(m.root, string(operatorLog))
		if openErr != nil {
			return res, fmt.Errorf("open staff audit log: %w", openErr)
		}
		defer log.Close()
		log.Event("Local operator requested "+req.Action, "activity", "staff.operator", "action", req.Action,
			"account_id", req.AccountID, "character_id", req.CharacterID, "outcome", "requested")
		defer func() {
			outcome := "completed"
			if err != nil {
				outcome = "rejected"
			}
			accountID := req.AccountID
			if res.Account != nil {
				accountID = res.Account.ID
			}
			log.Event(fmt.Sprintf("Local operator %s %s (account %d, character %d)", req.Action, outcome, accountID, req.CharacterID),
				"activity", "staff.operator", "action", req.Action, "account_id", accountID,
				"character_id", req.CharacterID, "outcome", outcome)
		}()
	}
	switch req.Action {
	case "accounts", "create", "details", "stats", "save-stats", "account", "password", "delete", "restore",
		"ban-account", "unban-account", "ip-bans", "ban-ip", "unban-ip":
		processes, err := m.processes()
		if err != nil {
			return res, err
		}
		if len(processes[World]) != 0 {
			return m.worldOperator(ctx, processes[World], req)
		}
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return res, err
	}
	defer clear(payload)
	name := "gx_console"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(m.root, "build", name), "operator-json")
	cmd.Dir, cmd.Stdin = m.root, bytes.NewReader(payload)
	detach(cmd)
	var output, problem bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &problem
	if err := cmd.Run(); err != nil {
		if detail := strings.TrimSpace(problem.String()); detail != "" {
			return res, errors.New(detail)
		}
		return res, fmt.Errorf("local operator command: %w", err)
	}
	err = json.Unmarshal(output.Bytes(), &res)
	return res, err
}

func (m *Manager) worldOperator(ctx context.Context, pids []int, req operatorwire.Request) (operatorwire.Response, error) {
	if len(pids) != 1 {
		return operatorwire.Response{}, errors.New("multiple World processes match this installation")
	}
	return controlplane.RunOperator(ctx, m.root, pids[0], m.binary(World), req)
}
