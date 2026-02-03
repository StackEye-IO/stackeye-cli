// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// probeWatchTimeout is the maximum time to wait for a single API response during watch.
const probeWatchTimeout = 15 * time.Second

// probeWatchFlags holds the flag values for the probe watch command.
type probeWatchFlags struct {
	interval time.Duration
	status   string
}

// NewProbeWatchCmd creates and returns the probe watch subcommand.
func NewProbeWatchCmd() *cobra.Command {
	flags := &probeWatchFlags{}

	cmd := &cobra.Command{
		Use:               "watch [id]",
		Short:             "Watch probe status with live updates",
		ValidArgsFunction: ProbeCompletion(),
		Long: `Watch probe status with live-updating display.

Polls the StackEye API at a configurable interval and displays current probe
status in a continuously refreshing table. Press Ctrl+C to stop watching.

When run without arguments, watches all probes. When a probe ID or name is
provided, watches that specific probe with additional detail.

The display automatically refreshes, showing current status, response times,
and last check timestamps. In non-interactive mode (piped output, JSON/YAML
format), a single snapshot is printed and the command exits.

Interval:
  The refresh interval controls how often the display is updated. The minimum
  interval is 1 second. Shorter intervals provide more responsive updates but
  increase API usage.

Examples:
  # Watch all probes (default 5s refresh)
  stackeye probe watch

  # Watch a specific probe by name
  stackeye probe watch "Production API"

  # Watch a specific probe by UUID
  stackeye probe watch <probe-uuid>

  # Watch with a faster refresh interval
  stackeye probe watch --interval 2s

  # Watch only probes that are down
  stackeye probe watch --status down

  # Watch with 10-second interval
  stackeye probe watch -i 10s

  # Single snapshot as JSON (non-interactive)
  stackeye probe watch -o json`,
		Aliases: []string{"w"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var idArg string
			if len(args) > 0 {
				idArg = args[0]
			}
			return runProbeWatch(cmd.Context(), idArg, flags)
		},
	}

	cmd.Flags().DurationVarP(&flags.interval, "interval", "i", 5*time.Second, "refresh interval (minimum 1s)")
	cmd.Flags().StringVarP(&flags.status, "status", "s", "", "filter by status: up, down, degraded, paused, pending")

	return cmd
}

// runProbeWatch executes the probe watch command logic.
func runProbeWatch(ctx context.Context, idArg string, flags *probeWatchFlags) error {
	// Validate interval
	if flags.interval < 1*time.Second {
		return fmt.Errorf("invalid interval %s: minimum is 1s", flags.interval)
	}

	// Validate status filter
	var probeStatus client.ProbeStatus
	if flags.status != "" {
		switch flags.status {
		case "up":
			probeStatus = client.ProbeStatusUp
		case "down":
			probeStatus = client.ProbeStatusDown
		case "degraded":
			probeStatus = client.ProbeStatusDegraded
		case "paused":
			probeStatus = client.ProbeStatusPaused
		case "pending":
			probeStatus = client.ProbeStatusPending
		default:
			return clierrors.InvalidValueError("--status", flags.status, clierrors.ValidProbeStatusFilters)
		}
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// If a specific probe is requested, resolve its ID
	if idArg != "" {
		return runProbeWatchSingle(ctx, apiClient, idArg, flags)
	}

	return runProbeWatchAll(ctx, apiClient, probeStatus, flags)
}

// runProbeWatchAll watches all probes with live updates.
func runProbeWatchAll(ctx context.Context, apiClient *client.Client, probeStatus client.ProbeStatus, flags *probeWatchFlags) error {
	// Non-interactive mode: print single snapshot and exit
	if !output.IsInteractive() {
		return fetchAndPrintAllProbes(ctx, apiClient, probeStatus)
	}

	// Interactive polling loop
	// Print initial data immediately
	if err := clearScreenAndPrintAllProbes(ctx, apiClient, probeStatus, flags.interval); err != nil {
		return err
	}

	ticker := time.NewTicker(flags.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Clean exit on Ctrl+C
			fmt.Fprintln(os.Stderr, "\nWatch stopped.")
			return nil
		case <-ticker.C:
			if err := clearScreenAndPrintAllProbes(ctx, apiClient, probeStatus, flags.interval); err != nil {
				// On API errors during watch, print error but continue watching
				fmt.Fprintf(os.Stderr, "\rError: %v (retrying...)\n", err)
			}
		}
	}
}

