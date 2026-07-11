//go:build integration

package adapter

import (
	"testing"
)

func TestAppImageAdapter_Search_Integration(t *testing.T) {
	a := &AppImageAdapter{}

	if !a.IsAvailable() {
		t.Skip("curl/wget not available")
	}

	pkgs, err := a.Search("etcher")
	if err != nil {
		t.Skipf("AppImageHub API unreachable: %v", err)
	}

	// May return 0 results — AppImageHub catalog changes over time
	t.Logf("found %d results for 'etcher'", len(pkgs))

	for _, p := range pkgs {
		if p.Source != "appimage" {
			t.Errorf("Source = %q, want 'appimage'", p.Source)
		}
	}
}

func TestAppImageAdapter_Info_Integration(t *testing.T) {
	a := &AppImageAdapter{}

	pkg := Package{Name: "Etcher", Source: "appimage"}
	_, err := a.Info(pkg)
	if err == nil {
		t.Error("Info() should return error for AppImage (not supported)")
	}
	t.Logf("Info() correctly returned error: %v", err)
}
