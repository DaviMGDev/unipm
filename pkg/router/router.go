// Package router provides the adapter registry and fan-out dispatch logic
// for unipm. It holds a map of available PackageManager implementations and
// exposes methods for searching across all adapters in parallel, looking up
// adapters by name, and listing available sources.
package router

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/DaviMGDev/unipm/pkg/adapter"
)

// searchResult pairs a search result slice with the adapter that produced it.
type searchResult struct {
	packages []adapter.Package
	source   string
	err      error
}

// Registry holds all available package manager adapters and provides
// methods for dispatching operations across them.
type Registry struct {
	adapters map[string]adapter.PackageManager
}

// New creates an empty Registry. Call Register() to add adapters, then
// use the dispatch methods (SearchAll, Get, Names, List).
func New() *Registry {
	return &Registry{
		adapters: make(map[string]adapter.PackageManager),
	}
}

// Register adds an adapter to the registry. If an adapter with the same
// name is already registered, it is silently replaced. Callers should
// check IsAvailable() before registering.
func (r *Registry) Register(a adapter.PackageManager) {
	r.adapters[a.Name()] = a
}

// Get returns the adapter registered under the given name. Returns an
// error if no adapter matches.
func (r *Registry) Get(name string) (adapter.PackageManager, error) {
	a, ok := r.adapters[name]
	if !ok {
		return nil, fmt.Errorf("source %q is not available", name)
	}
	return a, nil
}

// Names returns a sorted slice of registered adapter names (alphabetical
// order, per ADR-0002 — no source ranking).
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// List returns all registered adapters in alphabetical order by name.
func (r *Registry) List() []adapter.PackageManager {
	names := r.Names()
	adapters := make([]adapter.PackageManager, 0, len(names))
	for _, name := range names {
		adapters = append(adapters, r.adapters[name])
	}
	return adapters
}

// IsEmpty returns true if no adapters are registered.
func (r *Registry) IsEmpty() bool {
	return len(r.adapters) == 0
}

// Count returns the number of registered adapters.
func (r *Registry) Count() int {
	return len(r.adapters)
}

// SearchAll fans out a search query to all registered adapters in parallel.
// Each adapter runs with the given timeout. Returns:
//   - merged results from all successful adapters, deduplicated by (source, name)
//   - list of source names that timed out
//   - list of source names that errored (non-timeout)
func (r *Registry) SearchAll(query string, timeout time.Duration) ([]adapter.Package, []string, []string) {
	adapters := r.List()
	if len(adapters) == 0 {
		return nil, nil, nil
	}

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
	var timedOut []string
	var errored []string

	for res := range results {
		if res.err != nil {
			if isTimeoutError(res.err) {
				timedOut = append(timedOut, res.source)
			} else {
				errored = append(errored, res.source)
			}
			continue
		}
		allPackages = append(allPackages, res.packages...)
	}

	// Deduplicate by (Source, Name)
	allPackages = deduplicate(allPackages)

	return allPackages, timedOut, errored
}

// deduplicate removes packages with duplicate (source, name) keys.
// The first occurrence is kept.
func deduplicate(packages []adapter.Package) []adapter.Package {
	seen := make(map[string]bool)
	var result []adapter.Package
	for _, p := range packages {
		key := p.Source + "/" + p.Name
		if !seen[key] {
			seen[key] = true
			result = append(result, p)
		}
	}
	return result
}

// isTimeoutError returns true if the error string contains "timeout".
func isTimeoutError(err error) bool {
	return err != nil && len(err.Error()) >= 7 && err.Error()[:7] == "timeout"
}
