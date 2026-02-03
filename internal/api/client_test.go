package api

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

// TestGetClient_NoConfigGetter tests that GetClient returns an error when
// no config getter is set.
func TestGetClient_NoConfigGetter(t *testing.T) {
	// Reset config getter
	SetConfigGetter(nil)

	client, err := GetClient()
	if client != nil {
		t.Error("expected nil client when config getter is nil")
	}
	if !errors.Is(err, ErrConfigNotLoaded) {
		t.Errorf("expected ErrConfigNotLoaded, got %v", err)
	}
}

// TestGetClient_NilConfig tests that GetClient returns an error when
// the config getter returns nil.
func TestGetClient_NilConfig(t *testing.T) {
	SetConfigGetter(func() *config.Config {
		return nil
	})

	client, err := GetClient()
	if client != nil {
		t.Error("expected nil client when config is nil")
	}
	if !errors.Is(err, ErrConfigNotLoaded) {
		t.Errorf("expected ErrConfigNotLoaded, got %v", err)
	}
}

// TestGetClient_NoCurrentContext tests that GetClient returns an error when
// no current context is set.
func TestGetClient_NoCurrentContext(t *testing.T) {
	cfg := config.NewConfig()
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	client, err := GetClient()
	if client != nil {
		t.Error("expected nil client when no current context")
	}
	if !errors.Is(err, ErrNoCurrentContext) {
		t.Errorf("expected ErrNoCurrentContext, got %v", err)
	}
}

// TestGetClient_ContextNotFound tests that GetClient returns an error when
// the current context doesn't exist in the config.
func TestGetClient_ContextNotFound(t *testing.T) {
	cfg := config.NewConfig()
	cfg.CurrentContext = "nonexistent"
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	client, err := GetClient()
	if client != nil {
		t.Error("expected nil client when context not found")
	}
	if !errors.Is(err, ErrContextNotFound) {
		t.Errorf("expected ErrContextNotFound, got %v", err)
	}
}

// TestGetClient_NoAPIKey tests that GetClient returns an error when
// the context has no API key configured.
func TestGetClient_NoAPIKey(t *testing.T) {
	cfg := config.NewConfig()
	cfg.CurrentContext = "test-context"
	cfg.SetContext("test-context", &config.Context{
		APIURL: "https://api.example.com",
		// No API key
	})
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	client, err := GetClient()
	if client != nil {
		t.Error("expected nil client when no API key")
	}
	if !errors.Is(err, ErrNoAPIKey) {
		t.Errorf("expected ErrNoAPIKey, got %v", err)
	}
}

