// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// newBashCompletionCmd creates the bash completion subcommand.
// This generates a bash completion script that can be sourced to enable
// tab-completion for the StackEye CLI.
func newBashCompletionCmd() *cobra.Command {
	var noDescriptions bool

	cmd := &cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion script",
		Long: `Generate bash completion script for StackEye CLI.

To enable bash completion, add the generated script to your bash configuration:

  # Linux: Install to user completion directory
  mkdir -p ~/.bash_completion.d
  stackeye completion bash > ~/.bash_completion.d/stackeye

  # Linux: Install system-wide (requires sudo)
  stackeye completion bash | sudo tee /etc/bash_completion.d/stackeye > /dev/null

  # macOS with Homebrew bash-completion:
  stackeye completion bash > $(brew --prefix)/etc/bash_completion.d/stackeye

  # macOS with bash-completion@2:
  stackeye completion bash > $(brew --prefix)/etc/bash_completion.d/stackeye.bash

Then restart your shell or source the completion script:

  source ~/.bash_completion.d/stackeye

Requirements:
  - bash-completion v2.0+ must be installed
  - On macOS: brew install bash-completion@2

The --no-descriptions flag disables completion descriptions for a cleaner
but less informative completion experience.
`,
		Example: `  # Generate completion script to stdout
  stackeye completion bash

  # Install to user completion directory
  stackeye completion bash > ~/.bash_completion.d/stackeye

  # Generate without descriptions (shorter output)
  stackeye completion bash --no-descriptions`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate completion script to stdout
			// GenBashCompletionV2 is the modern completion generator that supports
			// descriptions and is compatible with bash-completion v2.
			return rootCmd.GenBashCompletionV2(os.Stdout, !noDescriptions)
		},
	}

	cmd.Flags().BoolVar(&noDescriptions, "no-descriptions", false, "disable completion descriptions")

	return cmd
}
