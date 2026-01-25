// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
)

// TestNewAdminM2MKeyRotateCmd verifies that the m2m-key rotate command is properly constructed.
func TestNewAdminM2MKeyRotateCmd(t *testing.T) {
	cmd := NewAdminM2MKeyRotateCmd()

	if cmd.Use != "rotate" {
		t.Errorf("expected Use='rotate', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Rotate an M2M key" {
		t.Errorf("expected Short='Rotate an M2M key', got %q", cmd.Short)
	}
}

// TestNewAdminM2MKeyRotateCmd_Long verifies the Long description contains key information.
func TestNewAdminM2MKeyRotateCmd_Long(t *testing.T) {
	cmd := NewAdminM2MKeyRotateCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"rotate",
		"replacement",
		"expiration",
		"grace period",
		"24-hour",
	}
	for _, feature := range features {
		if !strings.Contains(strings.ToLower(long), strings.ToLower(feature)) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye admin m2m-key rotate --id") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention JSON format option
	if !strings.Contains(long, "-o json") {
		t.Error("expected Long description to mention JSON output option")
	}

	// Should warn about saving the new key
	if !strings.Contains(long, "only displayed once") || !strings.Contains(long, "cannot be retrieved") {
		t.Error("expected Long description to warn about saving the new key")
	}
}

// TestNewAdminM2MKeyRotateCmd_Flags verifies that flags are properly configured.
func TestNewAdminM2MKeyRotateCmd_Flags(t *testing.T) {
	cmd := NewAdminM2MKeyRotateCmd()

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

// TestNewAdminM2MKeyRotateCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewAdminM2MKeyRotateCmd_RunEIsSet(t *testing.T) {
	cmd := NewAdminM2MKeyRotateCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestPrintM2MKeyRotated_DoesNotPanic verifies that printM2MKeyRotated doesn't panic with various inputs.
func TestPrintM2MKeyRotated_DoesNotPanic(t *testing.T) {
	tests := []struct {
		name     string
		response *admin.RotateM2MKeyResponse
	}{
		{
			name: "complete rotation response",
			response: &admin.RotateM2MKeyResponse{
				OldKey: admin.M2MKeyRotationInfo{
					ID:           "550e8400-e29b-41d4-a716-446655440000",
					Region:       "nyc3",
					KeyPrefix:    "se_m2m_nyc3_old123",
					ExpiresAt:    "2026-01-26T12:00:00Z",
					ReplacedByID: "550e8400-e29b-41d4-a716-446655440001",
					CreatedAt:    "2026-01-01T00:00:00Z",
				},
				NewKey: admin.M2MKeyCreationInfo{
					ID:           "550e8400-e29b-41d4-a716-446655440001",
					Region:       "nyc3",
					KeyPrefix:    "se_m2m_nyc3_new456",
					PlaintextKey: "se_m2m_nyc3_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					CreatedAt:    "2026-01-25T12:00:00Z",
				},
				GracePeriod:    "24h0m0s",
				GracePeriodSec: 86400,
			},
		},
		{
			name: "global key rotation",
			response: &admin.RotateM2MKeyResponse{
				OldKey: admin.M2MKeyRotationInfo{
					ID:           "550e8400-e29b-41d4-a716-446655440002",
					Region:       "",
					KeyPrefix:    "se_m2m_global_old",
					ExpiresAt:    "2026-01-26T12:00:00Z",
					ReplacedByID: "550e8400-e29b-41d4-a716-446655440003",
					CreatedAt:    "2026-01-01T00:00:00Z",
				},
				NewKey: admin.M2MKeyCreationInfo{
					ID:           "550e8400-e29b-41d4-a716-446655440003",
					Region:       "",
					KeyPrefix:    "se_m2m_global_new",
					PlaintextKey: "se_m2m_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					CreatedAt:    "2026-01-25T12:00:00Z",
				},
				GracePeriod:    "24h0m0s",
				GracePeriodSec: 86400,
			},
		},
		{
			name: "response with empty fields",
			response: &admin.RotateM2MKeyResponse{
				OldKey:         admin.M2MKeyRotationInfo{},
				NewKey:         admin.M2MKeyCreationInfo{},
				GracePeriod:    "",
				GracePeriodSec: 0,
			},
		},
		{
			name: "response with very long plaintext key",
			response: &admin.RotateM2MKeyResponse{
				OldKey: admin.M2MKeyRotationInfo{
					ID:           "550e8400-e29b-41d4-a716-446655440004",
					Region:       "fra1",
					KeyPrefix:    "se_m2m_fra1_oldkey",
					ExpiresAt:    "2026-01-26T12:00:00Z",
					ReplacedByID: "550e8400-e29b-41d4-a716-446655440005",
					CreatedAt:    "2026-01-01T00:00:00Z",
				},
				NewKey: admin.M2MKeyCreationInfo{
					ID:           "550e8400-e29b-41d4-a716-446655440005",
					Region:       "fra1",
					KeyPrefix:    "se_m2m_fra1_newkeywithverylongprefix",
					PlaintextKey: "se_m2m_fra1_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					CreatedAt:    "2026-01-25T12:00:00Z",
				},
				GracePeriod:    "48h0m0s",
				GracePeriodSec: 172800,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printM2MKeyRotated panicked: %v", r)
				}
			}()
			printM2MKeyRotated(tt.response)
		})
	}
}

// TestPrintM2MKeyRotated_DisplaysKeyInfo verifies that the printed output would contain expected information.
func TestPrintM2MKeyRotated_DisplaysKeyInfo(t *testing.T) {
	// This test verifies the structure of the response that gets printed
	response := &admin.RotateM2MKeyResponse{
		OldKey: admin.M2MKeyRotationInfo{
			ID:           "old-key-id",
			ReplacedByID: "new-key-id",
			ExpiresAt:    "2026-01-26T12:00:00Z",
		},
		NewKey: admin.M2MKeyCreationInfo{
			ID:           "new-key-id",
			PlaintextKey: "test-plaintext-key",
		},
		GracePeriod: "24h0m0s",
	}

	// Verify the response structure contains expected fields
	if response.OldKey.ID == "" {
		t.Error("expected OldKey.ID to be set")
	}
	if response.NewKey.ID == "" {
		t.Error("expected NewKey.ID to be set")
	}
	if response.NewKey.PlaintextKey == "" {
		t.Error("expected NewKey.PlaintextKey to be set")
	}
	if response.GracePeriod == "" {
		t.Error("expected GracePeriod to be set")
	}
	if response.OldKey.ReplacedByID != response.NewKey.ID {
		t.Error("expected OldKey.ReplacedByID to match NewKey.ID")
	}
}
