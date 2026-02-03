// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-cli/internal/config"
	"github.com/spf13/cobra"
)

func TestContextCompletion_NoArgs(t *testing.T) {
	// Create a temporary config file with test contexts
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `current_context: acme-prod
contexts:
  acme-prod:
    api_key: "sk_test_prod"
    api_url: "https://api.stackeye.io"
    organization_name: "Acme Production"
  acme-staging:
    api_key: "sk_test_staging"
    api_url: "https://api.stackeye.io"
    organization_name: "Acme Staging"
  beta-org:
    api_key: "sk_test_beta"
    api_url: "https://api.stackeye.io"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment to use our test config
	originalEnv := os.Getenv("STACKEYE_CONFIG")
	os.Setenv("STACKEYE_CONFIG", configPath)
	defer os.Setenv("STACKEYE_CONFIG", originalEnv)

	// Reset the config manager to pick up new config
	config.ResetManager()

	// Create a test command
	cmd := &cobra.Command{Use: "test"}

	// Get the completion function
	completionFunc := ContextCompletion()

	// Test completion with no args and empty prefix
	completions, directive := completionFunc(cmd, []string{}, "")

	// Verify directive
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}

	// Verify we got all 3 contexts
	if len(completions) != 3 {
		t.Errorf("Expected 3 completions, got %d: %v", len(completions), completions)
	}

	// Verify completions include expected context names
	foundProd := false
	foundStaging := false
	foundBeta := false

	for _, c := range completions {
		if strings.HasPrefix(c, "acme-prod\t") {
			foundProd = true
			if !strings.Contains(c, "Acme Production") {
				t.Errorf("acme-prod completion missing org name: %s", c)
			}
			if !strings.Contains(c, "[current]") {
				t.Errorf("acme-prod should be marked as [current]: %s", c)
			}
		}
		if strings.HasPrefix(c, "acme-staging\t") {
			foundStaging = true
			if !strings.Contains(c, "Acme Staging") {
				t.Errorf("acme-staging completion missing org name: %s", c)
			}
		}
		if strings.HasPrefix(c, "beta-org\t") {
			foundBeta = true
			if !strings.Contains(c, "(no org)") {
				t.Errorf("beta-org should show (no org): %s", c)
			}
		}
	}

	if !foundProd {
		t.Error("Missing acme-prod in completions")
	}
	if !foundStaging {
		t.Error("Missing acme-staging in completions")
	}
	if !foundBeta {
		t.Error("Missing beta-org in completions")
	}
}

func TestContextCompletion_WithPrefix(t *testing.T) {
	// Create a temporary config file with test contexts
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `current_context: acme-prod
contexts:
  acme-prod:
    api_key: "sk_test_prod"
    organization_name: "Acme Production"
  acme-staging:
    api_key: "sk_test_staging"
    organization_name: "Acme Staging"
  beta-org:
    api_key: "sk_test_beta"
    organization_name: "Beta Organization"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment to use our test config
	originalEnv := os.Getenv("STACKEYE_CONFIG")
	os.Setenv("STACKEYE_CONFIG", configPath)
	defer os.Setenv("STACKEYE_CONFIG", originalEnv)

	// Reset the config manager to pick up new config
	config.ResetManager()

	// Create a test command
	cmd := &cobra.Command{Use: "test"}

	// Get the completion function
	completionFunc := ContextCompletion()

	// Test completion with "acme" prefix
	completions, directive := completionFunc(cmd, []string{}, "acme")

	// Verify directive
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}

	// Verify we got only 2 contexts (acme-prod and acme-staging)
	if len(completions) != 2 {
		t.Errorf("Expected 2 completions for 'acme' prefix, got %d: %v", len(completions), completions)
	}

	// Verify beta-org is not included
	for _, c := range completions {
		if strings.HasPrefix(c, "beta-org") {
			t.Errorf("beta-org should not be in completions for 'acme' prefix: %v", completions)
		}
	}
}

