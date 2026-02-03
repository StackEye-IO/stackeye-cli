package telemetry

import (
	"testing"
	"time"
)

func TestSanitizeCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple command",
			input:    "probe list",
			expected: "probe list",
		},
		{
			name:     "command with flags",
			input:    "probe create --name test",
			expected: "probe create",
		},
		{
			name:     "command with short flags",
			input:    "probe list -o json",
			expected: "probe list",
		},
		{
			name:     "single word",
			input:    "version",
			expected: "version",
		},
		{
			name:     "empty command",
			input:    "",
			expected: "unknown",
		},
		{
			name:     "command with multiple flags",
			input:    "channel create --name test --type slack --webhook-url http://example.com",
			expected: "channel create",
		},
		{
			name:     "nested subcommand",
			input:    "admin m2m-key create --global",
			expected: "admin m2m-key create",
		},
		{
			name:     "command with equal sign flag",
			input:    "probe get --id=abc123",
			expected: "probe get",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeCommand(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHashOrgID(t *testing.T) {
	tests := []struct {
		name  string
		orgID string
	}{
		{
			name:  "standard org ID",
			orgID: "org_abc123def456",
		},
		{
			name:  "empty org ID",
			orgID: "",
		},
		{
			name:  "long org ID",
			orgID: "org_very_long_organization_identifier_that_is_quite_lengthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hashOrgID(tt.orgID)
			// Hash prefix should be 16 characters
			if len(result) != 16 {
				t.Errorf("hashOrgID(%q) length = %d, want 16", tt.orgID, len(result))
			}
			// Same input should produce same output
			result2 := hashOrgID(tt.orgID)
			if result != result2 {
				t.Errorf("hashOrgID(%q) not deterministic: %q != %q", tt.orgID, result, result2)
			}
		})
	}
}

func TestHashOrgID_Consistency(t *testing.T) {
	// Verify consistent results for same input
	orgID := "org_test123"
	hash1 := hashOrgID(orgID)
	hash2 := hashOrgID(orgID)

	if hash1 != hash2 {
		t.Errorf("hashOrgID not deterministic: %q != %q", hash1, hash2)
	}
}

func TestHashOrgID_UniqueForDifferentInputs(t *testing.T) {
	hash1 := hashOrgID("org_abc")
	hash2 := hashOrgID("org_def")

	if hash1 == hash2 {
		t.Errorf("hashOrgID produced same hash for different inputs: %q", hash1)
	}
}

func TestClient_IsEnabled(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client := GetClient()

	// Initially disabled (no config)
	if client.IsEnabled() {
		t.Error("expected telemetry to be disabled by default")
	}

	// Enable
	client.SetEnabled(true)
	if !client.IsEnabled() {
		t.Error("expected telemetry to be enabled after SetEnabled(true)")
	}

	// Disable
	client.SetEnabled(false)
	if client.IsEnabled() {
		t.Error("expected telemetry to be disabled after SetEnabled(false)")
	}
}

func TestClient_Singleton(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client1 := GetClient()
	client2 := GetClient()

	if client1 != client2 {
		t.Error("GetClient should return the same instance")
	}
}

func TestClient_SetEndpoint(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client := GetClient()
	customEndpoint := "http://localhost:8080/telemetry"
	client.SetEndpoint(customEndpoint)

	// Verify endpoint was set (indirectly by checking the client)
	if client.endpoint != customEndpoint {
		t.Errorf("SetEndpoint failed: got %q, want %q", client.endpoint, customEndpoint)
	}
}

func TestEvent_Fields(t *testing.T) {
	now := time.Now()
	event := &Event{
		CLIVersion: "1.0.0",
		Command:    "probe list",
		ExitCode:   0,
		DurationMs: 1234,
		OS:         "linux",
		Arch:       "amd64",
		OrgIDHash:  "abcdef1234567890",
	}

	if event.CLIVersion != "1.0.0" {
		t.Errorf("expected CLIVersion='1.0.0', got %q", event.CLIVersion)
	}
	if event.Command != "probe list" {
		t.Errorf("expected Command='probe list', got %q", event.Command)
	}
	if event.ExitCode != 0 {
		t.Errorf("expected ExitCode=0, got %d", event.ExitCode)
	}
	if event.DurationMs != 1234 {
		t.Errorf("expected DurationMs=1234, got %d", event.DurationMs)
	}
	if event.OS != "linux" {
		t.Errorf("expected OS='linux', got %q", event.OS)
	}
	if event.Arch != "amd64" {
		t.Errorf("expected Arch='amd64', got %q", event.Arch)
	}
	if event.OrgIDHash != "abcdef1234567890" {
		t.Errorf("expected OrgIDHash='abcdef1234567890', got %q", event.OrgIDHash)
	}

	_ = now // Silence unused variable warning
}

func TestClient_Track_WhenDisabled(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client := GetClient()
	client.SetEnabled(false)

	// Track should be a no-op when disabled
	client.Track(nil, "test command", 0, time.Second)

	// Should not panic and should not set lastEvent
	if client.GetLastEvent() != nil {
		t.Error("expected no event to be tracked when telemetry is disabled")
	}
}
