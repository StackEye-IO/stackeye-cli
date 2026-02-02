// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

func TestNewIncidentCreateCmd(t *testing.T) {
	cmd := NewIncidentCreateCmd()

	if cmd.Use != "create" {
		t.Errorf("expected Use='create', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Create a new incident for a status page" {
		t.Errorf("expected Short='Create a new incident for a status page', got %q", cmd.Short)
	}
}

func TestNewIncidentCreateCmd_Flags(t *testing.T) {
	cmd := NewIncidentCreateCmd()

	// Verify expected flags exist with correct defaults
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"status-page-id", "0"},
		{"title", ""},
		{"impact", ""},
		{"message", ""},
		{"status", "investigating"},
		{"from-file", ""},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("expected flag %q to exist", f.name)
			continue
		}
		if flag.DefValue != f.defaultValue {
			t.Errorf("flag %q: expected default %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestNewIncidentCreateCmd_Long(t *testing.T) {
	cmd := NewIncidentCreateCmd()

	long := cmd.Long

	// Should contain required flags documentation
	requiredFlags := []string{"--status-page-id", "--title", "--impact"}
	for _, flag := range requiredFlags {
		if !strings.Contains(long, flag) {
			t.Errorf("expected Long description to mention flag %q", flag)
		}
	}

	// Should contain status workflow
	statusValues := []string{"investigating", "identified", "monitoring", "resolved"}
	for _, status := range statusValues {
		if !strings.Contains(long, status) {
			t.Errorf("expected Long description to mention status %q", status)
		}
	}

	// Should contain impact levels
	impactValues := []string{"none", "minor", "major", "critical"}
	for _, impact := range impactValues {
		if !strings.Contains(long, impact) {
			t.Errorf("expected Long description to mention impact %q", impact)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye incident create") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention YAML file support
	if !strings.Contains(long, "--from-file") {
		t.Error("expected Long description to mention --from-file")
	}
	if !strings.Contains(long, "YAML File Format") {
		t.Error("expected Long description to contain YAML file format documentation")
	}
}

func TestBuildIncidentRequestFromFlags_RequiredTitle(t *testing.T) {
	flags := &incidentCreateFlags{
		title:  "",
		impact: "minor",
		status: "investigating",
	}

	_, err := buildIncidentRequestFromFlags(flags)
	if err == nil {
		t.Error("expected error for empty title")
	}
	if !strings.Contains(err.Error(), "--title is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBuildIncidentRequestFromFlags_RequiredMessage(t *testing.T) {
	flags := &incidentCreateFlags{
		title:   "Test Incident",
		message: "",
		impact:  "minor",
		status:  "investigating",
	}

	_, err := buildIncidentRequestFromFlags(flags)
	if err == nil {
		t.Error("expected error for empty message")
	}
	if !strings.Contains(err.Error(), "--message is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBuildIncidentRequestFromFlags_RequiredImpact(t *testing.T) {
	flags := &incidentCreateFlags{
		title:   "Test Incident",
		message: "Test message",
		impact:  "",
		status:  "investigating",
	}

	_, err := buildIncidentRequestFromFlags(flags)
	if err == nil {
		t.Error("expected error for empty impact")
	}
	if !strings.Contains(err.Error(), "--impact is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBuildIncidentRequestFromFlags_InvalidImpact(t *testing.T) {
	tests := []struct {
		name   string
		impact string
	}{
		{"invalid string", "invalid"},
		{"typo", "minnor"},
		{"partial", "crit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &incidentCreateFlags{
				title:   "Test Incident",
				message: "Test message",
				impact:  tt.impact,
				status:  "investigating",
			}

			_, err := buildIncidentRequestFromFlags(flags)
			if err == nil {
				t.Errorf("expected error for invalid impact %q", tt.impact)
			}
		})
	}
}

func TestBuildIncidentRequestFromFlags_InvalidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{"invalid string", "invalid"},
		{"typo", "investgating"},
		{"partial", "resolv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &incidentCreateFlags{
				title:   "Test Incident",
				message: "Test message",
				impact:  "minor",
				status:  tt.status,
			}

			_, err := buildIncidentRequestFromFlags(flags)
			if err == nil {
				t.Errorf("expected error for invalid status %q", tt.status)
			}
		})
	}
}

func TestBuildIncidentRequestFromFlags_ValidImpacts(t *testing.T) {
	validImpacts := []string{"none", "minor", "major", "critical", "None", "MINOR", "Major", "CRITICAL"}

	for _, impact := range validImpacts {
		t.Run(impact, func(t *testing.T) {
			flags := &incidentCreateFlags{
				title:   "Test Incident",
				message: "Test message",
				impact:  impact,
				status:  "investigating",
			}

			req, err := buildIncidentRequestFromFlags(flags)
			if err != nil {
				t.Errorf("unexpected error for valid impact %q: %v", impact, err)
			}
			// Verify impact is normalized to lowercase
			if req.Impact != strings.ToLower(impact) {
				t.Errorf("expected impact to be normalized to lowercase, got %q", req.Impact)
			}
		})
	}
}

func TestBuildIncidentRequestFromFlags_ValidStatuses(t *testing.T) {
	validStatuses := []string{"investigating", "identified", "monitoring", "resolved", "Investigating", "IDENTIFIED", "Monitoring", "RESOLVED"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			flags := &incidentCreateFlags{
				title:   "Test Incident",
				message: "Test message",
				impact:  "minor",
				status:  status,
			}

			req, err := buildIncidentRequestFromFlags(flags)
			if err != nil {
				t.Errorf("unexpected error for valid status %q: %v", status, err)
			}
			// Verify status is normalized to lowercase
			if req.Status != strings.ToLower(status) {
				t.Errorf("expected status to be normalized to lowercase, got %q", req.Status)
			}
		})
	}
}

