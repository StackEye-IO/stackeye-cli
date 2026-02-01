// Package config provides CLI-specific configuration management.
//
// This package wraps the stackeye-go-sdk/config package to add CLI-specific
// functionality such as environment variable overrides and authentication
// validation helpers. The SDK package handles persistence and context
// management; this package adds runtime configuration concerns.
//
// Environment variables (override config file values):
//
//	STACKEYE_API_KEY    - API key for authentication
//	STACKEYE_API_URL    - API server URL
//	STACKEYE_CONFIG     - Path to config file
//	STACKEYE_CONTEXT    - Context name to use
//
// Usage:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    return err
//	}
//
//	// Get API key with env override
//	apiKey := config.GetAPIKey(cfg)
//
//	// Require valid authentication
//	if err := config.RequireAuth(cfg); err != nil {
//	    return err
//	}
package config

import (
	"errors"
	"fmt"
	"os"
	"sync"

	sdkconfig "github.com/StackEye-IO/stackeye-go-sdk/config"
)

// Environment variable names for configuration overrides.
const (
	// EnvAPIKey overrides the API key from config.
	EnvAPIKey = "STACKEYE_API_KEY"

	// EnvAPIURL overrides the API URL from config.
	EnvAPIURL = "STACKEYE_API_URL"

	// EnvConfig specifies a custom config file path.
	EnvConfig = "STACKEYE_CONFIG"

	// EnvContext overrides the current context.
	EnvContext = "STACKEYE_CONTEXT"
)

// Re-export SDK types for convenience.
// This allows CLI commands to import only the internal config package.
type (
	// Config is the root configuration structure.
	Config = sdkconfig.Config

	// Context represents a named configuration context.
	Context = sdkconfig.Context

	// Preferences holds user preferences for CLI behavior.
	Preferences = sdkconfig.Preferences

	// OutputFormat specifies the output format for commands.
	OutputFormat = sdkconfig.OutputFormat

	// ColorMode specifies when to use colored output.
	ColorMode = sdkconfig.ColorMode

	// ContextInfo provides display information about a context.
	ContextInfo = sdkconfig.ContextInfo
)

// Re-export SDK constants.
const (
	// Output formats
	OutputFormatTable = sdkconfig.OutputFormatTable
	OutputFormatJSON  = sdkconfig.OutputFormatJSON
	OutputFormatYAML  = sdkconfig.OutputFormatYAML
	OutputFormatWide  = sdkconfig.OutputFormatWide

	// Color modes
	ColorModeAuto   = sdkconfig.ColorModeAuto
	ColorModeAlways = sdkconfig.ColorModeAlways
	ColorModeNever  = sdkconfig.ColorModeNever

	// Default values
	DefaultAPIURL = sdkconfig.DefaultAPIURL
)

// Re-export SDK errors.
var (
	ErrNoCurrentContext = sdkconfig.ErrNoCurrentContext
	ErrContextNotFound  = sdkconfig.ErrContextNotFound
	ErrContextExists    = sdkconfig.ErrContextExists
)

// CLI-specific errors.
var (
	// ErrNotAuthenticated is returned when authentication is required but not configured.
	ErrNotAuthenticated = errors.New("not authenticated: run 'stackeye login' to authenticate")

	// ErrInvalidAPIKey is returned when the API key format is invalid.
	ErrInvalidAPIKey = errors.New("invalid API key format")
)

// Manager provides thread-safe access to configuration.
// Use GetManager() to obtain the singleton instance.
type Manager struct {
	mu     sync.RWMutex
	config *Config
	loaded bool
}

var (
	manager     *Manager
	managerOnce sync.Once
)

// GetManager returns the singleton configuration manager.
func GetManager() *Manager {
	managerOnce.Do(func() {
		manager = &Manager{}
	})
	return manager
}

// ResetManager resets the singleton manager for testing purposes.
// This should only be used in tests to ensure test isolation.
func ResetManager() {
	managerOnce = sync.Once{}
	manager = nil
}

// Load loads the configuration, applying environment variable overrides.
// If STACKEYE_CONFIG is set, loads from that path; otherwise uses default location.
func (m *Manager) Load() (*Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.loaded && m.config != nil {
		return m.config, nil
	}

	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	m.config = cfg
	m.loaded = true
	return cfg, nil
}

