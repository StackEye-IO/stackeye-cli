// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/spf13/cobra"
)

// apiKeyCreateTimeout is the maximum time to wait for the API response.
const apiKeyCreateTimeout = 30 * time.Second

// apiKeyCreateFlags holds the flag values for the api-key create command.
type apiKeyCreateFlags struct {
	name        string
	permissions string
	expiresIn   string
}

// NewAPIKeyCreateCmd creates and returns the api-key create subcommand.
func NewAPIKeyCreateCmd() *cobra.Command {
	flags := &apiKeyCreateFlags{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API key",
		Long: `Create a new API key for programmatic access to the StackEye API.

IMPORTANT: The full API key is displayed ONLY ONCE after creation.
You must save it immediately - it cannot be retrieved later.

API keys use the format: se_<32_random_characters>

Required Flags:
  --name       Human-readable name for the API key

Optional Flags:
  --permissions   Comma-separated permissions (e.g., "read:probes,write:probes")
  --expires-in    Duration until expiration (e.g., "30d", "90d", "1y")

Security Best Practices:
  - Use descriptive names to identify what each key is used for
  - Create separate keys for different applications or environments
  - Store keys securely (environment variables, secret managers)
  - Set expiration dates for keys used in temporary contexts
  - Revoke keys immediately if they may have been compromised

Examples:
  # Create a basic API key
  stackeye api-key create --name "CI Pipeline"

  # Create a key with specific permissions
  stackeye api-key create --name "Read-Only Dashboard" --permissions "read:probes,read:alerts"

  # Create a key that expires in 90 days
  stackeye api-key create --name "Contractor Access" --expires-in 90d

  # Create a key with JSON output for scripting
  stackeye api-key create --name "Deploy Script" -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPIKeyCreate(cmd.Context(), flags)
		},
	}

	// Required flags
	cmd.Flags().StringVar(&flags.name, "name", "", "name for the API key (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		// This should never happen since we're marking a flag we just defined
		panic(fmt.Sprintf("failed to mark name flag as required: %v", err))
	}

	// Optional flags
	cmd.Flags().StringVar(&flags.permissions, "permissions", "", "comma-separated permissions (e.g., read:probes,write:probes)")
	cmd.Flags().StringVar(&flags.expiresIn, "expires-in", "", "duration until expiration (e.g., 30d, 90d, 1y)")

	return cmd
}

// runAPIKeyCreate executes the api-key create command logic.
func runAPIKeyCreate(ctx context.Context, flags *apiKeyCreateFlags) error {
	// Build the request
	req := &client.CreateAPIKeyRequest{
		Name:        flags.name,
		Permissions: flags.permissions,
	}

	// Parse expiration duration if provided
	if flags.expiresIn != "" {
		duration, err := parseDuration(flags.expiresIn)
		if err != nil {
			return fmt.Errorf("invalid --expires-in value: %w", err)
		}
		expiresAt := time.Now().Add(duration)
		req.ExpiresAt = &expiresAt
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		details := []string{
			"Name", flags.name,
		}
		if flags.permissions != "" {
			details = append(details, "Permissions", flags.permissions)
		}
		if flags.expiresIn != "" {
			details = append(details, "Expires In", flags.expiresIn)
		}
		dryrun.PrintAction("create", "API key", details...)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to create API key with timeout
	reqCtx, cancel := context.WithTimeout(ctx, apiKeyCreateTimeout)
	defer cancel()

	result, err := client.CreateAPIKey(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	// Print warning and the created key
	return printAPIKeyCreated(result)
}

// parseDuration parses a duration string like "30d", "90d", "1y".
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("duration too short: %q", s)
	}

	// Parse the numeric part
	var num int
	var unit string
	_, err := fmt.Sscanf(s, "%d%s", &num, &unit)
	if err != nil {
		return 0, fmt.Errorf("invalid format %q: expected number followed by unit (d, w, m, y)", s)
	}

	if num <= 0 {
		return 0, fmt.Errorf("duration must be positive: %d", num)
	}

	// Convert to duration based on unit
	switch unit {
	case "d":
		return time.Duration(num) * 24 * time.Hour, nil
	case "w":
		return time.Duration(num) * 7 * 24 * time.Hour, nil
	case "m":
		// Approximate months as 30 days
		return time.Duration(num) * 30 * 24 * time.Hour, nil
	case "y":
		// Approximate years as 365 days
		return time.Duration(num) * 365 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown unit %q: use d (days), w (weeks), m (months), or y (years)", unit)
	}
}

// truncateString truncates a string to the specified length, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// printAPIKeyCreated prints the created API key with appropriate warnings.
func printAPIKeyCreated(key *client.CreateAPIKeyResponse) error {
	printer := output.NewPrinter(nil)

	// For JSON/YAML output, just print the key data
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(key)
	}

	// For table output, print the key prominently with warnings
	// API keys are 67 chars: se_ + 64 hex characters
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                       API KEY CREATED SUCCESSFULLY                        ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                                           ║")
	fmt.Printf("║  Name:   %-63s  ║\n", truncateString(key.Name, 63))
	fmt.Printf("║  ID:     %-63s  ║\n", key.ID.String())
	fmt.Println("║                                                                           ║")
	fmt.Println("║  Your API Key:                                                            ║")
	fmt.Printf("║  %-71s  ║\n", key.Key)
	fmt.Println("║                                                                           ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  ⚠  WARNING: This key will NOT be shown again!                           ║")
	fmt.Println("║  Save it now in a secure location.                                        ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	return nil
}
