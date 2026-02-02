// Package cmd implements the CLI commands for StackEye.
// Task #8068
package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

func TestNewProbeLabelCmd(t *testing.T) {
	cmd := NewProbeLabelCmd()

	if cmd.Use != "label <probe-id> <labels...>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "label <probe-id> <labels...>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestProbeLabelCmd_NoArgs(t *testing.T) {
	cmd := NewProbeLabelCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no arguments provided, got nil")
	}

	// Cobra's MinimumNArgs(2) produces a specific error message
	expectedMsg := "requires at least 2 arg"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestProbeLabelCmd_OnlyProbeID(t *testing.T) {
	cmd := NewProbeLabelCmd()
	cmd.SetArgs([]string{"api-health"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when only probe ID provided, got nil")
	}

	// Cobra's MinimumNArgs(2) produces a specific error message
	expectedMsg := "requires at least 2 arg"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error = %q, want to contain %q", err.Error(), expectedMsg)
	}
}

func TestParseLabelArgs_Empty(t *testing.T) {
	labels, err := parseLabelArgs([]string{})
	if err == nil {
		t.Error("Expected error for empty labels, got nil")
	}
	if labels != nil {
		t.Errorf("Expected nil labels, got %v", labels)
	}
}

func TestParseLabelArgs_KeyValueFormat(t *testing.T) {
	labels, err := parseLabelArgs([]string{"env=production"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(labels) != 1 {
		t.Fatalf("Expected 1 label, got %d", len(labels))
	}

	if labels[0].Key != "env" {
		t.Errorf("Key = %q, want %q", labels[0].Key, "env")
	}

	if labels[0].Value == nil || *labels[0].Value != "production" {
		t.Errorf("Value = %v, want %q", labels[0].Value, "production")
	}
}

func TestParseLabelArgs_KeyOnlyFormat(t *testing.T) {
	labels, err := parseLabelArgs([]string{"pci"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(labels) != 1 {
		t.Fatalf("Expected 1 label, got %d", len(labels))
	}

	if labels[0].Key != "pci" {
		t.Errorf("Key = %q, want %q", labels[0].Key, "pci")
	}

	if labels[0].Value != nil {
		t.Errorf("Value = %v, want nil for key-only label", labels[0].Value)
	}
}

func TestParseLabelArgs_MultipleLabels(t *testing.T) {
	labels, err := parseLabelArgs([]string{"env=production", "tier=backend", "pci"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(labels) != 3 {
		t.Fatalf("Expected 3 labels, got %d", len(labels))
	}

	// Check first label (key=value)
	if labels[0].Key != "env" || labels[0].Value == nil || *labels[0].Value != "production" {
		t.Errorf("First label = {%q, %v}, want {env, production}", labels[0].Key, labels[0].Value)
	}

	// Check second label (key=value)
	if labels[1].Key != "tier" || labels[1].Value == nil || *labels[1].Value != "backend" {
		t.Errorf("Second label = {%q, %v}, want {tier, backend}", labels[1].Key, labels[1].Value)
	}

	// Check third label (key-only)
	if labels[2].Key != "pci" || labels[2].Value != nil {
		t.Errorf("Third label = {%q, %v}, want {pci, nil}", labels[2].Key, labels[2].Value)
	}
}

func TestParseLabelArgs_EmptyValue(t *testing.T) {
	labels, err := parseLabelArgs([]string{"env="})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(labels) != 1 {
		t.Fatalf("Expected 1 label, got %d", len(labels))
	}

	if labels[0].Key != "env" {
		t.Errorf("Key = %q, want %q", labels[0].Key, "env")
	}

	// Empty value is allowed (key= format)
	if labels[0].Value == nil || *labels[0].Value != "" {
		t.Errorf("Value = %v, want empty string", labels[0].Value)
	}
}

func TestParseSingleLabel_InvalidKey(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantErr string
	}{
		{
			name:    "uppercase key",
			arg:     "ENV=production",
			wantErr: "invalid key format",
		},
		{
			name:    "key with spaces",
			arg:     "env name=production",
			wantErr: "invalid key format",
		},
		{
			name:    "key too long",
			arg:     strings.Repeat("a", 64) + "=value",
			wantErr: "at most 63 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseSingleLabel(tt.arg)
			if err == nil {
				t.Errorf("Expected error for %q, got nil", tt.arg)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseSingleLabel_ValidKeys(t *testing.T) {
	tests := []struct {
		name      string
		arg       string
		wantKey   string
		wantValue *string
	}{
		{
			name:      "simple key-value",
			arg:       "env=production",
			wantKey:   "env",
			wantValue: strPtr("production"),
		},
		{
			name:      "key-only",
			arg:       "pci",
			wantKey:   "pci",
			wantValue: nil,
		},
		{
			name:      "key with hyphens",
			arg:       "service-tier=web-frontend",
			wantKey:   "service-tier",
			wantValue: strPtr("web-frontend"),
		},
		{
			name:      "numeric value",
			arg:       "version=123",
			wantKey:   "version",
			wantValue: strPtr("123"),
		},
		{
			name:      "value with dots",
			arg:       "domain=api.example.com",
			wantKey:   "domain",
			wantValue: strPtr("api.example.com"),
		},
		{
			name:      "value with underscores",
			arg:       "region=us_east_1",
			wantKey:   "region",
			wantValue: strPtr("us_east_1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, err := parseSingleLabel(tt.arg)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if label.Key != tt.wantKey {
				t.Errorf("Key = %q, want %q", label.Key, tt.wantKey)
			}

			if tt.wantValue == nil {
				if label.Value != nil {
					t.Errorf("Value = %v, want nil", label.Value)
				}
			} else {
				if label.Value == nil || *label.Value != *tt.wantValue {
					t.Errorf("Value = %v, want %q", label.Value, *tt.wantValue)
				}
			}
		})
	}
}

func TestValidateLabelValue(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"simple value", "production", false},
		{"with hyphens", "web-frontend", false},
		{"with underscores", "us_east_1", false},
		{"with dots", "api.v2", false},
		{"numeric", "123", false},
		{"mixed", "v1.2.3-beta_1", false},
		{"max length", strings.Repeat("a", 63), false},
		{"too long", strings.Repeat("a", 64), true},
		{"with spaces", "my value", true},
		{"with special chars", "value@test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLabelValue(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLabelValue(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestProbeLabelCmd_Aliases(t *testing.T) {
	cmd := NewProbeLabelCmd()

	// label command shouldn't have aliases (avoid confusion with label key management)
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for label command, got %v", cmd.Aliases)
	}
}

// strPtr is a helper to create string pointers for test cases.
func strPtr(s string) *string {
	return &s
}

// Ensure ProbeLabelInput is compatible with our parsing
func TestProbeLabelInput_Compatibility(t *testing.T) {
	// This test ensures our parsed labels are compatible with the SDK type
	labels, err := parseLabelArgs([]string{"env=production", "pci"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The result should be a slice of client.ProbeLabelInput
	var _ []client.ProbeLabelInput = labels

	// Verify types
	if labels[0].Key == "" {
		t.Error("First label key should not be empty")
	}
}
