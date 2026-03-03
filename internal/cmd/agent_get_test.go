// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewAgentGetCmd verifies the get command is correctly configured.
func TestNewAgentGetCmd(t *testing.T) {
	cmd := NewAgentGetCmd()

	if cmd.Use != "get" {
		t.Errorf("expected Use='get', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewAgentGetCmd_RequiredFlag verifies --id is required.
func TestNewAgentGetCmd_RequiredFlag(t *testing.T) {
	cmd := NewAgentGetCmd()

	flag := cmd.Flags().Lookup("id")
	if flag == nil {
		t.Fatal("expected --id flag to be defined")
	}

	annotations := flag.Annotations
	if annotations == nil {
		t.Fatal("expected flag annotations to be set")
	}
	if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
		t.Error("expected --id to be marked as required")
	}
}

// TestNewAgentGetCmd_Aliases verifies aliases are set.
func TestNewAgentGetCmd_Aliases(t *testing.T) {
	cmd := NewAgentGetCmd()

	wantAliases := []string{"show", "status"}
	for _, want := range wantAliases {
		found := false
		for _, a := range cmd.Aliases {
			if a == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected alias %q to be set", want)
		}
	}
}

// TestNewAgentGetCmd_Long verifies example commands appear in Long.
func TestNewAgentGetCmd_Long(t *testing.T) {
	cmd := NewAgentGetCmd()

	if !strings.Contains(cmd.Long, "stackeye agent get") {
		t.Error("expected Long description to contain example commands")
	}
	if !strings.Contains(cmd.Long, "-o json") {
		t.Error("expected Long description to mention JSON output option")
	}
}

// TestPrintAgentDetail_DoesNotPanic verifies printAgentDetail is panic-safe.
func TestPrintAgentDetail_DoesNotPanic(t *testing.T) {
	desc := "Production web server"
	host := "prod-web-01"
	ip := "10.0.0.5"
	ver := "1.2.0"
	lastSeen := "2026-01-01T00:00:00Z"

	tests := []struct {
		name  string
		agent client.Agent
	}{
		{
			name: "active agent with all fields",
			agent: client.Agent{
				ID:           "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				Name:         "prod-web-01",
				Description:  &desc,
				KeyPrefix:    "se_ag_ab12",
				LastSeenAt:   &lastSeen,
				AgentVersion: &ver,
				Hostname:     &host,
				IPAddress:    &ip,
				IsActive:     true,
				CreatedAt:    "2026-01-01T00:00:00Z",
				UpdatedAt:    "2026-01-01T00:00:00Z",
			},
		},
		{
			name: "inactive agent with minimal fields",
			agent: client.Agent{
				ID:        "b2c3d4e5-f6a7-8901-bcde-f12345678901",
				Name:      "old-agent",
				KeyPrefix: "se_ag_cd34",
				IsActive:  false,
				CreatedAt: "2026-01-01T00:00:00Z",
				UpdatedAt: "2026-01-01T00:00:00Z",
			},
		},
		{
			name: "pending agent (never seen)",
			agent: client.Agent{
				ID:        "c3d4e5f6-a7b8-9012-cdef-012345678902",
				Name:      "new-agent",
				KeyPrefix: "se_ag_ef56",
				IsActive:  true,
				CreatedAt: "2026-01-01T00:00:00Z",
				UpdatedAt: "2026-01-01T00:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printAgentDetail panicked: %v", r)
				}
			}()
			printAgentDetail(&tt.agent)
		})
	}
}