func TestContextCompletion_CaseInsensitive(t *testing.T) {
	// Create a temporary config file with test contexts
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `current_context: ACME-PROD
contexts:
  ACME-PROD:
    api_key: "sk_test_prod"
    organization_name: "Acme Production"
  acme-staging:
    api_key: "sk_test_staging"
    organization_name: "Acme Staging"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment to use our test config
	originalEnv := os.Getenv("STACKEYE_CONFIG")
	os.Setenv("STACKEYE_CONFIG", configPath)
	defer os.Setenv("STACKEYE_CONFIG", originalEnv)

	// Reset the config manager to pick up new config
	config.ResetManager()

	// Create a test command
	cmd := &cobra.Command{Use: "test"}

	// Get the completion function
	completionFunc := ContextCompletion()

	// Test completion with lowercase "acme" should match uppercase ACME-PROD
	completions, _ := completionFunc(cmd, []string{}, "acme")

	// Should match both contexts (case-insensitive)
	if len(completions) != 2 {
		t.Errorf("Expected 2 completions (case-insensitive), got %d: %v", len(completions), completions)
	}
}

func TestContextCompletion_AlreadyHasArg(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `current_context: acme-prod
contexts:
  acme-prod:
    api_key: "sk_test_prod"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment to use our test config
	originalEnv := os.Getenv("STACKEYE_CONFIG")
	os.Setenv("STACKEYE_CONFIG", configPath)
	defer os.Setenv("STACKEYE_CONFIG", originalEnv)

	// Reset the config manager
	config.ResetManager()

	// Create a test command
	cmd := &cobra.Command{Use: "test"}

	// Get the completion function
	completionFunc := ContextCompletion()

	// Test completion when an argument is already provided
	completions, directive := completionFunc(cmd, []string{"existing-arg"}, "")

	// Should return no completions
	if len(completions) != 0 {
		t.Errorf("Expected no completions when arg already provided, got %d: %v", len(completions), completions)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

func TestContextCompletion_EmptyConfig(t *testing.T) {
	// Create a temporary config file with no contexts
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `current_context: ""
contexts: {}
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment to use our test config
	originalEnv := os.Getenv("STACKEYE_CONFIG")
	os.Setenv("STACKEYE_CONFIG", configPath)
	defer os.Setenv("STACKEYE_CONFIG", originalEnv)

	// Reset the config manager
	config.ResetManager()

	// Create a test command
	cmd := &cobra.Command{Use: "test"}

	// Get the completion function
	completionFunc := ContextCompletion()

	// Test completion with empty config
	completions, directive := completionFunc(cmd, []string{}, "")

	// Should return no completions
	if len(completions) != 0 {
		t.Errorf("Expected no completions for empty config, got %d: %v", len(completions), completions)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

func TestContextCompletion_NoMatchingPrefix(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `current_context: acme-prod
contexts:
  acme-prod:
    api_key: "sk_test_prod"
    organization_name: "Acme Production"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment to use our test config
	originalEnv := os.Getenv("STACKEYE_CONFIG")
	os.Setenv("STACKEYE_CONFIG", configPath)
	defer os.Setenv("STACKEYE_CONFIG", originalEnv)

	// Reset the config manager
	config.ResetManager()

	// Create a test command
	cmd := &cobra.Command{Use: "test"}

	// Get the completion function
	completionFunc := ContextCompletion()

	// Test completion with prefix that doesn't match any context
	completions, directive := completionFunc(cmd, []string{}, "xyz")

	// Should return no completions
	if len(completions) != 0 {
		t.Errorf("Expected no completions for non-matching prefix, got %d: %v", len(completions), completions)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

func TestContextCompletion_ConfigLoadError(t *testing.T) {
	// Set environment to use a non-existent config path
	originalEnv := os.Getenv("STACKEYE_CONFIG")
	os.Setenv("STACKEYE_CONFIG", "/non/existent/path/config.yaml")
	defer os.Setenv("STACKEYE_CONFIG", originalEnv)

	// Reset the config manager to trigger the error
	config.ResetManager()

	// Create a test command
	cmd := &cobra.Command{Use: "test"}

	// Get the completion function
	completionFunc := ContextCompletion()

	// Test completion should gracefully handle config load error
	completions, directive := completionFunc(cmd, []string{}, "")

	// Should return no completions silently
	if len(completions) != 0 {
		t.Errorf("Expected no completions on config error, got %d: %v", len(completions), completions)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

func TestContextNameCompletion_Alias(t *testing.T) {
	// Verify ContextNameCompletion is an alias for ContextCompletion
	if ContextNameCompletion == nil {
		t.Error("ContextNameCompletion should not be nil")
	}

	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `current_context: test-ctx
contexts:
  test-ctx:
    api_key: "sk_test"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	originalEnv := os.Getenv("STACKEYE_CONFIG")
	os.Setenv("STACKEYE_CONFIG", configPath)
	defer os.Setenv("STACKEYE_CONFIG", originalEnv)

	config.ResetManager()

	cmd := &cobra.Command{Use: "test"}

	// Both should produce the same results
	completionFunc1 := ContextCompletion()
	completionFunc2 := ContextNameCompletion()

	comp1, dir1 := completionFunc1(cmd, []string{}, "")
	comp2, dir2 := completionFunc2(cmd, []string{}, "")

	if dir1 != dir2 {
		t.Errorf("Directives should match: %v vs %v", dir1, dir2)
	}

	if len(comp1) != len(comp2) {
		t.Errorf("Completion counts should match: %d vs %d", len(comp1), len(comp2))
	}
}

func TestNewContextUseCmd_HasValidArgsFunction(t *testing.T) {
	cmd := newContextUseCmd()

	if cmd.ValidArgsFunction == nil {
		t.Error("context use command should have ValidArgsFunction set")
	}
}

func TestNewContextDeleteCmd_HasValidArgsFunction(t *testing.T) {
	cmd := newContextDeleteCmd()

	if cmd.ValidArgsFunction == nil {
		t.Error("context delete command should have ValidArgsFunction set")
	}
}
