// Package config handles reading, writing, and default generation of the
// unipm configuration file (~/.unipm/config.yaml).
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DistroboxConfig defines a distrobox container that unipm can use as a
// package source.
type DistroboxConfig struct {
	// ContainerName is the name of the distrobox container (must match an
	// existing container).
	ContainerName string `yaml:"container_name"`

	// PackageManager is the package manager to use inside the container.
	// One of: apt, pacman, yay, dnf, zypper.
	PackageManager string `yaml:"package_manager"`
}

// Config holds all user-configurable settings for unipm.
type Config struct {
	// Distrobox maps container nicknames to their distrobox configuration.
	Distrobox map[string]DistroboxConfig `yaml:"distrobox"`

	// CacheTTL is the time-to-live in seconds for the tab-completion cache.
	CacheTTL int `yaml:"cache_ttl"`

	// SearchTimeout is the default per-adapter timeout in seconds for
	// search queries.
	SearchTimeout int `yaml:"search_timeout"`
}

// Defaults returns a Config populated with sensible defaults.
func Defaults() Config {
	return Config{
		Distrobox:     make(map[string]DistroboxConfig),
		CacheTTL:      86400, // 24 hours
		SearchTimeout: 10,    // seconds
	}
}

// dir returns the unipm config directory path (~/.unipm).
func dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, ".unipm"), nil
}

// EnsureDir creates the ~/.unipm directory with 0700 permissions if it
// does not already exist.
func EnsureDir() error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return fmt.Errorf("create config directory %s: %w", d, err)
	}
	return nil
}

// path returns the full path to the config file (~/.unipm/config.yaml).
func path() (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.yaml"), nil
}

// Load reads the config file from ~/.unipm/config.yaml. If the file does
// not exist, it returns a Config with default values. If the file exists
// but cannot be parsed, it returns an error.
func Load() (Config, error) {
	p, err := path()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaults(), nil
		}
		return Config{}, fmt.Errorf("read config file %s: %w", p, err)
	}

	cfg := Defaults()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file %s: %w", p, err)
	}

	// Ensure defaults for zero-value fields
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 86400
	}
	if cfg.SearchTimeout <= 0 {
		cfg.SearchTimeout = 10
	}
	if cfg.Distrobox == nil {
		cfg.Distrobox = make(map[string]DistroboxConfig)
	}

	return cfg, nil
}

// Save writes the config to ~/.unipm/config.yaml with 0600 permissions.
// It creates the config directory if needed.
func Save(cfg Config) error {
	if err := EnsureDir(); err != nil {
		return err
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	p, err := path()
	if err != nil {
		return err
	}

	if err := os.WriteFile(p, data, 0o600); err != nil {
		return fmt.Errorf("write config file %s: %w", p, err)
	}

	return nil
}
