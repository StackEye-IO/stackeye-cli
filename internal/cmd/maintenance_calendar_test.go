package cmd

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

func TestNewMaintenanceCalendarCmd(t *testing.T) {
	cmd := NewMaintenanceCalendarCmd()

	if cmd.Use != "calendar" {
		t.Errorf("expected Use='calendar', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Show maintenance windows in calendar view" {
		t.Errorf("expected Short='Show maintenance windows in calendar view', got %q", cmd.Short)
	}
}

func TestNewMaintenanceCalendarCmd_Aliases(t *testing.T) {
	cmd := NewMaintenanceCalendarCmd()

	if len(cmd.Aliases) != 1 {
		t.Errorf("expected 1 alias, got %d", len(cmd.Aliases))
	}

	if cmd.Aliases[0] != "cal" {
		t.Errorf("expected alias 'cal', got %q", cmd.Aliases[0])
	}
}

func TestNewMaintenanceCalendarCmd_Flags(t *testing.T) {
	cmd := NewMaintenanceCalendarCmd()

	// Check --month flag
	monthFlag := cmd.Flags().Lookup("month")
	if monthFlag == nil {
		t.Error("expected --month flag to be defined")
	} else {
		if monthFlag.DefValue != "false" {
			t.Errorf("expected --month default to be 'false', got %q", monthFlag.DefValue)
		}
	}

	// Check --from flag
	fromFlag := cmd.Flags().Lookup("from")
	if fromFlag == nil {
		t.Error("expected --from flag to be defined")
	} else {
		if fromFlag.DefValue != "" {
			t.Errorf("expected --from default to be empty, got %q", fromFlag.DefValue)
		}
	}

	// Check --include-expired flag
	expiredFlag := cmd.Flags().Lookup("include-expired")
	if expiredFlag == nil {
		t.Error("expected --include-expired flag to be defined")
	} else {
		if expiredFlag.DefValue != "false" {
			t.Errorf("expected --include-expired default to be 'false', got %q", expiredFlag.DefValue)
		}
	}
}

func TestNewMaintenanceCalendarCmd_Long(t *testing.T) {
	cmd := NewMaintenanceCalendarCmd()

	long := cmd.Long

	// Should describe view modes
	modes := []string{"Week", "Month"}
	for _, mode := range modes {
		if !strings.Contains(long, mode) {
			t.Errorf("expected Long description to mention view mode %q", mode)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye maintenance calendar") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention key flags
	flags := []string{"--month", "--from", "--include-expired"}
	for _, flag := range flags {
		if !strings.Contains(long, flag) {
			t.Errorf("expected Long description to mention flag %q", flag)
		}
	}
}

func TestCalculateDateRange_WeekView(t *testing.T) {
	// Use a known date: Wednesday, January 15, 2025
	baseDate := time.Date(2025, 1, 15, 12, 0, 0, 0, time.Local)

	start, end := calculateDateRange(baseDate, false)

	// Should return Monday of that week
	expectedStart := time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local)
	if !start.Equal(expectedStart) {
		t.Errorf("expected start=%v, got %v", expectedStart, start)
	}

	// Should return Sunday of that week
	expectedEnd := time.Date(2025, 1, 19, 0, 0, 0, 0, time.Local)
	if !end.Equal(expectedEnd) {
		t.Errorf("expected end=%v, got %v", expectedEnd, end)
	}
}

func TestCalculateDateRange_WeekView_Sunday(t *testing.T) {
	// Test when baseDate is Sunday - should still show that same week
	baseDate := time.Date(2025, 1, 19, 12, 0, 0, 0, time.Local) // Sunday

	start, end := calculateDateRange(baseDate, false)

	expectedStart := time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local) // Monday
	if !start.Equal(expectedStart) {
		t.Errorf("expected start=%v, got %v", expectedStart, start)
	}

	expectedEnd := time.Date(2025, 1, 19, 0, 0, 0, 0, time.Local) // Sunday
	if !end.Equal(expectedEnd) {
		t.Errorf("expected end=%v, got %v", expectedEnd, end)
	}
}

func TestCalculateDateRange_WeekView_Monday(t *testing.T) {
	// Test when baseDate is Monday
	baseDate := time.Date(2025, 1, 13, 12, 0, 0, 0, time.Local) // Monday

	start, end := calculateDateRange(baseDate, false)

	expectedStart := time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local) // Same Monday
	if !start.Equal(expectedStart) {
		t.Errorf("expected start=%v, got %v", expectedStart, start)
	}

	expectedEnd := time.Date(2025, 1, 19, 0, 0, 0, 0, time.Local) // Sunday
	if !end.Equal(expectedEnd) {
		t.Errorf("expected end=%v, got %v", expectedEnd, end)
	}
}

func TestCalculateDateRange_MonthView(t *testing.T) {
	// Use a known date: January 15, 2025
	baseDate := time.Date(2025, 1, 15, 12, 0, 0, 0, time.Local)

	start, end := calculateDateRange(baseDate, true)

	// Should return first day of month
	expectedStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local)
	if !start.Equal(expectedStart) {
		t.Errorf("expected start=%v, got %v", expectedStart, start)
	}

	// Should return last day of month
	expectedEnd := time.Date(2025, 1, 31, 0, 0, 0, 0, time.Local)
	if !end.Equal(expectedEnd) {
		t.Errorf("expected end=%v, got %v", expectedEnd, end)
	}
}

