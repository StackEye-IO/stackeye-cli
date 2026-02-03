package errors

import (
	"strings"
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1, s2   string
		expected int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "adc", 1},
		{"abc", "dbc", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"http", "htpp", 1},
		{"http", "htp", 1},
		{"json", "jsno", 2},
		{"yaml", "ymal", 2},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			dist := levenshteinDistance(tt.s1, tt.s2)
			if dist != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, dist, tt.expected)
			}
		})
	}
}

func TestSuggestFromOptions(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		validOptions []string
		maxDistance  int
		expected     string
	}{
		{
			name:         "exact match returns empty",
			input:        "http",
			validOptions: ValidCheckTypes,
			maxDistance:  2,
			expected:     "",
		},
		{
			name:         "typo in check type",
			input:        "htpp",
			validOptions: ValidCheckTypes,
			maxDistance:  2,
			expected:     "http",
		},
		{
			name:         "typo in check type tcp",
			input:        "tpc",
			validOptions: ValidCheckTypes,
			maxDistance:  2,
			expected:     "tcp",
		},
		{
			name:         "typo in method",
			input:        "GTE",
			validOptions: ValidHTTPMethods,
			maxDistance:  2,
			expected:     "GET",
		},
		{
			name:         "typo in method PSOT",
			input:        "PSOT",
			validOptions: ValidHTTPMethods,
			maxDistance:  2,
			expected:     "POST",
		},
		{
			name:         "typo in output format",
			input:        "jsno",
			validOptions: ValidOutputFormats,
			maxDistance:  2,
			expected:     "json",
		},
		{
			name:         "typo in output format yaml",
			input:        "ymal",
			validOptions: ValidOutputFormats,
			maxDistance:  2,
			expected:     "yaml",
		},
		{
			name:         "case insensitive",
			input:        "HTTP",
			validOptions: ValidCheckTypes,
			maxDistance:  2,
			expected:     "", // Exact match (case-insensitive)
		},
		{
			name:         "too different",
			input:        "xyz",
			validOptions: ValidCheckTypes,
			maxDistance:  2,
			expected:     "",
		},
		{
			name:         "dns_resolve typo",
			input:        "dns_reslove", //nolint:misspell // intentional typo for testing
			validOptions: ValidCheckTypes,
			maxDistance:  2,
			expected:     "dns_resolve",
		},
		{
			name:         "period typo",
			input:        "7D",
			validOptions: ValidPeriods,
			maxDistance:  2,
			expected:     "", // Case-insensitive exact match
		},
		{
			name:         "period typo 2",
			input:        "24",
			validOptions: ValidPeriods,
			maxDistance:  2,
			expected:     "24h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuggestFromOptions(tt.input, tt.validOptions, tt.maxDistance)
			if result != tt.expected {
				t.Errorf("SuggestFromOptions(%q, %v, %d) = %q, want %q",
					tt.input, tt.validOptions, tt.maxDistance, result, tt.expected)
			}
		})
	}
}

func TestInvalidValueError(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		value        string
		validOptions []string
		wantSuggest  bool
		wantContains string
	}{
		{
			name:         "with suggestion",
			flagName:     "--check-type",
			value:        "htpp",
			validOptions: ValidCheckTypes,
			wantSuggest:  true,
			wantContains: "Did you mean",
		},
		{
			name:         "without suggestion",
			flagName:     "--check-type",
			value:        "xyz",
			validOptions: ValidCheckTypes,
			wantSuggest:  false,
			wantContains: "must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidValueError(tt.flagName, tt.value, tt.validOptions)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			errStr := err.Error()
			if !strings.Contains(errStr, tt.wantContains) {
				t.Errorf("error %q should contain %q", errStr, tt.wantContains)
			}

			hasSuggestion := strings.Contains(errStr, "Did you mean")
			if hasSuggestion != tt.wantSuggest {
				t.Errorf("error has suggestion: %v, want: %v", hasSuggestion, tt.wantSuggest)
			}
		})
	}
}

func TestInvalidValueWithHintError(t *testing.T) {
	err := InvalidValueWithHintError("--interval", "abc", "must be a number between 30 and 3600")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "--interval") {
		t.Errorf("error should contain flag name")
	}
	if !strings.Contains(errStr, "abc") {
		t.Errorf("error should contain invalid value")
	}
	if !strings.Contains(errStr, "must be a number") {
		t.Errorf("error should contain hint")
	}
}

func TestRequiredFlagError(t *testing.T) {
	err := RequiredFlagError("--name")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "--name") {
		t.Errorf("error should contain flag name")
	}
	if !strings.Contains(errStr, "required") {
		t.Errorf("error should mention 'required'")
	}
}

func TestRequiredArgError(t *testing.T) {
	err := RequiredArgError("probe-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "probe-id") {
		t.Errorf("error should contain arg name")
	}
	if !strings.Contains(errStr, "required") {
		t.Errorf("error should mention 'required'")
	}
}
