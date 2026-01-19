// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewContextCmd creates and returns the context parent command.
func NewContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage organization contexts",
		Long: `Manage organization contexts for the StackEye CLI.

Contexts allow you to switch between different organizations or API endpoints,
similar to kubectl contexts for Kubernetes clusters.

Each context stores authentication credentials and configuration for a specific
organization. Use 'stackeye context use <name>' to switch between contexts.

Examples:
  # List all contexts
  stackeye context list

  # Use a specific context
  stackeye context use myorg

  # Show current context
  stackeye context current`,
	}

	// Add subcommands
	cmd.AddCommand(newContextListCmd())

	return cmd
}

// newContextListCmd creates the list subcommand to display all contexts.
func newContextListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured contexts",
		Long: `List all configured organization contexts.

The current context is marked with an asterisk (*). Each context shows
the organization name (if set) and the API URL it connects to.

Examples:
  # List all contexts
  stackeye context list`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContextList()
		},
	}

	return cmd
}

// runContextList executes the context list command.
func runContextList() error {
	cfg := GetConfig()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Get sorted context names
	names := cfg.ContextNames()

	if len(names) == 0 {
		fmt.Println("No contexts configured.")
		fmt.Println("")
		fmt.Println("Run 'stackeye login' to authenticate and create a context,")
		fmt.Println("or use 'stackeye config set-key' to configure an API key.")
		return nil
	}

	// Print header
	fmt.Printf("%-2s %-20s %-25s %s\n", "", "NAME", "ORGANIZATION", "API URL")

	// Print each context
	for _, name := range names {
		ctx, err := cfg.GetContext(name)
		if err != nil || ctx == nil {
			continue // Skip invalid or nil contexts
		}

		// Determine marker for current context
		marker := ""
		if cfg.CurrentContext == name {
			marker = "*"
		}

		// Get display values with truncation for alignment
		displayName := truncateStr(name, 20)
		orgName := ctx.OrganizationName
		if orgName == "" {
			orgName = "(not set)"
		}
		displayOrg := truncateStr(orgName, 25)

		apiURL := ctx.EffectiveAPIURL()

		// Print context row
		fmt.Printf("%-2s %-20s %-25s %s\n", marker, displayName, displayOrg, apiURL)
	}

	return nil
}

// truncateStr truncates a string to maxLen, adding "..." if truncated.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
