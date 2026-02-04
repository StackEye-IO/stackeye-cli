// Package output provides CLI output helpers for StackEye commands.
// Task #7170
package output

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

func TestNewAPIKeyTableFormatter(t *testing.T) {
	formatter := NewAPIKeyTableFormatter(sdkoutput.ColorNever, false)

	if formatter == nil {
		t.Fatal("expected formatter to be non-nil")
	}

	if formatter.colorMgr == nil {
		t.Error("expected colorMgr to be non-nil")
	}

	if formatter.isWide {
		t.Error("expected isWide to be false")
	}
}

func TestNewAPIKeyTableFormatter_WideMode(t *testing.T) {
	formatter := NewAPIKeyTableFormatter(sdkoutput.ColorNever, true)

	if !formatter.isWide {
		t.Error("expected isWide to be true")
	}
}

func TestAPIKeyTableFormatter_FormatAPIKey(t *testing.T) {
	formatter := NewAPIKeyTableFormatter(sdkoutput.ColorNever, false)

	lastUsed := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	expiresAt := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	keyID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	key := client.APIKey{
		ID:          keyID,
		Name:        "production-key",
		KeyPrefix:   "se_abc1",
		Permissions: "read:probes,write:probes",
		LastUsedAt:  &lastUsed,
		ExpiresAt:   &expiresAt,
		CreatedAt:   createdAt,
	}

	row := formatter.FormatAPIKey(key)

	if row.Name != "production-key" {
		t.Errorf("expected Name='production-key', got %q", row.Name)
	}

	if row.KeyPrefix != "se_abc1..." {
		t.Errorf("expected KeyPrefix='se_abc1...', got %q", row.KeyPrefix)
	}

	if row.Permissions != "read:probes,write:probes" {
		t.Errorf("expected Permissions='read:probes,write:probes', got %q", row.Permissions)
	}

	if row.LastUsed != "2026-01-15" {
		t.Errorf("expected LastUsed='2026-01-15', got %q", row.LastUsed)
	}

	if row.Expires != "2026-06-15" {
		t.Errorf("expected Expires='2026-06-15', got %q", row.Expires)
	}

	if row.Created != "2026-01-01" {
		t.Errorf("expected Created='2026-01-01', got %q", row.Created)
	}

	if row.ID != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected ID='11111111-1111-1111-1111-111111111111', got %q", row.ID)
	}
}

func TestAPIKeyTableFormatter_FormatAPIKeys(t *testing.T) {
	formatter := NewAPIKeyTableFormatter(sdkoutput.ColorNever, false)

	keys := []client.APIKey{
		{
			ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:      "key-1",
			KeyPrefix: "se_aaa",
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:      "key-2",
			KeyPrefix: "se_bbb",
			CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	rows := formatter.FormatAPIKeys(keys)

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0].Name != "key-1" {
		t.Errorf("expected first row Name='key-1', got %q", rows[0].Name)
	}

	if rows[1].Name != "key-2" {
		t.Errorf("expected second row Name='key-2', got %q", rows[1].Name)
	}
}

func TestAPIKeyTableFormatter_FormatAPIKeys_Empty(t *testing.T) {
	formatter := NewAPIKeyTableFormatter(sdkoutput.ColorNever, false)

	rows := formatter.FormatAPIKeys([]client.APIKey{})

	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}
}

func TestFormatKeyPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "normal prefix", input: "se_abc1", expected: "se_abc1..."},
		{name: "empty prefix", input: "", expected: "-"},
		{name: "short prefix", input: "se_", expected: "se_..."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatKeyPrefix(tc.input)
			if result != tc.expected {
				t.Errorf("formatKeyPrefix(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatPermissions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty", input: "", expected: "-"},
		{name: "short", input: "read:probes", expected: "read:probes"},
		{name: "exactly 30 chars", input: "123456789012345678901234567890", expected: "123456789012345678901234567890"},
		{name: "over 30 chars", input: "read:probes,write:probes,read:alerts,write:alerts", expected: "read:probes,write:probes,re..."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatPermissions(tc.input)
			if result != tc.expected {
				t.Errorf("formatPermissions(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatLastUsed(t *testing.T) {
	used := time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    *time.Time
		expected string
	}{
		{name: "nil (never used)", input: nil, expected: "Never"},
		{name: "valid time", input: &used, expected: "2026-03-15"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatLastUsed(tc.input)
			if result != tc.expected {
				t.Errorf("formatLastUsed() = %q, expected %q", result, tc.expected)
			}
		})
	}
}

func TestFormatExpires(t *testing.T) {
	expires := time.Date(2026, 12, 31, 23, 59, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    *time.Time
		expected string
	}{
		{name: "nil (no expiration)", input: nil, expected: "Never"},
		{name: "valid time", input: &expires, expected: "2026-12-31"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatExpires(tc.input)
			if result != tc.expected {
				t.Errorf("formatExpires() = %q, expected %q", result, tc.expected)
			}
		})
	}
}

func TestAPIKeyTableFormatter_NilFields(t *testing.T) {
	formatter := NewAPIKeyTableFormatter(sdkoutput.ColorNever, false)

	key := client.APIKey{
		ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:        "minimal-key",
		KeyPrefix:   "",
		Permissions: "",
		LastUsedAt:  nil,
		ExpiresAt:   nil,
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	row := formatter.FormatAPIKey(key)

	if row.KeyPrefix != "-" {
		t.Errorf("expected KeyPrefix='-' for empty prefix, got %q", row.KeyPrefix)
	}

	if row.Permissions != "-" {
		t.Errorf("expected Permissions='-' for empty perms, got %q", row.Permissions)
	}

	if row.LastUsed != "Never" {
		t.Errorf("expected LastUsed='Never' for nil, got %q", row.LastUsed)
	}

	if row.Expires != "Never" {
		t.Errorf("expected Expires='Never' for nil, got %q", row.Expires)
	}
}
