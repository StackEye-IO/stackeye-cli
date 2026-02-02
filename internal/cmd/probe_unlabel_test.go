// Package cmd implements the CLI commands for StackEye.
// Task #8069
package cmd

import (
	"strings"
	"testing"
)

func TestNewProbeUnlabelCmd(t *testing.T) {
	cmd := NewProbeUnlabelCmd()

	if cmd.Use != "unlabel <probe-id> <keys...>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "unlabel <probe-id> <keys...>")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestProbeUnlabelCmd_NoArgs(t *testing.T) {
	cmd := NewProbeUnlabelCmd()
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

func TestProbeUnlabelCmd_OnlyProbeID(t *testing.T) {
	cmd := NewProbeUnlabelCmd()
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

func TestValidateLabelKeyForRemoval_Empty(t *testing.T) {
	err := validateLabelKeyForRemoval("")
	if err == nil {
		t.Error("Expected error for empty key, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Error = %q, want to contain 'cannot be empty'", err.Error())
	}
}

func TestValidateLabelKeyForRemoval_TooLong(t *testing.T) {
	longKey := strings.Repeat("a", 64)
	err := validateLabelKeyForRemoval(longKey)
	if err == nil {
		t.Error("Expected error for key exceeding 63 characters, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds maximum length") {
		t.Errorf("Error = %q, want to contain 'exceeds maximum length'", err.Error())
	}
}

func TestValidateLabelKeyForRemoval_ValidKeys(t *testing.T) {
	tests := []string{
		"env",
		"tier",
		"pci",
		"service-tier",
		"region-us-east-1",
		strings.Repeat("a", 63), // Max length
	}

	for _, key := range tests {
		t.Run(key, func(t *testing.T) {
			err := validateLabelKeyForRemoval(key)
			if err != nil {
				t.Errorf("validateLabelKeyForRemoval(%q) unexpected error: %v", key, err)
			}
		})
	}
}

func TestIsLabelNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "nil error",
			errMsg:   "",
			expected: false,
		},
		{
			name:     "404 error",
			errMsg:   "HTTP 404: label not found",
			expected: true,
		},
		{
			name:     "label_not_found code",
			errMsg:   "error: label_not_found",
			expected: true,
		},
		{
			name:     "generic not found",
			errMsg:   "resource not found",
			expected: true,
		},
		{
			name:     "unrelated error",
			errMsg:   "connection timeout",
			expected: false,
		},
		{
			name:     "auth error",
			errMsg:   "HTTP 401: unauthorized",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMsg != "" {
				err = &testError{msg: tt.errMsg}
			}
			result := isLabelNotFoundError(err)
			if result != tt.expected {
				t.Errorf("isLabelNotFoundError(%q) = %v, want %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}

// testError is a simple error type for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestContainsUnlabel(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"", "", true},
		{"hello", "", true},
		{"", "hello", false},
		{"HTTP 404", "404", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			result := containsUnlabel(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("containsUnlabel(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestProbeUnlabelCmd_Aliases(t *testing.T) {
	cmd := NewProbeUnlabelCmd()

	// unlabel command shouldn't have aliases
	if len(cmd.Aliases) != 0 {
		t.Errorf("Expected no aliases for unlabel command, got %v", cmd.Aliases)
	}
}
