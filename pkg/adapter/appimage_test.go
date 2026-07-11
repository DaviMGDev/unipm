package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAppImageAdapter_Name(t *testing.T) {
	a := &AppImageAdapter{}
	if name := a.Name(); name != "appimage" {
		t.Errorf("Name() = %q, want %q", name, "appimage")
	}
}

func TestAppImageAdapter_IsAvailable(t *testing.T) {
	a := &AppImageAdapter{}
	available := a.IsAvailable()
	t.Logf("appimage IsAvailable() = %v", available)
}

func TestParseAppImageResults_GoldenFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "appimage_search_htop.json"))
	if err != nil {
		t.Skipf("golden fixture not available: %v", err)
	}

	var entries []appImageHubEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	packages := parseAppImageResults(entries)

	if len(packages) == 0 {
		t.Fatal("expected at least one package")
	}

	if packages[0].Name == "" {
		t.Error("Name should not be empty")
	}
	if packages[0].Source != "appimage" {
		t.Errorf("Source = %q, want %q", packages[0].Source, "appimage")
	}

	t.Logf("first result: name=%s version=%s description=%s",
		packages[0].Name, packages[0].Version, packages[0].Description)
}

func TestParseAppImageResults_Empty(t *testing.T) {
	packages := parseAppImageResults(nil)
	if len(packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(packages))
	}
}

func TestExtractVersionFromURL(t *testing.T) {
	tests := []struct {
		url     string
		wantVer string
	}{
		{
			url:     "https://example.com/MyApp-2.3.0-x86_64.AppImage",
			wantVer: "2.3.0-x86_64",
		},
		{
			url:     "https://example.com/App-v1.0.0-x86_64.AppImage",
			wantVer: "v1.0.0-x86_64",
		},
		{
			url:     "https://example.com/SimpleApp.AppImage",
			wantVer: "",
		},
		{
			url:     "",
			wantVer: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := extractVersionFromURL(tt.url)
			if got != tt.wantVer {
				t.Errorf("extractVersionFromURL(%q) = %q, want %q", tt.url, got, tt.wantVer)
			}
		})
	}
}

func TestAppImageAdapter_Info_NotSupported(t *testing.T) {
	a := &AppImageAdapter{}
	_, err := a.Info(Package{Name: "MyApp", Source: "appimage"})
	if err == nil {
		t.Error("Info() should return error for AppImage (not supported)")
	}
}

func TestAppImageAdapter_FlagConstruction(t *testing.T) {
	a := &AppImageAdapter{}

	pkg := Package{Name: "MyApp", Source: "appimage"}
	if pkg.Source != a.Name() {
		t.Errorf("package source %q != adapter name %q", pkg.Source, a.Name())
	}
}
