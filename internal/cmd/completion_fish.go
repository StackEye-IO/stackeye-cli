// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// newFishCompletionCmd creates the fish completion subcommand.
// This generates a fish completion script that can be sourced to enable
// tab-completion for the StackEye CLI.
func newFishCompletionCmd() *cobra.Command {
	var noDescriptions bool

	cmd := &cobra.Command{
		Use:   "fish",
		Short: "Generate fish completion script",
		Long: `Generate fish completion script for StackEye CLI.

To enable fish completion, save the generated script to your fish completions directory:

  # User-specific completions (recommended)
  mkdir -p ~/.config/fish/completions
  stackeye completion fish > ~/.config/fish/completions/stackeye.fish

  # System-wide completions (requires sudo)
  stackeye completion fish | sudo tee /usr/share/fish/vendor_completions.d/stackeye.fish > /dev/null

  # macOS with Homebrew fish:
  stackeye completion fish > $(brew --prefix)/share/fish/vendor_completions.d/stackeye.fish

After installation, completions will be available in new fish shell sessions.
To use immediately without restarting:

  source ~/.config/fish/completions/stackeye.fish

The --no-descriptions flag disables completion descriptions for a cleaner
but less informative completion experience.
`,
		Example: `  # Generate completion script to stdout
  stackeye completion fish

  # Install to user completions directory
  mkdir -p ~/.config/fish/completions
  stackeye completion fish > ~/.config/fish/completions/stackeye.fish

  # Install system-wide (Linux)
  stackeye completion fish | sudo tee /usr/share/fish/vendor_completions.d/stackeye.fish > /dev/null

  # Generate without descriptions (shorter output)
  stackeye completion fish --no-descriptions`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate completion script to stdout
			// GenFishCompletion generates a full script with descriptions
			// The second parameter controls whether to include descriptions
			return rootCmd.GenFishCompletion(os.Stdout, !noDescriptions)
		},
	}

	cmd.Flags().BoolVar(&noDescriptions, "no-descriptions", false, "disable completion descriptions")

	return cmd
}
