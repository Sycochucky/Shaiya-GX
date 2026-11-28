//go:build !windows

package controlplane

// os.CreateTemp creates the descriptor with 0600 permissions.
func protectDescriptor(string) error { return nil }
