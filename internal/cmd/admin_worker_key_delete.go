// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
	"github.com/spf13/cobra"
)

// adminWorkerKeyDeleteTimeout is the maximum time to wait for the API response.
const adminWorkerKeyDeleteTimeout = 30 * time.Second

// workerKeyDeleteForce is the force flag value (skip confirmation).
var workerKeyDeleteForce bool

// NewAdminWorkerKeyDeleteCmd creates and returns the worker-key delete command.
func NewAdminWorkerKeyDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <key-id>",
		Short: "Delete a worker key",
		Long: `Permanently delete a worker authentication key.

This command permanently removes a worker key from the system. Workers
using this key will immediately fail authentication on their next request.

WARNING: This action is irreversible. For audit trail purposes, consider
using the deactivate command instead, which preserves the key record.

A confirmation prompt will be shown unless --force is specified.

Examples:
  # Delete a worker key (with confirmation prompt)
  stackeye admin worker-key delete 550e8400-e29b-41d4-a716-446655440000

  # Delete without confirmation (for scripting)
  stackeye admin worker-key delete 550e8400-e29b-41d4-a716-446655440000 --force

  # Using short flags
  stackeye admin wk delete <key-id> -f`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdminWorkerKeyDelete(cmd.Context(), args[0])
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&workerKeyDeleteForce, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

// runAdminWorkerKeyDelete executes the worker-key delete command logic.
func runAdminWorkerKeyDelete(ctx context.Context, keyID string) error {
	// Dry-run check: after arg parsing, before API calls
	if GetDryRun() {
		dryrun.PrintAction("delete", "worker key",
			"ID", keyID,
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Confirmation prompt (unless --force is set)
	if !workerKeyDeleteForce {
		confirmed, promptErr := confirmWorkerKeyDeletion(keyID)
		if promptErr != nil {
			return fmt.Errorf("failed to read confirmation: %w", promptErr)
		}
		if !confirmed {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// Call SDK to delete worker key with timeout
	reqCtx, cancel := context.WithTimeout(ctx, adminWorkerKeyDeleteTimeout)
	defer cancel()

	err = admin.DeleteWorkerKey(reqCtx, apiClient, keyID)
	if err != nil {
		return fmt.Errorf("failed to delete worker key: %w", err)
	}

	// Print success message
	printWorkerKeyDeleted(keyID)
	return nil
}

// confirmWorkerKeyDeletion prompts the user to confirm deletion.
func confirmWorkerKeyDeletion(keyID string) (bool, error) {
	fmt.Printf("Are you sure you want to permanently delete worker key %s?\n", keyID)
	fmt.Print("This action cannot be undone. Type 'yes' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "yes", nil
}

// printWorkerKeyDeleted formats and prints a success message after deletion.
func printWorkerKeyDeleted(keyID string) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              WORKER KEY DELETED SUCCESSFULLY               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Key ID: %s\n", keyID)
	fmt.Println()
	fmt.Println("  The worker key has been permanently deleted.")
	fmt.Println("  Workers using this key will fail authentication immediately.")
	fmt.Println()
}
