// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"testing"
)

func TestValidateProbeURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https url",
			url:     "https://api.example.com/health",
			wantErr: false,
		},
		{
			name:    "valid http url",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "valid url with port",
			url:     "https://api.example.com:8443/health",
			wantErr: false,
		},
		{
			name:    "valid url with path and query",
			url:     "https://api.example.com/health?check=true",
			wantErr: false,
		},
		{
			name:    "missing scheme",
			url:     "example.com",
			wantErr: true,
		},
		{
			name:    "invalid scheme",
			url:     "ftp://files.example.com",
			wantErr: true,
		},
		{
			name:    "missing host",
			url:     "https://",
			wantErr: true,
		},
		{
			name:    "empty url",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProbeURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProbeURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCheckType(t *testing.T) {
	tests := []struct {
		name      string
		checkType string
		wantErr   bool
	}{
		{name: "http", checkType: "http", wantErr: false},
		{name: "ping", checkType: "ping", wantErr: false},
		{name: "tcp", checkType: "tcp", wantErr: false},
		{name: "dns_resolve", checkType: "dns_resolve", wantErr: false},
		{name: "invalid", checkType: "websocket", wantErr: true},
		{name: "empty", checkType: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCheckType(tt.checkType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCheckType(%q) error = %v, wantErr %v", tt.checkType, err, tt.wantErr)
			}
		})
	}
}

func TestValidateHTTPMethod(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr bool
	}{
		{name: "GET", method: "GET", wantErr: false},
		{name: "POST", method: "POST", wantErr: false},
		{name: "PUT", method: "PUT", wantErr: false},
		{name: "DELETE", method: "DELETE", wantErr: false},
		{name: "PATCH", method: "PATCH", wantErr: false},
		{name: "HEAD", method: "HEAD", wantErr: false},
		{name: "OPTIONS", method: "OPTIONS", wantErr: false},
		{name: "lowercase get", method: "get", wantErr: false},
		{name: "invalid", method: "CONNECT", wantErr: true},
		{name: "empty", method: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHTTPMethod(tt.method)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHTTPMethod(%q) error = %v, wantErr %v", tt.method, err, tt.wantErr)
			}
		})
	}
}

func TestValidateKeywordCheckType(t *testing.T) {
	tests := []struct {
		name      string
		checkType string
		wantErr   bool
	}{
		{name: "contains", checkType: "contains", wantErr: false},
		{name: "not_contains", checkType: "not_contains", wantErr: false},
		{name: "invalid", checkType: "matches", wantErr: true},
		{name: "empty", checkType: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKeywordCheckType(tt.checkType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateKeywordCheckType(%q) error = %v, wantErr %v", tt.checkType, err, tt.wantErr)
			}
		})
	}
}

func TestParseStatusCodes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{
			name:    "single code",
			input:   "200",
			want:    []int{200},
			wantErr: false,
		},
		{
			name:    "multiple codes",
			input:   "200,201,204",
			want:    []int{200, 201, 204},
			wantErr: false,
		},
		{
			name:    "codes with spaces",
			input:   "200, 201, 204",
			want:    []int{200, 201, 204},
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    []int{200},
			wantErr: false,
		},
		{
			name:    "invalid code",
			input:   "abc",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "code out of range low",
			input:   "50",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "code out of range high",
			input:   "600",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "valid 3xx codes",
			input:   "301,302,304",
			want:    []int{301, 302, 304},
			wantErr: false,
		},
		{
			name:    "valid 4xx and 5xx codes",
			input:   "400,404,500,503",
			want:    []int{400, 404, 500, 503},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStatusCodes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStatusCodes(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("parseStatusCodes(%q) = %v, want %v", tt.input, got, tt.want)
					return
				}
				for i, v := range got {
					if v != tt.want[i] {
						t.Errorf("parseStatusCodes(%q) = %v, want %v", tt.input, got, tt.want)
						return
					}
				}
			}
		})
	}
}

func TestBuildCreateProbeRequest(t *testing.T) {
	tests := []struct {
		name          string
		flags         *probeCreateFlags
		expectedCodes []int
		wantTimeoutMs int
	}{
		{
			name: "basic request",
			flags: &probeCreateFlags{
				name:            "Test Probe",
				url:             "https://example.com",
				checkType:       "http",
				method:          "GET",
				intervalSeconds: 60,
				timeoutSeconds:  10,
				followRedirects: true,
				maxRedirects:    10,
			},
			expectedCodes: []int{200},
			wantTimeoutMs: 10000,
		},
		{
			name: "timeout conversion",
			flags: &probeCreateFlags{
				name:            "Timeout Test",
				url:             "https://example.com",
				checkType:       "http",
				method:          "GET",
				intervalSeconds: 30,
				timeoutSeconds:  5,
				followRedirects: true,
				maxRedirects:    5,
			},
			expectedCodes: []int{200, 201},
			wantTimeoutMs: 5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildCreateProbeRequest(tt.flags, tt.expectedCodes)

			if req.Name != tt.flags.name {
				t.Errorf("Name = %q, want %q", req.Name, tt.flags.name)
			}
			if req.URL != tt.flags.url {
				t.Errorf("URL = %q, want %q", req.URL, tt.flags.url)
			}
			if req.TimeoutMs != tt.wantTimeoutMs {
				t.Errorf("TimeoutMs = %d, want %d", req.TimeoutMs, tt.wantTimeoutMs)
			}
			if req.IntervalSeconds != tt.flags.intervalSeconds {
				t.Errorf("IntervalSeconds = %d, want %d", req.IntervalSeconds, tt.flags.intervalSeconds)
			}
			if len(req.ExpectedStatusCodes) != len(tt.expectedCodes) {
				t.Errorf("ExpectedStatusCodes length = %d, want %d", len(req.ExpectedStatusCodes), len(tt.expectedCodes))
			}
		})
	}
}

func TestNewProbeCreateCmd(t *testing.T) {
	cmd := NewProbeCreateCmd()

	if cmd.Use != "create" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create")
	}

	// Check required flags exist
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("Expected --name flag to exist")
	}

	urlFlag := cmd.Flags().Lookup("url")
	if urlFlag == nil {
		t.Error("Expected --url flag to exist")
	}

	// Check optional flags exist
	optionalFlags := []string{
		"check-type", "method", "interval", "timeout", "regions",
		"headers", "body", "expected-status-codes", "follow-redirects",
		"max-redirects", "keyword-check", "keyword-check-type",
		"json-path-check", "json-path-expected", "ssl-check-enabled",
		"ssl-expiry-threshold-days",
	}

	for _, flagName := range optionalFlags {
		if cmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected --%s flag to exist", flagName)
		}
	}
}
