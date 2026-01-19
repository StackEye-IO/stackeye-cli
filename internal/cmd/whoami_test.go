package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

func TestNewWhoamiCmd(t *testing.T) {
	cmd := NewWhoamiCmd()

	if cmd.Use != "whoami" {
		t.Errorf("NewWhoamiCmd().Use = %q, want %q", cmd.Use, "whoami")
	}

	if cmd.Short == "" {
		t.Error("NewWhoamiCmd().Short should not be empty")
	}

	if cmd.Long == "" {
		t.Error("NewWhoamiCmd().Long should not be empty")
	}
}

func TestWhoamiCmd_HelpText(t *testing.T) {
	cmd := NewWhoamiCmd()

	// Verify help text contains important information
	if !strings.Contains(cmd.Long, "user") {
		t.Error("Long description should mention user")
	}

	if !strings.Contains(cmd.Long, "context") {
		t.Error("Long description should mention context")
	}

	if !strings.Contains(cmd.Long, "stackeye login") {
		t.Error("Long description should mention how to log in")
	}

	if !strings.Contains(cmd.Long, "--context") {
		t.Error("Long description examples should show --context flag usage")
	}
}

func TestRunWhoami_NilConfig(t *testing.T) {
	// Save and restore global config
	originalConfig := loadedConfig
	defer func() { loadedConfig = originalConfig }()

	// Set config to nil
	loadedConfig = nil

	// Run whoami - should not error, just print message
	err := runWhoami()
	if err != nil {
		t.Errorf("runWhoami() with nil config should return nil, got error: %v", err)
	}
}

func TestRunWhoami_EmptyCurrentContext(t *testing.T) {
	// Save and restore global config
	originalConfig := loadedConfig
	defer func() { loadedConfig = originalConfig }()

	// Create config with empty current context
	cfg := &configForTest{
		currentContext: "",
	}
	loadedConfig = cfg.toConfig()

	// Run whoami - should not error, just print message
	err := runWhoami()
	if err != nil {
		t.Errorf("runWhoami() with empty current context should return nil, got error: %v", err)
	}
}

// configForTest is a helper to create test configurations.
type configForTest struct {
	currentContext string
	contextExists  bool
	hasAPIKey      bool
	apiURL         string
	orgName        string
	orgID          string
}

func (c *configForTest) toConfig() *config.Config {
	cfg := config.NewConfig()
	cfg.CurrentContext = c.currentContext

	if c.contextExists {
		ctx := &config.Context{
			APIURL:           c.apiURL,
			OrganizationID:   c.orgID,
			OrganizationName: c.orgName,
		}
		if c.hasAPIKey {
			ctx.APIKey = "se_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		}
		cfg.SetContext(c.currentContext, ctx)
	}

	return cfg
}

func TestRunWhoami_ContextNotFound(t *testing.T) {
	// Save and restore global config
	originalConfig := loadedConfig
	defer func() { loadedConfig = originalConfig }()

	// Create config with current context pointing to non-existent context
	cfg := config.NewConfig()
	cfg.CurrentContext = "nonexistent"
	// Don't add any contexts
	loadedConfig = cfg

	// Run whoami - should not error, just print message
	err := runWhoami()
	if err != nil {
		t.Errorf("runWhoami() with nonexistent context should return nil, got error: %v", err)
	}
}

func TestRunWhoami_NoAPIKey(t *testing.T) {
	// Save and restore global config
	originalConfig := loadedConfig
	defer func() { loadedConfig = originalConfig }()

	// Create config with context but no API key
	cfg := &configForTest{
		currentContext: "test-context",
		contextExists:  true,
		hasAPIKey:      false,
		apiURL:         "https://api.stackeye.io",
		orgName:        "Test Org",
	}
	loadedConfig = cfg.toConfig()

	// Run whoami - should not error, just print message
	err := runWhoami()
	if err != nil {
		t.Errorf("runWhoami() with no API key should return nil, got error: %v", err)
	}
}
