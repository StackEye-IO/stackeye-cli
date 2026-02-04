package testutil

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/stretchr/testify/require"
)

// TestEnv encapsulates all test dependencies.
type TestEnv struct {
	// T is the testing context.
	T *testing.T
	// Server is the mock API server.
	Server *MockServer
	// Config is the test configuration.
	Config *TestConfig
	// CLI is the command runner.
	CLI *CLIRunner
}

// NewTestEnv creates a complete test environment with mock server, config, and CLI runner.
// Call Cleanup() when done to release resources.
func NewTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	// Create mock server
	server := NewMockServer()
	server.RegisterDefaultRoutes()

	// Create test config pointing to mock server
	config, err := NewTestConfig(server.BaseURL)
	require.NoError(t, err, "failed to create test config")

	// Create CLI runner with XDG_CONFIG_HOME set to test directory
	// This ensures Save() writes to the test config, not user's actual config
	cli := NewCLIRunner(config.ConfigPath, config.Dir)

	return &TestEnv{
		T:      t,
		Server: server,
		Config: config,
		CLI:    cli,
	}
}

// NewTestEnvWithMultipleContexts creates a test environment with multiple contexts.
func NewTestEnvWithMultipleContexts(t *testing.T) *TestEnv {
	t.Helper()

	// Create mock server
	server := NewMockServer()
	server.RegisterDefaultRoutes()

	// Create test config with multiple contexts
	config, err := NewTestConfigWithMultipleContexts(server.BaseURL)
	require.NoError(t, err, "failed to create test config")

	// Create CLI runner with XDG_CONFIG_HOME set to test directory
	// This ensures Save() writes to the test config, not user's actual config
	cli := NewCLIRunner(config.ConfigPath, config.Dir)

	return &TestEnv{
		T:      t,
		Server: server,
		Config: config,
		CLI:    cli,
	}
}

// NewLiveTestEnv creates a test environment using the user's real CLI config.
// This connects to the live API instead of a mock server.
// Requires: User must be authenticated (run `stackeye login` first).
func NewLiveTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	if !IsLiveTestEnabled() {
		t.Skip("skipping live integration test: set STACKEYE_E2E_LIVE=true to run")
	}

	// Get the user's actual config path
	configPath := config.ConfigPath()
	require.NotEmpty(t, configPath, "failed to determine config path - home directory unavailable")

	// Load existing config to verify it exists and is configured
	cfg, err := config.Load()
	require.NoError(t, err, "failed to load config - run 'stackeye login' first")
	require.NotEmpty(t, cfg.CurrentContext, "no active context - run 'stackeye login' first")

	// Verify the current context has an API key
	ctx, err := cfg.GetContext(cfg.CurrentContext)
	require.NoError(t, err, "current context %q not found in config", cfg.CurrentContext)
	require.NotEmpty(t, ctx.APIKey, "no API key in context %q - run 'stackeye login' first", cfg.CurrentContext)

	// Get config directory (parent of config.yaml) for XDG_CONFIG_HOME
	configDir := filepath.Dir(configPath)
	xdgConfigHome := filepath.Dir(configDir) // Go up one more level since config is at $XDG_CONFIG_HOME/stackeye/config.yaml

	// Create CLI runner using real config
	cli := NewCLIRunner(configPath, xdgConfigHome)

	return &TestEnv{
		T:      t,
		Server: nil, // No mock server for live tests
		Config: nil, // Using user's real config, not managed by TestConfig
		CLI:    cli,
	}
}

// IsLiveMode returns true if running against live API (Server is nil).
func (e *TestEnv) IsLiveMode() bool {
	return e.Server == nil
}

// Cleanup releases all test resources.
func (e *TestEnv) Cleanup() {
	if e.Server != nil {
		e.Server.Close()
	}
	if e.Config != nil {
		e.Config.Cleanup()
	}
	// No cleanup needed for live mode - we don't modify user's config
}

