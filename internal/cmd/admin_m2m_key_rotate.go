// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
	"github.com/spf13/cobra"
)

// adminM2MKeyRotateTimeout is the maximum time to wait for the API response.
const adminM2MKeyRotateTimeout = 30 * time.Second

// NewAdminM2MKeyRotateCmd creates and returns the m2m-key rotate command.
func NewAdminM2MKeyRotateCmd() *cobra.Command {
	var keyID string

	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate an M2M key",
		Long: `Rotate an M2M key by creating a new replacement key.

The old key is marked for expiration with a 24-hour grace period,
and a new key is created inheriting the region from the old key.
During the grace period, both keys are valid to allow smooth transitions.

IMPORTANT: The new plaintext key is only displayed once at creation time.
Store it securely as it cannot be retrieved later.

Examples:
  # Rotate an M2M key
  stackeye admin m2m-key rotate --id abc12345-6789-def0-1234-567890abcdef

  # Output in JSON format
  stackeye admin m2m-key rotate --id abc12345 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdminM2MKeyRotate(cmd.Context(), keyID)
		},
	}

	cmd.Flags().StringVarP(&keyID, "id", "i", "", "M2M key ID to rotate (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// runAdminM2MKeyRotate executes the m2m-key rotate command logic.
func runAdminM2MKeyRotate(ctx context.Context, keyID string) error {
	// Dry-run check: after flag parsing (cobra validates required flags), before API calls
	if GetDryRun() {
		dryrun.PrintAction("rotate", "M2M key",
			"ID", keyID,
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to rotate M2M key with timeout
	reqCtx, cancel := context.WithTimeout(ctx, adminM2MKeyRotateTimeout)
	defer cancel()

	response, err := admin.RotateM2MKey(reqCtx, apiClient, keyID)
	if err != nil {
		return fmt.Errorf("failed to rotate M2M key: %w", err)
	}

	// Check output format - use JSON/YAML if requested, otherwise pretty print
	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	// Pretty print for table format (default)
	printM2MKeyRotated(response)
	return nil
}

// printM2MKeyRotated formats and prints the rotation result.
func printM2MKeyRotated(response *admin.RotateM2MKeyResponse) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                  M2M KEY ROTATED                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  New Key:")
	fmt.Printf("    ID:      %s\n", response.NewKey.ID)
	fmt.Printf("    Prefix:  %s\n", response.NewKey.KeyPrefix)
	fmt.Printf("    Created: %s\n", response.NewKey.CreatedAt)
	fmt.Println()
	fmt.Println("  Plaintext Key:")
	fmt.Println("  ┌────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("    %s\n", response.NewKey.PlaintextKey)
	fmt.Println("  └────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  Old Key (expiring):")
	fmt.Printf("    ID:          %s\n", response.OldKey.ID)
	fmt.Printf("    Prefix:      %s\n", response.OldKey.KeyPrefix)
	fmt.Printf("    Expires:     %s\n", response.OldKey.ExpiresAt)
	fmt.Printf("    Replaced By: %s\n", response.OldKey.ReplacedByID)
	fmt.Println()
	fmt.Printf("  Grace Period: %s\n", response.GracePeriod)
	fmt.Println()
	fmt.Println("  ⚠ WARNING: Save the new key now - it cannot be retrieved later!")
	fmt.Println()
}
