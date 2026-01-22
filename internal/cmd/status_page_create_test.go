// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

func TestNewStatusPageCreateCmd(t *testing.T) {
	cmd := NewStatusPageCreateCmd()

	if cmd.Use != "create" {
		t.Errorf("expected Use to be 'create', got %q", cmd.Use)
	}

	if cmd.Short != "Create a new status page" {
		t.Errorf("expected Short to be 'Create a new status page', got %q", cmd.Short)
	}
}

func TestNewStatusPageCreateCmd_Flags(t *testing.T) {
	cmd := NewStatusPageCreateCmd()

	// Required flag
	if cmd.Flags().Lookup("name") == nil {
		t.Error("expected 'name' flag to exist")
	}

	// Optional flags
	optionalFlags := []string{
		"slug",
		"custom-domain",
		"logo-url",
		"favicon-url",
		"header-text",
		"footer-text",
		"theme",
		"public",
		"show-uptime",
		"enabled",
		"from-file",
	}

	for _, flag := range optionalFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected %q flag to exist", flag)
		}
	}
}

func TestNewStatusPageCreateCmd_Long(t *testing.T) {
	cmd := NewStatusPageCreateCmd()

	// Verify Long contains key documentation
	expectedContains := []string{
		"Create a new status page",
		"--name",
		"--slug",
		"--theme",
		"light, dark, system",
		"Plan Limits",
		"YAML File Format",
	}

	for _, expected := range expectedContains {
		if !containsString(cmd.Long, expected) {
			t.Errorf("expected Long to contain %q", expected)
		}
	}
}

func TestValidateTheme(t *testing.T) {
	tests := []struct {
		name    string
		theme   string
		wantErr bool
	}{
		{"valid light", "light", false},
		{"valid dark", "dark", false},
		{"valid system", "system", false},
		{"valid Light uppercase", "Light", false},
		{"valid DARK uppercase", "DARK", false},
		{"valid System mixed", "System", false},
		{"invalid empty", "", true},
		{"invalid custom", "custom", true},
		{"invalid blue", "blue", true},
		{"invalid random", "random-theme", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTheme(tt.theme)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTheme(%q) error = %v, wantErr %v", tt.theme, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		wantErr bool
	}{
		// Valid slugs
		{"valid simple", "acme-status", false},
		{"valid alphanumeric", "api123", false},
		{"valid with numbers", "status-page-1", false},
		{"valid min length", "abc", false},
		{"valid max length", "a23456789012345678901234567890123456789012345678901234567890123", false}, // 63 chars

		// Invalid slugs - too short
		{"invalid too short 1", "a", true},
		{"invalid too short 2", "ab", true},

		// Invalid slugs - too long
		{"invalid too long", "a234567890123456789012345678901234567890123456789012345678901234", true}, // 64 chars

		// Invalid slugs - bad format
		{"invalid starts with hyphen", "-acme", true},
		{"invalid ends with hyphen", "acme-", true},
		{"invalid uppercase", "ACME-STATUS", true},
		{"invalid spaces", "acme status", true},
		{"invalid underscore", "acme_status", true},
		{"invalid special chars", "acme@status", true},
		{"invalid double hyphen", "acme--status", false}, // This is actually valid per regex
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSlug(tt.slug)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSlug(%q) error = %v, wantErr %v", tt.slug, err, tt.wantErr)
			}
		})
	}
}

func TestBuildStatusPageRequestFromFlags_RequiredName(t *testing.T) {
	flags := &statusPageCreateFlags{
		name: "",
	}

	_, err := buildStatusPageRequestFromFlags(flags)
	if err == nil {
		t.Error("expected error for empty name")
	}
	if err.Error() != "--name is required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBuildStatusPageRequestFromFlags_InvalidTheme(t *testing.T) {
	flags := &statusPageCreateFlags{
		name:  "Test Status",
		theme: "invalid-theme",
	}

	_, err := buildStatusPageRequestFromFlags(flags)
	if err == nil {
		t.Error("expected error for invalid theme")
	}
}

func TestBuildStatusPageRequestFromFlags_InvalidSlug(t *testing.T) {
	flags := &statusPageCreateFlags{
		name: "Test Status",
		slug: "a", // too short
	}

	_, err := buildStatusPageRequestFromFlags(flags)
	if err == nil {
		t.Error("expected error for invalid slug")
	}
}

