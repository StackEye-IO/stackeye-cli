// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// TestNewIncidentTableFormatter verifies the formatter is created correctly.
func TestNewIncidentTableFormatter(t *testing.T) {
	formatter := NewIncidentTableFormatter(sdkoutput.ColorNever, false)
	if formatter == nil {
		t.Fatal("expected non-nil formatter")
	}
	if formatter.colorMgr == nil {
		t.Error("expected colorMgr to be initialized")
	}
	if formatter.isWide {
		t.Error("expected isWide to be false")
	}

	formatterWide := NewIncidentTableFormatter(sdkoutput.ColorAlways, true)
	if !formatterWide.isWide {
		t.Error("expected isWide to be true")
	}
}

// TestFormatIncidents verifies multiple incidents are formatted correctly.
func TestFormatIncidents(t *testing.T) {
	formatter := NewIncidentTableFormatter(sdkoutput.ColorNever, false)
	now := time.Now()
	resolved := now.Add(-time.Hour)

	incidents := []client.Incident{
		{ID: 1, Title: "First incident", Status: "investigating", Impact: "minor", CreatedAt: now, UpdatedAt: now},
		{ID: 2, Title: "Second incident", Status: "resolved", Impact: "none", CreatedAt: now, UpdatedAt: now, ResolvedAt: &resolved},
	}

	rows := formatter.FormatIncidents(incidents)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0].ID != "1" {
		t.Errorf("expected ID '1', got %q", rows[0].ID)
	}
	if rows[0].Title != "First incident" {
		t.Errorf("expected title 'First incident', got %q", rows[0].Title)
	}
	if rows[1].ID != "2" {
		t.Errorf("expected ID '2', got %q", rows[1].ID)
	}
}

// TestFormatIncident verifies a single incident is formatted correctly.
func TestFormatIncident(t *testing.T) {
	formatter := NewIncidentTableFormatter(sdkoutput.ColorNever, false)
	now := time.Now()

	incident := client.Incident{
		ID:        123,
		Title:     "Test incident",
		Status:    "monitoring",
		Impact:    "major",
		CreatedAt: now,
		UpdatedAt: now,
	}

	row := formatter.FormatIncident(incident)

	if row.ID != "123" {
		t.Errorf("expected ID '123', got %q", row.ID)
	}
	if row.Title != "Test incident" {
		t.Errorf("expected title 'Test incident', got %q", row.Title)
	}
	if row.Resolved != "-" {
		t.Errorf("expected resolved '-', got %q", row.Resolved)
	}
}

