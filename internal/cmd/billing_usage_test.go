// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewBillingUsageCmd verifies that the billing usage command is properly constructed.
func TestNewBillingUsageCmd(t *testing.T) {
	cmd := NewBillingUsageCmd()

	if cmd.Use != "usage" {
		t.Errorf("expected Use='usage', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Show current resource usage against plan limits" {
		t.Errorf("expected Short='Show current resource usage against plan limits', got %q", cmd.Short)
	}
}

// TestNewBillingUsageCmd_Aliases verifies that aliases are set correctly.
func TestNewBillingUsageCmd_Aliases(t *testing.T) {
	cmd := NewBillingUsageCmd()

	expectedAliases := []string{"metrics", "stats"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, alias := range expectedAliases {
		if !slices.Contains(cmd.Aliases, alias) {
			t.Errorf("expected alias %q not found", alias)
		}
	}
}

// TestNewBillingUsageCmd_Long verifies the Long description contains key information.
func TestNewBillingUsageCmd_Long(t *testing.T) {
	cmd := NewBillingUsageCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"Monitor usage",
		"Team member usage",
		"Probe checks",
		"billing period",
	}
	for _, feature := range features {
		if !strings.Contains(long, feature) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye billing usage") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention output formats
	if !strings.Contains(long, "json") {
		t.Error("expected Long description to mention JSON output option")
	}
}

// TestNewBillingUsageCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewBillingUsageCmd_RunEIsSet(t *testing.T) {
	cmd := NewBillingUsageCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestCalculateUsagePercent verifies percentage calculation.
func TestCalculateUsagePercent(t *testing.T) {
	tests := []struct {
		name     string
		used     int
		limit    int
		expected float64
	}{
		{"zero usage", 0, 100, 0},
		{"half usage", 50, 100, 50},
		{"full usage", 100, 100, 100},
		{"over limit", 150, 100, 150},
		{"zero limit", 50, 0, 0},
		{"negative limit", 50, -10, 0},
		{"small numbers", 1, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateUsagePercent(tt.used, tt.limit)
			if result != tt.expected {
				t.Errorf("calculateUsagePercent(%d, %d) = %f, want %f", tt.used, tt.limit, result, tt.expected)
			}
		})
	}
}

// TestFormatUsageBar verifies progress bar formatting.
func TestFormatUsageBar(t *testing.T) {
	tests := []struct {
		name        string
		percent     float64
		width       int
		expectedLen int
		hasWarning  bool
	}{
		{"empty bar", 0, 10, 10, false},
		{"half full", 50, 10, 10, false},
		{"full bar", 100, 10, 10, false},
		{"over 100", 150, 10, 10, false},
		{"negative", -10, 10, 10, false},
		{"minimum width", 50, 2, 4, false},
		{"high usage warning", 95, 10, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUsageBar(tt.percent, tt.width)
			runeLen := utf8.RuneCountInString(result)
			if runeLen != tt.expectedLen {
				t.Errorf("formatUsageBar(%f, %d) rune length = %d, want %d", tt.percent, tt.width, runeLen, tt.expectedLen)
			}
			// Check it starts with [ and ends with ]
			if !strings.HasPrefix(result, "[") {
				t.Errorf("formatUsageBar should start with '[', got %q", result)
			}
			if !strings.HasSuffix(result, "]") {
				t.Errorf("formatUsageBar should end with ']', got %q", result)
			}
			// Check warning character for high usage
			if tt.hasWarning && !strings.Contains(result, "▓") {
				t.Errorf("formatUsageBar at %f%% should contain warning character ▓", tt.percent)
			}
		})
	}
}

// TestFormatLargeNumber verifies thousand separator formatting.
func TestFormatLargeNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero", 0, "0"},
		{"single digit", 5, "5"},
		{"three digits", 123, "123"},
		{"four digits", 1234, "1,234"},
		{"six digits", 123456, "123,456"},
		{"seven digits", 1234567, "1,234,567"},
		{"millions", 12345678, "12,345,678"},
		{"negative", -1234, "-1,234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatLargeNumber(tt.input)
			if result != tt.expected {
				t.Errorf("formatLargeNumber(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPadRight verifies padding function.
func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"shorter string", "hi", 5, "hi   "},
		{"exact length", "hello", 5, "hello"},
		{"longer string", "hello world", 5, "hello world"},
		{"empty string", "", 3, "   "},
		{"zero width", "hi", 0, "hi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padRight(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.width, result, tt.expected)
			}
		})
	}
}

// TestPrintUsageInfo_DoesNotPanic verifies that printUsageInfo doesn't panic with various inputs.
func TestPrintUsageInfo_DoesNotPanic(t *testing.T) {
	tests := []struct {
		name  string
		usage *client.UsageInfo
	}{
		{
			name: "minimal usage",
			usage: &client.UsageInfo{
				MonitorsCount:    0,
				MonitorsLimit:    10,
				TeamMembersCount: 0,
				TeamMembersLimit: 5,
				ChecksCount:      0,
				PeriodStart:      "2026-01-01T00:00:00Z",
				PeriodEnd:        "2026-02-01T00:00:00Z",
			},
		},
		{
			name: "normal usage",
			usage: &client.UsageInfo{
				MonitorsCount:    45,
				MonitorsLimit:    100,
				TeamMembersCount: 8,
				TeamMembersLimit: 25,
				ChecksCount:      125000,
				PeriodStart:      "2026-01-01T00:00:00Z",
				PeriodEnd:        "2026-02-01T00:00:00Z",
			},
		},
		{
			name: "high usage warning",
			usage: &client.UsageInfo{
				MonitorsCount:    95,
				MonitorsLimit:    100,
				TeamMembersCount: 24,
				TeamMembersLimit: 25,
				ChecksCount:      9999999,
				PeriodStart:      "2026-01-01T00:00:00Z",
				PeriodEnd:        "2026-02-01T00:00:00Z",
			},
		},
		{
			name: "at limit",
			usage: &client.UsageInfo{
				MonitorsCount:    100,
				MonitorsLimit:    100,
				TeamMembersCount: 25,
				TeamMembersLimit: 25,
				ChecksCount:      0,
				PeriodStart:      "2026-01-01T00:00:00Z",
				PeriodEnd:        "2026-02-01T00:00:00Z",
			},
		},
		{
			name: "zero limits",
			usage: &client.UsageInfo{
				MonitorsCount:    0,
				MonitorsLimit:    0,
				TeamMembersCount: 0,
				TeamMembersLimit: 0,
				ChecksCount:      0,
				PeriodStart:      "",
				PeriodEnd:        "",
			},
		},
		{
			name: "large numbers",
			usage: &client.UsageInfo{
				MonitorsCount:    500,
				MonitorsLimit:    1000,
				TeamMembersCount: 100,
				TeamMembersLimit: 500,
				ChecksCount:      1234567890,
				PeriodStart:      "2026-01-01T00:00:00Z",
				PeriodEnd:        "2026-02-01T00:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printUsageInfo panicked: %v", r)
				}
			}()
			printUsageInfo(tt.usage)
		})
	}
}

// TestNewBillingCmd_HasUsageSubcommand verifies that usage subcommand is registered.
func TestNewBillingCmd_HasUsageSubcommand(t *testing.T) {
	cmd := NewBillingCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 2 {
		t.Error("expected billing command to have at least 2 subcommands (info and usage)")
	}

	// Verify usage subcommand is registered
	found := false
	for _, sub := range subcommands {
		if sub.Use == "usage" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'usage' subcommand to be registered")
	}
}
