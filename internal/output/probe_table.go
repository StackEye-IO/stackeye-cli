// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// ProbeTableRow represents a row in the probe table output.
// The struct tags control column headers and wide mode display.
type ProbeTableRow struct {
	Status    string `table:"STATUS"`
	Name      string `table:"NAME"`
	URL       string `table:"URL"`
	Deps      string `table:"DEPS"`
	Interval  string `table:"INTERVAL"`
	LastCheck string `table:"LAST CHECK"`
	// Wide mode columns
	Type    string `table:"TYPE,wide"`
	Uptime  string `table:"UPTIME,wide"`
	AvgResp string `table:"AVG RESP,wide"`
	Regions string `table:"REGIONS,wide"`
	Labels  string `table:"LABELS,wide"` // Task #8070
	ID      string `table:"ID,wide"`
}

// ProbeTableFormatter converts SDK Probe types to table-displayable rows
// with status coloring support.
type ProbeTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewProbeTableFormatter creates a new formatter for probe table output.
// The colorMode parameter controls whether status colors are applied.
// Set isWide to true for extended output with additional columns.
func NewProbeTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *ProbeTableFormatter {
	return &ProbeTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatProbes converts a slice of SDK Probes into table-displayable rows.
// Status fields are colored based on probe state:
//   - up/healthy: green
//   - down/error: red
//   - degraded: yellow
//   - paused/pending: no color (plain text)
func (f *ProbeTableFormatter) FormatProbes(probes []client.Probe) []ProbeTableRow {
	rows := make([]ProbeTableRow, 0, len(probes))
	for _, p := range probes {
		rows = append(rows, f.formatProbe(p))
	}
	return rows
}

// FormatProbe converts a single SDK Probe into a table-displayable row.
func (f *ProbeTableFormatter) FormatProbe(probe client.Probe) ProbeTableRow {
	return f.formatProbe(probe)
}

// formatProbe is the internal conversion function.
func (f *ProbeTableFormatter) formatProbe(p client.Probe) ProbeTableRow {
	// Use "unreachable" status if probe is marked as unreachable due to parent dependency
	status := p.Status
	if p.IsUnreachable {
		status = "unreachable"
	}

	return ProbeTableRow{
		Status:    f.formatStatus(status),
		Name:      p.Name,
		URL:       truncateURL(p.URL, 50),
		Deps:      formatDeps(p.ParentCount, p.ChildCount),
		Interval:  formatInterval(p.IntervalSeconds),
		LastCheck: formatLastCheck(p.LastCheckedAt),
		// Wide mode fields
		Type:    string(p.CheckType),
		Uptime:  formatUptime(p.Uptime),
		AvgResp: formatResponseTime(p.AvgResponseTimeMs),
		Regions: formatRegions(p.Regions),
		Labels:  formatProbeLabels(p.Labels), // Task #8070
		ID:      p.ID.String(),
	}
}

// formatProbeLabels converts probe labels to a display string.
// Shows up to 3 labels, with "+N" indicator for additional labels.
// Task #8070
func formatProbeLabels(labels []client.ProbeLabel) string {
	if len(labels) == 0 {
		return "-"
	}

	const maxDisplayLabels = 3
	const maxTotalLength = 40

	parts := make([]string, 0, len(labels))
	for _, l := range labels {
		if l.Value != nil && *l.Value != "" {
			parts = append(parts, l.Key+"="+*l.Value)
		} else {
			parts = append(parts, l.Key)
		}
	}

	if len(parts) <= maxDisplayLabels {
		result := strings.Join(parts, ", ")
		if len(result) > maxTotalLength {
			return result[:maxTotalLength-3] + "..."
		}
		return result
	}

	// Show first 3 labels with "+N" indicator
	displayed := strings.Join(parts[:maxDisplayLabels], ", ")
	remaining := len(parts) - maxDisplayLabels
	return fmt.Sprintf("%s +%d", displayed, remaining)
}

// formatStatus applies color based on probe status.
func (f *ProbeTableFormatter) formatStatus(status string) string {
	upperStatus := strings.ToUpper(status)
	return f.colorMgr.StatusColor(upperStatus)
}

// truncateURL shortens a URL to fit in the table column.
func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}

// formatInterval converts seconds to a human-readable interval.
func formatInterval(seconds int) string {
	switch {
	case seconds < 60:
		return fmt.Sprintf("%ds", seconds)
	case seconds < 3600:
		mins := seconds / 60
		return fmt.Sprintf("%dm", mins)
	default:
		hours := seconds / 3600
		return fmt.Sprintf("%dh", hours)
	}
}

// formatLastCheck converts a timestamp to relative or absolute time.
func formatLastCheck(t *time.Time) string {
	if t == nil {
		return "never"
	}

	elapsed := time.Since(*t)

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

// formatUptime converts uptime percentage to a display string.
func formatUptime(uptime float64) string {
	if uptime == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f%%", uptime)
}

// formatResponseTime converts milliseconds to a display string.
func formatResponseTime(ms float64) string {
	if ms == 0 {
		return "-"
	}
	if ms < 1000 {
		return fmt.Sprintf("%.0fms", ms)
	}
	return fmt.Sprintf("%.2fs", ms/1000)
}

// formatRegions joins region codes into a comma-separated list.
func formatRegions(regions []string) string {
	if len(regions) == 0 {
		return "-"
	}
	if len(regions) > 3 {
		return fmt.Sprintf("%s +%d", strings.Join(regions[:3], ","), len(regions)-3)
	}
	return strings.Join(regions, ",")
}

// formatDeps formats the dependency counts as "parents/children".
// Returns "-" when probe has no dependencies (both counts are 0).
func formatDeps(parentCount, childCount int) string {
	if parentCount == 0 && childCount == 0 {
		return "-"
	}
	return fmt.Sprintf("%d/%d", parentCount, childCount)
}

// PrintProbes is a convenience function that formats and prints probes
// using the CLI's configured output format. It handles status coloring
// and wide mode automatically based on configuration.
func PrintProbes(probes []client.Probe) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewProbeTableFormatter(colorMode, isWide)
	rows := formatter.FormatProbes(probes)

	return printer.Print(rows)
}

// PrintProbe is a convenience function that formats and prints a single probe.
func PrintProbe(probe client.Probe) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewProbeTableFormatter(colorMode, isWide)
	row := formatter.FormatProbe(probe)

	return printer.Print(row)
}
