// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewPrivateRegionRotateCmd verifies the rotate command is correctly configured.
func TestNewPrivateRegionRotateCmd(t *testing.T) {
	cmd := NewPrivateRegionRotateCmd()

	if cmd.Use != "rotate" {
		t.Errorf("expected Use='rotate', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewPrivateRegionRotateCmd_FlagsExist verifies --id is declared.
func TestNewPrivateRegionRotateCmd_FlagsExist(t *testing.T) {
	cmd := NewPrivateRegionRotateCmd()

	if cmd.Flags().Lookup("id") == nil {
		t.Error("expected flag --id to be defined")
	}
	if cmd.Flags().Lookup("display-name") == nil {
		t.Error("expected flag --display-name to be defined")
	}
}

// TestNewPrivateRegionRotateCmd_Long verifies example commands in Long.
func TestNewPrivateRegionRotateCmd_Long(t *testing.T) {
	cmd := NewPrivateRegionRotateCmd()

	if !strings.Contains(cmd.Long, "stackeye private-region rotate") {
		t.Error("expected Long description to contain example commands")
	}
	if !strings.Contains(strings.ToLower(cmd.Long), "plaintext") {
		t.Error("expected Long description to warn about plaintext key shown once")
	}
}

// TestPrintPrivateRegionRotated_DoesNotPanic verifies printPrivateRegionRotated is panic-safe.
func TestPrintPrivateRegionRotated_DoesNotPanic(t *testing.T) {
	plaintext := "se_prv_abc123secretkey"

	tests := []struct {
		name     string
		response *client.PrivateRegionRotateKeyResponse
	}{
		{
			name: "rotation with plaintext key",
			response: &client.PrivateRegionRotateKeyResponse{
				Status: "success",
				Data: struct {
					NewKey        client.PrivateRegionKey `json:"new_key"`
					RevokedKeyIDs []string                `json:"revoked_key_ids"`
				}{
					NewKey: client.PrivateRegionKey{
						ID:           "new-key-uuid",
						RegionID:     "prv-nyc-office",
						KeyPrefix:    "se_prv_",
						DisplayName:  "Rotated Key",
						IsActive:     true,
						CreatedAt:    "2026-03-01T00:00:00Z",
						PlaintextKey: &plaintext,
					},
					RevokedKeyIDs: []string{"old-key-uuid-1", "old-key-uuid-2"},
				},
			},
		},
		{
			name: "rotation without plaintext (not first response)",
			response: &client.PrivateRegionRotateKeyResponse{
				Status: "success",
				Data: struct {
					NewKey        client.PrivateRegionKey `json:"new_key"`
					RevokedKeyIDs []string                `json:"revoked_key_ids"`
				}{
					NewKey: client.PrivateRegionKey{
						ID:           "new-key-uuid",
						RegionID:     "prv-nyc-office",
						KeyPrefix:    "se_prv_",
						IsActive:     true,
						CreatedAt:    "2026-03-01T00:00:00Z",
						PlaintextKey: nil,
					},
					RevokedKeyIDs: []string{},
				},
			},
		},
		{
			name: "rotation with no revoked keys",
			response: &client.PrivateRegionRotateKeyResponse{
				Status: "success",
				Data: struct {
					NewKey        client.PrivateRegionKey `json:"new_key"`
					RevokedKeyIDs []string                `json:"revoked_key_ids"`
				}{
					NewKey: client.PrivateRegionKey{
						ID:           "new-key-uuid",
						IsActive:     true,
						PlaintextKey: &plaintext,
					},
					RevokedKeyIDs: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printPrivateRegionRotated panicked: %v", r)
				}
			}()
			printPrivateRegionRotated(tt.response)
		})
	}
}
