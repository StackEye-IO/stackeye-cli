// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// AlertStatsRow represents a row in the alert stats table output.
// Each row shows a metric name and its value.
type AlertStatsRow struct {
	Metric string `table:"METRIC"`
	Value  string `table:"VALUE"`
}

// AlertStatsFormatter converts SDK AlertStats into table-displayable rows
// with severity coloring support for the by-severity breakdown.
type AlertStatsFormatter struct {
	colorMgr *sdkoutput.ColorManager
}

// NewAlertStatsFormatter creates a new formatter for alert stats output.
// The colorMode parameter controls whether severity colors are applied.
func NewAlertStatsFormatter(colorMode sdkoutput.ColorMode) *AlertStatsFormatter {
	return &AlertStatsFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
	}
}

// FormatAlertStats converts SDK AlertStats into table-displayable rows.
// The output shows summary statistics followed by severity breakdown.
func (f *AlertStatsFormatter) FormatAlertStats(stats *client.AlertStats) []AlertStatsRow {
	rows := []AlertStatsRow{
		{Metric: "Period", Value: stats.Period},
		{Metric: "Total Alerts", Value: fmt.Sprintf("%d", stats.TotalAlerts)},
		{Metric: "Active", Value: f.colorMgr.StatusDown(fmt.Sprintf("%d", stats.ActiveAlerts))},
		{Metric: "Acknowledged", Value: f.colorMgr.StatusWarning(fmt.Sprintf("%d", stats.AcknowledgedAlerts))},
		{Metric: "Resolved", Value: f.colorMgr.StatusUp(fmt.Sprintf("%d", stats.ResolvedAlerts))},
		{Metric: "", Value: ""}, // Separator row
		{Metric: "By Severity", Value: ""},
		{Metric: "  Critical", Value: f.colorMgr.StatusDown(fmt.Sprintf("%d", stats.CriticalAlerts))},
		{Metric: "  Warning", Value: f.colorMgr.StatusWarning(fmt.Sprintf("%d", stats.WarningAlerts))},
		{Metric: "  Info", Value: f.colorMgr.StatusInfo(fmt.Sprintf("%d", stats.InfoAlerts))},
	}

	// Add MTTA if available
	if stats.MTTA != nil {
		rows = append(rows, AlertStatsRow{
			Metric: "",
			Value:  "",
		})
		rows = append(rows, AlertStatsRow{
			Metric: "MTTA (Mean Time To Acknowledge)",
			Value:  formatDurationSeconds(*stats.MTTA),
		})
	}

	// Add MTTR if available
	if stats.MTTR != nil {
		if stats.MTTA == nil {
			rows = append(rows, AlertStatsRow{
				Metric: "",
				Value:  "",
			})
		}
		rows = append(rows, AlertStatsRow{
			Metric: "MTTR (Mean Time To Resolve)",
			Value:  formatDurationSeconds(*stats.MTTR),
		})
	}

	return rows
}

// formatDurationSeconds converts seconds to a human-readable format.
func formatDurationSeconds(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		mins := seconds / 60
		secs := seconds % 60
		if secs > 0 {
			return fmt.Sprintf("%dm%ds", mins, secs)
		}
		return fmt.Sprintf("%dm", mins)
	}
	hours := seconds / 3600
	mins := (seconds % 3600) / 60
	if mins > 0 {
		return fmt.Sprintf("%dh%dm", hours, mins)
	}
	return fmt.Sprintf("%dh", hours)
}

// PrintAlertStats is a convenience function that formats and prints alert statistics
// using the CLI's configured output format. It handles severity coloring
// automatically based on configuration.
func PrintAlertStats(stats *client.AlertStats) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	// For JSON/YAML output, print the raw stats object
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(stats)
	}

	// For table output, format as key-value pairs
	formatter := NewAlertStatsFormatter(colorMode)
	rows := formatter.FormatAlertStats(stats)

	return printer.Print(rows)
}
