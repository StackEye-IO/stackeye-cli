// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
)

// TestNewAdminM2MKeyDeactivateCmd verifies that the m2m-key deactivate command is properly constructed.
func TestNewAdminM2MKeyDeactivateCmd(t *testing.T) {
	cmd := NewAdminM2MKeyDeactivateCmd()

	if cmd.Use != "deactivate" {
		t.Errorf("expected Use='deactivate', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Deactivate an M2M key immediately" {
		t.Errorf("expected Short='Deactivate an M2M key immediately', got %q", cmd.Short)
	}
}

// TestNewAdminM2MKeyDeactivateCmd_Long verifies the Long description contains key information.
func TestNewAdminM2MKeyDeactivateCmd_Long(t *testing.T) {
	cmd := NewAdminM2MKeyDeactivateCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"deactivate",
		"immediate",
		"invalid",
		"authentication",
		"audit",
	}
	for _, feature := range features {
		if !strings.Contains(strings.ToLower(long), strings.ToLower(feature)) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye admin m2m-key deactivate --id") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention JSON format option
	if !strings.Contains(long, "-o json") {
		t.Error("expected Long description to mention JSON output option")
	}

	// Should recommend using rotate for graceful replacement
	if !strings.Contains(strings.ToLower(long), "rotate") {
		t.Error("expected Long description to recommend using rotate for graceful replacement")
	}

	// Should note that there is no grace period
	if !strings.Contains(strings.ToLower(long), "no grace") {
		t.Error("expected Long description to note that there is no grace period")
	}
}

// TestNewAdminM2MKeyDeactivateCmd_Flags verifies that flags are properly configured.
func TestNewAdminM2MKeyDeactivateCmd_Flags(t *testing.T) {
	cmd := NewAdminM2MKeyDeactivateCmd()

	// Check --id flag
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Fatal("expected --id flag to be defined")
	}

	// Verify usage contains "required"
	if !strings.Contains(strings.ToLower(idFlag.Usage), "required") {
		t.Error("expected --id flag usage to indicate it's required")
	}

	// Verify the -i shorthand is configured
	if idFlag.Shorthand != "i" {
		t.Errorf("expected --id shorthand to be 'i', got %q", idFlag.Shorthand)
	}
}

// TestNewAdminM2MKeyDeactivateCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewAdminM2MKeyDeactivateCmd_RunEIsSet(t *testing.T) {
	cmd := NewAdminM2MKeyDeactivateCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestPrintM2MKeyDeactivated_DoesNotPanic verifies that printM2MKeyDeactivated doesn't panic with various inputs.
func TestPrintM2MKeyDeactivated_DoesNotPanic(t *testing.T) {
	tests := []struct {
		name     string
		response *admin.DeactivateM2MKeyResponse
	}{
		{
			name: "standard deactivation message",
			response: &admin.DeactivateM2MKeyResponse{
				Message: "M2M key successfully deactivated",
			},
		},
		{
			name: "empty message",
			response: &admin.DeactivateM2MKeyResponse{
				Message: "",
			},
		},
		{
			name: "long message",
			response: &admin.DeactivateM2MKeyResponse{
				Message: "M2M key 550e8400-e29b-41d4-a716-446655440000 has been successfully deactivated. The key is now invalid and cannot be used for authentication. All services using this key will fail on their next request.",
			},
		},
		{
			name: "message with special characters",
			response: &admin.DeactivateM2MKeyResponse{
				Message: "Key deactivated: se_m2m_nyc3_abc123... [ID: 550e8400]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printM2MKeyDeactivated panicked: %v", r)
				}
			}()
			printM2MKeyDeactivated(tt.response)
		})
	}
}

// TestPrintM2MKeyDeactivated_ResponseStructure verifies the response structure.
func TestPrintM2MKeyDeactivated_ResponseStructure(t *testing.T) {
	// Verify the DeactivateM2MKeyResponse has the expected structure
	response := &admin.DeactivateM2MKeyResponse{
		Message: "test message",
	}

	if response.Message != "test message" {
		t.Errorf("expected Message='test message', got %q", response.Message)
	}
}

// TestDeactivateVsRotateRecommendation verifies that the command documentation recommends rotate.
func TestDeactivateVsRotateRecommendation(t *testing.T) {
	cmd := NewAdminM2MKeyDeactivateCmd()

	// The Long description should mention 'rotate' as an alternative
	if !strings.Contains(cmd.Long, "rotate") {
		t.Error("expected Long description to mention 'rotate' as an alternative for graceful key replacement")
	}

	// Should mention that rotate provides a grace period
	if !strings.Contains(strings.ToLower(cmd.Long), "grace period") {
		t.Error("expected Long description to mention grace period in context of rotate alternative")
	}
}
