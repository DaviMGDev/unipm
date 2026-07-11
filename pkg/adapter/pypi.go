package adapter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

// pypiResponse represents the JSON response from the PyPI JSON API
// (https://pypi.org/pypi/<name>/json).
type pypiResponse struct {
	Info pypiPackageInfo `json:"info"`
}

type pypiPackageInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Summary     string `json:"summary"`
	HomePage    string `json:"home_page"`
	License     string `json:"license"`
	Description string `json:"description"`
}

// PypiAdapter implements PackageManager for PyPI (Python Package Index).
type PypiAdapter struct{}

// Name returns the adapter identifier.
func (a *PypiAdapter) Name() string {
	return "pypi"
}

// IsAvailable checks whether the pip3 binary is on $PATH.
func (a *PypiAdapter) IsAvailable() bool {
	_, err := exec.LookPath("pip3")
	return err == nil
}

// Search queries the PyPI JSON API for a package by name. It tries an
// exact-match lookup against /pypi/<query>/json and returns a single
// Package if found. If the package does not exist, it returns an empty
// slice (no error — the package simply isn't on PyPI).
func (a *PypiAdapter) Search(query string) ([]Package, error) {
	searchURL := fmt.Sprintf("https://pypi.org/pypi/%s/json", url.PathEscape(query))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("pypi search %s: %w", query, err)
	}
	defer resp.Body.Close()

	// 404 means the package doesn't exist — not an error.
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pypi search %s: HTTP %d", query, resp.StatusCode)
	}

	var result pypiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("pypi search %s: parse response: %w", query, err)
	}

	if result.Info.Name == "" {
		return nil, nil
	}

	return []Package{{
		Name:        result.Info.Name,
		Source:      "pypi",
		Version:     result.Info.Version,
		Description: result.Info.Summary,
	}}, nil
}

// Install delegates to `pip3 install --user <name>`.
func (a *PypiAdapter) Install(pkg Package) error {
	cmd := exec.Command("pip3", "install", "--user", pkg.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip3 install --user %s: %s — %w", pkg.Name, string(output), err)
	}
	return nil
}

// Uninstall delegates to `pip3 uninstall -y <name>`.
func (a *PypiAdapter) Uninstall(pkg Package) error {
	cmd := exec.Command("pip3", "uninstall", "-y", pkg.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip3 uninstall %s: %s — %w", pkg.Name, string(output), err)
	}
	return nil
}

// Info delegates to `pip3 show <name>` and parses the output.
func (a *PypiAdapter) Info(pkg Package) (Details, error) {
	cmd := exec.Command("pip3", "show", pkg.Name)
	output, err := cmd.Output()
	if err != nil {
		return Details{}, fmt.Errorf("pip3 show %s: %w", pkg.Name, err)
	}

	return parsePipShow(string(output)), nil
}

// parsePipShow parses `pip3 show` output into a Details struct.
// The output format is RFC 2822-like header fields:
//
//	Name: requests
//	Version: 2.31.0
//	Summary: Python HTTP for Humans
//	Home-page: https://...
//	License: Apache 2.0
func parsePipShow(output string) Details {
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
		case "Name":
			d.Name = value
		case "Version":
			d.Version = value
		case "Summary":
			d.Description = value
		case "Home-page":
			d.Homepage = value
		case "License":
			d.License = value
		}
	}

	return d
}
