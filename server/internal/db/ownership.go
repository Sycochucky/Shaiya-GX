package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/filelock"
)

// AcquireWorld enforces one authoritative simulation process per runtime. The
// OS releases ownership on crash; a new owner removes stale online leases.
func AcquireWorld(ctx context.Context) (func(), error) {
	release, err := lockRuntime(ctx, "world")
	if err != nil {
		return nil, err
	}
	pool, err := Pool(GameData)
	if err != nil {
		release()
		return nil, err
	}
	cleanupCtx, cancelCleanup := context.WithTimeout(ctx, Timeout)
	defer cancelCleanup()
	if _, err = pool.ExecContext(cleanupCtx, `DELETE FROM GX_ActiveSessions`); err != nil {
		release()
		return nil, fmt.Errorf("clear stale world sessions: %w", err)
	}
	return release, nil
}

func AcquireLogin(ctx context.Context) (func(), error) {
	return lockRuntime(ctx, "login")
}

// AcquireWorldAdministration prevents offline edits from racing World startup.
func AcquireWorldAdministration(ctx context.Context) (func(), error) {
	return lockRuntime(ctx, "world")
}

func lockRuntime(ctx context.Context, service string) (func(), error) {
	path := config.DatabasePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	if real, err := filepath.EvalSymlinks(path); err == nil {
		path = real
	} else if real, err := filepath.EvalSymlinks(dir); err == nil {
		path = filepath.Join(real, filepath.Base(path))
	}
	leaseCtx, cancelLease := context.WithTimeout(ctx, time.Second)
	defer cancelLease()
	release, err := filelock.Acquire(leaseCtx, path+"."+service+".lock")
	if err != nil {
		return nil, fmt.Errorf("acquire %s runtime lease: %w", service, err)
	}
	return release, nil
}
func AcquireAccount(uid uint32, id int64) (bool, error) {
	pool, err := Pool(UserData)
	if err != nil {
		return false, err
	}
	c, cancel := ctx()
	defer cancel()
	result, err := pool.ExecContext(c, `INSERT INTO GX_ActiveSessions(UserUID,SessionID) VALUES(?,?) ON CONFLICT(UserUID) DO NOTHING`, uid, id)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}
func ReleaseAccount(uid uint32, id int64) error {
	pool, err := Pool(UserData)
	if err != nil {
		return err
	}
	c, cancel := ctx()
	defer cancel()
	_, err = pool.ExecContext(c, `DELETE FROM GX_ActiveSessions WHERE UserUID=? AND SessionID=?`, uid, id)
	return err
}

var releases = struct {
	sync.Mutex
	pending map[uint32]int64
}{pending: map[uint32]int64{}}

// Failed disconnect cleanup is retried with its exact session identity, never
// deleting a newer account lease. Cardinality is bounded by admission capacity.
func RetryAccountRelease(uid uint32, id int64) {
	releases.Lock()
	if len(releases.pending) < 4096 {
		releases.pending[uid] = id
	}
	releases.Unlock()
}
func RetryAccountReleases() {
	releases.Lock()
	pending := make(map[uint32]int64, len(releases.pending))
	for uid, id := range releases.pending {
		pending[uid] = id
	}
	releases.Unlock()
	for uid, id := range pending {
		if err := ReleaseAccount(uid, id); err == nil {
			releases.Lock()
			if releases.pending[uid] == id {
				delete(releases.pending, uid)
			}
			releases.Unlock()
		}
	}
}
