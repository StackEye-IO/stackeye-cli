// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

func TestNewProbeDepsWizardCmd(t *testing.T) {
	cmd := NewProbeDepsWizardCmd()

	if cmd.Use != "wizard" {
		t.Errorf("expected Use to be 'wizard', got %q", cmd.Use)
	}

	if !strings.Contains(cmd.Short, "Interactive") {
		t.Errorf("expected Short to contain 'Interactive', got %q", cmd.Short)
	}

	// Verify the command exists and has expected behavior
	if cmd.RunE == nil {
		t.Error("expected RunE to be defined")
	}
}

func TestProbeDepsWizardCmd_HelpText(t *testing.T) {
	cmd := NewProbeDepsWizardCmd()

	// Verify help contains key information
	help := cmd.Long

	expectedTerms := []string{
		"infrastructure",
		"application",
		"dependencies",
		"hierarchical alerting",
	}

	for _, term := range expectedTerms {
		if !strings.Contains(strings.ToLower(help), term) {
			t.Errorf("expected Long help to mention %q", term)
		}
	}
}

func TestRunProbeDepsWizardNonInteractive(t *testing.T) {
	// Test the non-interactive mode directly
	// Create a minimal test to verify the function exists and returns expected output
	err := runProbeDepsWizardNonInteractive()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// The function prints to stdout, so we just verify it doesn't error
}

func TestInfrastructurePatterns(t *testing.T) {
	// Test that the infrastructure pattern matches expected probe names
	// Note: The pattern uses simple substring matching, so names like
	// "database-main" will match because "main" is in the infrastructure pattern.
	// In the wizard, both patterns are checked and the name classification
	// depends on which pattern is checked first.
	cases := []struct {
		name  string
		match bool
	}{
		{"router-primary", true},
		{"firewall-01", true},
		{"gateway-west", true},
		{"vpn-server", true},
		{"dns-resolver", true},
		{"network-core", true},
		{"load-balancer", true},
		{"lb-frontend", true},
		{"haproxy-main", true},
		{"nginx-proxy", true},
		{"traefik-ingress", true},
		{"api-server", false},       // Contains "server" but not infra keywords
		{"web-frontend", false},     // No infrastructure keywords
		{"my-random-probe", false},  // No keywords at all
		{"endpoint-check", false},   // No infrastructure keywords
	}

	for _, tc := range cases {
		matched := infrastructurePatterns.MatchString(tc.name)
		if matched != tc.match {
			t.Errorf("infrastructurePatterns.MatchString(%q) = %v, want %v", tc.name, matched, tc.match)
		}
	}
}

func TestApplicationPatterns(t *testing.T) {
	// Test that the application pattern matches expected probe names
	cases := []struct {
		name  string
		match bool
	}{
		{"api-server", true},
		{"app-backend", true},
		{"web-frontend", true},
		{"user-service", true},
		{"database-01", true},
		{"db-primary", true},
		{"cache-redis", true},
		{"redis-cluster", true},
		{"postgres-main", true},
		{"mysql-replica", true},
		{"mongo-shard1", true},
		{"backend-api", true},
		{"frontend-web", true},
		{"endpoint-api", true},
		{"router-probe", false},     // Infrastructure, no app keywords
		{"firewall-01", false},      // Infrastructure, no app keywords
		{"gateway-west", false},     // Infrastructure, no app keywords
		{"my-random-probe", false},  // No keywords at all
	}

	for _, tc := range cases {
		matched := applicationPatterns.MatchString(tc.name)
		if matched != tc.match {
			t.Errorf("applicationPatterns.MatchString(%q) = %v, want %v", tc.name, matched, tc.match)
		}
	}
}

func TestParseSelectedProbeIDs(t *testing.T) {
	probe1ID := uuid.New()
	probe2ID := uuid.New()
	probe3ID := uuid.New()

	probes := []client.Probe{
		{ID: probe1ID, Name: "Probe 1"},
		{ID: probe2ID, Name: "Probe 2"},
		{ID: probe3ID, Name: "Probe 3"},
	}

	// Test parsing selected options
	selected := []string{
		"Probe 1 (" + probe1ID.String() + ")",
		"Probe 3 (" + probe3ID.String() + ")",
	}

	ids := parseSelectedProbeIDs(selected, probes)

	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs, got %d", len(ids))
	}

	// Verify the correct IDs were parsed
	found1, found3 := false, false
	for _, id := range ids {
		if id == probe1ID {
			found1 = true
		}
		if id == probe3ID {
			found3 = true
		}
	}

	if !found1 {
		t.Error("expected probe1ID to be in parsed results")
	}
	if !found3 {
		t.Error("expected probe3ID to be in parsed results")
	}
}

func TestParseSelectedProbeIDs_Empty(t *testing.T) {
	probes := []client.Probe{
		{ID: uuid.New(), Name: "Probe 1"},
	}

	ids := parseSelectedProbeIDs([]string{}, probes)

	if len(ids) != 0 {
		t.Errorf("expected 0 IDs for empty selection, got %d", len(ids))
	}
}

func TestParseSelectedProbeIDs_NoMatch(t *testing.T) {
	probes := []client.Probe{
		{ID: uuid.New(), Name: "Probe 1"},
	}

	// Non-matching selection
	ids := parseSelectedProbeIDs([]string{"Unknown Probe (some-id)"}, probes)

	if len(ids) != 0 {
		t.Errorf("expected 0 IDs for non-matching selection, got %d", len(ids))
	}
}

func TestGetProbeNameByID(t *testing.T) {
	probe1ID := uuid.New()
	probe2ID := uuid.New()

	probes := []client.Probe{
		{ID: probe1ID, Name: "Router"},
		{ID: probe2ID, Name: "Database"},
	}

	// Test finding existing probe
	name := getProbeNameByID(probe1ID, probes)
	if name != "Router" {
		t.Errorf("expected 'Router', got %q", name)
	}

	// Test finding another probe
	name = getProbeNameByID(probe2ID, probes)
	if name != "Database" {
		t.Errorf("expected 'Database', got %q", name)
	}

	// Test non-existent probe returns ID string
	unknownID := uuid.New()
	name = getProbeNameByID(unknownID, probes)
	if name != unknownID.String() {
		t.Errorf("expected ID string for unknown probe, got %q", name)
	}
}

func TestParseUUID(t *testing.T) {
	// Test valid UUID
	validID := uuid.New()
	parsed := parseUUID(validID.String())
	if parsed != validID {
		t.Errorf("expected %s, got %s", validID, parsed)
	}
}

func TestParseUUID_Invalid(t *testing.T) {
	// Test that invalid UUID panics (as expected for internal validation)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid UUID")
		}
	}()

	parseUUID("not-a-valid-uuid")
}
