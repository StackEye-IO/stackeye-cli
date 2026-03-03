// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewAgentRegisterCmd verifies the register command is correctly configured.
func TestNewAgentRegisterCmd(t *testing.T) {
	cmd := NewAgentRegisterCmd()

	if cmd.Use != "register" {
		t.Errorf("expected Use='register', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewAgentRegisterCmd_RequiredFlag verifies --name is required.
func TestNewAgentRegisterCmd_RequiredFlag(t *testing.T) {
	cmd := NewAgentRegisterCmd()

	flag := cmd.Flags().Lookup("name")
	if flag == nil {
		t.Fatal("expected --name flag to be defined")
	}

	annotations := flag.Annotations
	if annotations == nil {
		t.Fatal("expected flag annotations to be set")
	}
	if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
		t.Error("expected --name to be marked as required")
	}
}

// TestNewAgentRegisterCmd_OptionalDescription verifies --description is optional.
func TestNewAgentRegisterCmd_OptionalDescription(t *testing.T) {
	cmd := NewAgentRegisterCmd()

	flag := cmd.Flags().Lookup("description")
	if flag == nil {
		t.Fatal("expected --description flag to be defined")
	}
	// Description flag should not be required
	if ann := flag.Annotations; ann != nil {
		if _, ok := ann["cobra_annotation_bash_completion_one_required_flag"]; ok {
			t.Error("--description should not be required")
		}
	}
}

// TestNewAgentRegisterCmd_Long verifies example commands appear in Long.
func TestNewAgentRegisterCmd_Long(t *testing.T) {
	cmd := NewAgentRegisterCmd()

	if !strings.Contains(cmd.Long, "stackeye agent register") {
		t.Error("expected Long description to contain example commands")
	}
	if !strings.Contains(cmd.Long, "--dry-run") {
		t.Error("expected Long description to mention --dry-run option")
	}
	if !strings.Contains(cmd.Long, "-o json") {
		t.Error("expected Long description to mention JSON output option")
	}
}

// TestPrintAgentRegistered_DoesNotPanic verifies printAgentRegistered is panic-safe.
func TestPrintAgentRegistered_DoesNotPanic(t *testing.T) {
	desc := "Production web server"
	agent := client.Agent{
		ID:          "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		Name:        "prod-web-01",
		Description: &desc,
		KeyPrefix:   "se_ag_ab12",
		IsActive:    true,
		CreatedAt:   "2026-01-01T00:00:00Z",
		UpdatedAt:   "2026-01-01T00:00:00Z",
	}

	tests := []struct {
		name     string
		response *client.AgentRegisterResponse
	}{
		{
			name: "newly registered agent",
			response: &client.AgentRegisterResponse{
				Status: "success",
				Data: struct {
					Agent  client.Agent `json:"agent"`
					APIKey string       `json:"api_key"`
				}{
					Agent:  agent,
					APIKey: "se_ag_deadbeef1234567890abcdef",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printAgentRegistered panicked: %v", r)
				}
			}()
			printAgentRegistered(tt.response)
		})
	}
}
