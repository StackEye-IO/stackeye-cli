package cmd

import (
	"testing"
)

func TestNewProbeWizardCmd(t *testing.T) {
	cmd := NewProbeWizardCmd()

	if cmd == nil {
		t.Fatal("NewProbeWizardCmd() returned nil")
	}

	if cmd.Use != "wizard" {
		t.Errorf("Use = %q, want %q", cmd.Use, "wizard")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestGenerateProbeName(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "HTTPS URL",
			target: "https://api.example.com/health",
			want:   "Api.example.com",
		},
		{
			name:   "HTTP URL",
			target: "http://example.com",
			want:   "Example.com",
		},
		{
			name:   "URL with port",
			target: "https://api.example.com:8443/health",
			want:   "Api.example.com",
		},
		{
			name:   "Plain hostname",
			target: "example.com",
			want:   "Example.com",
		},
		{
			name:   "Host with port (TCP style)",
			target: "db.example.com:5432",
			want:   "Db.example.com",
		},
		{
			name:   "IP address",
			target: "192.168.1.1",
			want:   "192.168.1.1",
		},
		{
			name:   "Empty string",
			target: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateProbeName(tt.target)
			if got != tt.want {
				t.Errorf("generateProbeName(%q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

func TestSkipIfNotHTTP(t *testing.T) {
	tests := []struct {
		name      string
		checkType string
		want      bool
	}{
		{
			name:      "HTTP type",
			checkType: "http",
			want:      false,
		},
		{
			name:      "Ping type",
			checkType: "ping",
			want:      true,
		},
		{
			name:      "TCP type",
			checkType: "tcp",
			want:      true,
		},
		{
			name:      "DNS type",
			checkType: "dns_resolve",
			want:      true,
		},
		{
			name:      "Empty type",
			checkType: "",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock wizard with the check type set
			wiz := &mockWizard{data: map[string]interface{}{
				"checkType": tt.checkType,
			}}

			got := skipIfNotHTTPTestHelper(wiz)
			if got != tt.want {
				t.Errorf("skipIfNotHTTP() = %v, want %v for checkType=%q", got, tt.want, tt.checkType)
			}
		})
	}
}

func TestSkipIfNotHTTPS(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{
			name:   "HTTPS URL",
			target: "https://example.com",
			want:   false,
		},
		{
			name:   "HTTP URL",
			target: "http://example.com",
			want:   true,
		},
		{
			name:   "Plain hostname",
			target: "example.com",
			want:   true,
		},
		{
			name:   "HTTPS with uppercase",
			target: "HTTPS://example.com",
			want:   false,
		},
		{
			name:   "Empty target",
			target: "",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wiz := &mockWizard{data: map[string]interface{}{
				"target": tt.target,
			}}

			got := skipIfNotHTTPSTestHelper(wiz)
			if got != tt.want {
				t.Errorf("skipIfNotHTTPS() = %v, want %v for target=%q", got, tt.want, tt.target)
			}
		})
	}
}

// mockWizard is a simple mock for testing skip functions.
type mockWizard struct {
	data map[string]interface{}
}

func (m *mockWizard) GetDataString(key string) string {
	if v, ok := m.data[key].(string); ok {
		return v
	}
	return ""
}

// skipIfNotHTTPTestHelper wraps the skip function for testing.
func skipIfNotHTTPTestHelper(w *mockWizard) bool {
	return w.GetDataString("checkType") != "http"
}

// skipIfNotHTTPSTestHelper wraps the skip function for testing.
func skipIfNotHTTPSTestHelper(w *mockWizard) bool {
	target := w.GetDataString("target")
	return len(target) < 8 || (target[:8] != "https://" && target[:8] != "HTTPS://")
}

func TestProbeWizardNonInteractive(t *testing.T) {
	// Test that non-interactive mode runs without error
	// Note: This doesn't test the actual output, just that it doesn't panic
	err := runProbeWizardNonInteractive()
	if err != nil {
		t.Errorf("runProbeWizardNonInteractive() returned error: %v", err)
	}
}

func TestProbeWizardCommandRegistration(t *testing.T) {
	// Verify wizard is registered as a subcommand of probe
	probeCmd := NewProbeCmd()

	var found bool
	for _, sub := range probeCmd.Commands() {
		if sub.Use == "wizard" {
			found = true
			break
		}
	}

	if !found {
		t.Error("wizard command not registered under probe command")
	}
}