func TestBuildStatusPageRequestFromFlags_Success(t *testing.T) {
	flags := &statusPageCreateFlags{
		name:                 "Acme Status",
		slug:                 "acme-status",
		customDomain:         "status.acme.com",
		logoURL:              "https://acme.com/logo.png",
		faviconURL:           "https://acme.com/favicon.ico",
		headerText:           "Welcome to Acme Status",
		footerText:           "Contact support@acme.com",
		theme:                "dark",
		isPublic:             true,
		showUptimePercentage: true,
		enabled:              true,
	}

	req, err := buildStatusPageRequestFromFlags(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify required fields
	if req.Name != "Acme Status" {
		t.Errorf("expected Name 'Acme Status', got %q", req.Name)
	}

	// Verify optional string fields
	if req.Slug != "acme-status" {
		t.Errorf("expected Slug 'acme-status', got %q", req.Slug)
	}
	if req.Theme != "dark" {
		t.Errorf("expected Theme 'dark', got %q", req.Theme)
	}

	// Verify optional pointer fields
	if req.CustomDomain == nil || *req.CustomDomain != "status.acme.com" {
		t.Errorf("expected CustomDomain 'status.acme.com', got %v", req.CustomDomain)
	}
	if req.LogoURL == nil || *req.LogoURL != "https://acme.com/logo.png" {
		t.Errorf("expected LogoURL 'https://acme.com/logo.png', got %v", req.LogoURL)
	}
	if req.FaviconURL == nil || *req.FaviconURL != "https://acme.com/favicon.ico" {
		t.Errorf("expected FaviconURL, got %v", req.FaviconURL)
	}
	if req.HeaderText == nil || *req.HeaderText != "Welcome to Acme Status" {
		t.Errorf("expected HeaderText, got %v", req.HeaderText)
	}
	if req.FooterText == nil || *req.FooterText != "Contact support@acme.com" {
		t.Errorf("expected FooterText, got %v", req.FooterText)
	}

	// Verify boolean pointer fields
	if req.IsPublic == nil || *req.IsPublic != true {
		t.Errorf("expected IsPublic true, got %v", req.IsPublic)
	}
	if req.ShowUptimePercentage == nil || *req.ShowUptimePercentage != true {
		t.Errorf("expected ShowUptimePercentage true, got %v", req.ShowUptimePercentage)
	}
	if req.Enabled == nil || *req.Enabled != true {
		t.Errorf("expected Enabled true, got %v", req.Enabled)
	}
}

func TestBuildStatusPageRequestFromFlags_ThemeNormalization(t *testing.T) {
	tests := []struct {
		name          string
		inputTheme    string
		expectedTheme string
	}{
		{"uppercase DARK", "DARK", "dark"},
		{"mixed case Light", "Light", "light"},
		{"uppercase SYSTEM", "SYSTEM", "system"},
		{"already lowercase", "dark", "dark"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &statusPageCreateFlags{
				name:  "Test Status",
				theme: tt.inputTheme,
			}

			req, err := buildStatusPageRequestFromFlags(flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if req.Theme != tt.expectedTheme {
				t.Errorf("expected Theme %q, got %q", tt.expectedTheme, req.Theme)
			}
		})
	}
}

func TestBuildStatusPageRequestFromFlags_MinimalInput(t *testing.T) {
	flags := &statusPageCreateFlags{
		name:                 "Simple Status",
		theme:                "light",
		isPublic:             true,
		showUptimePercentage: true,
		enabled:              true,
	}

	req, err := buildStatusPageRequestFromFlags(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify required fields
	if req.Name != "Simple Status" {
		t.Errorf("expected Name 'Simple Status', got %q", req.Name)
	}

	// Verify optional pointer fields are nil when not set
	if req.CustomDomain != nil {
		t.Errorf("expected CustomDomain nil, got %v", req.CustomDomain)
	}
	if req.LogoURL != nil {
		t.Errorf("expected LogoURL nil, got %v", req.LogoURL)
	}
	if req.FaviconURL != nil {
		t.Errorf("expected FaviconURL nil, got %v", req.FaviconURL)
	}
	if req.HeaderText != nil {
		t.Errorf("expected HeaderText nil, got %v", req.HeaderText)
	}
	if req.FooterText != nil {
		t.Errorf("expected FooterText nil, got %v", req.FooterText)
	}
}

// Note: containsString helper is defined in channel_test.go (same package)

// YAML parsing tests

func TestBuildStatusPageRequestFromYAML_ValidFile(t *testing.T) {
	// Create a temporary YAML file
	content := `name: "Acme Status"
slug: "acme-status"
theme: "dark"
is_public: true
show_uptime_percentage: true
enabled: true
header_text: "Welcome"
footer_text: "Contact us"
`
	tmpFile := createTempYAMLFile(t, content)
	defer os.Remove(tmpFile)

	req, err := buildStatusPageRequestFromYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Name != "Acme Status" {
		t.Errorf("expected Name 'Acme Status', got %q", req.Name)
	}
	if req.Slug != "acme-status" {
		t.Errorf("expected Slug 'acme-status', got %q", req.Slug)
	}
	if req.Theme != "dark" {
		t.Errorf("expected Theme 'dark', got %q", req.Theme)
	}
	if req.IsPublic == nil || !*req.IsPublic {
		t.Errorf("expected IsPublic true, got %v", req.IsPublic)
	}
	if req.HeaderText == nil || *req.HeaderText != "Welcome" {
		t.Errorf("expected HeaderText 'Welcome', got %v", req.HeaderText)
	}
}

func TestBuildStatusPageRequestFromYAML_MinimalFile(t *testing.T) {
	// Only required field
	content := `name: "Simple Status"
`
	tmpFile := createTempYAMLFile(t, content)
	defer os.Remove(tmpFile)

	req, err := buildStatusPageRequestFromYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Name != "Simple Status" {
		t.Errorf("expected Name 'Simple Status', got %q", req.Name)
	}
	// Optional fields should be nil/empty
	if req.CustomDomain != nil {
		t.Errorf("expected CustomDomain nil, got %v", req.CustomDomain)
	}
}

func TestBuildStatusPageRequestFromYAML_MissingName(t *testing.T) {
	// Missing required name field
	content := `slug: "acme-status"
theme: "dark"
`
	tmpFile := createTempYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildStatusPageRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for missing name")
	}
	if !containsString(err.Error(), "name") {
		t.Errorf("error should mention 'name', got: %v", err)
	}
}

func TestBuildStatusPageRequestFromYAML_InvalidTheme(t *testing.T) {
	content := `name: "Test Status"
theme: "invalid-theme"
`
	tmpFile := createTempYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildStatusPageRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for invalid theme")
	}
	if !containsString(err.Error(), "theme") {
		t.Errorf("error should mention 'theme', got: %v", err)
	}
}

