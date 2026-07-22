// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	cliinteractive "github.com/StackEye-IO/stackeye-cli/internal/interactive"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// enrollmentKeyRevokeTimeout is the maximum time to wait for a single API response.
const enrollmentKeyRevokeTimeout = 30 * time.Second

// enrollmentKeyRevokeFlags holds the flag values for the enrollment-key revoke command.
type enrollmentKeyRevokeFlags struct {
	yes bool // Skip confirmation prompt
}

// NewEnrollmentKeyRevokeCmd creates and returns the enrollment-key revoke subcommand.
func NewEnrollmentKeyRevokeCmd() *cobra.Command {
	flags := &enrollmentKeyRevokeFlags{}

	cmd := &cobra.Command{
		Use:   "revoke <id>",
		Short: "Revoke a Station enrollment key",
		Long: `Revoke a Station enrollment key by its ID.

This stops the key from authenticating NEW enrollments immediately.
Stations already enrolled with this key are UNTOUCHED — this only closes the
bootstrap door, it does not affect live Stations.

By default, the command will prompt for confirmation before revoking. Use
--yes to skip the confirmation prompt for scripting or automation.

Examples:
  # Revoke a key (with confirmation)
  stackeye enrollment-key revoke a1b2c3d4-e5f6-7890-abcd-ef1234567890

  # Revoke without confirmation
  stackeye enrollment-key revoke a1b2c3d4-e5f6-7890-abcd-ef1234567890 --yes`,
		Aliases: []string{"rm", "remove"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnrollmentKeyRevoke(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// runEnrollmentKeyRevoke executes the enrollment-key revoke command logic.
func runEnrollmentKeyRevoke(ctx context.Context, keyID string, flags *enrollmentKeyRevokeFlags) error {
	if GetDryRun() {
		dryrun.PrintAction("revoke", "enrollment key",
			"ID", keyID,
		)
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	message := fmt.Sprintf("Are you sure you want to revoke enrollment key %q? New enrollments will stop immediately.", keyID)
	confirmed, err := cliinteractive.Confirm(message, cliinteractive.WithYesFlag(flags.yes))
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Revoke cancelled.")
		return nil
	}

	reqCtx, cancel := context.WithTimeout(ctx, enrollmentKeyRevokeTimeout)
	defer cancel()

	if _, err := client.RevokeEnrollmentKey(reqCtx, apiClient, keyID); err != nil {
		return fmt.Errorf("failed to revoke enrollment key: %w", err)
	}

	fmt.Printf("Revoked enrollment key %s\n", keyID)
	return nil
}
