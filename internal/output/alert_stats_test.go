// Package output provides CLI output helpers for StackEye commands.
// Task #7170
package output

import (
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

func TestNewAlertStatsFormatter(t *testing.T) {
	formatter := NewAlertStatsFormatter(sdkoutput.ColorNever)

	if formatter == nil {
		t.Fatal("expected formatter to be non-nil")
	}

	if formatter.colorMgr == nil {
		t.Error("expected colorMgr to be non-nil")
	}
}

func TestAlertStatsFormatter_FormatAlertStats_Basic(t *testing.T) {
	formatter := NewAlertStatsFormatter(sdkoutput.ColorNever)

	stats := &client.AlertStats{
		Period:             "24h",
		TotalAlerts:        42,
		ActiveAlerts:       5,
		AcknowledgedAlerts: 3,
		ResolvedAlerts:     34,
		CriticalAlerts:     8,
		WarningAlerts:      20,
		InfoAlerts:         14,
	}

	rows := formatter.FormatAlertStats(stats)

	// Should have 10 rows: Period, Total, Active, Acknowledged, Resolved, separator, By Severity header, Critical, Warning, Info
	if len(rows) != 10 {
		t.Fatalf("expected 10 rows, got %d", len(rows))
	}

	// Verify key rows
	if rows[0].Metric != "Period" || rows[0].Value != "24h" {
		t.Errorf("row 0: expected Period=24h, got %q=%q", rows[0].Metric, rows[0].Value)
	}

	if rows[1].Metric != "Total Alerts" || rows[1].Value != "42" {
		t.Errorf("row 1: expected Total Alerts=42, got %q=%q", rows[1].Metric, rows[1].Value)
	}

	if rows[2].Metric != "Active" || rows[2].Value != "5" {
		t.Errorf("row 2: expected Active=5, got %q=%q", rows[2].Metric, rows[2].Value)
	}

	if rows[3].Metric != "Acknowledged" || rows[3].Value != "3" {
		t.Errorf("row 3: expected Acknowledged=3, got %q=%q", rows[3].Metric, rows[3].Value)
	}

	if rows[4].Metric != "Resolved" || rows[4].Value != "34" {
		t.Errorf("row 4: expected Resolved=34, got %q=%q", rows[4].Metric, rows[4].Value)
	}

	// Row 5 is separator
	if rows[5].Metric != "" || rows[5].Value != "" {
		t.Errorf("row 5: expected empty separator, got %q=%q", rows[5].Metric, rows[5].Value)
	}

	// Row 6 is By Severity header
	if rows[6].Metric != "By Severity" {
		t.Errorf("row 6: expected By Severity header, got %q", rows[6].Metric)
	}

	if rows[7].Metric != "  Critical" || rows[7].Value != "8" {
		t.Errorf("row 7: expected Critical=8, got %q=%q", rows[7].Metric, rows[7].Value)
	}

	if rows[8].Metric != "  Warning" || rows[8].Value != "20" {
		t.Errorf("row 8: expected Warning=20, got %q=%q", rows[8].Metric, rows[8].Value)
	}

	if rows[9].Metric != "  Info" || rows[9].Value != "14" {
		t.Errorf("row 9: expected Info=14, got %q=%q", rows[9].Metric, rows[9].Value)
	}
}

func TestAlertStatsFormatter_FormatAlertStats_WithMTTA(t *testing.T) {
	formatter := NewAlertStatsFormatter(sdkoutput.ColorNever)

	mtta := int64(300) // 5 minutes
	stats := &client.AlertStats{
		Period:             "7d",
		TotalAlerts:        10,
		ActiveAlerts:       1,
		AcknowledgedAlerts: 2,
		ResolvedAlerts:     7,
		CriticalAlerts:     3,
		WarningAlerts:      4,
		InfoAlerts:         3,
		MTTA:               &mtta,
	}

	rows := formatter.FormatAlertStats(stats)

	// 10 base rows + separator + MTTA = 12
	if len(rows) != 12 {
		t.Fatalf("expected 12 rows, got %d", len(rows))
	}

	// Row 10 is separator before MTTA
	if rows[10].Metric != "" || rows[10].Value != "" {
		t.Errorf("row 10: expected separator, got %q=%q", rows[10].Metric, rows[10].Value)
	}

	if rows[11].Metric != "MTTA (Mean Time To Acknowledge)" {
		t.Errorf("row 11: expected MTTA metric, got %q", rows[11].Metric)
	}

	if rows[11].Value != "5m" {
		t.Errorf("row 11: expected MTTA value 5m, got %q", rows[11].Value)
	}
}

func TestAlertStatsFormatter_FormatAlertStats_WithMTTR(t *testing.T) {
	formatter := NewAlertStatsFormatter(sdkoutput.ColorNever)

	mttr := int64(7200) // 2 hours
	stats := &client.AlertStats{
		Period:             "30d",
		TotalAlerts:        50,
		ActiveAlerts:       0,
		AcknowledgedAlerts: 0,
		ResolvedAlerts:     50,
		CriticalAlerts:     10,
		WarningAlerts:      25,
		InfoAlerts:         15,
		MTTR:               &mttr,
	}

	rows := formatter.FormatAlertStats(stats)

	// 10 base rows + separator + MTTR = 12
	if len(rows) != 12 {
		t.Fatalf("expected 12 rows, got %d", len(rows))
	}

	lastRow := rows[len(rows)-1]
	if lastRow.Metric != "MTTR (Mean Time To Resolve)" {
		t.Errorf("last row: expected MTTR metric, got %q", lastRow.Metric)
	}

	if lastRow.Value != "2h" {
		t.Errorf("last row: expected MTTR value 2h, got %q", lastRow.Value)
	}
}

func TestAlertStatsFormatter_FormatAlertStats_WithBothMTTAAndMTTR(t *testing.T) {
	formatter := NewAlertStatsFormatter(sdkoutput.ColorNever)

	mtta := int64(120)  // 2 minutes
	mttr := int64(3660) // 1h1m
	stats := &client.AlertStats{
		Period:             "24h",
		TotalAlerts:        5,
		ActiveAlerts:       1,
		AcknowledgedAlerts: 1,
		ResolvedAlerts:     3,
		CriticalAlerts:     2,
		WarningAlerts:      2,
		InfoAlerts:         1,
		MTTA:               &mtta,
		MTTR:               &mttr,
	}

	rows := formatter.FormatAlertStats(stats)

	// 10 base rows + separator + MTTA + MTTR = 13
	if len(rows) != 13 {
		t.Fatalf("expected 13 rows, got %d", len(rows))
	}

	if rows[11].Metric != "MTTA (Mean Time To Acknowledge)" {
		t.Errorf("row 11: expected MTTA, got %q", rows[11].Metric)
	}
	if rows[11].Value != "2m" {
		t.Errorf("row 11: expected 2m, got %q", rows[11].Value)
	}

	if rows[12].Metric != "MTTR (Mean Time To Resolve)" {
		t.Errorf("row 12: expected MTTR, got %q", rows[12].Metric)
	}
	if rows[12].Value != "1h1m" {
		t.Errorf("row 12: expected 1h1m, got %q", rows[12].Value)
	}
}

func TestFormatDurationSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{name: "zero seconds", input: 0, expected: "0s"},
		{name: "30 seconds", input: 30, expected: "30s"},
		{name: "59 seconds", input: 59, expected: "59s"},
		{name: "exactly 1 minute", input: 60, expected: "1m"},
		{name: "1 minute 30 seconds", input: 90, expected: "1m30s"},
		{name: "5 minutes", input: 300, expected: "5m"},
		{name: "59 minutes 59 seconds", input: 3599, expected: "59m59s"},
		{name: "exactly 1 hour", input: 3600, expected: "1h"},
		{name: "1 hour 30 minutes", input: 5400, expected: "1h30m"},
		{name: "2 hours", input: 7200, expected: "2h"},
		{name: "2 hours 15 minutes", input: 8100, expected: "2h15m"},
		{name: "24 hours", input: 86400, expected: "24h"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatDurationSeconds(tc.input)
			if result != tc.expected {
				t.Errorf("formatDurationSeconds(%d) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestAlertStatsFormatter_ZeroValues(t *testing.T) {
	formatter := NewAlertStatsFormatter(sdkoutput.ColorNever)

	stats := &client.AlertStats{
		Period:             "",
		TotalAlerts:        0,
		ActiveAlerts:       0,
		AcknowledgedAlerts: 0,
		ResolvedAlerts:     0,
		CriticalAlerts:     0,
		WarningAlerts:      0,
		InfoAlerts:         0,
	}

	rows := formatter.FormatAlertStats(stats)

	if len(rows) != 10 {
		t.Fatalf("expected 10 rows, got %d", len(rows))
	}

	if rows[0].Value != "" {
		t.Errorf("expected empty period, got %q", rows[0].Value)
	}

	if rows[1].Value != "0" {
		t.Errorf("expected Total Alerts=0, got %q", rows[1].Value)
	}
}
