// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// probeLogsTimeout is the maximum time to wait for a single API response.
const probeLogsTimeout = 30 * time.Second

// probeLogsFollowInterval is the default polling interval for --follow mode.
const probeLogsFollowInterval = 5 * time.Second

// probeLogsFlags holds the flag values for the probe logs command.
type probeLogsFlags struct {
	limit  int
	since  string
	until  string
	region string
	status string
	follow bool
}

// NewProbeLogsCmd creates and returns the probe logs subcommand.
func NewProbeLogsCmd() *cobra.Command {
	flags := &probeLogsFlags{}

	cmd := &cobra.Command{
		Use:               "logs <id>",
		Short:             "View recent check logs for a probe",
		ValidArgsFunction: ProbeCompletion(),
		Long: `View recent check logs for a monitoring probe.

Displays check results including timestamp, region, status, response time,
and any error messages. Results are ordered by most recent first.

Time Range Filtering:
  Use --since and --until to filter results by time range.
  Both flags accept duration strings (e.g., "1h", "24h", "7d") which are
  interpreted as relative to the current time. RFC3339 timestamps are also
  accepted for --until (e.g., "2025-01-15T10:00:00Z").

Follow Mode:
  Use --follow (-f) to continuously poll for new check results. The command
  will print new results as they arrive, similar to "tail -f". Press Ctrl+C
  to stop following.

  In non-interactive mode (piped output), --follow is ignored and a single
  batch of results is printed.

The probe can be specified by UUID or by name.

Examples:
  # View last 50 check results (default)
  stackeye probe logs "Production API"

  # View results from the last hour
  stackeye probe logs "Production API" --since 1h

  # View results from the last 7 days, limited to 100
  stackeye probe logs "Production API" --since 7d --limit 100

  # View only failures
  stackeye probe logs "Production API" --status failure

  # View results from a specific time window
  stackeye probe logs "Production API" --since 24h --until 12h

  # Follow new results as they arrive
  stackeye probe logs "Production API" --follow

  # Follow only failures in us-east-1
  stackeye probe logs "Production API" -f --status failure --region us-east-1

  # Output as JSON for scripting
  stackeye probe logs "Production API" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeLogs(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().IntVar(&flags.limit, "limit", 50, "maximum number of results to return (1-1000)")
	cmd.Flags().StringVar(&flags.since, "since", "", "show results since duration ago (e.g., 1h, 24h, 7d)")
	cmd.Flags().StringVar(&flags.until, "until", "", "show results until duration ago or RFC3339 timestamp")
	cmd.Flags().StringVar(&flags.region, "region", "", "filter by region")
	cmd.Flags().StringVar(&flags.status, "status", "", "filter by status: success, failure")
	cmd.Flags().BoolVarP(&flags.follow, "follow", "f", false, "follow new results (poll every 5s)")

	return cmd
}

// runProbeLogs executes the probe logs command logic.
func runProbeLogs(ctx context.Context, idArg string, flags *probeLogsFlags) error {
	if err := validateProbeLogsFlags(flags); err != nil {
		return err
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	probeID, err := ResolveProbeID(ctx, apiClient, idArg)
	if err != nil {
		return err
	}

	// Resolve time range from flags
	from, to, err := resolveLogsTimeRange(flags)
	if err != nil {
		return err
	}

	if flags.follow && output.IsInteractive() {
		return runProbeLogsFollow(ctx, apiClient, probeID, flags, from)
	}

	return runProbeLogsBatch(ctx, apiClient, probeID, flags, from, to)
}

// validateProbeLogsFlags validates all flag values before making API calls.
func validateProbeLogsFlags(flags *probeLogsFlags) error {
	if flags.limit < 1 || flags.limit > 1000 {
		return fmt.Errorf("invalid limit %d: must be between 1 and 1000", flags.limit)
	}

	if flags.status != "" {
		switch flags.status {
		case "success", "failure":
			// Valid
		default:
			return fmt.Errorf("invalid status %q: must be 'success' or 'failure'", flags.status)
		}
	}

	if flags.since != "" {
		if _, err := parseSinceDuration(flags.since); err != nil {
			return fmt.Errorf("invalid --since value %q: %w", flags.since, err)
		}
	}

	if flags.until != "" {
		if _, err := parseUntilValue(flags.until); err != nil {
			return fmt.Errorf("invalid --until value %q: %w", flags.until, err)
		}
	}

	return nil
}

// resolveLogsTimeRange converts --since and --until flags into absolute time values.
// Returns zero-value times when the corresponding flag is not set.
func resolveLogsTimeRange(flags *probeLogsFlags) (from, to time.Time, err error) {
	now := time.Now()

	if flags.since != "" {
		d, err := parseSinceDuration(flags.since)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = now.Add(-d)
	}

	if flags.until != "" {
		to, err = parseUntilValue(flags.until)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	// Validate that from is before to when both are set
	if !from.IsZero() && !to.IsZero() && !from.Before(to) {
		return time.Time{}, time.Time{}, fmt.Errorf("--since time (%s) must be before --until time (%s)",
			from.Format(time.RFC3339), to.Format(time.RFC3339))
	}

	return from, to, nil
}

// parseUntilValue parses the --until flag value as either a duration string
// (relative to now) or an RFC3339 timestamp.
func parseUntilValue(until string) (time.Time, error) {
	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, until); err == nil {
		return t, nil
	}

	// Try as a duration (relative to now)
	d, err := parseSinceDuration(until)
	if err != nil {
		return time.Time{}, fmt.Errorf("must be a duration (e.g., 1h, 7d) or RFC3339 timestamp: %w", err)
	}

	return time.Now().Add(-d), nil
}

// runProbeLogsBatch fetches and prints a single batch of log results.
func runProbeLogsBatch(ctx context.Context, apiClient *client.Client, probeID uuid.UUID, flags *probeLogsFlags, from, to time.Time) error {
	opts := &client.ListProbeResultsOptions{
		Page:   1,
		Limit:  flags.limit,
		Region: flags.region,
		Status: flags.status,
		From:   from,
		To:     to,
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeLogsTimeout)
	defer cancel()

	results, err := client.GetProbeResults(reqCtx, apiClient, probeID, opts)
	if err != nil {
		return fmt.Errorf("failed to get probe logs: %w", err)
	}

	if len(results.Results) == 0 {
		return output.PrintEmpty("No check logs found for this probe")
	}

	logsOutput := convertToLogsOutput(probeID, results)
	return output.Print(logsOutput)
}

// runProbeLogsFollow polls for new results and prints them incrementally.
func runProbeLogsFollow(ctx context.Context, apiClient *client.Client, probeID uuid.UUID, flags *probeLogsFlags, from time.Time) error {
	// If no --since was provided, start from now
	if from.IsZero() {
		from = time.Now()
	}

	// Print initial batch of recent results
	initialOpts := &client.ListProbeResultsOptions{
		Page:   1,
		Limit:  flags.limit,
		Region: flags.region,
		Status: flags.status,
		From:   from,
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeLogsTimeout)
	results, err := client.GetProbeResults(reqCtx, apiClient, probeID, initialOpts)
	cancel()

	if err != nil {
		return fmt.Errorf("failed to get initial probe logs: %w", err)
	}

	// Track the latest timestamp we've seen to avoid duplicates
	latestSeen := from
	if len(results.Results) > 0 {
		logsOutput := convertToLogsOutput(probeID, results)
		if err := output.Print(logsOutput); err != nil {
			return err
		}
		// Find the most recent timestamp
		for _, r := range results.Results {
			if r.CheckedAt.After(latestSeen) {
				latestSeen = r.CheckedAt
			}
		}
	}

	fmt.Fprintf(os.Stderr, "\n--- Following new results (Ctrl+C to stop) ---\n\n")

	ticker := time.NewTicker(probeLogsFollowInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "\nFollow stopped.")
			return nil
		case <-ticker.C:
			// Query for results newer than what we've already seen
			// Add 1 nanosecond to avoid re-fetching the last result
			followFrom := latestSeen.Add(1 * time.Nanosecond)

			followOpts := &client.ListProbeResultsOptions{
				Page:   1,
				Limit:  100,
				Region: flags.region,
				Status: flags.status,
				From:   followFrom,
			}

			followCtx, followCancel := context.WithTimeout(ctx, probeLogsTimeout)
			newResults, err := client.GetProbeResults(followCtx, apiClient, probeID, followOpts)
			followCancel()

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error polling: %v (retrying...)\n", err)
				continue
			}

			if len(newResults.Results) == 0 {
				continue
			}

			logsOutput := convertToLogsOutput(probeID, newResults)
			if err := output.Print(logsOutput); err != nil {
				fmt.Fprintf(os.Stderr, "Error printing: %v\n", err)
				continue
			}

			for _, r := range newResults.Results {
				if r.CheckedAt.After(latestSeen) {
					latestSeen = r.CheckedAt
				}
			}
		}
	}
}

// ProbeLogsOutput wraps probe results for the logs command output.
type ProbeLogsOutput struct {
	ProbeID uuid.UUID       `json:"probe_id" yaml:"probe_id"`
	Results []ProbeLogEntry `json:"results" yaml:"results"`
	Total   int64           `json:"total" yaml:"total"`
}

// ProbeLogEntry represents a single check result in the logs output.
type ProbeLogEntry struct {
	CheckedAt      time.Time `json:"checked_at" yaml:"checked_at"`
	Region         string    `json:"region" yaml:"region"`
	Status         string    `json:"status" yaml:"status"`
	ResponseTimeMs int       `json:"response_time_ms" yaml:"response_time_ms"`
	StatusCode     *int      `json:"status_code,omitempty" yaml:"status_code,omitempty"`
	ErrorMessage   *string   `json:"error_message,omitempty" yaml:"error_message,omitempty"`
}

// convertToLogsOutput converts SDK probe results to the logs output format.
func convertToLogsOutput(probeID uuid.UUID, results *client.ProbeResultListResponse) *ProbeLogsOutput {
	entries := make([]ProbeLogEntry, 0, len(results.Results))

	for _, r := range results.Results {
		entry := ProbeLogEntry{
			CheckedAt:      r.CheckedAt,
			Region:         r.Region,
			Status:         r.Status,
			ResponseTimeMs: r.ResponseTimeMs,
			StatusCode:     r.StatusCode,
			ErrorMessage:   r.ErrorMessage,
		}
		entries = append(entries, entry)
	}

	return &ProbeLogsOutput{
		ProbeID: probeID,
		Results: entries,
		Total:   results.Total,
	}
}
