// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// newZshCompletionCmd creates the zsh completion subcommand.
// This generates a zsh completion script that can be sourced to enable
// tab-completion for the StackEye CLI.
func newZshCompletionCmd() *cobra.Command {
	var noDescriptions bool

	cmd := &cobra.Command{
		Use:   "zsh",
		Short: "Generate zsh completion script",
		Long: `Generate zsh completion script for StackEye CLI.

To enable zsh completion, add the generated script to your zsh configuration:

  # Option 1: Add to your fpath (recommended)
  # First, ensure a completions directory exists and is in fpath:
  mkdir -p ~/.zsh/completions
  stackeye completion zsh > ~/.zsh/completions/_stackeye

  # Add to ~/.zshrc if not already present:
  fpath=(~/.zsh/completions $fpath)
  autoload -Uz compinit && compinit

  # Option 2: Source directly in ~/.zshrc
  source <(stackeye completion zsh)

  # Option 3: Install system-wide (macOS with Homebrew)
  stackeye completion zsh > $(brew --prefix)/share/zsh/site-functions/_stackeye

  # Option 4: Install system-wide (Linux)
  stackeye completion zsh | sudo tee /usr/local/share/zsh/site-functions/_stackeye > /dev/null

After installation, restart your shell or run:

  autoload -Uz compinit && compinit

Note: If completions don't work immediately, you may need to delete the
completion cache:

  rm -f ~/.zcompdump*

The --no-descriptions flag disables completion descriptions for a cleaner
but less informative completion experience.
`,
		Example: `  # Generate completion script to stdout
  stackeye completion zsh

  # Install to user completions directory
  mkdir -p ~/.zsh/completions
  stackeye completion zsh > ~/.zsh/completions/_stackeye

  # Install with Homebrew (macOS)
  stackeye completion zsh > $(brew --prefix)/share/zsh/site-functions/_stackeye

  # Generate without descriptions (shorter output)
  stackeye completion zsh --no-descriptions`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate completion script to stdout
			// GenZshCompletionNoDesc generates a minimal script without descriptions
			// GenZshCompletion generates a full script with descriptions
			if noDescriptions {
				return rootCmd.GenZshCompletionNoDesc(os.Stdout)
			}
			return rootCmd.GenZshCompletion(os.Stdout)
		},
	}

	cmd.Flags().BoolVar(&noDescriptions, "no-descriptions", false, "disable completion descriptions")

	return cmd
}
