// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func TestNewStatusPageUpdateCmd(t *testing.T) {
	cmd := NewStatusPageUpdateCmd()

	if cmd.Use != "update <id>" {
		t.Errorf("expected Use to be 'update <id>', got %q", cmd.Use)
	}

	if cmd.Short != "Update an existing status page" {
		t.Errorf("expected Short to be 'Update an existing status page', got %q", cmd.Short)
	}
}

func TestNewStatusPageUpdateCmd_RequiresArg(t *testing.T) {
	cmd := NewStatusPageUpdateCmd()

	// Should require exactly 1 argument
	if cmd.Args == nil {
		t.Fatal("expected Args validation to be set")
	}

	// Test that it requires exactly 1 arg
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"123"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"123", "456"})
	if err == nil {
		t.Error("expected error when 2 args provided")
	}
}

func TestNewStatusPageUpdateCmd_Flags(t *testing.T) {
	cmd := NewStatusPageUpdateCmd()

	// All flags should exist
	expectedFlags := []string{
		"name",
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

	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected %q flag to exist", flag)
		}
	}
}

func TestNewStatusPageUpdateCmd_Long(t *testing.T) {
	cmd := NewStatusPageUpdateCmd()

	// Verify Long contains key documentation
	expectedContains := []string{
		"Update an existing status page",
		"partial updates",
		"--name",
		"--theme",
		"--enabled",
		"YAML File Format",
	}

	for _, expected := range expectedContains {
		if !containsString(cmd.Long, expected) {
			t.Errorf("expected Long to contain %q", expected)
		}
	}
}

func TestBuildStatusPageUpdateRequest_NoFlags(t *testing.T) {
	cmd := createTestUpdateCmd()

	flags := &statusPageUpdateFlags{}
	initStatusPageUpdateFlagsForTest(flags)

	_, err := buildStatusPageUpdateRequest(cmd, flags)
	if err == nil {
		t.Error("expected error when no flags specified")
	}
	if !containsString(err.Error(), "no update flags") {
		t.Errorf("expected error about no update flags, got: %v", err)
	}
}

func TestBuildStatusPageUpdateRequest_EmptyName(t *testing.T) {
	cmd := createTestUpdateCmd()
	_ = cmd.Flags().Set("name", "")

	flags := &statusPageUpdateFlags{}
	initStatusPageUpdateFlagsForTest(flags)

	_, err := buildStatusPageUpdateRequest(cmd, flags)
	if err == nil {
		t.Error("expected error for empty name")
	}
	if !containsString(err.Error(), "--name cannot be empty") {
		t.Errorf("expected error about empty name, got: %v", err)
	}
}

func TestBuildStatusPageUpdateRequest_EmptySlug(t *testing.T) {
	cmd := createTestUpdateCmd()
	_ = cmd.Flags().Set("slug", "")

	flags := &statusPageUpdateFlags{}
	initStatusPageUpdateFlagsForTest(flags)

	_, err := buildStatusPageUpdateRequest(cmd, flags)
	if err == nil {
		t.Error("expected error for empty slug")
	}
	if !containsString(err.Error(), "--slug cannot be empty") {
		t.Errorf("expected error about empty slug, got: %v", err)
	}
}

func TestBuildStatusPageUpdateRequest_InvalidSlug(t *testing.T) {
	cmd := createTestUpdateCmd()
	_ = cmd.Flags().Set("slug", "a") // too short

	flags := &statusPageUpdateFlags{}
	initStatusPageUpdateFlagsForTest(flags)
	*flags.slug = "a"

	_, err := buildStatusPageUpdateRequest(cmd, flags)
	if err == nil {
		t.Error("expected error for invalid slug")
	}
	if !containsString(err.Error(), "slug") {
		t.Errorf("expected error about slug, got: %v", err)
	}
}

func TestBuildStatusPageUpdateRequest_InvalidTheme(t *testing.T) {
	cmd := createTestUpdateCmd()
	_ = cmd.Flags().Set("theme", "invalid-theme")

	flags := &statusPageUpdateFlags{}
	initStatusPageUpdateFlagsForTest(flags)
	*flags.theme = "invalid-theme"

	_, err := buildStatusPageUpdateRequest(cmd, flags)
	if err == nil {
		t.Error("expected error for invalid theme")
	}
	if !containsString(err.Error(), "theme") {
		t.Errorf("expected error about theme, got: %v", err)
	}
}

