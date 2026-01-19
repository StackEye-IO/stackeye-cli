package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

func TestNewContextCmd(t *testing.T) {
	cmd := NewContextCmd()

	if cmd.Use != "context" {
		t.Errorf("expected Use='context', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify list subcommand is registered
	listCmd, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Errorf("list subcommand not found: %v", err)
	}
	if listCmd.Use != "list" {
		t.Errorf("expected list Use='list', got %q", listCmd.Use)
	}
}

func TestNewContextListCmd_Alias(t *testing.T) {
	cmd := NewContextCmd()
	listCmd, _, err := cmd.Find([]string{"ls"})
	if err != nil {
		t.Errorf("ls alias not found: %v", err)
	}
	if listCmd.Use != "list" {
		t.Errorf("expected alias to resolve to list, got %q", listCmd.Use)
	}
}

func TestRunContextList_NoContexts(t *testing.T) {
	// Create empty config
	cfg := config.NewConfig()
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextList()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No contexts configured") {
		t.Errorf("expected output to contain 'No contexts configured', got %q", output)
	}
}

func TestRunContextList_WithContexts(t *testing.T) {
	// Create config with contexts
	cfg := config.NewConfig()
	cfg.CurrentContext = "prod"
	cfg.SetContext("prod", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Acme Corp",
		OrganizationID:   "org_123",
		APIKey:           "se_test123",
	})
	cfg.SetContext("dev", &config.Context{
		APIURL:           "https://api.dev.stackeye.io",
		OrganizationName: "Acme Dev",
		OrganizationID:   "org_456",
		APIKey:           "se_dev456",
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextList()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check header
	if !strings.Contains(output, "NAME") {
		t.Error("expected output to contain NAME header")
	}
	if !strings.Contains(output, "ORGANIZATION") {
		t.Error("expected output to contain ORGANIZATION header")
	}
	if !strings.Contains(output, "API URL") {
		t.Error("expected output to contain API URL header")
	}

	// Check contexts are listed (sorted alphabetically)
	if !strings.Contains(output, "dev") {
		t.Error("expected output to contain dev context")
	}
	if !strings.Contains(output, "prod") {
		t.Error("expected output to contain prod context")
	}

	// Check organization names
	if !strings.Contains(output, "Acme Corp") {
		t.Error("expected output to contain Acme Corp")
	}
	if !strings.Contains(output, "Acme Dev") {
		t.Error("expected output to contain Acme Dev")
	}

	// Check current context marker - prod should have asterisk
	if !strings.Contains(output, "*") {
		t.Error("expected output to contain asterisk for current context")
	}
}

func TestRunContextList_NoOrgName(t *testing.T) {
	// Create config with context without org name
	cfg := config.NewConfig()
	cfg.CurrentContext = "default"
	cfg.SetContext("default", &config.Context{
		APIURL: "https://api.stackeye.io",
		APIKey: "se_test123",
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextList()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check that (not set) is displayed for missing org name
	if !strings.Contains(output, "(not set)") {
		t.Error("expected output to contain '(not set)' for missing org name")
	}
}

func TestRunContextList_NilConfig(t *testing.T) {
	loadedConfig = nil

	err := runContextList()

	if err == nil {
		t.Error("expected error for nil config, got nil")
	}
	if !strings.Contains(err.Error(), "configuration not loaded") {
		t.Errorf("expected error to contain 'configuration not loaded', got %q", err.Error())
	}
}

func TestRunContextList_DefaultAPIURL(t *testing.T) {
	// Create config with context using default API URL
	cfg := config.NewConfig()
	cfg.CurrentContext = "default"
	cfg.SetContext("default", &config.Context{
		// APIURL intentionally left empty to test default
		OrganizationName: "Test Org",
		APIKey:           "se_test123",
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextList()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check that default API URL is shown
	if !strings.Contains(output, config.DefaultAPIURL) {
		t.Errorf("expected output to contain default API URL %q", config.DefaultAPIURL)
	}
}

func TestRunContextList_NilContextValue(t *testing.T) {
	// Create config with a nil context value (edge case from YAML parsing)
	cfg := config.NewConfig()
	cfg.CurrentContext = "valid"
	cfg.SetContext("valid", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Valid Org",
		APIKey:           "se_valid123",
	})
	// Simulate nil context value (can happen with empty YAML entry)
	cfg.Contexts["nilctx"] = nil
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextList()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should list valid context without panicking
	if !strings.Contains(output, "valid") {
		t.Error("expected output to contain valid context")
	}
	if !strings.Contains(output, "Valid Org") {
		t.Error("expected output to contain Valid Org")
	}
}

func TestRunContextList_LongNames(t *testing.T) {
	// Test that long names are truncated properly
	cfg := config.NewConfig()
	cfg.CurrentContext = "very-long-context-name-exceeds-limit"
	cfg.SetContext("very-long-context-name-exceeds-limit", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Super Long Organization Name That Exceeds Column Width",
		APIKey:           "se_test123",
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextList()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should contain truncated name with ellipsis
	if !strings.Contains(output, "...") {
		t.Error("expected truncated names to contain ellipsis")
	}
}
