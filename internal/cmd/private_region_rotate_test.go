// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewPrivateRegionRotateCmd verifies the rotate command is correctly configured.
func TestNewPrivateRegionRotateCmd(t *testing.T) {
	cmd := NewPrivateRegionRotateCmd()

	if cmd.Use != "rotate-key" {
		t.Errorf("expected Use='rotate-key', got %q", cmd.Use)
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

	if !strings.Contains(cmd.Long, "stackeye private-region rotate-key") {
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

// TestRunPrivateRegionRotate_PlanTierRequired verifies that a 402 response
// from the API is surfaced as an error (plan tier gate enforcement).
func TestRunPrivateRegionRotate_PlanTierRequired(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/private-regions/prv-nyc-office/rotate-key" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   "PLAN_LIMIT_EXCEEDED",
			"message": "Private regions require a Team plan or higher",
		})
	})

	_, cleanup := setupMockAPIServer(t, handler)
	defer cleanup()

	err := runPrivateRegionRotate(context.Background(), "prv-nyc-office", nil)
	if err == nil {
		t.Fatal("expected error for plan_tier_required (402), got nil")
	}
	if !strings.Contains(err.Error(), "failed to rotate bootstrap key") {
		t.Errorf("expected error to contain 'failed to rotate bootstrap key', got: %v", err)
	}
}

// TestRunPrivateRegionRotate_Unauthorized verifies that a 401 response
// from the API is surfaced as an error (missing or invalid API key).
func TestRunPrivateRegionRotate_Unauthorized(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/private-regions/prv-nyc-office/rotate-key" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   "UNAUTHORIZED",
			"message": "Invalid or missing API key",
		})
	})

	_, cleanup := setupMockAPIServer(t, handler)
	defer cleanup()

	err := runPrivateRegionRotate(context.Background(), "prv-nyc-office", nil)
	if err == nil {
		t.Fatal("expected error for unauthorized (401), got nil")
	}
	if !strings.Contains(err.Error(), "failed to rotate bootstrap key") {
		t.Errorf("expected error to contain 'failed to rotate bootstrap key', got: %v", err)
	}
}