func TestBuildIncidentRequestFromFlags_Success(t *testing.T) {
	flags := &incidentCreateFlags{
		title:   "API Degradation",
		message: "Users experiencing slow responses",
		status:  "investigating",
		impact:  "minor",
	}

	req, err := buildIncidentRequestFromFlags(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Title != "API Degradation" {
		t.Errorf("expected Title 'API Degradation', got %q", req.Title)
	}
	if req.Message != "Users experiencing slow responses" {
		t.Errorf("expected Message 'Users experiencing slow responses', got %q", req.Message)
	}
	if req.Status != "investigating" {
		t.Errorf("expected Status 'investigating', got %q", req.Status)
	}
	if req.Impact != "minor" {
		t.Errorf("expected Impact 'minor', got %q", req.Impact)
	}
}

func TestBuildIncidentRequestFromFlags_MinimalInput(t *testing.T) {
	flags := &incidentCreateFlags{
		title:   "Service Outage",
		message: "Investigating service disruption",
		impact:  "critical",
		status:  "investigating",
	}

	req, err := buildIncidentRequestFromFlags(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Title != "Service Outage" {
		t.Errorf("expected Title 'Service Outage', got %q", req.Title)
	}
	if req.Impact != "critical" {
		t.Errorf("expected Impact 'critical', got %q", req.Impact)
	}
	if req.Message != "Investigating service disruption" {
		t.Errorf("expected Message 'Investigating service disruption', got %q", req.Message)
	}
}

// YAML parsing tests

func TestBuildIncidentRequestFromYAML_ValidFile(t *testing.T) {
	content := `title: "Database Connectivity Issues"
message: "Users may experience intermittent connection errors"
status: "investigating"
impact: "major"
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	req, err := buildIncidentRequestFromYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Title != "Database Connectivity Issues" {
		t.Errorf("expected Title 'Database Connectivity Issues', got %q", req.Title)
	}
	if req.Message != "Users may experience intermittent connection errors" {
		t.Errorf("expected Message, got %q", req.Message)
	}
	if req.Status != "investigating" {
		t.Errorf("expected Status 'investigating', got %q", req.Status)
	}
	if req.Impact != "major" {
		t.Errorf("expected Impact 'major', got %q", req.Impact)
	}
}

func TestBuildIncidentRequestFromYAML_MinimalFile(t *testing.T) {
	// Required fields only (without optional status)
	content := `title: "Simple Incident"
message: "Brief description of the incident"
impact: "minor"
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	req, err := buildIncidentRequestFromYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Title != "Simple Incident" {
		t.Errorf("expected Title 'Simple Incident', got %q", req.Title)
	}
	if req.Message != "Brief description of the incident" {
		t.Errorf("expected Message 'Brief description of the incident', got %q", req.Message)
	}
	if req.Impact != "minor" {
		t.Errorf("expected Impact 'minor', got %q", req.Impact)
	}
	// Status defaults to investigating
	if req.Status != "investigating" {
		t.Errorf("expected default Status 'investigating', got %q", req.Status)
	}
}

func TestBuildIncidentRequestFromYAML_MissingTitle(t *testing.T) {
	content := `impact: "minor"
status: "investigating"
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildIncidentRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for missing title")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error should mention 'title', got: %v", err)
	}
}

func TestBuildIncidentRequestFromYAML_MissingMessage(t *testing.T) {
	content := `title: "Test Incident"
impact: "minor"
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildIncidentRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for missing message")
	}
	if !strings.Contains(err.Error(), "message") {
		t.Errorf("error should mention 'message', got: %v", err)
	}
}

func TestBuildIncidentRequestFromYAML_MissingImpact(t *testing.T) {
	content := `title: "Test Incident"
message: "Test message"
status: "investigating"
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildIncidentRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for missing impact")
	}
	if !strings.Contains(err.Error(), "impact") {
		t.Errorf("error should mention 'impact', got: %v", err)
	}
}

func TestBuildIncidentRequestFromYAML_InvalidImpact(t *testing.T) {
	content := `title: "Test Incident"
message: "Test message"
impact: "invalid-impact"
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildIncidentRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for invalid impact")
	}
	if !strings.Contains(err.Error(), "invalid impact") {
		t.Errorf("error should mention 'invalid impact', got: %v", err)
	}
}