// TestGetClient_Success tests that GetClient returns a valid client when
// all configuration is correct.
func TestGetClient_Success(t *testing.T) {
	cfg := config.NewConfig()
	cfg.CurrentContext = "test-context"
	cfg.SetContext("test-context", &config.Context{
		APIURL: "https://api.example.com",
		APIKey: "se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	})
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	client, err := GetClient()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
	if client.BaseURL() != "https://api.example.com" {
		t.Errorf("expected base URL https://api.example.com, got %s", client.BaseURL())
	}
}

// TestGetClient_DefaultAPIURL tests that GetClient uses the default API URL
// when none is specified in the context.
func TestGetClient_DefaultAPIURL(t *testing.T) {
	cfg := config.NewConfig()
	cfg.CurrentContext = "test-context"
	cfg.SetContext("test-context", &config.Context{
		// No APIURL - should use default
		APIKey: "se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	})
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	client, err := GetClient()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
	if client.BaseURL() != config.DefaultAPIURL {
		t.Errorf("expected default API URL %s, got %s", config.DefaultAPIURL, client.BaseURL())
	}
}

// TestGetClientWithContext_Success tests that GetClientWithContext returns
// a valid client for a specific named context.
func TestGetClientWithContext_Success(t *testing.T) {
	cfg := config.NewConfig()
	cfg.CurrentContext = "default"
	cfg.SetContext("default", &config.Context{
		APIURL: "https://api.default.com",
		APIKey: "se_default000000000000000000000000000000000000000000000000000000000",
	})
	cfg.SetContext("other", &config.Context{
		APIURL: "https://api.other.com",
		APIKey: "se_other0000000000000000000000000000000000000000000000000000000000",
	})
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	// Get client for "other" context instead of current "default"
	client, err := GetClientWithContext("other")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
	if client.BaseURL() != "https://api.other.com" {
		t.Errorf("expected base URL https://api.other.com, got %s", client.BaseURL())
	}
}

// TestGetClientWithContext_NotFound tests that GetClientWithContext returns
// an error when the specified context doesn't exist.
func TestGetClientWithContext_NotFound(t *testing.T) {
	cfg := config.NewConfig()
	cfg.CurrentContext = "default"
	cfg.SetContext("default", &config.Context{
		APIKey: "se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	})
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	client, err := GetClientWithContext("nonexistent")
	if client != nil {
		t.Error("expected nil client when context not found")
	}
	if !errors.Is(err, ErrContextNotFound) {
		t.Errorf("expected ErrContextNotFound, got %v", err)
	}
}

// TestGetClientWithContext_NoAPIKey tests that GetClientWithContext returns
// an error when the specified context has no API key.
func TestGetClientWithContext_NoAPIKey(t *testing.T) {
	cfg := config.NewConfig()
	cfg.CurrentContext = "default"
	cfg.SetContext("default", &config.Context{
		APIKey: "se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	})
	cfg.SetContext("nokey", &config.Context{
		APIURL: "https://api.nokey.com",
		// No API key
	})
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	c, err := GetClientWithContext("nokey")
	if c != nil {
		t.Error("expected nil client when no API key")
	}
	if !errors.Is(err, ErrNoAPIKey) {
		t.Errorf("expected ErrNoAPIKey, got %v", err)
	}
}

// TestBuildClientOptions_TimeoutFromFlag tests that the --timeout flag
// value is passed through to the SDK client.
func TestBuildClientOptions_TimeoutFromFlag(t *testing.T) {
	// Set up config with default timeout
	cfg := config.NewConfig()
	cfg.Preferences.DefaultTimeout = 30
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	// Set timeout getter to simulate --timeout=45 flag
	SetTimeoutGetter(func() int { return 45 })
	defer SetTimeoutGetter(nil)

	opts := buildClientOptions()
	c := client.New("se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "", opts...)

	if c.Timeout() != 45*time.Second {
		t.Errorf("expected timeout 45s from flag, got %v", c.Timeout())
	}
}

// TestBuildClientOptions_TimeoutFromConfig tests that the config preference
// timeout is used when no flag is set.
func TestBuildClientOptions_TimeoutFromConfig(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Preferences.DefaultTimeout = 60
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	// No flag timeout set
	SetTimeoutGetter(func() int { return 0 })
	defer SetTimeoutGetter(nil)

	opts := buildClientOptions()
	c := client.New("se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "", opts...)

	if c.Timeout() != 60*time.Second {
		t.Errorf("expected timeout 60s from config, got %v", c.Timeout())
	}
}

// TestBuildClientOptions_TimeoutFlagOverridesConfig tests that the --timeout
// flag takes precedence over the config preference.
func TestBuildClientOptions_TimeoutFlagOverridesConfig(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Preferences.DefaultTimeout = 60
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	// Flag overrides config
	SetTimeoutGetter(func() int { return 10 })
	defer SetTimeoutGetter(nil)

	opts := buildClientOptions()
	c := client.New("se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "", opts...)

	if c.Timeout() != 10*time.Second {
		t.Errorf("expected timeout 10s from flag (overriding config 60s), got %v", c.Timeout())
	}
}

// TestBuildClientOptions_NoTimeoutUsesDefault tests that when neither flag
// nor config is set, the SDK default timeout (30s) is used.
func TestBuildClientOptions_NoTimeoutUsesDefault(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Preferences.DefaultTimeout = 0 // No config preference
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	SetTimeoutGetter(func() int { return 0 })
	defer SetTimeoutGetter(nil)

	opts := buildClientOptions()
	c := client.New("se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "", opts...)

	// SDK default is 30s
	if c.Timeout() != client.DefaultTimeout {
		t.Errorf("expected SDK default timeout %v, got %v", client.DefaultTimeout, c.Timeout())
	}
}

// captureStderr redirects os.Stderr to capture output during a function call.
func captureStderr(fn func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// TestWarnExcessiveTimeout_NoWarningBelowThreshold verifies no warning for reasonable timeouts.
func TestWarnExcessiveTimeout_NoWarningBelowThreshold(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
	}{
		{"1 second", 1},
		{"30 seconds (default)", 30},
		{"60 seconds", 60},
		{"300 seconds (exactly 5 min)", 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStderr(func() {
				warnExcessiveTimeout(tt.seconds)
			})
			if output != "" {
				t.Errorf("expected no warning for %d seconds, got: %s", tt.seconds, output)
			}
		})
	}
}

// TestWarnExcessiveTimeout_WarningAboveThreshold verifies warning for excessive timeouts.
func TestWarnExcessiveTimeout_WarningAboveThreshold(t *testing.T) {
	tests := []struct {
		name        string
		seconds     int
		expectedMin int
	}{
		{"301 seconds", 301, 5},
		{"600 seconds (10 min)", 600, 10},
		{"3000 seconds (typo for 30)", 3000, 50},
		{"3600 seconds (1 hour)", 3600, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStderr(func() {
				warnExcessiveTimeout(tt.seconds)
			})
			if output == "" {
				t.Errorf("expected warning for %d seconds, got none", tt.seconds)
				return
			}
			expectedSecondsStr := fmt.Sprintf("%d seconds", tt.seconds)
			if !strings.Contains(output, expectedSecondsStr) {
				t.Errorf("warning should mention %q, got: %s", expectedSecondsStr, output)
			}
			expectedMinStr := fmt.Sprintf("%d minutes", tt.expectedMin)
			if !strings.Contains(output, expectedMinStr) {
				t.Errorf("warning should mention %q, got: %s", expectedMinStr, output)
			}
			if !strings.Contains(output, "unusually high") {
				t.Errorf("warning should say 'unusually high', got: %s", output)
			}
		})
	}
}

// TestBuildClientOptions_WarnsOnExcessiveTimeoutFromFlag tests that buildClientOptions
// warns when the --timeout flag provides an excessive value.
func TestBuildClientOptions_WarnsOnExcessiveTimeoutFromFlag(t *testing.T) {
	cfg := config.NewConfig()
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	SetTimeoutGetter(func() int { return 3000 })
	defer SetTimeoutGetter(nil)

	output := captureStderr(func() {
		opts := buildClientOptions()
		c := client.New("se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "", opts...)
		if c.Timeout() != 3000*time.Second {
			t.Errorf("expected timeout 3000s, got %v", c.Timeout())
		}
	})

	if !strings.Contains(output, "3000 seconds") {
		t.Errorf("expected warning about 3000 seconds, got: %s", output)
	}
}

// TestBuildClientOptions_NoWarningForNormalTimeout tests that buildClientOptions
// does not warn for a normal timeout value.
func TestBuildClientOptions_NoWarningForNormalTimeout(t *testing.T) {
	cfg := config.NewConfig()
	SetConfigGetter(func() *config.Config {
		return cfg
	})

	SetTimeoutGetter(func() int { return 30 })
	defer SetTimeoutGetter(nil)

	output := captureStderr(func() {
		buildClientOptions()
	})

	if output != "" {
		t.Errorf("expected no warning for 30s timeout, got: %s", output)
	}
}
