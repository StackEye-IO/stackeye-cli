// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewCompletionCmd creates and returns the completion parent command.
// This command provides shell completion script generation.
//
// Usage:
//
//	stackeye completion bash > ~/.bash_completion.d/stackeye
func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for StackEye CLI.

Shell completion enables tab-completion for commands, subcommands, and flags
in your terminal.

Currently supported shells:
  - bash

Bash:

  # Linux: Add to ~/.bashrc or install to system completion directory
  stackeye completion bash > ~/.bash_completion.d/stackeye
  # Or install system-wide (requires sudo)
  stackeye completion bash | sudo tee /etc/bash_completion.d/stackeye > /dev/null

  # macOS with Homebrew:
  stackeye completion bash > $(brew --prefix)/etc/bash_completion.d/stackeye

  Restart your shell or source the completion script to activate.
`,
		// Skip config loading - completion commands must work without authentication
		// and should have minimal latency since they run on every tab press.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	// Add shell-specific subcommands
	cmd.AddCommand(newBashCompletionCmd())

	return cmd
}
