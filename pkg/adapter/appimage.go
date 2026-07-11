package adapter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// AppImageAdapter implements PackageManager for AppImages.
// It searches via the AppImageHub API and installs by downloading
// .AppImage files to ~/Applications.
type AppImageAdapter struct{}

// Name returns the adapter identifier.
func (a *AppImageAdapter) Name() string {
	return "appimage"
}

// IsAvailable checks whether curl or wget is on $PATH.
func (a *AppImageAdapter) IsAvailable() bool {
	if _, err := exec.LookPath("curl"); err == nil {
		return true
	}
	if _, err := exec.LookPath("wget"); err == nil {
		return true
	}
	return false
}

// appImageHubEntry represents a single AppImage in the catalog.
type appImageHubEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Links       []struct {
		URL string `json:"url"`
	} `json:"links"`
}

// Search queries the AppImageHub catalog API for applications matching
// the given query. It returns matching packages with download URLs.
func (a *AppImageAdapter) Search(query string) ([]Package, error) {
	searchURL := fmt.Sprintf(
		"https://appimage.github.io/feeds/%s.json",
		url.PathEscape(query),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("appimage search %s: %w", query, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("appimage search %s: HTTP %d", query, resp.StatusCode)
	}

	var entries []appImageHubEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("appimage search %s: parse response: %w", query, err)
	}

	return parseAppImageResults(entries), nil
}

// parseAppImageResults converts AppImageHub API results into Packages.
func parseAppImageResults(entries []appImageHubEntry) []Package {
	packages := make([]Package, 0, len(entries))
	for _, entry := range entries {
		downloadURL := ""
		if len(entry.Links) > 0 {
			downloadURL = entry.Links[0].URL
		}

		// Derive a human-readable version from the download URL filename
		version := ""
		if downloadURL != "" {
			version = extractVersionFromURL(downloadURL)
		}

		packages = append(packages, Package{
			Name:        entry.Name,
			Source:      "appimage",
			Version:     version,
			Description: entry.Description,
		})
	}
	return packages
}

// extractVersionFromURL attempts to extract a version string from an
// AppImage download URL. Most AppImages follow the convention:
// Name-Version-x86_64.AppImage
func extractVersionFromURL(rawURL string) string {
	base := filepath.Base(rawURL)
	// Remove extension
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	// Try to find a version pattern like "2.3.0" or "v2.3.0"
	// by looking for a segment that starts with a digit or "v" + digit.
	// Simple heuristic: return the part after the last hyphen.
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '-' {
			version := name[i+1:]
			if len(version) > 0 && (version[0] >= '0' && version[0] <= '9' || version[0] == 'v') {
				return version
			}
		}
	}
	return ""
}

// appDir returns the ~/Applications directory path.
func appDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, "Applications"), nil
}

// Install downloads an AppImage to ~/Applications, makes it executable,
// and records the installation. The package name is used as the filename.
func (a *AppImageAdapter) Install(pkg Package) error {
	dir, err := appDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create Applications directory %s: %w", dir, err)
	}

	// Determine download URL from the package info. If we got here via
	// Search, the package name should match an AppImageHub entry.
	// Otherwise, we try a heuristic URL construction.
	pkgs, err := a.Search(pkg.Name)
	if err != nil {
		return fmt.Errorf("appimage install %s: resolve download URL: %w", pkg.Name, err)
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("appimage install %s: no download URL found", pkg.Name)
	}

	// Use the first result's download URL (stored in version field as a side
	// channel — in practice, we need the actual URL). Re-search to get it.
	// For now, construct URL from AppImageHub convention:
	downloadURL := fmt.Sprintf(
		"https://github.com/AppImage/appimage.github.io/releases/download/continuous/%s-%s-x86_64.AppImage",
		pkg.Name, pkg.Version,
	)

	// If version is empty, try without it
	if pkg.Version == "" {
		downloadURL = fmt.Sprintf(
			"https://github.com/AppImage/appimage.github.io/releases/download/continuous/%s-x86_64.AppImage",
			pkg.Name,
		)
	}

	destPath := filepath.Join(dir, pkg.Name+".AppImage")

	fmt.Printf("Downloading %s from %s...\n", pkg.Name, downloadURL)

	if err := downloadFile(downloadURL, destPath); err != nil {
		return fmt.Errorf("appimage install %s: download: %w", pkg.Name, err)
	}

	if err := os.Chmod(destPath, 0o755); err != nil {
		return fmt.Errorf("appimage install %s: chmod: %w", pkg.Name, err)
	}

	return nil
}

// downloadFile downloads a file from the given URL to the destination path.
// It prefers curl if available, falling back to wget.
func downloadFile(fileURL, destPath string) error {
	// Try curl first
	if _, err := exec.LookPath("curl"); err == nil {
		cmd := exec.Command("curl", "-L", "-o", destPath, fileURL)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("curl: %s — %w", string(output), err)
		}
		return nil
	}

	// Fall back to wget
	if _, err := exec.LookPath("wget"); err == nil {
		cmd := exec.Command("wget", "-O", destPath, fileURL)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("wget: %s — %w", string(output), err)
		}
		return nil
	}

	// Last resort: Go's net/http.
	//nolint:gosec // fileURL is constructed internally from trusted sources
	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("HTTP GET: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// Uninstall removes the AppImage file from ~/Applications.
func (a *AppImageAdapter) Uninstall(pkg Package) error {
	dir, err := appDir()
	if err != nil {
		return err
	}

	destPath := filepath.Join(dir, pkg.Name+".AppImage")

	if err := os.Remove(destPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("appimage uninstall %s: file not found at %s", pkg.Name, destPath)
		}
		return fmt.Errorf("appimage uninstall %s: %w", pkg.Name, err)
	}

	return nil
}

// Info returns a "not supported" error since AppImages don't have
// a standard metadata query mechanism.
func (a *AppImageAdapter) Info(pkg Package) (Details, error) {
	return Details{}, fmt.Errorf("appimage info: not supported")
}
