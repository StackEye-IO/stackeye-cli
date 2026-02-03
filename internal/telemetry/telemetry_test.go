package telemetry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
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
	client.Track(context.TODO(), "test command", 0, time.Second)

	// Should not panic and should not set lastEvent
	if client.GetLastEvent() != nil {
		t.Error("expected no event to be tracked when telemetry is disabled")
	}
}

func TestClient_Track_WhenEnabled(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client := GetClient()
	client.SetEnabled(true)

	// Use a non-existent endpoint to avoid actual network calls
	client.SetEndpoint("http://127.0.0.1:1")

	// Track should store the event
	client.Track(context.TODO(), "probe list --verbose", 0, 500*time.Millisecond)

	// Give async goroutine time to run (even though it will fail)
	time.Sleep(10 * time.Millisecond)

	// Should set lastEvent
	event := client.GetLastEvent()
	if event == nil {
		t.Fatal("expected event to be tracked when telemetry is enabled")
	}

	if event.Command != "probe list" {
		t.Errorf("expected command='probe list', got %q", event.Command)
	}
	if event.ExitCode != 0 {
		t.Errorf("expected exit code=0, got %d", event.ExitCode)
	}
	if event.DurationMs != 500 {
		t.Errorf("expected duration=500, got %d", event.DurationMs)
	}
	if event.OS == "" {
		t.Error("expected OS to be set")
	}
	if event.Arch == "" {
		t.Error("expected Arch to be set")
	}
}

func TestClient_GetLastEvent_NilInitially(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client := GetClient()
	if client.GetLastEvent() != nil {
		t.Error("expected nil for last event initially")
	}
}

func TestClient_BuildEvent(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client := GetClient()
	event := client.buildEvent("admin worker-key list", 1, 2*time.Second)

	if event.Command != "admin worker-key list" {
		t.Errorf("expected command='admin worker-key list', got %q", event.Command)
	}
	if event.ExitCode != 1 {
		t.Errorf("expected exit code=1, got %d", event.ExitCode)
	}
	if event.DurationMs != 2000 {
		t.Errorf("expected duration=2000, got %d", event.DurationMs)
	}
	if event.CLIVersion == "" {
		t.Error("expected CLIVersion to be set")
	}
	if event.OS == "" {
		t.Error("expected OS to be set")
	}
	if event.Arch == "" {
		t.Error("expected Arch to be set")
	}
}

func TestClient_Reload(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client := GetClient()

	// Should not panic
	client.Reload()

	// Should still be disabled after reload (no config)
	if client.IsEnabled() {
		t.Error("expected telemetry to remain disabled after reload with no config")
	}
}

func TestDefaultEndpoint(t *testing.T) {
	if DefaultEndpoint != "https://api.stackeye.io/v1/telemetry/cli" {
		t.Errorf("expected default endpoint to be production URL, got %q", DefaultEndpoint)
	}
}

func TestEnvTelemetryConstant(t *testing.T) {
	if EnvTelemetry != "STACKEYE_TELEMETRY" {
		t.Errorf("expected env constant to be STACKEYE_TELEMETRY, got %q", EnvTelemetry)
	}
}

func TestClient_LoadConfig_NoConfig(t *testing.T) {
	ResetClient()
	defer ResetClient()

	// Fresh client with no config should be disabled
	client := GetClient()
	if client.IsEnabled() {
		t.Error("expected telemetry to be disabled when no config exists")
	}
}

