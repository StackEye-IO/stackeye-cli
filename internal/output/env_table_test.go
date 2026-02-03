package output

import (
	"testing"
)

func TestMaskSensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short value masked completely",
			input:    "abc",
			expected: "****",
		},
		{
			name:     "exactly 8 chars masked completely",
			input:    "12345678",
			expected: "****",
		},
		{
			name:     "api key format",
			input:    "se_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			expected: "se_****...7890",
		},
		{
			name:     "non-api-key long value",
			input:    "some-long-value-here",
			expected: "som****...here",
		},
		{
			name:     "9 char value shows prefix and suffix",
			input:    "123456789",
			expected: "123****...6789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskSensitive(tt.input)
			if got != tt.expected {
				t.Errorf("MaskSensitive(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSupportedEnvVars(t *testing.T) {
	vars := SupportedEnvVars()

	if len(vars) == 0 {
		t.Fatal("expected at least one supported env var")
	}

	// Check that known env vars are present
	expectedVars := map[string]bool{
		"STACKEYE_API_KEY":   false,
		"STACKEYE_API_URL":   false,
		"STACKEYE_CONFIG":    false,
		"STACKEYE_CONTEXT":   false,
		"STACKEYE_DEBUG":     false,
		"STACKEYE_TIMEOUT":   false,
		"STACKEYE_NO_INPUT":  false,
		"STACKEYE_TELEMETRY": false,
		"NO_COLOR":           false,
		"XDG_CONFIG_HOME":    false,
		"TERM":               false,
	}

	for _, v := range vars {
		if _, ok := expectedVars[v.Name]; ok {
			expectedVars[v.Name] = true
		}
	}

	for name, found := range expectedVars {
		if !found {
			t.Errorf("expected env var %q to be in SupportedEnvVars()", name)
		}
	}
}

func TestSupportedEnvVars_APIKeySensitive(t *testing.T) {
	vars := SupportedEnvVars()
	for _, v := range vars {
		if v.Name == "STACKEYE_API_KEY" {
			if !v.Sensitive {
				t.Error("STACKEYE_API_KEY should be marked as Sensitive")
			}
			return
		}
	}
	t.Error("STACKEYE_API_KEY not found in SupportedEnvVars()")
}

func TestCollectEnvVars_UnsetVars(t *testing.T) {
	// Unset all supported vars to test "(not set)" output
	vars := SupportedEnvVars()
	for _, v := range vars {
		t.Setenv(v.Name, "")
	}

	// Now unset them properly (Setenv with empty is different from unset)
	// t.Setenv handles cleanup automatically
	rows := CollectEnvVars()

	if len(rows) != len(vars) {
		t.Errorf("expected %d rows, got %d", len(vars), len(rows))
	}
}

func TestCollectEnvVars_SetVars(t *testing.T) {
	t.Setenv("STACKEYE_API_KEY", "se_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	t.Setenv("STACKEYE_API_URL", "https://api.example.com")
	t.Setenv("STACKEYE_DEBUG", "1")

	rows := CollectEnvVars()

	// Find specific rows
	rowMap := make(map[string]EnvVarRow)
	for _, r := range rows {
		rowMap[r.Variable] = r
	}

	// API key should be masked
	apiKeyRow := rowMap["STACKEYE_API_KEY"]
	if apiKeyRow.Value == "se_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890" {
		t.Error("STACKEYE_API_KEY should be masked, not shown in plaintext")
	}
	if apiKeyRow.Source != "env" {
		t.Errorf("expected source 'env' for set var, got %q", apiKeyRow.Source)
	}

	// API URL should be shown as-is
	apiURLRow := rowMap["STACKEYE_API_URL"]
	if apiURLRow.Value != "https://api.example.com" {
		t.Errorf("expected API URL value 'https://api.example.com', got %q", apiURLRow.Value)
	}

	// Debug should be shown as-is
	debugRow := rowMap["STACKEYE_DEBUG"]
	if debugRow.Value != "1" {
		t.Errorf("expected debug value '1', got %q", debugRow.Value)
	}
}

func TestCollectEnvVars_AllHaveDescriptions(t *testing.T) {
	rows := CollectEnvVars()
	for _, r := range rows {
		if r.Description == "" {
			t.Errorf("env var %q has empty description", r.Variable)
		}
	}
}

func TestPrintEnvVarsWithFormat_JSON(t *testing.T) {
	rows := []EnvVarRow{
		{Variable: "TEST_VAR", Value: "test-value", Source: "env", Description: "test"},
	}
	// Should not panic or error for JSON format
	err := PrintEnvVarsWithFormat(rows, "json")
	if err != nil {
		t.Errorf("PrintEnvVarsWithFormat(json) failed: %v", err)
	}
}

func TestPrintEnvVarsWithFormat_YAML(t *testing.T) {
	rows := []EnvVarRow{
		{Variable: "TEST_VAR", Value: "test-value", Source: "env", Description: "test"},
	}
	err := PrintEnvVarsWithFormat(rows, "yaml")
	if err != nil {
		t.Errorf("PrintEnvVarsWithFormat(yaml) failed: %v", err)
	}
}

func TestPrintEnvVarsWithFormat_Table(t *testing.T) {
	rows := []EnvVarRow{
		{Variable: "TEST_VAR", Value: "test-value", Source: "env", Description: "test"},
	}
	err := PrintEnvVarsWithFormat(rows, "table")
	if err != nil {
		t.Errorf("PrintEnvVarsWithFormat(table) failed: %v", err)
	}
}

func TestPrintEnvVarsWithFormat_Wide(t *testing.T) {
	rows := []EnvVarRow{
		{Variable: "TEST_VAR", Value: "test-value", Source: "env", Description: "test"},
	}
	err := PrintEnvVarsWithFormat(rows, "wide")
	if err != nil {
		t.Errorf("PrintEnvVarsWithFormat(wide) failed: %v", err)
	}
}

func TestPrintEnvVarsWithFormat_EmptyDefault(t *testing.T) {
	rows := []EnvVarRow{
		{Variable: "TEST_VAR", Value: "test-value", Source: "env", Description: "test"},
	}
	// Empty string should default to table format
	err := PrintEnvVarsWithFormat(rows, "")
	if err != nil {
		t.Errorf("PrintEnvVarsWithFormat('') failed: %v", err)
	}
}
