// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// statusPageDomainVerifyTimeout is the maximum time to wait for the API response.
const statusPageDomainVerifyTimeout = 30 * time.Second

// NewStatusPageDomainVerifyCmd creates and returns the status-page domain-verify subcommand.
func NewStatusPageDomainVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domain-verify <id>",
		Short: "Get DNS verification record for custom domain",
		Long: `Get the DNS TXT record required to verify custom domain ownership for a status page.

When you configure a custom domain for your status page, you must prove domain
ownership by creating a DNS TXT record with the values provided by this command.

Output Fields:
  HOST               The DNS record name (e.g., _stackeye-verify.status.example.com)
  VALUE              The TXT record value to set

DNS Configuration Steps:
  1. Run this command to get the verification record
  2. Add a TXT record to your DNS provider with the provided host and value
  3. Wait for DNS propagation (typically 5-60 minutes)
  4. Update your status page to use the custom domain

Examples:
  # Get DNS verification record for a status page
  stackeye status-page domain-verify 123

  # Output as JSON for scripting
  stackeye status-page domain-verify 123 -o json

  # Output as YAML
  stackeye status-page domain-verify 123 -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusPageDomainVerify(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runStatusPageDomainVerify executes the status-page domain-verify command logic.
func runStatusPageDomainVerify(ctx context.Context, idArg string) error {
	// Parse and validate status page ID (uint)
	id, err := strconv.ParseUint(idArg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid status page ID %q: must be a positive integer", idArg)
	}
	if id == 0 {
		return fmt.Errorf("invalid status page ID: must be greater than 0")
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get domain verification record with timeout
	reqCtx, cancel := context.WithTimeout(ctx, statusPageDomainVerifyTimeout)
	defer cancel()

	verification, err := client.GetDomainVerification(reqCtx, apiClient, uint(id))
	if err != nil {
		return fmt.Errorf("failed to get domain verification record: %w", err)
	}

	// Defensive check for nil response
	if verification == nil {
		return fmt.Errorf("status page %d not found", id)
	}

	// Check if no custom domain is configured
	if verification.Host == "" && verification.Value == "" {
		return fmt.Errorf("status page %d has no custom domain configured", id)
	}

	// Print the domain verification record using the configured output format
	return output.PrintDomainVerification(*verification)
}
