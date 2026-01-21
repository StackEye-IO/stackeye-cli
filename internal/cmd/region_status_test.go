// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewRegionStatusCmd verifies that the region status command is properly constructed.
func TestNewRegionStatusCmd(t *testing.T) {
	cmd := NewRegionStatusCmd()

	if cmd.Use != "status" {
		t.Errorf("expected Use to be 'status', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify Long description contains key information
	if !strings.Contains(cmd.Long, "health status") {
		t.Error("expected Long description to mention 'health status'")
	}

	// Verify aliases
	expectedAliases := []string{"health"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	} else {
		for i, alias := range expectedAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("expected alias[%d] to be %q, got %q", i, alias, cmd.Aliases[i])
			}
		}
	}
}

// TestNewRegionStatusCmd_HelpContainsStatusValues verifies that status values are documented.
func TestNewRegionStatusCmd_HelpContainsStatusValues(t *testing.T) {
	cmd := NewRegionStatusCmd()

	// Verify all status values are documented
	statusValues := []string{"active", "maintenance", "disabled"}
	for _, status := range statusValues {
		if !strings.Contains(cmd.Long, status) {
			t.Errorf("expected Long description to document status value %q", status)
		}
	}
}

// TestNewRegionStatusCmd_HelpContainsHealthIndicators verifies that health indicators are documented.
func TestNewRegionStatusCmd_HelpContainsHealthIndicators(t *testing.T) {
	cmd := NewRegionStatusCmd()

	// Verify all health indicators are documented
	healthIndicators := []string{"healthy", "warning", "degraded", "unknown"}
	for _, health := range healthIndicators {
		if !strings.Contains(cmd.Long, health) {
			t.Errorf("expected Long description to document health indicator %q", health)
		}
	}
}

// TestNewRegionStatusCmd_HelpContainsExamples verifies that usage examples are provided.
func TestNewRegionStatusCmd_HelpContainsExamples(t *testing.T) {
	cmd := NewRegionStatusCmd()

	// Verify examples are present
	examples := []string{
		"stackeye region status",
		"--region",
		"-o json",
		"-o wide",
	}
	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}

// TestNewRegionStatusCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewRegionStatusCmd_RunEIsSet(t *testing.T) {
	cmd := NewRegionStatusCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewRegionStatusCmd_RegionFlagExists verifies that the region flag is defined.
func TestNewRegionStatusCmd_RegionFlagExists(t *testing.T) {
	cmd := NewRegionStatusCmd()

	flag := cmd.Flags().Lookup("region")
	if flag == nil {
		t.Error("expected --region flag to be defined")
	}

	if flag != nil && flag.DefValue != "" {
		t.Errorf("expected --region flag to have empty default, got %q", flag.DefValue)
	}
}

// TestNewRegionStatusCmd_RegionFlagUsage verifies that the region flag has a description.
func TestNewRegionStatusCmd_RegionFlagUsage(t *testing.T) {
	cmd := NewRegionStatusCmd()

	flag := cmd.Flags().Lookup("region")
	if flag == nil {
		t.Fatal("expected --region flag to be defined")
	}

	if flag.Usage == "" {
		t.Error("expected --region flag to have usage description")
	}

	if !strings.Contains(flag.Usage, "region") {
		t.Error("expected --region flag usage to mention 'region'")
	}
}

// TestRegionStatusTimeout verifies that the timeout constant is reasonable.
func TestRegionStatusTimeout(t *testing.T) {
	// Timeout should be at least 30 seconds for network operations
	if regionStatusTimeout.Seconds() < 30 {
		t.Errorf("expected regionStatusTimeout to be at least 30 seconds, got %v", regionStatusTimeout)
	}

	// Timeout should not exceed 2 minutes
	if regionStatusTimeout.Seconds() > 120 {
		t.Errorf("expected regionStatusTimeout to not exceed 2 minutes, got %v", regionStatusTimeout)
	}
}
