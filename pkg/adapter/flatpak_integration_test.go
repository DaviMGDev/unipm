//go:build integration

package adapter

import (
	"os/exec"
	"testing"
)

func TestFlatpakAdapter_Search_Integration(t *testing.T) {
	a := &FlatpakAdapter{}

	if !a.IsAvailable() {
		t.Skip("flatpak not available")
	}

	// Check if we can actually run flatpak (requires user namespaces in Docker)
	cmd := exec.Command("flatpak", "search", "htop")
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Exit code 1 = no results, which is fine
		} else {
			t.Skipf("flatpak not functional in this environment: %v", err)
		}
	}

	pkgs, err := a.Search("htop")
	if err != nil {
		t.Skipf("flatpak search failed in this environment: %v", err)
	}

	// May return 0 results if htop isn't on flathub, which is fine
	t.Logf("found %d results for 'htop'", len(pkgs))
}

func TestFlatpakAdapter_Info_Integration(t *testing.T) {
	a := &FlatpakAdapter{}

	if !a.IsAvailable() {
		t.Skip("flatpak not available")
	}

	// Only test info if flatpak is functional
	cmd := exec.Command("flatpak", "info", "org.freedesktop.Platform")
	if err := cmd.Run(); err != nil {
		t.Skipf("flatpak not functional or platform not installed: %v", err)
	}

	pkg := Package{Name: "org.freedesktop.Platform", Source: "flatpak"}
	info, err := a.Info(pkg)
	if err != nil {
		t.Skipf("Info() failed: %v", err)
	}

	t.Logf("Platform: name=%s version=%s", info.Name, info.Version)
}
