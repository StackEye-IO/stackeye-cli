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

func TestNewProbeDepsAddCmd(t *testing.T) {
	cmd := NewProbeDepsAddCmd()

	if cmd.Use != "add <probe-id> --parent <parent-probe-id>" {
		t.Errorf("expected Use to be 'add <probe-id> --parent <parent-probe-id>', got %q", cmd.Use)
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

	// Verify --force flag exists
	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("expected --force flag to be defined")
	}

	// Verify -f shorthand
	if forceFlag != nil && forceFlag.Shorthand != "f" {
		t.Errorf("expected --force shorthand to be 'f', got %q", forceFlag.Shorthand)
	}
}

func TestProbeDepsAddCmd_Args(t *testing.T) {
	cmd := NewProbeDepsAddCmd()

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

func TestProbeDepsAddCmd_MissingParentFlag(t *testing.T) {
	cmd := NewProbeDepsAddCmd()

	// Create a parent command to hold the flag
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set args without --parent flag
	root.SetArgs([]string{"add", "550e8400-e29b-41d4-a716-446655440000"})

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

func TestProbeDepsAddCmd_InvalidProbeID(t *testing.T) {
	cmd := NewProbeDepsAddCmd()

	// Create a parent command to hold the flag
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set an invalid probe ID
	root.SetArgs([]string{"add", "not-a-valid-uuid", "--parent", "550e8400-e29b-41d4-a716-446655440000"})

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

func TestProbeDepsAddCmd_InvalidParentID(t *testing.T) {
	cmd := NewProbeDepsAddCmd()

	// Create a parent command to hold the flag
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set an invalid parent probe ID
	root.SetArgs([]string{"add", "550e8400-e29b-41d4-a716-446655440000", "--parent", "not-a-valid-uuid"})

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

func TestProbeDepsAddCmd_SameProbeID(t *testing.T) {
	cmd := NewProbeDepsAddCmd()

	// Create a parent command to hold the flag
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set same probe ID for both child and parent
	sameID := "550e8400-e29b-41d4-a716-446655440000"
	root.SetArgs([]string{"add", sameID, "--parent", sameID})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := root.Execute()
	if err == nil {
		t.Error("expected error when probe depends on itself")
	}

	if !strings.Contains(err.Error(), "cannot depend on itself") {
		t.Errorf("expected 'cannot depend on itself' error, got: %v", err)
	}
}

func TestHandleAddDependencyError(t *testing.T) {
	tests := []struct {
		name       string
		errMsg     string
		probeName  string
		parentName string
		wantMsg    string
	}{
		{
			name:       "cyclic dependency",
			errMsg:     "cyclic_dependency: would create a cycle",
			probeName:  "Web Server",
			parentName: "Database",
			wantMsg:    "circular dependency detected",
		},
		{
			name:       "dependency exists",
			errMsg:     "dependency_exists: already present",
			probeName:  "Web Server",
			parentName: "Database",
			wantMsg:    "dependency already exists",
		},
		{
			name:       "same probe",
			errMsg:     "same_probe: cannot depend on self",
			probeName:  "Web Server",
			parentName: "Web Server",
			wantMsg:    "cannot depend on itself",
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
			wantMsg:    "failed to add dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleAddDependencyError(
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

// mockError implements error interface for testing.
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantRes bool
	}{
		{
			name:    "nil error",
			err:     nil,
			wantRes: false,
		},
		{
			name:    "not found error",
			err:     &mockError{msg: "resource not found"},
			wantRes: true,
		},
		{
			name:    "404 error",
			err:     &mockError{msg: "HTTP 404 error"},
			wantRes: true,
		},
		{
			name:    "other error",
			err:     &mockError{msg: "connection timeout"},
			wantRes: false,
		},
		{
			name:    "ErrNotFound sentinel",
			err:     client.ErrNotFound,
			wantRes: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNotFoundError(tt.err)
			if got != tt.wantRes {
				t.Errorf("isNotFoundError() = %v, want %v", got, tt.wantRes)
			}
		})
	}
}

func TestProbeDepsAddIntegration_MockServer(t *testing.T) {
	// This test verifies the mock server setup for integration testing.
	// It validates the expected API contract for adding dependencies.
	probeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	parentID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")

	mockResponse := client.AddDependencyResponse{
		Status:        "success",
		Message:       "Dependency created successfully",
		ProbeID:       probeID,
		ParentProbeID: parentID,
	}

	var requestReceived bool
	var requestBody client.AddDependencyRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Verify the request path
		expectedPath := "/v1/probes/" + probeID.String() + "/dependencies"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// Verify request body
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if requestBody.ParentProbeID != parentID {
			t.Errorf("expected parent_probe_id %s, got %s", parentID, requestBody.ParentProbeID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Make a direct HTTP request to verify the mock server works
	reqBody := client.AddDependencyRequest{ParentProbeID: parentID}
	reqBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		server.URL+"/v1/probes/"+probeID.String()+"/dependencies",
		"application/json",
		bytes.NewReader(reqBytes),
	)
	if err != nil {
		t.Fatalf("failed to make request to mock server: %v", err)
	}
	defer resp.Body.Close()

	if !requestReceived {
		t.Error("mock server did not receive request")
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var result client.AddDependencyResponse
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
