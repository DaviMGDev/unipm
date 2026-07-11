package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/DaviMGDev/unipm/pkg/adapter"
	"github.com/DaviMGDev/unipm/pkg/state"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <package>",
	Short: "Uninstall a package tracked by unipm",
	Long:  `Uninstall a package that was installed via unipm by looking up its source in the local state file.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUninstall,
}

func runUninstall(cmd *cobra.Command, args []string) error {
	pkgName := args[0]

	rec, err := state.Get(pkgName)
	if err != nil {
		// Check if the error is a "not found" error
		if strings.Contains(err.Error(), "was not installed via unipm") {
			return err
		}
		return fmt.Errorf("lookup %s: %w", pkgName, err)
	}

	a, err := findAdapterByName(rec.Source)
	if err != nil {
		return fmt.Errorf("uninstall %s: %w", pkgName, err)
	}

	pkg := adapter.Package{
		Name:   pkgName,
		Source: rec.Source,
	}

	fmt.Printf("Removing %s from %s...\n", pkgName, rec.Source)

	if err := a.Uninstall(pkg); err != nil {
		// Offer to clean state record even if native removal fails
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: backend removal failed: %v\n", err)
		fmt.Printf("Remove %s from unipm tracking anyway? [y/N] ", pkgName)

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			return fmt.Errorf("uninstall cancelled: %w", err)
		}
	}

	if err := state.Remove(pkgName); err != nil {
		return fmt.Errorf("state cleanup failed: %w", err)
	}

	fmt.Printf("✓ %s removed from %s\n", pkgName, rec.Source)
	return nil
}

// nowUTC returns the current time as an RFC 3339 string.
func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}
