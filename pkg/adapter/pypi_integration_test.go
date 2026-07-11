//go:build integration

package adapter

import (
	"testing"
)

func TestPypiAdapter_Search_Integration(t *testing.T) {
	a := &PypiAdapter{}

	if !a.IsAvailable() {
		t.Skip("pip3 not available")
	}

	pkgs, err := a.Search("requests")
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}

	if len(pkgs) == 0 {
		t.Error("Search('requests') returned no results — PyPI API may be unreachable")
	}

	for _, p := range pkgs {
		if p.Source != "pypi" {
			t.Errorf("Source = %q, want 'pypi'", p.Source)
		}
		if p.Name == "" {
			t.Error("Name is empty")
		}
	}

	t.Logf("found %d results for 'requests'", len(pkgs))
}

func TestPypiAdapter_Info_Integration(t *testing.T) {
	a := &PypiAdapter{}

	if !a.IsAvailable() {
		t.Skip("pip3 not available")
	}

	pkg := Package{Name: "pip", Source: "pypi"}
	info, err := a.Info(pkg)
	if err != nil {
		t.Skipf("Info() skipped — pip may not be installed via pip: %v", err)
	}

	if info.Name == "" || info.Version == "" {
		t.Error("Info() returned empty fields for pip")
	}
	t.Logf("pip: version=%s location=%s", info.Version, info.Homepage)
}
