// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func TestNewProbeDepsRemoveCmd(t *testing.T) {
	cmd := NewProbeDepsRemoveCmd()

	if cmd.Use != "remove <probe-id> --parent <parent-probe-id>" {
		t.Errorf("expected Use to be 'remove <probe-id> --parent <parent-probe-id>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Verify RunE is set
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}

	// Verify --parent flag exists and is required
	parentFlag := cmd.Flags().Lookup("parent")
	if parentFlag == nil {
		t.Error("expected --parent flag to be defined")
	}

	// Verify -p shorthand
	if parentFlag != nil && parentFlag.Shorthand != "p" {
		t.Errorf("expected --parent shorthand to be 'p', got %q", parentFlag.Shorthand)
	}

	// Verify --yes flag exists
	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Error("expected --yes flag to be defined")
	}

	// Verify -y shorthand
	if yesFlag != nil && yesFlag.Shorthand != "y" {
		t.Errorf("expected --yes shorthand to be 'y', got %q", yesFlag.Shorthand)
	}
}

func TestProbeDepsRemoveCmd_Args(t *testing.T) {
	cmd := NewProbeDepsRemoveCmd()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one valid argument",
			args:    []string{"550e8400-e29b-41d4-a716-446655440000"},
			wantErr: false,
		},
		{
			name:    "too many arguments",
			args:    []string{"550e8400-e29b-41d4-a716-446655440000", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProbeDepsRemoveCmd_MissingParentFlag(t *testing.T) {
	cmd := NewProbeDepsRemoveCmd()

	// Create a parent command to hold the flag
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set args without --parent flag
	root.SetArgs([]string{"remove", "550e8400-e29b-41d4-a716-446655440000"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing --parent flag")
	}

	if !strings.Contains(err.Error(), "parent") {
		t.Errorf("expected error about missing parent flag, got: %v", err)
	}
}

func TestProbeDepsRemoveCmd_InvalidProbeID(t *testing.T) {
	cmd := NewProbeDepsRemoveCmd()

	// Create a parent command to hold the flag
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set an invalid probe ID
	root.SetArgs([]string{"remove", "not-a-valid-uuid", "--parent", "550e8400-e29b-41d4-a716-446655440000", "--yes"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid probe ID")
	}

	if !strings.Contains(err.Error(), "invalid probe ID") {
		t.Errorf("expected 'invalid probe ID' error, got: %v", err)
	}
}

func TestProbeDepsRemoveCmd_InvalidParentID(t *testing.T) {
	cmd := NewProbeDepsRemoveCmd()

	// Create a parent command to hold the flag
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set an invalid parent probe ID
	root.SetArgs([]string{"remove", "550e8400-e29b-41d4-a716-446655440000", "--parent", "not-a-valid-uuid", "--yes"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid parent probe ID")
	}

	if !strings.Contains(err.Error(), "invalid parent probe ID") {
		t.Errorf("expected 'invalid parent probe ID' error, got: %v", err)
	}
}

func TestHandleRemoveDependencyError(t *testing.T) {
	tests := []struct {
		name       string
		errMsg     string
		probeName  string
		parentName string
		wantMsg    string
	}{
		{
			name:       "dependency not found",
			errMsg:     "dependency_not_found: no such dependency",
			probeName:  "Web Server",
			parentName: "Database",
			wantMsg:    "dependency not found",
		},
		{
			name:       "not found error",
			errMsg:     "resource not found",
			probeName:  "Web Server",
			parentName: "Database",
			wantMsg:    "dependency not found",
		},
		{
			name:       "probe not found",
			errMsg:     "probe_not_found: invalid probe ID",
			probeName:  "Web Server",
			parentName: "Database",
			wantMsg:    "one or more probes not found",
		},
		{
			name:       "unknown error",
			errMsg:     "some random error",
			probeName:  "Web Server",
			parentName: "Database",
			wantMsg:    "failed to remove dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleRemoveDependencyError(
				&mockError{msg: tt.errMsg},
				tt.probeName,
				tt.parentName,
			)
			if err == nil {
				t.Error("expected error to be returned")
				return
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("expected error to contain %q, got %q", tt.wantMsg, err.Error())
			}
		})
	}
}

func TestProbeDepsRemoveIntegration_MockServer(t *testing.T) {
	// This test verifies the mock server setup for integration testing.
	// It validates the expected API contract for removing dependencies.
	probeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	parentID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")

	mockResponse := client.RemoveDependencyResponse{
		Status:        "success",
		Message:       "Dependency removed successfully",
		ProbeID:       probeID,
		ParentProbeID: parentID,
	}

	var requestReceived bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Verify the request path
		expectedPath := "/v1/probes/" + probeID.String() + "/dependencies/" + parentID.String()
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify method
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Make a direct HTTP request to verify the mock server works
	req, _ := http.NewRequest(
		http.MethodDelete,
		server.URL+"/v1/probes/"+probeID.String()+"/dependencies/"+parentID.String(),
		nil,
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request to mock server: %v", err)
	}
	defer resp.Body.Close()

	if !requestReceived {
		t.Error("mock server did not receive request")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result client.RemoveDependencyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Status != "success" {
		t.Errorf("expected status 'success', got %q", result.Status)
	}

	if result.ProbeID != probeID {
		t.Errorf("expected probe_id %s, got %s", probeID, result.ProbeID)
	}

	if result.ParentProbeID != parentID {
		t.Errorf("expected parent_probe_id %s, got %s", parentID, result.ParentProbeID)
	}
}