func TestSanitizeCommand_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "only spaces",
			input:    "   ",
			expected: "unknown",
		},
		{
			name:     "command with multiple spaces",
			input:    "probe  list",
			expected: "probe list",
		},
		{
			name:     "leading space",
			input:    " probe list",
			expected: "probe list",
		},
		{
			name:     "trailing space",
			input:    "probe list ",
			expected: "probe list",
		},
		{
			name:     "flag at start",
			input:    "--help",
			expected: "unknown",
		},
		{
			name:     "slash at start",
			input:    "/path/to/file",
			expected: "unknown",
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

func TestClient_ConcurrentAccess(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client := GetClient()
	client.SetEnabled(true)
	client.SetEndpoint("http://127.0.0.1:1")

	// Run concurrent operations to test thread safety
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			client.IsEnabled()
			client.SetEnabled(idx%2 == 0)
			client.GetLastEvent()
			client.Track(context.TODO(), "test", idx, time.Millisecond)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	// Test passes if no race conditions or panics
}

func TestLoadConfig_EnvVarDisable(t *testing.T) {
	ResetClient()
	defer ResetClient()
	defer os.Unsetenv(EnvTelemetry)

	// Test various disable values
	disableValues := []string{"0", "false", "no", "off", "FALSE", "NO", "OFF"}
	for _, val := range disableValues {
		t.Run("disable_"+val, func(t *testing.T) {
			ResetClient()
			os.Setenv(EnvTelemetry, val)
			client := GetClient()
			if client.IsEnabled() {
				t.Errorf("expected telemetry disabled with STACKEYE_TELEMETRY=%s", val)
			}
		})
	}
}

func TestLoadConfig_EnvVarEnable(t *testing.T) {
	ResetClient()
	defer ResetClient()
	defer os.Unsetenv(EnvTelemetry)

	// Test various enable values
	enableValues := []string{"1", "true", "yes", "on", "TRUE", "YES", "ON"}
	for _, val := range enableValues {
		t.Run("enable_"+val, func(t *testing.T) {
			ResetClient()
			os.Setenv(EnvTelemetry, val)
			client := GetClient()
			if !client.IsEnabled() {
				t.Errorf("expected telemetry enabled with STACKEYE_TELEMETRY=%s", val)
			}
		})
	}
}

func TestLoadConfig_EnvVarUnknownValue(t *testing.T) {
	ResetClient()
	defer ResetClient()
	defer os.Unsetenv(EnvTelemetry)

	// Unknown env value should fall through to config
	os.Setenv(EnvTelemetry, "maybe")
	client := GetClient()
	// Without config, should be disabled
	if client.IsEnabled() {
		t.Error("expected telemetry disabled with unknown env value and no config")
	}
}

func TestClient_Reload_ReloadsFromEnv(t *testing.T) {
	ResetClient()
	defer ResetClient()
	defer os.Unsetenv(EnvTelemetry)

	client := GetClient()

	// Initially disabled
	if client.IsEnabled() {
		t.Error("expected initially disabled")
	}

	// Set env and reload
	os.Setenv(EnvTelemetry, "1")
	client.Reload()

	if !client.IsEnabled() {
		t.Error("expected enabled after reload with env=1")
	}

	// Disable via env and reload
	os.Setenv(EnvTelemetry, "0")
	client.Reload()

	if client.IsEnabled() {
		t.Error("expected disabled after reload with env=0")
	}
}

func TestClient_Send_Success(t *testing.T) {
	ResetClient()
	defer ResetClient()

	// Create a test server that accepts telemetry events
	var receivedEvent *Event
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("expected User-Agent header")
		}

		// Decode the event
		receivedEvent = &Event{}
		if err := json.NewDecoder(r.Body).Decode(receivedEvent); err != nil {
			t.Errorf("failed to decode event: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := GetClient()
	client.SetEndpoint(server.URL)

	event := &Event{
		CLIVersion: "1.0.0",
		Command:    "test command",
		ExitCode:   0,
		DurationMs: 100,
		OS:         "linux",
		Arch:       "amd64",
	}

	err := client.send(context.Background(), event)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	// Verify event was received
	if receivedEvent == nil {
		t.Fatal("expected event to be received")
	}
	if receivedEvent.Command != "test command" {
		t.Errorf("expected command='test command', got %q", receivedEvent.Command)
	}
}

func TestClient_Send_ServerError(t *testing.T) {
	ResetClient()
	defer ResetClient()

	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := GetClient()
	client.SetEndpoint(server.URL)

	event := &Event{
		CLIVersion: "1.0.0",
		Command:    "test",
		ExitCode:   0,
	}

	// Send should not return error for server errors (fire and forget)
	err := client.send(context.Background(), event)
	if err != nil {
		t.Fatalf("send should not fail on server error: %v", err)
	}
}

func TestClient_Send_InvalidEndpoint(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client := GetClient()
	client.SetEndpoint("http://127.0.0.1:1") // Invalid port

	event := &Event{
		CLIVersion: "1.0.0",
		Command:    "test",
		ExitCode:   0,
	}

	// Send should return error for network failure
	err := client.send(context.Background(), event)
	if err == nil {
		t.Error("expected error for invalid endpoint")
	}
}

func TestClient_Send_ContextCanceled(t *testing.T) {
	ResetClient()
	defer ResetClient()

	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := GetClient()
	client.SetEndpoint(server.URL)

	event := &Event{
		CLIVersion: "1.0.0",
		Command:    "test",
		ExitCode:   0,
	}

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.send(ctx, event)
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestClient_Track_StoresLastEvent(t *testing.T) {
	ResetClient()
	defer ResetClient()

	client := GetClient()
	client.SetEnabled(true)
	client.SetEndpoint("http://127.0.0.1:1") // Won't actually connect

	client.Track(context.TODO(), "probe list", 0, 500*time.Millisecond)

	// Give async goroutine time to set lastEvent
	time.Sleep(20 * time.Millisecond)

	event := client.GetLastEvent()
	if event == nil {
		t.Fatal("expected last event to be set")
	}
	if event.Command != "probe list" {
		t.Errorf("expected command='probe list', got %q", event.Command)
	}
}

func TestClient_Flush_WaitsForPendingSends(t *testing.T) {
	ResetClient()
	defer ResetClient()

	// Create a server that responds slowly
	var sendCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&sendCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := GetClient()
	client.SetEnabled(true)
	client.SetEndpoint(server.URL)

	// Track multiple events
	client.Track(context.TODO(), "test1", 0, time.Millisecond)
	client.Track(context.TODO(), "test2", 0, time.Millisecond)
	client.Track(context.TODO(), "test3", 0, time.Millisecond)

	// Flush should wait for all sends to complete
	client.Flush(5 * time.Second)

	// All events should have been sent
	count := atomic.LoadInt32(&sendCount)
	if count != 3 {
		t.Errorf("expected 3 sends, got %d", count)
	}
}

func TestClient_Flush_RespectsTimeout(t *testing.T) {
	ResetClient()
	defer ResetClient()

	// Use an unresponsive endpoint instead of a slow server
	// This avoids the httptest server blocking on Close
	client := GetClient()
	client.SetEnabled(true)
	client.SetEndpoint("http://10.255.255.1:1") // Non-routable IP

	client.Track(context.TODO(), "slow", 0, time.Millisecond)

	// Flush with short timeout should return quickly
	start := time.Now()
	client.Flush(100 * time.Millisecond)
	elapsed := time.Since(start)

	// Should return after timeout, not wait for connection
	if elapsed > 500*time.Millisecond {
		t.Errorf("Flush should have returned after timeout, took %v", elapsed)
	}
}

func TestClient_Reload_ThreadSafe(t *testing.T) {
	ResetClient()
	defer ResetClient()
	defer os.Unsetenv(EnvTelemetry)

	client := GetClient()

	// Concurrent reloads should not race
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			if idx%2 == 0 {
				os.Setenv(EnvTelemetry, "1")
			} else {
				os.Setenv(EnvTelemetry, "0")
			}
			client.Reload()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	// Test passes if no race conditions
}
