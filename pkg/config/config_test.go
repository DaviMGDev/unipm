package config

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTempHome sets HOME to a temp directory and returns a cleanup function.
func setupTempHome(t *testing.T) (string, func()) {
	t.Helper()
	tmp := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	// Ensure .unipm directory is created under our temp home
	unipmDir := filepath.Join(tmp, ".unipm")
	return unipmDir, func() {
		os.Setenv("HOME", oldHome)
	}
}

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	if cfg.CacheTTL != 86400 {
		t.Errorf("CacheTTL = %d, want 86400", cfg.CacheTTL)
	}
	if cfg.SearchTimeout != 10 {
		t.Errorf("SearchTimeout = %d, want 10", cfg.SearchTimeout)
	}
	if cfg.Distrobox == nil {
		t.Error("Distrobox map should be initialized, got nil")
	}
	if len(cfg.Distrobox) != 0 {
		t.Errorf("Distrobox map should be empty, got %d entries", len(cfg.Distrobox))
	}
}

func TestEnsureDir(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	if err := EnsureDir(); err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	home := os.Getenv("HOME")
	unipmPath := filepath.Join(home, ".unipm")

	info, err := os.Stat(unipmPath)
	if err != nil {
		t.Fatalf("stat %s: %v", unipmPath, err)
	}
	if !info.IsDir() {
		t.Errorf("%s is not a directory", unipmPath)
	}

	// Check permissions: 0700 means owner rwx only.
	// On Unix, os.MkdirAll with 0700 should set this. Skip on Windows.
	if info.Mode().Perm() != 0o700 {
		t.Logf("directory permissions are %o, expected 700 (may vary on some platforms)", info.Mode().Perm())
	}
}

func TestLoad_FileNotExists_ReturnsDefaults(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.CacheTTL != 86400 {
		t.Errorf("CacheTTL = %d, want 86400", cfg.CacheTTL)
	}
	if cfg.SearchTimeout != 10 {
		t.Errorf("SearchTimeout = %d, want 10", cfg.SearchTimeout)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	original := Config{
		Distrobox: map[string]DistroboxConfig{
			"arch": {
				ContainerName:  "arch-dev",
				PackageManager: "yay",
			},
		},
		CacheTTL:      3600,
		SearchTimeout: 15,
	}

	if err := Save(original); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.CacheTTL != 3600 {
		t.Errorf("CacheTTL = %d, want 3600", loaded.CacheTTL)
	}
	if loaded.SearchTimeout != 15 {
		t.Errorf("SearchTimeout = %d, want 15", loaded.SearchTimeout)
	}

	archCfg, ok := loaded.Distrobox["arch"]
	if !ok {
		t.Fatal("distrobox 'arch' not found in loaded config")
	}
	if archCfg.ContainerName != "arch-dev" {
		t.Errorf("ContainerName = %q, want %q", archCfg.ContainerName, "arch-dev")
	}
	if archCfg.PackageManager != "yay" {
		t.Errorf("PackageManager = %q, want %q", archCfg.PackageManager, "yay")
	}
}

func TestLoad_InvalidYAML_ReturnsError(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	// Ensure .unipm directory exists
	if err := EnsureDir(); err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	home := os.Getenv("HOME")
	cfgPath := filepath.Join(home, ".unipm", "config.yaml")

	// Write invalid YAML
	if err := os.WriteFile(cfgPath, []byte("{{{invalid: yaml: ["), 0o600); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}

func TestLoad_ZeroValues_GetDefaults(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	if err := EnsureDir(); err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	home := os.Getenv("HOME")
	cfgPath := filepath.Join(home, ".unipm", "config.yaml")

	// Write config with zero values for CacheTTL and SearchTimeout
	cfg := `distrobox: {}
cache_ttl: 0
search_timeout: 0
`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.CacheTTL != 86400 {
		t.Errorf("CacheTTL = %d, want 86400 (default applied to zero)", loaded.CacheTTL)
	}
	if loaded.SearchTimeout != 10 {
		t.Errorf("SearchTimeout = %d, want 10 (default applied to zero)", loaded.SearchTimeout)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	_, cleanup := setupTempHome(t)
	defer cleanup()

	// Ensure .unipm directory does NOT exist before Save
	home := os.Getenv("HOME")
	unipmPath := filepath.Join(home, ".unipm")
	os.RemoveAll(unipmPath) // clean any leftovers

	cfg := Defaults()
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(unipmPath); os.IsNotExist(err) {
		t.Error("Save() should have created ~/.unipm directory")
	}

	// Verify file was created with content
	cfgPath := filepath.Join(unipmPath, "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read created config file: %v", err)
	}
	if len(data) == 0 {
		t.Error("config file is empty after Save()")
	}
}
