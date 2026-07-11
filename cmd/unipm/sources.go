package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "List available package sources and their status",
	Long:  `Display every compiled-in adapter with its availability status (available or not found on $PATH).`,
	Args:  cobra.NoArgs,
	RunE:  runSources,
}

func runSources(cmd *cobra.Command, args []string) error {
	// Collect all adapters (available and unavailable)
	all := allAdapterStatuses()

	if len(all) == 0 {
		return fmt.Errorf(
			"no package managers are compiled into this build.\n" +
				"This is a bug — please report it.",
		)
	}

	// Check if any are available
	anyAvailable := appRouter.Count() > 0

	if !anyAvailable {
		fmt.Println("No package managers are available on your system.")
		fmt.Println("\nInstall at least one of the following and try again:")
		for _, s := range all {
			fmt.Printf("  • %s\n", s.name)
		}
		return fmt.Errorf("no package managers available")
	}

	// Print status table
	maxLen := 0
	for _, s := range all {
		if len(s.name) > maxLen {
			maxLen = len(s.name)
		}
	}

	for _, s := range all {
		status := "✗ not found on $PATH"
		if s.available {
			status = "✓ available"
		}
		fmt.Printf("%-*s  %s\n", maxLen, s.name, status)
	}

	return nil
}

// adapterStatus holds the name and availability of an adapter.
type adapterStatus struct {
	name      string
	available bool
}

// allAdapterStatuses returns statuses for all compiled-in adapters.
// This is the canonical list of adapters — must be updated when adding
// new adapters.
func allAdapterStatuses() []adapterStatus {
	// Check availability via the router for registered adapters
	avail := func(name string) bool {
		_, err := appRouter.Get(name)
		return err == nil
	}

	return []adapterStatus{
		{name: "apt", available: avail("apt")},
		{name: "npm", available: avail("npm")},
		{name: "pypi", available: avail("pypi")},
		// More adapters added in Phase 3:
		// {name: "flatpak", available: ...},
		// {name: "brew", available: ...},
		// {name: "appimage", available: ...},
	}
}
