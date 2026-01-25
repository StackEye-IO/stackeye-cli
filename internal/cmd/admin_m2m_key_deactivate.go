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

// adminM2MKeyDeactivateTimeout is the maximum time to wait for the API response.
const adminM2MKeyDeactivateTimeout = 30 * time.Second

// NewAdminM2MKeyDeactivateCmd creates and returns the m2m-key deactivate command.
func NewAdminM2MKeyDeactivateCmd() *cobra.Command {
	var keyID string

	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate an M2M key immediately",
		Long: `Deactivate an M2M key with immediate effect.

The key becomes invalid immediately - services using this key will fail
authentication on their next request. Unlike rotation, there is no grace
period. The key record is preserved for audit purposes.

For graceful key replacement, use 'rotate' instead which provides a
grace period during which both old and new keys are valid.

Examples:
  # Deactivate an M2M key
  stackeye admin m2m-key deactivate --id abc12345-6789-def0-1234-567890abcdef

  # Output in JSON format
  stackeye admin m2m-key deactivate --id abc12345 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdminM2MKeyDeactivate(cmd.Context(), keyID)
		},
	}

	cmd.Flags().StringVarP(&keyID, "id", "i", "", "M2M key ID to deactivate (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// runAdminM2MKeyDeactivate executes the m2m-key deactivate command logic.
func runAdminM2MKeyDeactivate(ctx context.Context, keyID string) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to deactivate M2M key with timeout
	reqCtx, cancel := context.WithTimeout(ctx, adminM2MKeyDeactivateTimeout)
	defer cancel()

	response, err := admin.DeactivateM2MKey(reqCtx, apiClient, keyID)
	if err != nil {
		return fmt.Errorf("failed to deactivate M2M key: %w", err)
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
	printM2MKeyDeactivated(response)
	return nil
}

// printM2MKeyDeactivated formats and prints the deactivation confirmation.
func printM2MKeyDeactivated(response *admin.DeactivateM2MKeyResponse) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                 M2M KEY DEACTIVATED                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  %s\n", response.Message)
	fmt.Println()
	fmt.Println("  The key is now invalid and cannot be used for authentication.")
	fmt.Println()
}
