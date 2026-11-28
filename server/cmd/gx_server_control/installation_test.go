package main

import (
	"path/filepath"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
)

func TestRememberInstallationKeepsMultipleUniqueRoots(t *testing.T) {
	home := t.TempDir()
	first := filepath.Join(home, "server-one")
	second := filepath.Join(home, "server-two")
	for _, chosen := range []string{first, second, first} {
		if err := rememberInstallation(home, chosen); err != nil {
			t.Fatal(err)
		}
	}
	stored := rememberedInstallations(home)
	if len(stored) != 2 {
		t.Fatalf("stored installations = %v", stored)
	}
	for _, want := range []string{first, second} {
		found := false
		for _, got := range stored {
			found = found || control.SameInstallation(got, want)
		}
		if !found {
			t.Fatalf("missing %s in %v", want, stored)
		}
	}
}
