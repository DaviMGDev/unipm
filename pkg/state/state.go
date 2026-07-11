// Package state manages the unipm state file (~/.unipm/state.json) that
// tracks every package installed through unipm. All writes are atomic
// (temp file + rename) to prevent corruption.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// StateRecord represents a single package installed through unipm.
type StateRecord struct {
	// Name is the package identifier (e.g., "htop", "httpie").
	Name string `json:"name"`

	// Source identifies the backend that installed this package
	// (e.g., "apt", "npm", "pypi").
	Source string `json:"source"`

	// Version is the installed version at install/update time.
	Version string `json:"version"`

	// InstalledAt is an RFC 3339 UTC timestamp of when the package was
	// installed or last updated.
	InstalledAt string `json:"installed_at"`
}

// StateFile is the on-disk representation of the state file.
type StateFile struct {
	// Version is the state file schema version. unipm rejects unknown
	// versions with a migration instruction.
	Version int `json:"version"`

	// Packages is the list of tracked package records.
	Packages []StateRecord `json:"packages"`
}

const (
	// CurrentVersion is the state file schema version that this version
	// of unipm can read and write.
	CurrentVersion = 1

	// dirName is the unipm config directory name under $HOME.
	dirName = ".unipm"

	// fileName is the state file name within the config directory.
	fileName = "state.json"
)

// dir returns the unipm config directory path (~/.unipm).
func dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, dirName), nil
}

// path returns the full path to the state file (~/.unipm/state.json).
func path() (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, fileName), nil
}

// ensureDir creates the ~/.unipm directory with 0700 permissions if it
// does not already exist.
func ensureDir() error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return fmt.Errorf("create state directory %s: %w", d, err)
	}
	return nil
}

// Load reads the state file from ~/.unipm/state.json. If the file does
// not exist, it returns an empty StateFile with the current version.
// If the file exists but has an unknown version, it returns an error.
func Load() (StateFile, error) {
	p, err := path()
	if err != nil {
		return StateFile{}, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return StateFile{
				Version:  CurrentVersion,
				Packages: []StateRecord{},
			}, nil
		}
		return StateFile{}, fmt.Errorf("read state file %s: %w", p, err)
	}

	var sf StateFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return StateFile{}, fmt.Errorf("parse state file %s: %w", p, err)
	}

	if sf.Version != CurrentVersion {
		return StateFile{}, fmt.Errorf(
			"unsupported state file version %d (expected %d). "+
				"Run 'unipm migrate' to upgrade your state file.",
			sf.Version, CurrentVersion,
		)
	}

	if sf.Packages == nil {
		sf.Packages = []StateRecord{}
	}

	return sf, nil
}

// Save writes the state file atomically to ~/.unipm/state.json with 0600
// permissions. It writes to a temp file first, then renames to prevent
// corruption on crash or power loss.
func Save(sf StateFile) error {
	sf.Version = CurrentVersion

	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state file: %w", err)
	}

	if err := ensureDir(); err != nil {
		return err
	}

	p, err := path()
	if err != nil {
		return err
	}

	// Atomic write: write to temp file then rename.
	tmpPath := p + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write temp state file %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, p); err != nil {
		// Clean up temp file on failure
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp state file to %s: %w", p, err)
	}

	return nil
}

// Add appends a new StateRecord to the state file. Returns an error if a
// package with the same name already exists.
func Add(record StateRecord) error {
	sf, err := Load()
	if err != nil {
		return fmt.Errorf("add %s: %w", record.Name, err)
	}

	// Check for duplicate names
	for _, p := range sf.Packages {
		if p.Name == record.Name {
			return fmt.Errorf("package %q is already tracked (installed from %s)", record.Name, p.Source)
		}
	}

	sf.Packages = append(sf.Packages, record)

	if err := Save(sf); err != nil {
		return fmt.Errorf("add %s: %w", record.Name, err)
	}

	return nil
}

// Remove deletes a StateRecord by name from the state file. Returns an
// error if the package is not found.
func Remove(name string) error {
	sf, err := Load()
	if err != nil {
		return fmt.Errorf("remove %s: %w", name, err)
	}

	found := false
	var filtered []StateRecord
	for _, p := range sf.Packages {
		if p.Name == name {
			found = true
			continue
		}
		filtered = append(filtered, p)
	}

	if !found {
		return fmt.Errorf("package %q was not installed via unipm", name)
	}

	sf.Packages = filtered

	if err := Save(sf); err != nil {
		return fmt.Errorf("remove %s: %w", name, err)
	}

	return nil
}

// Get finds a StateRecord by name. Returns an error if the package is not
// found.
func Get(name string) (StateRecord, error) {
	sf, err := Load()
	if err != nil {
		return StateRecord{}, fmt.Errorf("get %s: %w", name, err)
	}

	for _, p := range sf.Packages {
		if p.Name == name {
			return p, nil
		}
	}

	return StateRecord{}, fmt.Errorf("package %q was not installed via unipm", name)
}

// List returns all StateRecords from the state file.
func List() ([]StateRecord, error) {
	sf, err := Load()
	if err != nil {
		return nil, fmt.Errorf("list packages: %w", err)
	}
	return sf.Packages, nil
}

// UpdateVersion refreshes the version and installed_at fields for a
// tracked package. Returns an error if the package is not found.
func UpdateVersion(name, version, installedAt string) error {
	sf, err := Load()
	if err != nil {
		return fmt.Errorf("update %s: %w", name, err)
	}

	found := false
	for i, p := range sf.Packages {
		if p.Name == name {
			sf.Packages[i].Version = version
			sf.Packages[i].InstalledAt = installedAt
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("package %q was not installed via unipm", name)
	}

	if err := Save(sf); err != nil {
		return fmt.Errorf("update %s: %w", name, err)
	}

	return nil
}
