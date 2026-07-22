// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func TestNewEnrollmentKeyTableFormatter(t *testing.T) {
	formatter := NewEnrollmentKeyTableFormatter(sdkoutput.ColorNever, false)
	if formatter == nil {
		t.Fatal("expected formatter to be non-nil")
	}
	if formatter.isWide {
		t.Error("expected isWide to be false")
	}
}

func TestEnrollmentKeyTableFormatter_FormatEnrollmentKey(t *testing.T) {
	formatter := NewEnrollmentKeyTableFormatter(sdkoutput.ColorNever, false)

	key := client.EnrollmentKey{
		ID:            "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		Name:          "Fleet bootstrap key",
		KeyPrefix:     "se_ek_ab12",
		CapabilitySet: []string{"host_monitoring", "private_relay"},
		Mode:          client.EnrollmentKeyModeFleet,
		Environment:   "production",
		MaxUses:       intPtr(10000),
		Uses:          42,
		ExpiresAt:     strPtr("2026-10-01T00:00:00Z"),
		CreatedAt:     "2026-07-01T00:00:00Z",
	}

	row := formatter.FormatEnrollmentKey(key)

	if row.Name != "Fleet bootstrap key" {
		t.Errorf("expected Name='Fleet bootstrap key', got %q", row.Name)
	}
	if row.KeyPrefix != "se_ek_ab12..." {
		t.Errorf("expected KeyPrefix='se_ek_ab12...', got %q", row.KeyPrefix)
	}
	if row.Mode != "fleet" {
		t.Errorf("expected Mode='fleet', got %q", row.Mode)
	}
	if row.Capabilities != "host_monitoring,private_relay" {
		t.Errorf("expected Capabilities='host_monitoring,private_relay', got %q", row.Capabilities)
	}
	if row.Uses != "42/10000" {
		t.Errorf("expected Uses='42/10000', got %q", row.Uses)
	}
	if row.Expires != "2026-10-01" {
		t.Errorf("expected Expires='2026-10-01', got %q", row.Expires)
	}
	if row.Environment != "production" {
		t.Errorf("expected Environment='production', got %q", row.Environment)
	}
	if row.Created != "2026-07-01" {
		t.Errorf("expected Created='2026-07-01', got %q", row.Created)
	}
	if row.ID != key.ID {
		t.Errorf("expected ID=%q, got %q", key.ID, row.ID)
	}
}

func TestEnrollmentKeyTableFormatter_NilFields(t *testing.T) {
	formatter := NewEnrollmentKeyTableFormatter(sdkoutput.ColorNever, false)

	key := client.EnrollmentKey{
		ID:        "key-id",
		CreatedAt: "2026-01-01T00:00:00Z",
	}

	row := formatter.FormatEnrollmentKey(key)

	if row.Name != "(unnamed)" {
		t.Errorf("expected Name='(unnamed)', got %q", row.Name)
	}
	if row.KeyPrefix != "-" {
		t.Errorf("expected KeyPrefix='-', got %q", row.KeyPrefix)
	}
	if row.Mode != "standard" {
		t.Errorf("expected Mode='standard' (default), got %q", row.Mode)
	}
	if row.Capabilities != "-" {
		t.Errorf("expected Capabilities='-', got %q", row.Capabilities)
	}
	if row.Uses != "0/unlimited" {
		t.Errorf("expected Uses='0/unlimited', got %q", row.Uses)
	}
	if row.Expires != "never" {
		t.Errorf("expected Expires='never', got %q", row.Expires)
	}
	if row.Environment != "-" {
		t.Errorf("expected Environment='-', got %q", row.Environment)
	}
}

func TestEnrollmentKeyTableFormatter_Revoked(t *testing.T) {
	formatter := NewEnrollmentKeyTableFormatter(sdkoutput.ColorNever, false)

	key := client.EnrollmentKey{
		ID:        "key-id",
		ExpiresAt: strPtr("2026-10-01T00:00:00Z"),
		RevokedAt: strPtr("2026-08-01T00:00:00Z"),
		CreatedAt: "2026-01-01T00:00:00Z",
	}

	row := formatter.FormatEnrollmentKey(key)
	if row.Expires != "revoked" {
		t.Errorf("expected Expires='revoked' to take priority over a future ExpiresAt, got %q", row.Expires)
	}
}

func TestEnrollmentKeyTableFormatter_FormatEnrollmentKeys(t *testing.T) {
	formatter := NewEnrollmentKeyTableFormatter(sdkoutput.ColorNever, false)

	keys := []client.EnrollmentKey{
		{ID: "key-1", Name: "one", CreatedAt: "2026-01-01T00:00:00Z"},
		{ID: "key-2", Name: "two", CreatedAt: "2026-01-02T00:00:00Z"},
	}

	rows := formatter.FormatEnrollmentKeys(keys)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].Name != "one" || rows[1].Name != "two" {
		t.Errorf("unexpected row order/names: %+v", rows)
	}
}

func TestEnrollmentKeyTableFormatter_FormatEnrollmentKeys_Empty(t *testing.T) {
	formatter := NewEnrollmentKeyTableFormatter(sdkoutput.ColorNever, false)

	rows := formatter.FormatEnrollmentKeys([]client.EnrollmentKey{})
	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}
}

func TestFormatEnrollmentKeyTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty", input: "", expected: "-"},
		{name: "valid RFC3339", input: "2026-06-15T10:30:00Z", expected: "2026-06-15"},
		{name: "unparseable falls back to raw", input: "not-a-date", expected: "not-a-date"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatEnrollmentKeyTimestamp(tc.input)
			if result != tc.expected {
				t.Errorf("formatEnrollmentKeyTimestamp(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}