func TestBuildIncidentRequestFromYAML_InvalidStatus(t *testing.T) {
	content := `title: "Test Incident"
message: "Test message"
impact: "minor"
status: "invalid-status"
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildIncidentRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("error should mention 'invalid status', got: %v", err)
	}
}

func TestBuildIncidentRequestFromYAML_MalformedYAML(t *testing.T) {
	content := `title: "Test Incident"
  invalid yaml syntax here
    - broken: indentation
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildIncidentRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
	if !strings.Contains(err.Error(), "YAML") {
		t.Errorf("error should mention 'YAML', got: %v", err)
	}
}

func TestBuildIncidentRequestFromYAML_FileNotFound(t *testing.T) {
	_, err := buildIncidentRequestFromYAML("/nonexistent/path/to/file.yaml")
	if err == nil {
		t.Error("expected error for file not found")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("error should mention file read failure, got: %v", err)
	}
}

func TestBuildIncidentRequestFromYAML_EmptyFile(t *testing.T) {
	content := ``
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildIncidentRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for empty file (missing title)")
	}
}

func TestBuildIncidentRequestFromYAML_StatusDefaultsToInvestigating(t *testing.T) {
	content := `title: "No Status Specified"
message: "Test message"
impact: "minor"
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	req, err := buildIncidentRequestFromYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Status != "investigating" {
		t.Errorf("expected default Status 'investigating', got %q", req.Status)
	}
}

func TestBuildIncidentRequestFromYAML_CaseNormalization(t *testing.T) {
	content := `title: "Test Incident"
message: "Test message"
impact: "MAJOR"
status: "IDENTIFIED"
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	req, err := buildIncidentRequestFromYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Impact != "major" {
		t.Errorf("expected Impact to be normalized to 'major', got %q", req.Impact)
	}
	if req.Status != "identified" {
		t.Errorf("expected Status to be normalized to 'identified', got %q", req.Status)
	}
}

// Command execution tests

func TestRunIncidentCreate_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	tests := []struct {
		name         string
		args         []string
		wantErrorMsg string
	}{
		{
			name:         "status-page-id required",
			args:         []string{"--title", "Test", "--message", "Test message", "--impact", "minor"},
			wantErrorMsg: "required flag(s) \"status-page-id\" not set",
		},
		{
			name:         "title required without from-file",
			args:         []string{"--status-page-id", "1", "--message", "Test message", "--impact", "minor"},
			wantErrorMsg: "--title is required",
		},
		{
			name:         "message required without from-file",
			args:         []string{"--status-page-id", "1", "--title", "Test", "--impact", "minor"},
			wantErrorMsg: "--message is required",
		},
		{
			name:         "impact required without from-file",
			args:         []string{"--status-page-id", "1", "--title", "Test", "--message", "Test message"},
			wantErrorMsg: "--impact is required",
		},
		{
			name:         "invalid impact",
			args:         []string{"--status-page-id", "1", "--title", "Test", "--message", "Test message", "--impact", "invalid"},
			wantErrorMsg: "invalid impact",
		},
		{
			name:         "invalid status",
			args:         []string{"--status-page-id", "1", "--title", "Test", "--message", "Test message", "--impact", "minor", "--status", "invalid"},
			wantErrorMsg: "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewIncidentCreateCmd()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErrorMsg)
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrorMsg) {
				t.Errorf("expected error containing %q, got %q", tt.wantErrorMsg, err.Error())
			}
		})
	}
}

