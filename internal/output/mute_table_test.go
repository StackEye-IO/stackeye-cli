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

func TestNewMuteTableFormatter(t *testing.T) {
	formatter := NewMuteTableFormatter(sdkoutput.ColorNever, false)

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

func TestNewMuteTableFormatter_WideMode(t *testing.T) {
	formatter := NewMuteTableFormatter(sdkoutput.ColorNever, true)

	if !formatter.isWide {
		t.Error("expected isWide to be true")
	}
}

func TestMuteTableFormatter_FormatMute_Active(t *testing.T) {
	formatter := NewMuteTableFormatter(sdkoutput.ColorNever, false)

	reason := "Scheduled maintenance"
	mute := client.AlertMute{
		ID:                  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		OrganizationID:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ScopeType:           client.MuteScopeOrganization,
		DurationMinutes:     60,
		Reason:              &reason,
		CreatedBy:           uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		StartsAt:            time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC),
		IsMaintenanceWindow: false,
		IsActive:            true,
		CreatedAt:           time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:           time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC),
	}

	row := formatter.FormatMute(mute)

	if row.Status != "ACTIVE" {
		t.Errorf("expected Status='ACTIVE', got %q", row.Status)
	}

	if row.Scope != "Organization" {
		t.Errorf("expected Scope='Organization', got %q", row.Scope)
	}

	if row.Target != "All" {
		t.Errorf("expected Target='All', got %q", row.Target)
	}

	if row.Duration != "1h" {
		t.Errorf("expected Duration='1h', got %q", row.Duration)
	}

	if row.Reason != "Scheduled maintenance" {
		t.Errorf("expected Reason='Scheduled maintenance', got %q", row.Reason)
	}

	if row.Maintenance != "No" {
		t.Errorf("expected Maintenance='No', got %q", row.Maintenance)
	}

	if row.ID != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected ID='11111111-1111-1111-1111-111111111111', got %q", row.ID)
	}
}

