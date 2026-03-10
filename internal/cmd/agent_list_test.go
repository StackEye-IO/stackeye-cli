// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewAgentListCmd verifies the list command is correctly configured.
func TestNewAgentListCmd(t *testing.T) {
	cmd := NewAgentListCmd()

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

// TestNewAgentListCmd_Aliases verifies the 'ls' alias is set.
func TestNewAgentListCmd_Aliases(t *testing.T) {
	cmd := NewAgentListCmd()

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

// TestNewAgentListCmd_Long verifies example commands appear in Long.
func TestNewAgentListCmd_Long(t *testing.T) {
	cmd := NewAgentListCmd()

	if !strings.Contains(cmd.Long, "stackeye agent list") {
		t.Error("expected Long description to contain example commands")
	}
	if !strings.Contains(cmd.Long, "-o json") {
		t.Error("expected Long description to mention JSON output option")
	}
}

// TestTruncateAgentField verifies field truncation.
func TestTruncateAgentField(t *testing.T) {
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
		{"agent name", "prod-web-01", 20, "prod-web-01"},
		{"long name truncated", "prod-very-long-hostname-example", 20, "prod-very-long-ho..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateAgentField(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateAgentField(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestFormatAgentDate verifies ISO 8601 date formatting.
func TestFormatAgentDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid timestamp", "2026-01-15T10:30:00Z", "Jan 15 10:30"},
		{"end of year", "2025-12-31T23:59:59Z", "Dec 31 23:59"},
		{"invalid input returned as-is", "not-a-date", "not-a-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAgentDate(tt.input)
			if result != tt.expected {
				t.Errorf("formatAgentDate(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestAgentStatus verifies the status derivation logic.
func TestAgentStatus(t *testing.T) {
	recentTime := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
	staleTime := time.Now().Add(-10 * time.Minute).Format(time.RFC3339)
	malformedTime := "not-a-valid-timestamp"

	tests := []struct {
		name     string
		agent    client.Agent
		expected string
	}{
		{
			name:     "inactive agent",
			agent:    client.Agent{IsActive: false},
			expected: "inactive",
		},
		{
			name:     "pending agent (never seen)",
			agent:    client.Agent{IsActive: true, LastSeenAt: nil},
			expected: "pending",
		},
		{
			name:     "online agent (seen recently)",
			agent:    client.Agent{IsActive: true, LastSeenAt: &recentTime},
			expected: "online",
		},
		{
			name:     "offline agent (stale heartbeat)",
			agent:    client.Agent{IsActive: true, LastSeenAt: &staleTime},
			expected: "offline",
		},
		{
			name:     "unknown status (malformed timestamp)",
			agent:    client.Agent{IsActive: true, LastSeenAt: &malformedTime},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agentStatus(&tt.agent)
			if result != tt.expected {
				t.Errorf("agentStatus() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestPrintAgentList_DoesNotPanic verifies printAgentList is panic-safe.
func TestPrintAgentList_DoesNotPanic(t *testing.T) {
	desc := "Test agent"
	host := "prod-web-01"
	ip := "10.0.0.5"
	ver := "1.0.0"
	lastSeen := "2026-01-01T00:00:00Z"

	tests := []struct {
		name     string
		response *client.AgentListResponse
	}{
		{
			name: "empty list",
			response: &client.AgentListResponse{
				Status: "success",
				Data:   []client.Agent{},
				Meta: struct {
					Total int `json:"total"`
				}{Total: 0},
			},
		},
		{
			name: "single active agent with all fields",
			response: &client.AgentListResponse{
				Status: "success",
				Data: []client.Agent{
					{
						ID:           "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
						Name:         "prod-web-01",
						Description:  &desc,
						KeyPrefix:    "se_ag_ab12",
						LastSeenAt:   &lastSeen,
						AgentVersion: &ver,
						Hostname:     &host,
						IPAddress:    &ip,
						IsActive:     true,
						CreatedAt:    "2026-01-01T00:00:00Z",
						UpdatedAt:    "2026-01-01T00:00:00Z",
					},
				},
				Meta: struct {
					Total int `json:"total"`
				}{Total: 1},
			},
		},
		{
			name: "inactive agent with minimal fields",
			response: &client.AgentListResponse{
				Status: "success",
				Data: []client.Agent{
					{
						ID:        "b2c3d4e5-f6a7-8901-bcde-f12345678901",
						Name:      "deactivated-agent",
						KeyPrefix: "se_ag_cd34",
						IsActive:  false,
						CreatedAt: "2026-01-01T00:00:00Z",
						UpdatedAt: "2026-01-01T00:00:00Z",
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
					t.Errorf("printAgentList panicked: %v", r)
				}
			}()
			printAgentList(tt.response)
		})
	}
}
