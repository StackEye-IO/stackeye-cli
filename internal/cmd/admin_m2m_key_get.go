// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
	"github.com/spf13/cobra"
)

// adminM2MKeyGetTimeout is the maximum time to wait for the API response.
const adminM2MKeyGetTimeout = 30 * time.Second

// NewAdminM2MKeyGetCmd creates and returns the m2m-key get command.
func NewAdminM2MKeyGetCmd() *cobra.Command {
	var keyID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a specific M2M key",
		Long: `Get detailed information about a specific M2M key.

Displays all metadata including ID, region, prefix, status, last seen
timestamp, and creation/update times. The plaintext key is never shown.

Examples:
  # Get M2M key details by ID
  stackeye admin m2m-key get --id abc12345-6789-def0-1234-567890abcdef

  # Output in JSON format
  stackeye admin m2m-key get --id abc12345 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdminM2MKeyGet(cmd.Context(), keyID)
		},
	}

	cmd.Flags().StringVarP(&keyID, "id", "i", "", "M2M key ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// runAdminM2MKeyGet executes the m2m-key get command logic.
func runAdminM2MKeyGet(ctx context.Context, keyID string) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get M2M key with timeout
	reqCtx, cancel := context.WithTimeout(ctx, adminM2MKeyGetTimeout)
	defer cancel()

	key, err := admin.GetM2MKey(reqCtx, apiClient, keyID)
	if err != nil {
		return fmt.Errorf("failed to get M2M key: %w", err)
	}

	// Check output format - use JSON/YAML if requested, otherwise pretty print
	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(key)
		}
	}

	// Pretty print for table format (default)
	printM2MKeyDetail(key)
	return nil
}

// printM2MKeyDetail formats and prints detailed M2M key information.
func printM2MKeyDetail(key *admin.M2MKey) {
	region := key.Region
	if region == "" {
		region = "global"
	}

	status := "Active"
	if !key.IsActive {
		status = "Inactive"
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    M2M KEY DETAILS                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  ID:         %s\n", key.ID)
	fmt.Printf("  Region:     %s\n", region)
	fmt.Printf("  Prefix:     %s\n", key.KeyPrefix)
	fmt.Printf("  Type:       %s\n", key.KeyType)
	fmt.Printf("  Status:     %s\n", status)
	fmt.Println()
	fmt.Println("  Activity:")
	if key.LastSeenAt != nil && !key.LastSeenAt.IsZero() {
		fmt.Printf("    Last Seen: %s (%s)\n",
			key.LastSeenAt.Format(time.RFC3339),
			formatM2MKeyLastSeen(key.LastSeenAt))
	} else {
		fmt.Println("    Last Seen: Never")
	}
	fmt.Printf("    Created:   %s\n", key.CreatedAt.Format(time.RFC3339))
	fmt.Println()
}
