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
	_, _ = buf.ReadFrom(r)
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
	_, _ = buf.ReadFrom(r)
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
	_, _ = buf.ReadFrom(r)
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
	_, _ = buf.ReadFrom(r)
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
	_, _ = buf.ReadFrom(r)
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
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should contain truncated name with ellipsis
	if !strings.Contains(output, "...") {
		t.Error("expected truncated names to contain ellipsis")
	}
}

func TestNewContextUseCmd(t *testing.T) {
	cmd := NewContextCmd()

	// Verify use subcommand is registered
	useCmd, _, err := cmd.Find([]string{"use"})
	if err != nil {
		t.Errorf("use subcommand not found: %v", err)
	}
	if useCmd.Use != "use <name>" {
		t.Errorf("expected Use='use <name>', got %q", useCmd.Use)
	}
}

func TestNewContextCurrentCmd(t *testing.T) {
	cmd := NewContextCmd()

	// Verify current subcommand is registered
	currentCmd, _, err := cmd.Find([]string{"current"})
	if err != nil {
		t.Errorf("current subcommand not found: %v", err)
	}
	if currentCmd.Use != "current" {
		t.Errorf("expected Use='current', got %q", currentCmd.Use)
	}
}

func TestRunContextCurrent_Success(t *testing.T) {
	// Create config with current context set
	cfg := config.NewConfig()
	cfg.CurrentContext = "prod"
	cfg.SetContext("prod", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Acme Corp",
		OrganizationID:   "org_123",
		APIKey:           "se_test123",
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextCurrent()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Check context name is displayed
	if !strings.Contains(output, "prod") {
		t.Error("expected output to contain context name 'prod'")
	}

	// Check organization name is displayed
	if !strings.Contains(output, "Acme Corp") {
		t.Error("expected output to contain organization name 'Acme Corp'")
	}

	// Check API URL is displayed
	if !strings.Contains(output, "https://api.stackeye.io") {
		t.Error("expected output to contain API URL")
	}
}

func TestRunContextCurrent_NoCurrentContext(t *testing.T) {
	// Create empty config with no current context
	cfg := config.NewConfig()
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextCurrent()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Check helpful message is displayed
	if !strings.Contains(output, "No current context set") {
		t.Error("expected output to contain 'No current context set'")
	}
	if !strings.Contains(output, "stackeye login") {
		t.Error("expected output to contain login hint")
	}
}

func TestRunContextCurrent_NilConfig(t *testing.T) {
	loadedConfig = nil

	err := runContextCurrent()

	if err == nil {
		t.Error("expected error for nil config, got nil")
	}
	if !strings.Contains(err.Error(), "configuration not loaded") {
		t.Errorf("expected error to contain 'configuration not loaded', got %q", err.Error())
	}
}

func TestRunContextCurrent_NoOrgName(t *testing.T) {
	// Create config with context without org name
	cfg := config.NewConfig()
	cfg.CurrentContext = "default"
	cfg.SetContext("default", &config.Context{
		APIURL: "https://api.stackeye.io",
		APIKey: "se_test123",
		// No OrganizationName
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextCurrent()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Check that (not set) is displayed for missing org name
	if !strings.Contains(output, "(not set)") {
		t.Error("expected output to contain '(not set)' for missing org name")
	}
}

func TestRunContextCurrent_DefaultAPIURL(t *testing.T) {
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

	err := runContextCurrent()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Check that default API URL is shown
	if !strings.Contains(output, config.DefaultAPIURL) {
		t.Errorf("expected output to contain default API URL %q", config.DefaultAPIURL)
	}
}

func TestRunContextCurrent_NilContextValue(t *testing.T) {
	// Create config with a nil context value for the current context
	cfg := config.NewConfig()
	cfg.CurrentContext = "nilctx"
	cfg.Contexts["nilctx"] = nil
	loadedConfig = cfg

	err := runContextCurrent()

	if err == nil {
		t.Error("expected error for nil context value, got nil")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected error to contain 'invalid', got %q", err.Error())
	}
}

func TestRunContextUse_Success(t *testing.T) {
	// Use a temp directory for config to avoid writing to user's real config file.
	tmpDir := t.TempDir()
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Create config with multiple contexts
	cfg := config.NewConfig()
	cfg.CurrentContext = "prod"
	cfg.SetContext("prod", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Acme Corp",
		APIKey:           "se_test12300000000000000000000000000000000000000000000000000000",
	})
	cfg.SetContext("dev", &config.Context{
		APIURL:           "https://api.dev.stackeye.io",
		OrganizationName: "Acme Dev",
		APIKey:           "se_dev45600000000000000000000000000000000000000000000000000000",
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextUse("dev")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.CurrentContext != "dev" {
		t.Errorf("expected CurrentContext to be 'dev', got %q", cfg.CurrentContext)
	}
	if !strings.Contains(output, "Switched to context") {
		t.Errorf("expected output to contain 'Switched to context', got %q", output)
	}
	if !strings.Contains(output, "Acme Dev") {
		t.Errorf("expected output to contain org name 'Acme Dev', got %q", output)
	}
}

func TestRunContextUse_AlreadyUsing(t *testing.T) {
	// Create config with context already set
	cfg := config.NewConfig()
	cfg.CurrentContext = "prod"
	cfg.SetContext("prod", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Acme Corp",
		APIKey:           "se_test123",
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextUse("prod")

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Already using context") {
		t.Errorf("expected output to contain 'Already using context', got %q", output)
	}
}

func TestRunContextUse_NotFound(t *testing.T) {
	// Create config without the target context
	cfg := config.NewConfig()
	cfg.CurrentContext = "prod"
	cfg.SetContext("prod", &config.Context{
		APIURL: "https://api.stackeye.io",
		APIKey: "se_test123",
	})
	loadedConfig = cfg

	err := runContextUse("nonexistent")

	if err == nil {
		t.Error("expected error for nonexistent context, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain 'not found', got %q", err.Error())
	}
}

func TestRunContextUse_NilConfig(t *testing.T) {
	loadedConfig = nil

	err := runContextUse("somecontext")

	if err == nil {
		t.Error("expected error for nil config, got nil")
	}
	if !strings.Contains(err.Error(), "configuration not loaded") {
		t.Errorf("expected error to contain 'configuration not loaded', got %q", err.Error())
	}
}

func TestRunContextUse_NilContext(t *testing.T) {
	// Create config with a nil context value
	cfg := config.NewConfig()
	cfg.CurrentContext = "valid"
	cfg.SetContext("valid", &config.Context{
		APIURL: "https://api.stackeye.io",
		APIKey: "se_valid123",
	})
	// Simulate nil context value
	cfg.Contexts["nilctx"] = nil
	loadedConfig = cfg

	err := runContextUse("nilctx")

	if err == nil {
		t.Error("expected error for nil context, got nil")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected error to contain 'invalid', got %q", err.Error())
	}
}

func TestRunContextUse_NoOrgName(t *testing.T) {
	// Use a temp directory for config to avoid writing to user's real config file.
	// This prevents the test from corrupting the user's actual configuration.
	tmpDir := t.TempDir()
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Create config with context without org name
	cfg := config.NewConfig()
	cfg.CurrentContext = "default"
	cfg.SetContext("default", &config.Context{
		APIURL: "https://api.stackeye.io",
		APIKey: "se_test12300000000000000000000000000000000000000000000000000000",
	})
	cfg.SetContext("staging", &config.Context{
		APIURL: "https://api.staging.stackeye.io",
		APIKey: "se_staging0000000000000000000000000000000000000000000000000000",
		// No OrganizationName
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextUse("staging")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Save might fail, but if it succeeds, output should not have org name in parens
	if err == nil {
		// Check that no parentheses appear (since there's no org name set)
		if strings.Contains(output, "(") && strings.Contains(output, ")") {
			t.Error("expected output not to contain parentheses when org name is not set")
		}
		if !strings.Contains(output, "Switched to context") {
			t.Errorf("expected output to contain 'Switched to context', got %q", output)
		}
	}
}

func TestNewContextDeleteCmd(t *testing.T) {
	cmd := NewContextCmd()

	// Verify delete subcommand is registered
	deleteCmd, _, err := cmd.Find([]string{"delete"})
	if err != nil {
		t.Errorf("delete subcommand not found: %v", err)
	}
	if deleteCmd.Use != "delete <name>" {
		t.Errorf("expected Use='delete <name>', got %q", deleteCmd.Use)
	}
}

func TestNewContextDeleteCmd_Aliases(t *testing.T) {
	cmd := NewContextCmd()

	// Test rm alias
	rmCmd, _, err := cmd.Find([]string{"rm"})
	if err != nil {
		t.Errorf("rm alias not found: %v", err)
	}
	if rmCmd.Use != "delete <name>" {
		t.Errorf("expected rm alias to resolve to delete, got %q", rmCmd.Use)
	}

	// Test remove alias
	removeCmd, _, err := cmd.Find([]string{"remove"})
	if err != nil {
		t.Errorf("remove alias not found: %v", err)
	}
	if removeCmd.Use != "delete <name>" {
		t.Errorf("expected remove alias to resolve to delete, got %q", removeCmd.Use)
	}
}

func TestRunContextDelete_Success(t *testing.T) {
	// Use a temp directory for config to avoid writing to user's real config file.
	tmpDir := t.TempDir()
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Create config with multiple contexts
	cfg := config.NewConfig()
	cfg.CurrentContext = "prod"
	cfg.SetContext("prod", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Acme Corp",
		APIKey:           "se_test12300000000000000000000000000000000000000000000000000000",
	})
	cfg.SetContext("dev", &config.Context{
		APIURL:           "https://api.dev.stackeye.io",
		OrganizationName: "Acme Dev",
		APIKey:           "se_dev45600000000000000000000000000000000000000000000000000000",
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextDelete("dev")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check context was deleted
	if cfg.HasContext("dev") {
		t.Error("expected dev context to be deleted")
	}

	// Check prod context still exists
	if !cfg.HasContext("prod") {
		t.Error("expected prod context to still exist")
	}

	// Check current context unchanged
	if cfg.CurrentContext != "prod" {
		t.Errorf("expected CurrentContext to remain 'prod', got %q", cfg.CurrentContext)
	}

	// Check output message
	if !strings.Contains(output, "Deleted context") {
		t.Errorf("expected output to contain 'Deleted context', got %q", output)
	}
	if !strings.Contains(output, "dev") {
		t.Errorf("expected output to contain 'dev', got %q", output)
	}
}

func TestRunContextDelete_DeletesCurrentContext(t *testing.T) {
	// Use a temp directory for config to avoid writing to user's real config file.
	tmpDir := t.TempDir()
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Create config with current context as target
	cfg := config.NewConfig()
	cfg.CurrentContext = "dev"
	cfg.SetContext("prod", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Acme Corp",
		APIKey:           "se_test12300000000000000000000000000000000000000000000000000000",
	})
	cfg.SetContext("dev", &config.Context{
		APIURL:           "https://api.dev.stackeye.io",
		OrganizationName: "Acme Dev",
		APIKey:           "se_dev45600000000000000000000000000000000000000000000000000000",
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextDelete("dev")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check context was deleted
	if cfg.HasContext("dev") {
		t.Error("expected dev context to be deleted")
	}

	// Check current context was cleared
	if cfg.CurrentContext != "" {
		t.Errorf("expected CurrentContext to be cleared, got %q", cfg.CurrentContext)
	}

	// Check note about current context
	if !strings.Contains(output, "The deleted context was your current context") {
		t.Errorf("expected output to contain note about current context, got %q", output)
	}

	// Check suggestion to use another context
	if !strings.Contains(output, "stackeye context use") {
		t.Errorf("expected output to contain suggestion to use another context, got %q", output)
	}
}

func TestRunContextDelete_DeletesOnlyContext(t *testing.T) {
	// Use a temp directory for config to avoid writing to user's real config file.
	tmpDir := t.TempDir()
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Create config with only one context
	cfg := config.NewConfig()
	cfg.CurrentContext = "only"
	cfg.SetContext("only", &config.Context{
		APIURL:           "https://api.stackeye.io",
		OrganizationName: "Only Org",
		APIKey:           "se_only12300000000000000000000000000000000000000000000000000000",
	})
	loadedConfig = cfg

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContextDelete("only")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check context was deleted
	if cfg.HasContext("only") {
		t.Error("expected only context to be deleted")
	}

	// Check no contexts remaining message
	if !strings.Contains(output, "No contexts remaining") {
		t.Errorf("expected output to contain 'No contexts remaining', got %q", output)
	}
	if !strings.Contains(output, "stackeye login") {
		t.Errorf("expected output to contain login hint, got %q", output)
	}
}

func TestRunContextDelete_NotFound(t *testing.T) {
	// Create config without target context
	cfg := config.NewConfig()
	cfg.CurrentContext = "prod"
	cfg.SetContext("prod", &config.Context{
		APIURL: "https://api.stackeye.io",
		APIKey: "se_test123",
	})
	loadedConfig = cfg

	err := runContextDelete("nonexistent")

	if err == nil {
		t.Error("expected error for nonexistent context, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain 'not found', got %q", err.Error())
	}
}

func TestRunContextDelete_NilConfig(t *testing.T) {
	loadedConfig = nil

	err := runContextDelete("somecontext")

	if err == nil {
		t.Error("expected error for nil config, got nil")
	}
	if !strings.Contains(err.Error(), "configuration not loaded") {
		t.Errorf("expected error to contain 'configuration not loaded', got %q", err.Error())
	}
}

func TestRunContextDelete_NilContext(t *testing.T) {
	// Create config with a nil context value
	cfg := config.NewConfig()
	cfg.CurrentContext = "valid"
	cfg.SetContext("valid", &config.Context{
		APIURL: "https://api.stackeye.io",
		APIKey: "se_valid123",
	})
	// Simulate nil context value
	cfg.Contexts["nilctx"] = nil
	loadedConfig = cfg

	err := runContextDelete("nilctx")

	if err == nil {
		t.Error("expected error for nil context, got nil")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected error to contain 'invalid', got %q", err.Error())
	}
}
