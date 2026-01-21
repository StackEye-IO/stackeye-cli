// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// probeStatsTimeout is the maximum time to wait for the API response.
const probeStatsTimeout = 30 * time.Second

// probeStatsFlags holds the flag values for the probe stats command.
type probeStatsFlags struct {
	period string
}

// ProbeStatsOutput wraps probe statistics for output formatting.
// This struct is exported to allow JSON/YAML serialization with proper field tags.
type ProbeStatsOutput struct {
	ProbeID       uuid.UUID `json:"probe_id" yaml:"probe_id"`
	Period        string    `json:"period" yaml:"period"`
	UptimePercent float64   `json:"uptime_percent" yaml:"uptime_percent"`
	TotalChecks   int64     `json:"total_checks" yaml:"total_checks"`
	SuccessChecks int64     `json:"success_checks" yaml:"success_checks"`
	FailureChecks int64     `json:"failure_checks" yaml:"failure_checks"`
	AvgResponseMs float64   `json:"avg_response_time_ms" yaml:"avg_response_time_ms"`
	P95ResponseMs float64   `json:"p95_response_time_ms" yaml:"p95_response_time_ms"`
	P99ResponseMs float64   `json:"p99_response_time_ms" yaml:"p99_response_time_ms"`
	MinResponseMs int       `json:"min_response_time_ms" yaml:"min_response_time_ms"`
	MaxResponseMs int       `json:"max_response_time_ms" yaml:"max_response_time_ms"`
	From          time.Time `json:"from" yaml:"from"`
	To            time.Time `json:"to" yaml:"to"`
	TimeBuckets   int       `json:"time_buckets" yaml:"time_buckets"`
}

