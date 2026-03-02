// Package cmd implements the CLI commands for StackEye.
package cmd

import (
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
