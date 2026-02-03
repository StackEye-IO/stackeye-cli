// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"errors"
	"fmt"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
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
	cmd.AddCommand(newContextUseCmd())
	cmd.AddCommand(newContextCurrentCmd())
	cmd.AddCommand(newContextDeleteCmd())

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

// newContextUseCmd creates the use subcommand to switch active context.
func newContextUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Switch to a different context",
		Long: `Switch to a different organization context.

This command changes the active context used for all subsequent commands.
The context must already exist in your configuration.

Examples:
  # Switch to the production context
  stackeye context use acme-prod

  # Switch to staging
  stackeye context use acme-staging`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: ContextCompletion(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContextUse(args[0])
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

// runContextUse executes the context use command.
func runContextUse(name string) error {
	cfg := GetConfig()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Validate the context exists
	ctx, err := cfg.GetContext(name)
	if err != nil {
		if errors.Is(err, config.ErrNilContext) {
			return fmt.Errorf("context %q is invalid", name)
		}
		return fmt.Errorf("context %q not found", name)
	}

	// Check if already using this context
	if cfg.CurrentContext == name {
		fmt.Printf("Already using context %q\n", name)
		return nil
	}

	// Update current context
	previousContext := cfg.CurrentContext
	cfg.CurrentContext = name

	// Save the configuration
	if err := cfg.Save(); err != nil {
		// Restore previous context on save failure
		cfg.CurrentContext = previousContext
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Print success message with context details
	fmt.Printf("Switched to context %q", name)
	if ctx.OrganizationName != "" {
		fmt.Printf(" (%s)", ctx.OrganizationName)
	}
	fmt.Println()

	return nil
}

// newContextCurrentCmd creates the current subcommand to display active context.
func newContextCurrentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Display the current context",
		Long: `Display the currently active organization context.

Shows the context name, organization name, and API URL for the active context.
Use 'stackeye context use <name>' to switch to a different context.

Examples:
  # Show current context
  stackeye context current`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContextCurrent()
		},
	}

	return cmd
}

// runContextCurrent executes the context current command.
func runContextCurrent() error {
	cfg := GetConfig()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Check if a current context is set
	if cfg.CurrentContext == "" {
		fmt.Println("No current context set.")
		fmt.Println("")
		fmt.Println("Run 'stackeye login' to authenticate and create a context,")
		fmt.Println("or use 'stackeye context use <name>' to switch to an existing context.")
		return nil
	}

	// Get the current context details
	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		if errors.Is(err, config.ErrNilContext) {
			return fmt.Errorf("current context %q is invalid", cfg.CurrentContext)
		}
		return fmt.Errorf("current context %q not found in configuration", cfg.CurrentContext)
	}

	// Display context information
	fmt.Printf("Current context: %s\n", cfg.CurrentContext)

	orgName := ctx.OrganizationName
	if orgName == "" {
		orgName = "(not set)"
	}
	fmt.Printf("Organization:    %s\n", orgName)

	fmt.Printf("API URL:         %s\n", ctx.EffectiveAPIURL())

	return nil
}

// newContextDeleteCmd creates the delete subcommand to remove a context.
func newContextDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a context",
		Long: `Delete a context from your configuration.

This removes the specified context entirely, including its API key and
organization information. If the deleted context is the current context,
you will need to switch to another context before running commands.

Examples:
  # Delete the staging context
  stackeye context delete acme-staging

  # Delete a context by name
  stackeye context delete old-org`,
		Aliases:           []string{"rm", "remove"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: ContextCompletion(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContextDelete(args[0])
		},
	}

	return cmd
}

// runContextDelete executes the context delete command.
func runContextDelete(name string) error {
	cfg := GetConfig()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Check if the context exists before attempting to delete
	_, err := cfg.GetContext(name)
	if err != nil {
		if errors.Is(err, config.ErrNilContext) {
			return fmt.Errorf("context %q is invalid", name)
		}
		return fmt.Errorf("context %q not found", name)
	}

	// Track if we're deleting the current context
	wasCurrentContext := cfg.CurrentContext == name

	// Remove the context using SDK method with validation
	if err := cfg.RemoveContext(name); err != nil {
		return fmt.Errorf("failed to delete context: %w", err)
	}

	// Save the configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Print success message
	fmt.Printf("Deleted context %q\n", name)

	if wasCurrentContext {
		fmt.Println("")
		fmt.Println("Note: The deleted context was your current context.")
		names := cfg.ContextNames()
		if len(names) > 0 {
			fmt.Printf("Use 'stackeye context use <name>' to switch to another context.\n")
			fmt.Printf("Available contexts: %v\n", names)
		} else {
			fmt.Println("No contexts remaining. Run 'stackeye login' to create a new context.")
		}
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