func TestBuildStatusPageRequestFromYAML_InvalidSlug(t *testing.T) {
	content := `name: "Test Status"
slug: "a"
`
	tmpFile := createTempYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildStatusPageRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for invalid slug")
	}
	if !containsString(err.Error(), "slug") {
		t.Errorf("error should mention 'slug', got: %v", err)
	}
}

func TestBuildStatusPageRequestFromYAML_MalformedYAML(t *testing.T) {
	content := `name: "Test Status"
  invalid yaml syntax here
    - broken: indentation
`
	tmpFile := createTempYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildStatusPageRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
	if !containsString(err.Error(), "YAML") {
		t.Errorf("error should mention 'YAML', got: %v", err)
	}
}

func TestBuildStatusPageRequestFromYAML_FileNotFound(t *testing.T) {
	_, err := buildStatusPageRequestFromYAML("/nonexistent/path/to/file.yaml")
	if err == nil {
		t.Error("expected error for file not found")
	}
	if !containsString(err.Error(), "failed to read file") {
		t.Errorf("error should mention file read failure, got: %v", err)
	}
}

func TestBuildStatusPageRequestFromYAML_EmptyFile(t *testing.T) {
	content := ``
	tmpFile := createTempYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildStatusPageRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for empty file (missing name)")
	}
}

// createTempYAMLFile creates a temporary YAML file for testing and returns its path.
func createTempYAMLFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-status-page.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}

// Compile-time check that client.CreateStatusPageRequest is used correctly.
var _ *client.CreateStatusPageRequest = (*client.CreateStatusPageRequest)(nil)
