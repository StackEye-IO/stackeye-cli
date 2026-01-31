// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func TestNewProbeDepsListCmd(t *testing.T) {
	cmd := NewProbeDepsListCmd()

	if cmd.Use != "list <probe-id>" {
		t.Errorf("expected Use to be 'list <probe-id>', got %q", cmd.Use)
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
}

func TestProbeDepsListCmd_Args(t *testing.T) {
	cmd := NewProbeDepsListCmd()

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

func TestProbeDepsListCmd_InvalidProbeID(t *testing.T) {
	cmd := NewProbeDepsListCmd()

	// Create a parent command to hold the flag
	root := &cobra.Command{}
	root.AddCommand(cmd)

	// Set an invalid probe ID
	root.SetArgs([]string{"list", "not-a-valid-uuid"})

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

func TestFormatDependencyStatus(t *testing.T) {
	// Use ColorNever to get predictable output without ANSI codes
	colorMgr := sdkoutput.NewColorManager(sdkoutput.ColorNever)

	tests := []struct {
		name          string
		status        string
		isUnreachable bool
		want          string
	}{
		{
			name:          "up status",
			status:        "up",
			isUnreachable: false,
			want:          "UP",
		},
		{
			name:          "down status",
			status:        "down",
			isUnreachable: false,
			want:          "DOWN",
		},
		{
			name:          "unreachable overrides status",
			status:        "up",
			isUnreachable: true,
			want:          "UNREACHABLE",
		},
		{
			name:          "unreachable with down status",
			status:        "down",
			isUnreachable: true,
			want:          "UNREACHABLE",
		},
		{
			name:          "pending status",
			status:        "pending",
			isUnreachable: false,
			want:          "PENDING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDependencyStatus(tt.status, tt.isUnreachable, colorMgr)
			if got != tt.want {
				t.Errorf("formatDependencyStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintDependenciesTable_NoParentsNoChildren(t *testing.T) {
	deps := &client.ProbeDependencyInfo{
		Parents:  []client.ProbeBasicInfo{},
		Children: []client.ProbeBasicInfo{},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printDependenciesTable(deps)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("printDependenciesTable() error = %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "PARENTS (0):") {
		t.Errorf("expected 'PARENTS (0):' in output, got: %s", output)
	}

	if !strings.Contains(output, "No parent dependencies") {
		t.Errorf("expected 'No parent dependencies' in output, got: %s", output)
	}

	if !strings.Contains(output, "CHILDREN (0):") {
		t.Errorf("expected 'CHILDREN (0):' in output, got: %s", output)
	}

	if !strings.Contains(output, "No child dependencies") {
		t.Errorf("expected 'No child dependencies' in output, got: %s", output)
	}
}

func TestPrintDependenciesTable_WithParentsAndChildren(t *testing.T) {
	parentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	childID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")

	deps := &client.ProbeDependencyInfo{
		Parents: []client.ProbeBasicInfo{
			{
				ID:            parentID,
				Name:          "Database Check",
				Status:        "up",
				IsUnreachable: false,
			},
		},
		Children: []client.ProbeBasicInfo{
			{
				ID:            childID,
				Name:          "Web Server HTTP",
				Status:        "down",
				IsUnreachable: true,
			},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printDependenciesTable(deps)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("printDependenciesTable() error = %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "PARENTS (1):") {
		t.Errorf("expected 'PARENTS (1):' in output, got: %s", output)
	}

	if !strings.Contains(output, "Database Check") {
		t.Errorf("expected 'Database Check' in output, got: %s", output)
	}

	if !strings.Contains(output, "UP") {
		t.Errorf("expected 'UP' status in output, got: %s", output)
	}

	if !strings.Contains(output, "CHILDREN (1):") {
		t.Errorf("expected 'CHILDREN (1):' in output, got: %s", output)
	}

	if !strings.Contains(output, "Web Server HTTP") {
		t.Errorf("expected 'Web Server HTTP' in output, got: %s", output)
	}

	if !strings.Contains(output, "UNREACHABLE") {
		t.Errorf("expected 'UNREACHABLE' status in output, got: %s", output)
	}
}

func TestPrintDependenciesTable_LongName(t *testing.T) {
	parentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	deps := &client.ProbeDependencyInfo{
		Parents: []client.ProbeBasicInfo{
			{
				ID:            parentID,
				Name:          "This is a very long probe name that exceeds thirty characters",
				Status:        "up",
				IsUnreachable: false,
			},
		},
		Children: []client.ProbeBasicInfo{},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printDependenciesTable(deps)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("printDependenciesTable() error = %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Name should be truncated with "..."
	if !strings.Contains(output, "...") {
		t.Errorf("expected truncated name with '...' in output, got: %s", output)
	}
}

func TestProbeDepsListIntegration_MockServer(t *testing.T) {
	// This test verifies the mock server setup for integration testing.
	// Full integration tests require the api.GetClient() setup which
	// depends on CLI configuration. This test validates the expected
	// API contract.
	probeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	parentID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")

	mockDeps := client.ProbeDependencyInfo{
		Parents: []client.ProbeBasicInfo{
			{
				ID:            parentID,
				Name:          "Parent Probe",
				Status:        "up",
				IsUnreachable: false,
			},
		},
		Children: []client.ProbeBasicInfo{},
	}

	var requestReceived bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Verify the request path
		expectedPath := "/v1/probes/" + probeID.String() + "/dependencies"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify method
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockDeps)
	}))
	defer server.Close()

	// Make a direct HTTP request to verify the mock server works
	resp, err := http.Get(server.URL + "/v1/probes/" + probeID.String() + "/dependencies")
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

	var result client.ProbeDependencyInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(result.Parents) != 1 {
		t.Errorf("expected 1 parent, got %d", len(result.Parents))
	}

	if result.Parents[0].Name != "Parent Probe" {
		t.Errorf("expected parent name 'Parent Probe', got %q", result.Parents[0].Name)
	}
}
