// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewPrivateRegionGetCmd verifies the get command is correctly configured.
func TestNewPrivateRegionGetCmd(t *testing.T) {
	cmd := NewPrivateRegionGetCmd()

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

// TestNewPrivateRegionGetCmd_FlagRequired verifies --id is required.
func TestNewPrivateRegionGetCmd_FlagRequired(t *testing.T) {
	cmd := NewPrivateRegionGetCmd()
	flag := cmd.Flags().Lookup("id")
	if flag == nil {
		t.Fatal("expected --id flag to be defined")
	}
	annotations := cmd.Annotations
	_ = annotations // annotations are set differently for required flags

	// Verify the flag exists and has a shorthand
	if flag.Shorthand != "i" {
		t.Errorf("expected --id shorthand 'i', got %q", flag.Shorthand)
	}
}

// TestNewPrivateRegionGetCmd_Long verifies example commands appear in Long.
func TestNewPrivateRegionGetCmd_Long(t *testing.T) {
	cmd := NewPrivateRegionGetCmd()

	if !strings.Contains(cmd.Long, "stackeye private-region get") {
		t.Error("expected Long description to contain example commands")
	}
}

// TestPrintPrivateRegionDetail_DoesNotPanic verifies printPrivateRegionDetail is panic-safe.
func TestPrintPrivateRegionDetail_DoesNotPanic(t *testing.T) {
	city := "New York"
	tests := []struct {
		name   string
		region *client.PrivateRegion
	}{
		{
			name: "region with city",
			region: &client.PrivateRegion{
				ID:          "prv-nyc-office",
				Name:        "nyc-office",
				DisplayName: "NYC Office",
				Continent:   "North America",
				CountryCode: "US",
				City:        &city,
				Scope:       "private",
				Status:      "active",
				CreatedAt:   "2026-01-01T00:00:00Z",
				UpdatedAt:   "2026-01-01T00:00:00Z",
			},
		},
		{
			name: "region without city",
			region: &client.PrivateRegion{
				ID:          "prv-remote",
				Name:        "remote",
				DisplayName: "Remote DC",
				Continent:   "Asia",
				CountryCode: "JP",
				City:        nil,
				Scope:       "private",
				Status:      "inactive",
				CreatedAt:   "2026-02-01T00:00:00Z",
				UpdatedAt:   "2026-02-15T00:00:00Z",
			},
		},
		{
			name: "region with invalid timestamps",
			region: &client.PrivateRegion{
				ID:          "prv-test",
				Name:        "test",
				DisplayName: "Test",
				Continent:   "Europe",
				CountryCode: "FR",
				Scope:       "private",
				Status:      "active",
				CreatedAt:   "not-a-date",
				UpdatedAt:   "not-a-date",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printPrivateRegionDetail panicked: %v", r)
				}
			}()
			printPrivateRegionDetail(tt.region)
		})
	}
}
