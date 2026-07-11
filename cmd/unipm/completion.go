package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion script",
	Long: `Generate a shell completion script for bash, zsh, or fish.

To load completions:

  Bash:
    eval "$(unipm completion bash)"

  Zsh:
    eval "$(unipm completion zsh)"

  Fish:
    unipm completion fish | source`,
	ValidArgs: []string{"bash", "zsh", "fish"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:      runCompletion,
}

func runCompletion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		return rootCmd.GenBashCompletion(cmd.OutOrStdout())
	case "zsh":
		return rootCmd.GenZshCompletion(cmd.OutOrStdout())
	case "fish":
		return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", args[0])
	}
}
