// Package cmd implements the CLI commands for StackEye.
// Task #8069
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// probeUnlabelTimeout is the maximum time to wait for each API response.
const probeUnlabelTimeout = 30 * time.Second

// NewProbeUnlabelCmd creates and returns the probe unlabel subcommand.
// Task #8069
func NewProbeUnlabelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "unlabel <probe-id> <keys...>",
		Short:             "Remove labels from a probe",
		ValidArgsFunction: ProbeCompletion(),
		Long: `Remove one or more labels from a probe by key name.

Label keys that don't exist on the probe are silently ignored (no error).
This operation does not affect other labels on the probe.

The probe can be specified by UUID or by name. If the name matches multiple
probes, you'll be prompted to use the UUID instead.

Examples:
  # Remove a single label
  stackeye probe unlabel api-health env

  # Remove multiple labels at once
  stackeye probe unlabel api-health env tier pci

  # Use probe UUID
  stackeye probe unlabel 550e8400-e29b-41d4-a716-446655440000 env`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeUnlabel(cmd.Context(), args[0], args[1:])
		},
	}

	return cmd
}

// runProbeUnlabel executes the probe unlabel command logic.
// Task #8069
func runProbeUnlabel(ctx context.Context, probeIDArg string, keys []string) error {
	if len(keys) == 0 {
		return fmt.Errorf("at least one label key is required")
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Resolve probe ID (accepts UUID or name)
	probeID, err := ResolveProbeID(ctx, apiClient, probeIDArg)
	if err != nil {
		return err
	}

	// Remove each label key sequentially
	// Per acceptance criteria: silently succeed if label not present
	for _, key := range keys {
		if err := validateLabelKeyForRemoval(key); err != nil {
			return err
		}

		reqCtx, cancel := context.WithTimeout(ctx, probeUnlabelTimeout)
		err := client.RemoveProbeLabel(reqCtx, apiClient, probeID, key)
		cancel()

		if err != nil {
			// Check if it's a "not found" error for the label - this is OK per acceptance criteria
			// The API returns 404 if the label key doesn't exist, but we want to silently succeed
			if isLabelNotFoundError(err) {
				continue // Silently succeed
			}
			return fmt.Errorf("failed to remove label %q: %w", key, err)
		}
	}

	// Print the updated labels
	reqCtx, cancel := context.WithTimeout(ctx, probeUnlabelTimeout)
	defer cancel()

	result, err := client.GetProbeLabels(reqCtx, apiClient, probeID)
	if err != nil {
		return fmt.Errorf("failed to get updated labels: %w", err)
	}

	return output.PrintProbeLabels(result.Labels)
}

// validateLabelKeyForRemoval validates a label key argument for removal.
// Less strict than creation since we just need to match existing keys.
// Task #8069
func validateLabelKeyForRemoval(key string) error {
	if key == "" {
		return fmt.Errorf("label key cannot be empty")
	}

	// Basic length check - keys can't be longer than 63 characters
	if len(key) > 63 {
		return fmt.Errorf("label key %q exceeds maximum length of 63 characters", key)
	}

	return nil
}

// isLabelNotFoundError checks if the error indicates the label was not found.
// This is used to silently succeed when removing a label that doesn't exist.
// Task #8069
func isLabelNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// The API returns 404 with code "label_not_found" when the label doesn't exist
	// Check for common patterns in the error message
	errStr := err.Error()
	return containsUnlabel(errStr, "404") ||
		containsUnlabel(errStr, "not found") ||
		containsUnlabel(errStr, "label_not_found")
}

// containsUnlabel checks if s contains substr.
// Task #8069
func containsUnlabel(s, substr string) bool {
	return len(s) >= len(substr) && searchStringUnlabel(s, substr)
}

// searchStringUnlabel is a simple substring search.
// Task #8069
func searchStringUnlabel(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
