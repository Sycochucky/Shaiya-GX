//go:build integration && linux

package control

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/controlplane"
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
)

func TestShieldManagerChangesBothServices(t *testing.T) {
	if service := os.Getenv("GX_SHIELD_TEST_SERVICE"); service != "" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		server, err := controlplane.Start(ctx, service, "", cancel, func() controlplane.ServiceStatus {
			return controlplane.ServiceStatus{Endpoint: "127.0.0.1:0"}
		}, netx.NewShield())
		if err != nil {
			t.Fatal(err)
		}
		defer server.Close()
		<-ctx.Done()
		return
	}

	root := t.TempDir()
	build := filepath.Join(root, "build")
	if err := os.Mkdir(build, 0700); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, service := range Services() {
		to := filepath.Join(build, string(service))
		from, err := os.Open(executable)
		if err != nil {
			t.Fatal(err)
		}
		output, err := os.OpenFile(to, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
		if err != nil {
			from.Close()
			t.Fatal(err)
		}
		_, copyErr := io.Copy(output, from)
		closeErr := output.Close()
		from.Close()
		if copyErr != nil || closeErr != nil {
			t.Fatalf("copy %s helper: %v %v", service, copyErr, closeErr)
		}
		cmd := exec.Command(to, "-test.run=^TestShieldManagerChangesBothServices$")
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GX_SHIELD_TEST_SERVICE="+string(service))
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		})
	}

	m := New(root)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		status := m.Snapshot(ctx)
		if status[Login].State == Ready && status[World].State == Ready {
			break
		}
		if err := ctx.Err(); err != nil {
			t.Fatalf("helpers did not become ready: %v; login=%+v world=%+v", err, status[Login], status[World])
		}
		time.Sleep(25 * time.Millisecond)
	}
	for _, minutes := range []int{1, 0} {
		if err := m.SetShield(ctx, minutes); err != nil {
			t.Fatal(err)
		}
		want := netx.ShieldStandard
		if minutes > 0 {
			want = netx.ShieldUnderAttack
		}
		for service, state := range m.Snapshot(ctx) {
			if state.Shield.Mode != want {
				t.Fatalf("%s shield mode = %q, want %q", service, state.Shield.Mode, want)
			}
		}
	}
}
