// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// probeDepsAddTimeout is the maximum time to wait for the API response.
const probeDepsAddTimeout = 30 * time.Second

// NewProbeDepsAddCmd creates and returns the probe deps add subcommand.
func NewProbeDepsAddCmd() *cobra.Command {
	var parentID string
	var force bool

	cmd := &cobra.Command{
		Use:               "add <probe-id> --parent <parent-probe-id>",
		Short:             "Add a parent dependency to a probe",
		ValidArgsFunction: ProbeCompletion(),
		Long: `Add a parent dependency so that when the parent probe is DOWN, the child probe
is marked as UNREACHABLE and its alerts are suppressed.

Probes can be specified by UUID or by name. If a name matches multiple probes,
you'll be prompted to use the UUID instead.

This is useful for reducing alert noise when a dependent service is down.
For example, if your database goes down, you don't want separate alerts
for every web server that depends on it.

Examples:
  # Add a dependency by name: web-server depends on database
  stackeye probe deps add "web-server" --parent "database"

  # Add a dependency by UUID
  stackeye probe deps add 550e8400-e29b-41d4-a716-446655440000 --parent 660e8400-e29b-41d4-a716-446655440001

  # Force add even if parent is currently DOWN
  stackeye probe deps add "web-server" --parent "database" --force

Common dependency patterns:
  Database -> Application Servers
  Load Balancer -> Backend Servers
  Core Router -> All Downstream Devices`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeDepsAddCmd(cmd.Context(), args[0], parentID, force)
		},
	}

	cmd.Flags().StringVarP(&parentID, "parent", "p", "", "Parent probe ID (required)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation when parent is DOWN")
	if err := cmd.MarkFlagRequired("parent"); err != nil {
		panic(fmt.Sprintf("failed to mark parent flag as required: %v", err))
	}

	return cmd
}

// runProbeDepsAddCmd executes the probe deps add command logic.
func runProbeDepsAddCmd(ctx context.Context, probeIDArg, parentIDArg string, force bool) error {
	// Get authenticated API client first (needed for name resolution)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Resolve child probe ID (accepts UUID or name)
	probeID, err := ResolveProbeID(ctx, apiClient, probeIDArg)
	if err != nil {
		return fmt.Errorf("failed to resolve probe: %w", err)
	}

	// Resolve parent probe ID (accepts UUID or name)
	parentID, err := ResolveProbeID(ctx, apiClient, parentIDArg)
	if err != nil {
		return fmt.Errorf("failed to resolve parent probe: %w", err)
	}

	// Cannot depend on itself
	if probeID == parentID {
		return fmt.Errorf("a probe cannot depend on itself")
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeDepsAddTimeout)
	defer cancel()

	// Get probe names for user-friendly output
	var probeName, parentName string
	probe, err := client.GetProbe(reqCtx, apiClient, probeID, "")
	if err != nil {
		// API will return proper error, but we can provide better context
		if isNotFoundError(err) {
			return fmt.Errorf("probe %q not found", probeIDArg)
		}
		return fmt.Errorf("failed to get probe: %w", err)
	}
	probeName = probe.Name

	// Reset context timeout for parent lookup
	reqCtx, cancel = context.WithTimeout(ctx, probeDepsAddTimeout)
	defer cancel()

	parent, err := client.GetProbe(reqCtx, apiClient, parentID, "")
	if err != nil {
		if isNotFoundError(err) {
			return fmt.Errorf("parent probe %q not found", parentIDArg)
		}
		return fmt.Errorf("failed to get parent probe: %w", err)
	}
	parentName = parent.Name

	// Warn if parent is currently DOWN (child would immediately become unreachable)
	if strings.ToLower(parent.Status) == "down" && !force {
		fmt.Printf("Warning: Parent probe %q is currently DOWN.\n", parentName)
		fmt.Printf("Adding this dependency will immediately mark %q as UNREACHABLE.\n\n", probeName)
		fmt.Printf("Use --force to proceed anyway, or resolve the parent probe issue first.\n")
		return fmt.Errorf("parent probe is DOWN (use --force to override)")
	}

	// Reset context timeout for add operation
	reqCtx, cancel = context.WithTimeout(ctx, probeDepsAddTimeout)
	defer cancel()

	// Add the dependency
	_, err = client.AddProbeDependency(reqCtx, apiClient, probeID, parentID)
	if err != nil {
		return handleAddDependencyError(err, probeName, parentName)
	}

	fmt.Printf("Dependency added: %q now depends on %q\n", probeName, parentName)
	return nil
}

// handleAddDependencyError maps API errors to user-friendly error messages.
func handleAddDependencyError(err error, probeName, parentName string) error {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "cyclic_dependency"):
		return fmt.Errorf("circular dependency detected: adding this dependency would create a cycle")
	case strings.Contains(errMsg, "dependency_exists"):
		return fmt.Errorf("dependency already exists: %q already depends on %q", probeName, parentName)
	case strings.Contains(errMsg, "same_probe"):
		return fmt.Errorf("a probe cannot depend on itself")
	case strings.Contains(errMsg, "probe_not_found"):
		return fmt.Errorf("one or more probes not found")
	default:
		return fmt.Errorf("failed to add dependency: %w", err)
	}
}

// isNotFoundError checks if an error indicates a resource was not found.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "404") ||
		errors.Is(err, client.ErrNotFound)
}
