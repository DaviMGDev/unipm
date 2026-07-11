package adapter

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/DaviMGDev/unipm/pkg/config"
)

// DistroboxAdapter implements PackageManager for a distrobox container
// with a specified internal package manager (apt, pacman, yay, dnf, zypper).
//
// One DistroboxAdapter is created per configured container at startup.
// The adapter name is "distrobox-<container_name>".
type DistroboxAdapter struct {
	// ContainerName is the name of the distrobox container.
	ContainerName string

	// PackageManager is the package manager inside the container
	// (e.g., "pacman", "yay", "dnf", "apt", "zypper").
	PackageManager string

	// Nickname is the user-facing key from config.yaml
	// (e.g., "arch", "fedora").
	Nickname string
}

// NewDistroboxAdapter creates a DistroboxAdapter from a distrobox config entry.
func NewDistroboxAdapter(nickname string, cfg config.DistroboxConfig) *DistroboxAdapter {
	return &DistroboxAdapter{
		ContainerName:  cfg.ContainerName,
		PackageManager: cfg.PackageManager,
		Nickname:       nickname,
	}
}

// Name returns the adapter identifier: "distrobox-<container_name>".
func (a *DistroboxAdapter) Name() string {
	return "distrobox-" + a.ContainerName
}

// IsAvailable checks that distrobox is on $PATH and the container exists.
func (a *DistroboxAdapter) IsAvailable() bool {
	if _, err := exec.LookPath("distrobox"); err != nil {
		return false
	}

	// Check if the container exists via `distrobox list`
	cmd := exec.Command("distrobox", "list", "--no-color")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// distrobox list output contains container names
	return strings.Contains(string(output), a.ContainerName)
}

// Search runs the container's package manager search command via distrobox.
func (a *DistroboxAdapter) Search(query string) ([]Package, error) {
	searchArgs := a.buildSearchArgs(query)
	if len(searchArgs) == 0 {
		return nil, fmt.Errorf("distrobox %s: unsupported package manager %q", a.ContainerName, a.PackageManager)
	}

	args := append([]string{"enter", a.ContainerName, "--"}, searchArgs...)
	cmd := exec.Command("distrobox", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("distrobox-%s search %s: %w", a.ContainerName, query, err)
	}

	return parseDistroboxSearch(string(output), a.ContainerName, a.PackageManager), nil
}

// buildSearchArgs returns the package-manager-specific search arguments.
func (a *DistroboxAdapter) buildSearchArgs(query string) []string {
	switch a.PackageManager {
	case "apt":
		return []string{"apt", "search", query}
	case "pacman":
		return []string{"pacman", "-Ss", query}
	case "yay":
		return []string{"yay", "-Ss", query}
	case "dnf":
		return []string{"dnf", "search", query}
	case "zypper":
		return []string{"zypper", "search", query}
	default:
		return nil
	}
}

// parseDistroboxSearch parses search output from various package managers.
// Output format varies by PM; we use a best-effort line-based parser.
func parseDistroboxSearch(output, container, pm string) []Package {
	lines := strings.Split(output, "\n")
	var packages []Package
	source := "distrobox-" + container

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		name, version, description := parsePMLine(trimmed, pm)
		if name == "" {
			continue
		}

		packages = append(packages, Package{
			Name:        name,
			Source:      source,
			Version:     version,
			Description: description,
		})
	}

	return packages
}

// parsePMLine extracts package name, version, and description from a single
// output line of a package manager search command. The format varies:
//
//	apt:    "htop/stable 3.4.1 amd64\n  description"
//	pacman: "core/htop 3.4.1-1\n    Interactive process viewer"
//	yay:    "aur/htop 3.4.1-1 (42)\n    Interactive process viewer"
//	dnf:    "htop.x86_64  3.4.1-1.fc40  fedora"
func parsePMLine(line, pm string) (name, version, description string) {
	switch pm {
	case "apt":
		// apt: "name/distribution version arch"
		// Ignore description lines (they come on separate indented lines,
		// handled by the stream parser in apt.go)
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			return "", "", ""
		}
		if !strings.Contains(line, "/") {
			return "", "", ""
		}
		parts := strings.SplitN(line, " ", 2)
		namePart := strings.SplitN(parts[0], "/", 2)
		name = namePart[0]
		if len(parts) == 2 {
			verParts := strings.Fields(parts[1])
			if len(verParts) > 0 {
				version = verParts[0]
			}
		}

	case "pacman", "yay":
		// pacman: "repository/name version"
		// yay:    "repository/name version (votes)"
		parts := strings.Fields(line)
		if len(parts) < 2 {
			return "", "", ""
		}
		repoName := strings.SplitN(parts[0], "/", 2)
		if len(repoName) == 2 {
			name = repoName[1]
		} else {
			name = parts[0]
		}
		if len(parts) >= 2 {
			version = strings.TrimRight(parts[1], ")")
		}
		if len(parts) >= 4 {
			description = strings.Join(parts[3:], " ")
		}

	case "dnf":
		// dnf: "name.arch  version  repo"
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			archName := strings.SplitN(parts[0], ".", 2)
			name = archName[0]
		}
		if len(parts) >= 2 {
			version = parts[1]
		}
	}

	return name, version, description
}