func TestCalculateDateRange_MonthView_February(t *testing.T) {
	// Test February in non-leap year (2025)
	baseDate := time.Date(2025, 2, 15, 12, 0, 0, 0, time.Local)

	start, end := calculateDateRange(baseDate, true)

	expectedStart := time.Date(2025, 2, 1, 0, 0, 0, 0, time.Local)
	if !start.Equal(expectedStart) {
		t.Errorf("expected start=%v, got %v", expectedStart, start)
	}

	expectedEnd := time.Date(2025, 2, 28, 0, 0, 0, 0, time.Local)
	if !end.Equal(expectedEnd) {
		t.Errorf("expected end=%v, got %v", expectedEnd, end)
	}
}

func TestCalculateDateRange_MonthView_LeapYear(t *testing.T) {
	// Test February in leap year (2024)
	baseDate := time.Date(2024, 2, 15, 12, 0, 0, 0, time.Local)

	start, end := calculateDateRange(baseDate, true)

	expectedStart := time.Date(2024, 2, 1, 0, 0, 0, 0, time.Local)
	if !start.Equal(expectedStart) {
		t.Errorf("expected start=%v, got %v", expectedStart, start)
	}

	expectedEnd := time.Date(2024, 2, 29, 0, 0, 0, 0, time.Local)
	if !end.Equal(expectedEnd) {
		t.Errorf("expected end=%v, got %v", expectedEnd, end)
	}
}

func TestFilterWindowsByDateRange_Overlapping(t *testing.T) {
	rangeStart := time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local)
	rangeEnd := time.Date(2025, 1, 19, 0, 0, 0, 0, time.Local)

	windows := []client.AlertMute{
		{
			StartsAt:        time.Date(2025, 1, 14, 10, 0, 0, 0, time.Local),
			DurationMinutes: 60, // Ends at 11:00 same day - within range
		},
		{
			StartsAt:        time.Date(2025, 1, 10, 10, 0, 0, 0, time.Local),
			DurationMinutes: 60, // Ends Jan 10 - before range
		},
		{
			StartsAt:        time.Date(2025, 1, 25, 10, 0, 0, 0, time.Local),
			DurationMinutes: 60, // Starts after range
		},
	}

	filtered := filterWindowsByDateRange(windows, rangeStart, rangeEnd)

	if len(filtered) != 1 {
		t.Errorf("expected 1 window in range, got %d", len(filtered))
	}
}

func TestFilterWindowsByDateRange_SpanningIntoRange(t *testing.T) {
	rangeStart := time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local)
	rangeEnd := time.Date(2025, 1, 19, 0, 0, 0, 0, time.Local)

	windows := []client.AlertMute{
		{
			// Window starts before range but extends into it
			StartsAt:        time.Date(2025, 1, 12, 22, 0, 0, 0, time.Local),
			DurationMinutes: 180, // 3 hours - ends at 01:00 on Jan 13
		},
	}

	filtered := filterWindowsByDateRange(windows, rangeStart, rangeEnd)

	if len(filtered) != 1 {
		t.Errorf("expected 1 window (spanning into range), got %d", len(filtered))
	}
}

