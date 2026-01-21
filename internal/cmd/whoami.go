// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
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

	// Create API client and verify credentials (with verbosity options)
	opts := api.GetClientOptions()
	apiClient := client.New(ctx.APIKey, ctx.EffectiveAPIURL(), opts...)

	reqCtx, cancel := context.WithTimeout(context.Background(), whoamiTimeout)
	defer cancel()

	// Use VerifyCLICredentials instead of GetCurrentUser
	// This endpoint works with API key auth and returns user info from the API key creator
	verifyResp, err := client.VerifyCLICredentials(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to verify credentials: %w", err)
	}

	// Print user and organization information
	printWhoamiInfoFromVerify(cfg.CurrentContext, ctx, verifyResp)

	return nil
}

// printWhoamiInfoFromVerify formats and prints the user and context information
// from the CLI credential verification response.
func printWhoamiInfoFromVerify(contextName string, ctx *config.Context, resp *client.CLIVerifyResponse) {
	fmt.Println()

	// User info (if available from API key creator)
	if resp.UserEmail != "" {
		fmt.Printf("User:         %s\n", resp.UserEmail)
		if resp.UserName != "" && resp.UserName != resp.UserEmail {
			fmt.Printf("Name:         %s\n", resp.UserName)
		}
		if resp.IsPlatformAdmin {
			fmt.Printf("Role:         Platform Admin\n")
		}
		fmt.Println()
	}

	// Context info
	fmt.Printf("Context:      %s\n", contextName)
	fmt.Printf("Organization: %s (%s)\n", resp.OrganizationName, resp.OrganizationID)
	fmt.Printf("API URL:      %s\n", ctx.EffectiveAPIURL())
	fmt.Printf("Auth Type:    %s\n", resp.AuthType)

	fmt.Println()
}
