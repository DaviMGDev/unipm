package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/DaviMGDev/unipm/pkg/adapter"
	"github.com/spf13/cobra"
)

func init() {
	searchCmd.Flags().Int("timeout", 10, "per-adapter timeout in seconds")
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for a package across all available sources",
	Long: `Query all available package backends in parallel and display
results in a unified table with Source, Name, Version, and Description.`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

// availableAdapters returns a slice of all adapters that are currently
// available on the system (based on $PATH checks).
func availableAdapters() []adapter.PackageManager {
	var adapters []adapter.PackageManager

	candidates := []adapter.PackageManager{
		&adapter.AptAdapter{},
		&adapter.NpmAdapter{},
	}

	for _, a := range candidates {
		if a.IsAvailable() {
			adapters = append(adapters, a)
		}
	}

	return adapters
}

// searchResult pairs a search result slice with the adapter that produced it.
type searchResult struct {
	packages []adapter.Package
	source   string
	err      error
}

// runSearch is the handler for the search command.
func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	timeoutSec, _ := cmd.Flags().GetInt("timeout")
	timeout := time.Duration(timeoutSec) * time.Second

	adapters := availableAdapters()
	if len(adapters) == 0 {
		return fmt.Errorf(
			"no package managers are available.\n\n" +
				"unipm requires at least one of: apt, npm.\n" +
				"Install one of these and try again, or run 'unipm sources' to see what's detected.",
		)
	}

	// Fan out to all adapters in parallel
	var wg sync.WaitGroup
	results := make(chan searchResult, len(adapters))

	for _, a := range adapters {
		wg.Add(1)
		go func(ad adapter.PackageManager) {
			defer wg.Done()

			done := make(chan searchResult, 1)
			go func() {
				pkgs, err := ad.Search(query)
				done <- searchResult{packages: pkgs, source: ad.Name(), err: err}
			}()

			select {
			case res := <-done:
				results <- res
			case <-time.After(timeout):
				results <- searchResult{
					source: ad.Name(),
					err:    fmt.Errorf("timeout after %v", timeout),
				}
			}
		}(a)
	}

	wg.Wait()
	close(results)

	// Collect and merge results
	var allPackages []adapter.Package
	var timedOutSources []string
	hadErrors := false

	for res := range results {
		if res.err != nil {
			if strings.Contains(res.err.Error(), "timeout") {
				timedOutSources = append(timedOutSources, res.source)
			} else {
				fmt.Fprintf(os.Stderr, "warning: %s search failed: %v\n", res.source, res.err)
				hadErrors = true
			}
			continue
		}
		allPackages = append(allPackages, res.packages...)
	}

	// Deduplicate by (Source, Name)
	allPackages = deduplicatePackages(allPackages)

	// Print warnings
	for _, source := range timedOutSources {
		fmt.Fprintf(os.Stderr, "warning: %s search timed out (showing partial results)\n", source)
	}

	if len(allPackages) == 0 {
		if hadErrors || len(timedOutSources) > 0 {
			fmt.Println("No results found. Some backends timed out or failed — try again or adjust --timeout.")
		} else {
			fmt.Printf("No packages found for %q.\n", query)
		}
		return nil
	}

	// Print results table
	printSearchTable(allPackages)

	return nil
}

// deduplicatePackages removes duplicate entries where both Source and Name
// match. The first occurrence is kept.
func deduplicatePackages(packages []adapter.Package) []adapter.Package {
	seen := make(map[string]bool)
	var deduped []adapter.Package
	for _, p := range packages {
		key := p.Source + "/" + p.Name
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, p)
		}
	}
	return deduped
}

// printSearchTable outputs search results as a simple aligned table.
func printSearchTable(packages []adapter.Package) {
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
