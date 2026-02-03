package cmd

import (
	"context"
	"slices"
	"strings"
	"testing"
)

func TestNewProbeListCmd(t *testing.T) {
	cmd := NewProbeListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use='list', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "List all monitoring probes" {
		t.Errorf("expected Short='List all monitoring probes', got %q", cmd.Short)
	}
}

func TestNewProbeListCmd_Aliases(t *testing.T) {
	cmd := NewProbeListCmd()

	expectedAliases := []string{"ls"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, expected := range expectedAliases {
		if !slices.Contains(cmd.Aliases, expected) {
			t.Errorf("expected alias %q not found", expected)
		}
	}
}

func TestNewProbeListCmd_Long(t *testing.T) {
	cmd := NewProbeListCmd()

	long := cmd.Long

	// Should contain status documentation
	statuses := []string{"up", "down", "degraded", "paused", "pending"}
	for _, status := range statuses {
		if !strings.Contains(long, status) {
			t.Errorf("expected Long description to mention status %q", status)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye probe list") {
		t.Error("expected Long description to contain example commands")
	}
}

func TestNewProbeListCmd_Flags(t *testing.T) {
	cmd := NewProbeListCmd()

	// Verify expected flags exist
	flags := []struct {
		name         string
		shorthand    string
		defaultValue string
	}{
		{"status", "", ""},
		{"page", "", "1"},
		{"limit", "", "20"},
		{"period", "", ""},
		{"labels", "", ""}, // Task #8070
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

func TestRunProbeList_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	// Since runProbeList requires an API client, validation happens first
	// and will fail before making API calls for invalid inputs.
	tests := []struct {
		name         string
		limit        int
		page         int
		status       string
		period       string
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
			name:         "invalid status",
			limit:        20,
			page:         1,
			status:       "badstatus",
			wantErrorMsg: `invalid value "badstatus" for --status: must be one of: up, down, degraded, paused, pending`,
		},
		{
			name:         "invalid period",
			limit:        20,
			page:         1,
			period:       "1h",
			wantErrorMsg: `invalid value "1h" for --period: must be one of: 24h, 7d, 30d`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &probeListFlags{
				page:   tt.page,
				limit:  tt.limit,
				status: tt.status,
				period: tt.period,
			}

			// Call runProbeList with a background context.
			// It should fail on validation before needing API client.
			err := runProbeList(context.Background(), flags)

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

func TestRunProbeList_ValidFlags(t *testing.T) {
	// Test that valid flags pass validation (will fail later on API client)
	flags := &probeListFlags{
		page:   1,
		limit:  20,
		status: "up",
		period: "24h",
	}

	err := runProbeList(context.Background(), flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	validationErrors := []string{
		"invalid limit",
		"invalid page",
		"invalid status",
		"invalid period",
	}
	for _, ve := range validationErrors {
		if strings.Contains(err.Error(), ve) {
			t.Errorf("got unexpected validation error: %s", err.Error())
		}
	}
}

// Task #8070: Tests for parseLabelFilters function
func TestParseLabelFilters(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:  "single key=value",
			input: "env=production",
			want:  map[string]string{"env": "production"},
		},
		{
			name:  "single key-only",
			input: "pci",
			want:  map[string]string{"pci": ""},
		},
		{
			name:  "multiple key=value",
			input: "env=production,tier=web",
			want:  map[string]string{"env": "production", "tier": "web"},
		},
		{
			name:  "mixed key=value and key-only",
			input: "env=production,pci,tier=web",
			want:  map[string]string{"env": "production", "pci": "", "tier": "web"},
		},
		{
			name:  "handles whitespace",
			input: "env=production, tier=web , pci",
			want:  map[string]string{"env": "production", "tier": "web", "pci": ""},
		},
		{
			name:  "empty value",
			input: "env=",
			want:  map[string]string{"env": ""},
		},
		{
			name:    "invalid key with spaces",
			input:   "env name=production",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLabelFilters(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLabelFilters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("parseLabelFilters() got %d filters, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if gotV, ok := got[k]; !ok || gotV != v {
					t.Errorf("parseLabelFilters()[%q] = %q, want %q", k, gotV, v)
				}
			}
		})
	}
}

func TestRunProbeList_ValidLabelsFlag(t *testing.T) {
	// Test that valid labels flag passes validation (will fail later on API client)
	flags := &probeListFlags{
		page:   1,
		limit:  20,
		labels: "env=production,tier=web",
	}

	err := runProbeList(context.Background(), flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a label validation error
	if strings.Contains(err.Error(), "label") {
		t.Errorf("got unexpected label validation error: %s", err.Error())
	}
}