// NewProbeStatsCmd creates and returns the probe stats subcommand.
func NewProbeStatsCmd() *cobra.Command {
	flags := &probeStatsFlags{}

	cmd := &cobra.Command{
		Use:   "stats <id>",
		Short: "View probe uptime and latency statistics",
		Long: `View aggregated uptime and response time statistics for a monitoring probe.

Displays uptime percentage, average response time, and percentile latencies (p95, p99)
for the specified time period. This provides a quick overview of probe performance.

Time Periods:
  24h - Statistics for the last 24 hours (hourly buckets)
  7d  - Statistics for the last 7 days (hourly buckets)
  30d - Statistics for the last 30 days (daily buckets)

Statistics Displayed:
  Uptime        - Percentage of successful checks over total checks
  Avg Latency   - Mean response time across all checks
  P95 Latency   - 95th percentile response time (estimated from buckets)
  P99 Latency   - 99th percentile response time (estimated from buckets)
  Min/Max       - Minimum and maximum response times observed

Examples:
  # View 24-hour statistics (default)
  stackeye probe stats 550e8400-e29b-41d4-a716-446655440000

  # View 7-day statistics
  stackeye probe stats 550e8400-e29b-41d4-a716-446655440000 --period 7d

  # View 30-day statistics
  stackeye probe stats 550e8400-e29b-41d4-a716-446655440000 --period 30d

  # Output as JSON for scripting
  stackeye probe stats 550e8400-e29b-41d4-a716-446655440000 -o json

  # Output as YAML
  stackeye probe stats 550e8400-e29b-41d4-a716-446655440000 --period 7d -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeStats(cmd.Context(), args[0], flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVar(&flags.period, "period", "24h", "statistics period: 24h, 7d, 30d")

	return cmd
}

// runProbeStats executes the probe stats command logic.
func runProbeStats(ctx context.Context, idArg string, flags *probeStatsFlags) error {
	// Parse and validate UUID
	probeID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid probe ID %q: must be a valid UUID", idArg)
	}

	// Validate period flag
	aggregate, from, to, err := parsePeriodToAggregateParams(flags.period)
	if err != nil {
		return err
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get aggregated probe results with timeout
	reqCtx, cancel := context.WithTimeout(ctx, probeStatsTimeout)
	defer cancel()

	results, err := client.GetAggregatedProbeResults(reqCtx, apiClient, probeID, aggregate, from, to)
	if err != nil {
		return fmt.Errorf("failed to get probe statistics: %w", err)
	}

	// Handle empty results
	if len(results.Results) == 0 {
		return output.PrintEmpty("No statistics available for this probe in the specified period")
	}

	// Calculate aggregated statistics from the time buckets
	statsOutput := calculateStatsFromBuckets(probeID, flags.period, results)

	// Print the results using the configured output format
	return output.Print(statsOutput)
}

// parsePeriodToAggregateParams converts a period string to SDK parameters.
// Returns aggregate interval, from time, to time, and any error.
func parsePeriodToAggregateParams(period string) (aggregate string, from time.Time, to time.Time, err error) {
	now := time.Now().UTC()
	to = now

	switch period {
	case "24h":
		aggregate = "1h" // Hourly buckets for 24h view
		from = now.Add(-24 * time.Hour)
	case "7d":
		aggregate = "1h" // Hourly buckets for 7d view
		from = now.Add(-7 * 24 * time.Hour)
	case "30d":
		aggregate = "1d" // Daily buckets for 30d view
		from = now.Add(-30 * 24 * time.Hour)
	default:
		err = fmt.Errorf("invalid period %q: must be 24h, 7d, or 30d", period)
	}

	return aggregate, from, to, err
}

// weightedValue pairs a response time with its weight (check count).
type weightedValue struct {
	value  float64
	weight int64
}

// calculateStatsFromBuckets computes overall statistics from time-bucketed results.
func calculateStatsFromBuckets(probeID uuid.UUID, period string, results *client.AggregatedResultListResponse) *ProbeStatsOutput {
	var (
		totalChecks   int64
		successChecks int64
		failureChecks int64
		sumResponseMs float64
		minResponseMs = int(^uint(0) >> 1) // Max int
		maxResponseMs = 0
	)

	// Collect bucket averages with their weights for memory-efficient percentile calculation.
	// Instead of expanding each bucket's avg N times (which could be 100K+ elements),
	// we store each bucket's avg once with its weight.
	weightedTimes := make([]weightedValue, 0, len(results.Results))

	for _, bucket := range results.Results {
		totalChecks += bucket.TotalChecks
		successChecks += bucket.SuccessChecks
		failureChecks += bucket.FailureChecks
		sumResponseMs += bucket.AvgResponseMs * float64(bucket.TotalChecks)

		if bucket.MinResponseMs < minResponseMs {
			minResponseMs = bucket.MinResponseMs
		}
		if bucket.MaxResponseMs > maxResponseMs {
			maxResponseMs = bucket.MaxResponseMs
		}

		// Store bucket average with its check count weight
		if bucket.TotalChecks > 0 {
			weightedTimes = append(weightedTimes, weightedValue{
				value:  bucket.AvgResponseMs,
				weight: bucket.TotalChecks,
			})
		}
	}

	// Handle edge case of no checks
	if totalChecks == 0 {
		return &ProbeStatsOutput{
			ProbeID:     probeID,
			Period:      period,
			From:        results.From,
			To:          results.To,
			TimeBuckets: len(results.Results),
		}
	}

	// Calculate overall uptime
	uptimePercent := float64(successChecks) / float64(totalChecks) * 100

	// Calculate overall average response time
	avgResponseMs := sumResponseMs / float64(totalChecks)

	// Calculate p95 and p99 using weighted percentile calculation.
	// This is an approximation since we use bucket averages, not individual measurements.
	p95ResponseMs, p99ResponseMs := calculateWeightedPercentiles(weightedTimes, totalChecks)

	// Handle case where no valid response times were recorded
	if minResponseMs == int(^uint(0)>>1) {
		minResponseMs = 0
	}

	return &ProbeStatsOutput{
		ProbeID:       probeID,
		Period:        period,
		UptimePercent: uptimePercent,
		TotalChecks:   totalChecks,
		SuccessChecks: successChecks,
		FailureChecks: failureChecks,
		AvgResponseMs: avgResponseMs,
		P95ResponseMs: p95ResponseMs,
		P99ResponseMs: p99ResponseMs,
		MinResponseMs: minResponseMs,
		MaxResponseMs: maxResponseMs,
		From:          results.From,
		To:            results.To,
		TimeBuckets:   len(results.Results),
	}
}

// calculateWeightedPercentiles computes p95 and p99 from weighted response time buckets.
// This is memory-efficient: instead of expanding bucket averages N times (potentially 100K+
// elements), we use weighted percentile calculation with O(buckets) space complexity.
// Note: This approximation uses bucket averages, not individual measurements.
func calculateWeightedPercentiles(times []weightedValue, totalWeight int64) (p95, p99 float64) {
	if len(times) == 0 || totalWeight == 0 {
		return 0, 0
	}

	// Sort by value for percentile calculation
	sort.Slice(times, func(i, j int) bool {
		return times[i].value < times[j].value
	})

	// Calculate the target cumulative weights for p95 and p99
	p95Target := float64(totalWeight) * 0.95
	p99Target := float64(totalWeight) * 0.99

	var cumulativeWeight int64
	for _, wv := range times {
		cumulativeWeight += wv.weight

		// First bucket that reaches p95 threshold
		if p95 == 0 && float64(cumulativeWeight) >= p95Target {
			p95 = wv.value
		}

		// First bucket that reaches p99 threshold
		if p99 == 0 && float64(cumulativeWeight) >= p99Target {
			p99 = wv.value
		}

		// Exit early if both percentiles found
		if p95 > 0 && p99 > 0 {
			break
		}
	}

	// If we didn't find values (edge case), use the last value
	if len(times) > 0 {
		if p95 == 0 {
			p95 = times[len(times)-1].value
		}
		if p99 == 0 {
			p99 = times[len(times)-1].value
		}
	}

	return p95, p99
}

// calculatePercentiles computes p95 and p99 from a slice of response times.
// This is an approximation since we're using bucket averages, not individual measurements.
// Kept for backward compatibility with tests.
func calculatePercentiles(times []float64) (p95, p99 float64) {
	if len(times) == 0 {
		return 0, 0
	}

	// Sort the times for percentile calculation
	sort.Float64s(times)

	// Calculate p95 index (95th percentile)
	p95Index := int(float64(len(times)) * 0.95)
	if p95Index >= len(times) {
		p95Index = len(times) - 1
	}
	p95 = times[p95Index]

	// Calculate p99 index (99th percentile)
	p99Index := int(float64(len(times)) * 0.99)
	if p99Index >= len(times) {
		p99Index = len(times) - 1
	}
	p99 = times[p99Index]

	return p95, p99
}
