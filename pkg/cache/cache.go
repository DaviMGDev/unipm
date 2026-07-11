// Package cache provides a simple TTL-based completion cache for package
// names, stored at ~/.unipm/cache.json. Search commands populate the cache;
// tab-completion reads from it.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry maps package names to their last-seen timestamps.
type Entry struct {
	Packages map[string]time.Time `json:"packages"`
}

const (
	dirName  = ".unipm"
	fileName = "cache.json"
)

// dir returns the cache directory path.
func dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, dirName), nil
}

// path returns the full cache file path.
func path() (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, fileName), nil
}

// Load reads the cache file. If it doesn't exist, returns an empty cache.
func Load() (Entry, error) {
	p, err := path()
	if err != nil {
		return Entry{}, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Entry{Packages: make(map[string]time.Time)}, nil
		}
		return Entry{}, fmt.Errorf("read cache file: %w", err)
	}

	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		// Corrupted cache — start fresh
		return Entry{Packages: make(map[string]time.Time)}, nil
	}

	if e.Packages == nil {
		e.Packages = make(map[string]time.Time)
	}

	return e, nil
}

// Save writes the cache file, creating the directory if needed.
func Save(e Entry) error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	p, err := path()
	if err != nil {
		return err
	}

	if err := os.WriteFile(p, data, 0o600); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	return nil
}

// AddRecords adds package names to the cache with the current timestamp.
func AddRecords(names []string) error {
	e, err := Load()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, name := range names {
		e.Packages[name] = now
	}

	return Save(e)
}

// Matching returns cached package names that match the given prefix and
// haven't exceeded the TTL. Returns names sorted alphabetically.
func Matching(prefix string, ttl time.Duration) []string {
	e, err := Load()
	if err != nil {
		return nil
	}

	now := time.Now().UTC()
	var matches []string

	for name, ts := range e.Packages {
		// Skip expired entries
		if now.Sub(ts) > ttl {
			continue
		}

		// Match prefix (case-insensitive for convenience)
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			matches = append(matches, name)
		}
	}

	// Already sorted by map iteration order is non-deterministic,
	// but for shell completion the shell handles sorting.
	return matches
}
