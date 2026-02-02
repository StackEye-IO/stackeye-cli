package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
)

func TestNewProbeHistoryCmd(t *testing.T) {
	cmd := NewProbeHistoryCmd()

	// Verify command basic properties
	if cmd.Use != "history <id>" {
		t.Errorf("expected Use to be 'history <id>', got %q", cmd.Use)
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

	// Verify flags exist
	flags := []struct {
		name     string
		defValue string
	}{
		{"limit", "20"},
		{"since", ""},
		{"region", ""},
		{"status", ""},
		{"page", "1"},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("expected '%s' flag to be defined", f.name)
		} else if flag.DefValue != f.defValue {
			t.Errorf("expected %s flag default to be %q, got %q", f.name, f.defValue, flag.DefValue)
		}
	}
}

func TestRunProbeHistory_NameResolution(t *testing.T) {
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
			flags := &probeHistoryFlags{limit: 20, page: 1}
			err := runProbeHistory(t.Context(), tt.idArg, flags)

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

func TestRunProbeHistory_InvalidLimit(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name    string
		limit   int
		wantErr string
	}{
		{
			name:    "zero limit",
			limit:   0,
			wantErr: "invalid limit",
		},
		{
			name:    "negative limit",
			limit:   -1,
			wantErr: "invalid limit",
		},
		{
			name:    "limit exceeds maximum",
			limit:   101,
			wantErr: "invalid limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &probeHistoryFlags{limit: tt.limit, page: 1}
			err := runProbeHistory(t.Context(), validUUID, flags)

			if err == nil {
				t.Error("expected error for invalid limit, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestRunProbeHistory_InvalidPage(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name    string
		page    int
		wantErr string
	}{
		{
			name:    "zero page",
			page:    0,
			wantErr: "invalid page",
		},
		{
			name:    "negative page",
			page:    -1,
			wantErr: "invalid page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &probeHistoryFlags{limit: 20, page: tt.page}
			err := runProbeHistory(t.Context(), validUUID, flags)

			if err == nil {
				t.Error("expected error for invalid page, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestRunProbeHistory_InvalidStatus(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name    string
		status  string
		wantErr string
	}{
		{
			name:    "invalid status value",
			status:  "invalid",
			wantErr: "invalid status",
		},
		{
			name:    "uppercase status",
			status:  "SUCCESS",
			wantErr: "invalid status",
		},
		{
			name:    "mixed case",
			status:  "Success",
			wantErr: "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &probeHistoryFlags{limit: 20, page: 1, status: tt.status}
			err := runProbeHistory(t.Context(), validUUID, flags)

			if err == nil {
				t.Error("expected error for invalid status, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestRunProbeHistory_ValidStatus(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	validStatuses := []string{"success", "failure"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			flags := &probeHistoryFlags{limit: 20, page: 1, status: status}
			err := runProbeHistory(t.Context(), validUUID, flags)

			// We expect an API client error since we're not mocking,
			// but we should NOT get a status validation error
			if err != nil && strings.Contains(err.Error(), "invalid status") {
				t.Errorf("status %q should be valid, got validation error: %v", status, err)
			}
		})
	}
}

func TestParseSinceDuration(t *testing.T) {
	tests := []struct {
		name    string
		since   string
		wantDur time.Duration
		wantErr bool
	}{
		{
			name:    "1 hour",
			since:   "1h",
			wantDur: time.Hour,
			wantErr: false,
		},
		{
			name:    "24 hours",
			since:   "24h",
			wantDur: 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "7 days",
			since:   "7d",
			wantDur: 7 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "30 days",
			since:   "30d",
			wantDur: 30 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "1 day",
			since:   "1d",
			wantDur: 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "30 minutes",
			since:   "30m",
			wantDur: 30 * time.Minute,
			wantErr: false,
		},
		{
			name:    "invalid format",
			since:   "invalid",
			wantDur: 0,
			wantErr: true,
		},
		{
			name:    "negative duration",
			since:   "-1h",
			wantDur: 0,
			wantErr: true,
		},
		{
			name:    "zero duration",
			since:   "0h",
			wantDur: 0,
			wantErr: true,
		},
		{
			name:    "empty string",
			since:   "",
			wantDur: 0,
			wantErr: true,
		},
		{
			name:    "days exceeds maximum",
			since:   "400d",
			wantDur: 0,
			wantErr: true,
		},
		{
			name:    "zero days",
			since:   "0d",
			wantDur: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSinceDuration(tt.since)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseSinceDuration(%q) error = %v, wantErr %v", tt.since, err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantDur {
				t.Errorf("parseSinceDuration(%q) = %v, want %v", tt.since, got, tt.wantDur)
			}
		})
	}
}

func TestConvertToHistoryOutput(t *testing.T) {
	probeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	resultID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")
	checkedAt := time.Now().Add(-1 * time.Hour)
	statusCode := 200

	results := &client.ProbeResultListResponse{
		Results: []client.ProbeResult{
			{
				ID:             resultID,
				ProbeID:        probeID,
				Region:         "us-east-1",
				Status:         "success",
				ResponseTimeMs: 150,
				StatusCode:     &statusCode,
				CheckedAt:      checkedAt,
			},
		},
		Total: 1,
		Page:  1,
		Limit: 20,
	}

	flags := &probeHistoryFlags{limit: 20, page: 1}
	output := convertToHistoryOutput(probeID, results, flags)

	if output.ProbeID != probeID {
		t.Errorf("expected ProbeID %s, got %s", probeID, output.ProbeID)
	}

	if len(output.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(output.Results))
		return
	}

	entry := output.Results[0]
	if entry.ID != resultID {
		t.Errorf("expected result ID %s, got %s", resultID, entry.ID)
	}
	if entry.Region != "us-east-1" {
		t.Errorf("expected region 'us-east-1', got %s", entry.Region)
	}
	if entry.Status != "success" {
		t.Errorf("expected status 'success', got %s", entry.Status)
	}
	if entry.ResponseTimeMs != 150 {
		t.Errorf("expected response time 150ms, got %d", entry.ResponseTimeMs)
	}
	if entry.StatusCode == nil || *entry.StatusCode != 200 {
		t.Errorf("expected status code 200, got %v", entry.StatusCode)
	}
}

func TestConvertToHistoryOutput_WithSinceFilter(t *testing.T) {
	probeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Create results with different timestamps
	now := time.Now()
	recentResult := client.ProbeResult{
		ID:             uuid.MustParse("660e8400-e29b-41d4-a716-446655440001"),
		ProbeID:        probeID,
		Region:         "us-east-1",
		Status:         "success",
		ResponseTimeMs: 150,
		CheckedAt:      now.Add(-30 * time.Minute), // 30 minutes ago
	}
	oldResult := client.ProbeResult{
		ID:             uuid.MustParse("660e8400-e29b-41d4-a716-446655440002"),
		ProbeID:        probeID,
		Region:         "us-west-2",
		Status:         "success",
		ResponseTimeMs: 200,
		CheckedAt:      now.Add(-2 * time.Hour), // 2 hours ago
	}

	results := &client.ProbeResultListResponse{
		Results: []client.ProbeResult{recentResult, oldResult},
		Total:   2,
		Page:    1,
		Limit:   20,
	}

	// Filter to last 1 hour - should only include recentResult
	flags := &probeHistoryFlags{limit: 20, page: 1, since: "1h"}
	output := convertToHistoryOutput(probeID, results, flags)

	if len(output.Results) != 1 {
		t.Errorf("expected 1 result after filtering, got %d", len(output.Results))
		return
	}

	if output.Results[0].Region != "us-east-1" {
		t.Errorf("expected only recent result (us-east-1), got %s", output.Results[0].Region)
	}
}

func TestRunProbeHistory_InvalidSince(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name    string
		since   string
		wantErr string
	}{
		{
			name:    "invalid format",
			since:   "invalid",
			wantErr: "invalid since value",
		},
		{
			name:    "negative duration",
			since:   "-1h",
			wantErr: "invalid since value",
		},
		{
			name:    "too many days",
			since:   "400d",
			wantErr: "invalid since value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &probeHistoryFlags{limit: 20, page: 1, since: tt.since}
			err := runProbeHistory(t.Context(), validUUID, flags)

			if err == nil {
				t.Error("expected error for invalid since, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestRunProbeHistory_ValidSince(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	validSinceValues := []string{"1h", "24h", "7d", "30d", "30m"}

	for _, since := range validSinceValues {
		t.Run(since, func(t *testing.T) {
			flags := &probeHistoryFlags{limit: 20, page: 1, since: since}
			err := runProbeHistory(t.Context(), validUUID, flags)

			// We expect an API client error since we're not mocking,
			// but we should NOT get a since validation error
			if err != nil && strings.Contains(err.Error(), "invalid since value") {
				t.Errorf("since %q should be valid, got validation error: %v", since, err)
			}
		})
	}
}
