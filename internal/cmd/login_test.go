// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"testing"

	"github.com/StackEye-IO/stackeye-cli/internal/auth"
)

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
