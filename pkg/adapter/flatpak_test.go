package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFlatpakAdapter_Name(t *testing.T) {
	a := &FlatpakAdapter{}
	if name := a.Name(); name != "flatpak" {
		t.Errorf("Name() = %q, want %q", name, "flatpak")
	}
}

func TestFlatpakAdapter_IsAvailable(t *testing.T) {
	a := &FlatpakAdapter{}
	available := a.IsAvailable()
	t.Logf("flatpak IsAvailable() = %v", available)
}

func TestParseFlatpakSearch_GoldenFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "flatpak_search_htop.txt"))
	if err != nil {
		t.Skipf("golden fixture not available: %v", err)
	}

	packages := parseFlatpakSearch(string(data))

	if len(packages) == 0 {
		t.Fatal("expected at least one package in search results")
	}

	// Check that Flatseal is in the results
	found := false
	for _, p := range packages {
		if p.Name == "com.github.tchx84.Flatseal" {
			found = true
			if p.Source != "flatpak" {
				t.Errorf("Source = %q, want %q", p.Source, "flatpak")
			}
			if p.Version == "" {
				t.Error("Version should not be empty for Flatseal")
			}
			if p.Description == "" {
				t.Error("Description should not be empty for Flatseal")
			}
			t.Logf("Flatseal: version=%s description=%s", p.Version, p.Description)
			break
		}
	}
	if !found {
		t.Error("com.github.tchx84.Flatseal not found in parsed search results")
	}
}

func TestParseFlatpakSearch_EmptyOutput(t *testing.T) {
	packages := parseFlatpakSearch("")
	if len(packages) != 0 {
		t.Errorf("expected 0 packages from empty output, got %d", len(packages))
	}
}

func TestParseFlatpakSearch_HeaderOnly(t *testing.T) {
	output := "Name\tDescription\tApplication ID\tVersion\tBranch\tRemotes\n"
	packages := parseFlatpakSearch(output)
	if len(packages) != 0 {
		t.Errorf("expected 0 packages from header-only output, got %d", len(packages))
	}
}

func TestParseFlatpakSearch_MalformedLine(t *testing.T) {
	output := "Name\tDescription\tApplication ID\tVersion\tBranch\tRemotes\n" +
		"just-one-column\n"
	packages := parseFlatpakSearch(output)
	if len(packages) != 0 {
		t.Errorf("expected 0 packages from malformed line, got %d", len(packages))
	}
}

func TestParseFlatpakInfo_GoldenFixture(t *testing.T) {
	// Inline fixture since flatpak info output is simple key-value.
	output := `          ID: com.github.tchx84.Flatseal
         Ref: app/com.github.tchx84.Flatseal/x86_64/stable
        Arch: x86_64
      Branch: stable
     Version: 2.3.0
`

	details := parseFlatpakInfo(output)

	if details.Name != "com.github.tchx84.Flatseal" {
		t.Errorf("Name = %q, want %q", details.Name, "com.github.tchx84.Flatseal")
	}
	if details.Version != "2.3.0" {
		t.Errorf("Version = %q, want %q", details.Version, "2.3.0")
	}
}

func TestParseFlatpakInfo_Empty(t *testing.T) {
	details := parseFlatpakInfo("")
	if details.Name != "" {
		t.Errorf("Name = %q, want empty", details.Name)
	}
}

func TestFlatpakAdapter_FlagConstruction(t *testing.T) {
	a := &FlatpakAdapter{}

	pkg := Package{Name: "com.github.tchx84.Flatseal", Source: "flatpak"}
	if pkg.Source != a.Name() {
		t.Errorf("package source %q != adapter name %q", pkg.Source, a.Name())
	}
}