func TestMuteTableFormatter_FormatMute_Expired(t *testing.T) {
	formatter := NewMuteTableFormatter(sdkoutput.ColorNever, false)

	mute := client.AlertMute{
		ID:              uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		OrganizationID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ScopeType:       client.MuteScopeOrganization,
		DurationMinutes: 30,
		IsActive:        false,
		CreatedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	row := formatter.FormatMute(mute)

	if row.Status != "EXPIRED" {
		t.Errorf("expected Status='EXPIRED', got %q", row.Status)
	}
}

func TestMuteTableFormatter_FormatMutes(t *testing.T) {
	formatter := NewMuteTableFormatter(sdkoutput.ColorNever, false)

	mutes := []client.AlertMute{
		{
			ID:              uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			OrganizationID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			ScopeType:       client.MuteScopeOrganization,
			DurationMinutes: 60,
			IsActive:        true,
			CreatedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:              uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			OrganizationID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			ScopeType:       client.MuteScopeProbe,
			DurationMinutes: 120,
			IsActive:        false,
			CreatedAt:       time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	rows := formatter.FormatMutes(mutes)

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0].Status != "ACTIVE" {
		t.Errorf("expected first row Status='ACTIVE', got %q", rows[0].Status)
	}

	if rows[1].Status != "EXPIRED" {
		t.Errorf("expected second row Status='EXPIRED', got %q", rows[1].Status)
	}
}

func TestMuteTableFormatter_FormatMutes_Empty(t *testing.T) {
	formatter := NewMuteTableFormatter(sdkoutput.ColorNever, false)

	rows := formatter.FormatMutes([]client.AlertMute{})

	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}
}

func TestFormatScopeType(t *testing.T) {
	tests := []struct {
		name     string
		input    client.MuteScopeType
		expected string
	}{
		{name: "organization", input: client.MuteScopeOrganization, expected: "Organization"},
		{name: "probe", input: client.MuteScopeProbe, expected: "Probe"},
		{name: "channel", input: client.MuteScopeChannel, expected: "Channel"},
		{name: "alert_type", input: client.MuteScopeAlertType, expected: "Alert Type"},
		{name: "unknown", input: client.MuteScopeType("custom"), expected: "custom"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatScopeType(tc.input)
			if result != tc.expected {
				t.Errorf("formatScopeType(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatMuteTarget(t *testing.T) {
	probeID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	channelID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	alertType := client.AlertTypeStatusDown

	tests := []struct {
		name     string
		input    client.AlertMute
		expected string
	}{
		{
			name:     "organization scope",
			input:    client.AlertMute{ScopeType: client.MuteScopeOrganization},
			expected: "All",
		},
		{
			name:     "probe scope with ID",
			input:    client.AlertMute{ScopeType: client.MuteScopeProbe, ProbeID: &probeID},
			expected: "aaaaaaaa...",
		},
		{
			name:     "probe scope without ID",
			input:    client.AlertMute{ScopeType: client.MuteScopeProbe},
			expected: "-",
		},
		{
			name:     "channel scope with ID",
			input:    client.AlertMute{ScopeType: client.MuteScopeChannel, ChannelID: &channelID},
			expected: "11111111...",
		},
		{
			name:     "channel scope without ID",
			input:    client.AlertMute{ScopeType: client.MuteScopeChannel},
			expected: "-",
		},
		{
			name:     "alert_type scope with type",
			input:    client.AlertMute{ScopeType: client.MuteScopeAlertType, AlertType: &alertType},
			expected: "Down",
		},
		{
			name:     "alert_type scope without type",
			input:    client.AlertMute{ScopeType: client.MuteScopeAlertType},
			expected: "-",
		},
		{
			name:     "unknown scope",
			input:    client.AlertMute{ScopeType: client.MuteScopeType("unknown")},
			expected: "-",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatMuteTarget(tc.input)
			if result != tc.expected {
				t.Errorf("formatMuteTarget() = %q, expected %q", result, tc.expected)
			}
		})
	}
}

func TestFormatMuteDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{name: "30 minutes", input: 30, expected: "30m"},
		{name: "59 minutes", input: 59, expected: "59m"},
		{name: "exactly 1 hour", input: 60, expected: "1h"},
		{name: "1 hour 30 minutes", input: 90, expected: "1h30m"},
		{name: "2 hours", input: 120, expected: "2h"},
		{name: "2 hours 15 minutes", input: 135, expected: "2h15m"},
		{name: "24 hours", input: 1440, expected: "24h"},
		{name: "1 minute", input: 1, expected: "1m"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatMuteDuration(tc.input)
			if result != tc.expected {
				t.Errorf("formatMuteDuration(%d) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatExpiresAt(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	in30min := time.Now().Add(30 * time.Minute)
	in5hours := time.Now().Add(5 * time.Hour)
	in3days := time.Now().Add(72 * time.Hour)

	tests := []struct {
		name     string
		input    *time.Time
		contains string // Use contains since relative times vary
	}{
		{name: "nil", input: nil, contains: "Never"},
		{name: "past", input: &past, contains: "Expired"},
		{name: "within hour", input: &in30min, contains: "in"},
		{name: "within day", input: &in5hours, contains: "in"},
		{name: "beyond day", input: &in3days, contains: "20"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatExpiresAt(tc.input)
			if len(result) == 0 {
				t.Error("expected non-empty result")
			}
			// For relative times, just check it contains expected substring
			if tc.contains != "" {
				found := false
				if len(result) >= len(tc.contains) {
					for i := 0; i <= len(result)-len(tc.contains); i++ {
						if result[i:i+len(tc.contains)] == tc.contains {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("formatExpiresAt() = %q, expected to contain %q", result, tc.contains)
				}
			}
		})
	}
}

func TestFormatReason(t *testing.T) {
	short := "Maintenance window"
	long := "This is a very long reason that exceeds the thirty character limit"
	empty := ""

	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{name: "nil reason", input: nil, expected: "-"},
		{name: "empty reason", input: &empty, expected: "-"},
		{name: "short reason", input: &short, expected: "Maintenance window"},
		{name: "long reason (truncated)", input: &long, expected: "This is a very long reason ..."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatReason(tc.input)
			if result != tc.expected {
				t.Errorf("formatReason() = %q, expected %q", result, tc.expected)
			}
		})
	}
}

func TestFormatMaintenance(t *testing.T) {
	name := "Weekly Patch Window"
	longName := "This Is A Very Long Maintenance Window Name"
	empty := ""

	tests := []struct {
		name                string
		isMaintenanceWindow bool
		maintenanceName     *string
		expected            string
	}{
		{name: "not maintenance", isMaintenanceWindow: false, maintenanceName: nil, expected: "No"},
		{name: "maintenance without name", isMaintenanceWindow: true, maintenanceName: nil, expected: "Yes"},
		{name: "maintenance with empty name", isMaintenanceWindow: true, maintenanceName: &empty, expected: "Yes"},
		{name: "maintenance with name", isMaintenanceWindow: true, maintenanceName: &name, expected: "Weekly Patch Window"},
		{name: "maintenance with long name", isMaintenanceWindow: true, maintenanceName: &longName, expected: "This Is A Very Lo..."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatMaintenance(tc.isMaintenanceWindow, tc.maintenanceName)
			if result != tc.expected {
				t.Errorf("formatMaintenance() = %q, expected %q", result, tc.expected)
			}
		})
	}
}

func TestTruncateUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "full UUID", input: "11111111-2222-3333-4444-555555555555", expected: "11111111..."},
		{name: "short string", input: "abc", expected: "abc"},
		{name: "exactly 8 chars", input: "12345678", expected: "12345678"},
		{name: "9 chars", input: "123456789", expected: "12345678..."},
		{name: "empty", input: "", expected: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := truncateUUID(tc.input)
			if result != tc.expected {
				t.Errorf("truncateUUID(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestMuteTableFormatter_ProbeScope(t *testing.T) {
	formatter := NewMuteTableFormatter(sdkoutput.ColorNever, false)
	probeID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	mute := client.AlertMute{
		ID:              uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		OrganizationID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ScopeType:       client.MuteScopeProbe,
		ProbeID:         &probeID,
		DurationMinutes: 45,
		IsActive:        true,
		CreatedAt:       time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC),
	}

	row := formatter.FormatMute(mute)

	if row.Scope != "Probe" {
		t.Errorf("expected Scope='Probe', got %q", row.Scope)
	}

	if row.Target != "aaaaaaaa..." {
		t.Errorf("expected Target='aaaaaaaa...', got %q", row.Target)
	}

	if row.Duration != "45m" {
		t.Errorf("expected Duration='45m', got %q", row.Duration)
	}
}

func TestMuteTableFormatter_MaintenanceWindow(t *testing.T) {
	formatter := NewMuteTableFormatter(sdkoutput.ColorNever, true)
	maintName := "DB Upgrade"

	mute := client.AlertMute{
		ID:                  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		OrganizationID:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ScopeType:           client.MuteScopeOrganization,
		DurationMinutes:     120,
		IsMaintenanceWindow: true,
		MaintenanceName:     &maintName,
		IsActive:            true,
		CreatedAt:           time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:           time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC),
	}

	row := formatter.FormatMute(mute)

	if row.Maintenance != "DB Upgrade" {
		t.Errorf("expected Maintenance='DB Upgrade', got %q", row.Maintenance)
	}
}
