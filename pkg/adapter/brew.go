package adapter

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// BrewAdapter implements PackageManager for Homebrew (Linuxbrew / macOS).
type BrewAdapter struct{}

// Name returns the adapter identifier.
func (a *BrewAdapter) Name() string {
	return "brew"
}

// IsAvailable checks whether the brew binary is on $PATH.
func (a *BrewAdapter) IsAvailable() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

// Search queries Homebrew for formulae/casks matching the given query.
// It runs `brew search <query>` and parses the output.
//
// brew search outputs one package per line. Lines containing "==>" are
// section headers (Formulae / Casks) and are skipped.
func (a *BrewAdapter) Search(query string) ([]Package, error) {
	cmd := exec.Command("brew", "search", query)
	output, err := cmd.Output()
	if err != nil {
		// brew search exits 0 even with no results, but if brew is broken,
		// return the error.
		return nil, fmt.Errorf("brew search %s: %w", query, err)
	}

	return parseBrewSearch(string(output)), nil
}

// parseBrewSearch parses `brew search` output into Packages.
// Output format:
//
//	==> Formulae
//	htop
//	htop-osx
//	==> Casks
//	htop-app
func parseBrewSearch(output string) []Package {
	lines := strings.Split(output, "\n")
	var packages []Package

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip section headers
		if strings.HasPrefix(trimmed, "==>") {
			continue
		}

		// brew search sometimes returns lines with descriptions when
		// searching with `--desc`, but the default search just lists names.
		packages = append(packages, Package{
			Name:    trimmed,
			Source:  "brew",
			Version: "", // brew search doesn't include versions
		})
	}

	return packages
}

// Install delegates to `brew install <name>`.
func (a *BrewAdapter) Install(pkg Package) error {
	cmd := exec.Command("brew", "install", pkg.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("brew install %s: %s — %w", pkg.Name, string(output), err)
	}
	return nil
}

// Uninstall delegates to `brew uninstall <name>`.
func (a *BrewAdapter) Uninstall(pkg Package) error {
	cmd := exec.Command("brew", "uninstall", pkg.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("brew uninstall %s: %s — %w", pkg.Name, string(output), err)
	}
	return nil
}

// Info delegates to `brew info --json <name>` and parses the JSON output.
func (a *BrewAdapter) Info(pkg Package) (Details, error) {
	cmd := exec.Command("brew", "info", "--json", pkg.Name)
	output, err := cmd.Output()
	if err != nil {
		return Details{}, fmt.Errorf("brew info %s: %w", pkg.Name, err)
	}

	return parseBrewInfo(string(output)), nil
}

// brewInfoResponse represents the JSON array output of `brew info --json`.
// brew info --json returns an array even for a single package.
type brewInfoResponse struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"desc"`
	Homepage    string `json:"homepage"`
	License     string `json:"license"`
	Versions    struct {
		Stable string `json:"stable"`
	} `json:"versions"`
}

// parseBrewInfo parses `brew info --json` output.
// Returns Details with available fields filled; empty Details if parsing fails.
func parseBrewInfo(output string) Details {
	// brew info --json returns an array: [{...}]
	var infos []brewInfoResponse
	if err := json.Unmarshal([]byte(output), &infos); err != nil {
		return Details{}
	}

	if len(infos) == 0 {
		return Details{}
	}

	info := infos[0]

	version := info.Versions.Stable
	if version == "" {
		// For casks, version might be elsewhere; try full_name
		version = "unknown"
	}

	return Details{
		Name:        info.Name,
		Version:     version,
		Description: info.Description,
		Homepage:    info.Homepage,
		License:     info.License,
	}
}
