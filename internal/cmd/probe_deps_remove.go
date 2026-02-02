// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// probeDepsRemoveTimeout is the maximum time to wait for the API response.
const probeDepsRemoveTimeout = 30 * time.Second

// NewProbeDepsRemoveCmd creates and returns the probe deps remove subcommand.
// Task #8025: Implements removal of probe dependencies.
func NewProbeDepsRemoveCmd() *cobra.Command {
	var parentID string
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "remove <probe-id> --parent <parent-probe-id>",
		Short: "Remove a parent dependency from a probe",
		Long: `Remove a parent dependency so that the child probe will no longer be
marked as UNREACHABLE when the parent is DOWN.

After removal, the child probe will be monitored independently and will
generate its own alerts regardless of the parent probe's status.

Examples:
  # Remove a dependency: web-server no longer depends on database
  stackeye probe deps remove 550e8400-e29b-41d4-a716-446655440000 --parent 660e8400-e29b-41d4-a716-446655440001

  # Skip confirmation prompt
  stackeye probe deps remove <probe-id> --parent <parent-id> --yes

Note: If the probe has a suppressed alert due to this parent being DOWN,
the alert will become active after the dependency is removed.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeDepsRemoveCmd(cmd.Context(), args[0], parentID, skipConfirm)
		},
	}

	cmd.Flags().StringVarP(&parentID, "parent", "p", "", "Parent probe ID to remove (required)")
	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
	if err := cmd.MarkFlagRequired("parent"); err != nil {
		panic(fmt.Sprintf("failed to mark parent flag as required: %v", err))
	}

	return cmd
}

// runProbeDepsRemoveCmd executes the probe deps remove command logic.
func runProbeDepsRemoveCmd(ctx context.Context, probeIDArg, parentIDArg string, skipConfirm bool) error {
	// Parse and validate child probe UUID
	probeID, err := uuid.Parse(probeIDArg)
	if err != nil {
		return fmt.Errorf("invalid probe ID %q: must be a valid UUID", probeIDArg)
	}

	// Parse and validate parent probe UUID
	parentID, err := uuid.Parse(parentIDArg)
	if err != nil {
		return fmt.Errorf("invalid parent probe ID %q: must be a valid UUID", parentIDArg)
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeDepsRemoveTimeout)
	defer cancel()

	// Get probe names for user-friendly output
	probe, err := client.GetProbe(reqCtx, apiClient, probeID, "")
	if err != nil {
		if isNotFoundError(err) {
			return fmt.Errorf("probe %q not found", probeIDArg)
		}
		return fmt.Errorf("failed to get probe: %w", err)
	}
	probeName := probe.Name

	// Reset context timeout for parent lookup
	reqCtx, cancel = context.WithTimeout(ctx, probeDepsRemoveTimeout)
	defer cancel()

	parent, err := client.GetProbe(reqCtx, apiClient, parentID, "")
	if err != nil {
		if isNotFoundError(err) {
			return fmt.Errorf("parent probe %q not found", parentIDArg)
		}
		return fmt.Errorf("failed to get parent probe: %w", err)
	}
	parentName := parent.Name

	// Check if parent is currently DOWN and probe might have suppressed alerts
	if strings.ToLower(parent.Status) == "down" && !skipConfirm {
		fmt.Printf("Warning: Parent probe %q is currently DOWN.\n", parentName)
		fmt.Printf("If %q has suppressed alerts, they will become active after removal.\n\n", probeName)

		confirmed, confirmErr := interactive.AskConfirm(&interactive.ConfirmPromptOptions{
			Message: "Remove this dependency?",
			Default: false,
		})
		if confirmErr != nil {
			return fmt.Errorf("failed to get confirmation: %w", confirmErr)
		}
		if !confirmed {
			fmt.Println("Operation cancelled.")
			return nil
		}
	} else if !skipConfirm {
		// Normal confirmation
		confirmed, confirmErr := interactive.AskConfirm(&interactive.ConfirmPromptOptions{
			Message: fmt.Sprintf("Remove dependency: %q will no longer depend on %q?", probeName, parentName),
			Default: false,
		})
		if confirmErr != nil {
			return fmt.Errorf("failed to get confirmation: %w", confirmErr)
		}
		if !confirmed {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// Reset context timeout for remove operation
	reqCtx, cancel = context.WithTimeout(ctx, probeDepsRemoveTimeout)
	defer cancel()

	// Remove the dependency
	_, err = client.RemoveProbeDependency(reqCtx, apiClient, probeID, parentID)
	if err != nil {
		return handleRemoveDependencyError(err, probeName, parentName)
	}

	fmt.Printf("Dependency removed: %q no longer depends on %q\n", probeName, parentName)
	return nil
}

// handleRemoveDependencyError maps API errors to user-friendly error messages.
func handleRemoveDependencyError(err error, probeName, parentName string) error {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "dependency_not_found") || strings.Contains(errMsg, "not found"):
		return fmt.Errorf("dependency not found: %q does not depend on %q", probeName, parentName)
	case strings.Contains(errMsg, "probe_not_found"):
		return fmt.Errorf("one or more probes not found")
	default:
		return fmt.Errorf("failed to remove dependency: %w", err)
	}
}
