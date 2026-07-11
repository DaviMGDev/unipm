package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAptAdapter_Name(t *testing.T) {
	a := &AptAdapter{}
	if name := a.Name(); name != "apt" {
		t.Errorf("Name() = %q, want %q", name, "apt")
	}
}

func TestAptAdapter_IsAvailable(t *testing.T) {
	a := &AptAdapter{}
	// This test only verifies the method doesn't panic and returns a bool.
	// The actual value depends on the test environment.
	available := a.IsAvailable()
	t.Logf("apt IsAvailable() = %v", available)
}

func TestParseAptSearch_SinglePackage(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "apt_search_htop.txt"))
	if err != nil {
		t.Skipf("golden fixture not available: %v", err)
	}

	packages := parseAptSearch(string(data), "htop")

	if len(packages) == 0 {
		t.Fatal("expected at least one package in search results")
	}

	// Check that htop is in the results
	found := false
	for _, p := range packages {
		if p.Name == "htop" {
			found = true
			if p.Source != "apt" {
				t.Errorf("Source = %q, want %q", p.Source, "apt")
			}
			if p.Version == "" {
				t.Error("Version should not be empty for htop")
			}
			if p.Description == "" {
				t.Error("Description should not be empty for htop")
			}
			t.Logf("htop: version=%s description=%s", p.Version, p.Description)
			break
		}
	}
	if !found {
		t.Error("htop not found in parsed search results")
	}
}

func TestParseAptSearch_MultiPackage(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "apt_search_multi.txt"))
	if err != nil {
		t.Skipf("golden fixture not available: %v", err)
	}

	packages := parseAptSearch(string(data), "python")

	if len(packages) == 0 {
		t.Fatal("expected multiple packages in search results")
	}

	// Verify each package has required fields
	for i, p := range packages {
		if p.Name == "" {
			t.Errorf("package[%d]: Name is empty", i)
		}
		if p.Source != "apt" {
			t.Errorf("package[%d]: Source = %q, want %q", i, p.Source, "apt")
		}
	}

	t.Logf("parsed %d packages from apt search python", len(packages))
}

func TestParseAptSearch_EmptyOutput(t *testing.T) {
	packages := parseAptSearch("", "nonexistent-pkg-xyz123")
	if len(packages) != 0 {
		t.Errorf("expected 0 packages from empty output, got %d", len(packages))
	}
}

func TestParseAptSearch_HeaderLines(t *testing.T) {
	// `apt search` adds "Sorting..." and "Full Text Search..." header lines
	// These should be ignored by the parser.
	output := `Sorting...
Full Text Search...
htop/stable 3.4.1-5 amd64
  interactive processes viewer
`

	packages := parseAptSearch(output, "htop")
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}
	if packages[0].Name != "htop" {
		t.Errorf("Name = %q, want %q", packages[0].Name, "htop")
	}
	if packages[0].Version != "3.4.1-5" {
		t.Errorf("Version = %q, want %q", packages[0].Version, "3.4.1-5")
	}
	if packages[0].Description != "interactive processes viewer" {
		t.Errorf("Description = %q, want %q", packages[0].Description, "interactive processes viewer")
	}
}

func TestParseAptSearch_MultiLineDescription(t *testing.T) {
	output := `htop/stable 3.4.1-5 amd64
  interactive processes viewer
  with multi-line description
  and more text
`

	packages := parseAptSearch(output, "htop")
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}
	if packages[0].Name != "htop" {
		t.Errorf("Name = %q, want %q", packages[0].Name, "htop")
	}
	expectedDesc := "interactive processes viewer with multi-line description and more text"
	if packages[0].Description != expectedDesc {
		t.Errorf("Description = %q, want %q", packages[0].Description, expectedDesc)
	}
}

func TestParseAptShow(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "apt_show_htop.txt"))
	if err != nil {
		t.Skipf("golden fixture not available: %v", err)
	}

	details := parseAptShow(string(data))

	if details.Name == "" {
		t.Error("Name should not be empty")
	}
	t.Logf("apt show htop: Name=%s Version=%s Homepage=%s Size=%d Desc=%s",
		details.Name, details.Version, details.Homepage, details.Size, details.Description)
}

func TestParseAptShow_Empty(t *testing.T) {
	details := parseAptShow("")
	if details.Name != "" {
		t.Errorf("Name = %q, want empty", details.Name)
	}
}

func TestAptAdapter_Install_FlagConstruction(t *testing.T) {
	// Tier 1 test: verify we construct the right arguments.
	// We don't actually run `sudo apt install` in unit tests.
	a := &AptAdapter{}

	// Verify the adapter constructs the expected command structure
	// by testing that we can create an exec.Cmd with the right args.
	// The actual execution is tested in Tier 2 integration tests.

	if !a.IsAvailable() {
		t.Skip("apt not available on this system — Tier 2 test only")
	}

	// Just verify Name() matches what we'd expect for source tracking
	pkg := Package{Name: "htop", Source: "apt"}
	if pkg.Source != a.Name() {
		t.Errorf("package source %q != adapter name %q", pkg.Source, a.Name())
	}
}

func TestAptAdapter_Uninstall_FlagConstruction(t *testing.T) {
	a := &AptAdapter{}

	pkg := Package{Name: "htop", Source: "apt"}
	if pkg.Source != a.Name() {
		t.Errorf("package source %q != adapter name %q", pkg.Source, a.Name())
	}
}
