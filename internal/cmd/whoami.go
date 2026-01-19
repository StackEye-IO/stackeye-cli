// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/spf13/cobra"
)

const (
	// whoamiTimeout is the maximum time to wait for the API response.
	whoamiTimeout = 30 * time.Second
)

// NewWhoamiCmd creates and returns the whoami command.
func NewWhoamiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Display current user and context information",
		Long: `Display information about the currently authenticated user and active context.

Shows the user's email, name, and organization details from the current
context. This is useful for verifying which account and organization
you're operating as before running commands.

If you're not logged in, the command will indicate that no credentials
are configured and suggest running 'stackeye login'.

Examples:
  # Show current user info
  stackeye whoami

  # Show user info for a specific context
  stackeye whoami --context production`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhoami()
		},
	}

	return cmd
}

// runWhoami executes the whoami command.
func runWhoami() error {
	// Get the loaded configuration
	cfg := GetConfig()
	if cfg == nil {
		fmt.Println("Not logged in.")
		fmt.Println("Run 'stackeye login' to authenticate.")
		return nil
	}

	// Check if there's a current context
	if cfg.CurrentContext == "" {
		fmt.Println("Not logged in.")
		fmt.Println("No context configured. Run 'stackeye login' to authenticate.")
		return nil
	}

	// Get the current context
	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		fmt.Println("Not logged in.")
		fmt.Printf("Context '%s' not found. Run 'stackeye login' to authenticate.\n", cfg.CurrentContext)
		return nil
	}

	// Check if there's an API key
	if ctx.APIKey == "" {
		fmt.Println("Not logged in.")
		fmt.Printf("No credentials for context '%s'. Run 'stackeye login' to authenticate.\n", cfg.CurrentContext)
		return nil
	}

	// Create API client and fetch user info
	apiClient := client.New(ctx.APIKey, ctx.EffectiveAPIURL())

	reqCtx, cancel := context.WithTimeout(context.Background(), whoamiTimeout)
	defer cancel()

	userResp, err := client.GetCurrentUser(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	// Print user information
	printWhoamiInfo(cfg.CurrentContext, ctx, userResp)

	return nil
}

// printWhoamiInfo formats and prints the user and context information.
func printWhoamiInfo(contextName string, ctx *config.Context, userResp *client.UserResponse) {
	user := userResp.User

	fmt.Println()

	// User info
	fmt.Printf("User:         %s\n", user.Email)
	if displayName := user.GetDisplayName(); displayName != "" && displayName != user.Email {
		fmt.Printf("Name:         %s\n", displayName)
	}

	// Platform admin badge
	if user.IsPlatformAdmin {
		fmt.Printf("Role:         Platform Admin\n")
	}

	fmt.Println()

	// Context info
	fmt.Printf("Context:      %s\n", contextName)
	if ctx.OrganizationName != "" {
		fmt.Printf("Organization: %s", ctx.OrganizationName)
		if ctx.OrganizationID != "" {
			fmt.Printf(" (%s)", ctx.OrganizationID)
		}
		fmt.Println()
	}
	fmt.Printf("API URL:      %s\n", ctx.EffectiveAPIURL())

	fmt.Println()
}
