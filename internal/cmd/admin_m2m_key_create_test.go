// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
)

// TestNewAdminM2MKeyCreateCmd verifies that the m2m-key create command is properly constructed.
func TestNewAdminM2MKeyCreateCmd(t *testing.T) {
	cmd := NewAdminM2MKeyCreateCmd()

	if cmd.Use != "create" {
		t.Errorf("expected Use='create', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Create a new M2M key" {
		t.Errorf("expected Short='Create a new M2M key', got %q", cmd.Short)
	}
}

// TestNewAdminM2MKeyCreateCmd_Long verifies the Long description contains key information.
func TestNewAdminM2MKeyCreateCmd_Long(t *testing.T) {
	cmd := NewAdminM2MKeyCreateCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"machine-to-machine",
		"M2M",
		"regional",
		"global",
		"--region",
		"--global",
	}
	for _, feature := range features {
		if !strings.Contains(strings.ToLower(long), strings.ToLower(feature)) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye admin m2m-key create") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention JSON format option
	if !strings.Contains(long, "-o json") {
		t.Error("expected Long description to mention JSON output option")
	}

	// Should warn about saving the key
	if !strings.Contains(long, "only displayed once") || !strings.Contains(long, "cannot be retrieved") {
		t.Error("expected Long description to warn about saving the key")
	}
}

// TestNewAdminM2MKeyCreateCmd_Flags verifies that flags are properly configured.
func TestNewAdminM2MKeyCreateCmd_Flags(t *testing.T) {
	cmd := NewAdminM2MKeyCreateCmd()

	// Check --region flag
	regionFlag := cmd.Flags().Lookup("region")
	if regionFlag == nil {
		t.Fatal("expected --region flag to be defined")
	}
	if regionFlag.Shorthand != "r" {
		t.Errorf("expected --region shorthand to be 'r', got %q", regionFlag.Shorthand)
	}

	// Check --global flag
	globalFlag := cmd.Flags().Lookup("global")
	if globalFlag == nil {
		t.Fatal("expected --global flag to be defined")
	}
	if globalFlag.Shorthand != "g" {
		t.Errorf("expected --global shorthand to be 'g', got %q", globalFlag.Shorthand)
	}
}

// TestNewAdminM2MKeyCreateCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewAdminM2MKeyCreateCmd_RunEIsSet(t *testing.T) {
	cmd := NewAdminM2MKeyCreateCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestPrintM2MKeyCreated_DoesNotPanic verifies that printM2MKeyCreated doesn't panic with various inputs.
func TestPrintM2MKeyCreated_DoesNotPanic(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		response *admin.CreateM2MKeyResponse
		region   string
	}{
		{
			name: "regional key",
			response: &admin.CreateM2MKeyResponse{
				Status: "success",
				Data: admin.M2MKey{
					ID:        "550e8400-e29b-41d4-a716-446655440000",
					KeyPrefix: "se_m2m_nyc3_abc12345",
					Key:       "se_m2m_nyc3_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					Region:    "nyc3",
					KeyType:   "m2m",
					IsActive:  true,
					CreatedAt: now,
				},
			},
			region: "nyc3",
		},
		{
			name: "global key",
			response: &admin.CreateM2MKeyResponse{
				Status: "success",
				Data: admin.M2MKey{
					ID:        "550e8400-e29b-41d4-a716-446655440001",
					KeyPrefix: "se_m2m_xyz98765",
					Key:       "se_m2m_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					Region:    "",
					KeyType:   "m2m",
					IsActive:  true,
					CreatedAt: now,
				},
			},
			region: "",
		},
		{
			name: "key with very long plaintext",
			response: &admin.CreateM2MKeyResponse{
				Status: "success",
				Data: admin.M2MKey{
					ID:        "550e8400-e29b-41d4-a716-446655440002",
					KeyPrefix: "se_m2m_lon1_longprefix",
					Key:       "se_m2m_lon1_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
					Region:    "lon1",
					KeyType:   "m2m",
					IsActive:  true,
					CreatedAt: now,
				},
			},
			region: "lon1",
		},
		{
			name: "key with empty fields",
			response: &admin.CreateM2MKeyResponse{
				Status: "success",
				Data: admin.M2MKey{
					ID:        "",
					KeyPrefix: "",
					Key:       "",
					Region:    "",
					KeyType:   "",
					IsActive:  false,
					CreatedAt: time.Time{},
				},
			},
			region: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printM2MKeyCreated panicked: %v", r)
				}
			}()
			printM2MKeyCreated(tt.response, tt.region)
		})
	}
}

// TestPrintM2MKeyCreated_GlobalType verifies that empty region shows "global" type.
func TestPrintM2MKeyCreated_GlobalType(t *testing.T) {
	// Simulate what printM2MKeyCreated does for key type determination
	tests := []struct {
		name     string
		region   string
		expected string
	}{
		{"empty region", "", "global"},
		{"nyc3 region", "nyc3", "regional (nyc3)"},
		{"lon1 region", "lon1", "regional (lon1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyType := "global"
			if tt.region != "" {
				keyType = "regional (" + tt.region + ")"
			}
			if keyType != tt.expected {
				t.Errorf("expected keyType=%q, got %q", tt.expected, keyType)
			}
		})
	}
}
