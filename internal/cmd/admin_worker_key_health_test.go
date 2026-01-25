// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
)

// TestNewAdminWorkerKeyHealthCmd verifies that the worker-key health command is properly constructed.
func TestNewAdminWorkerKeyHealthCmd(t *testing.T) {
	cmd := NewAdminWorkerKeyHealthCmd()

	if cmd.Use != "health" {
		t.Errorf("expected Use='health', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Check worker health status" {
		t.Errorf("expected Short='Check worker health status', got %q", cmd.Short)
	}
}

// TestNewAdminWorkerKeyHealthCmd_Aliases verifies the command aliases.
func TestNewAdminWorkerKeyHealthCmd_Aliases(t *testing.T) {
	cmd := NewAdminWorkerKeyHealthCmd()

	expectedAliases := []string{"status"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for i, alias := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != alias {
			t.Errorf("expected alias %q at index %d", alias, i)
		}
	}
}

// TestNewAdminWorkerKeyHealthCmd_Long verifies the Long description contains key information.
func TestNewAdminWorkerKeyHealthCmd_Long(t *testing.T) {
	cmd := NewAdminWorkerKeyHealthCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"health status",
		"healthy",
		"stale",
		"threshold",
		"heartbeat",
	}
	for _, feature := range features {
		if !strings.Contains(strings.ToLower(long), strings.ToLower(feature)) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye admin worker-key health") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention threshold flag
	if !strings.Contains(long, "--threshold") {
		t.Error("expected Long description to mention --threshold flag")
	}
}

// TestNewAdminWorkerKeyHealthCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewAdminWorkerKeyHealthCmd_RunEIsSet(t *testing.T) {
	cmd := NewAdminWorkerKeyHealthCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewAdminWorkerKeyHealthCmd_ThresholdFlag verifies that the threshold flag is registered.
func TestNewAdminWorkerKeyHealthCmd_ThresholdFlag(t *testing.T) {
	cmd := NewAdminWorkerKeyHealthCmd()

	flag := cmd.Flags().Lookup("threshold")
	if flag == nil {
		t.Fatal("expected --threshold flag to be registered")
	}

	if flag.Shorthand != "t" {
		t.Errorf("expected threshold flag shorthand to be 't', got %q", flag.Shorthand)
	}

	if flag.DefValue != "0" {
		t.Errorf("expected threshold flag default value to be '0', got %q", flag.DefValue)
	}

	// Check flag usage contains description
	if flag.Usage == "" {
		t.Error("expected threshold flag to have a usage description")
	}
}

// TestPrintWorkerHealth_DoesNotPanic verifies that printWorkerHealth doesn't panic with various inputs.
func TestPrintWorkerHealth_DoesNotPanic(t *testing.T) {
	now := time.Now()
	hourAgo := now.Add(-1 * time.Hour)
	dayAgo := now.Add(-24 * time.Hour)

	tests := []struct {
		name   string
		status *admin.WorkerHealthResponse
	}{
		{
			name: "empty status",
			status: &admin.WorkerHealthResponse{
				TotalActive:    0,
				TotalHealthy:   0,
				TotalStale:     0,
				HealthyWorkers: nil,
				StaleWorkers:   nil,
				CheckedAt:      now,
			},
		},
		{
			name: "healthy workers only",
			status: &admin.WorkerHealthResponse{
				TotalActive:  2,
				TotalHealthy: 2,
				TotalStale:   0,
				HealthyWorkers: []admin.WorkerHealthWorker{
					{ID: "1", Region: "nyc3", KeyPrefix: "wk_abc", IsActive: true, LastSeenAt: &now},
					{ID: "2", Region: "lon1", KeyPrefix: "wk_def", IsActive: true, LastSeenAt: &hourAgo},
				},
				StaleWorkers: nil,
				CheckedAt:    now,
			},
		},
		{
			name: "stale workers only",
			status: &admin.WorkerHealthResponse{
				TotalActive:    1,
				TotalHealthy:   0,
				TotalStale:     1,
				HealthyWorkers: nil,
				StaleWorkers: []admin.WorkerHealthWorker{
					{ID: "3", Region: "fra1", KeyPrefix: "wk_ghi", IsActive: true, LastSeenAt: &dayAgo},
				},
				CheckedAt: now,
			},
		},
		{
			name: "mixed healthy and stale",
			status: &admin.WorkerHealthResponse{
				TotalActive:  3,
				TotalHealthy: 2,
				TotalStale:   1,
				HealthyWorkers: []admin.WorkerHealthWorker{
					{ID: "1", Region: "nyc3", KeyPrefix: "wk_abc", IsActive: true, LastSeenAt: &now},
					{ID: "2", Region: "lon1", KeyPrefix: "wk_def", IsActive: true, LastSeenAt: &hourAgo},
				},
				StaleWorkers: []admin.WorkerHealthWorker{
					{ID: "3", Region: "fra1", KeyPrefix: "wk_ghi", IsActive: true, LastSeenAt: &dayAgo},
				},
				CheckedAt: now,
			},
		},
		{
			name: "nil last seen times",
			status: &admin.WorkerHealthResponse{
				TotalActive:  1,
				TotalHealthy: 0,
				TotalStale:   1,
				StaleWorkers: []admin.WorkerHealthWorker{
					{ID: "4", Region: "sfo1", KeyPrefix: "wk_jkl", IsActive: true, LastSeenAt: nil},
				},
				CheckedAt: now,
			},
		},
		{
			name: "zero checked at",
			status: &admin.WorkerHealthResponse{
				TotalActive:  0,
				TotalHealthy: 0,
				TotalStale:   0,
				CheckedAt:    time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printWorkerHealth panicked: %v", r)
				}
			}()
			printWorkerHealth(tt.status)
		})
	}
}