// Get returns the loaded configuration.
// Returns nil if Load() hasn't been called.
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Reload forces a reload of the configuration from disk.
func (m *Manager) Reload() (*Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.loaded = false
	m.config = nil

	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	m.config = cfg
	m.loaded = true
	return cfg, nil
}

// Save persists the current configuration to disk.
// Uses ConfigPath() to respect the STACKEYE_CONFIG environment variable.
func (m *Manager) Save() error {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if cfg == nil {
		return errors.New("no configuration loaded")
	}

	// Use CLI's ConfigPath() which respects STACKEYE_CONFIG env var,
	// instead of the SDK's Save() which uses the SDK's ConfigPath().
	return cfg.SaveTo(ConfigPath())
}

// loadConfig loads configuration from file and applies environment overrides.
func loadConfig() (*Config, error) {
	var cfg *Config
	var err error

	// Check for custom config path
	if configPath := os.Getenv(EnvConfig); configPath != "" {
		cfg, err = sdkconfig.LoadFrom(configPath)
	} else {
		cfg, err = sdkconfig.Load()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Apply context override from environment
	if ctxName := os.Getenv(EnvContext); ctxName != "" {
		if err := cfg.UseContext(ctxName); err != nil {
			return nil, fmt.Errorf("failed to use context %q: %w", ctxName, err)
		}
	}

	return cfg, nil
}

// Load loads configuration using the default manager.
// This is a convenience function for simple use cases.
func Load() (*Config, error) {
	return GetManager().Load()
}

// Save saves configuration using the default manager.
func Save() error {
	return GetManager().Save()
}

// GetAPIKey returns the effective API key, checking environment first.
// Returns empty string if no API key is configured.
func GetAPIKey(cfg *Config) string {
	// Environment variable takes precedence
	if envKey := os.Getenv(EnvAPIKey); envKey != "" {
		return envKey
	}

	// Fall back to config
	if cfg == nil {
		return ""
	}

	ctx, err := cfg.GetCurrentContext()
	if err != nil || ctx == nil {
		return ""
	}

	return ctx.APIKey
}

// GetAPIURL returns the effective API URL, checking environment first.
// Returns DefaultAPIURL if no URL is configured.
func GetAPIURL(cfg *Config) string {
	// Environment variable takes precedence
	if envURL := os.Getenv(EnvAPIURL); envURL != "" {
		return envURL
	}

	// Fall back to config
	if cfg == nil {
		return DefaultAPIURL
	}

	ctx, err := cfg.GetCurrentContext()
	if err != nil || ctx == nil {
		return DefaultAPIURL
	}

	if ctx.APIURL != "" {
		return ctx.APIURL
	}

	return DefaultAPIURL
}

// IsAuthenticated returns true if valid authentication is configured.
// Checks both environment variables and config file.
func IsAuthenticated(cfg *Config) bool {
	apiKey := GetAPIKey(cfg)
	return apiKey != ""
}

// RequireAuth returns an error if no valid authentication is configured.
// Use this at the start of commands that require authentication.
func RequireAuth(cfg *Config) error {
	if !IsAuthenticated(cfg) {
		return ErrNotAuthenticated
	}
	return nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist.
// Returns the path to the config directory.
func EnsureConfigDir() (string, error) {
	dir := sdkconfig.ConfigDir()

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return dir, nil
}

// ConfigPath returns the path to the configuration file.
// If STACKEYE_CONFIG is set, returns that path; otherwise returns default.
func ConfigPath() string {
	if configPath := os.Getenv(EnvConfig); configPath != "" {
		return configPath
	}
	return sdkconfig.ConfigPath()
}

// ConfigDir returns the path to the configuration directory.
func ConfigDir() string {
	return sdkconfig.ConfigDir()
}

// NewConfig creates a new empty configuration.
func NewConfig() *Config {
	return sdkconfig.NewConfig()
}

// NewContext creates a new context with default values.
func NewContext() *Context {
	return sdkconfig.NewContext()
}

// NewPreferences creates new preferences with default values.
func NewPreferences() *Preferences {
	return sdkconfig.NewPreferences()
}

// ValidateContextName checks if a context name is valid.
func ValidateContextName(name string) error {
	return sdkconfig.ValidateContextName(name)
}