func TestFilterWindowsByDateRange_SpanningOutOfRange(t *testing.T) {
	rangeStart := time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local)
	rangeEnd := time.Date(2025, 1, 19, 0, 0, 0, 0, time.Local)

	windows := []client.AlertMute{
		{
			// Window starts within range but extends beyond it
			StartsAt:        time.Date(2025, 1, 19, 22, 0, 0, 0, time.Local),
			DurationMinutes: 180, // Ends at 01:00 on Jan 20
		},
	}

	filtered := filterWindowsByDateRange(windows, rangeStart, rangeEnd)

	if len(filtered) != 1 {
		t.Errorf("expected 1 window (spanning out of range), got %d", len(filtered))
	}
}

func TestFilterWindowsByDateRange_Empty(t *testing.T) {
	rangeStart := time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local)
	rangeEnd := time.Date(2025, 1, 19, 0, 0, 0, 0, time.Local)

	windows := []client.AlertMute{}

	filtered := filterWindowsByDateRange(windows, rangeStart, rangeEnd)

	if len(filtered) != 0 {
		t.Errorf("expected 0 windows, got %d", len(filtered))
	}
}

func TestGroupWindowsByDay_SingleDay(t *testing.T) {
	windows := []client.AlertMute{
		{
			StartsAt:        time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local),
			DurationMinutes: 60, // 1 hour, ends same day
		},
	}

	grouped := groupWindowsByDay(windows)

	if len(grouped) != 1 {
		t.Errorf("expected 1 day, got %d", len(grouped))
	}

	dayKey := "2025-01-15"
	if dayWindows, ok := grouped[dayKey]; !ok {
		t.Errorf("expected windows for %s", dayKey)
	} else if len(dayWindows) != 1 {
		t.Errorf("expected 1 window on %s, got %d", dayKey, len(dayWindows))
	}
}

func TestGroupWindowsByDay_MultiDay(t *testing.T) {
	windows := []client.AlertMute{
		{
			// 36 hour window spanning 2 days
			StartsAt:        time.Date(2025, 1, 15, 12, 0, 0, 0, time.Local),
			DurationMinutes: 36 * 60,
		},
	}

	grouped := groupWindowsByDay(windows)

	// Should appear on Jan 15, 16, and 17
	expectedDays := []string{"2025-01-15", "2025-01-16", "2025-01-17"}
	for _, day := range expectedDays {
		if _, ok := grouped[day]; !ok {
			t.Errorf("expected window to appear on %s", day)
		}
	}
}

func TestGroupWindowsByDay_MultipleWindows(t *testing.T) {
	windows := []client.AlertMute{
		{
			StartsAt:        time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local),
			DurationMinutes: 60,
		},
		{
			StartsAt:        time.Date(2025, 1, 15, 14, 0, 0, 0, time.Local),
			DurationMinutes: 60,
		},
	}

	grouped := groupWindowsByDay(windows)

	dayKey := "2025-01-15"
	if dayWindows, ok := grouped[dayKey]; !ok {
		t.Errorf("expected windows for %s", dayKey)
	} else if len(dayWindows) != 2 {
		t.Errorf("expected 2 windows on %s, got %d", dayKey, len(dayWindows))
	}
}

func TestFormatCalendarDuration_Minutes(t *testing.T) {
	tests := []struct {
		minutes  int
		expected string
	}{
		{15, "15m"},
		{30, "30m"},
		{45, "45m"},
		{59, "59m"},
	}

	for _, tt := range tests {
		result := formatCalendarDuration(tt.minutes)
		if result != tt.expected {
			t.Errorf("formatCalendarDuration(%d) = %q, expected %q", tt.minutes, result, tt.expected)
		}
	}
}

func TestFormatCalendarDuration_Hours(t *testing.T) {
	tests := []struct {
		minutes  int
		expected string
	}{
		{60, "1h"},
		{120, "2h"},
		{180, "3h"},
		{1440, "24h"},
	}

	for _, tt := range tests {
		result := formatCalendarDuration(tt.minutes)
		if result != tt.expected {
			t.Errorf("formatCalendarDuration(%d) = %q, expected %q", tt.minutes, result, tt.expected)
		}
	}
}

