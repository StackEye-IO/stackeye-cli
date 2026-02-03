// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/google/uuid"
)

// AlertTableRow represents a row in the alert table output.
// The struct tags control column headers and wide mode display.
type AlertTableRow struct {
	Severity  string `table:"SEVERITY"`
	Status    string `table:"STATUS"`
	Type      string `table:"TYPE"`
	Probe     string `table:"PROBE"`
	Triggered string `table:"TRIGGERED"`
	Duration  string `table:"DURATION"`
	// Wide mode columns
	AckBy   string `table:"ACK BY,wide"`
	Message string `table:"MESSAGE,wide"`
	ID      string `table:"ID,wide"`
}

// AlertTableFormatter converts SDK Alert types to table-displayable rows
// with severity and status coloring support.
type AlertTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewAlertTableFormatter creates a new formatter for alert table output.
// The colorMode parameter controls whether severity/status colors are applied.
// Set isWide to true for extended output with additional columns.
func NewAlertTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *AlertTableFormatter {
	return &AlertTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatAlerts converts a slice of SDK Alerts into table-displayable rows.
// Severity fields are colored based on alert level:
//   - critical: red
//   - warning: yellow
//   - info: cyan
//
// Status fields are colored based on alert state:
//   - active: red
//   - acknowledged: yellow
//   - resolved: green
func (f *AlertTableFormatter) FormatAlerts(alerts []client.Alert) []AlertTableRow {
	rows := make([]AlertTableRow, 0, len(alerts))
	for _, a := range alerts {
		rows = append(rows, f.formatAlert(a))
	}
	return rows
}

// FormatAlert converts a single SDK Alert into a table-displayable row.
func (f *AlertTableFormatter) FormatAlert(alert client.Alert) AlertTableRow {
	return f.formatAlert(alert)
}

// formatAlert is the internal conversion function.
func (f *AlertTableFormatter) formatAlert(a client.Alert) AlertTableRow {
	return AlertTableRow{
		Severity:  f.formatSeverity(a.Severity),
		Status:    f.formatStatus(a.Status),
		Type:      formatAlertType(a.AlertType),
		Probe:     formatProbeName(a.Probe),
		Triggered: formatTriggeredTime(a.TriggeredAt),
		Duration:  formatAlertDuration(a.TriggeredAt, a.ResolvedAt, a.DurationSeconds),
		// Wide mode fields
		AckBy:   formatAcknowledgedBy(a.AcknowledgedBy),
		Message: truncateMessage(a.Message, 40),
		ID:      a.ID.String(),
	}
}

// formatSeverity applies color based on alert severity.
func (f *AlertTableFormatter) formatSeverity(severity client.AlertSeverity) string {
	upperSeverity := strings.ToUpper(string(severity))
	switch severity {
	case client.AlertSeverityCritical:
		return f.colorMgr.StatusDown(upperSeverity)
	case client.AlertSeverityWarning:
		return f.colorMgr.StatusWarning(upperSeverity)
	case client.AlertSeverityInfo:
		return f.colorMgr.StatusInfo(upperSeverity)
	default:
		return upperSeverity
	}
}

// formatStatus applies color based on alert status.
func (f *AlertTableFormatter) formatStatus(status client.AlertStatus) string {
	upperStatus := strings.ToUpper(string(status))
	switch status {
	case client.AlertStatusActive:
		return f.colorMgr.StatusDown(upperStatus)
	case client.AlertStatusAcknowledged:
		return f.colorMgr.StatusWarning(upperStatus)
	case client.AlertStatusResolved:
		return f.colorMgr.StatusUp(upperStatus)
	default:
		return upperStatus
	}
}

// formatAlertType converts the alert type to a human-readable string.
func formatAlertType(alertType client.AlertType) string {
	switch alertType {
	case client.AlertTypeStatusDown:
		return "Down"
	case client.AlertTypeSSLExpiry:
		return "SSL Expiry"
	case client.AlertTypeSSLInvalid:
		return "SSL Invalid"
	case client.AlertTypeSlowResponse:
		return "Slow Response"
	default:
		return string(alertType)
	}
}

// formatProbeName extracts the probe name from the nested probe info.
func formatProbeName(probe *client.AlertProbe) string {
	if probe == nil {
		return "-"
	}
	return truncateString(probe.Name, 30)
}

// formatTriggeredTime converts the triggered timestamp to relative time.
func formatTriggeredTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	elapsed := time.Since(t)

	switch {
	case elapsed < time.Minute:
		return "just now"
	case elapsed < time.Hour:
		mins := int(elapsed.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case elapsed < 24*time.Hour:
		hours := int(elapsed.Hours())
		return fmt.Sprintf("%dh ago", hours)
	default:
		days := int(elapsed.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}

// formatAlertDuration calculates and formats the alert duration.
// For resolved alerts, uses the stored duration or calculates from timestamps.
// For active alerts, calculates elapsed time since trigger.
func formatAlertDuration(triggered time.Time, resolved *time.Time, durationSeconds *int) string {
	if triggered.IsZero() {
		return "-"
	}

	var duration time.Duration

	if durationSeconds != nil && *durationSeconds > 0 {
		duration = time.Duration(*durationSeconds) * time.Second
	} else if resolved != nil {
		duration = resolved.Sub(triggered)
	} else {
		duration = time.Since(triggered)
	}

	return formatDuration(duration)
}

// formatDuration converts a duration to a human-readable string.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		mins := int(d.Minutes())
		return fmt.Sprintf("%dm", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		if mins > 0 {
			return fmt.Sprintf("%dh%dm", hours, mins)
		}
		return fmt.Sprintf("%dh", hours)
	default:
		days := int(d.Hours() / 24)
		hours := int(d.Hours()) % 24
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}
}

// formatAcknowledgedBy formats the acknowledged by UUID for display.
func formatAcknowledgedBy(ackBy *uuid.UUID) string {
	if ackBy == nil {
		return "-"
	}
	return ackBy.String()[:8] + "..."
}

// truncateMessage shortens a message to fit in the table column.
func truncateMessage(msg *string, maxLen int) string {
	if msg == nil || *msg == "" {
		return "-"
	}
	return truncateString(*msg, maxLen)
}

// truncateString shortens a string to the specified maximum length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// PrintAlerts is a convenience function that formats and prints alerts
// using the CLI's configured output format. It handles severity/status
// coloring and wide mode automatically based on configuration.
func PrintAlerts(alerts []client.Alert) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewAlertTableFormatter(colorMode, isWide)
	rows := formatter.FormatAlerts(alerts)

	return printer.Print(rows)
}

// PrintAlert is a convenience function that formats and prints a single alert.
func PrintAlert(alert client.Alert) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewAlertTableFormatter(colorMode, isWide)
	row := formatter.FormatAlert(alert)

	return printer.Print(row)
}
