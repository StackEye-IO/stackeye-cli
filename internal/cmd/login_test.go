// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIURLToWebURL(t *testing.T) {
	tests := []struct {
		name    string
		apiURL  string
		want    string
		wantErr bool
	}{
		{
			name:    "production URL",
			apiURL:  "https://api.stackeye.io",
			want:    "https://app.stackeye.io",
			wantErr: false,
		},
		{
			name:    "dev URL",
			apiURL:  "https://api.dev.stackeye.io",
			want:    "https://app.dev.stackeye.io",
			wantErr: false,
		},
		{
			name:    "staging URL",
			apiURL:  "https://api.stg.stackeye.io",
			want:    "https://app.stg.stackeye.io",
			wantErr: false,
		},
		{
			name:    "custom domain without api prefix",
			apiURL:  "https://custom.example.com",
			want:    "https://custom.example.com",
			wantErr: false,
		},
		{
			name:    "localhost URL",
			apiURL:  "http://localhost:8080",
			want:    "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "with trailing path",
			apiURL:  "https://api.stackeye.io/v1",
			want:    "https://app.stackeye.io",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := apiURLToWebURL(tt.apiURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("apiURLToWebURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("apiURLToWebURL() = %v, want %v", got, tt.want)
			}
		})
	}
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

func TestBuildWebUIURL(t *testing.T) {
	tests := []struct {
		name        string
		apiURL      string
		callbackURL string
		want        string
		wantErr     bool
	}{
		{
			name:        "production with callback",
			apiURL:      "https://api.stackeye.io",
			callbackURL: "http://127.0.0.1:12345/callback",
			want:        "https://app.stackeye.io/cli-auth?callback=http%3A%2F%2F127.0.0.1%3A12345%2Fcallback",
			wantErr:     false,
		},
		{
			name:        "dev environment",
			apiURL:      "https://api.dev.stackeye.io",
			callbackURL: "http://127.0.0.1:54321/callback",
			want:        "https://app.dev.stackeye.io/cli-auth?callback=http%3A%2F%2F127.0.0.1%3A54321%2Fcallback",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildWebUIURL(tt.apiURL, tt.callbackURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildWebUIURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildWebUIURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCallbackHandler_Success(t *testing.T) {
	// Create a channel to receive the result
	resultCh := make(chan loginResult, 1)

	// Create the handler
	handler := makeCallbackHandler(resultCh)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/callback?api_key=se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef&org_id=org_123&org_name=Acme%20Corp", nil)
	// Set RemoteAddr to localhost
	req.RemoteAddr = "127.0.0.1:54321"

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response status
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Check the result was sent
	select {
	case result := <-resultCh:
		if result.err != nil {
			t.Errorf("unexpected error in result: %v", result.err)
		}
		if result.apiKey != "se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" {
			t.Errorf("unexpected API key: got %v", result.apiKey)
		}
		if result.orgID != "org_123" {
			t.Errorf("unexpected org ID: got %v want org_123", result.orgID)
		}
		if result.orgName != "Acme Corp" {
			t.Errorf("unexpected org name: got %v want Acme Corp", result.orgName)
		}
	default:
		t.Error("no result received on channel")
	}
}

func TestCallbackHandler_MissingKey(t *testing.T) {
	// Create a channel to receive the result
	resultCh := make(chan loginResult, 1)

	// Create the handler
	handler := makeCallbackHandler(resultCh)

	// Create a test request without api_key
	req := httptest.NewRequest(http.MethodGet, "/callback", nil)
	req.RemoteAddr = "127.0.0.1:54321"

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response status
	if rr.Code != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusBadRequest)
	}

	// Check the result was sent with error
	select {
	case result := <-resultCh:
		if result.err == nil {
			t.Error("expected error in result, got nil")
		}
	default:
		t.Error("no result received on channel")
	}
}

func TestCallbackHandler_NonLocalhost(t *testing.T) {
	// Create a channel to receive the result
	resultCh := make(chan loginResult, 1)

	// Create the handler
	handler := makeCallbackHandler(resultCh)

	// Create a test request from non-localhost IP
	req := httptest.NewRequest(http.MethodGet, "/callback?api_key=se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", nil)
	req.RemoteAddr = "192.168.1.100:54321"

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response status - should be Forbidden
	if rr.Code != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusForbidden)
	}

	// Check the result was sent with error
	select {
	case result := <-resultCh:
		if result.err == nil {
			t.Error("expected error in result, got nil")
		}
	default:
		t.Error("no result received on channel")
	}
}

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"127.0.1.1", true},
		{"::1", true},
		{"[::1]", true},
		{"192.168.1.1", false},
		{"10.0.0.1", false},
		{"8.8.8.8", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			if got := isLocalhost(tt.ip); got != tt.want {
				t.Errorf("isLocalhost(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		remoteAddr string
		want       string
	}{
		{"127.0.0.1:54321", "127.0.0.1"},
		{"192.168.1.1:8080", "192.168.1.1"},
		{"[::1]:54321", "::1"},
		{"127.0.0.1", "127.0.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.remoteAddr, func(t *testing.T) {
			if got := extractIP(tt.remoteAddr); got != tt.want {
				t.Errorf("extractIP(%q) = %v, want %v", tt.remoteAddr, got, tt.want)
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

func TestExtractOrgFromEmail(t *testing.T) {
	tests := []struct {
		email string
		want  string
	}{
		{"user@acme.com", "acme"},
		{"user@company.co.uk", "company"},
		{"user@example.org", "example"},
		{"invalid-email", "default"},
		{"", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := extractOrgFromEmail(tt.email); got != tt.want {
				t.Errorf("extractOrgFromEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestCallbackHandler_IPv6Localhost(t *testing.T) {
	// Create a channel to receive the result
	resultCh := make(chan loginResult, 1)

	// Create the handler
	handler := makeCallbackHandler(resultCh)

	// Create a test request with IPv6 localhost
	req := httptest.NewRequest(http.MethodGet, "/callback?api_key=se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", nil)
	req.RemoteAddr = "[::1]:54321"

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response status - should succeed from IPv6 localhost
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Check the result was sent
	select {
	case result := <-resultCh:
		if result.err != nil {
			t.Errorf("unexpected error in result: %v", result.err)
		}
	default:
		t.Error("no result received on channel")
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
	} else if flag.DefValue != defaultAPIURL {
		t.Errorf("unexpected default for --api-url: got %v want %v", flag.DefValue, defaultAPIURL)
	}
}

