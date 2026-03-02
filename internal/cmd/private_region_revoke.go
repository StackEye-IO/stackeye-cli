// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// privateRegionRevokeTimeout is the maximum time to wait for the API response.
const privateRegionRevokeTimeout = 30 * time.Second

// NewPrivateRegionRevokeCmd creates and returns the private-region revoke command.
func NewPrivateRegionRevokeCmd() *cobra.Command {
	var regionID string
	var keyID string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a single bootstrap key for a private region",
		Long: `Revoke a specific bootstrap key by its UUID.

The appliance using this key will lose connectivity immediately.
This is useful for rotating out a single compromised key without
disturbing other keys.

To replace all keys at once, use 'rotate' instead.

Examples:
  # Revoke a specific key
  stackeye private-region revoke \
      --region-id prv-nyc-office \
      --key-id a1b2c3d4-e5f6-7890-abcd-ef1234567890

  # Preview without revoking
  stackeye private-region revoke \
      --region-id prv-nyc-office \
      --key-id a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
      --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrivateRegionRevoke(cmd.Context(), regionID, keyID)
		},
	}

	cmd.Flags().StringVarP(&regionID, "region-id", "r", "", "Private region ID (e.g., prv-nyc-office) (required)")
	cmd.Flags().StringVarP(&keyID, "key-id", "k", "", "Bootstrap key UUID to revoke (required)")
	_ = cmd.MarkFlagRequired("region-id")
	_ = cmd.MarkFlagRequired("key-id")

	return cmd
}

// runPrivateRegionRevoke executes the private-region revoke command logic.
func runPrivateRegionRevoke(ctx context.Context, regionID, keyID string) error {
	if GetDryRun() {
		dryrun.PrintAction("revoke", "private region bootstrap key",
			"Region ID", regionID,
			"Key ID", keyID,
		)
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, privateRegionRevokeTimeout)
	defer cancel()

	if err := client.RevokePrivateRegionKey(reqCtx, apiClient, regionID, keyID); err != nil {
		return fmt.Errorf("failed to revoke bootstrap key: %w", err)
	}

	fmt.Println()
	fmt.Println("  Bootstrap key revoked successfully.")
	fmt.Printf("  Key ID: %s\n", keyID)
	fmt.Println()
	fmt.Println("  The appliance using this key has lost connectivity.")
	fmt.Println()

	return nil
}