// Run executes a CLI command and returns the result.
func (e *TestEnv) Run(args ...string) *RunResult {
	return e.CLI.Run(args...)
}

// RunSuccess executes and asserts success.
func (e *TestEnv) RunSuccess(args ...string) *RunResult {
	e.T.Helper()
	result := e.CLI.Run(args...)
	require.True(e.T, result.Success(), "expected success, got: %s", result)
	return result
}

// RunError executes and asserts failure.
func (e *TestEnv) RunError(args ...string) *RunResult {
	e.T.Helper()
	result := e.CLI.Run(args...)
	require.True(e.T, result.Failed(), "expected failure, got: %s", result)
	return result
}

// AssertContains asserts that stdout contains the substring.
func (e *TestEnv) AssertContains(result *RunResult, substr string) {
	e.T.Helper()
	require.Contains(e.T, result.Stdout, substr, "expected stdout to contain %q", substr)
}

// AssertNotContains asserts that stdout does not contain the substring.
func (e *TestEnv) AssertNotContains(result *RunResult, substr string) {
	e.T.Helper()
	require.NotContains(e.T, result.Stdout, substr, "expected stdout to not contain %q", substr)
}

// AssertStderrContains asserts that stderr contains the substring.
func (e *TestEnv) AssertStderrContains(result *RunResult, substr string) {
	e.T.Helper()
	require.Contains(e.T, result.Stderr, substr, "expected stderr to contain %q", substr)
}

// AssertAPICall asserts that a specific API call was made.
// In live mode, this is a no-op (we can't inspect real API calls).
func (e *TestEnv) AssertAPICall(method, pathPrefix string) {
	e.T.Helper()
	if e.IsLiveMode() {
		// Skip in live mode - can't intercept real API calls
		return
	}
	require.True(e.T, e.Server.HasCall(method, pathPrefix),
		"expected API call %s %s, got: %v", method, pathPrefix, e.Server.GetCalls())
}

// AssertNoAPICall asserts that no API call matching the pattern was made.
// In live mode, this is a no-op (we can't inspect real API calls).
func (e *TestEnv) AssertNoAPICall(method, pathPrefix string) {
	e.T.Helper()
	if e.IsLiveMode() {
		// Skip in live mode - can't intercept real API calls
		return
	}
	require.False(e.T, e.Server.HasCall(method, pathPrefix),
		"expected no API call %s %s, but found it", method, pathPrefix)
}

// AssertJSONOutput asserts that stdout is valid JSON and unmarshals it.
func (e *TestEnv) AssertJSONOutput(result *RunResult, v any) {
	e.T.Helper()
	err := json.Unmarshal([]byte(result.Stdout), v)
	require.NoError(e.T, err, "expected valid JSON output: %s", result.Stdout)
}

// AssertTableHasRows asserts that table output has at least the expected number of rows.
// This accounts for header row(s).
func (e *TestEnv) AssertTableHasRows(result *RunResult, minRows int) {
	e.T.Helper()
	lines := result.Lines()
	// Filter out empty lines and divider lines
	dataLines := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "---") && !strings.HasPrefix(trimmed, "===") {
			dataLines = append(dataLines, line)
		}
	}
	require.GreaterOrEqual(e.T, len(dataLines), minRows+1,
		"expected at least %d data rows plus header, got %d lines", minRows, len(dataLines))
}

// ClearAPICalls clears the recorded API calls.
// In live mode, this is a no-op.
func (e *TestEnv) ClearAPICalls() {
	if e.IsLiveMode() {
		return
	}
	e.Server.ClearCalls()
}

// GetAPICalls returns all recorded API calls.
// In live mode, returns an empty slice (we can't intercept real API calls).
func (e *TestEnv) GetAPICalls() []RecordedCall {
	if e.IsLiveMode() {
		return nil
	}
	return e.Server.GetCalls()
}
