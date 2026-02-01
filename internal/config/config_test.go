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

// setupTestConfig creates a temp config file and sets up environment for testing.
// Returns cleanup function that must be called with defer.
func setupTestConfig(t *testing.T, content string) (configPath string, cleanup func()) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath = filepath.Join(tmpDir, "config.yaml")

	if content != "" {
		if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}
	}

	// Save original env vars
	origConfig := os.Getenv(EnvConfig)
	origContext := os.Getenv(EnvContext)
	origAPIKey := os.Getenv(EnvAPIKey)
	origAPIURL := os.Getenv(EnvAPIURL)

	// Set config path to our temp file
	os.Setenv(EnvConfig, configPath)

	cleanup = func() {
		os.Setenv(EnvConfig, origConfig)
		os.Setenv(EnvContext, origContext)
		os.Setenv(EnvAPIKey, origAPIKey)
		os.Setenv(EnvAPIURL, origAPIURL)
		ResetManager()
	}

	// Reset manager to ensure fresh state
	ResetManager()

	return configPath, cleanup
}

func TestManagerLoad(t *testing.T) {
	validConfig := `
version: 1
current_context: test
contexts:
  test:
    api_key: se_testkey123
    api_url: https://api.test.io
`

	t.Run("loads valid config file", func(t *testing.T) {
		_, cleanup := setupTestConfig(t, validConfig)
		defer cleanup()

		m := GetManager()
		cfg, err := m.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("Load() returned nil config")
		}
		if cfg.CurrentContext != "test" {
			t.Errorf("CurrentContext = %q, want %q", cfg.CurrentContext, "test")
		}
	})

	t.Run("returns cached config on second call", func(t *testing.T) {
		_, cleanup := setupTestConfig(t, validConfig)
		defer cleanup()

		m := GetManager()
		cfg1, err := m.Load()
		if err != nil {
			t.Fatalf("first Load() error = %v", err)
		}

		cfg2, err := m.Load()
		if err != nil {
			t.Fatalf("second Load() error = %v", err)
		}

		if cfg1 != cfg2 {
			t.Error("Load() should return cached config on second call")
		}
	})

	t.Run("returns empty config when file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentPath := filepath.Join(tmpDir, "nonexistent", "config.yaml")

		origConfig := os.Getenv(EnvConfig)
		os.Setenv(EnvConfig, nonExistentPath)
		defer os.Setenv(EnvConfig, origConfig)
		ResetManager()
		defer ResetManager()

		m := GetManager()
		cfg, err := m.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("Load() returned nil config")
		}
		// Empty config should have no current context
		if cfg.CurrentContext != "" {
			t.Errorf("CurrentContext = %q, want empty", cfg.CurrentContext)
		}
	})

	t.Run("applies STACKEYE_CONTEXT override", func(t *testing.T) {
		multiContextConfig := `
version: 1
current_context: default
contexts:
  default:
    api_key: se_defaultkey
  production:
    api_key: se_prodkey
`
		_, cleanup := setupTestConfig(t, multiContextConfig)
		defer cleanup()

		os.Setenv(EnvContext, "production")

		m := GetManager()
		cfg, err := m.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.CurrentContext != "production" {
			t.Errorf("CurrentContext = %q, want %q", cfg.CurrentContext, "production")
		}
	})

	t.Run("returns error for invalid context override", func(t *testing.T) {
		_, cleanup := setupTestConfig(t, validConfig)
		defer cleanup()

		os.Setenv(EnvContext, "nonexistent")

		m := GetManager()
		_, err := m.Load()
		if err == nil {
			t.Error("Load() should return error for invalid context")
		}
	})
}

func TestManagerGet(t *testing.T) {
	validConfig := `
version: 1
current_context: test
contexts:
  test:
    api_key: se_testkey123
`

	t.Run("returns nil before Load", func(t *testing.T) {
		_, cleanup := setupTestConfig(t, validConfig)
		defer cleanup()

		m := GetManager()
		cfg := m.Get()
		if cfg != nil {
			t.Error("Get() should return nil before Load()")
		}
	})

	t.Run("returns config after Load", func(t *testing.T) {
		_, cleanup := setupTestConfig(t, validConfig)
		defer cleanup()

		m := GetManager()
		_, err := m.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		cfg := m.Get()
		if cfg == nil {
			t.Fatal("Get() should return config after Load()")
		}
		if cfg.CurrentContext != "test" {
			t.Errorf("CurrentContext = %q, want %q", cfg.CurrentContext, "test")
		}
	})
}

