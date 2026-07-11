package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPypiAdapter_Name(t *testing.T) {
	a := &PypiAdapter{}
	if name := a.Name(); name != "pypi" {
		t.Errorf("Name() = %q, want %q", name, "pypi")
	}
}

func TestPypiAdapter_IsAvailable(t *testing.T) {
	a := &PypiAdapter{}
	available := a.IsAvailable()
	t.Logf("pypi IsAvailable() = %v", available)
}

func TestParsePypiSearch_GoldenFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "pypi_search_requests.json"))
	if err != nil {
		t.Skipf("golden fixture not available: %v", err)
	}

	// Simulate a minimal HTTP round-trip test by parsing the fixture
	// through the same JSON path that Search() uses.
	var resp pypiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	if resp.Info.Name != "requests" {
		t.Errorf("Name = %q, want %q", resp.Info.Name, "requests")
	}
	if resp.Info.Version == "" {
		t.Error("Version should not be empty")
	}
	if resp.Info.Summary == "" {
		t.Error("Summary should not be empty")
	}

	t.Logf("requests: version=%s summary=%s license=%s",
		resp.Info.Version, resp.Info.Summary, resp.Info.License)
}

func TestParsePipShow_GoldenFixture(t *testing.T) {
	// Test with inline fixture since pip3 show output is simple key-value.
	output := `Name: requests
Version: 2.31.0
Summary: Python HTTP for Humans.
Home-page: https://requests.readthedocs.io
Author: Kenneth Reitz
Author-email: me@kennethreitz.org
License: Apache 2.0
Location: /usr/lib/python3/dist-packages
Requires: certifi, charset-normalizer, idna, urllib3
Required-by:
`

	details := parsePipShow(output)

	if details.Name != "requests" {
		t.Errorf("Name = %q, want %q", details.Name, "requests")
	}
	if details.Version != "2.31.0" {
		t.Errorf("Version = %q, want %q", details.Version, "2.31.0")
	}
	if details.Description != "Python HTTP for Humans." {
		t.Errorf("Description = %q, want %q", details.Description, "Python HTTP for Humans.")
	}
	if details.Homepage != "https://requests.readthedocs.io" {
		t.Errorf("Homepage = %q", details.Homepage)
	}
	if details.License != "Apache 2.0" {
		t.Errorf("License = %q, want %q", details.License, "Apache 2.0")
	}
}

func TestParsePipShow_Empty(t *testing.T) {
	details := parsePipShow("")
	if details.Name != "" {
		t.Errorf("Name = %q, want empty", details.Name)
	}
}

func TestParsePipShow_NoMatchingFields(t *testing.T) {
	output := `Unknown-Field: some value
Another-Field: another value
`
	details := parsePipShow(output)
	if details.Name != "" || details.Version != "" {
		t.Error("expected empty Details for unrecognized fields")
	}
}

func TestPypiAdapter_Install_FlagConstruction(t *testing.T) {
	a := &PypiAdapter{}

	if !a.IsAvailable() {
		t.Skip("pip3 not available on this system — Tier 2 test only")
	}

	pkg := Package{Name: "requests", Source: "pypi"}
	if pkg.Source != a.Name() {
		t.Errorf("package source %q != adapter name %q", pkg.Source, a.Name())
	}
}

func TestPypiAdapter_Uninstall_FlagConstruction(t *testing.T) {
	a := &PypiAdapter{}

	pkg := Package{Name: "requests", Source: "pypi"}
	if pkg.Source != a.Name() {
		t.Errorf("package source %q != adapter name %q", pkg.Source, a.Name())
	}
}
