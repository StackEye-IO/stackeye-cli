// Package auth provides browser-based authentication for the StackEye CLI.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBuildWebUIURL(t *testing.T) {
	tests := []struct {
		name        string
		apiURL      string
		callbackURL string
		want        string
		wantErr     bool
	}{
		{
			name:        "production API",
			apiURL:      "https://api.stackeye.io",
			callbackURL: "http://127.0.0.1:12345/callback",
			want:        "https://app.stackeye.io/cli-auth?callback=http%3A%2F%2F127.0.0.1%3A12345%2Fcallback",
			wantErr:     false,
		},
		{
			name:        "dev API",
			apiURL:      "https://api-dev.stackeye.io",
			callbackURL: "http://127.0.0.1:54321/callback",
			want:        "https://app-dev.stackeye.io/cli-auth?callback=http%3A%2F%2F127.0.0.1%3A54321%2Fcallback",
			wantErr:     false,
		},
		{
			name:        "staging API",
			apiURL:      "https://api-staging.stackeye.io",
			callbackURL: "http://127.0.0.1:8080/callback",
			want:        "https://app-staging.stackeye.io/cli-auth?callback=http%3A%2F%2F127.0.0.1%3A8080%2Fcallback",
			wantErr:     false,
		},
		{
			name:        "localhost API unchanged",
			apiURL:      "http://localhost:8080",
			callbackURL: "http://127.0.0.1:9000/callback",
			want:        "http://localhost:8080/cli-auth?callback=http%3A%2F%2F127.0.0.1%3A9000%2Fcallback",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildWebUIURL(tt.apiURL, tt.callbackURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWebUIURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BuildWebUIURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIURLToWebURL(t *testing.T) {
	tests := []struct {
		name    string
		apiURL  string
		want    string
		wantErr bool
	}{
		{
			name:    "production",
			apiURL:  "https://api.stackeye.io",
			want:    "https://app.stackeye.io",
			wantErr: false,
		},
		{
			name:    "dev environment",
			apiURL:  "https://api-dev.stackeye.io",
			want:    "https://app-dev.stackeye.io",
			wantErr: false,
		},
		{
			name:    "staging environment",
			apiURL:  "https://api-staging.stackeye.io",
			want:    "https://app-staging.stackeye.io",
			wantErr: false,
		},
		{
			name:    "localhost unchanged",
			apiURL:  "http://localhost:8080",
			want:    "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "custom domain unchanged",
			apiURL:  "https://mycompany.internal:9000",
			want:    "https://mycompany.internal:9000",
			wantErr: false,
		},
		{
			name:    "http scheme preserved",
			apiURL:  "http://api.stackeye.io",
			want:    "http://app.stackeye.io",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := APIURLToWebURL(tt.apiURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("APIURLToWebURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("APIURLToWebURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{
			name: "127.0.0.1",
			ip:   "127.0.0.1",
			want: true,
		},
		{
			name: "127.0.0.2",
			ip:   "127.0.0.2",
			want: true,
		},
		{
			name: "127.255.255.255",
			ip:   "127.255.255.255",
			want: true,
		},
		{
			name: "IPv6 localhost",
			ip:   "::1",
			want: true,
		},
		{
			name: "IPv6 localhost bracketed",
			ip:   "[::1]",
			want: true,
		},
		{
			name: "public IP",
			ip:   "8.8.8.8",
			want: false,
		},
		{
			name: "private IP",
			ip:   "192.168.1.1",
			want: false,
		},
		{
			name: "empty string",
			ip:   "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLocalhost(tt.ip); got != tt.want {
				t.Errorf("IsLocalhost(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestCallbackHandler(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		remoteAddr     string
		wantStatus     int
		wantAPIKey     string
		wantOrgID      string
		wantOrgName    string
		wantErr        bool
		wantErrContain string
	}{
		{
			name:        "successful callback",
			query:       "?api_key=se_abc123&org_id=org_1&org_name=Acme",
			remoteAddr:  "127.0.0.1:54321",
			wantStatus:  http.StatusOK,
			wantAPIKey:  "se_abc123",
			wantOrgID:   "org_1",
			wantOrgName: "Acme",
			wantErr:     false,
		},
		{
			name:       "missing api_key",
			query:      "?org_id=org_1",
			remoteAddr: "127.0.0.1:54321",
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:           "non-localhost rejected",
			query:          "?api_key=se_abc123",
			remoteAddr:     "8.8.8.8:54321",
			wantStatus:     http.StatusForbidden,
			wantErr:        true,
			wantErrContain: "non-localhost",
		},
		{
			name:        "api_key only",
			query:       "?api_key=se_xyz789",
			remoteAddr:  "127.0.0.1:12345",
			wantStatus:  http.StatusOK,
			wantAPIKey:  "se_xyz789",
			wantOrgID:   "",
			wantOrgName: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultCh := make(chan callbackResult, 1)
			handler := makeCallbackHandler(resultCh)

			req := httptest.NewRequest(http.MethodGet, "/callback"+tt.query, nil)
			req.RemoteAddr = tt.remoteAddr
			rec := httptest.NewRecorder()

			handler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("handler returned status %d, want %d", rec.Code, tt.wantStatus)
			}

			// Check result channel
			select {
			case result := <-resultCh:
				if tt.wantErr {
					if result.err == nil {
						t.Error("expected error, got nil")
					} else if tt.wantErrContain != "" && !containsString(result.err.Error(), tt.wantErrContain) {
						t.Errorf("error %q should contain %q", result.err.Error(), tt.wantErrContain)
					}
				} else {
					if result.err != nil {
						t.Errorf("unexpected error: %v", result.err)
					}
					if result.apiKey != tt.wantAPIKey {
						t.Errorf("apiKey = %q, want %q", result.apiKey, tt.wantAPIKey)
					}
					if result.orgID != tt.wantOrgID {
						t.Errorf("orgID = %q, want %q", result.orgID, tt.wantOrgID)
					}
					if result.orgName != tt.wantOrgName {
						t.Errorf("orgName = %q, want %q", result.orgName, tt.wantOrgName)
					}
				}
			default:
				t.Error("no result received from handler")
			}
		})
	}
}

func TestBrowserLoginWithContext_Timeout(t *testing.T) {
	// Test that BrowserLoginWithContext properly times out
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := BrowserLoginWithContext(ctx, Options{
		APIURL:  "https://api.stackeye.io",
		Timeout: 50 * time.Millisecond,
		OnBrowserOpen: func(url string) {
			// Don't actually open browser in tests
		},
		OnWaiting: func() {
			// Silent in tests
		},
	})

	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	// Should be a timeout or canceled error
	if err.Error() == "" {
		t.Error("expected error message")
	}
}

func TestBrowserLoginWithContext_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	_, err := BrowserLoginWithContext(ctx, Options{
		APIURL:  "https://api.stackeye.io",
		Timeout: 5 * time.Minute,
		OnBrowserOpen: func(url string) {
			// Don't actually open browser in tests
		},
		OnWaiting: func() {
			// Silent in tests
		},
	})

	if err == nil {
		t.Error("expected canceled error, got nil")
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{
			name:       "IPv4 with port",
			remoteAddr: "127.0.0.1:54321",
			want:       "127.0.0.1",
		},
		{
			name:       "IPv4 without port",
			remoteAddr: "127.0.0.1",
			want:       "127.0.0.1",
		},
		{
			name:       "IPv6 with port",
			remoteAddr: "[::1]:54321",
			want:       "::1",
		},
		{
			name:       "public IP with port",
			remoteAddr: "8.8.8.8:12345",
			want:       "8.8.8.8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractIP(tt.remoteAddr); got != tt.want {
				t.Errorf("extractIP(%q) = %q, want %q", tt.remoteAddr, got, tt.want)
			}
		})
	}
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
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

func TestOptions_Defaults(t *testing.T) {
	// Verify that BrowserLoginWithContext applies defaults correctly
	// We can't fully test this without mocking, but we can verify the option struct

	opts := Options{}

	if opts.APIURL != "" {
		t.Error("APIURL should be empty by default")
	}
	if opts.Timeout != 0 {
		t.Error("Timeout should be zero by default")
	}
	if opts.OnBrowserOpen != nil {
		t.Error("OnBrowserOpen should be nil by default")
	}
	if opts.OnWaiting != nil {
		t.Error("OnWaiting should be nil by default")
	}
}

func TestResult_Fields(t *testing.T) {
	result := Result{
		APIKey:  "se_test123",
		OrgID:   "org_456",
		OrgName: "Test Org",
	}

	if result.APIKey != "se_test123" {
		t.Errorf("APIKey = %q, want %q", result.APIKey, "se_test123")
	}
	if result.OrgID != "org_456" {
		t.Errorf("OrgID = %q, want %q", result.OrgID, "org_456")
	}
	if result.OrgName != "Test Org" {
		t.Errorf("OrgName = %q, want %q", result.OrgName, "Test Org")
	}
}

func TestErrors(t *testing.T) {
	// Verify error messages are meaningful
	errors := []struct {
		err  error
		want string
	}{
		{ErrTimeout, "auth: login timed out waiting for browser callback"},
		{ErrCanceled, "auth: login canceled"},
		{ErrInvalidAPIKey, "auth: received invalid API key format"},
		{ErrMissingAPIKey, "auth: callback missing api_key parameter"},
		{ErrForbidden, "auth: request from non-localhost IP rejected"},
	}

	for _, e := range errors {
		if e.err.Error() != e.want {
			t.Errorf("error = %q, want %q", e.err.Error(), e.want)
		}
	}
}

func TestSuccessHTML(t *testing.T) {
	// Verify successHTML contains expected content
	if successHTML == "" {
		t.Error("successHTML should not be empty")
	}

	expectedContents := []string{
		"<!DOCTYPE html>",
		"Login Successful",
		"close this window",
	}

	for _, expected := range expectedContents {
		if !containsString(successHTML, expected) {
			t.Errorf("successHTML should contain %q", expected)
		}
	}
}

func TestSetDebug(t *testing.T) {
	// Initially, debug should be disabled
	SetDebug(false)

	// Enable debug
	SetDebug(true)
	if debugLogger == nil {
		t.Error("expected debugLogger to be non-nil after SetDebug(true)")
	}

	// Disable debug
	SetDebug(false)
	if debugLogger != nil {
		t.Error("expected debugLogger to be nil after SetDebug(false)")
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "normal API key",
			key:  "se_abc123def456ghi789",
			want: "se_a...i789",
		},
		{
			name: "short key",
			key:  "short",
			want: "***",
		},
		{
			name: "exactly 8 chars",
			key:  "12345678",
			want: "***",
		},
		{
			name: "9 chars",
			key:  "123456789",
			want: "1234...6789",
		},
		{
			name: "empty key",
			key:  "",
			want: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maskAPIKey(tt.key); got != tt.want {
				t.Errorf("maskAPIKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// Example shows basic usage of the auth package.
func Example() {
	result, err := BrowserLogin(Options{
		APIURL:  "https://api.stackeye.io",
		Timeout: 5 * time.Minute,
		OnBrowserOpen: func(url string) {
			fmt.Printf("Please visit: %s\n", url)
		},
		OnWaiting: func() {
			fmt.Println("Waiting for authentication...")
		},
	})
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}

	fmt.Printf("Logged in to org: %s\n", result.OrgName)
}
