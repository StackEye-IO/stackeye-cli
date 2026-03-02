// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewPrivateRegionListCmd verifies the list command is correctly configured.
func TestNewPrivateRegionListCmd(t *testing.T) {
	cmd := NewPrivateRegionListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use='list', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewPrivateRegionListCmd_Aliases verifies the 'ls' alias is set.
func TestNewPrivateRegionListCmd_Aliases(t *testing.T) {
	cmd := NewPrivateRegionListCmd()

	found := false
	for _, a := range cmd.Aliases {
		if a == "ls" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'ls' alias to be set")
	}
}

// TestNewPrivateRegionListCmd_Long verifies example commands appear in Long.
func TestNewPrivateRegionListCmd_Long(t *testing.T) {
	cmd := NewPrivateRegionListCmd()

	if !strings.Contains(cmd.Long, "stackeye private-region list") {
		t.Error("expected Long description to contain example commands")
	}
	if !strings.Contains(cmd.Long, "-o json") {
		t.Error("expected Long description to mention JSON output option")
	}
}

// TestTruncatePrivateRegionField verifies field truncation.
func TestTruncatePrivateRegionField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"shorter than max", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"longer truncated", "hello world", 8, "hello..."},
		{"very short max", "hello", 3, "hel"},
		{"empty string", "", 10, ""},
		{"one char max", "hello", 1, "h"},
		{"two char max", "hello", 2, "he"},
		{"region id", "prv-nyc-office", 20, "prv-nyc-office"},
		{"region id truncated", "prv-very-long-region-name", 20, "prv-very-long-reg..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncatePrivateRegionField(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncatePrivateRegionField(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestFormatPrivateRegionDate verifies ISO 8601 date formatting.
func TestFormatPrivateRegionDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid timestamp", "2026-01-15T10:30:00Z", "Jan 15, 2026"},
		{"different month", "2026-06-01T00:00:00Z", "Jun 01, 2026"},
		{"end of year", "2025-12-31T23:59:59Z", "Dec 31, 2025"},
		{"invalid input returned as-is", "not-a-date", "not-a-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPrivateRegionDate(tt.input)
			if result != tt.expected {
				t.Errorf("formatPrivateRegionDate(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestPrintPrivateRegionList_DoesNotPanic verifies printPrivateRegionList is panic-safe.
func TestPrintPrivateRegionList_DoesNotPanic(t *testing.T) {
	city := "New York"
	tests := []struct {
		name     string
		response *client.PrivateRegionListResponse
	}{
		{
			name: "empty list",
			response: &client.PrivateRegionListResponse{
				Status: "success",
				Data:   []client.PrivateRegion{},
				Meta: struct {
					Total int `json:"total"`
				}{Total: 0},
			},
		},
		{
			name: "single region with city",
			response: &client.PrivateRegionListResponse{
				Status: "success",
				Data: []client.PrivateRegion{
					{
						ID:          "prv-nyc-office",
						Name:        "nyc-office",
						DisplayName: "NYC Office",
						Continent:   "North America",
						CountryCode: "US",
						City:        &city,
						Scope:       "private",
						Status:      "active",
						CreatedAt:   "2026-01-01T00:00:00Z",
						UpdatedAt:   "2026-01-01T00:00:00Z",
					},
				},
				Meta: struct {
					Total int `json:"total"`
				}{Total: 1},
			},
		},
		{
			name: "region without city",
			response: &client.PrivateRegionListResponse{
				Status: "success",
				Data: []client.PrivateRegion{
					{
						ID:          "prv-remote",
						Name:        "remote",
						DisplayName: "Remote DC",
						Continent:   "North America",
						CountryCode: "US",
						City:        nil,
						Scope:       "private",
						Status:      "active",
						CreatedAt:   "2026-01-01T00:00:00Z",
						UpdatedAt:   "2026-01-01T00:00:00Z",
					},
				},
				Meta: struct {
					Total int `json:"total"`
				}{Total: 1},
			},
		},
		{
			name: "region with very long display name",
			response: &client.PrivateRegionListResponse{
				Status: "success",
				Data: []client.PrivateRegion{
					{
						ID:          "prv-very-long-region-id",
						Name:        "very-long-region-id",
						DisplayName: "A Very Long Display Name That Exceeds Normal Length",
						Continent:   "Europe",
						CountryCode: "GB",
						Scope:       "private",
						Status:      "inactive",
						CreatedAt:   "2026-01-01T00:00:00Z",
						UpdatedAt:   "2026-01-01T00:00:00Z",
					},
				},
				Meta: struct {
					Total int `json:"total"`
				}{Total: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printPrivateRegionList panicked: %v", r)
				}
			}()
			printPrivateRegionList(tt.response)
		})
	}
}
