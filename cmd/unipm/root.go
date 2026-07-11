package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DaviMGDev/unipm/pkg/config"
	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags.
// Example: go build -ldflags="-X main.version=0.1.0" ./cmd/unipm
var version = "dev"

// rootCmd is the top-level unipm command.
var rootCmd = &cobra.Command{
	Use:   "unipm",
	Short: "Universal Package Manager",
	Long: `unipm is a meta package manager that unifies apt, npm, pypi, flatpak,
brew, appimage, and pacman/yay (via Distrobox) under a single CLI.

Search, install, and remove software from any ecosystem without
memorizing a dozen different flags.`,
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip directory creation for help and completion commands
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}
		return ensureConfigDir()
	},
}

func init() {
	// Persistent flags
	configPath := filepath.Join(unipmHome(), "config.yaml")
	rootCmd.PersistentFlags().String("config", configPath, "path to config file")

	// Register subcommands (defined in their own files)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(sourcesCmd)
	rootCmd.AddCommand(completionCmd)
}

// unipmHome returns the path to the ~/.unipm directory.
func unipmHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".unipm"
	}
	return filepath.Join(home, ".unipm")
}

// ensureConfigDir creates the ~/.unipm directory if it doesn't exist.
func ensureConfigDir() error {
	if err := config.EnsureDir(); err != nil {
		return fmt.Errorf("unipm: %w", err)
	}
	return nil
}
