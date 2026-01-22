package testutil

import (
	"os"
	"path/filepath"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

// Environment variables for live testing.
const (
	// EnvLiveTest enables live API testing when set to "true" or "1".
	EnvLiveTest = "STACKEYE_E2E_LIVE"
	// EnvLiveAPIURL overrides the API URL for live testing.
	EnvLiveAPIURL = "STACKEYE_E2E_API_URL"
	// DefaultLiveAPIURL is the default API URL for live testing.
	DefaultLiveAPIURL = "https://api-dev.stackeye.io"
)

// IsLiveTestEnabled returns true if live testing is enabled via environment.
func IsLiveTestEnabled() bool {
	v := os.Getenv(EnvLiveTest)
	return v == "true" || v == "1"
}

// GetLiveAPIURL returns the API URL for live testing.
func GetLiveAPIURL() string {
	if url := os.Getenv(EnvLiveAPIURL); url != "" {
		return url
	}
	return DefaultLiveAPIURL
}

// TestConfig manages temporary configuration for E2E tests.
type TestConfig struct {
	// Dir is the temporary config directory.
	Dir string
	// ConfigPath is the path to the config file.
	ConfigPath string
	// Config is the configuration object.
	Config *config.Config
}

// TestAPIKey is a fixed API key for testing.
const TestAPIKey = "se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// TestContextName is the default test context name.
const TestContextName = "test"

// NewTestConfig creates a new test configuration with a temporary directory.
// The returned TestConfig must be cleaned up with Cleanup().
//
// The directory structure matches XDG_CONFIG_HOME expectations:
// - Dir (XDG_CONFIG_HOME): /tmp/stackeye-e2e-xxx/
// - ConfigPath: /tmp/stackeye-e2e-xxx/stackeye/config.yaml
//
// This allows the CLI's Save() function to work correctly when
// XDG_CONFIG_HOME is set to Dir.
func NewTestConfig(apiURL string) (*TestConfig, error) {
	// Create temp directory (will be XDG_CONFIG_HOME)
	dir, err := os.MkdirTemp("", "stackeye-e2e-*")
	if err != nil {
		return nil, err
	}

	// Create stackeye subdirectory to match XDG_CONFIG_HOME/stackeye/config.yaml
	stackeyeDir := filepath.Join(dir, "stackeye")
	if err := os.MkdirAll(stackeyeDir, 0700); err != nil {
		os.RemoveAll(dir)
		return nil, err
	}

	configPath := filepath.Join(stackeyeDir, "config.yaml")

	tc := &TestConfig{
		Dir:        dir,
		ConfigPath: configPath,
		Config:     config.NewConfig(),
	}

	// Create a test context pointing to the mock server
	ctx := config.NewContext()
	ctx.APIURL = apiURL
	ctx.APIKey = TestAPIKey
	ctx.OrganizationID = "org_test123"
	ctx.OrganizationName = "Test Organization"

	tc.Config.SetContext(TestContextName, ctx)
	tc.Config.CurrentContext = TestContextName

	// Save the config
	if err := tc.Config.SaveTo(configPath); err != nil {
		os.RemoveAll(dir)
		return nil, err
	}

	return tc, nil
}

// NewTestConfigWithMultipleContexts creates a test config with multiple contexts.
func NewTestConfigWithMultipleContexts(apiURL string) (*TestConfig, error) {
	tc, err := NewTestConfig(apiURL)
	if err != nil {
		return nil, err
	}

	// Add additional contexts
	ctx2 := config.NewContext()
	ctx2.APIURL = apiURL
	ctx2.APIKey = "se_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	ctx2.OrganizationID = "org_staging123"
	ctx2.OrganizationName = "Staging Organization"
	tc.Config.SetContext("staging", ctx2)

	ctx3 := config.NewContext()
	ctx3.APIURL = apiURL
	ctx3.APIKey = "se_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	ctx3.OrganizationID = "org_prod456"
	ctx3.OrganizationName = "Production Organization"
	tc.Config.SetContext("production", ctx3)

	// Save the config
	if err := tc.Config.SaveTo(tc.ConfigPath); err != nil {
		tc.Cleanup()
		return nil, err
	}

	return tc, nil
}

// Cleanup removes the temporary configuration directory.
func (tc *TestConfig) Cleanup() {
	if tc.Dir != "" {
		os.RemoveAll(tc.Dir)
	}
}

// SetCurrentContext changes the current context and saves the config.
func (tc *TestConfig) SetCurrentContext(name string) error {
	tc.Config.CurrentContext = name
	return tc.Config.SaveTo(tc.ConfigPath)
}

// AddContext adds a new context and saves the config.
func (tc *TestConfig) AddContext(name string, ctx *config.Context) error {
	tc.Config.SetContext(name, ctx)
	return tc.Config.SaveTo(tc.ConfigPath)
}

// RemoveContext removes a context and saves the config.
func (tc *TestConfig) RemoveContext(name string) error {
	tc.Config.DeleteContext(name)
	return tc.Config.SaveTo(tc.ConfigPath)
}

// Reload reloads the config from disk.
func (tc *TestConfig) Reload() error {
	cfg, err := config.LoadFrom(tc.ConfigPath)
	if err != nil {
		return err
	}
	tc.Config = cfg
	return nil
}
