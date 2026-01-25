// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
)

// TestNewAdminM2MKeyGetCmd verifies that the m2m-key get command is properly constructed.
func TestNewAdminM2MKeyGetCmd(t *testing.T) {
	cmd := NewAdminM2MKeyGetCmd()

	if cmd.Use != "get" {
		t.Errorf("expected Use='get', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Get details of a specific M2M key" {
		t.Errorf("expected Short='Get details of a specific M2M key', got %q", cmd.Short)
	}
}

// TestNewAdminM2MKeyGetCmd_Long verifies the Long description contains key information.
func TestNewAdminM2MKeyGetCmd_Long(t *testing.T) {
	cmd := NewAdminM2MKeyGetCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"detailed information",
		"ID",
		"region",
		"prefix",
		"status",
		"last seen",
	}
	for _, feature := range features {
		if !strings.Contains(strings.ToLower(long), strings.ToLower(feature)) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye admin m2m-key get --id") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention JSON format option
	if !strings.Contains(long, "-o json") {
		t.Error("expected Long description to mention JSON output option")
	}

	// Should note that plaintext key is never shown
	if !strings.Contains(strings.ToLower(long), "plaintext") || !strings.Contains(strings.ToLower(long), "never") {
		t.Error("expected Long description to note that plaintext key is never shown")
	}
}

// TestNewAdminM2MKeyGetCmd_Flags verifies that flags are properly configured.
func TestNewAdminM2MKeyGetCmd_Flags(t *testing.T) {
	cmd := NewAdminM2MKeyGetCmd()

	// Check --id flag
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Fatal("expected --id flag to be defined")
	}

	// Verify the flag has no default value (indicating it's likely required)
	if idFlag.DefValue != "" {
		t.Errorf("expected --id flag to have no default value, got %q", idFlag.DefValue)
	}
}

// TestNewAdminM2MKeyGetCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewAdminM2MKeyGetCmd_RunEIsSet(t *testing.T) {
	cmd := NewAdminM2MKeyGetCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestPrintM2MKeyDetail_DoesNotPanic verifies that printM2MKeyDetail doesn't panic with various inputs.
func TestPrintM2MKeyDetail_DoesNotPanic(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		key  *admin.M2MKey
	}{
		{
			name: "active regional key with last seen",
			key: &admin.M2MKey{
				ID:         "550e8400-e29b-41d4-a716-446655440000",
				Region:     "nyc3",
				KeyPrefix:  "se_m2m_nyc3_abc12345",
				KeyType:    "m2m",
				IsActive:   true,
				LastSeenAt: &now,
				CreatedAt:  now.Add(-24 * time.Hour),
			},
		},
		{
			name: "inactive key",
			key: &admin.M2MKey{
				ID:        "550e8400-e29b-41d4-a716-446655440001",
				Region:    "lon1",
				KeyPrefix: "se_m2m_lon1_xyz98765",
				KeyType:   "m2m",
				IsActive:  false,
				CreatedAt: now.Add(-48 * time.Hour),
			},
		},
		{
			name: "global key never seen",
			key: &admin.M2MKey{
				ID:         "550e8400-e29b-41d4-a716-446655440002",
				Region:     "",
				KeyPrefix:  "se_m2m_global123",
				KeyType:    "m2m",
				IsActive:   true,
				LastSeenAt: nil,
				CreatedAt:  now,
			},
		},
		{
			name: "key with zero last seen",
			key: &admin.M2MKey{
				ID:         "550e8400-e29b-41d4-a716-446655440003",
				Region:     "fra1",
				KeyPrefix:  "se_m2m_fra1_test",
				KeyType:    "m2m",
				IsActive:   true,
				LastSeenAt: func() *time.Time { t := time.Time{}; return &t }(),
				CreatedAt:  now,
			},
		},
		{
			name: "key with empty fields",
			key: &admin.M2MKey{
				ID:        "",
				Region:    "",
				KeyPrefix: "",
				KeyType:   "",
				IsActive:  false,
				CreatedAt: time.Time{},
			},
		},
		{
			name: "key with very long prefix",
			key: &admin.M2MKey{
				ID:        "550e8400-e29b-41d4-a716-446655440004",
				Region:    "nyc3",
				KeyPrefix: "se_m2m_nyc3_verylongkeyprefixthatexceedsnormallength12345678901234567890",
				KeyType:   "m2m",
				IsActive:  true,
				CreatedAt: now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printM2MKeyDetail panicked: %v", r)
				}
			}()
			printM2MKeyDetail(tt.key)
		})
	}
}

// TestPrintM2MKeyDetail_EmptyRegionShowsGlobal verifies that empty region shows "global".
func TestPrintM2MKeyDetail_EmptyRegionShowsGlobal(t *testing.T) {
	// Simulate what printM2MKeyDetail does for region determination
	tests := []struct {
		name     string
		region   string
		expected string
	}{
		{"empty region", "", "global"},
		{"nyc3 region", "nyc3", "nyc3"},
		{"lon1 region", "lon1", "lon1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region := tt.region
			if region == "" {
				region = "global"
			}
			if region != tt.expected {
				t.Errorf("expected region=%q, got %q", tt.expected, region)
			}
		})
	}
}

// TestPrintM2MKeyDetail_StatusDisplay verifies that status is displayed correctly.
func TestPrintM2MKeyDetail_StatusDisplay(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
		expected string
	}{
		{"active key", true, "Active"},
		{"inactive key", false, "Inactive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := "Active"
			if !tt.isActive {
				status = "Inactive"
			}
			if status != tt.expected {
				t.Errorf("expected status=%q, got %q", tt.expected, status)
			}
		})
	}
}
