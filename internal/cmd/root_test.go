package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

func TestLoadConfig_DefaultPath(t *testing.T) {
	// Reset global state
	loadedConfig = nil
	configFile = ""
	debugFlag = false
	outputFormat = ""
	noColor = false

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

	// Reset global state
	loadedConfig = nil
	configFile = tempFile
	debugFlag = false
	outputFormat = ""
	noColor = false

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

	// Cleanup
	configFile = ""
}

func TestLoadConfig_FlagOverrides(t *testing.T) {
	// Reset global state
	loadedConfig = nil
	configFile = ""
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

	// Cleanup
	debugFlag = false
	outputFormat = ""
	noColor = false
}

func TestLoadConfig_InvalidOutputFormat(t *testing.T) {
	// Reset global state
	loadedConfig = nil
	configFile = ""
	debugFlag = false
	outputFormat = "invalid"
	noColor = false

	// Load config should fail
	err := loadConfig()
	if err == nil {
		t.Fatal("Expected error for invalid output format")
	}

	// Cleanup
	outputFormat = ""
}

func TestGetConfigOrFail_NilConfig(t *testing.T) {
	// Reset global state
	loadedConfig = nil

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
