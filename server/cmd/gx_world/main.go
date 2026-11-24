// Command gx_world is the Shaiya 5.4 game server on the original
// EP4.5 tables translated to SQLite: character select, world simulation (maps,
// NPCs, mobs, combat, skills), items, trade and the item mall.
package main

import (
	"context"
	"fmt"
	"github.com/Spelunkern/Shaiya-GX/internal/transport"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/controlplane"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/world"
)

var nextConnection atomic.Int64

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	if err := config.LoadRuntimeEnv(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	policy, err := config.Network("world")
	if err != nil {
		return err
	}
	logger, err := logx.New("gx_world")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	defer logger.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	resetFailure := db.OnStorageFailure(func(err error) { logger.Errorf("storage unavailable; stopping service: %v", err); stop() })
	defer resetFailure()
	defer db.Close()
	logger.Printf("Starting GX Server World")
	if err := db.Open(db.UserData, db.GameData, db.GameDefs); err != nil {
		logger.Printf("GX Server World database error: %v", err)
		return err
	}
	ownership, cancelOwnership := context.WithTimeout(context.Background(), time.Second)
	release, err := db.AcquireWorld(ownership)
	cancelOwnership()
	if err != nil {
		logger.Printf("world runtime already owned or unavailable: %v", err)
		return err
	}
	defer release()
	defs, err := db.Definitions()
	if err != nil {
		logger.Printf("GX Server World failed to load PS_GameDefs: %v", err)
		return err
	}
	logger.Printf("GX Server World loaded definitions items=%d skills=%d mobs=%d products=%d status=%d base_items=%d",
		len(defs.Items), len(defs.Skills), len(defs.Mobs), len(defs.Products), len(defs.Status), len(defs.BaseItems))
	if err := world.LoadGameConfig(logger); err != nil {
		logger.Printf("GX Server World failed to load data/game (copy the EP4.5 PS_Game/Data files there): %v", err)
		return err
	}
	world.InitWorldCalendar(logger)
	world.InitKillCount(logger)
	if err := world.LoadGuilds(logger); err != nil {
		logger.Printf("GX Server World failed to load guilds: %v", err)
		return err
	}
	if err := world.LoadSvmaps(logger); err != nil {
		logger.Printf("GX Server World failed to load maps: %v", err)
		return err
	}
	if err := world.LoadRelics(logger); err != nil {
		logger.Printf("GX Server World failed to load obelisks: %v", err)
		return err
	}

	listener, err := transport.Listen(config.WorldListen())
	if err != nil {
		logger.Errorf("GX Server World failed to bind %s: %v", config.WorldListen(), err)
		return err
	}
	logger.Printf("GX Server World listening on %s", config.WorldListen())

	engine := world.StartSweeps(ctx, logger, stop)
	defer engine.Close()
	admin, err := controlplane.Start(ctx, "gx_world", os.Getenv("RAPTURE_WORLD_DIAGNOSTICS"), stop, func() controlplane.ServiceStatus {
		return controlplane.ServiceStatus{Endpoint: listener.Addr().String(), Players: world.PlayerCount()}
	}, db.Observe)
	if err != nil {
		listener.Close()
		logger.Printf("local control startup failed: %v", err)
		return err
	}
	defer admin.Close()
	logger.Event("local diagnostics ready", "address", admin.Address())
	if err := netx.Serve(ctx, listener, policy.Connections, policy.PerIP, func(conn net.Conn) {
		log := logger.Tagged(fmt.Sprintf("conn%d", nextConnection.Add(1)))
		defer func() {
			if r := recover(); r != nil {
				log.Printf("world session panic: %v", r)
			}
		}()
		if err := world.Serve(conn, logger, log); err != nil {
			log.Printf("world client error: %v", err)
		}
	}, netx.NewRateLimiter(4096, policy.Establishments, time.Minute)); err != nil {
		logger.Printf("server accept error: %v", err)
		return err
	}

	return nil
}
