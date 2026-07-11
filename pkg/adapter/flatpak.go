package adapter

import (
	"fmt"
	"os/exec"
	"strings"
)

// FlatpakAdapter implements PackageManager for Flatpak (flatpak CLI).
type FlatpakAdapter struct{}

// Name returns the adapter identifier.
func (a *FlatpakAdapter) Name() string {
	return "flatpak"
}

// IsAvailable checks whether the flatpak binary is on $PATH.
func (a *FlatpakAdapter) IsAvailable() bool {
	_, err := exec.LookPath("flatpak")
	return err == nil
}

// Search queries Flatpak for applications matching the given query.
// It runs `flatpak search <query>` and parses the tabular output.
//
// flatpak search output format (columns separated by tabs):
//
//	Name        Description                                      Application ID              Version  Branch Remotes
//	Flatseal    Manage Flatpak permissions                       com.github.tchx84.Flatseal   2.3.0    stable flathub
func (a *FlatpakAdapter) Search(query string) ([]Package, error) {
	cmd := exec.Command("flatpak", "search", query)
	output, err := cmd.Output()
	if err != nil {
		// flatpak search returns non-zero when no results found — not an error.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("flatpak search %s: %w", query, err)
	}

	return parseFlatpakSearch(string(output)), nil
}

// parseFlatpakSearch parses `flatpak search` output into Packages.
// The first line is a header row; all following tab-separated lines
// contain: Name, Description, Application ID, Version, Branch, Remotes.
func parseFlatpakSearch(output string) []Package {
	lines := strings.Split(output, "\n")
	var packages []Package

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip header row (first non-empty line)
		if i == 0 {
			continue
		}

		// Columns: Name \t Description \t Application ID \t Version \t Branch \t Remotes
		fields := strings.Split(trimmed, "\t")
		if len(fields) < 3 {
			continue
		}

		name := strings.TrimSpace(fields[0])
		description := ""
		if len(fields) > 1 {
			description = strings.TrimSpace(fields[1])
		}
		appID := ""
		if len(fields) > 2 {
			appID = strings.TrimSpace(fields[2])
		}
		version := ""
		if len(fields) > 3 {
			version = strings.TrimSpace(fields[3])
		}

		// Use Application ID as the canonical name since that's what
		// flatpak install/uninstall expects.
		if appID == "" {
			continue
		}

		packages = append(packages, Package{
			Name:        appID,
			Source:      "flatpak",
			Version:     version,
			Description: description,
		})
		_ = name // human-readable name is informational only
	}

	return packages
}

// Install delegates to `flatpak install -y flathub <name>`.
func (a *FlatpakAdapter) Install(pkg Package) error {
	cmd := exec.Command("flatpak", "install", "-y", "flathub", pkg.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("flatpak install %s: %s — %w", pkg.Name, string(output), err)
	}
	return nil
}

// Uninstall delegates to `flatpak uninstall -y <name>`.
func (a *FlatpakAdapter) Uninstall(pkg Package) error {
	cmd := exec.Command("flatpak", "uninstall", "-y", pkg.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("flatpak uninstall %s: %s — %w", pkg.Name, string(output), err)
	}
	return nil
}

// Info delegates to `flatpak info <name>` and parses the output.
func (a *FlatpakAdapter) Info(pkg Package) (Details, error) {
	cmd := exec.Command("flatpak", "info", pkg.Name)
	output, err := cmd.Output()
	if err != nil {
		return Details{}, fmt.Errorf("flatpak info %s: %w", pkg.Name, err)
	}

	return parseFlatpakInfo(string(output)), nil
}

// parseFlatpakInfo parses `flatpak info` output.
// Format is key-value with ": " separator:
//
//	     ID: com.github.tchx84.Flatseal
//	    Ref: app/com.github.tchx84.Flatseal/x86_64/stable
//	   Arch: x86_64
//	 Branch: stable
//	Version: 2.3.0
func parseFlatpakInfo(output string) Details {
	d := Details{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		parts := strings.SplitN(trimmed, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "ID":
			d.Name = value
		case "Version":
			d.Version = value
		}
	}

	return d
}
