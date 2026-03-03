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

// TestNewPrivateRegionCreateCmd verifies the create command is correctly configured.
func TestNewPrivateRegionCreateCmd(t *testing.T) {
	cmd := NewPrivateRegionCreateCmd()

	if cmd.Use != "create" {
		t.Errorf("expected Use='create', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewPrivateRegionCreateCmd_FlagsExist verifies all required flags are declared.
func TestNewPrivateRegionCreateCmd_FlagsExist(t *testing.T) {
	cmd := NewPrivateRegionCreateCmd()

	requiredFlags := []string{"slug", "display-name", "continent", "country-code"}
	for _, name := range requiredFlags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s to be defined", name)
		}
	}

	// city is optional
	if cmd.Flags().Lookup("city") == nil {
		t.Error("expected flag --city to be defined")
	}
}

// TestNewPrivateRegionCreateCmd_Long verifies example commands in Long.
func TestNewPrivateRegionCreateCmd_Long(t *testing.T) {
	cmd := NewPrivateRegionCreateCmd()

	if !strings.Contains(cmd.Long, "stackeye private-region create") {
		t.Error("expected Long description to contain example commands")
	}
	if !strings.Contains(cmd.Long, "--dry-run") {
		t.Error("expected Long description to mention --dry-run")
	}
}

// TestPrintPrivateRegionCreated_DoesNotPanic verifies printPrivateRegionCreated is panic-safe.
func TestPrintPrivateRegionCreated_DoesNotPanic(t *testing.T) {
	tests := []struct {
		name   string
		region *client.PrivateRegion
	}{
		{
			name: "newly created region",
			region: &client.PrivateRegion{
				ID:          "prv-nyc-office",
				Name:        "nyc-office",
				DisplayName: "NYC Office",
				Continent:   "North America",
				CountryCode: "US",
				Scope:       "private",
				Status:      "active",
				CreatedAt:   "2026-01-01T00:00:00Z",
				UpdatedAt:   "2026-01-01T00:00:00Z",
			},
		},
		{
			name: "region with empty display name",
			region: &client.PrivateRegion{
				ID:     "prv-test",
				Status: "active",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printPrivateRegionCreated panicked: %v", r)
				}
			}()
			printPrivateRegionCreated(tt.region)
		})
	}
}

// TestRunPrivateRegionCreate_PlanTierRequired verifies that a 402 response
// from the API is surfaced as an error (plan tier gate enforcement).
func TestRunPrivateRegionCreate_PlanTierRequired(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/private-regions" || r.Method != http.MethodPost {
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

	err := runPrivateRegionCreate(context.Background(), "nyc-office", "NYC Office", "North America", "US", nil)
	if err == nil {
		t.Fatal("expected error for plan_tier_required (402), got nil")
	}
	if !strings.Contains(err.Error(), "failed to create private region") {
		t.Errorf("expected error to contain 'failed to create private region', got: %v", err)
	}
}

// TestRunPrivateRegionCreate_Unauthorized verifies that a 401 response
// from the API is surfaced as an error (missing or invalid API key).
func TestRunPrivateRegionCreate_Unauthorized(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/private-regions" || r.Method != http.MethodPost {
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

	err := runPrivateRegionCreate(context.Background(), "nyc-office", "NYC Office", "North America", "US", nil)
	if err == nil {
		t.Fatal("expected error for unauthorized (401), got nil")
	}
	if !strings.Contains(err.Error(), "failed to create private region") {
		t.Errorf("expected error to contain 'failed to create private region', got: %v", err)
	}
}
