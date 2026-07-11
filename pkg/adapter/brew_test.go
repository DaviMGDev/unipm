package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBrewAdapter_Name(t *testing.T) {
	a := &BrewAdapter{}
	if name := a.Name(); name != "brew" {
		t.Errorf("Name() = %q, want %q", name, "brew")
	}
}

func TestBrewAdapter_IsAvailable(t *testing.T) {
	a := &BrewAdapter{}
	available := a.IsAvailable()
	t.Logf("brew IsAvailable() = %v", available)
}

func TestParseBrewSearch_GoldenFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "brew_search_htop.txt"))
	if err != nil {
		t.Skipf("golden fixture not available: %v", err)
	}

	packages := parseBrewSearch(string(data))

	if len(packages) < 2 {
		t.Fatalf("expected at least 2 packages, got %d", len(packages))
	}

	// Check that htop is in the results
	found := false
	for _, p := range packages {
		if p.Name == "htop" {
			found = true
			if p.Source != "brew" {
				t.Errorf("Source = %q, want %q", p.Source, "brew")
			}
			t.Logf("htop found in brew search results")
			break
		}
	}
	if !found {
		t.Error("htop not found in parsed brew search results")
	}
}

func TestParseBrewSearch_EmptyOutput(t *testing.T) {
	packages := parseBrewSearch("")
	if len(packages) != 0 {
		t.Errorf("expected 0 packages from empty output, got %d", len(packages))
	}
}

func TestParseBrewSearch_HeadersOnly(t *testing.T) {
	output := "==> Formulae\n==> Casks\n"
	packages := parseBrewSearch(output)
	if len(packages) != 0 {
		t.Errorf("expected 0 packages from headers-only output, got %d", len(packages))
	}
}

func TestParseBrewSearch_SkipsEmptyLines(t *testing.T) {
	output := "\n\nhtop\n\n\n"
	packages := parseBrewSearch(output)
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}
	if packages[0].Name != "htop" {
		t.Errorf("Name = %q, want %q", packages[0].Name, "htop")
	}
}

func TestParseBrewInfo_GoldenFixture(t *testing.T) {
	// brew info --json returns a JSON array.
	output := `[{
  "name": "htop",
  "full_name": "htop",
  "desc": "Improved top (interactive process viewer)",
  "homepage": "https://htop.dev/",
  "license": "GPL-2.0-or-later",
  "versions": {
    "stable": "3.4.1",
    "head": "HEAD",
    "bottle": true
  }
}]`

	details := parseBrewInfo(output)

	if details.Name != "htop" {
		t.Errorf("Name = %q, want %q", details.Name, "htop")
	}
	if details.Version != "3.4.1" {
		t.Errorf("Version = %q, want %q", details.Version, "3.4.1")
	}
	if details.Description != "Improved top (interactive process viewer)" {
		t.Errorf("Description = %q", details.Description)
	}
	if details.Homepage != "https://htop.dev/" {
		t.Errorf("Homepage = %q", details.Homepage)
	}
	if details.License != "GPL-2.0-or-later" {
		t.Errorf("License = %q", details.License)
	}
}

func TestParseBrewInfo_Empty(t *testing.T) {
	details := parseBrewInfo("[]")
	if details.Name != "" {
		t.Errorf("Name = %q, want empty", details.Name)
	}
}

func TestParseBrewInfo_InvalidJSON(t *testing.T) {
	details := parseBrewInfo("not json")
	if details.Name != "" {
		t.Errorf("Name = %q, want empty", details.Name)
	}
}

func TestBrewAdapter_FlagConstruction(t *testing.T) {
	a := &BrewAdapter{}

	pkg := Package{Name: "htop", Source: "brew"}
	if pkg.Source != a.Name() {
		t.Errorf("package source %q != adapter name %q", pkg.Source, a.Name())
	}
}
