package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetAPIKey(t *testing.T) {
	// Save and restore environment
	origKey := os.Getenv(EnvAPIKey)
	defer os.Setenv(EnvAPIKey, origKey)

	tests := []struct {
		name     string
		envValue string
		cfg      *Config
		want     string
	}{
		{
			name:     "env takes precedence",
			envValue: "se_env_key_123",
			cfg:      configWithAPIKey("se_config_key_456"),
			want:     "se_env_key_123",
		},
		{
			name:     "falls back to config",
			envValue: "",
			cfg:      configWithAPIKey("se_config_key_456"),
			want:     "se_config_key_456",
		},
		{
			name:     "empty when both missing",
			envValue: "",
			cfg:      nil,
			want:     "",
		},
		{
			name:     "empty when config has no context",
			envValue: "",
			cfg:      NewConfig(),
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(EnvAPIKey, tt.envValue)
			got := GetAPIKey(tt.cfg)
			if got != tt.want {
				t.Errorf("GetAPIKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetAPIURL(t *testing.T) {
	// Save and restore environment
	origURL := os.Getenv(EnvAPIURL)
	defer os.Setenv(EnvAPIURL, origURL)

	tests := []struct {
		name     string
		envValue string
		cfg      *Config
		want     string
	}{
		{
			name:     "env takes precedence",
			envValue: "https://api.custom.io",
			cfg:      configWithAPIURL("https://api.config.io"),
			want:     "https://api.custom.io",
		},
		{
			name:     "falls back to config",
			envValue: "",
			cfg:      configWithAPIURL("https://api.config.io"),
			want:     "https://api.config.io",
		},
		{
			name:     "default when both missing",
			envValue: "",
			cfg:      nil,
			want:     DefaultAPIURL,
		},
		{
			name:     "default when config has no context",
			envValue: "",
			cfg:      NewConfig(),
			want:     DefaultAPIURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(EnvAPIURL, tt.envValue)
			got := GetAPIURL(tt.cfg)
			if got != tt.want {
				t.Errorf("GetAPIURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsAuthenticated(t *testing.T) {
	// Save and restore environment
	origKey := os.Getenv(EnvAPIKey)
	defer os.Setenv(EnvAPIKey, origKey)

	tests := []struct {
		name     string
		envValue string
		cfg      *Config
		want     bool
	}{
		{
			name:     "authenticated via env",
			envValue: "se_test_key",
			cfg:      nil,
			want:     true,
		},
		{
			name:     "authenticated via config",
			envValue: "",
			cfg:      configWithAPIKey("se_config_key"),
			want:     true,
		},
		{
			name:     "not authenticated",
			envValue: "",
			cfg:      nil,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(EnvAPIKey, tt.envValue)
			got := IsAuthenticated(tt.cfg)
			if got != tt.want {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequireAuth(t *testing.T) {
	// Save and restore environment
	origKey := os.Getenv(EnvAPIKey)
	defer os.Setenv(EnvAPIKey, origKey)

	tests := []struct {
		name      string
		envValue  string
		cfg       *Config
		wantError bool
	}{
		{
			name:      "no error when authenticated via env",
			envValue:  "se_test_key",
			cfg:       nil,
			wantError: false,
		},
		{
			name:      "no error when authenticated via config",
			envValue:  "",
			cfg:       configWithAPIKey("se_config_key"),
			wantError: false,
		},
		{
			name:      "error when not authenticated",
			envValue:  "",
			cfg:       nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(EnvAPIKey, tt.envValue)
			err := RequireAuth(tt.cfg)
			if (err != nil) != tt.wantError {
				t.Errorf("RequireAuth() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && err != ErrNotAuthenticated {
				t.Errorf("RequireAuth() error = %v, want %v", err, ErrNotAuthenticated)
			}
		})
	}
}

func TestConfigPath(t *testing.T) {
	// Save and restore environment
	origConfig := os.Getenv(EnvConfig)
	defer os.Setenv(EnvConfig, origConfig)

	t.Run("returns env value when set", func(t *testing.T) {
		customPath := "/custom/path/config.yaml"
		os.Setenv(EnvConfig, customPath)
		got := ConfigPath()
		if got != customPath {
			t.Errorf("ConfigPath() = %q, want %q", got, customPath)
		}
	})

	t.Run("returns default when env not set", func(t *testing.T) {
		os.Setenv(EnvConfig, "")
		got := ConfigPath()
		if got == "" {
			t.Error("ConfigPath() returned empty string")
		}
		// Should end with config.yaml
		if filepath.Base(got) != "config.yaml" {
			t.Errorf("ConfigPath() = %q, expected to end with config.yaml", got)
		}
	})
}

func TestEnsureConfigDir(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()

	// Save and restore environment
	origConfig := os.Getenv(EnvConfig)
	defer os.Setenv(EnvConfig, origConfig)

	// Set config path to a file in the temp directory
	testConfigPath := filepath.Join(tmpDir, "stackeye", "config.yaml")
	os.Setenv(EnvConfig, testConfigPath)

	// Note: EnsureConfigDir uses the SDK's ConfigDir which won't respect
	// our env override, so we test that it returns a valid directory path
	dir, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}
	if dir == "" {
		t.Error("EnsureConfigDir() returned empty path")
	}

	// Verify the directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("failed to stat config dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("EnsureConfigDir() path %q is not a directory", dir)
	}
}

func TestManagerSingleton(t *testing.T) {
	m1 := GetManager()
	m2 := GetManager()

	if m1 != m2 {
		t.Error("GetManager() should return the same instance")
	}
}

func TestEnvironmentConstants(t *testing.T) {
	// Verify environment variable names are correct
	tests := []struct {
		name string
		env  string
		want string
	}{
		{"EnvAPIKey", EnvAPIKey, "STACKEYE_API_KEY"},
		{"EnvAPIURL", EnvAPIURL, "STACKEYE_API_URL"},
		{"EnvConfig", EnvConfig, "STACKEYE_CONFIG"},
		{"EnvContext", EnvContext, "STACKEYE_CONTEXT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.env, tt.want)
			}
		})
	}
}

func TestReExportedTypes(t *testing.T) {
	// Verify re-exported types work correctly
	cfg := NewConfig()
	if cfg == nil {
		t.Error("NewConfig() returned nil")
	}

	ctx := NewContext()
	if ctx == nil {
		t.Error("NewContext() returned nil")
	}

	prefs := NewPreferences()
	if prefs == nil {
		t.Error("NewPreferences() returned nil")
	}
}

func TestReExportedConstants(t *testing.T) {
	// Verify re-exported constants have expected values
	if OutputFormatTable != "table" {
		t.Errorf("OutputFormatTable = %q, want %q", OutputFormatTable, "table")
	}
	if OutputFormatJSON != "json" {
		t.Errorf("OutputFormatJSON = %q, want %q", OutputFormatJSON, "json")
	}
	if OutputFormatYAML != "yaml" {
		t.Errorf("OutputFormatYAML = %q, want %q", OutputFormatYAML, "yaml")
	}
	if OutputFormatWide != "wide" {
		t.Errorf("OutputFormatWide = %q, want %q", OutputFormatWide, "wide")
	}

	if ColorModeAuto != "auto" {
		t.Errorf("ColorModeAuto = %q, want %q", ColorModeAuto, "auto")
	}
	if ColorModeAlways != "always" {
		t.Errorf("ColorModeAlways = %q, want %q", ColorModeAlways, "always")
	}
	if ColorModeNever != "never" {
		t.Errorf("ColorModeNever = %q, want %q", ColorModeNever, "never")
	}
}

func TestValidateContextName(t *testing.T) {
	tests := []struct {
		name      string
		ctxName   string
		wantError bool
	}{
		{"valid simple name", "production", false},
		{"valid with hyphen", "prod-us-east", false},
		{"valid with underscore", "prod_backup", false},
		{"valid with period", "v1.0", false},
		{"empty name", "", true},
		{"invalid with space", "my context", true},
		{"invalid with special char", "prod@main", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContextName(tt.ctxName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateContextName(%q) error = %v, wantError %v", tt.ctxName, err, tt.wantError)
			}
		})
	}
}

// Helper functions

func configWithAPIKey(apiKey string) *Config {
	cfg := NewConfig()
	ctx := NewContext()
	ctx.APIKey = apiKey
	cfg.Contexts = map[string]*Context{"default": ctx}
	cfg.CurrentContext = "default"
	return cfg
}

func configWithAPIURL(apiURL string) *Config {
	cfg := NewConfig()
	ctx := NewContext()
	ctx.APIURL = apiURL
	cfg.Contexts = map[string]*Context{"default": ctx}
	cfg.CurrentContext = "default"
	return cfg
}
