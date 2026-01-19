package api

import (
	"errors"
	"testing"

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

	client, err := GetClientWithContext("nokey")
	if client != nil {
		t.Error("expected nil client when no API key")
	}
	if !errors.Is(err, ErrNoAPIKey) {
		t.Errorf("expected ErrNoAPIKey, got %v", err)
	}
}
