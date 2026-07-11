package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTempHome sets HOME to t.TempDir() and ensures a clean .unipm directory.
func setupTempHome(t *testing.T) func() {
	t.Helper()
	tmp := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	return func() {
		os.Setenv("HOME", oldHome)
	}
}

// writeStateFile writes a StateFile to the state path for testing.
func writeStateFile(t *testing.T, sf StateFile) {
	t.Helper()
	home := os.Getenv("HOME")
	dir := filepath.Join(home, dirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create state dir: %v", err)
	}
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		t.Fatalf("marshal state: %v", err)
	}
	p := filepath.Join(dir, fileName)
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatalf("write state: %v", err)
	}
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func TestLoad_FileNotExists_ReturnsEmpty(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	sf, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if sf.Version != CurrentVersion {
		t.Errorf("Version = %d, want %d", sf.Version, CurrentVersion)
	}
	if len(sf.Packages) != 0 {
		t.Errorf("Packages length = %d, want 0", len(sf.Packages))
	}
}

func TestLoad_UnknownVersion_ReturnsError(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	writeStateFile(t, StateFile{
		Version:  999,
		Packages: []StateRecord{},
	})

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for unknown version")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	sf := StateFile{
		Version: CurrentVersion,
		Packages: []StateRecord{
			{
				Name:        "htop",
				Source:      "apt",
				Version:     "3.3.0",
				InstalledAt: now(),
			},
			{
				Name:        "httpie",
				Source:      "pypi",
				Version:     "3.2.1",
				InstalledAt: now(),
			},
		},
	}

	if err := Save(sf); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Version != CurrentVersion {
		t.Errorf("Version = %d, want %d", loaded.Version, CurrentVersion)
	}
	if len(loaded.Packages) != 2 {
		t.Fatalf("Packages length = %d, want 2", len(loaded.Packages))
	}
	if loaded.Packages[0].Name != "htop" {
		t.Errorf("Packages[0].Name = %q, want %q", loaded.Packages[0].Name, "htop")
	}
	if loaded.Packages[0].Source != "apt" {
		t.Errorf("Packages[0].Source = %q, want %q", loaded.Packages[0].Source, "apt")
	}
	if loaded.Packages[1].Name != "httpie" {
		t.Errorf("Packages[1].Name = %q, want %q", loaded.Packages[1].Name, "httpie")
	}
}

func TestAdd_AppendsRecord(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	record := StateRecord{
		Name:        "htop",
		Source:      "apt",
		Version:     "3.3.0",
		InstalledAt: now(),
	}

	if err := Add(record); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Verify it's in the state file
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Packages) != 1 {
		t.Fatalf("Packages length = %d, want 1", len(loaded.Packages))
	}
	if loaded.Packages[0].Name != "htop" {
		t.Errorf("Name = %q, want %q", loaded.Packages[0].Name, "htop")
	}
}

func TestAdd_DuplicateName_ReturnsError(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	record := StateRecord{
		Name:        "htop",
		Source:      "apt",
		Version:     "3.3.0",
		InstalledAt: now(),
	}

	if err := Add(record); err != nil {
		t.Fatalf("first Add() error = %v", err)
	}

	// Adding same name again should fail
	err := Add(record)
	if err == nil {
		t.Error("Add() should return error for duplicate package name")
	}
}

func TestRemove_DeletesRecord(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	writeStateFile(t, StateFile{
		Version: CurrentVersion,
		Packages: []StateRecord{
			{Name: "htop", Source: "apt", Version: "3.3.0", InstalledAt: now()},
			{Name: "httpie", Source: "pypi", Version: "3.2.1", InstalledAt: now()},
		},
	})

	if err := Remove("htop"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Packages) != 1 {
		t.Fatalf("Packages length = %d, want 1", len(loaded.Packages))
	}
	if loaded.Packages[0].Name != "httpie" {
		t.Errorf("remaining package Name = %q, want %q", loaded.Packages[0].Name, "httpie")
	}
}

func TestRemove_NotFound_ReturnsError(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	writeStateFile(t, StateFile{
		Version:  CurrentVersion,
		Packages: []StateRecord{},
	})

	err := Remove("nonexistent")
	if err == nil {
		t.Error("Remove() should return error for untracked package")
	}
}

func TestGet_FindsRecord(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	writeStateFile(t, StateFile{
		Version: CurrentVersion,
		Packages: []StateRecord{
			{Name: "htop", Source: "apt", Version: "3.3.0", InstalledAt: now()},
		},
	})

	rec, err := Get("htop")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if rec.Name != "htop" {
		t.Errorf("Name = %q, want %q", rec.Name, "htop")
	}
	if rec.Source != "apt" {
		t.Errorf("Source = %q, want %q", rec.Source, "apt")
	}
}

func TestGet_NotFound_ReturnsError(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	writeStateFile(t, StateFile{
		Version:  CurrentVersion,
		Packages: []StateRecord{},
	})

	_, err := Get("nonexistent")
	if err == nil {
		t.Error("Get() should return error for untracked package")
	}
}

func TestList_ReturnsAllRecords(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	writeStateFile(t, StateFile{
		Version: CurrentVersion,
		Packages: []StateRecord{
			{Name: "htop", Source: "apt", Version: "3.3.0", InstalledAt: now()},
			{Name: "httpie", Source: "pypi", Version: "3.2.1", InstalledAt: now()},
			{Name: "ripgrep", Source: "brew", Version: "14.1.0", InstalledAt: now()},
		},
	})

	records, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("List() returned %d records, want 3", len(records))
	}
}

func TestList_EmptyState_ReturnsEmptySlice(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	records, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(records) != 0 {
		t.Errorf("List() returned %d records, want 0", len(records))
	}
}

func TestUpdateVersion_RefreshesRecord(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	oldTime := "2026-01-01T00:00:00Z"
	writeStateFile(t, StateFile{
		Version: CurrentVersion,
		Packages: []StateRecord{
			{Name: "htop", Source: "apt", Version: "3.2.0", InstalledAt: oldTime},
		},
	})

	newTime := now()
	if err := UpdateVersion("htop", "3.3.0", newTime); err != nil {
		t.Fatalf("UpdateVersion() error = %v", err)
	}

	rec, err := Get("htop")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if rec.Version != "3.3.0" {
		t.Errorf("Version = %q, want %q", rec.Version, "3.3.0")
	}
	if rec.InstalledAt != newTime {
		t.Errorf("InstalledAt = %q, want %q", rec.InstalledAt, newTime)
	}
}

func TestUpdateVersion_NotFound_ReturnsError(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	err := UpdateVersion("nonexistent", "1.0", now())
	if err == nil {
		t.Error("UpdateVersion() should return error for untracked package")
	}
}

func TestSave_AtomicWrite_NoCorruption(t *testing.T) {
	cleanup := setupTempHome(t)
	defer cleanup()

	sf := StateFile{
		Version: CurrentVersion,
		Packages: []StateRecord{
			{Name: "htop", Source: "apt", Version: "3.3.0", InstalledAt: now()},
		},
	}

	// Write initial state
	if err := Save(sf); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify no temp file is left behind
	home := os.Getenv("HOME")
	tmpPath := filepath.Join(home, dirName, fileName+".tmp")
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("temp file %s should not exist after successful save", tmpPath)
	}

	// Verify the actual state file is intact
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Packages) != 1 || loaded.Packages[0].Name != "htop" {
		t.Error("state file corrupted after atomic write — content mismatch")
	}
}
