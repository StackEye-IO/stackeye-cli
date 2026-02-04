// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-cli/internal/auth"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

// mockAuthenticator is a test fake that implements the Authenticator interface.
type mockAuthenticator struct {
	result *auth.Result
	err    error
	called bool
	opts   auth.Options
}

func (m *mockAuthenticator) Login(opts auth.Options) (*auth.Result, error) {
	m.called = true
	m.opts = opts
	return m.result, m.err
}

// setupTestConfigDir sets XDG_CONFIG_HOME to a temp directory so config.Load()
// and config.Save() operate on an isolated config file. Returns a cleanup function.
func setupTestConfigDir(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
}

// newMockVerifyServer creates an httptest.Server that responds to GET /v1/cli-auth/verify
// with the provided CLIVerifyResponse (or an error status).
func newMockVerifyServer(t *testing.T, resp *client.CLIVerifyResponse, statusCode int) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/cli-auth/verify" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if resp != nil {
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestGenerateContextName(t *testing.T) {
	tests := []struct {
		name    string
		orgName string
		apiURL  string
		want    string
	}{
		{
			name:    "simple org name production",
			orgName: "Acme Corp",
			apiURL:  "https://api.stackeye.io",
			want:    "acme-corp",
		},
		{
			name:    "org name with dev environment",
			orgName: "Acme Corp",
			apiURL:  "https://api.dev.stackeye.io",
			want:    "acme-corp-dev",
		},
		{
			name:    "org name with staging environment",
			orgName: "Acme Corp",
			apiURL:  "https://api.stg.stackeye.io",
			want:    "acme-corp-stg",
		},
		{
			name:    "org name with special characters",
			orgName: "Acme's Corp & Co.",
			apiURL:  "https://api.stackeye.io",
			want:    "acmes-corp-co",
		},
		{
			name:    "org name with underscores",
			orgName: "acme_corp_inc",
			apiURL:  "https://api.stackeye.io",
			want:    "acme-corp-inc",
		},
		{
			name:    "empty org name",
			orgName: "",
			apiURL:  "https://api.stackeye.io",
			want:    "default",
		},
		{
			name:    "org name with only special chars",
			orgName: "!@#$%",
			apiURL:  "https://api.stackeye.io",
			want:    "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateContextName(tt.orgName, tt.apiURL)
			if got != tt.want {
				t.Errorf("generateContextName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeContextName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Acme Corp", "acme-corp"},
		{"UPPERCASE", "uppercase"},
		{"with_underscore", "with-underscore"},
		{"multiple   spaces", "multiple-spaces"},
		{"special!@#chars", "specialchars"},
		{"---leading-trailing---", "leading-trailing"},
		{"", "default"},
		{"   ", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeContextName(tt.name); got != tt.want {
				t.Errorf("sanitizeContextName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestExtractEnvironment(t *testing.T) {
	tests := []struct {
		apiURL string
		want   string
	}{
		{"https://api.stackeye.io", ""},
		{"https://api.dev.stackeye.io", "dev"},
		{"https://api.stg.stackeye.io", "stg"},
		{"https://api.staging.stackeye.io", "stg"},
		{"http://localhost:8080", ""},
		{"invalid-url", ""},
	}

	for _, tt := range tests {
		t.Run(tt.apiURL, func(t *testing.T) {
			if got := extractEnvironment(tt.apiURL); got != tt.want {
				t.Errorf("extractEnvironment(%q) = %v, want %v", tt.apiURL, got, tt.want)
			}
		})
	}
}

func TestNewLoginCmd(t *testing.T) {
	cmd := NewLoginCmd()

	if cmd == nil {
		t.Fatal("NewLoginCmd() returned nil")
	}

	if cmd.Use != "login" {
		t.Errorf("unexpected command Use: got %v want login", cmd.Use)
	}

	// Check that --api-url flag exists
	flag := cmd.Flags().Lookup("api-url")
	if flag == nil {
		t.Error("expected --api-url flag to exist")
	} else if flag.DefValue != auth.DefaultAPIURL {
		t.Errorf("unexpected default for --api-url: got %v want %v", flag.DefValue, auth.DefaultAPIURL)
	}
}

func TestCompleteLogin_Success(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	server := newMockVerifyServer(t, &client.CLIVerifyResponse{
		Valid:            true,
		OrganizationID:   "org-123",
		OrganizationName: "Test Org",
		AuthType:         "api_key",
	}, http.StatusOK)

	apiKey := "se_" + "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	err := completeLogin(server.URL, apiKey, "org-123", "Test Org", false)
	if err != nil {
		t.Fatalf("completeLogin() unexpected error: %v", err)
	}

	// Verify config was saved correctly
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}

	if cfg.CurrentContext != "test-org" {
		t.Errorf("CurrentContext = %q, want %q", cfg.CurrentContext, "test-org")
	}

	ctx, err := cfg.GetContext("test-org")
	if err != nil {
		t.Fatalf("GetContext(\"test-org\") failed: %v", err)
	}
	if ctx.APIKey != apiKey {
		t.Errorf("APIKey = %q, want %q", ctx.APIKey, apiKey)
	}
	if ctx.APIURL != server.URL {
		t.Errorf("APIURL = %q, want %q", ctx.APIURL, server.URL)
	}
	if ctx.OrganizationID != "org-123" {
		t.Errorf("OrganizationID = %q, want %q", ctx.OrganizationID, "org-123")
	}
	if ctx.OrganizationName != "Test Org" {
		t.Errorf("OrganizationName = %q, want %q", ctx.OrganizationName, "Test Org")
	}
}

func TestCompleteLogin_VerifyFailure(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	server := newMockVerifyServer(t, nil, http.StatusUnauthorized)

	apiKey := "se_" + "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	err := completeLogin(server.URL, apiKey, "org-123", "Test Org", false)
	if err == nil {
		t.Fatal("completeLogin() expected error for 401 response, got nil")
	}
}

func TestCompleteLogin_ContextNameDeduplication(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	server := newMockVerifyServer(t, &client.CLIVerifyResponse{
		Valid:            true,
		OrganizationID:   "org-123",
		OrganizationName: "Acme Corp",
		AuthType:         "api_key",
	}, http.StatusOK)

	apiKey := "se_" + "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	// Pre-populate config with an existing "acme-corp" context
	cfg := config.NewConfig()
	cfg.SetContext("acme-corp", &config.Context{
		APIURL:           "https://api.stackeye.io",
		APIKey:           "se_b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2",
		OrganizationID:   "org-old",
		OrganizationName: "Acme Corp",
	})
	cfg.CurrentContext = "acme-corp"
	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to pre-populate config: %v", err)
	}

	// Login again with same org name - should get "acme-corp-2"
	err := completeLogin(server.URL, apiKey, "org-123", "Acme Corp", false)
	if err != nil {
		t.Fatalf("completeLogin() unexpected error: %v", err)
	}

	cfg, err = config.Load()
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}

	if cfg.CurrentContext != "acme-corp-2" {
		t.Errorf("CurrentContext = %q, want %q", cfg.CurrentContext, "acme-corp-2")
	}

	// Both contexts should exist
	if _, err := cfg.GetContext("acme-corp"); err != nil {
		t.Error("original context 'acme-corp' should still exist")
	}
	if _, err := cfg.GetContext("acme-corp-2"); err != nil {
		t.Error("new context 'acme-corp-2' should exist")
	}
}

func TestCompleteLogin_OrgNameFallback(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	server := newMockVerifyServer(t, &client.CLIVerifyResponse{
		Valid:            true,
		OrganizationID:   "org-456",
		OrganizationName: "From Verify",
		AuthType:         "api_key",
	}, http.StatusOK)

	apiKey := "se_" + "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	// Pass empty org name and org ID - should fall back to verify response values
	err := completeLogin(server.URL, apiKey, "", "", false)
	if err != nil {
		t.Fatalf("completeLogin() unexpected error: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}

	ctx, err := cfg.GetContext(cfg.CurrentContext)
	if err != nil {
		t.Fatalf("GetContext(%q) failed: %v", cfg.CurrentContext, err)
	}

	if ctx.OrganizationName != "From Verify" {
		t.Errorf("OrganizationName = %q, want %q", ctx.OrganizationName, "From Verify")
	}
	if ctx.OrganizationID != "org-456" {
		t.Errorf("OrganizationID = %q, want %q", ctx.OrganizationID, "org-456")
	}
}

func TestCompleteLogin_DebugMode(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	server := newMockVerifyServer(t, &client.CLIVerifyResponse{
		Valid:            true,
		OrganizationID:   "org-789",
		OrganizationName: "Debug Org",
		AuthType:         "api_key",
	}, http.StatusOK)

	apiKey := "se_" + "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	// Should not panic or error in debug mode
	err := completeLogin(server.URL, apiKey, "org-789", "Debug Org", true)
	if err != nil {
		t.Fatalf("completeLogin() with debug=true unexpected error: %v", err)
	}
}

func TestCheckExistingAuth_NoConfig(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	// No config file exists - should return nil (proceed with login)
	err := checkExistingAuth("https://api.stackeye.io")
	if err != nil {
		t.Fatalf("checkExistingAuth() unexpected error: %v", err)
	}
}

func TestCheckExistingAuth_NoInputMode(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)
	noInput = true

	// Create config with existing auth for this URL
	cfg := config.NewConfig()
	cfg.SetContext("existing", &config.Context{
		APIURL: "https://api.stackeye.io",
		APIKey: "se_b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2",
	})
	cfg.CurrentContext = "existing"
	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to pre-populate config: %v", err)
	}

	err := checkExistingAuth("https://api.stackeye.io")
	if err == nil {
		t.Fatal("checkExistingAuth() expected error in no-input mode with existing auth, got nil")
	}
}

func TestCheckExistingAuth_DifferentURL(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	// Create config with auth for a different URL
	cfg := config.NewConfig()
	cfg.SetContext("other", &config.Context{
		APIURL: "https://api.dev.stackeye.io",
		APIKey: "se_b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2",
	})
	cfg.CurrentContext = "other"
	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to pre-populate config: %v", err)
	}

	// Checking a different URL - should return nil (no existing auth for this URL)
	err := checkExistingAuth("https://api.stackeye.io")
	if err != nil {
		t.Fatalf("checkExistingAuth() unexpected error: %v", err)
	}
}

func TestRunLogin_Success(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	// Set up mock verify server
	server := newMockVerifyServer(t, &client.CLIVerifyResponse{
		Valid:            true,
		OrganizationID:   "org-mock-1",
		OrganizationName: "Mock Org",
		AuthType:         "api_key",
	}, http.StatusOK)

	apiKey := "se_" + "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	// Inject mock authenticator that returns a successful result
	mock := &mockAuthenticator{
		result: &auth.Result{
			APIKey:  apiKey,
			OrgID:   "org-mock-1",
			OrgName: "Mock Org",
		},
	}
	defaultAuthenticator = mock

	flags := &loginFlags{
		apiURL: server.URL,
		debug:  false,
	}

	err := runLogin(flags)
	if err != nil {
		t.Fatalf("runLogin() unexpected error: %v", err)
	}

	// Verify the authenticator was called
	if !mock.called {
		t.Error("expected authenticator.Login to be called")
	}

	// Verify the authenticator received the correct API URL
	if mock.opts.APIURL != server.URL {
		t.Errorf("authenticator received APIURL = %q, want %q", mock.opts.APIURL, server.URL)
	}

	// Verify config was saved correctly
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}

	if cfg.CurrentContext != "mock-org" {
		t.Errorf("CurrentContext = %q, want %q", cfg.CurrentContext, "mock-org")
	}

	ctx, err := cfg.GetContext("mock-org")
	if err != nil {
		t.Fatalf("GetContext(\"mock-org\") failed: %v", err)
	}
	if ctx.APIKey != apiKey {
		t.Errorf("APIKey = %q, want %q", ctx.APIKey, apiKey)
	}
	if ctx.OrganizationID != "org-mock-1" {
		t.Errorf("OrganizationID = %q, want %q", ctx.OrganizationID, "org-mock-1")
	}
}

func TestRunLogin_AuthenticatorError(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	// Inject mock authenticator that returns a timeout error
	mock := &mockAuthenticator{
		err: fmt.Errorf("connection refused: dial tcp 127.0.0.1:0: connect: connection refused"),
	}
	defaultAuthenticator = mock

	flags := &loginFlags{
		apiURL: "https://api.unreachable.example.com",
		debug:  false,
	}

	err := runLogin(flags)
	if err == nil {
		t.Fatal("runLogin() expected error for authenticator failure, got nil")
	}

	// Verify the error is wrapped with "login failed:"
	want := "login failed:"
	if got := err.Error(); !strings.Contains(got, want) {
		t.Errorf("error = %q, want to contain %q", got, want)
	}

	if !mock.called {
		t.Error("expected authenticator.Login to be called")
	}
}

func TestRunLogin_DefaultAPIURL(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	// Inject mock authenticator to capture the options
	mock := &mockAuthenticator{
		err: fmt.Errorf("intentional error to stop flow"),
	}
	defaultAuthenticator = mock

	// Pass empty apiURL - should default to auth.DefaultAPIURL
	flags := &loginFlags{
		apiURL: "",
		debug:  false,
	}

	_ = runLogin(flags)

	if !mock.called {
		t.Fatal("expected authenticator.Login to be called")
	}

	if mock.opts.APIURL != auth.DefaultAPIURL {
		t.Errorf("authenticator received APIURL = %q, want %q", mock.opts.APIURL, auth.DefaultAPIURL)
	}
}

func TestRunLogin_DebugMode(t *testing.T) {
	resetGlobalState()
	setupTestConfigDir(t)

	// Inject mock authenticator
	mock := &mockAuthenticator{
		err: fmt.Errorf("intentional error"),
	}
	defaultAuthenticator = mock

	flags := &loginFlags{
		apiURL: "https://api.test.example.com",
		debug:  true,
	}

	// Should not panic in debug mode even if authenticator fails
	_ = runLogin(flags)

	if !mock.called {
		t.Error("expected authenticator.Login to be called")
	}
}

