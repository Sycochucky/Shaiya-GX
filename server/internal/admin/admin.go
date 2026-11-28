// Package admin supplies local operator commands without a web admin service.
package admin

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"golang.org/x/term"
	"io"
	"os"
)

func secret() (string, error) {
	if value, ok := os.LookupEnv("RAPTURE_ACCOUNT_PASSWORD"); ok {
		return value, nil
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(os.Stderr, "Password: ")
		value, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		return string(value), err
	}
	value, err := bufio.NewReader(io.LimitReader(os.Stdin, 1024)).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	if len(value) > 0 && value[len(value)-1] == '\n' {
		value = value[:len(value)-1]
		if len(value) > 0 && value[len(value)-1] == '\r' {
			value = value[:len(value)-1]
		}
	}
	return value, nil
}
func Run(args []string) error {
	if len(args) == 0 {
		return errors.New("use account create NAME [--admin], account password NAME, account upgrade-legacy, or tls-init")
	}
	if err := config.LoadRuntimeEnv(); err != nil {
		return err
	}
	if args[0] == "tls-init" {
		return certificates(args[1:])
	}
	if len(args) == 1 && args[0] == "operator-json" {
		return runOperator(os.Stdin, os.Stdout)
	}
	if len(args) == 2 && args[0] == "account" && args[1] == "upgrade-legacy" {
		defer db.Close()
		upgraded, needsReset, err := db.UpgradeLegacyPasswords()
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]int{"upgraded": upgraded, "needs_reset": needsReset})
	}
	if args[0] != "account" || len(args) < 3 {
		return errors.New("use account create NAME [--admin] or account password NAME")
	}
	defer db.Close()
	name := args[2]
	flags := flag.NewFlagSet("account", flag.ContinueOnError)
	admin := flags.Bool("admin", false, "create a GM account")
	if err := flags.Parse(args[3:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected arguments")
	}
	value, err := secret()
	if err != nil {
		return err
	}
	switch args[1] {
	case "create":
		uid, err := db.CreateAccount(name, value, *admin)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"created": name, "uid": uid})
	case "password":
		if *admin {
			return errors.New("password command cannot change roles")
		}
		if err := db.SetAccountPassword(name, value); err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"password_updated": name, "sessions_revoked": true})
	default:
		return errors.New("unknown account command")
	}
}