func TestFormatCalendarDuration_Mixed(t *testing.T) {
	tests := []struct {
		minutes  int
		expected string
	}{
		{90, "1h30m"},
		{150, "2h30m"},
		{75, "1h15m"},
		{1470, "24h30m"},
	}

	for _, tt := range tests {
		result := formatCalendarDuration(tt.minutes)
		if result != tt.expected {
			t.Errorf("formatCalendarDuration(%d) = %q, expected %q", tt.minutes, result, tt.expected)
		}
	}
}

func TestFormatCalendarScope(t *testing.T) {
	tests := []struct {
		scopeType client.MuteScopeType
		expected  string
	}{
		{client.MuteScopeOrganization, "Org"},
		{client.MuteScopeProbe, "Probe"},
		{client.MuteScopeChannel, "Channel"},
		{client.MuteScopeAlertType, "Type"},
	}

	for _, tt := range tests {
		w := client.AlertMute{ScopeType: tt.scopeType}
		result := formatCalendarScope(w)
		if result != tt.expected {
			t.Errorf("formatCalendarScope(%v) = %q, expected %q", tt.scopeType, result, tt.expected)
		}
	}
}

// Integration tests for runMaintenanceCalendar

func TestRunMaintenanceCalendar_InvalidDateFormat(t *testing.T) {
	tests := []struct {
		name         string
		fromDate     string
		wantErrorMsg string
	}{
		{
			name:         "invalid format - slash separators",
			fromDate:     "2024/01/15",
			wantErrorMsg: `invalid date format "2024/01/15": use YYYY-MM-DD`,
		},
		{
			name:         "invalid format - US style",
			fromDate:     "01-15-2024",
			wantErrorMsg: `invalid date format "01-15-2024": use YYYY-MM-DD`,
		},
		{
			name:         "invalid format - text",
			fromDate:     "january",
			wantErrorMsg: `invalid date format "january": use YYYY-MM-DD`,
		},
		{
			name:         "invalid format - incomplete date",
			fromDate:     "2024-01",
			wantErrorMsg: `invalid date format "2024-01": use YYYY-MM-DD`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &maintenanceCalendarFlags{
				from: tt.fromDate,
			}

			err := runMaintenanceCalendar(context.Background(), flags)

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErrorMsg)
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrorMsg) {
				t.Errorf("expected error containing %q, got %q", tt.wantErrorMsg, err.Error())
			}
		})
	}
}

func TestRunMaintenanceCalendar_ValidDateFormat(t *testing.T) {
	// Test that valid date formats pass validation (will fail later on API client)
	tests := []struct {
		name     string
		fromDate string
	}{
		{
			name:     "standard format",
			fromDate: "2024-01-15",
		},
		{
			name:     "end of month",
			fromDate: "2024-12-31",
		},
		{
			name:     "leap day",
			fromDate: "2024-02-29",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &maintenanceCalendarFlags{
				from: tt.fromDate,
			}

			err := runMaintenanceCalendar(context.Background(), flags)

			// Should fail on API client initialization, not validation
			if err == nil {
				t.Error("expected error (no API client configured), got nil")
				return
			}

			// Error should NOT be a date format error
			if strings.Contains(err.Error(), "invalid date format") {
				t.Errorf("got unexpected date format error for valid date %q: %s", tt.fromDate, err.Error())
			}
		})
	}
}

func TestRunMaintenanceCalendar_DefaultValues(t *testing.T) {
	// Test with all default flags (empty from, no month view, no expired)
	flags := &maintenanceCalendarFlags{
		viewMonth:      false,
		from:           "",
		includeExpired: false,
	}

	err := runMaintenanceCalendar(context.Background(), flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should be about API client, not flag validation
	if strings.Contains(err.Error(), "invalid") {
		t.Errorf("got unexpected validation error with default flags: %s", err.Error())
	}
}

func TestRunMaintenanceCalendar_MonthViewFlag(t *testing.T) {
	// Test month view flag is properly handled
	flags := &maintenanceCalendarFlags{
		viewMonth: true,
		from:      "2024-06-15",
	}

	err := runMaintenanceCalendar(context.Background(), flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	if strings.Contains(err.Error(), "invalid") {
		t.Errorf("got unexpected validation error: %s", err.Error())
	}
}
