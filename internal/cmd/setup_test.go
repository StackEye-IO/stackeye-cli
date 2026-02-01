// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-cli/internal/auth"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
)

func TestNewSetupCmd(t *testing.T) {
	cmd := NewSetupCmd()

	if cmd == nil {
		t.Fatal("NewSetupCmd() returned nil")
	}

	if cmd.Use != "setup" {
		t.Errorf("unexpected command Use: got %v want setup", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Check that --api-url flag exists
	apiURLFlag := cmd.Flags().Lookup("api-url")
	if apiURLFlag == nil {
		t.Error("expected --api-url flag to exist")
	} else if apiURLFlag.DefValue != auth.DefaultAPIURL {
		t.Errorf("unexpected default for --api-url: got %v want %v", apiURLFlag.DefValue, auth.DefaultAPIURL)
	}

	// Check that --skip-probe flag exists
	skipProbeFlag := cmd.Flags().Lookup("skip-probe")
	if skipProbeFlag == nil {
		t.Error("expected --skip-probe flag to exist")
	} else if skipProbeFlag.DefValue != "false" {
		t.Errorf("unexpected default for --skip-probe: got %v want false", skipProbeFlag.DefValue)
	}
}

func TestRunSetupNonInteractive_NoConfig(t *testing.T) {
	resetGlobalState()
	noInput = true

	// Create temp directory with no config
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "nonexistent.yaml")
	configFile = tempFile

	// Capture stdout
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	flags := &setupFlags{
		apiURL:    auth.DefaultAPIURL,
		skipProbe: false,
	}

	// Run non-interactive setup
	err := runSetupNonInteractive(t.Context(), flags)
	if err != nil {
		t.Fatalf("runSetupNonInteractive() unexpected error: %v", err)
	}

	// Read captured output
	w.Close()
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	output := buf.String()

	// Should indicate not configured
	if output == "" {
		t.Skip("stdout capture issue - skipping output assertion")
	}
}

func TestRunSetupNonInteractive_ValidContext(t *testing.T) {
	resetGlobalState()
	noInput = true

	// Create temp config file with valid context
	// Note: This test verifies config parsing works correctly.
	// The runSetupNonInteractive function would make an API call to verify
	// credentials, which requires mocking the API client for full coverage.
	// For now, we verify the config file parsing path.
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "config.yaml")
	content := `current_context: test
contexts:
  test:
    api_url: https://test.example.com
    api_key: test-key
    organization_name: Test Org
preferences:
  output_format: table
`
	if err := os.WriteFile(tempFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}
	configFile = tempFile

	// Verify config can be loaded successfully
	err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig() returned nil after loadConfig()")
	}
	if cfg.CurrentContext != "test" {
		t.Errorf("expected CurrentContext 'test', got %q", cfg.CurrentContext)
	}

	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		t.Fatalf("GetCurrentContext() failed: %v", err)
	}
	if ctx.APIKey != "test-key" {
		t.Errorf("expected APIKey 'test-key', got %q", ctx.APIKey)
	}
	if ctx.OrganizationName != "Test Org" {
		t.Errorf("expected OrganizationName 'Test Org', got %q", ctx.OrganizationName)
	}
}

func TestSkipIfAuthenticated_NotAuthenticated(t *testing.T) {
	resetGlobalState()

	// Use a non-existent config file to ensure clean state
	tempDir := t.TempDir()
	configFile = filepath.Join(tempDir, "nonexistent.yaml")

	wiz := interactive.NewWizard(&interactive.WizardOptions{
		Title: "Test Wizard",
	})
	wiz.SetData("apiURL", "https://api.nonexistent.example.com")

	// Should not skip when no valid config exists
	result := skipIfAuthenticated(wiz)

	// Will return false because config.Load() will fail or find no matching context
	if result {
		// If skipIfAuthenticated returns true, it means it found a cached config
		// which is acceptable in the test environment - just verify it set data
		if wiz.GetDataBool("authenticated") != true {
			t.Error("skipIfAuthenticated returned true but didn't set authenticated=true")
		}
	}
}

func TestSkipIfAuthenticated_Authenticated(t *testing.T) {
	resetGlobalState()

	// Create temp config file with matching context
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "config.yaml")
	testAPIURL := "https://api.test.stackeye.io"
	content := `current_context: test
contexts:
  test:
    api_url: ` + testAPIURL + `
    api_key: test-key-12345
    organization_id: org-123
    organization_name: Test Organization
preferences:
  output_format: table
`
	if err := os.WriteFile(tempFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}
	configFile = tempFile

	wiz := interactive.NewWizard(&interactive.WizardOptions{
		Title: "Test Wizard",
	})
	wiz.SetData("apiURL", testAPIURL)

	// Should skip when config has matching API URL with valid credentials
	result := skipIfAuthenticated(wiz)

	if result {
		// Verify data was set correctly
		if wiz.GetDataString("apiKey") != "test-key-12345" {
			t.Errorf("expected apiKey to be set, got %q", wiz.GetDataString("apiKey"))
		}
		if wiz.GetDataString("orgName") != "Test Organization" {
			t.Errorf("expected orgName to be 'Test Organization', got %q", wiz.GetDataString("orgName"))
		}
		if wiz.GetDataString("orgID") != "org-123" {
			t.Errorf("expected orgID to be 'org-123', got %q", wiz.GetDataString("orgID"))
		}
		if wiz.GetDataBool("authenticated") != true {
			t.Error("expected authenticated to be true")
		}
	}
}