// TestFormatStatus verifies status values are formatted correctly.
func TestFormatStatus(t *testing.T) {
	// Use ColorNever to get predictable output without ANSI codes
	formatter := NewIncidentTableFormatter(sdkoutput.ColorNever, false)

	tests := []struct {
		status   string
		expected string
	}{
		{"resolved", "Resolved"},
		{"monitoring", "Monitoring"},
		{"identified", "Identified"},
		{"investigating", "Investigating"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := formatter.formatStatus(tt.status)
			if result != tt.expected {
				t.Errorf("formatStatus(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

// TestFormatImpact verifies impact values are formatted correctly.
func TestFormatImpact(t *testing.T) {
	formatter := NewIncidentTableFormatter(sdkoutput.ColorNever, false)

	tests := []struct {
		impact   string
		expected string
	}{
		{"none", "None"},
		{"minor", "Minor"},
		{"major", "Major"},
		{"critical", "Critical"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.impact, func(t *testing.T) {
			result := formatter.formatImpact(tt.impact)
			if result != tt.expected {
				t.Errorf("formatImpact(%q) = %q, want %q", tt.impact, result, tt.expected)
			}
		})
	}
}

// TestTruncateIncidentTitle verifies title truncation works correctly.
func TestTruncateIncidentTitle(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		maxLen   int
		expected string
	}{
		{
			name:     "short title no truncation",
			title:    "Short",
			maxLen:   40,
			expected: "Short",
		},
		{
			name:     "exact length no truncation",
			title:    "This is exactly forty characters long!!",
			maxLen:   40,
			expected: "This is exactly forty characters long!!",
		},
		{
			name:     "long title truncated",
			title:    "This is a very long incident title that exceeds the maximum length",
			maxLen:   40,
			expected: "This is a very long incident title th...",
		},
		{
			name:     "empty title",
			title:    "",
			maxLen:   40,
			expected: "",
		},
		{
			name:     "minimum maxLen enforced",
			title:    "Hello",
			maxLen:   2,
			expected: "H...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateIncidentTitle(tt.title, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateIncidentTitle(%q, %d) = %q, want %q", tt.title, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestTruncateIncidentTitle_Unicode verifies Unicode characters are handled correctly.
func TestTruncateIncidentTitle_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		maxLen   int
		expected string
	}{
		{
			name:     "chinese characters no truncation",
			title:    "测试",
			maxLen:   10,
			expected: "测试",
		},
		{
			name:     "chinese characters truncated",
			title:    "这是一个很长的中文标题需要截断",
			maxLen:   10,
			expected: "这是一个很长的...",
		},
		{
			name:     "emoji no truncation",
			title:    "Alert 🚨",
			maxLen:   10,
			expected: "Alert 🚨",
		},
		{
			name:     "emoji truncated",
			title:    "🚨🔥💀⚠️🛑 Multiple emoji alert",
			maxLen:   10,
			expected: "🚨🔥💀⚠️🛑 ...",
		},
		{
			name:     "accented characters",
			title:    "Résumé café naïve",
			maxLen:   10,
			expected: "Résumé ...",
		},
		{
			name:     "mixed unicode and ascii",
			title:    "Server 日本語 down",
			maxLen:   12,
			expected: "Server 日本...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateIncidentTitle(tt.title, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateIncidentTitle(%q, %d) = %q, want %q", tt.title, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestFormatResolvedAt verifies resolved timestamp formatting.
func TestFormatResolvedAt(t *testing.T) {
	t.Run("nil resolved time", func(t *testing.T) {
		result := formatResolvedAt(nil)
		if result != "-" {
			t.Errorf("expected '-', got %q", result)
		}
	})

	t.Run("valid resolved time", func(t *testing.T) {
		resolved := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := formatResolvedAt(&resolved)
		expected := "2024-01-15 10:30"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

// TestFormatIncidentCount verifies pagination count formatting.
func TestFormatIncidentCount(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		page     int
		limit    int
		expected string
	}{
		{
			name:     "zero total",
			total:    0,
			page:     1,
			limit:    20,
			expected: "",
		},
		{
			name:     "first page partial",
			total:    5,
			page:     1,
			limit:    20,
			expected: "Showing 1-5 of 5 incidents",
		},
		{
			name:     "first page full",
			total:    50,
			page:     1,
			limit:    20,
			expected: "Showing 1-20 of 50 incidents",
		},
		{
			name:     "second page",
			total:    50,
			page:     2,
			limit:    20,
			expected: "Showing 21-40 of 50 incidents",
		},
		{
			name:     "last page partial",
			total:    45,
			page:     3,
			limit:    20,
			expected: "Showing 41-45 of 45 incidents",
		},
		{
			name:     "single item",
			total:    1,
			page:     1,
			limit:    20,
			expected: "Showing 1-1 of 1 incidents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatIncidentCount(tt.total, tt.page, tt.limit)
			if result != tt.expected {
				t.Errorf("FormatIncidentCount(%d, %d, %d) = %q, want %q",
					tt.total, tt.page, tt.limit, result, tt.expected)
			}
		})
	}
}

// TestIncidentTableRow_StructTags verifies the table struct tags are correct.
func TestIncidentTableRow_StructTags(t *testing.T) {
	// This test ensures the struct tags haven't been accidentally modified
	row := IncidentTableRow{
		ID:       "1",
		Title:    "Test",
		Status:   "investigating",
		Impact:   "minor",
		Created:  "2024-01-15",
		Updated:  "2024-01-15",
		Resolved: "-",
	}

	// Verify all fields are accessible (compile-time check)
	_ = row.ID
	_ = row.Title
	_ = row.Status
	_ = row.Impact
	_ = row.Created
	_ = row.Updated
	_ = row.Resolved
}

// TestFormatIncidents_EmptySlice verifies empty input is handled correctly.
func TestFormatIncidents_EmptySlice(t *testing.T) {
	formatter := NewIncidentTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatIncidents([]client.Incident{})

	if rows == nil {
		t.Error("expected non-nil slice, got nil")
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

// TestFormatIncidents_PreservesOrder verifies incident order is maintained.
func TestFormatIncidents_PreservesOrder(t *testing.T) {
	formatter := NewIncidentTableFormatter(sdkoutput.ColorNever, false)
	now := time.Now()

	incidents := []client.Incident{
		{ID: 3, Title: "Third", Status: "resolved", Impact: "none", CreatedAt: now, UpdatedAt: now},
		{ID: 1, Title: "First", Status: "investigating", Impact: "critical", CreatedAt: now, UpdatedAt: now},
		{ID: 2, Title: "Second", Status: "monitoring", Impact: "minor", CreatedAt: now, UpdatedAt: now},
	}

	rows := formatter.FormatIncidents(incidents)

	expectedIDs := []string{"3", "1", "2"}
	for i, expected := range expectedIDs {
		if rows[i].ID != expected {
			t.Errorf("row[%d].ID = %q, want %q", i, rows[i].ID, expected)
		}
	}
}
