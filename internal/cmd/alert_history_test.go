package cmd

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestNewAlertHistoryCmd(t *testing.T) {
	cmd := NewAlertHistoryCmd()

	if cmd.Use != "history" {
		t.Errorf("expected Use='history', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Show historical alerts" {
		t.Errorf("expected Short='Show historical alerts', got %q", cmd.Short)
	}
}

func TestNewAlertHistoryCmd_Aliases(t *testing.T) {
	cmd := NewAlertHistoryCmd()

	expectedAliases := []string{"hist"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, expected := range expectedAliases {
		if !slices.Contains(cmd.Aliases, expected) {
			t.Errorf("expected alias %q not found", expected)
		}
	}
}

func TestNewAlertHistoryCmd_Long(t *testing.T) {
	cmd := NewAlertHistoryCmd()

	long := cmd.Long

	// Should contain time range documentation
	timeFormats := []string{"24h", "7d", "RFC3339"}
	for _, format := range timeFormats {
		if !strings.Contains(long, format) {
			t.Errorf("expected Long description to mention time format %q", format)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye alert history") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention --since and --until flags
	if !strings.Contains(long, "--since") {
		t.Error("expected Long description to document --since flag")
	}
	if !strings.Contains(long, "--until") {
		t.Error("expected Long description to document --until flag")
	}
}

func TestNewAlertHistoryCmd_Flags(t *testing.T) {
	cmd := NewAlertHistoryCmd()

	// Verify expected flags exist
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"probe", ""},
		{"since", ""},
		{"until", ""},
		{"page", "1"},
		{"limit", "20"},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("expected flag %q to exist", f.name)
			continue
		}
		if flag.DefValue != f.defaultValue {
			t.Errorf("flag %q: expected default %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestParseTimeFlag_Empty(t *testing.T) {
	result, err := parseTimeFlag("")
	if err != nil {
		t.Errorf("expected no error for empty string, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result for empty string")
	}
}

func TestParseTimeFlag_RelativeHours(t *testing.T) {
	before := time.Now()
	result, err := parseTimeFlag("24h")
	after := time.Now()

	if err != nil {
		t.Errorf("expected no error for '24h', got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for '24h'")
	}

	// Result should be approximately 24 hours ago
	expectedMin := before.Add(-24 * time.Hour)
	expectedMax := after.Add(-24 * time.Hour)

	if result.Before(expectedMin) || result.After(expectedMax) {
		t.Errorf("result %v not within expected range [%v, %v]", result, expectedMin, expectedMax)
	}
}

func TestParseTimeFlag_RelativeDays(t *testing.T) {
	before := time.Now()
	result, err := parseTimeFlag("7d")
	after := time.Now()

	if err != nil {
		t.Errorf("expected no error for '7d', got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for '7d'")
	}

	// Result should be approximately 7 days ago
	expectedMin := before.Add(-7 * 24 * time.Hour)
	expectedMax := after.Add(-7 * 24 * time.Hour)

	if result.Before(expectedMin) || result.After(expectedMax) {
		t.Errorf("result %v not within expected range [%v, %v]", result, expectedMin, expectedMax)
	}
}

func TestParseTimeFlag_RFC3339(t *testing.T) {
	input := "2024-01-15T10:30:00Z"
	result, err := parseTimeFlag(input)

	if err != nil {
		t.Errorf("expected no error for RFC3339 time, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for RFC3339 time")
	}

	expected, _ := time.Parse(time.RFC3339, input)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParseTimeFlag_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid relative", "24x"},
		{"invalid date", "2024-13-45"},
		{"random string", "yesterday"},
		{"negative days", "-7d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTimeFlag(tt.input)
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
		})
	}
}

func TestRunAlertHistory_Validation(t *testing.T) {
	tests := []struct {
		name         string
		limit        int
		page         int
		probeID      string
		since        string
		until        string
		wantErrorMsg string
	}{
		{
			name:         "limit too low",
			limit:        0,
			page:         1,
			wantErrorMsg: "invalid limit 0: must be between 1 and 100",
		},
		{
			name:         "limit too high",
			limit:        101,
			page:         1,
			wantErrorMsg: "invalid limit 101: must be between 1 and 100",
		},
		{
			name:         "page too low",
			limit:        20,
			page:         0,
			wantErrorMsg: "invalid page 0: must be at least 1",
		},
		{
			name:         "invalid probe ID",
			limit:        20,
			page:         1,
			probeID:      "not-a-uuid",
			wantErrorMsg: `invalid probe ID "not-a-uuid"`,
		},
		{
			name:         "invalid since format",
			limit:        20,
			page:         1,
			since:        "badtime",
			wantErrorMsg: "invalid --since flag",
		},
		{
			name:         "invalid until format",
			limit:        20,
			page:         1,
			until:        "badtime",
			wantErrorMsg: "invalid --until flag",
		},
		{
			name:         "since after until",
			limit:        20,
			page:         1,
			since:        "2024-01-15T00:00:00Z",
			until:        "2024-01-10T00:00:00Z",
			wantErrorMsg: "invalid time range: --since",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &alertHistoryFlags{
				page:    tt.page,
				limit:   tt.limit,
				probeID: tt.probeID,
				since:   tt.since,
				until:   tt.until,
			}

			err := runAlertHistory(context.Background(), flags)

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErrorMsg)
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrorMsg) {
				t.Errorf("expected error containing %q, got %q", tt.wantErrorMsg, err.Error())
			}
		})
	}
}

func TestRunAlertHistory_ValidFlags(t *testing.T) {
	// Test that valid flags pass validation (will fail later on API client)
	flags := &alertHistoryFlags{
		page:  1,
		limit: 20,
		since: "7d",
	}

	err := runAlertHistory(context.Background(), flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	validationErrors := []string{
		"invalid limit",
		"invalid page",
		"invalid probe ID",
		"invalid --since",
		"invalid --until",
		"invalid time range",
	}
	for _, ve := range validationErrors {
		if strings.Contains(err.Error(), ve) {
			t.Errorf("got unexpected validation error: %s", err.Error())
		}
	}
}

func TestRunAlertHistory_ValidProbeID(t *testing.T) {
	// Test that a valid UUID probe ID passes validation
	flags := &alertHistoryFlags{
		page:    1,
		limit:   20,
		probeID: "123e4567-e89b-12d3-a456-426614174000", // Valid UUID
	}

	err := runAlertHistory(context.Background(), flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a probe ID validation error
	if strings.Contains(err.Error(), "invalid probe ID") {
		t.Errorf("got unexpected validation error for valid probe ID: %s", err.Error())
	}
}

func TestRunAlertHistory_ValidTimeRange(t *testing.T) {
	// Test that a valid time range passes validation
	flags := &alertHistoryFlags{
		page:  1,
		limit: 20,
		since: "2024-01-01T00:00:00Z",
		until: "2024-01-31T23:59:59Z",
	}

	err := runAlertHistory(context.Background(), flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a time range validation error
	if strings.Contains(err.Error(), "invalid time range") {
		t.Errorf("got unexpected validation error for valid time range: %s", err.Error())
	}
}
