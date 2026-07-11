package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/DaviMGDev/unipm/pkg/adapter"
	"github.com/DaviMGDev/unipm/pkg/cache"
	"github.com/DaviMGDev/unipm/pkg/config"
	"github.com/spf13/cobra"
)

func init() {
	searchCmd.Flags().Int("timeout", 10, "per-adapter timeout in seconds")
	searchCmd.Flags().StringP("source", "s", "", "source(s) to search (comma-separated, e.g. apt,npm)")

	// Register completion for the positional package-name argument
	searchCmd.ValidArgsFunction = packageNameCompletion
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for a package across all available sources",
	Long: `Query all available package backends in parallel and display
results in a unified table with Source, Name, Version, and Description.

Use --source to limit the search to specific backends:
  unipm search htop --source apt
  unipm search react -s npm`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

// runSearch is the handler for the search command.
func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	timeoutSec, _ := cmd.Flags().GetInt("timeout")
	timeout := time.Duration(timeoutSec) * time.Second
	sourceFlag, _ := cmd.Flags().GetString("source")

	if appRouter.IsEmpty() {
		return fmt.Errorf(
			"no package managers are available.\n\n" +
				"unipm requires at least one of: apt, npm.\n" +
				"Install one of these and try again, or run 'unipm sources' to see what's detected.",
		)
	}

	// If --source is given, search only named sources
	if sourceFlag != "" {
		return searchFromSources(query, sourceFlag, timeout)
	}

	// Fan out to all registered adapters via the router
	allPackages, timedOutSources, erroredSources := appRouter.SearchAll(query, timeout)

	// Print warnings for timed-out adapters
	for _, source := range timedOutSources {
		fmt.Fprintf(os.Stderr, "warning: %s search timed out (showing partial results)\n", source)
	}

	// Print warnings for errored adapters
	for _, source := range erroredSources {
		fmt.Fprintf(os.Stderr, "warning: %s search failed\n", source)
	}

	hadFailures := len(timedOutSources) > 0 || len(erroredSources) > 0

	if len(allPackages) == 0 {
		if hadFailures {
			fmt.Println("No results found. Some backends timed out or failed — try again or adjust --timeout.")
		} else {
			fmt.Printf("No packages found for %q.\n", query)
		}
		return nil
	}

	// Print results table
	printSearchTable(allPackages)

	// Populate completion cache with search results
	populateCompletionCache(allPackages)

	return nil
}

// searchFromSources searches only the named adapter(s) synchronously.
func searchFromSources(query, sourceFlag string, timeout time.Duration) error {
	sourceNames := strings.Split(sourceFlag, ",")

	for _, name := range sourceNames {
		name = strings.TrimSpace(name)
		a, err := appRouter.Get(name)
		if err != nil {
			return err
		}

		pkgs, err := a.Search(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s search failed: %v\n", name, err)
			continue
		}

		if len(pkgs) > 0 {
			fmt.Printf("\n── %s ──\n", name)
			printSearchTable(pkgs)
		}
	}

	return nil
}

// printSearchTable outputs search results as a simple aligned table.
func printSearchTable(packages []adapter.Package) {
	if len(packages) == 0 {
		fmt.Println("(no results)")
		return
	}

	// Sort by source then name for consistent output
	sort.Slice(packages, func(i, j int) bool {
		if packages[i].Source != packages[j].Source {
			return packages[i].Source < packages[j].Source
		}
		return packages[i].Name < packages[j].Name
	})

	// Calculate column widths
	maxSource := len("Source")
	maxName := len("Name")
	maxVersion := len("Version")
	for _, p := range packages {
		if len(p.Source) > maxSource {
			maxSource = len(p.Source)
		}
		if len(p.Name) > maxName {
			maxName = len(p.Name)
		}
		if len(p.Version) > maxVersion {
			maxVersion = len(p.Version)
		}
	}

	// Print header
	fmt.Printf(
		"%-*s  %-*s  %-*s  %s\n",
		maxSource, "Source",
		maxName, "Name",
		maxVersion, "Version",
		"Description",
	)
	fmt.Printf(
		"%s  %s  %s  %s\n",
		strings.Repeat("─", maxSource),
		strings.Repeat("─", maxName),
		strings.Repeat("─", maxVersion),
		strings.Repeat("─", 40),
	)

	// Print rows
	for _, p := range packages {
		fmt.Printf(
			"%-*s  %-*s  %-*s  %s\n",
			maxSource, p.Source,
			maxName, p.Name,
			maxVersion, p.Version,
			p.Description,
		)
	}
}

// populateCompletionCache adds search result package names to the
// tab-completion cache for future shell completions.
func populateCompletionCache(packages []adapter.Package) {
	names := make([]string, 0, len(packages))
	for _, p := range packages {
		names = append(names, p.Name)
	}
	if len(names) > 0 {
		_ = cache.AddRecords(names)
	}
}

// packageNameCompletion provides shell completion for package names
// from the completion cache.
func packageNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Don't attempt network completions for queries shorter than 3 chars
	if len(toComplete) < 3 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ttl := time.Duration(cfg.CacheTTL) * time.Second
	matches := cache.Matching(toComplete, ttl)

	if len(matches) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}
