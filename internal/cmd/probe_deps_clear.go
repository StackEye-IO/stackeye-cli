// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/spf13/cobra"
)

// probeDepsClearTimeout is the maximum time to wait for the API response.
const probeDepsClearTimeout = 30 * time.Second

// NewProbeDepsClearCmd creates and returns the probe deps clear subcommand.
// Task #8026: Implements bulk removal of probe dependencies.
func NewProbeDepsClearCmd() *cobra.Command {
	var direction string
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "clear <probe-id>",
		Short: "Remove all dependencies from a probe",
		Long: `Remove all dependencies from a probe in the specified direction(s).

The probe can be specified by UUID or by name. If the name matches multiple
probes, you'll be prompted to use the UUID instead.

This command is useful for quickly resetting a probe's dependency configuration
or removing a probe from the dependency tree entirely.

Directions:
  parents  - Remove all probes that this probe depends on
  children - Remove all probes that depend on this probe
  both     - Remove both parent and child dependencies (default)

Examples:
  # Remove all dependencies by name (parents and children)
  stackeye probe deps clear "web-server"

  # Remove all dependencies by UUID
  stackeye probe deps clear 550e8400-e29b-41d4-a716-446655440000

  # Remove only parent dependencies
  stackeye probe deps clear "web-server" --direction parents

  # Remove only child dependencies (make this probe a leaf node)
  stackeye probe deps clear "web-server" --direction children

  # Skip confirmation prompt
  stackeye probe deps clear "web-server" --yes

Note: Removing dependencies may cause alerts that were suppressed to become active.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeDepsClearCmd(cmd.Context(), args[0], direction, skipConfirm)
		},
	}

	cmd.Flags().StringVarP(&direction, "direction", "d", "both", "Direction to clear: parents, children, or both")
	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// runProbeDepsClearCmd executes the probe deps clear command logic.
func runProbeDepsClearCmd(ctx context.Context, probeIDArg, direction string, skipConfirm bool) error {
	// Validate direction
	if direction != "parents" && direction != "children" && direction != "both" {
		return fmt.Errorf("invalid direction %q: must be 'parents', 'children', or 'both'", direction)
	}

	// Get authenticated API client first (needed for name resolution)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Resolve probe ID (accepts UUID or name)
	probeID, err := ResolveProbeID(ctx, apiClient, probeIDArg)
	if err != nil {
		return err
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeDepsClearTimeout)
	defer cancel()

	// Get probe info for user-friendly output
	probe, err := client.GetProbe(reqCtx, apiClient, probeID, "")
	if err != nil {
		if isNotFoundError(err) {
			return fmt.Errorf("probe %q not found", probeIDArg)
		}
		return fmt.Errorf("failed to get probe: %w", err)
	}
	probeName := probe.Name

	// Get current dependencies
	reqCtx, cancel = context.WithTimeout(ctx, probeDepsClearTimeout)
	defer cancel()

	deps, err := client.GetProbeDependencies(reqCtx, apiClient, probeID)
	if err != nil {
		return fmt.Errorf("failed to get probe dependencies: %w", err)
	}

	// Determine what we're going to remove
	var parentsToRemove []client.ProbeBasicInfo
	var childrenToRemove []client.ProbeBasicInfo

	if direction == "parents" || direction == "both" {
		parentsToRemove = deps.Parents
	}
	if direction == "children" || direction == "both" {
		childrenToRemove = deps.Children
	}

	// Check if there's anything to remove
	totalToRemove := len(parentsToRemove) + len(childrenToRemove)
	if totalToRemove == 0 {
		fmt.Printf("Probe %q has no dependencies in the specified direction(s).\n", probeName)
		return nil
	}

	// Show what will be removed and confirm
	if !skipConfirm {
		fmt.Printf("This will remove the following dependencies for %q:\n\n", probeName)

		if len(parentsToRemove) > 0 {
			fmt.Printf("Parent dependencies (%d):\n", len(parentsToRemove))
			for _, p := range parentsToRemove {
				fmt.Printf("  - %s (%s)\n", p.Name, p.ID)
			}
			fmt.Println()
		}

		if len(childrenToRemove) > 0 {
			fmt.Printf("Child dependencies (%d):\n", len(childrenToRemove))
			for _, c := range childrenToRemove {
				fmt.Printf("  - %s (%s)\n", c.Name, c.ID)
			}
			fmt.Println()
		}

		confirmed, confirmErr := interactive.AskConfirm(&interactive.ConfirmPromptOptions{
			Message: fmt.Sprintf("Remove %d dependencies?", totalToRemove),
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

	// Remove parent dependencies
	var removedCount int
	var errors []string

	for _, parent := range parentsToRemove {
		reqCtx, cancel = context.WithTimeout(ctx, probeDepsClearTimeout)
		_, removeErr := client.RemoveProbeDependency(reqCtx, apiClient, probeID, parent.ID)
		cancel()

		if removeErr != nil {
			errors = append(errors, fmt.Sprintf("failed to remove parent %q: %v", parent.Name, removeErr))
		} else {
			removedCount++
		}
	}

	// Remove child dependencies (remove this probe as parent from each child)
	for _, child := range childrenToRemove {
		reqCtx, cancel = context.WithTimeout(ctx, probeDepsClearTimeout)
		_, removeErr := client.RemoveProbeDependency(reqCtx, apiClient, child.ID, probeID)
		cancel()

		if removeErr != nil {
			errors = append(errors, fmt.Sprintf("failed to remove child %q: %v", child.Name, removeErr))
		} else {
			removedCount++
		}
	}

	// Report results
	if len(errors) > 0 {
		fmt.Printf("Partially completed: removed %d of %d dependencies.\n", removedCount, totalToRemove)
		fmt.Println("Errors:")
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("failed to remove %d dependencies", len(errors))
	}

	fmt.Printf("Successfully cleared %d dependencies from %q.\n", removedCount, probeName)
	return nil
}
