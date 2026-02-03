// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// probeGetTimeout is the maximum time to wait for the API response.
const probeGetTimeout = 30 * time.Second

// Note: Probe ID resolution (UUID or name) is handled by ResolveProbeID in resolve.go

// probeGetFlags holds the flag values for the probe get command.
type probeGetFlags struct {
	period string
}

// NewProbeGetCmd creates and returns the probe get subcommand.
func NewProbeGetCmd() *cobra.Command {
	flags := &probeGetFlags{}

	cmd := &cobra.Command{
		Use:               "get <id>",
		Short:             "Get details of a monitoring probe",
		ValidArgsFunction: ProbeCompletion(),
		Long: `Get detailed information about a specific monitoring probe.

Displays the full probe configuration including target URL, check type, interval,
regions, expected status codes, SSL settings, and current status.

Use the --period flag to include uptime statistics for the specified time range.

The probe can be specified by UUID or by name. If the name matches multiple
probes, you'll be prompted to use the UUID instead.

Examples:
  # Get probe details by ID
  stackeye probe get 550e8400-e29b-41d4-a716-446655440000

  # Get probe details by name
  stackeye probe get "Production API"

  # Get probe with 7-day uptime statistics
  stackeye probe get "Production API" --period 7d

  # Output as JSON for scripting
  stackeye probe get 550e8400-e29b-41d4-a716-446655440000 -o json

  # Get probe with 30-day statistics in YAML format
  stackeye probe get "Production API" --period 30d -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeGet(cmd.Context(), args[0], flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVar(&flags.period, "period", "", "include uptime stats for period: 24h, 7d, 30d")

	return cmd
}

// runProbeGet executes the probe get command logic.
func runProbeGet(ctx context.Context, idArg string, flags *probeGetFlags) error {
	// Validate period flag if provided
	if flags.period != "" {
		switch flags.period {
		case "24h", "7d", "30d":
			// Valid period
		default:
			return clierrors.InvalidValueError("--period", flags.period, clierrors.ValidPeriods)
		}
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Resolve probe ID (accepts UUID or name)
	probeID, err := ResolveProbeID(ctx, apiClient, idArg)
	if err != nil {
		return err
	}

	// Call SDK to get probe with timeout
	reqCtx, cancel := context.WithTimeout(ctx, probeGetTimeout)
	defer cancel()

	probe, err := client.GetProbe(reqCtx, apiClient, probeID, flags.period)
	if err != nil {
		return fmt.Errorf("failed to get probe: %w", err)
	}

	// Print the probe using the configured output format
	return output.Print(probe)
}
