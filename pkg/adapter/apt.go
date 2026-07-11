package adapter

import (
	"fmt"
	"os/exec"
	"strings"
)

// AptAdapter implements PackageManager for apt-based systems (Debian,
// Ubuntu, etc.).
type AptAdapter struct{}

// Name returns the adapter identifier.
func (a *AptAdapter) Name() string {
	return "apt"
}

// IsAvailable checks whether the apt binary is on $PATH.
func (a *AptAdapter) IsAvailable() bool {
	_, err := exec.LookPath("apt")
	return err == nil
}

// Search queries apt for packages matching the given query. It runs
// `apt search <query>` and parses the output into Package structs.
func (a *AptAdapter) Search(query string) ([]Package, error) {
	cmd := exec.Command("apt", "search", query)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("apt search %s: %w", query, err)
	}

	return parseAptSearch(string(output), query), nil
}

// parseAptSearch parses the output of `apt search`. The format is:
//
//	package/distribution version arch
//	  description line 1
//	  description line 2
//
// Lines without indentation start a new package entry.
func parseAptSearch(output, query string) []Package {
	var packages []Package
	lines := strings.Split(output, "\n")

	var current *Package
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Indented lines are description continuations
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if current != nil && trimmed != "" {
				if current.Description != "" {
					current.Description += " "
				}
				current.Description += trimmed
			}
			continue
		}

		// Non-indented line: finish previous package, start new one
		if current != nil {
			packages = append(packages, *current)
			current = nil
		}

		// Skip apt search header lines ("Sorting...", "Full Text Search...")
		if !strings.Contains(trimmed, "/") {
			continue
		}

		// Format: "name/distribution version arch"
		parts := strings.SplitN(trimmed, " ", 2)
		if len(parts) == 0 {
			continue
		}

		// Split name from distribution
		namePart := parts[0]
		nameDist := strings.SplitN(namePart, "/", 2)
		name := nameDist[0]

		version := ""
		if len(parts) == 2 {
			// Second part might be "version arch" or just "version"
			verParts := strings.Fields(parts[1])
			if len(verParts) > 0 {
				version = verParts[0]
			}
		}

		current = &Package{
			Name:    name,
			Source:  "apt",
			Version: version,
		}
	}

	// Append the last package (if any)
	if current != nil {
		packages = append(packages, *current)
	}

	return packages
}

// Install delegates to `sudo apt install -y <name>`.
func (a *AptAdapter) Install(pkg Package) error {
	cmd := exec.Command("sudo", "apt", "install", "-y", pkg.Name)
	cmd.Stdout = nil // let caller handle output
	cmd.Stderr = nil
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apt install %s: %s — %w", pkg.Name, string(output), err)
	}
	return nil
}

// Uninstall delegates to `sudo apt remove -y <name>`.
func (a *AptAdapter) Uninstall(pkg Package) error {
	cmd := exec.Command("sudo", "apt", "remove", "-y", pkg.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apt remove %s: %s — %w", pkg.Name, string(output), err)
	}
	return nil
}

// Info delegates to `apt show <name>` and parses the output.
func (a *AptAdapter) Info(pkg Package) (Details, error) {
	cmd := exec.Command("apt", "show", pkg.Name)
	output, err := cmd.Output()
	if err != nil {
		return Details{}, fmt.Errorf("apt show %s: %w", pkg.Name, err)
	}

	return parseAptShow(string(output)), nil
}

// parseAptShow parses `apt show` output into a Details struct.
func parseAptShow(output string) Details {
	d := Details{}
	lines := strings.Split(output, "\n")
	var descLines []string
	inDescription := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if trimmed == "Description:" {
			inDescription = true
			continue
		}

		if inDescription {
			if strings.HasPrefix(trimmed, "Description-") {
				inDescription = false
				continue
			}
			if strings.Contains(line, ":") && !strings.HasPrefix(line, " ") {
				// Next field started
				inDescription = false
				// fall through to parse the field
			}
		}

		if !inDescription {
			parts := strings.SplitN(trimmed, ": ", 2)
			if len(parts) != 2 {
				continue
			}
			key := parts[0]
			value := parts[1]

			switch key {
			case "Package":
				d.Name = value
			case "Version":
				d.Version = value
			case "Homepage":
				d.Homepage = value
			case "Installed-Size":
				// Installed-Size is in KB
				var sizeKB int64
				_, _ = fmt.Sscanf(value, "%d", &sizeKB)
				d.Size = sizeKB * 1024
			}
		} else {
			descLines = append(descLines, trimmed)
		}
	}

	d.Description = strings.Join(descLines, " ")
	return d
}