func TestRunIncidentCreate_ValidFlags(t *testing.T) {
	// Test that valid flags pass validation (will fail later on API client)
	cmd := NewIncidentCreateCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--title", "Test Incident", "--message", "Test message", "--impact", "minor"})

	err := cmd.Execute()

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	validationErrors := []string{
		"--title is required",
		"--message is required",
		"--impact is required",
		"invalid impact",
		"invalid status",
	}
	for _, ve := range validationErrors {
		if strings.Contains(err.Error(), ve) {
			t.Errorf("got unexpected validation error: %s", err.Error())
		}
	}
}

func TestRunIncidentCreate_AllImpactLevels(t *testing.T) {
	// Test all valid impact levels
	impacts := []string{"none", "minor", "major", "critical"}

	for _, impact := range impacts {
		t.Run(impact, func(t *testing.T) {
			cmd := NewIncidentCreateCmd()
			cmd.SetArgs([]string{"--status-page-id", "1", "--title", "Test", "--message", "Test message", "--impact", impact})

			err := cmd.Execute()

			// Should fail on API client, not validation
			if err == nil {
				t.Error("expected error (no API client configured), got nil")
				return
			}

			// Error should not be about invalid impact
			if strings.Contains(err.Error(), "invalid impact") {
				t.Errorf("impact %q should be valid, got error: %s", impact, err.Error())
			}
		})
	}
}

func TestRunIncidentCreate_AllStatusValues(t *testing.T) {
	// Test all valid status values
	statuses := []string{"investigating", "identified", "monitoring", "resolved"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			cmd := NewIncidentCreateCmd()
			cmd.SetArgs([]string{"--status-page-id", "1", "--title", "Test", "--message", "Test message", "--impact", "minor", "--status", status})

			err := cmd.Execute()

			// Should fail on API client, not validation
			if err == nil {
				t.Error("expected error (no API client configured), got nil")
				return
			}

			// Error should not be about invalid status
			if strings.Contains(err.Error(), "invalid status") {
				t.Errorf("status %q should be valid, got error: %s", status, err.Error())
			}
		})
	}
}

func TestRunIncidentCreate_MessageIsRequired(t *testing.T) {
	// Verify that --message is required
	cmd := NewIncidentCreateCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--title", "Test", "--impact", "minor"}) // No --message flag

	err := cmd.Execute()

	// Should fail with message required error
	if err == nil {
		t.Error("expected error for missing message")
		return
	}

	// Error should be about missing message
	if !strings.Contains(err.Error(), "--message is required") {
		t.Errorf("expected '--message is required' error, got: %s", err.Error())
	}
}

func TestRunIncidentCreate_FromFile(t *testing.T) {
	// Create a valid YAML file
	content := `title: "Database Connectivity Issues"
message: "Users may experience intermittent connection errors"
impact: "major"
status: "investigating"
`
	tmpFile := createTempIncidentYAMLFile(t, content)
	defer os.Remove(tmpFile)

	cmd := NewIncidentCreateCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--from-file", tmpFile})

	err := cmd.Execute()

	// Should fail on API client, not YAML parsing
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should not be about YAML parsing
	yamlErrors := []string{
		"failed to read file",
		"failed to parse YAML",
		"must contain",
		"invalid impact",
		"invalid status",
	}
	for _, ye := range yamlErrors {
		if strings.Contains(err.Error(), ye) {
			t.Errorf("got unexpected YAML parsing error: %s", err.Error())
		}
	}
}

func TestRunIncidentCreate_FromFileInvalidPath(t *testing.T) {
	cmd := NewIncidentCreateCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--from-file", "/nonexistent/path/to/file.yaml"})

	err := cmd.Execute()

	if err == nil {
		t.Error("expected error for nonexistent file")
		return
	}

	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("expected file read error, got: %s", err.Error())
	}
}

// createTempIncidentYAMLFile creates a temporary YAML file for testing and returns its path.
func createTempIncidentYAMLFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-incident.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}

// Compile-time check that client.CreateIncidentRequest is used correctly.
var _ *client.CreateIncidentRequest = (*client.CreateIncidentRequest)(nil)