// runProbeWatchSingle watches a single probe with live updates.
func runProbeWatchSingle(ctx context.Context, apiClient *client.Client, idArg string, flags *probeWatchFlags) error {
	// Resolve probe ID once
	probeID, err := ResolveProbeID(ctx, apiClient, idArg)
	if err != nil {
		return err
	}

	// Non-interactive mode: print single snapshot and exit
	if !output.IsInteractive() {
		return fetchAndPrintSingleProbe(ctx, apiClient, probeID.String())
	}

	// Interactive polling loop
	if err := clearScreenAndPrintSingleProbe(ctx, apiClient, probeID.String(), flags.interval); err != nil {
		return err
	}

	ticker := time.NewTicker(flags.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "\nWatch stopped.")
			return nil
		case <-ticker.C:
			if err := clearScreenAndPrintSingleProbe(ctx, apiClient, probeID.String(), flags.interval); err != nil {
				fmt.Fprintf(os.Stderr, "\rError: %v (retrying...)\n", err)
			}
		}
	}
}

// clearScreenAndPrintAllProbes clears the terminal and prints all probes.
func clearScreenAndPrintAllProbes(ctx context.Context, apiClient *client.Client, probeStatus client.ProbeStatus, interval time.Duration) error {
	opts := &client.ListProbesOptions{
		Page:   1,
		Limit:  100,
		Status: probeStatus,
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeWatchTimeout)
	defer cancel()

	result, err := client.ListProbes(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to list probes: %w", err)
	}

	// Clear screen and move cursor to top
	fmt.Fprint(os.Stdout, "\033[2J\033[H")

	// Print header with timestamp
	fmt.Fprintf(os.Stdout, "Every %s: stackeye probe watch    %s\n\n",
		interval, time.Now().Format("2006-01-02 15:04:05"))

	if len(result.Probes) == 0 {
		fmt.Fprintln(os.Stdout, "No probes found.")
		return nil
	}

	return output.PrintProbes(result.Probes)
}

// clearScreenAndPrintSingleProbe clears the terminal and prints a single probe.
func clearScreenAndPrintSingleProbe(ctx context.Context, apiClient *client.Client, probeID string, interval time.Duration) error {
	reqCtx, cancel := context.WithTimeout(ctx, probeWatchTimeout)
	defer cancel()

	probe, err := client.GetProbe(reqCtx, apiClient, mustParseUUID(probeID), "24h")
	if err != nil {
		return fmt.Errorf("failed to get probe: %w", err)
	}

	// Clear screen and move cursor to top
	fmt.Fprint(os.Stdout, "\033[2J\033[H")

	// Print header with timestamp
	fmt.Fprintf(os.Stdout, "Every %s: stackeye probe watch %s    %s\n\n",
		interval, probe.Name, time.Now().Format("2006-01-02 15:04:05"))

	return output.PrintProbe(*probe)
}

// fetchAndPrintAllProbes fetches and prints all probes once (non-interactive).
func fetchAndPrintAllProbes(ctx context.Context, apiClient *client.Client, probeStatus client.ProbeStatus) error {
	opts := &client.ListProbesOptions{
		Page:   1,
		Limit:  100,
		Status: probeStatus,
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeWatchTimeout)
	defer cancel()

	result, err := client.ListProbes(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to list probes: %w", err)
	}

	if len(result.Probes) == 0 {
		return output.PrintEmpty("No probes found.")
	}

	return output.PrintProbes(result.Probes)
}

// fetchAndPrintSingleProbe fetches and prints a single probe once (non-interactive).
func fetchAndPrintSingleProbe(ctx context.Context, apiClient *client.Client, probeID string) error {
	reqCtx, cancel := context.WithTimeout(ctx, probeWatchTimeout)
	defer cancel()

	probe, err := client.GetProbe(reqCtx, apiClient, mustParseUUID(probeID), "24h")
	if err != nil {
		return fmt.Errorf("failed to get probe: %w", err)
	}

	return output.PrintProbe(*probe)
}

// mustParseUUID parses a UUID string, panicking on failure.
// Only use after the UUID has been validated (e.g., via ResolveProbeID).
func mustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("mustParseUUID: invalid UUID %q: %v", s, err))
	}
	return id
}
