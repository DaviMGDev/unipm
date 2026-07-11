// Package adapter defines the PackageManager interface that every backend
// package manager must implement, along with shared data types used across
// the unipm system.
package adapter

// Package represents a package as returned by any backend adapter. It is the
// common currency of the router — all adapters produce and consume Package
// values.
type Package struct {
	// Name is the package identifier as understood by the source backend
	// (e.g., "htop", "python3-requests", "@angular/core").
	Name string

	// Source identifies the backend that provided this package
	// (e.g., "apt", "npm", "pypi"). Must match a registered adapter name.
	Source string

	// Version is the available or installed version string. Format is
	// backend-specific and may be empty if the backend doesn't report
	// versions during search.
	Version string

	// Description is a human-readable summary of the package. May be
	// truncated or empty depending on the backend.
	Description string
}

// Details provides extended metadata for a package beyond the basic search
// result, returned by the Info method.
type Details struct {
	// Name is the canonical package name.
	Name string

	// Version is the currently installed or latest version.
	Version string

	// Description is a full (non-truncated) description.
	Description string

	// Homepage is the upstream project URL, if available.
	Homepage string

	// License is the SPDX license identifier, if available.
	License string

	// Size is the installed size in bytes, if available.
	Size int64
}

// PackageManager is the interface every backend adapter must implement. The
// router holds a registry of PackageManager implementations and dispatches
// commands to them polymorphically.
//
// Invariants:
//   - IsAvailable() must be called before Search(), Install(), or Uninstall().
//     The router enforces this at startup by only registering available adapters.
//   - Install() and Uninstall() receive a Package whose Source matches the
//     adapter's Name().
//   - Methods that cannot be meaningfully implemented must return a clear
//     "not supported" error rather than silently succeeding.
type PackageManager interface {
	// Name returns the adapter identifier (e.g., "apt", "npm", "pypi").
	Name() string

	// Search queries the backend for packages matching the given query string
	// and returns matching packages. Returns an empty slice if no matches are
	// found.
	Search(query string) ([]Package, error)

	// Install delegates installation of the given package to the native
	// package manager. The caller must ensure pkg.Source matches the adapter's
	// Name().
	Install(pkg Package) error

	// Uninstall delegates removal of the given package to the native package
	// manager. The caller must ensure pkg.Source matches the adapter's Name().
	Uninstall(pkg Package) error

	// Info returns extended metadata for a package. Returns a "not supported"
	// error if the backend cannot provide this information (e.g., AppImage).
	Info(pkg Package) (Details, error)

	// IsAvailable checks whether the required backend binary is present on
	// $PATH and is executable.
	IsAvailable() bool
}
