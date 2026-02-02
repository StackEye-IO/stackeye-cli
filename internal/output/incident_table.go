// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"
	"strconv"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// IncidentTableRow represents a row in the incident table output.
// The struct tags control column headers and wide mode display.
type IncidentTableRow struct {
	ID      string `table:"ID"`
	Title   string `table:"TITLE"`
	Status  string `table:"STATUS"`
	Impact  string `table:"IMPACT"`
	Created string `table:"CREATED"`
	// Wide mode columns
	Updated  string `table:"UPDATED,wide"`
	Resolved string `table:"RESOLVED,wide"`
}

// IncidentTableFormatter converts SDK Incident types to table-displayable rows
// with status coloring support.
type IncidentTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewIncidentTableFormatter creates a new formatter for incident table output.
// The colorMode parameter controls whether status colors are applied.
// Set isWide to true for extended output with additional columns.
func NewIncidentTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *IncidentTableFormatter {
	return &IncidentTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatIncidents converts a slice of SDK Incidents into table-displayable rows.
func (f *IncidentTableFormatter) FormatIncidents(incidents []client.Incident) []IncidentTableRow {
	rows := make([]IncidentTableRow, 0, len(incidents))
	for _, inc := range incidents {
		rows = append(rows, f.formatIncident(inc))
	}
	return rows
}

// FormatIncident converts a single SDK Incident into a table-displayable row.
func (f *IncidentTableFormatter) FormatIncident(incident client.Incident) IncidentTableRow {
	return f.formatIncident(incident)
}

// formatIncident is the internal conversion function.
func (f *IncidentTableFormatter) formatIncident(inc client.Incident) IncidentTableRow {
	return IncidentTableRow{
		ID:       strconv.FormatUint(uint64(inc.ID), 10),
		Title:    truncateIncidentTitle(inc.Title, 40),
		Status:   f.formatStatus(inc.Status),
		Impact:   f.formatImpact(inc.Impact),
		Created:  inc.CreatedAt.Format("2006-01-02 15:04"),
		Updated:  inc.UpdatedAt.Format("2006-01-02 15:04"),
		Resolved: formatResolvedAt(inc.ResolvedAt),
	}
}

// formatStatus applies color based on incident status.
// Statuses: investigating, identified, monitoring, resolved
func (f *IncidentTableFormatter) formatStatus(status string) string {
	switch status {
	case "resolved":
		return f.colorMgr.StatusUp("Resolved")
	case "monitoring":
		return f.colorMgr.StatusUp("Monitoring")
	case "identified":
		return f.colorMgr.Warning("Identified")
	case "investigating":
		return f.colorMgr.StatusDown("Investigating")
	default:
		return f.colorMgr.Dim(status)
	}
}

// formatImpact applies color based on incident impact level.
// Impact levels: none, minor, major, critical
func (f *IncidentTableFormatter) formatImpact(impact string) string {
	switch impact {
	case "none":
		return f.colorMgr.Dim("None")
	case "minor":
		return f.colorMgr.Warning("Minor")
	case "major":
		return f.colorMgr.StatusDown("Major")
	case "critical":
		return f.colorMgr.StatusDown("Critical")
	default:
		return f.colorMgr.Dim(impact)
	}
}

// truncateIncidentTitle truncates an incident title for display.
// Uses rune-based slicing to properly handle Unicode characters.
func truncateIncidentTitle(title string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	runes := []rune(title)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return title
}

// formatResolvedAt formats the resolved timestamp for display.
func formatResolvedAt(resolvedAt *time.Time) string {
	if resolvedAt == nil {
		return "-"
	}
	return resolvedAt.Format("2006-01-02 15:04")
}

// PrintIncidents is a convenience function that formats and prints incidents
// using the CLI's configured output format. It handles status coloring
// and wide mode automatically based on configuration.
func PrintIncidents(incidents []client.Incident) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewIncidentTableFormatter(colorMode, isWide)
	rows := formatter.FormatIncidents(incidents)

	return printer.Print(rows)
}

// PrintIncident is a convenience function that formats and prints a single incident.
func PrintIncident(incident client.Incident) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewIncidentTableFormatter(colorMode, isWide)
	row := formatter.FormatIncident(incident)

	return printer.Print(row)
}

// FormatIncidentCount formats the total count for pagination display.
func FormatIncidentCount(total int64, page, limit int) string {
	if total == 0 {
		return ""
	}
	start := (page-1)*limit + 1
	end := start + limit - 1
	if int64(end) > total {
		end = int(total)
	}
	return fmt.Sprintf("Showing %d-%d of %d incidents", start, end, total)
}

// IncidentDetail represents a detailed view of a single incident.
// Used for displaying full incident information including the message body.
type IncidentDetail struct {
	ID         string
	Title      string
	Status     string
	Impact     string
	Message    string
	CreatedAt  string
	UpdatedAt  string
	ResolvedAt string
	Duration   string
}

// FormatIncidentDetail converts a single SDK Incident into a detailed view
// showing the full message and timeline information.
func (f *IncidentTableFormatter) FormatIncidentDetail(incident client.Incident) IncidentDetail {
	return IncidentDetail{
		ID:         strconv.FormatUint(uint64(incident.ID), 10),
		Title:      incident.Title,
		Status:     f.formatStatus(incident.Status),
		Impact:     f.formatImpact(incident.Impact),
		Message:    incident.Message,
		CreatedAt:  incident.CreatedAt.Format("2006-01-02 15:04:05 MST"),
		UpdatedAt:  incident.UpdatedAt.Format("2006-01-02 15:04:05 MST"),
		ResolvedAt: formatResolvedAtDetail(incident.ResolvedAt),
		Duration:   formatIncidentDuration(incident.CreatedAt, incident.ResolvedAt),
	}
}

// formatResolvedAtDetail formats the resolved timestamp for detailed display.
func formatResolvedAtDetail(resolvedAt *time.Time) string {
	if resolvedAt == nil {
		return "Not resolved"
	}
	return resolvedAt.Format("2006-01-02 15:04:05 MST")
}

// formatIncidentDuration calculates and formats the incident duration.
func formatIncidentDuration(createdAt time.Time, resolvedAt *time.Time) string {
	var end time.Time
	if resolvedAt != nil {
		end = *resolvedAt
	} else {
		end = time.Now()
	}

	duration := end.Sub(createdAt)
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		mins := int(duration.Minutes()) % 60
		if mins > 0 {
			return fmt.Sprintf("%dh %dm", hours, mins)
		}
		return fmt.Sprintf("%dh", hours)
	}
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}

// PrintIncidentDetail prints a single incident in detailed format.
// This shows the full message and timeline information.
func PrintIncidentDetail(incident client.Incident) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	// For JSON output, return the full incident data
	if printer.Format() == sdkoutput.FormatJSON {
		return printer.Print(incident)
	}

	formatter := NewIncidentTableFormatter(colorMode, false)
	detail := formatter.FormatIncidentDetail(incident)

	// Build formatted output for table/text mode
	output := fmt.Sprintf(`Incident #%s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Title:      %s
Status:     %s
Impact:     %s

Message:
%s

Timeline:
  Created:  %s
  Updated:  %s
  Resolved: %s
  Duration: %s
`,
		detail.ID,
		detail.Title,
		detail.Status,
		detail.Impact,
		detail.Message,
		detail.CreatedAt,
		detail.UpdatedAt,
		detail.ResolvedAt,
		detail.Duration,
	)

	fmt.Print(output)
	return nil
}
