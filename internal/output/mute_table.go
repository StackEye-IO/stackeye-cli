// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// MuteTableRow represents a row in the mute table output.
// The struct tags control column headers and wide mode display.
type MuteTableRow struct {
	Status    string `table:"STATUS"`
	Scope     string `table:"SCOPE"`
	Target    string `table:"TARGET"`
	Duration  string `table:"DURATION"`
	ExpiresAt string `table:"EXPIRES"`
	Reason    string `table:"REASON"`
	// Wide mode columns
	Maintenance string `table:"MAINTENANCE,wide"`
	Created     string `table:"CREATED,wide"`
	ID          string `table:"ID,wide"`
}

// MuteTableFormatter converts SDK AlertMute types to table-displayable rows
// with status coloring support.
type MuteTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewMuteTableFormatter creates a new formatter for mute table output.
// The colorMode parameter controls whether status colors are applied.
// Set isWide to true for extended output with additional columns.
func NewMuteTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *MuteTableFormatter {
	return &MuteTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatMutes converts a slice of SDK AlertMutes into table-displayable rows.
// Status fields are colored based on mute state:
//   - active: green (ACTIVE)
//   - expired: yellow (EXPIRED)
func (f *MuteTableFormatter) FormatMutes(mutes []client.AlertMute) []MuteTableRow {
	rows := make([]MuteTableRow, 0, len(mutes))
	for _, m := range mutes {
		rows = append(rows, f.formatMute(m))
	}
	return rows
}

// FormatMute converts a single SDK AlertMute into a table-displayable row.
func (f *MuteTableFormatter) FormatMute(mute client.AlertMute) MuteTableRow {
	return f.formatMute(mute)
}

// formatMute is the internal conversion function.
func (f *MuteTableFormatter) formatMute(m client.AlertMute) MuteTableRow {
	return MuteTableRow{
		Status:      f.formatStatus(m.IsActive),
		Scope:       formatScopeType(m.ScopeType),
		Target:      formatMuteTarget(m),
		Duration:    formatMuteDuration(m.DurationMinutes),
		ExpiresAt:   formatExpiresAt(m.ExpiresAt),
		Reason:      formatReason(m.Reason),
		Maintenance: formatMaintenance(m.IsMaintenanceWindow, m.MaintenanceName),
		Created:     m.CreatedAt.Format("2006-01-02 15:04"),
		ID:          m.ID.String(),
	}
}

// formatStatus applies color based on mute active state.
func (f *MuteTableFormatter) formatStatus(isActive bool) string {
	if isActive {
		return f.colorMgr.StatusUp("ACTIVE")
	}
	return f.colorMgr.Warning("EXPIRED")
}

// formatScopeType converts the scope type to a human-readable string.
func formatScopeType(scopeType client.MuteScopeType) string {
	switch scopeType {
	case client.MuteScopeOrganization:
		return "Organization"
	case client.MuteScopeProbe:
		return "Probe"
	case client.MuteScopeChannel:
		return "Channel"
	case client.MuteScopeAlertType:
		return "Alert Type"
	default:
		return string(scopeType)
	}
}

// formatMuteTarget extracts and formats the target based on scope type.
func formatMuteTarget(m client.AlertMute) string {
	switch m.ScopeType {
	case client.MuteScopeOrganization:
		return "All"
	case client.MuteScopeProbe:
		if m.ProbeID != nil {
			return truncateUUID(m.ProbeID.String())
		}
		return "-"
	case client.MuteScopeChannel:
		if m.ChannelID != nil {
			return truncateUUID(m.ChannelID.String())
		}
		return "-"
	case client.MuteScopeAlertType:
		if m.AlertType != nil {
			// Use the formatAlertType from alert_table.go
			return formatAlertType(*m.AlertType)
		}
		return "-"
	default:
		return "-"
	}
}

// formatMuteDuration formats duration in minutes to human-readable string.
func formatMuteDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	remainingMins := minutes % 60
	if remainingMins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, remainingMins)
}

// formatExpiresAt formats the expiration time for display.
func formatExpiresAt(expiresAt *time.Time) string {
	if expiresAt == nil {
		return "Never"
	}
	now := time.Now()
	if expiresAt.Before(now) {
		return "Expired"
	}
	// Show relative time if within 24 hours
	remaining := time.Until(*expiresAt)
	if remaining < time.Hour {
		return fmt.Sprintf("in %dm", int(remaining.Minutes()))
	}
	if remaining < 24*time.Hour {
		return fmt.Sprintf("in %dh", int(remaining.Hours()))
	}
	return expiresAt.Format("2006-01-02 15:04")
}

// formatReason truncates reason for display.
func formatReason(reason *string) string {
	if reason == nil || *reason == "" {
		return "-"
	}
	r := *reason
	if len(r) > 30 {
		return r[:27] + "..."
	}
	return r
}

// formatMaintenance formats the maintenance window info.
func formatMaintenance(isMaintenanceWindow bool, name *string) string {
	if !isMaintenanceWindow {
		return "No"
	}
	if name != nil && *name != "" {
		n := *name
		if len(n) > 20 {
			return n[:17] + "..."
		}
		return n
	}
	return "Yes"
}

// truncateUUID shortens a UUID for display (first 8 chars).
func truncateUUID(id string) string {
	if len(id) > 8 {
		return id[:8] + "..."
	}
	return id
}

// PrintMutes is a convenience function that formats and prints mutes
// using the CLI's configured output format. It handles status coloring
// and wide mode automatically based on configuration.
func PrintMutes(mutes []client.AlertMute) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewMuteTableFormatter(colorMode, isWide)
	rows := formatter.FormatMutes(mutes)

	return printer.Print(rows)
}

// PrintMute is a convenience function that formats and prints a single mute.
func PrintMute(mute client.AlertMute) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewMuteTableFormatter(colorMode, isWide)
	row := formatter.FormatMute(mute)

	return printer.Print(row)
}
