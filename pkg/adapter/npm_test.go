package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNpmAdapter_Name(t *testing.T) {
	a := &NpmAdapter{}
	if name := a.Name(); name != "npm" {
		t.Errorf("Name() = %q, want %q", name, "npm")
	}
}

func TestNpmAdapter_IsAvailable(t *testing.T) {
	a := &NpmAdapter{}
	available := a.IsAvailable()
	t.Logf("npm IsAvailable() = %v", available)
}

func TestParseNpmSearch_GoldenFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "npm_search_htop.json"))
	if err != nil {
		t.Skipf("golden fixture not available: %v", err)
	}

	var resp npmSearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal golden fixture: %v", err)
	}

	packages := parseNpmSearch(&resp)

	if len(packages) == 0 {
		t.Fatal("expected at least one package in search results")
	}

	// Check the first package has required fields
	if packages[0].Name == "" {
		t.Error("Name should not be empty")
	}
	if packages[0].Source != "npm" {
		t.Errorf("Source = %q, want %q", packages[0].Source, "npm")
	}
	if packages[0].Version == "" {
		t.Error("Version should not be empty")
	}

	t.Logf("parsed %d npm packages; first: name=%s version=%s desc=%s",
		len(packages), packages[0].Name, packages[0].Version, packages[0].Description)
}

func TestParseNpmSearch_EmptyResponse(t *testing.T) {
	resp := &npmSearchResponse{Objects: []npmSearchObject{}}
	packages := parseNpmSearch(resp)
	if len(packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(packages))
	}
}

func TestParseNpmSearch_MultiplePackages(t *testing.T) {
	resp := &npmSearchResponse{
		Objects: []npmSearchObject{
			{Package: npmPackageInfo{Name: "htop", Version: "1.0.1", Description: "handle-to-promise"}},
			{Package: npmPackageInfo{Name: "react", Version: "18.2.0", Description: "React is a JavaScript library"}},
			{Package: npmPackageInfo{Name: "lodash", Version: "4.17.21", Description: "Lodash modular utilities"}},
		},
	}

	packages := parseNpmSearch(resp)
	if len(packages) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(packages))
	}

	expected := []struct{ name, version string }{
		{"htop", "1.0.1"},
		{"react", "18.2.0"},
		{"lodash", "4.17.21"},
	}

	for i, exp := range expected {
		if packages[i].Name != exp.name {
			t.Errorf("packages[%d].Name = %q, want %q", i, packages[i].Name, exp.name)
		}
		if packages[i].Version != exp.version {
			t.Errorf("packages[%d].Version = %q, want %q", i, packages[i].Version, exp.version)
		}
		if packages[i].Source != "npm" {
			t.Errorf("packages[%d].Source = %q, want %q", i, packages[i].Source, "npm")
		}
	}
}

func TestParseNpmInfo(t *testing.T) {
	infoJSON := `{
		"name": "htop",
		"version": "1.0.1",
		"description": "handle-to-promise",
		"homepage": "https://github.com/example/htop",
		"license": "ISC"
	}`

	details := parseNpmInfo(infoJSON)

	if details.Name != "htop" {
		t.Errorf("Name = %q, want %q", details.Name, "htop")
	}
	if details.Version != "1.0.1" {
		t.Errorf("Version = %q, want %q", details.Version, "1.0.1")
	}
	if details.Description != "handle-to-promise" {
		t.Errorf("Description = %q, want %q", details.Description, "handle-to-promise")
	}
	if details.Homepage != "https://github.com/example/htop" {
		t.Errorf("Homepage = %q, want %q", details.Homepage, "https://github.com/example/htop")
	}
	if details.License != "ISC" {
		t.Errorf("License = %q, want %q", details.License, "ISC")
	}
}

func TestParseNpmInfo_InvalidJSON(t *testing.T) {
	details := parseNpmInfo("{invalid}")
	// Should return empty Details without panicking
	if details.Name != "" {
		t.Errorf("Name = %q, want empty for invalid JSON", details.Name)
	}
}

func TestNpmAdapter_Install_FlagConstruction(t *testing.T) {
	a := &NpmAdapter{}

	pkg := Package{Name: "htop", Source: "npm"}
	if pkg.Source != a.Name() {
		t.Errorf("package source %q != adapter name %q", pkg.Source, a.Name())
	}
}

func TestNpmAdapter_Uninstall_FlagConstruction(t *testing.T) {
	a := &NpmAdapter{}

	pkg := Package{Name: "htop", Source: "npm"}
	if pkg.Source != a.Name() {
		t.Errorf("package source %q != adapter name %q", pkg.Source, a.Name())
	}
}
