// Command gx_console controls the detached GX services from a terminal.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Spelunkern/Shaiya-GX/server/internal/admin"
	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) > 1 {
		return admin.Run(os.Args[1:])
	}
	manager, err := control.Installation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	show := func() {
		status := manager.Snapshot(ctx)
		for _, service := range control.Services() {
			info := status[service]
			fmt.Printf("%s: %s", service.Label(), info.State)
			if info.Detail != "" {
				fmt.Printf(" (%s)", info.Detail)
			}
			fmt.Println()
		}
	}
	commands := map[string]control.Action{
		"start server":    control.Start,
		"shutdown server": control.Stop,
		"restart server":  control.Restart,
	}
	fmt.Println("GX Server Console")
	fmt.Println(`Commands: "status" | "start server" | "shutdown server" | "restart server"`)
	show()
	input := bufio.NewScanner(os.Stdin)
	for fmt.Print("> "); input.Scan(); fmt.Print("> ") {
		command := strings.ToLower(strings.TrimSpace(input.Text()))
		action, ok := commands[command]
		switch {
		case ok:
			if err := manager.Apply(ctx, action, control.Services(), nil); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			show()
		case command == "status":
			show()
		case command != "":
			fmt.Println("Unknown command.")
		}
	}
	return input.Err()
}
