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

// enrollmentKeyRotateTimeout is the maximum time to wait for the API response.
const enrollmentKeyRotateTimeout = 30 * time.Second

// NewEnrollmentKeyRotateCmd creates and returns the enrollment-key rotate subcommand.
func NewEnrollmentKeyRotateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate <id>",
		Short: "Rotate a Station enrollment key",
		Long: `Rotate a Station enrollment key by its ID.

Mints a fresh se_ek_ key inheriting the old key's scoping (capabilities,
mode, environment, group) and revokes the old key in the SAME operation.
The new plaintext key is returned ONCE.

IMPORTANT: Update any bootstrap tooling (install scripts, DaemonSet
manifests) with the new key immediately — the old key stops authenticating
new enrollments as soon as rotation completes. Already-enrolled Stations
are untouched.

Examples:
  # Rotate a key
  stackeye enrollment-key rotate a1b2c3d4-e5f6-7890-abcd-ef1234567890

  # Preview without rotating
  stackeye enrollment-key rotate a1b2c3d4-e5f6-7890-abcd-ef1234567890 --dry-run

  # JSON output for scripting
  stackeye enrollment-key rotate a1b2c3d4-e5f6-7890-abcd-ef1234567890 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnrollmentKeyRotate(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runEnrollmentKeyRotate executes the enrollment-key rotate command logic.
func runEnrollmentKeyRotate(ctx context.Context, keyID string) error {
	if GetDryRun() {
		dryrun.PrintAction("rotate", "enrollment key",
			"ID", keyID,
		)
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, enrollmentKeyRotateTimeout)
	defer cancel()

	result, err := client.RotateEnrollmentKey(reqCtx, apiClient, keyID)
	if err != nil {
		return fmt.Errorf("failed to rotate enrollment key: %w", err)
	}

	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(result)
		}
	}

	printEnrollmentKeyMinted("ENROLLMENT KEY ROTATED", result)
	return nil
}
