package cache

import (
	"os"
	"testing"
	"time"
)

func setupTempHome(t *testing.T) func() {
	t.Helper()
	tmp := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	return func() {
		os.Setenv("HOME", oldHome)
	}
}

func TestLoad_EmptyCache(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	e, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if e.Packages == nil {
		t.Error("Packages should not be nil")
	}
	if len(e.Packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(e.Packages))
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	e := Entry{
		Packages: map[string]time.Time{
			"htop":    time.Now().UTC(),
			"ripgrep": time.Now().UTC(),
		},
	}

	if err := Save(e); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(loaded.Packages))
	}
	if _, ok := loaded.Packages["htop"]; !ok {
		t.Error("htop not found in loaded cache")
	}
	if _, ok := loaded.Packages["ripgrep"]; !ok {
		t.Error("ripgrep not found in loaded cache")
	}
}

func TestAddRecords(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	if err := AddRecords([]string{"htop", "ripgrep", "httpie"}); err != nil {
		t.Fatalf("AddRecords() error = %v", err)
	}

	e, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(e.Packages) != 3 {
		t.Errorf("expected 3 packages, got %d", len(e.Packages))
	}
}

func TestMatching_ExactPrefix(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	AddRecords([]string{"htop", "httpie", "ripgrep", "httrack"})

	matches := Matching("ht", 24*time.Hour)

	if len(matches) != 3 {
		t.Errorf("expected 3 matches for 'ht', got %d: %v", len(matches), matches)
	}
}

func TestMatching_Expired(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	// Add entries with old timestamp by manipulating the map directly
	e := Entry{
		Packages: map[string]time.Time{
			"htop":    time.Now().UTC().Add(-48 * time.Hour),
			"ripgrep": time.Now().UTC(),
		},
	}
	Save(e)

	matches := Matching("", 24*time.Hour)

	// htop should be expired, ripgrep should not
	if len(matches) != 1 {
		t.Errorf("expected 1 non-expired match, got %d: %v", len(matches), matches)
	}
	if len(matches) > 0 && matches[0] != "ripgrep" {
		t.Errorf("expected 'ripgrep', got %q", matches[0])
	}
}

func TestMatching_NoMatch(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	AddRecords([]string{"htop", "ripgrep"})

	matches := Matching("zzz", 24*time.Hour)

	if len(matches) != 0 {
		t.Errorf("expected 0 matches for 'zzz', got %d", len(matches))
	}
}

func TestMatching_EmptyPrefix(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	AddRecords([]string{"htop", "ripgrep"})

	matches := Matching("", 24*time.Hour)

	if len(matches) != 2 {
		t.Errorf("expected 2 matches for empty prefix, got %d", len(matches))
	}
}
