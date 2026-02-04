package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

// resetGlobalState resets all global flag variables to their default values.
// This should be called at the start of each test to prevent test pollution.
func resetGlobalState() {
	loadedConfig = nil
	configFile = ""
	contextOverride = ""
	debugFlag = false
	verboseFlag = false
	verbosity = 0
	outputFormat = ""
	noColor = false
	noInput = false
	dryRun = false
	timeoutSeconds = 0
	noUpdateCheck = false
	defaultAuthenticator = browserAuthenticator{}
}

func TestLoadConfig_DefaultPath(t *testing.T) {
	resetGlobalState()

	// Load config from default path
	err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig() returned nil after successful loadConfig()")
	}

	// Verify defaults
	if cfg.Preferences == nil {
		t.Fatal("cfg.Preferences is nil")
	}
	if cfg.Preferences.OutputFormat != config.OutputFormatTable {
		t.Errorf("Expected output format 'table', got %q", cfg.Preferences.OutputFormat)
	}
}

func TestLoadConfig_CustomPath(t *testing.T) {
	// Create temp config file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "config.yaml")
	content := `current_context: test
contexts:
  test:
    api_url: https://test.example.com
preferences:
  output_format: json
`
	if err := os.WriteFile(tempFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	resetGlobalState()
	configFile = tempFile

	// Load config
	err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if cfg.CurrentContext != "test" {
		t.Errorf("Expected current_context 'test', got %q", cfg.CurrentContext)
	}
	if cfg.Preferences.OutputFormat != config.OutputFormatJSON {
		t.Errorf("Expected output format 'json', got %q", cfg.Preferences.OutputFormat)
	}

}

func TestLoadConfig_FlagOverrides(t *testing.T) {
	resetGlobalState()
	debugFlag = true
	outputFormat = "yaml"
	noColor = true

	// Load config
	err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig() returned nil")
	}

	// Verify flag overrides
	if !cfg.Preferences.Debug {
		t.Error("Expected Debug=true from --debug flag")
	}
	if cfg.Preferences.OutputFormat != config.OutputFormatYAML {
		t.Errorf("Expected output format 'yaml', got %q", cfg.Preferences.OutputFormat)
	}
	if cfg.Preferences.Color != config.ColorModeNever {
		t.Errorf("Expected color mode 'never', got %q", cfg.Preferences.Color)
	}
}

func TestLoadConfig_InvalidOutputFormat(t *testing.T) {
	resetGlobalState()
	outputFormat = "invalid"

	// Load config should fail
	err := loadConfig()
	if err == nil {
		t.Fatal("Expected error for invalid output format")
	}
}

func TestLoadConfig_ContextOverride(t *testing.T) {
	// Create temp config file with multiple contexts
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "config.yaml")
	content := `current_context: default
contexts:
  default:
    api_url: https://default.example.com
  production:
    api_url: https://prod.example.com
`
	if err := os.WriteFile(tempFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	resetGlobalState()
	configFile = tempFile
	contextOverride = "production"

	// Load config
	err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if cfg.CurrentContext != "production" {
		t.Errorf("Expected current_context 'production', got %q", cfg.CurrentContext)
	}
}

func TestLoadConfig_ContextOverride_Invalid(t *testing.T) {
	resetGlobalState()
	contextOverride = "nonexistent"

	// Load config should fail with invalid context
	err := loadConfig()
	if err == nil {
		t.Fatal("Expected error for nonexistent context")
	}
}

func TestGetNoInput(t *testing.T) {
	resetGlobalState()

	// Default should be false
	if GetNoInput() {
		t.Error("Expected GetNoInput() = false when flag not set")
	}

	// Set to true
	noInput = true
	if !GetNoInput() {
		t.Error("Expected GetNoInput() = true when flag set")
	}
}

func TestGetDryRun(t *testing.T) {
	resetGlobalState()

	// Default should be false
	if GetDryRun() {
		t.Error("Expected GetDryRun() = false when flag not set")
	}

	// Set to true
	dryRun = true
	if !GetDryRun() {
		t.Error("Expected GetDryRun() = true when flag set")
	}
}

func TestGetConfigOrFail_NilConfig(t *testing.T) {
	resetGlobalState()

	// GetConfigOrFail would call os.Exit, so we just test GetConfig
	cfg := GetConfig()
	if cfg != nil {
		t.Error("Expected nil config before loadConfig()")
	}
}

func TestRootCmd(t *testing.T) {
	cmd := RootCmd()
	if cmd == nil {
		t.Fatal("RootCmd() returned nil")
	}
	if cmd.Use != "stackeye" {
		t.Errorf("Expected Use='stackeye', got %q", cmd.Use)
	}
}

func TestLoadConfig_VerboseFlag(t *testing.T) {
	resetGlobalState()
	verboseFlag = true

	err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	if verbosity != 5 {
		t.Errorf("Expected verbosity=5 with --verbose, got %d", verbosity)
	}

	cfg := GetConfig()
	if !cfg.Preferences.Debug {
		t.Error("Expected Debug=true with --verbose flag")
	}
}

func TestLoadConfig_VerboseAndDebug(t *testing.T) {
	resetGlobalState()
	verboseFlag = true
	debugFlag = true

	err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	// --debug (level 6) should take precedence over --verbose (level 5)
	if verbosity != 6 {
		t.Errorf("Expected verbosity=6 with --debug+--verbose, got %d", verbosity)
	}
}

func TestLoadConfig_VerboseWithExplicitV(t *testing.T) {
	resetGlobalState()
	verboseFlag = true
	verbosity = 3

	err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	// Explicit --v=3 should take precedence over --verbose
	if verbosity != 3 {
		t.Errorf("Expected verbosity=3 with explicit --v=3, got %d", verbosity)
	}
}

func TestLoadConfig_DebugWithExplicitV(t *testing.T) {
	resetGlobalState()
	debugFlag = true
	verbosity = 3

	err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	// Explicit --v=3 should take precedence over --debug
	if verbosity != 3 {
		t.Errorf("Expected verbosity=3 with explicit --v=3, got %d", verbosity)
	}
}
