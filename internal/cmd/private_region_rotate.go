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
	"github.com/spf13/cobra"
)

// privateRegionRotateTimeout is the maximum time to wait for the API response.
const privateRegionRotateTimeout = 30 * time.Second

// NewPrivateRegionRotateCmd creates and returns the private-region rotate command.
func NewPrivateRegionRotateCmd() *cobra.Command {
	var regionID string
	var displayName string

	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate the bootstrap key for a private region",
		Long: `Rotate all active bootstrap keys for a private monitoring region.

All existing active keys are revoked atomically and a new bootstrap key
is issued. The plaintext key is returned exactly once — it cannot be
retrieved after this response.

IMPORTANT: The appliance must be reconfigured with the new key immediately
after rotation. Until it reconnects with the new key, monitoring will be
interrupted.

Examples:
  # Rotate keys for a private region
  stackeye private-region rotate --id prv-nyc-office

  # Rotate and label the new key
  stackeye private-region rotate --id prv-nyc-office --display-name "Rotated 2026-03"

  # Preview without rotating
  stackeye private-region rotate --id prv-nyc-office --dry-run

  # Output in JSON format
  stackeye private-region rotate --id prv-nyc-office -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var displayNamePtr *string
			if cmd.Flags().Changed("display-name") {
				displayNamePtr = &displayName
			}
			return runPrivateRegionRotate(cmd.Context(), regionID, displayNamePtr)
		},
	}

	cmd.Flags().StringVarP(&regionID, "id", "i", "", "Private region ID (e.g., prv-nyc-office) (required)")
	cmd.Flags().StringVarP(&displayName, "display-name", "n", "", "Optional label for the new key")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// runPrivateRegionRotate executes the private-region rotate command logic.
func runPrivateRegionRotate(ctx context.Context, regionID string, displayName *string) error {
	if GetDryRun() {
		label := "(none)"
		if displayName != nil {
			label = *displayName
		}
		dryrun.PrintAction("rotate", "private region bootstrap key",
			"Region ID", regionID,
			"New Key Label", label,
		)
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, privateRegionRotateTimeout)
	defer cancel()

	response, err := client.RotatePrivateRegionKey(reqCtx, apiClient, regionID, displayName)
	if err != nil {
		return fmt.Errorf("failed to rotate bootstrap key: %w", err)
	}

	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	printPrivateRegionRotated(response)
	return nil
}

// printPrivateRegionRotated formats and prints the key rotation result.
func printPrivateRegionRotated(response *client.PrivateRegionRotateKeyResponse) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║             PRIVATE REGION KEY ROTATED                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	newKey := response.Data.NewKey
	fmt.Println("  New Bootstrap Key:")
	fmt.Printf("    ID:      %s\n", newKey.ID)
	fmt.Printf("    Prefix:  %s\n", newKey.KeyPrefix)
	if newKey.DisplayName != "" {
		fmt.Printf("    Label:   %s\n", newKey.DisplayName)
	}
	fmt.Println()

	if newKey.PlaintextKey != nil {
		fmt.Println("  Plaintext Key (save this now — shown only once):")
		fmt.Println("  ┌──────────────────────────────────────────────────────────────────────────────────┐")
		fmt.Printf("    %s\n", *newKey.PlaintextKey)
		fmt.Println("  └──────────────────────────────────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	if len(response.Data.RevokedKeyIDs) > 0 {
		fmt.Printf("  Revoked Keys: %d\n", len(response.Data.RevokedKeyIDs))
		for _, id := range response.Data.RevokedKeyIDs {
			fmt.Printf("    - %s\n", id)
		}
		fmt.Println()
	}

	fmt.Println("  ⚠ WARNING: Save the new key now — it cannot be retrieved later!")
	fmt.Println("  ⚠ Reconfigure your appliance immediately to restore monitoring.")
	fmt.Println()
}
