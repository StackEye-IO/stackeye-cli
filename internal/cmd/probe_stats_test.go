package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
)

func TestNewProbeStatsCmd(t *testing.T) {
	cmd := NewProbeStatsCmd()

	// Verify command basic properties
	if cmd.Use != "stats <id>" {
		t.Errorf("expected Use to be 'stats <id>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Verify required argument count
	if cmd.Args == nil {
		t.Error("expected Args validator to be set")
	}

	// Verify flags exist with correct defaults
	periodFlag := cmd.Flags().Lookup("period")
	if periodFlag == nil {
		t.Error("expected 'period' flag to be defined")
	} else if periodFlag.DefValue != "24h" {
		t.Errorf("expected period flag default to be '24h', got %q", periodFlag.DefValue)
	}
}

func TestRunProbeStats_NameResolution(t *testing.T) {
	// Since probe name resolution was added, non-UUID inputs are now treated as
	// potential probe names that need API resolution. Without a configured API
	// client, these will fail with an API client initialization error.
	tests := []struct {
		name    string
		idArg   string
		wantErr string
	}{
		{
			name:    "probe name requires API client",
			idArg:   "api-health",
			wantErr: "failed to initialize API client",
		},
		{
			name:    "partial UUID treated as name",
			idArg:   "550e8400-e29b-41d4",
			wantErr: "failed to initialize API client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &probeStatsFlags{period: "24h"}
			err := runProbeStats(t.Context(), tt.idArg, flags)

			if err == nil {
				t.Error("expected error when API client not configured, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestRunProbeStats_InvalidPeriod(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name    string
		period  string
		wantErr string
	}{
		{
			name:    "invalid period value",
			period:  "invalid",
			wantErr: "invalid period",
		},
		{
			name:    "uppercase period",
			period:  "24H",
			wantErr: "invalid period",
		},
		{
			name:    "unsupported duration",
			period:  "1h",
			wantErr: "invalid period",
		},
		{
			name:    "unsupported days",
			period:  "14d",
			wantErr: "invalid period",
		},
		{
			name:    "empty period",
			period:  "",
			wantErr: "invalid period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &probeStatsFlags{period: tt.period}
			err := runProbeStats(t.Context(), validUUID, flags)

			if err == nil {
				t.Error("expected error for invalid period, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestRunProbeStats_ValidPeriod(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	validPeriods := []string{"24h", "7d", "30d"}

	for _, period := range validPeriods {
		t.Run(period, func(t *testing.T) {
			flags := &probeStatsFlags{period: period}
			err := runProbeStats(t.Context(), validUUID, flags)

			// We expect an API client error since we're not mocking,
			// but we should NOT get a period validation error
			if err != nil && strings.Contains(err.Error(), "invalid period") {
				t.Errorf("period %q should be valid, got validation error: %v", period, err)
			}
		})
	}
}

func TestParsePeriodToAggregateParams(t *testing.T) {
	tests := []struct {
		name          string
		period        string
		wantAggregate string
		wantErr       bool
	}{
		{
			name:          "24 hours",
			period:        "24h",
			wantAggregate: "1h",
			wantErr:       false,
		},
		{
			name:          "7 days",
			period:        "7d",
			wantAggregate: "1h",
			wantErr:       false,
		},
		{
			name:          "30 days",
			period:        "30d",
			wantAggregate: "1d",
			wantErr:       false,
		},
		{
			name:          "invalid period",
			period:        "invalid",
			wantAggregate: "",
			wantErr:       true,
		},
		{
			name:          "unsupported period",
			period:        "1h",
			wantAggregate: "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregate, from, to, err := parsePeriodToAggregateParams(tt.period)

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePeriodToAggregateParams(%q) error = %v, wantErr %v", tt.period, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if aggregate != tt.wantAggregate {
					t.Errorf("parsePeriodToAggregateParams(%q) aggregate = %q, want %q", tt.period, aggregate, tt.wantAggregate)
				}

				// Verify time range is reasonable
				if from.IsZero() {
					t.Error("expected from time to be set")
				}
				if to.IsZero() {
					t.Error("expected to time to be set")
				}
				if !from.Before(to) {
					t.Error("expected from to be before to")
				}
			}
		})
	}
}

func TestCalculateStatsFromBuckets(t *testing.T) {
	probeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now().UTC()

	t.Run("basic statistics calculation", func(t *testing.T) {
		results := &client.AggregatedResultListResponse{
			Results: []client.AggregatedResult{
				{
					TimeBucket:    now.Add(-2 * time.Hour),
					TotalChecks:   100,
					SuccessChecks: 95,
					FailureChecks: 5,
					AvgResponseMs: 150.0,
					MinResponseMs: 50,
					MaxResponseMs: 300,
					UptimePercent: 95.0,
				},
				{
					TimeBucket:    now.Add(-1 * time.Hour),
					TotalChecks:   100,
					SuccessChecks: 98,
					FailureChecks: 2,
					AvgResponseMs: 120.0,
					MinResponseMs: 40,
					MaxResponseMs: 250,
					UptimePercent: 98.0,
				},
			},
			Aggregate: "1h",
			From:      now.Add(-24 * time.Hour),
			To:        now,
		}

		output := calculateStatsFromBuckets(probeID, "24h", results)

		// Verify basic fields
		if output.ProbeID != probeID {
			t.Errorf("expected ProbeID %s, got %s", probeID, output.ProbeID)
		}
		if output.Period != "24h" {
			t.Errorf("expected period '24h', got %s", output.Period)
		}

		// Verify aggregated counts
		if output.TotalChecks != 200 {
			t.Errorf("expected total checks 200, got %d", output.TotalChecks)
		}
		if output.SuccessChecks != 193 {
			t.Errorf("expected success checks 193, got %d", output.SuccessChecks)
		}
		if output.FailureChecks != 7 {
			t.Errorf("expected failure checks 7, got %d", output.FailureChecks)
		}

		// Verify uptime calculation (193/200 = 96.5%)
		expectedUptime := 96.5
		if output.UptimePercent != expectedUptime {
			t.Errorf("expected uptime %.1f%%, got %.1f%%", expectedUptime, output.UptimePercent)
		}

		// Verify min/max
		if output.MinResponseMs != 40 {
			t.Errorf("expected min response 40ms, got %d", output.MinResponseMs)
		}
		if output.MaxResponseMs != 300 {
			t.Errorf("expected max response 300ms, got %d", output.MaxResponseMs)
		}

		// Verify time buckets count
		if output.TimeBuckets != 2 {
			t.Errorf("expected 2 time buckets, got %d", output.TimeBuckets)
		}
	})

	t.Run("empty results", func(t *testing.T) {
		results := &client.AggregatedResultListResponse{
			Results:   []client.AggregatedResult{},
			Aggregate: "1h",
			From:      now.Add(-24 * time.Hour),
			To:        now,
		}

		output := calculateStatsFromBuckets(probeID, "24h", results)

		if output.TotalChecks != 0 {
			t.Errorf("expected total checks 0, got %d", output.TotalChecks)
		}
		if output.UptimePercent != 0 {
			t.Errorf("expected uptime 0%%, got %.1f%%", output.UptimePercent)
		}
	})

	t.Run("all failures", func(t *testing.T) {
		results := &client.AggregatedResultListResponse{
			Results: []client.AggregatedResult{
				{
					TimeBucket:    now.Add(-1 * time.Hour),
					TotalChecks:   10,
					SuccessChecks: 0,
					FailureChecks: 10,
					AvgResponseMs: 0,
					MinResponseMs: 0,
					MaxResponseMs: 0,
					UptimePercent: 0,
				},
			},
			Aggregate: "1h",
			From:      now.Add(-24 * time.Hour),
			To:        now,
		}

		output := calculateStatsFromBuckets(probeID, "24h", results)

		if output.UptimePercent != 0 {
			t.Errorf("expected uptime 0%%, got %.1f%%", output.UptimePercent)
		}
		if output.FailureChecks != 10 {
			t.Errorf("expected failure checks 10, got %d", output.FailureChecks)
		}
	})

	t.Run("perfect uptime", func(t *testing.T) {
		results := &client.AggregatedResultListResponse{
			Results: []client.AggregatedResult{
				{
					TimeBucket:    now.Add(-1 * time.Hour),
					TotalChecks:   100,
					SuccessChecks: 100,
					FailureChecks: 0,
					AvgResponseMs: 100.0,
					MinResponseMs: 50,
					MaxResponseMs: 150,
					UptimePercent: 100.0,
				},
			},
			Aggregate: "1h",
			From:      now.Add(-24 * time.Hour),
			To:        now,
		}

		output := calculateStatsFromBuckets(probeID, "24h", results)

		if output.UptimePercent != 100 {
			t.Errorf("expected uptime 100%%, got %.1f%%", output.UptimePercent)
		}
		if output.FailureChecks != 0 {
			t.Errorf("expected failure checks 0, got %d", output.FailureChecks)
		}
	})
}

func TestCalculateWeightedPercentiles(t *testing.T) {
	tests := []struct {
		name        string
		times       []weightedValue
		totalWeight int64
		wantP95     float64
		wantP99     float64
	}{
		{
			name:        "empty slice",
			times:       []weightedValue{},
			totalWeight: 0,
			wantP95:     0,
			wantP99:     0,
		},
		{
			name: "single value",
			times: []weightedValue{
				{value: 100, weight: 10},
			},
			totalWeight: 10,
			wantP95:     100,
			wantP99:     100,
		},
		{
			name: "two buckets equal weight",
			times: []weightedValue{
				{value: 100, weight: 50},
				{value: 200, weight: 50},
			},
			totalWeight: 100,
			wantP95:     200, // 95th percentile lands in bucket 2
			wantP99:     200, // 99th percentile lands in bucket 2
		},
		{
			name: "weighted distribution favoring low values",
			times: []weightedValue{
				{value: 50, weight: 90},  // 90% of checks have 50ms avg
				{value: 200, weight: 10}, // 10% of checks have 200ms avg
			},
			totalWeight: 100,
			wantP95:     200, // 95th percentile is in the 200ms bucket
			wantP99:     200, // 99th percentile is in the 200ms bucket
		},
		{
			name: "heavily weighted toward fast responses",
			times: []weightedValue{
				{value: 50, weight: 98},
				{value: 500, weight: 2},
			},
			totalWeight: 100,
			wantP95:     50,  // 95th percentile still in fast bucket
			wantP99:     500, // 99th percentile in slow bucket
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p95, p99 := calculateWeightedPercentiles(tt.times, tt.totalWeight)

			if p95 != tt.wantP95 {
				t.Errorf("p95 = %.1f, want %.1f", p95, tt.wantP95)
			}
			if p99 != tt.wantP99 {
				t.Errorf("p99 = %.1f, want %.1f", p99, tt.wantP99)
			}
		})
	}
}

func TestCalculatePercentiles(t *testing.T) {
	tests := []struct {
		name    string
		times   []float64
		wantP95 float64
		wantP99 float64
	}{
		{
			name:    "empty slice",
			times:   []float64{},
			wantP95: 0,
			wantP99: 0,
		},
		{
			name:    "single value",
			times:   []float64{100},
			wantP95: 100,
			wantP99: 100,
		},
		{
			name:    "two values",
			times:   []float64{100, 200},
			wantP95: 200,
			wantP99: 200,
		},
		{
			name:    "sequential values 1-100",
			times:   makeSequence(1, 100),
			wantP95: 96,  // Index 95 (0.95 * 100) = value 96
			wantP99: 100, // Index 99 (0.99 * 100) = value 100
		},
		{
			name:    "all same values",
			times:   []float64{100, 100, 100, 100, 100},
			wantP95: 100,
			wantP99: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p95, p99 := calculatePercentiles(tt.times)

			if p95 != tt.wantP95 {
				t.Errorf("p95 = %.1f, want %.1f", p95, tt.wantP95)
			}
			if p99 != tt.wantP99 {
				t.Errorf("p99 = %.1f, want %.1f", p99, tt.wantP99)
			}
		})
	}
}

// makeSequence creates a slice of floats from start to end (inclusive).
func makeSequence(start, end int) []float64 {
	result := make([]float64, end-start+1)
	for i := start; i <= end; i++ {
		result[i-start] = float64(i)
	}
	return result
}
