// Package output provides CLI output helpers for StackEye commands.
// Task #8065
package output

import (
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

func TestLabelKeyTableFormatter_FormatLabelKeys(t *testing.T) {
	formatter := NewLabelKeyTableFormatter(sdkoutput.ColorNever)

	displayName := "Environment"
	description := "Deployment environment"

	testCases := []struct {
		name     string
		input    []client.LabelKey
		expected []LabelKeyTableRow
	}{
		{
			name:     "empty slice",
			input:    []client.LabelKey{},
			expected: []LabelKeyTableRow{},
		},
		{
			name: "single label key with values",
			input: []client.LabelKey{
				{
					ID:             1,
					OrganizationID: "org-123",
					Key:            "env",
					DisplayName:    &displayName,
					Description:    &description,
					Color:          "#10B981",
					ValuesInUse:    []string{"production", "staging", "dev"},
					ProbeCount:     15,
					CreatedAt:      "2025-01-15T10:30:00Z",
					UpdatedAt:      "2025-01-15T14:20:00Z",
				},
			},
			expected: []LabelKeyTableRow{
				{
					Key:         "env",
					DisplayName: "Environment",
					Color:       "\u25CF",
					ValuesInUse: "production, staging, dev",
					Probes:      "15",
				},
			},
		},
		{
			name: "label key without values (tag-only)",
			input: []client.LabelKey{
				{
					ID:             2,
					OrganizationID: "org-123",
					Key:            "pci",
					DisplayName:    nil,
					Color:          "#6B7280",
					ValuesInUse:    []string{},
					ProbeCount:     5,
				},
			},
			expected: []LabelKeyTableRow{
				{
					Key:         "pci",
					DisplayName: "pci",
					Color:       "\u25CF",
					ValuesInUse: "(key-only)",
					Probes:      "5",
				},
			},
		},
		{
			name: "label key with many values (truncated)",
			input: []client.LabelKey{
				{
					Key:         "team",
					DisplayName: nil,
					Color:       "#3B82F6",
					ValuesInUse: []string{"alpha", "beta", "gamma", "delta", "epsilon"},
					ProbeCount:  25,
				},
			},
			expected: []LabelKeyTableRow{
				{
					Key:         "team",
					DisplayName: "team",
					Color:       "\u25CF",
					ValuesInUse: "alpha, beta, gamma +2",
					Probes:      "25",
				},
			},
		},
		{
			name: "label key with zero probes",
			input: []client.LabelKey{
				{
					Key:         "unused",
					DisplayName: nil,
					Color:       "#6B7280",
					ValuesInUse: []string{},
					ProbeCount:  0,
				},
			},
			expected: []LabelKeyTableRow{
				{
					Key:         "unused",
					DisplayName: "unused",
					Color:       "\u25CF",
					ValuesInUse: "(key-only)",
					Probes:      "0",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatter.FormatLabelKeys(tc.input)

			if len(result) != len(tc.expected) {
				t.Fatalf("expected %d rows, got %d", len(tc.expected), len(result))
			}

			for i, row := range result {
				expected := tc.expected[i]
				if row.Key != expected.Key {
					t.Errorf("row %d: expected Key=%q, got %q", i, expected.Key, row.Key)
				}
				if row.DisplayName != expected.DisplayName {
					t.Errorf("row %d: expected DisplayName=%q, got %q", i, expected.DisplayName, row.DisplayName)
				}
				if row.Color != expected.Color {
					t.Errorf("row %d: expected Color=%q, got %q", i, expected.Color, row.Color)
				}
				if row.ValuesInUse != expected.ValuesInUse {
					t.Errorf("row %d: expected ValuesInUse=%q, got %q", i, expected.ValuesInUse, row.ValuesInUse)
				}
				if row.Probes != expected.Probes {
					t.Errorf("row %d: expected Probes=%q, got %q", i, expected.Probes, row.Probes)
				}
			}
		})
	}
}

func TestFormatValues(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: "(key-only)",
		},
		{
			name:     "single value",
			input:    []string{"production"},
			expected: "production",
		},
		{
			name:     "two values",
			input:    []string{"production", "staging"},
			expected: "production, staging",
		},
		{
			name:     "three values",
			input:    []string{"production", "staging", "dev"},
			expected: "production, staging, dev",
		},
		{
			name:     "four values - truncated",
			input:    []string{"production", "staging", "dev", "test"},
			expected: "production, staging, dev +1",
		},
		{
			name:     "five values - truncated",
			input:    []string{"a", "b", "c", "d", "e"},
			expected: "a, b, c +2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatValues(tc.input)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestFormatInt64(t *testing.T) {
	testCases := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{100, "100"},
		{1234, "1234"},
		{-1, "-1"},
		{-100, "-100"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatInt64(tc.input)
			if result != tc.expected {
				t.Errorf("formatInt64(%d) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}
