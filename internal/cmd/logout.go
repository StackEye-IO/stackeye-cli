// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"fmt"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/spf13/cobra"
)

// logoutFlags holds the command flags for the logout command.
type logoutFlags struct {
	all bool
}

// NewLogoutCmd creates and returns the logout command.
func NewLogoutCmd() *cobra.Command {
	flags := &logoutFlags{}

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out from your StackEye account",
		Long: `Log out from your StackEye account by clearing stored credentials.

By default, this clears the API key from the current context only.
Use --all to clear API keys from all configured contexts.

The context configuration (API URL, organization name) is preserved,
allowing you to quickly log back in with 'stackeye login'.

Examples:
  # Log out from the current context
  stackeye logout

  # Log out from all contexts
  stackeye logout --all`,
		// Override PersistentPreRunE to skip config loading requirement.
		// The logout command should work without valid authentication.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(flags)
		},
	}

	cmd.Flags().BoolVar(&flags.all, "all", false, "clear credentials from all contexts")

	return cmd
}

// runLogout executes the logout operation.
func runLogout(flags *logoutFlags) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if there are any contexts to log out from
	if len(cfg.Contexts) == 0 {
		fmt.Println("Not logged in to any context.")
		return nil
	}

	if flags.all {
		return logoutAll(cfg)
	}

	return logoutCurrent(cfg)
}

// logoutCurrent clears the API key from the current context.
func logoutCurrent(cfg *config.Config) error {
	// Check if there's a current context set
	if cfg.CurrentContext == "" {
		fmt.Println("No current context set. Use 'stackeye logout --all' to clear all contexts.")
		return nil
	}

	// Get the current context
	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		fmt.Printf("Context '%s' not found. Use 'stackeye logout --all' to clear all contexts.\n", cfg.CurrentContext)
		return nil
	}

	// Check if already logged out
	if ctx.APIKey == "" {
		fmt.Printf("Already logged out from context '%s'.\n", cfg.CurrentContext)
		return nil
	}

	// Clear the API key
	contextName := cfg.CurrentContext
	orgName := ctx.OrganizationName
	ctx.APIKey = ""

	// Save configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Print success message
	fmt.Printf("Logged out from context '%s'", contextName)
	if orgName != "" {
		fmt.Printf(" (%s)", orgName)
	}
	fmt.Println(".")

	return nil
}

// logoutAll clears API keys from all contexts.
func logoutAll(cfg *config.Config) error {
	// Count how many contexts have credentials
	var loggedInCount int
	var loggedOutContexts []string

	for name, ctx := range cfg.Contexts {
		if ctx != nil && ctx.APIKey != "" {
			ctx.APIKey = ""
			loggedInCount++
			loggedOutContexts = append(loggedOutContexts, name)
		}
	}

	// Check if there was anything to log out from
	if loggedInCount == 0 {
		fmt.Println("Not logged in to any context.")
		return nil
	}

	// Save configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Print success message
	if loggedInCount == 1 {
		fmt.Printf("Logged out from 1 context: %s\n", loggedOutContexts[0])
	} else {
		fmt.Printf("Logged out from %d contexts:\n", loggedInCount)
		for _, name := range loggedOutContexts {
			fmt.Printf("  - %s\n", name)
		}
	}

	return nil
}
