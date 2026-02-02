// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/spf13/cobra"
)

// probeDepsListTimeout is the maximum time to wait for the API response.
const probeDepsListTimeout = 30 * time.Second

// NewProbeDepsListCmd creates and returns the probe deps list subcommand.
func NewProbeDepsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <probe-id>",
		Short: "List dependencies for a probe",
		Long: `List parent and child dependencies for a specific probe.

The probe can be specified by UUID or by name. If the name matches multiple
probes, you'll be prompted to use the UUID instead.

Parents are probes that this probe depends on. When a parent is DOWN,
this probe will be marked as UNREACHABLE and its alerts are suppressed.

Children are probes that depend on this probe. When this probe is DOWN,
its children will be marked as UNREACHABLE.

Examples:
  # List dependencies by name
  stackeye probe deps list "web-server"

  # List dependencies by UUID
  stackeye probe deps list 550e8400-e29b-41d4-a716-446655440000

  # Output as JSON for scripting
  stackeye probe deps list "web-server" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeDepsListCmd(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runProbeDepsListCmd executes the probe deps list command logic.
func runProbeDepsListCmd(ctx context.Context, probeIDArg string) error {
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

	// Call SDK to get probe dependencies with timeout
	reqCtx, cancel := context.WithTimeout(ctx, probeDepsListTimeout)
	defer cancel()

	deps, err := client.GetProbeDependencies(reqCtx, apiClient, probeID)
	if err != nil {
		return fmt.Errorf("failed to get probe dependencies: %w", err)
	}

	// Check output format - use the printer to determine current format
	printer := output.NewPrinter(nil)
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return output.Print(deps)
	}

	// Print table format
	return printDependenciesTable(deps)
}

// printDependenciesTable prints dependencies in a human-readable table format.
func printDependenciesTable(deps *client.ProbeDependencyInfo) error {
	// Get color manager from the printer configuration
	colorMgr := getColorManager()

	// Print parents section
	if len(deps.Parents) == 0 {
		fmt.Println("PARENTS (0):")
		fmt.Println("  No parent dependencies")
	} else {
		fmt.Printf("PARENTS (%d):\n", len(deps.Parents))
		fmt.Printf("  %-36s  %-30s  %s\n", "ID", "NAME", "STATUS")
		for _, p := range deps.Parents {
			status := formatDependencyStatus(p.Status, p.IsUnreachable, colorMgr)
			// Truncate name if too long
			name := p.Name
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			fmt.Printf("  %-36s  %-30s  %s\n", p.ID, name, status)
		}
	}

	fmt.Println()

	// Print children section
	if len(deps.Children) == 0 {
		fmt.Println("CHILDREN (0):")
		fmt.Println("  No child dependencies")
	} else {
		fmt.Printf("CHILDREN (%d):\n", len(deps.Children))
		fmt.Printf("  %-36s  %-30s  %s\n", "ID", "NAME", "STATUS")
		for _, c := range deps.Children {
			status := formatDependencyStatus(c.Status, c.IsUnreachable, colorMgr)
			// Truncate name if too long
			name := c.Name
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			fmt.Printf("  %-36s  %-30s  %s\n", c.ID, name, status)
		}
	}

	return nil
}

// getColorManager returns a ColorManager based on CLI configuration.
func getColorManager() *sdkoutput.ColorManager {
	// Use auto color mode - respects --no-color flag and terminal detection
	return sdkoutput.NewColorManager(sdkoutput.ColorAuto)
}

// formatDependencyStatus formats the probe status for display with color support.
// Shows UNREACHABLE in uppercase if the probe is unreachable due to parent being DOWN.
// Colors: UP=green, DOWN=red, UNREACHABLE=yellow, others=plain
func formatDependencyStatus(status string, isUnreachable bool, colorMgr *sdkoutput.ColorManager) string {
	if isUnreachable {
		return colorMgr.StatusColor("UNREACHABLE")
	}
	return colorMgr.StatusColor(strings.ToUpper(status))
}