func TestBuildStatusPageUpdateRequest_ValidName(t *testing.T) {
	cmd := createTestUpdateCmd()
	_ = cmd.Flags().Set("name", "Updated Name")

	flags := &statusPageUpdateFlags{}
	initStatusPageUpdateFlagsForTest(flags)
	*flags.name = "Updated Name"

	req, err := buildStatusPageUpdateRequest(cmd, flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Name == nil || *req.Name != "Updated Name" {
		t.Errorf("expected Name 'Updated Name', got %v", req.Name)
	}
	// Other fields should be nil (not updated)
	if req.Slug != nil {
		t.Errorf("expected Slug nil, got %v", req.Slug)
	}
	if req.Theme != nil {
		t.Errorf("expected Theme nil, got %v", req.Theme)
	}
}

func TestBuildStatusPageUpdateRequest_ValidTheme(t *testing.T) {
	tests := []struct {
		name          string
		inputTheme    string
		expectedTheme string
	}{
		{"lowercase dark", "dark", "dark"},
		{"uppercase LIGHT", "LIGHT", "light"},
		{"mixed System", "System", "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestUpdateCmd()
			_ = cmd.Flags().Set("theme", tt.inputTheme)

			flags := &statusPageUpdateFlags{}
			initStatusPageUpdateFlagsForTest(flags)
			*flags.theme = tt.inputTheme

			req, err := buildStatusPageUpdateRequest(cmd, flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if req.Theme == nil || *req.Theme != tt.expectedTheme {
				t.Errorf("expected Theme %q, got %v", tt.expectedTheme, req.Theme)
			}
		})
	}
}

func TestBuildStatusPageUpdateRequest_BooleanFlags(t *testing.T) {
	cmd := createTestUpdateCmd()
	_ = cmd.Flags().Set("enabled", "false")
	_ = cmd.Flags().Set("public", "true")
	_ = cmd.Flags().Set("show-uptime", "false")

	flags := &statusPageUpdateFlags{}
	initStatusPageUpdateFlagsForTest(flags)
	*flags.enabled = false
	*flags.isPublic = true
	*flags.showUptimePercentage = false

	req, err := buildStatusPageUpdateRequest(cmd, flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Enabled == nil || *req.Enabled != false {
		t.Errorf("expected Enabled false, got %v", req.Enabled)
	}
	if req.IsPublic == nil || *req.IsPublic != true {
		t.Errorf("expected IsPublic true, got %v", req.IsPublic)
	}
	if req.ShowUptimePercentage == nil || *req.ShowUptimePercentage != false {
		t.Errorf("expected ShowUptimePercentage false, got %v", req.ShowUptimePercentage)
	}
}

func TestBuildStatusPageUpdateRequest_MultipleFlags(t *testing.T) {
	cmd := createTestUpdateCmd()
	_ = cmd.Flags().Set("name", "New Name")
	_ = cmd.Flags().Set("theme", "dark")
	_ = cmd.Flags().Set("header-text", "New Header")
	_ = cmd.Flags().Set("enabled", "true")

	flags := &statusPageUpdateFlags{}
	initStatusPageUpdateFlagsForTest(flags)
	*flags.name = "New Name"
	*flags.theme = "dark"
	*flags.headerText = "New Header"
	*flags.enabled = true

	req, err := buildStatusPageUpdateRequest(cmd, flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Name == nil || *req.Name != "New Name" {
		t.Errorf("expected Name 'New Name', got %v", req.Name)
	}
	if req.Theme == nil || *req.Theme != "dark" {
		t.Errorf("expected Theme 'dark', got %v", req.Theme)
	}
	if req.HeaderText == nil || *req.HeaderText != "New Header" {
		t.Errorf("expected HeaderText 'New Header', got %v", req.HeaderText)
	}
	if req.Enabled == nil || *req.Enabled != true {
		t.Errorf("expected Enabled true, got %v", req.Enabled)
	}
}

// YAML update tests

func TestStatusPageUpdateFromYAML_ValidFile(t *testing.T) {
	content := `name: "Updated Name"
theme: "dark"
enabled: true
`
	tmpFile := createTempUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	cfg, err := parseStatusPageUpdateYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name == nil || *cfg.Name != "Updated Name" {
		t.Errorf("expected Name 'Updated Name', got %v", cfg.Name)
	}
	if cfg.Theme == nil || *cfg.Theme != "dark" {
		t.Errorf("expected Theme 'dark', got %v", cfg.Theme)
	}
	if cfg.Enabled == nil || *cfg.Enabled != true {
		t.Errorf("expected Enabled true, got %v", cfg.Enabled)
	}
}

func TestStatusPageUpdateFromYAML_PartialUpdate(t *testing.T) {
	// Only updating a single field
	content := `header_text: "Maintenance scheduled"
`
	tmpFile := createTempUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	cfg, err := parseStatusPageUpdateYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HeaderText == nil || *cfg.HeaderText != "Maintenance scheduled" {
		t.Errorf("expected HeaderText 'Maintenance scheduled', got %v", cfg.HeaderText)
	}
	// Other fields should be nil
	if cfg.Name != nil {
		t.Errorf("expected Name nil, got %v", cfg.Name)
	}
	if cfg.Theme != nil {
		t.Errorf("expected Theme nil, got %v", cfg.Theme)
	}
}

func TestStatusPageUpdateFromYAML_EmptyFile(t *testing.T) {
	content := ``
	tmpFile := createTempUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	cfg, err := parseStatusPageUpdateYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All fields should be nil for empty file
	if cfg.Name != nil {
		t.Errorf("expected Name nil for empty file, got %v", cfg.Name)
	}
}

func TestStatusPageUpdateFromYAML_MalformedYAML(t *testing.T) {
	content := `name: "Test"
  invalid yaml syntax here
    - broken: indentation
`
	tmpFile := createTempUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := parseStatusPageUpdateYAML(tmpFile)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
}

func TestStatusPageUpdateFromYAML_FileNotFound(t *testing.T) {
	_, err := parseStatusPageUpdateYAML("/nonexistent/path/to/file.yaml")
	if err == nil {
		t.Error("expected error for file not found")
	}
}

// Helper functions

// createTestUpdateCmd creates a cobra command with the status-page update flags for testing.
func createTestUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{}

	// Add all flags that would be checked
	var nameStr, slugStr, customDomainStr, logoURLStr, faviconURLStr string
	var headerTextStr, footerTextStr, themeStr string
	var isPublicBool, showUptimeBool, enabledBool bool

	cmd.Flags().StringVar(&nameStr, "name", "", "")
	cmd.Flags().StringVar(&slugStr, "slug", "", "")
	cmd.Flags().StringVar(&customDomainStr, "custom-domain", "", "")
	cmd.Flags().StringVar(&logoURLStr, "logo-url", "", "")
	cmd.Flags().StringVar(&faviconURLStr, "favicon-url", "", "")
	cmd.Flags().StringVar(&headerTextStr, "header-text", "", "")
	cmd.Flags().StringVar(&footerTextStr, "footer-text", "", "")
	cmd.Flags().StringVar(&themeStr, "theme", "", "")
	cmd.Flags().BoolVar(&isPublicBool, "public", false, "")
	cmd.Flags().BoolVar(&showUptimeBool, "show-uptime", false, "")
	cmd.Flags().BoolVar(&enabledBool, "enabled", false, "")

	return cmd
}

// initStatusPageUpdateFlagsForTest initializes the pointer fields in statusPageUpdateFlags.
func initStatusPageUpdateFlagsForTest(flags *statusPageUpdateFlags) {
	flags.name = new(string)
	flags.slug = new(string)
	flags.customDomain = new(string)
	flags.logoURL = new(string)
	flags.faviconURL = new(string)
	flags.headerText = new(string)
	flags.footerText = new(string)
	flags.theme = new(string)
	flags.isPublic = new(bool)
	flags.showUptimePercentage = new(bool)
	flags.enabled = new(bool)
}

// parseStatusPageUpdateYAML parses a YAML file for testing.
func parseStatusPageUpdateYAML(filePath string) (*statusPageUpdateYAMLConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg statusPageUpdateYAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// createTempUpdateYAMLFile creates a temporary YAML file for testing and returns its path.
func createTempUpdateYAMLFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-status-page-update.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}

// Compile-time check that client.UpdateStatusPageRequest is used correctly.
var _ *client.UpdateStatusPageRequest = (*client.UpdateStatusPageRequest)(nil)
