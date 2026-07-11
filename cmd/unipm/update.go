package main

import (
	"fmt"
	"strings"

	"github.com/DaviMGDev/unipm/pkg/adapter"
	"github.com/DaviMGDev/unipm/pkg/state"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [package]",
	Short: "Update packages managed by unipm",
	Long: `Update all tracked packages or a single named package.
Without an argument, all packages in state.json are updated.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		return updateSingle(args[0])
	}
	return updateAll()
}

// updateAll iterates all records in state.json and updates each one.
func updateAll() error {
	records, err := state.List()
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}

	if len(records) == 0 {
		fmt.Println("No packages are tracked by unipm.")
		return nil
	}

	hadFailure := false
	for _, rec := range records {
		a, err := findAdapterByName(rec.Source)
		if err != nil {
			fmt.Printf("✗ %s (%s): adapter not available — %v\n", rec.Name, rec.Source, err)
			hadFailure = true
			continue
		}

		pkg := adapter.Package{Name: rec.Name, Source: rec.Source}
		if err := a.Install(pkg); err != nil {
			fmt.Printf("✗ %s (%s): %v\n", rec.Name, rec.Source, err)
			hadFailure = true
			continue
		}

		// Refresh version in state
		version := rec.Version
		if info, err := a.Info(pkg); err == nil && info.Version != "" {
			version = info.Version
		}

		if err := state.UpdateVersion(rec.Name, version, nowUTC()); err != nil {
			fmt.Printf("warning: %s version update failed: %v\n", rec.Name, err)
		}

		fmt.Printf("✓ %s updated to %s (%s)\n", rec.Name, version, rec.Source)
	}

	if hadFailure {
		return fmt.Errorf("some packages failed to update")
	}
	return nil
}

// updateSingle updates a single named package from the state file.
func updateSingle(pkgName string) error {
	rec, err := state.Get(pkgName)
	if err != nil {
		if strings.Contains(err.Error(), "was not installed via unipm") {
			return err
		}
		return fmt.Errorf("lookup %s: %w", pkgName, err)
	}

	a, err := findAdapterByName(rec.Source)
	if err != nil {
		return fmt.Errorf("update %s: %w", pkgName, err)
	}

	pkg := adapter.Package{Name: rec.Name, Source: rec.Source}
	if err := a.Install(pkg); err != nil {
		return fmt.Errorf("update %s: %w", pkgName, err)
	}

	version := rec.Version
	if info, err := a.Info(pkg); err == nil && info.Version != "" {
		version = info.Version
	}

	if err := state.UpdateVersion(rec.Name, version, nowUTC()); err != nil {
		return fmt.Errorf("version update failed: %w", err)
	}

	fmt.Printf("✓ %s updated to %s (%s)\n", rec.Name, version, rec.Source)
	return nil
}
