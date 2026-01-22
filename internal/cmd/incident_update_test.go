// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

func TestNewIncidentUpdateCmd(t *testing.T) {
	cmd := NewIncidentUpdateCmd()

	if cmd.Use != "update" {
		t.Errorf("expected Use='update', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Update an existing incident on a status page" {
		t.Errorf("expected Short='Update an existing incident on a status page', got %q", cmd.Short)
	}
}

func TestNewIncidentUpdateCmd_Flags(t *testing.T) {
	cmd := NewIncidentUpdateCmd()

	// Verify expected flags exist with correct defaults
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"status-page-id", "0"},
		{"incident-id", "0"},
		{"title", ""},
		{"message", ""},
		{"status", ""},
		{"impact", ""},
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

func TestNewIncidentUpdateCmd_Long(t *testing.T) {
	cmd := NewIncidentUpdateCmd()

	long := cmd.Long

	// Should contain required flags documentation
	requiredFlags := []string{"--status-page-id", "--incident-id"}
	for _, flag := range requiredFlags {
		if !strings.Contains(long, flag) {
			t.Errorf("expected Long description to mention flag %q", flag)
		}
	}

	// Should contain optional flags documentation
	optionalFlags := []string{"--title", "--message", "--status", "--impact", "--from-file"}
	for _, flag := range optionalFlags {
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
	if !strings.Contains(long, "stackeye incident update") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention YAML file support
	if !strings.Contains(long, "YAML File Format") {
		t.Error("expected Long description to contain YAML file format documentation")
	}
}

func TestBuildIncidentUpdateRequestFromFlags_NoFieldsProvided(t *testing.T) {
	flags := &incidentUpdateFlags{
		statusPageID: 123,
		incidentID:   456,
	}

	req, err := buildIncidentUpdateRequestFromFlags(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All fields should be nil when not provided
	if req.Title != nil {
		t.Error("expected Title to be nil")
	}
	if req.Message != nil {
		t.Error("expected Message to be nil")
	}
	if req.Status != nil {
		t.Error("expected Status to be nil")
	}
	if req.Impact != nil {
		t.Error("expected Impact to be nil")
	}
}

func TestBuildIncidentUpdateRequestFromFlags_InvalidStatus(t *testing.T) {
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
			flags := &incidentUpdateFlags{
				statusPageID: 123,
				incidentID:   456,
				status:       tt.status,
			}

			_, err := buildIncidentUpdateRequestFromFlags(flags)
			if err == nil {
				t.Errorf("expected error for invalid status %q", tt.status)
			}
		})
	}
}

func TestBuildIncidentUpdateRequestFromFlags_InvalidImpact(t *testing.T) {
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
			flags := &incidentUpdateFlags{
				statusPageID: 123,
				incidentID:   456,
				impact:       tt.impact,
			}

			_, err := buildIncidentUpdateRequestFromFlags(flags)
			if err == nil {
				t.Errorf("expected error for invalid impact %q", tt.impact)
			}
		})
	}
}

func TestBuildIncidentUpdateRequestFromFlags_ValidStatuses(t *testing.T) {
	validStatuses := []string{"investigating", "identified", "monitoring", "resolved", "Investigating", "IDENTIFIED", "Monitoring", "RESOLVED"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			flags := &incidentUpdateFlags{
				statusPageID: 123,
				incidentID:   456,
				status:       status,
			}

			req, err := buildIncidentUpdateRequestFromFlags(flags)
			if err != nil {
				t.Errorf("unexpected error for valid status %q: %v", status, err)
			}
			// Verify status is normalized to lowercase
			if req.Status == nil {
				t.Fatal("expected Status to be set")
			}
			if *req.Status != strings.ToLower(status) {
				t.Errorf("expected status to be normalized to lowercase, got %q", *req.Status)
			}
		})
	}
}

