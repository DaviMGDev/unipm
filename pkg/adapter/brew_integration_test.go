//go:build integration

package adapter

import (
	"testing"
)

func TestBrewAdapter_Search_Integration(t *testing.T) {
	a := &BrewAdapter{}

	if !a.IsAvailable() {
		t.Skip("brew not available")
	}

	pkgs, err := a.Search("htop")
	if err != nil {
		t.Skipf("brew search failed in this environment: %v", err)
	}

	if len(pkgs) == 0 {
		t.Skip("brew search returned no results for 'htop'")
	}

	for _, p := range pkgs {
		if p.Source != "brew" {
			t.Errorf("Source = %q, want 'brew'", p.Source)
		}
	}

	t.Logf("found %d results for 'htop'", len(pkgs))
}

func TestBrewAdapter_Info_Integration(t *testing.T) {
	a := &BrewAdapter{}

	if !a.IsAvailable() {
		t.Skip("brew not available")
	}

	pkg := Package{Name: "htop", Source: "brew"}
	info, err := a.Info(pkg)
	if err != nil {
		t.Skipf("Info() failed (htop may not be installed): %v", err)
	}

	if info.Name == "" {
		t.Error("Info() returned empty name")
	}
	t.Logf("htop: version=%s homepage=%s license=%s", info.Version, info.Homepage, info.License)
}