// Install delegates to `distrobox enter <container> -- sudo <pm> <install-flags> <name>`.
func (a *DistroboxAdapter) Install(pkg Package) error {
	installArgs := a.buildInstallArgs(pkg.Name)
	if len(installArgs) == 0 {
		return fmt.Errorf("distrobox %s: unsupported package manager %q", a.ContainerName, a.PackageManager)
	}

	args := append([]string{"enter", a.ContainerName, "--"}, installArgs...)
	cmd := exec.Command("distrobox", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("distrobox-%s install %s: %s — %w", a.ContainerName, pkg.Name, string(output), err)
	}
	return nil
}

// buildInstallArgs returns the package-manager-specific install arguments.
func (a *DistroboxAdapter) buildInstallArgs(name string) []string {
	switch a.PackageManager {
	case "apt":
		return []string{"sudo", "apt", "install", "-y", name}
	case "pacman":
		return []string{"sudo", "pacman", "-S", "--noconfirm", name}
	case "yay":
		return []string{"yay", "-S", "--noconfirm", name}
	case "dnf":
		return []string{"sudo", "dnf", "install", "-y", name}
	case "zypper":
		return []string{"sudo", "zypper", "install", "-y", name}
	default:
		return nil
	}
}

// Uninstall delegates to `distrobox enter <container> -- sudo <pm> <remove-flags> <name>`.
func (a *DistroboxAdapter) Uninstall(pkg Package) error {
	uninstallArgs := a.buildUninstallArgs(pkg.Name)
	if len(uninstallArgs) == 0 {
		return fmt.Errorf("distrobox %s: unsupported package manager %q", a.ContainerName, a.PackageManager)
	}

	args := append([]string{"enter", a.ContainerName, "--"}, uninstallArgs...)
	cmd := exec.Command("distrobox", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("distrobox-%s uninstall %s: %s — %w", a.ContainerName, pkg.Name, string(output), err)
	}
	return nil
}

// buildUninstallArgs returns the package-manager-specific uninstall arguments.
func (a *DistroboxAdapter) buildUninstallArgs(name string) []string {
	switch a.PackageManager {
	case "apt":
		return []string{"sudo", "apt", "remove", "-y", name}
	case "pacman":
		return []string{"sudo", "pacman", "-R", "--noconfirm", name}
	case "yay":
		return []string{"yay", "-R", "--noconfirm", name}
	case "dnf":
		return []string{"sudo", "dnf", "remove", "-y", name}
	case "zypper":
		return []string{"sudo", "zypper", "remove", "-y", name}
	default:
		return nil
	}
}

// Info delegates to `distrobox enter <container> -- <pm> <info-flags> <name>`.
func (a *DistroboxAdapter) Info(pkg Package) (Details, error) {
	infoArgs := a.buildInfoArgs(pkg.Name)
	if len(infoArgs) == 0 {
		return Details{}, fmt.Errorf("distrobox %s: unsupported package manager %q", a.ContainerName, a.PackageManager)
	}

	args := append([]string{"enter", a.ContainerName, "--"}, infoArgs...)
	cmd := exec.Command("distrobox", args...)
	output, err := cmd.Output()
	if err != nil {
		return Details{}, fmt.Errorf("distrobox-%s info %s: %w", a.ContainerName, pkg.Name, err)
	}

	return parseDistroboxInfo(string(output), a.PackageManager), nil
}

// buildInfoArgs returns the package-manager-specific info arguments.
func (a *DistroboxAdapter) buildInfoArgs(name string) []string {
	switch a.PackageManager {
	case "apt":
		return []string{"apt", "show", name}
	case "pacman":
		return []string{"pacman", "-Qi", name}
	case "yay":
		return []string{"yay", "-Qi", name}
	case "dnf":
		return []string{"dnf", "info", name}
	case "zypper":
		return []string{"zypper", "info", name}
	default:
		return nil
	}
}

// parseDistroboxInfo parses package info output from various PMs.
func parseDistroboxInfo(output, pm string) Details {
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

		switch {
		case key == "Package" || key == "Name":
			d.Name = value
		case key == "Version":
			d.Version = value
		case key == "Homepage" || key == "URL":
			d.Homepage = value
		case key == "Description" || key == "Summary":
			if d.Description == "" {
				d.Description = value
			}
		case key == "License":
			d.License = value
		}
	}

	return d
}
