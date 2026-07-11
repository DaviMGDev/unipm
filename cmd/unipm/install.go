package main

import (
	"fmt"
	"strings"

	"github.com/DaviMGDev/unipm/pkg/adapter"
	"github.com/DaviMGDev/unipm/pkg/state"
	"github.com/DaviMGDev/unipm/pkg/ui"
	"github.com/spf13/cobra"
)

func init() {
	installCmd.Flags().StringP("source", "s", "", "source to install from (e.g., apt, npm, pypi)")
}

var installCmd = &cobra.Command{
	Use:   "install <package>",
	Short: "Install a package using the appropriate backend",
	Long: `Install a package. If the package exists in multiple sources,
an interactive TUI opens for selection. Use --source to bypass the prompt.

Examples:
  unipm install htop           # interactive if ambiguous
  unipm install htop --source apt   # install from apt only
  unipm install htop -s apt,npm     # install from multiple sources`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

// runInstall is the handler for the install command.
func runInstall(cmd *cobra.Command, args []string) error {
	pkgName := args[0]
	sourceFlag, _ := cmd.Flags().GetString("source")

	// If --source is given, install from named sources directly
	if sourceFlag != "" {
		return installFromSources(pkgName, sourceFlag)
	}

	// No --source: search all adapters and handle collision
	return installWithCollision(pkgName)
}

// installFromSources installs the package from the named source(s).
func installFromSources(pkgName, sourceFlag string) error {
	sourceNames := strings.Split(sourceFlag, ",")

	for _, name := range sourceNames {
		name = strings.TrimSpace(name)
		a, err := appRouter.Get(name)
		if err != nil {
			available := strings.Join(appRouter.Names(), ", ")
			return fmt.Errorf(
				"%q is not available.\n\nAvailable sources: %s\n\nRun 'unipm sources' to see all detected package managers.",
				name, available,
			)
		}

		if err := installSingle(a, pkgName); err != nil {
			return err
		}
	}

	return nil
}

// installWithCollision searches all adapters and handles the result
// based on how many matches are found.
func installWithCollision(pkgName string) error {
	if appRouter.IsEmpty() {
		return fmt.Errorf("no package managers available. Run 'unipm sources' to check.")
	}

	adapters := appRouter.List()

	// Search all adapters synchronously
	var matches []adapter.Package
	for _, a := range adapters {
		pkgs, err := a.Search(pkgName)
		if err != nil {
			fmt.Printf("warning: %s search failed: %v\n", a.Name(), err)
			continue
		}
		matches = append(matches, pkgs...)
	}

	if len(matches) == 0 {
		return fmt.Errorf(
			"no package named %q found in any source.\n\n"+
				"Try a different query, or run 'unipm sources' to see available package managers.",
			pkgName,
		)
	}

	if len(matches) == 1 {
		a, err := appRouter.Get(matches[0].Source)
		if err != nil {
			return err
		}
		return installSingle(a, pkgName)
	}

	// Multiple matches — collision. Open TUI for selection.
	selected, err := ui.RunSelection(matches)
	if err != nil {
		return err
	}

	a, err := appRouter.Get(selected.Source)
	if err != nil {
		return err
	}

	return installSingle(a, pkgName)
}

// installSingle installs a package from a specific adapter and records it.
func installSingle(a adapter.PackageManager, pkgName string) error {
	pkg := adapter.Package{
		Name:   pkgName,
		Source: a.Name(),
	}

	fmt.Printf("Installing %s from %s...\n", pkgName, a.Name())

	if err := a.Install(pkg); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Get version info for the state record
	version := "unknown"
	info, err := a.Info(pkg)
	if err == nil && info.Version != "" {
		version = info.Version
	}

	record := state.Record{
		Name:        pkgName,
		Source:      a.Name(),
		Version:     version,
		InstalledAt: nowUTC(),
	}

	if err := state.Add(record); err != nil {
		return fmt.Errorf("state recording failed: %w", err)
	}

	fmt.Printf("✓ %s %s installed from %s\n", pkgName, version, a.Name())
	return nil
}
