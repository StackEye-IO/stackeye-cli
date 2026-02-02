// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// probeHistoryTimeout is the maximum time to wait for the API response.
const probeHistoryTimeout = 30 * time.Second

// probeHistoryFlags holds the flag values for the probe history command.
type probeHistoryFlags struct {
	limit  int
	since  string
	region string
	status string
	page   int
}

// ProbeHistoryOutput wraps probe results for output formatting.
// This struct is exported to allow JSON/YAML serialization with proper field tags.
type ProbeHistoryOutput struct {
	ProbeID uuid.UUID           `json:"probe_id" yaml:"probe_id"`
	Results []ProbeHistoryEntry `json:"results" yaml:"results"`
	Total   int64               `json:"total" yaml:"total"`
	Page    int                 `json:"page" yaml:"page"`
	Limit   int                 `json:"limit" yaml:"limit"`
}

// ProbeHistoryEntry represents a single check result in the history output.
type ProbeHistoryEntry struct {
	ID             uuid.UUID `json:"id" yaml:"id"`
	CheckedAt      time.Time `json:"checked_at" yaml:"checked_at"`
	Region         string    `json:"region" yaml:"region"`
	Status         string    `json:"status" yaml:"status"`
	ResponseTimeMs int       `json:"response_time_ms" yaml:"response_time_ms"`
	StatusCode     *int      `json:"status_code,omitempty" yaml:"status_code,omitempty"`
	ErrorMessage   *string   `json:"error_message,omitempty" yaml:"error_message,omitempty"`
	SSLExpiryDays  *int      `json:"ssl_expiry_days,omitempty" yaml:"ssl_expiry_days,omitempty"`
}

// NewProbeHistoryCmd creates and returns the probe history subcommand.
func NewProbeHistoryCmd() *cobra.Command {
	flags := &probeHistoryFlags{}

	cmd := &cobra.Command{
		Use:   "history <id>",
		Short: "View probe check history",
		Long: `View the historical check results for a monitoring probe.

Displays recent check results including timestamp, region, status, response time,
and any error messages. Use filters to narrow down results by time range, region,
or status.

Time Range Filtering:
  The --since flag accepts duration strings like "1h", "24h", "7d", "30d".
  This filters results to only show checks from that time period.

Status Filtering:
  Use --status to filter by result status:
    success - Only show successful checks
    failure - Only show failed checks

The probe can be specified by UUID or by name. If the name matches multiple
probes, you'll be prompted to use the UUID instead.

Examples:
  # View last 20 check results by name (default)
  stackeye probe history "Production API"

  # View last 20 check results by UUID
  stackeye probe history 550e8400-e29b-41d4-a716-446655440000

  # View last 50 results
  stackeye probe history "Production API" --limit 50

  # View results from the last 24 hours
  stackeye probe history "Production API" --since 24h

  # View only failures from the last 7 days
  stackeye probe history "Production API" --since 7d --status failure

  # Filter by region
  stackeye probe history "Production API" --region us-east-1

  # Output as JSON for scripting
  stackeye probe history 550e8400-e29b-41d4-a716-446655440000 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeHistory(cmd.Context(), args[0], flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "maximum number of results to return (default 20, max 100)")
	cmd.Flags().StringVar(&flags.since, "since", "", "time range filter (e.g., 1h, 24h, 7d, 30d)")
	cmd.Flags().StringVar(&flags.region, "region", "", "filter by region")
	cmd.Flags().StringVar(&flags.status, "status", "", "filter by status: success, failure")
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")

	return cmd
}

// runProbeHistory executes the probe history command logic.
func runProbeHistory(ctx context.Context, idArg string, flags *probeHistoryFlags) error {
	// Validate limit
	if flags.limit < 1 {
		return fmt.Errorf("invalid limit %d: must be at least 1", flags.limit)
	}
	if flags.limit > 100 {
		return fmt.Errorf("invalid limit %d: maximum is 100", flags.limit)
	}

	// Validate page
	if flags.page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.page)
	}

	// Validate status flag if provided
	if flags.status != "" {
		switch flags.status {
		case "success", "failure":
			// Valid status
		default:
			return fmt.Errorf("invalid status %q: must be 'success' or 'failure'", flags.status)
		}
	}

	// Validate since flag if provided (check format but actual filtering
	// may need to be done client-side if API doesn't support it directly)
	if flags.since != "" {
		_, err := parseSinceDuration(flags.since)
		if err != nil {
			return fmt.Errorf("invalid since value %q: %w", flags.since, err)
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

	// Build options for the API call
	opts := &client.ListProbeResultsOptions{
		Page:   flags.page,
		Limit:  flags.limit,
		Region: flags.region,
		Status: flags.status,
	}

	// Call SDK to get probe results with timeout
	reqCtx, cancel := context.WithTimeout(ctx, probeHistoryTimeout)
	defer cancel()

	results, err := client.GetProbeResults(reqCtx, apiClient, probeID, opts)
	if err != nil {
		return fmt.Errorf("failed to get probe history: %w", err)
	}

	// Handle empty results
	if len(results.Results) == 0 {
		return output.PrintEmpty("No check history found for this probe")
	}

	// Convert SDK results to output format
	historyOutput := convertToHistoryOutput(probeID, results, flags)

	// Print the results using the configured output format
	return output.Print(historyOutput)
}

// parseSinceDuration parses a duration string like "1h", "24h", "7d", "30d" into a time.Duration.
// Returns the duration and any parsing error.
func parseSinceDuration(since string) (time.Duration, error) {
	// Handle day-based durations (Go's time.ParseDuration doesn't support 'd')
	if len(since) > 1 && since[len(since)-1] == 'd' {
		var days int
		_, err := fmt.Sscanf(since, "%dd", &days)
		if err != nil {
			return 0, fmt.Errorf("invalid duration format: %w", err)
		}
		if days < 1 || days > 365 {
			return 0, fmt.Errorf("days must be between 1 and 365")
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	// Use standard Go duration parsing for hours, minutes, etc.
	d, err := time.ParseDuration(since)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format (use 1h, 24h, 7d, etc.): %w", err)
	}

	if d <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}

	return d, nil
}

// convertToHistoryOutput converts SDK probe results to the CLI output format.
func convertToHistoryOutput(probeID uuid.UUID, results *client.ProbeResultListResponse, flags *probeHistoryFlags) *ProbeHistoryOutput {
	entries := make([]ProbeHistoryEntry, 0, len(results.Results))

	// Calculate cutoff time if --since flag was provided
	var cutoffTime time.Time
	if flags.since != "" {
		duration, _ := parseSinceDuration(flags.since) // Already validated
		cutoffTime = time.Now().Add(-duration)
	}

	for _, r := range results.Results {
		// Apply client-side time filtering if the API doesn't support it
		if !cutoffTime.IsZero() && r.CheckedAt.Before(cutoffTime) {
			continue
		}

		entry := ProbeHistoryEntry{
			ID:             r.ID,
			CheckedAt:      r.CheckedAt,
			Region:         r.Region,
			Status:         r.Status,
			ResponseTimeMs: r.ResponseTimeMs,
			StatusCode:     r.StatusCode,
			ErrorMessage:   r.ErrorMessage,
			SSLExpiryDays:  r.SSLExpiryDays,
		}
		entries = append(entries, entry)
	}

	return &ProbeHistoryOutput{
		ProbeID: probeID,
		Results: entries,
		Total:   results.Total,
		Page:    results.Page,
		Limit:   results.Limit,
	}
}