func TestManagerReload(t *testing.T) {
	initialConfig := `
version: 1
current_context: initial
contexts:
  initial:
    api_key: se_initialkey
`

	updatedConfig := `
version: 1
current_context: updated
contexts:
  updated:
    api_key: se_updatedkey
`

	t.Run("reloads config from disk", func(t *testing.T) {
		configPath, cleanup := setupTestConfig(t, initialConfig)
		defer cleanup()

		m := GetManager()
		cfg1, err := m.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg1.CurrentContext != "initial" {
			t.Errorf("initial CurrentContext = %q, want %q", cfg1.CurrentContext, "initial")
		}

		// Update the config file
		if err := os.WriteFile(configPath, []byte(updatedConfig), 0600); err != nil {
			t.Fatalf("failed to update config: %v", err)
		}

		// Reload should pick up the changes
		cfg2, err := m.Reload()
		if err != nil {
			t.Fatalf("Reload() error = %v", err)
		}
		if cfg2.CurrentContext != "updated" {
			t.Errorf("reloaded CurrentContext = %q, want %q", cfg2.CurrentContext, "updated")
		}

		// Get should return the new config
		cfg3 := m.Get()
		if cfg3.CurrentContext != "updated" {
			t.Errorf("Get() after Reload CurrentContext = %q, want %q", cfg3.CurrentContext, "updated")
		}
	})

	t.Run("returns new config when file is deleted", func(t *testing.T) {
		configPath, cleanup := setupTestConfig(t, initialConfig)
		defer cleanup()

		m := GetManager()
		_, err := m.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Delete the config file
		os.Remove(configPath)

		// Reload should return empty config (file not found is not an error)
		cfg, err := m.Reload()
		if err != nil {
			t.Fatalf("Reload() error = %v", err)
		}
		if cfg.CurrentContext != "" {
			t.Errorf("CurrentContext = %q, want empty", cfg.CurrentContext)
		}
	})
}

func TestManagerSave(t *testing.T) {
	initialConfig := `
version: 1
current_context: test
contexts:
  test:
    api_key: se_testkey123
`

	t.Run("saves modified config to disk", func(t *testing.T) {
		configPath, cleanup := setupTestConfig(t, initialConfig)
		defer cleanup()

		m := GetManager()
		cfg, err := m.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Modify the config via the Contexts map directly (pointer-based)
		if cfg.Contexts == nil {
			t.Fatal("Contexts map is nil")
		}
		ctx := cfg.Contexts["test"]
		if ctx == nil {
			t.Fatal("test context is nil")
		}
		ctx.APIKey = "se_newkey456"

		// Save
		if err := m.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Verify file was saved with correct content
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}
		if !filepath.IsAbs(configPath) {
			t.Error("config path should be absolute")
		}
		if len(data) == 0 {
			t.Error("saved config file should not be empty")
		}
		// Check the new key is in the file
		if !contains(string(data), "se_newkey456") {
			t.Errorf("saved file does not contain new key, content:\n%s", string(data))
		}

		// Reload and verify
		cfg2, err := m.Reload()
		if err != nil {
			t.Fatalf("Reload() error = %v", err)
		}
		ctx2, _ := cfg2.GetCurrentContext()
		if ctx2.APIKey != "se_newkey456" {
			t.Errorf("APIKey = %q, want %q", ctx2.APIKey, "se_newkey456")
		}
	})

	t.Run("returns error when no config loaded", func(t *testing.T) {
		_, cleanup := setupTestConfig(t, "")
		defer cleanup()

		m := GetManager()
		// Don't call Load()

		err := m.Save()
		if err == nil {
			t.Error("Save() should return error when no config loaded")
		}
	})
}

func TestPackageLevelLoadAndSave(t *testing.T) {
	validConfig := `
version: 1
current_context: test
contexts:
  test:
    api_key: se_testkey123
`

	t.Run("Load uses default manager", func(t *testing.T) {
		_, cleanup := setupTestConfig(t, validConfig)
		defer cleanup()

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("Load() returned nil")
		}
		if cfg.CurrentContext != "test" {
			t.Errorf("CurrentContext = %q, want %q", cfg.CurrentContext, "test")
		}
	})

	t.Run("Save uses default manager", func(t *testing.T) {
		_, cleanup := setupTestConfig(t, validConfig)
		defer cleanup()

		_, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		err = Save()
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	})
}

func TestConfigDir(t *testing.T) {
	t.Run("returns non-empty directory path", func(t *testing.T) {
		dir := ConfigDir()
		if dir == "" {
			t.Error("ConfigDir() returned empty string")
		}
	})

	t.Run("is parent of ConfigPath", func(t *testing.T) {
		// Clear env override to test default behavior
		origEnv := os.Getenv(EnvConfig)
		os.Setenv(EnvConfig, "")
		defer os.Setenv(EnvConfig, origEnv)

		dir := ConfigDir()
		path := ConfigPath()

		if filepath.Dir(path) != dir {
			t.Errorf("ConfigDir() = %q, but ConfigPath() parent = %q", dir, filepath.Dir(path))
		}
	})
}

func TestResetManager(t *testing.T) {
	validConfig := `
version: 1
current_context: test
contexts:
  test:
    api_key: se_testkey123
`
	_, cleanup := setupTestConfig(t, validConfig)
	defer cleanup()

	// Load config
	m1 := GetManager()
	_, err := m1.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify config is loaded
	if m1.Get() == nil {
		t.Fatal("config should be loaded")
	}

	// Reset
	ResetManager()

	// Get manager again - should be a new instance
	m2 := GetManager()
	if m2.Get() != nil {
		t.Error("new manager should have nil config")
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLoadConfigWithInvalidYAML(t *testing.T) {
	invalidConfig := `
this is not: valid: yaml: content
  broken indentation
`
	_, cleanup := setupTestConfig(t, invalidConfig)
	defer cleanup()

	m := GetManager()
	_, err := m.Load()
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}
