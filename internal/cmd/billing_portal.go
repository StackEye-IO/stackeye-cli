// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/auth"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// billingPortalTimeout is the maximum time to wait for the API response.
const billingPortalTimeout = 30 * time.Second

// portalFlags holds flag values for the portal command.
var portalFlags struct {
	urlOnly bool
}

// browserOpener is the function used to open URLs in the browser.
// It can be overridden in tests.
var browserOpener = auth.OpenBrowser

// NewBillingPortalCmd creates and returns the billing portal command.
func NewBillingPortalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "portal",
		Short: "Open Stripe billing portal",
		Long: `Open the Stripe customer portal for managing your subscription.

The customer portal allows you to:
  - Update payment methods
  - View and download invoices
  - Manage subscription plans
  - Cancel or modify subscriptions

By default, this command opens the portal in your default browser.
Use --url-only to print the URL without opening the browser.

Examples:
  # Open billing portal in browser
  stackeye billing portal

  # Get portal URL without opening browser (for scripting)
  stackeye billing portal --url-only

  # Output as JSON for automation
  stackeye billing portal --url-only -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBillingPortal(cmd.Context())
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&portalFlags.urlOnly, "url-only", false, "Print portal URL without opening browser")

	return cmd
}

// runBillingPortal executes the billing portal command logic.
func runBillingPortal(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to create portal session with timeout
	reqCtx, cancel := context.WithTimeout(ctx, billingPortalTimeout)
	defer cancel()

	// Create portal session request with a return URL
	// The return URL is a placeholder since CLI doesn't have a web callback
	req := &client.CreatePortalSessionRequest{
		ReturnURL: "https://app.stackeye.io/settings/billing",
	}

	response, err := client.CreatePortalSession(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create billing portal session: %w", err)
	}

	// Check output format - use JSON/YAML if requested
	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	// URL-only mode: just print the URL
	if portalFlags.urlOnly {
		fmt.Println(response.SessionURL)
		return nil
	}

	// Default mode: open in browser
	fmt.Println()
	fmt.Println("Opening Stripe billing portal in your browser...")
	fmt.Println()

	if err := browserOpener(response.SessionURL); err != nil {
		// If browser fails to open, fall back to printing URL
		fmt.Println("Could not open browser automatically.")
		fmt.Println()
		fmt.Println("Please open this URL manually:")
		fmt.Printf("  %s\n", response.SessionURL)
		fmt.Println()
		return nil
	}

	fmt.Println("Portal URL:")
	fmt.Printf("  %s\n", response.SessionURL)
	fmt.Println()
	fmt.Println("If the browser did not open, copy the URL above and paste it in your browser.")
	fmt.Println()

	return nil
}