// TestFormatWorkerLastSeen verifies the relative time formatting.
func TestFormatWorkerLastSeen(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     *time.Time
		expected string
	}{
		{"nil time", nil, "Never"},
		{"zero time", func() *time.Time { t := time.Time{}; return &t }(), "Never"},
		{"just now", &now, "Just now"},
		{"1 minute ago", func() *time.Time { t := now.Add(-1 * time.Minute); return &t }(), "1 min ago"},
		{"5 minutes ago", func() *time.Time { t := now.Add(-5 * time.Minute); return &t }(), "5 mins ago"},
		{"1 hour ago", func() *time.Time { t := now.Add(-1 * time.Hour); return &t }(), "1 hour ago"},
		{"2 hours ago", func() *time.Time { t := now.Add(-2 * time.Hour); return &t }(), "2 hours ago"},
		{"1 day ago", func() *time.Time { t := now.Add(-24 * time.Hour); return &t }(), "1 day ago"},
		{"3 days ago", func() *time.Time { t := now.Add(-72 * time.Hour); return &t }(), "3 days ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatWorkerLastSeen(tt.time)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestTruncateHealthField verifies string truncation logic.
func TestTruncateHealthField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "abc", 10, "abc"},
		{"exact length", "abcdefgh", 8, "abcdefgh"},
		{"needs truncation", "abcdefghijklmnop", 10, "abcdefg..."},
		{"very short maxLen", "abcd", 2, "ab"},
		{"empty string", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateHealthField(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatHealthCheckTime verifies timestamp formatting.
func TestFormatHealthCheckTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"zero time", time.Time{}, "Unknown"},
		{"valid time", time.Date(2026, 1, 25, 14, 30, 45, 0, time.UTC), "14:30:45 UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHealthCheckTime(tt.time)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestNewAdminWorkerKeyCmd_HasHealthSubcommand verifies that health subcommand is registered.
func TestNewAdminWorkerKeyCmd_HasHealthSubcommand(t *testing.T) {
	cmd := NewAdminWorkerKeyCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 3 {
		t.Errorf("expected worker-key command to have at least 3 subcommands (create, delete, health), got %d", len(subcommands))
	}

	// Verify health subcommand is registered
	found := false
	for _, sub := range subcommands {
		if sub.Use == "health" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'health' subcommand to be registered")
	}
}
