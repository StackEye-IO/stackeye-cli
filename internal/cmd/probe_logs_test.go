package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
)

func TestNewProbeLogsCmd(t *testing.T) {
	cmd := NewProbeLogsCmd()

	if cmd.Use != "logs <id>" {
		t.Errorf("expected Use to be 'logs <id>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	if cmd.Args == nil {
		t.Error("expected Args validator to be set")
	}

	// Verify flags exist with expected defaults
	flags := []struct {
		name     string
		defValue string
	}{
		{"limit", "50"},
		{"since", ""},
		{"until", ""},
		{"region", ""},
		{"status", ""},
		{"follow", "false"},
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

func TestNewProbeLogsCmd_FollowShorthand(t *testing.T) {
	cmd := NewProbeLogsCmd()

	flag := cmd.Flags().ShorthandLookup("f")
	if flag == nil {
		t.Error("expected -f shorthand for --follow flag")
	} else if flag.Name != "follow" {
		t.Errorf("expected -f to map to 'follow', got %q", flag.Name)
	}
}

func TestValidateProbeLogsFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   *probeLogsFlags
		wantErr string
	}{
		{
			name:    "valid defaults",
			flags:   &probeLogsFlags{limit: 50},
			wantErr: "",
		},
		{
			name:    "valid with all flags",
			flags:   &probeLogsFlags{limit: 100, since: "24h", until: "1h", region: "us-east-1", status: "failure"},
			wantErr: "",
		},
		{
			name:    "limit too low",
			flags:   &probeLogsFlags{limit: 0},
			wantErr: "invalid limit",
		},
		{
			name:    "limit too high",
			flags:   &probeLogsFlags{limit: 1001},
			wantErr: "invalid limit",
		},
		{
			name:    "negative limit",
			flags:   &probeLogsFlags{limit: -5},
			wantErr: "invalid limit",
		},
		{
			name:    "invalid status",
			flags:   &probeLogsFlags{limit: 50, status: "unknown"},
			wantErr: "invalid status",
		},
		{
			name:    "uppercase status rejected",
			flags:   &probeLogsFlags{limit: 50, status: "SUCCESS"},
			wantErr: "invalid status",
		},
		{
			name:    "valid success status",
			flags:   &probeLogsFlags{limit: 50, status: "success"},
			wantErr: "",
		},
		{
			name:    "valid failure status",
			flags:   &probeLogsFlags{limit: 50, status: "failure"},
			wantErr: "",
		},
		{
			name:    "invalid since",
			flags:   &probeLogsFlags{limit: 50, since: "invalid"},
			wantErr: "invalid --since",
		},
		{
			name:    "invalid until",
			flags:   &probeLogsFlags{limit: 50, until: "notadate"},
			wantErr: "invalid --until",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProbeLogsFlags(tt.flags)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErr)
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestParseUntilValue(t *testing.T) {
	tests := []struct {
		name    string
		until   string
		wantErr bool
	}{
		{
			name:    "duration 1h",
			until:   "1h",
			wantErr: false,
		},
		{
			name:    "duration 7d",
			until:   "7d",
			wantErr: false,
		},
		{
			name:    "RFC3339 timestamp",
			until:   "2025-01-15T10:00:00Z",
			wantErr: false,
		},
		{
			name:    "RFC3339 with timezone offset",
			until:   "2025-01-15T10:00:00+05:00",
			wantErr: false,
		},
		{
			name:    "invalid value",
			until:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "empty string",
			until:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseUntilValue(tt.until)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseUntilValue(%q) error = %v, wantErr %v", tt.until, err, tt.wantErr)
				return
			}

			if !tt.wantErr && result.IsZero() {
				t.Errorf("parseUntilValue(%q) returned zero time", tt.until)
			}
		})
	}
}

func TestParseUntilValue_DurationRelative(t *testing.T) {
	before := time.Now()
	result, err := parseUntilValue("1h")
	after := time.Now()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedEarliest := before.Add(-1 * time.Hour)
	expectedLatest := after.Add(-1 * time.Hour)

	if result.Before(expectedEarliest) || result.After(expectedLatest) {
		t.Errorf("parseUntilValue(\"1h\") = %v, expected between %v and %v",
			result, expectedEarliest, expectedLatest)
	}
}

func TestResolveLogsTimeRange(t *testing.T) {
	tests := []struct {
		name    string
		flags   *probeLogsFlags
		wantErr string
	}{
		{
			name:    "no time flags",
			flags:   &probeLogsFlags{limit: 50},
			wantErr: "",
		},
		{
			name:    "since only",
			flags:   &probeLogsFlags{limit: 50, since: "1h"},
			wantErr: "",
		},
		{
			name:    "until only",
			flags:   &probeLogsFlags{limit: 50, until: "30m"},
			wantErr: "",
		},
		{
			name:    "since before until",
			flags:   &probeLogsFlags{limit: 50, since: "24h", until: "1h"},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to, err := resolveLogsTimeRange(tt.flags)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}

				// Verify from is before to when both are set
				if !from.IsZero() && !to.IsZero() {
					if !from.Before(to) {
						t.Errorf("expected from (%v) to be before to (%v)", from, to)
					}
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}

func TestResolveLogsTimeRange_SinceAfterUntil(t *testing.T) {
	// since=1h and until=24h means from=now-1h and to=now-24h,
	// which is from > to, so this should error
	flags := &probeLogsFlags{limit: 50, since: "1h", until: "24h"}
	_, _, err := resolveLogsTimeRange(flags)

	if err == nil {
		t.Error("expected error when since is after until")
		return
	}

	if !strings.Contains(err.Error(), "must be before") {
		t.Errorf("expected error about time ordering, got %q", err.Error())
	}
}

func TestRunProbeLogs_NameResolution(t *testing.T) {
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
			flags := &probeLogsFlags{limit: 50}
			err := runProbeLogs(t.Context(), tt.idArg, flags)

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

func TestRunProbeLogs_InvalidFlags(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name    string
		flags   *probeLogsFlags
		wantErr string
	}{
		{
			name:    "limit too low",
			flags:   &probeLogsFlags{limit: 0},
			wantErr: "invalid limit",
		},
		{
			name:    "limit too high",
			flags:   &probeLogsFlags{limit: 1001},
			wantErr: "invalid limit",
		},
		{
			name:    "invalid status",
			flags:   &probeLogsFlags{limit: 50, status: "invalid"},
			wantErr: "invalid status",
		},
		{
			name:    "invalid since",
			flags:   &probeLogsFlags{limit: 50, since: "bad"},
			wantErr: "invalid --since",
		},
		{
			name:    "invalid until",
			flags:   &probeLogsFlags{limit: 50, until: "bad"},
			wantErr: "invalid --until",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runProbeLogs(t.Context(), validUUID, tt.flags)

			if err == nil {
				t.Error("expected error, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestConvertToLogsOutput(t *testing.T) {
	probeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	resultID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")
	checkedAt := time.Now().Add(-1 * time.Hour)
	statusCode := 200
	errMsg := "connection timeout"

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
			{
				ID:             uuid.MustParse("660e8400-e29b-41d4-a716-446655440002"),
				ProbeID:        probeID,
				Region:         "eu-west-1",
				Status:         "failure",
				ResponseTimeMs: 5000,
				ErrorMessage:   &errMsg,
				CheckedAt:      checkedAt.Add(-5 * time.Minute),
			},
		},
		Total: 2,
		Page:  1,
		Limit: 50,
	}

	out := convertToLogsOutput(probeID, results)

	if out.ProbeID != probeID {
		t.Errorf("expected ProbeID %s, got %s", probeID, out.ProbeID)
	}

	if out.Total != 2 {
		t.Errorf("expected Total 2, got %d", out.Total)
	}

	if len(out.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out.Results))
	}

	// Verify first entry
	entry := out.Results[0]
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
	if entry.ErrorMessage != nil {
		t.Errorf("expected nil error message, got %v", entry.ErrorMessage)
	}

	// Verify second entry (failure)
	entry2 := out.Results[1]
	if entry2.Status != "failure" {
		t.Errorf("expected status 'failure', got %s", entry2.Status)
	}
	if entry2.ErrorMessage == nil || *entry2.ErrorMessage != "connection timeout" {
		t.Errorf("expected error message 'connection timeout', got %v", entry2.ErrorMessage)
	}
}

func TestConvertToLogsOutput_Empty(t *testing.T) {
	probeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	results := &client.ProbeResultListResponse{
		Results: []client.ProbeResult{},
		Total:   0,
		Page:    1,
		Limit:   50,
	}

	out := convertToLogsOutput(probeID, results)

	if len(out.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(out.Results))
	}

	if out.Total != 0 {
		t.Errorf("expected Total 0, got %d", out.Total)
	}
}
