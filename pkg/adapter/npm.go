package adapter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"time"
)

// npmSearchResponse represents the JSON response from the npm registry
// search API (https://registry.npmjs.org/-/v1/search).
type npmSearchResponse struct {
	Objects []npmSearchObject `json:"objects"`
	Total   int               `json:"total"`
	Time    string            `json:"time"`
}

type npmSearchObject struct {
	Package npmPackageInfo `json:"package"`
}

type npmPackageInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// NpmAdapter implements PackageManager for npm (Node Package Manager).
type NpmAdapter struct{}

// Name returns the adapter identifier.
func (a *NpmAdapter) Name() string {
	return "npm"
}

// IsAvailable checks whether the npm binary is on $PATH.
func (a *NpmAdapter) IsAvailable() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

// Search queries the npm registry API for packages matching the given
// query. It uses the public npm search endpoint.
func (a *NpmAdapter) Search(query string) ([]Package, error) {
	searchURL := fmt.Sprintf("https://registry.npmjs.org/-/v1/search?text=%s&size=20", url.QueryEscape(query))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("npm search %s: %w", query, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("npm search %s: HTTP %d", query, resp.StatusCode)
	}

	var result npmSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("npm search %s: parse response: %w", query, err)
	}

	return parseNpmSearch(&result), nil
}

// parseNpmSearch converts an npmSearchResponse into a slice of Packages.
func parseNpmSearch(resp *npmSearchResponse) []Package {
	packages := make([]Package, 0, len(resp.Objects))
	for _, obj := range resp.Objects {
		packages = append(packages, Package{
			Name:        obj.Package.Name,
			Source:      "npm",
			Version:     obj.Package.Version,
			Description: obj.Package.Description,
		})
	}
	return packages
}

// Install delegates to `npm install -g <name>`.
func (a *NpmAdapter) Install(pkg Package) error {
	cmd := exec.Command("npm", "install", "-g", pkg.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm install -g %s: %s — %w", pkg.Name, string(output), err)
	}
	return nil
}

// Uninstall delegates to `npm uninstall -g <name>`.
func (a *NpmAdapter) Uninstall(pkg Package) error {
	cmd := exec.Command("npm", "uninstall", "-g", pkg.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm uninstall -g %s: %s — %w", pkg.Name, string(output), err)
	}
	return nil
}

// Info delegates to `npm info <name> --json` and parses the output.
func (a *NpmAdapter) Info(pkg Package) (Details, error) {
	cmd := exec.Command("npm", "info", pkg.Name, "--json")
	output, err := cmd.Output()
	if err != nil {
		return Details{}, fmt.Errorf("npm info %s: %w", pkg.Name, err)
	}

	return parseNpmInfo(string(output)), nil
}

// npmInfoResponse represents the JSON output of `npm info <name> --json`.
type npmInfoResponse struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
	License     string `json:"license"`
}

// parseNpmInfo parses `npm info --json` output.
func parseNpmInfo(output string) Details {
	var info npmInfoResponse
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		return Details{}
	}
	return Details{
		Name:        info.Name,
		Version:     info.Version,
		Description: info.Description,
		Homepage:    info.Homepage,
		License:     info.License,
	}
}
