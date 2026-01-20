package cmd

import (
	"testing"
)

func TestNewProbeGetCmd(t *testing.T) {
	cmd := NewProbeGetCmd()

	// Verify command basic properties
	if cmd.Use != "get <id>" {
		t.Errorf("expected Use to be 'get <id>', got %q", cmd.Use)
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

	// Verify period flag exists
	flag := cmd.Flags().Lookup("period")
	if flag == nil {
		t.Error("expected 'period' flag to be defined")
	} else {
		if flag.DefValue != "" {
			t.Errorf("expected period flag default to be empty, got %q", flag.DefValue)
		}
	}
}

func TestRunProbeGet_InvalidUUID(t *testing.T) {
	tests := []struct {
		name    string
		idArg   string
		wantErr string
	}{
		{
			name:    "empty string",
			idArg:   "",
			wantErr: "invalid probe ID",
		},
		{
			name:    "not a UUID",
			idArg:   "not-a-uuid",
			wantErr: "invalid probe ID",
		},
		{
			name:    "partial UUID",
			idArg:   "550e8400-e29b-41d4",
			wantErr: "invalid probe ID",
		},
		{
			name:    "probe name instead of UUID",
			idArg:   "api-health",
			wantErr: "invalid probe ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &probeGetFlags{}
			err := runProbeGet(t.Context(), tt.idArg, flags)

			if err == nil {
				t.Error("expected error for invalid UUID, got nil")
				return
			}

			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestRunProbeGet_InvalidPeriod(t *testing.T) {
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
			name:    "wrong format",
			period:  "7days",
			wantErr: "invalid period",
		},
		{
			name:    "numeric only",
			period:  "7",
			wantErr: "invalid period",
		},
		{
			name:    "unsupported period",
			period:  "1h",
			wantErr: "invalid period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &probeGetFlags{period: tt.period}
			err := runProbeGet(t.Context(), validUUID, flags)

			if err == nil {
				t.Error("expected error for invalid period, got nil")
				return
			}

			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestRunProbeGet_ValidPeriods(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	validPeriods := []string{"24h", "7d", "30d"}

	for _, period := range validPeriods {
		t.Run(period, func(t *testing.T) {
			flags := &probeGetFlags{period: period}
			err := runProbeGet(t.Context(), validUUID, flags)

			// We expect an API client error since we're not mocking,
			// but we should NOT get a period validation error
			if err != nil && contains(err.Error(), "invalid period") {
				t.Errorf("period %q should be valid, got validation error: %v", period, err)
			}
		})
	}
}

// Note: contains() helper is defined in config_test.go within the same package
