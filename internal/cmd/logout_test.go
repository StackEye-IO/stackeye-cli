package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

func TestNewLogoutCmd(t *testing.T) {
	cmd := NewLogoutCmd()

	if cmd.Use != "logout" {
		t.Errorf("NewLogoutCmd().Use = %q, want %q", cmd.Use, "logout")
	}

	if cmd.Short == "" {
		t.Error("NewLogoutCmd().Short should not be empty")
	}

	// Verify --all flag exists
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Error("NewLogoutCmd() should have --all flag")
	}
}

func TestLogoutCurrent_WithCredentials(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create a config with credentials
	cfg := config.NewConfig()
	cfg.CurrentContext = "test-context"
	cfg.SetContext("test-context", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationID:   "org_123",
		OrganizationName: "Test Org",
		APIKey:           "se_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	})

	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Set environment to use our test config
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Create config dir structure
	stackeyeDir := filepath.Join(tmpDir, "stackeye")
	if err := os.MkdirAll(stackeyeDir, 0700); err != nil {
		t.Fatalf("Failed to create stackeye dir: %v", err)
	}
	if err := os.Rename(configPath, filepath.Join(stackeyeDir, "config.yaml")); err != nil {
		t.Fatalf("Failed to move config: %v", err)
	}

	// Run logout
	flags := &logoutFlags{all: false}
	err := runLogout(flags)
	if err != nil {
		t.Fatalf("runLogout() error = %v", err)
	}

	// Verify the API key was cleared
	reloadedCfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	ctx, err := reloadedCfg.GetCurrentContext()
	if err != nil {
		t.Fatalf("Failed to get current context: %v", err)
	}

	if ctx.APIKey != "" {
		t.Errorf("API key should be cleared, got %q", ctx.APIKey)
	}

	// Verify other context data is preserved
	if ctx.OrganizationName != "Test Org" {
		t.Errorf("Organization name should be preserved, got %q", ctx.OrganizationName)
	}
}

func TestLogoutCurrent_AlreadyLoggedOut(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()

	// Create a config without credentials
	cfg := config.NewConfig()
	cfg.CurrentContext = "test-context"
	cfg.SetContext("test-context", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Test Org",
		APIKey:           "", // Already logged out
	})

	// Create config dir structure
	stackeyeDir := filepath.Join(tmpDir, "stackeye")
	if err := os.MkdirAll(stackeyeDir, 0700); err != nil {
		t.Fatalf("Failed to create stackeye dir: %v", err)
	}
	configPath := filepath.Join(stackeyeDir, "config.yaml")
	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Set environment to use our test config
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Run logout - should not error
	flags := &logoutFlags{all: false}
	err := runLogout(flags)
	if err != nil {
		t.Fatalf("runLogout() on already logged out context should not error, got %v", err)
	}
}

func TestLogoutCurrent_NoCurrentContext(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()

	// Create a config with no current context
	cfg := config.NewConfig()
	cfg.SetContext("some-context", &config.Context{
		APIURL: "https://api.stackeye.io",
		APIKey: "se_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	})
	// CurrentContext is intentionally empty

	// Create config dir structure
	stackeyeDir := filepath.Join(tmpDir, "stackeye")
	if err := os.MkdirAll(stackeyeDir, 0700); err != nil {
		t.Fatalf("Failed to create stackeye dir: %v", err)
	}
	configPath := filepath.Join(stackeyeDir, "config.yaml")
	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Set environment to use our test config
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Run logout - should not error
	flags := &logoutFlags{all: false}
	err := runLogout(flags)
	if err != nil {
		t.Fatalf("runLogout() with no current context should not error, got %v", err)
	}
}

func TestLogoutAll_MultipleContexts(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()

	// Create a config with multiple contexts with credentials
	cfg := config.NewConfig()
	cfg.CurrentContext = "context-1"
	cfg.SetContext("context-1", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Org 1",
		APIKey:           "se_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	})
	cfg.SetContext("context-2", &config.Context{
		APIURL:           "https://api.dev.stackeye.io",
		OrganizationName: "Org 2",
		APIKey:           "se_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
	})
	cfg.SetContext("context-3", &config.Context{
		APIURL:           "https://api.stg.stackeye.io",
		OrganizationName: "Org 3",
		APIKey:           "", // Already logged out
	})

	// Create config dir structure
	stackeyeDir := filepath.Join(tmpDir, "stackeye")
	if err := os.MkdirAll(stackeyeDir, 0700); err != nil {
		t.Fatalf("Failed to create stackeye dir: %v", err)
	}
	configPath := filepath.Join(stackeyeDir, "config.yaml")
	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Set environment to use our test config
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Run logout --all
	flags := &logoutFlags{all: true}
	err := runLogout(flags)
	if err != nil {
		t.Fatalf("runLogout(--all) error = %v", err)
	}

	// Verify all API keys were cleared
	reloadedCfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	for name, ctx := range reloadedCfg.Contexts {
		if ctx.APIKey != "" {
			t.Errorf("Context %q API key should be cleared, got %q", name, ctx.APIKey)
		}
	}

	// Verify context data is preserved
	ctx1, _ := reloadedCfg.GetContext("context-1")
	if ctx1.OrganizationName != "Org 1" {
		t.Errorf("Context 1 organization name should be preserved, got %q", ctx1.OrganizationName)
	}
}

func TestLogoutAll_NoContexts(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()

	// Create an empty config
	cfg := config.NewConfig()

	// Create config dir structure
	stackeyeDir := filepath.Join(tmpDir, "stackeye")
	if err := os.MkdirAll(stackeyeDir, 0700); err != nil {
		t.Fatalf("Failed to create stackeye dir: %v", err)
	}
	configPath := filepath.Join(stackeyeDir, "config.yaml")
	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Set environment to use our test config
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Run logout --all
	flags := &logoutFlags{all: true}
	err := runLogout(flags)
	if err != nil {
		t.Fatalf("runLogout(--all) with no contexts should not error, got %v", err)
	}
}

func TestLogoutCmd_Integration(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()

	// Create a config with credentials
	cfg := config.NewConfig()
	cfg.CurrentContext = "test"
	cfg.SetContext("test", &config.Context{
		APIURL: "https://api.stackeye.io",
		APIKey: "se_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	})

	// Create config dir structure
	stackeyeDir := filepath.Join(tmpDir, "stackeye")
	if err := os.MkdirAll(stackeyeDir, 0700); err != nil {
		t.Fatalf("Failed to create stackeye dir: %v", err)
	}
	configPath := filepath.Join(stackeyeDir, "config.yaml")
	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Set environment to use our test config
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Execute the command
	cmd := NewLogoutCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("logout command execution error = %v", err)
	}
}

func TestLogoutCmd_HelpText(t *testing.T) {
	cmd := NewLogoutCmd()

	// Verify help text contains important information
	if !strings.Contains(cmd.Long, "logout") {
		t.Error("Long description should mention logout")
	}

	if !strings.Contains(cmd.Long, "--all") {
		t.Error("Long description should mention --all flag")
	}

	if !strings.Contains(cmd.Long, "stackeye login") {
		t.Error("Long description should mention how to log back in")
	}
}