func TestBuildIncidentUpdateRequestFromFlags_ValidImpacts(t *testing.T) {
	validImpacts := []string{"none", "minor", "major", "critical", "None", "MINOR", "Major", "CRITICAL"}

	for _, impact := range validImpacts {
		t.Run(impact, func(t *testing.T) {
			flags := &incidentUpdateFlags{
				statusPageID: 123,
				incidentID:   456,
				impact:       impact,
			}

			req, err := buildIncidentUpdateRequestFromFlags(flags)
			if err != nil {
				t.Errorf("unexpected error for valid impact %q: %v", impact, err)
			}
			// Verify impact is normalized to lowercase
			if req.Impact == nil {
				t.Fatal("expected Impact to be set")
			}
			if *req.Impact != strings.ToLower(impact) {
				t.Errorf("expected impact to be normalized to lowercase, got %q", *req.Impact)
			}
		})
	}
}

func TestBuildIncidentUpdateRequestFromFlags_SingleField(t *testing.T) {
	tests := []struct {
		name    string
		flags   *incidentUpdateFlags
		checkFn func(t *testing.T, req *client.UpdateIncidentRequest)
	}{
		{
			name: "title only",
			flags: &incidentUpdateFlags{
				statusPageID: 123,
				incidentID:   456,
				title:        "New Title",
			},
			checkFn: func(t *testing.T, req *client.UpdateIncidentRequest) {
				if req.Title == nil || *req.Title != "New Title" {
					t.Error("expected Title to be set to 'New Title'")
				}
				if req.Message != nil || req.Status != nil || req.Impact != nil {
					t.Error("expected other fields to be nil")
				}
			},
		},
		{
			name: "message only",
			flags: &incidentUpdateFlags{
				statusPageID: 123,
				incidentID:   456,
				message:      "Updated message",
			},
			checkFn: func(t *testing.T, req *client.UpdateIncidentRequest) {
				if req.Message == nil || *req.Message != "Updated message" {
					t.Error("expected Message to be set")
				}
				if req.Title != nil || req.Status != nil || req.Impact != nil {
					t.Error("expected other fields to be nil")
				}
			},
		},
		{
			name: "status only",
			flags: &incidentUpdateFlags{
				statusPageID: 123,
				incidentID:   456,
				status:       "identified",
			},
			checkFn: func(t *testing.T, req *client.UpdateIncidentRequest) {
				if req.Status == nil || *req.Status != "identified" {
					t.Error("expected Status to be set to 'identified'")
				}
				if req.Title != nil || req.Message != nil || req.Impact != nil {
					t.Error("expected other fields to be nil")
				}
			},
		},
		{
			name: "impact only",
			flags: &incidentUpdateFlags{
				statusPageID: 123,
				incidentID:   456,
				impact:       "major",
			},
			checkFn: func(t *testing.T, req *client.UpdateIncidentRequest) {
				if req.Impact == nil || *req.Impact != "major" {
					t.Error("expected Impact to be set to 'major'")
				}
				if req.Title != nil || req.Message != nil || req.Status != nil {
					t.Error("expected other fields to be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := buildIncidentUpdateRequestFromFlags(tt.flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.checkFn(t, req)
		})
	}
}

func TestBuildIncidentUpdateRequestFromFlags_AllFields(t *testing.T) {
	flags := &incidentUpdateFlags{
		statusPageID: 123,
		incidentID:   456,
		title:        "New Title",
		message:      "Updated message",
		status:       "monitoring",
		impact:       "minor",
	}

	req, err := buildIncidentUpdateRequestFromFlags(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Title == nil || *req.Title != "New Title" {
		t.Errorf("expected Title 'New Title', got %v", req.Title)
	}
	if req.Message == nil || *req.Message != "Updated message" {
		t.Errorf("expected Message 'Updated message', got %v", req.Message)
	}
	if req.Status == nil || *req.Status != "monitoring" {
		t.Errorf("expected Status 'monitoring', got %v", req.Status)
	}
	if req.Impact == nil || *req.Impact != "minor" {
		t.Errorf("expected Impact 'minor', got %v", req.Impact)
	}
}

// YAML parsing tests

func TestBuildIncidentUpdateRequestFromYAML_AllFields(t *testing.T) {
	content := `title: "Database Connectivity Issues"
message: "Users may experience intermittent connection errors"
status: "investigating"
impact: "major"
`
	tmpFile := createTempIncidentUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	req, err := buildIncidentUpdateRequestFromYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Title == nil || *req.Title != "Database Connectivity Issues" {
		t.Errorf("expected Title 'Database Connectivity Issues', got %v", req.Title)
	}
	if req.Message == nil || *req.Message != "Users may experience intermittent connection errors" {
		t.Errorf("expected Message, got %v", req.Message)
	}
	if req.Status == nil || *req.Status != "investigating" {
		t.Errorf("expected Status 'investigating', got %v", req.Status)
	}
	if req.Impact == nil || *req.Impact != "major" {
		t.Errorf("expected Impact 'major', got %v", req.Impact)
	}
}

func TestBuildIncidentUpdateRequestFromYAML_SingleField(t *testing.T) {
	content := `status: "identified"
`
	tmpFile := createTempIncidentUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	req, err := buildIncidentUpdateRequestFromYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Status == nil || *req.Status != "identified" {
		t.Errorf("expected Status 'identified', got %v", req.Status)
	}
	if req.Title != nil || req.Message != nil || req.Impact != nil {
		t.Error("expected other fields to be nil")
	}
}

func TestBuildIncidentUpdateRequestFromYAML_EmptyFile(t *testing.T) {
	content := ``
	tmpFile := createTempIncidentUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	req, err := buildIncidentUpdateRequestFromYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty file is valid - will be caught by runIncidentUpdate validation
	if req.Title != nil || req.Message != nil || req.Status != nil || req.Impact != nil {
		t.Error("expected all fields to be nil for empty file")
	}
}

func TestBuildIncidentUpdateRequestFromYAML_InvalidStatus(t *testing.T) {
	content := `status: "invalid-status"
`
	tmpFile := createTempIncidentUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildIncidentUpdateRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("error should mention 'invalid status', got: %v", err)
	}
}

func TestBuildIncidentUpdateRequestFromYAML_InvalidImpact(t *testing.T) {
	content := `impact: "invalid-impact"
`
	tmpFile := createTempIncidentUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildIncidentUpdateRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for invalid impact")
	}
	if !strings.Contains(err.Error(), "invalid impact") {
		t.Errorf("error should mention 'invalid impact', got: %v", err)
	}
}

func TestBuildIncidentUpdateRequestFromYAML_MalformedYAML(t *testing.T) {
	content := `status: "investigating"
  invalid yaml syntax here
    - broken: indentation
`
	tmpFile := createTempIncidentUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	_, err := buildIncidentUpdateRequestFromYAML(tmpFile)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
	if !strings.Contains(err.Error(), "YAML") {
		t.Errorf("error should mention 'YAML', got: %v", err)
	}
}

func TestBuildIncidentUpdateRequestFromYAML_FileNotFound(t *testing.T) {
	_, err := buildIncidentUpdateRequestFromYAML("/nonexistent/path/to/file.yaml")
	if err == nil {
		t.Error("expected error for file not found")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("error should mention file read failure, got: %v", err)
	}
}

func TestBuildIncidentUpdateRequestFromYAML_CaseNormalization(t *testing.T) {
	content := `status: "IDENTIFIED"
impact: "MAJOR"
`
	tmpFile := createTempIncidentUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	req, err := buildIncidentUpdateRequestFromYAML(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Status == nil || *req.Status != "identified" {
		t.Errorf("expected Status to be normalized to 'identified', got %v", req.Status)
	}
	if req.Impact == nil || *req.Impact != "major" {
		t.Errorf("expected Impact to be normalized to 'major', got %v", req.Impact)
	}
}

// Command execution tests

func TestRunIncidentUpdate_RequiredFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErrorMsg string
	}{
		{
			name:         "status-page-id required",
			args:         []string{"--incident-id", "456", "--status", "identified"},
			wantErrorMsg: "required flag(s) \"status-page-id\" not set",
		},
		{
			name:         "incident-id required",
			args:         []string{"--status-page-id", "123", "--status", "identified"},
			wantErrorMsg: "required flag(s) \"incident-id\" not set",
		},
		{
			name:         "at least one update field required",
			args:         []string{"--status-page-id", "123", "--incident-id", "456"},
			wantErrorMsg: "at least one update field is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewIncidentUpdateCmd()
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

func TestRunIncidentUpdate_InvalidValues(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErrorMsg string
	}{
		{
			name:         "invalid status",
			args:         []string{"--status-page-id", "123", "--incident-id", "456", "--status", "invalid"},
			wantErrorMsg: "invalid status",
		},
		{
			name:         "invalid impact",
			args:         []string{"--status-page-id", "123", "--incident-id", "456", "--impact", "invalid"},
			wantErrorMsg: "invalid impact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewIncidentUpdateCmd()
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

func TestRunIncidentUpdate_ValidFlags(t *testing.T) {
	// Test that valid flags pass validation (will fail later on API client)
	cmd := NewIncidentUpdateCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--incident-id", "456", "--status", "identified"})

	err := cmd.Execute()

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	validationErrors := []string{
		"at least one update field is required",
		"invalid status",
		"invalid impact",
	}
	for _, ve := range validationErrors {
		if strings.Contains(err.Error(), ve) {
			t.Errorf("got unexpected validation error: %s", err.Error())
		}
	}
}

func TestRunIncidentUpdate_AllStatusValues(t *testing.T) {
	statuses := []string{"investigating", "identified", "monitoring", "resolved"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			cmd := NewIncidentUpdateCmd()
			cmd.SetArgs([]string{"--status-page-id", "1", "--incident-id", "1", "--status", status})

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

func TestRunIncidentUpdate_AllImpactLevels(t *testing.T) {
	impacts := []string{"none", "minor", "major", "critical"}

	for _, impact := range impacts {
		t.Run(impact, func(t *testing.T) {
			cmd := NewIncidentUpdateCmd()
			cmd.SetArgs([]string{"--status-page-id", "1", "--incident-id", "1", "--impact", impact})

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

func TestRunIncidentUpdate_FromFile(t *testing.T) {
	content := `status: "monitoring"
message: "Fix deployed, monitoring for stability"
`
	tmpFile := createTempIncidentUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	cmd := NewIncidentUpdateCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--incident-id", "456", "--from-file", tmpFile})

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
		"invalid status",
		"invalid impact",
	}
	for _, ye := range yamlErrors {
		if strings.Contains(err.Error(), ye) {
			t.Errorf("got unexpected YAML parsing error: %s", err.Error())
		}
	}
}

func TestRunIncidentUpdate_FromFileInvalidPath(t *testing.T) {
	cmd := NewIncidentUpdateCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--incident-id", "456", "--from-file", "/nonexistent/path/to/file.yaml"})

	err := cmd.Execute()

	if err == nil {
		t.Error("expected error for nonexistent file")
		return
	}

	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("expected file read error, got: %s", err.Error())
	}
}

func TestRunIncidentUpdate_FromFileEmptyFile(t *testing.T) {
	content := ``
	tmpFile := createTempIncidentUpdateYAMLFile(t, content)
	defer os.Remove(tmpFile)

	cmd := NewIncidentUpdateCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--incident-id", "456", "--from-file", tmpFile})

	err := cmd.Execute()

	if err == nil {
		t.Error("expected error for empty file (no update fields)")
		return
	}

	if !strings.Contains(err.Error(), "at least one update field is required") {
		t.Errorf("expected 'at least one update field is required' error, got: %s", err.Error())
	}
}

// createTempIncidentUpdateYAMLFile creates a temporary YAML file for testing and returns its path.
func createTempIncidentUpdateYAMLFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-incident-update.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}

// Compile-time check that client.UpdateIncidentRequest is used correctly.
var _ *client.UpdateIncidentRequest = (*client.UpdateIncidentRequest)(nil)