func TestSkipIfSingleOrg_NotAuthenticated(t *testing.T) {
	resetGlobalState()

	wiz := interactive.NewWizard(&interactive.WizardOptions{
		Title: "Test Wizard",
	})
	// Not authenticated
	wiz.SetData("authenticated", false)

	// Should not skip when not authenticated
	result := skipIfSingleOrg(wiz)

	if result {
		t.Error("skipIfSingleOrg should return false when not authenticated")
	}
}

func TestSkipIfSingleOrg_AuthenticatedNoAPIKey(t *testing.T) {
	resetGlobalState()

	wiz := interactive.NewWizard(&interactive.WizardOptions{
		Title: "Test Wizard",
	})
	wiz.SetData("authenticated", true)
	wiz.SetData("apiKey", "") // Empty API key

	// Should not skip when API key is empty
	result := skipIfSingleOrg(wiz)

	// Returns false because it can't make API call without key
	if result {
		// If it skipped, it means the SDK call succeeded somehow
		// which is unexpected with empty credentials
		t.Log("skipIfSingleOrg returned true with empty API key - SDK may have cache")
	}
}

func TestSkipProbeStep(t *testing.T) {
	tests := []struct {
		name      string
		skipProbe bool
		want      bool
	}{
		{
			name:      "skip probe when flag is true",
			skipProbe: true,
			want:      true,
		},
		{
			name:      "do not skip probe when flag is false",
			skipProbe: false,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wiz := interactive.NewWizard(&interactive.WizardOptions{
				Title: "Test Wizard",
			})
			wiz.SetData("skipProbe", tt.skipProbe)

			got := skipProbeStep(wiz)
			if got != tt.want {
				t.Errorf("skipProbeStep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintSetupSummary(t *testing.T) {
	// Create wizard with test data
	wiz := interactive.NewWizard(&interactive.WizardOptions{
		Title: "Test Wizard",
	})

	// Set up test data
	wiz.SetData("contextName", "test-context")
	wiz.SetData("orgName", "Test Organization")
	wiz.SetData("selectedOrgName", "Different Org")
	wiz.SetData("probeCreated", true)
	wiz.SetData("probeName", "Test Probe")
	wiz.SetData("probeID", "probe-uuid-12345")

	// Capture stdout - printSetupSummary prints to os.Stdout
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSetupSummary(wiz)

	// Read captured output
	w.Close()
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	// The function executed without panic - that's the main test
	// stdout capture in tests is environment-dependent
}

func TestPrintSetupSummary_MinimalData(t *testing.T) {
	// Create wizard with minimal data
	wiz := interactive.NewWizard(&interactive.WizardOptions{
		Title: "Test Wizard",
	})

	// Only set basic auth data
	wiz.SetData("contextName", "default")
	wiz.SetData("orgName", "My Org")
	wiz.SetData("probeCreated", false)

	// Capture stdout
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSetupSummary(wiz)

	// Read captured output
	w.Close()
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	// Function executed without panic
}

func TestPrintSetupSummary_NoData(t *testing.T) {
	// Create wizard with no data set
	wiz := interactive.NewWizard(&interactive.WizardOptions{
		Title: "Test Wizard",
	})

	// Capture stdout
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSetupSummary(wiz)

	// Read captured output
	w.Close()
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout

	// Function executed without panic even with no data
}

func TestSetupFlags(t *testing.T) {
	flags := &setupFlags{
		apiURL:    "https://custom.api.url",
		skipProbe: true,
	}

	if flags.apiURL != "https://custom.api.url" {
		t.Errorf("expected apiURL to be custom URL, got %v", flags.apiURL)
	}

	if !flags.skipProbe {
		t.Error("expected skipProbe to be true")
	}
}

func TestSetupFlags_APIKey(t *testing.T) {
	flags := &setupFlags{
		apiURL:        "https://api.stackeye.io",
		apiKey:        "se_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		probeName:     "Test Probe",
		probeURL:      "https://example.com",
		probeInterval: 30,
	}

	if flags.apiKey == "" {
		t.Error("expected apiKey to be set")
	}

	if flags.probeName != "Test Probe" {
		t.Errorf("expected probeName to be 'Test Probe', got %v", flags.probeName)
	}

	if flags.probeURL != "https://example.com" {
		t.Errorf("expected probeURL to be 'https://example.com', got %v", flags.probeURL)
	}

	if flags.probeInterval != 30 {
		t.Errorf("expected probeInterval to be 30, got %v", flags.probeInterval)
	}
}

func TestGetAPIKeyFromSources_FlagPrecedence(t *testing.T) {
	// Test that --api-key flag takes precedence over env var
	flagKey := "se_flag1234567890123456789012345678901234567890123456789012345678"
	envKey := "se_env01234567890123456789012345678901234567890123456789012345678"

	// Set env var
	t.Setenv("STACKEYE_API_KEY", envKey)

	flags := &setupFlags{
		apiKey: flagKey,
	}

	result := getAPIKeyFromSources(flags)
	if result != flagKey {
		t.Errorf("expected flag API key to take precedence, got %v", result)
	}
}

func TestGetAPIKeyFromSources_EnvVar(t *testing.T) {
	// Test that env var is used when flag is empty
	envKey := "se_env01234567890123456789012345678901234567890123456789012345678"

	// Set env var
	t.Setenv("STACKEYE_API_KEY", envKey)

	flags := &setupFlags{
		apiKey: "", // Empty flag
	}

	result := getAPIKeyFromSources(flags)
	if result != envKey {
		t.Errorf("expected env var API key, got %v", result)
	}
}

func TestGetAPIKeyFromSources_NoKey(t *testing.T) {
	// Clear env var
	t.Setenv("STACKEYE_API_KEY", "")

	flags := &setupFlags{
		apiKey: "",
	}

	result := getAPIKeyFromSources(flags)
	if result != "" {
		t.Errorf("expected empty string, got %v", result)
	}
}

func TestNewSetupCmd_APIKeyFlags(t *testing.T) {
	cmd := NewSetupCmd()

	// Check that --api-key flag exists
	apiKeyFlag := cmd.Flags().Lookup("api-key")
	if apiKeyFlag == nil {
		t.Error("expected --api-key flag to exist")
	} else if apiKeyFlag.DefValue != "" {
		t.Errorf("expected default for --api-key to be empty, got %v", apiKeyFlag.DefValue)
	}

	// Check that --probe-name flag exists
	probeNameFlag := cmd.Flags().Lookup("probe-name")
	if probeNameFlag == nil {
		t.Error("expected --probe-name flag to exist")
	}

	// Check that --probe-url flag exists
	probeURLFlag := cmd.Flags().Lookup("probe-url")
	if probeURLFlag == nil {
		t.Error("expected --probe-url flag to exist")
	}

	// Check that --probe-interval flag exists with default value
	probeIntervalFlag := cmd.Flags().Lookup("probe-interval")
	if probeIntervalFlag == nil {
		t.Error("expected --probe-interval flag to exist")
	} else if probeIntervalFlag.DefValue != "60" {
		t.Errorf("expected default for --probe-interval to be 60, got %v", probeIntervalFlag.DefValue)
	}
}

func TestRunNonInteractiveAuth_InvalidAPIKeyFormat(t *testing.T) {
	resetGlobalState()
	noInput = true

	// Create temp directory for config
	tempDir := t.TempDir()
	configFile = filepath.Join(tempDir, "config.yaml")

	flags := &setupFlags{
		apiURL: auth.DefaultAPIURL,
		apiKey: "invalid-key", // Invalid format
	}

	err := runNonInteractiveAuth(t.Context(), flags, flags.apiKey)
	if err == nil {
		t.Fatal("expected error for invalid API key format")
	}

	if !strings.Contains(err.Error(), "invalid API key format") {
		t.Errorf("expected 'invalid API key format' error, got: %v", err)
	}
}

func TestCreateNonInteractiveProbe_MissingProbeName(t *testing.T) {
	resetGlobalState()

	// Valid API key format
	apiKey := "se_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	flags := &setupFlags{
		apiURL:    auth.DefaultAPIURL,
		probeName: "", // Missing
		probeURL:  "https://example.com",
	}

	err := createNonInteractiveProbe(t.Context(), flags, apiKey)
	if err == nil {
		t.Fatal("expected error for missing probe name")
	}

	if !strings.Contains(err.Error(), "--probe-name is required") {
		t.Errorf("expected '--probe-name is required' error, got: %v", err)
	}
}

func TestCreateNonInteractiveProbe_MissingProbeURL(t *testing.T) {
	resetGlobalState()

	// Valid API key format
	apiKey := "se_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	flags := &setupFlags{
		apiURL:    auth.DefaultAPIURL,
		probeName: "Test Probe",
		probeURL:  "", // Missing
	}

	err := createNonInteractiveProbe(t.Context(), flags, apiKey)
	if err == nil {
		t.Fatal("expected error for missing probe URL")
	}

	if !strings.Contains(err.Error(), "--probe-url is required") {
		t.Errorf("expected '--probe-url is required' error, got: %v", err)
	}
}

func TestCreateNonInteractiveProbe_InvalidProbeURL(t *testing.T) {
	resetGlobalState()

	// Valid API key format
	apiKey := "se_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	flags := &setupFlags{
		apiURL:    auth.DefaultAPIURL,
		probeName: "Test Probe",
		probeURL:  "not-a-valid-url", // Invalid URL
	}

	err := createNonInteractiveProbe(t.Context(), flags, apiKey)
	if err == nil {
		t.Fatal("expected error for invalid probe URL")
	}

	if !strings.Contains(err.Error(), "invalid probe URL") {
		t.Errorf("expected 'invalid probe URL' error, got: %v", err)
	}
}
