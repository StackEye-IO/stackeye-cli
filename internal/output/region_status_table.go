// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"sort"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// RegionStatusTableRow represents a row in the region status table output.
// The struct tags control column headers and wide mode display.
type RegionStatusTableRow struct {
	Code        string `table:"CODE"`
	Name        string `table:"NAME"`
	Status      string `table:"STATUS"`
	Health      string `table:"HEALTH"`
	Maintenance string `table:"MAINTENANCE,wide"`
}

// RegionStatusFormatter converts SDK RegionStatus types to table-displayable rows.
type RegionStatusFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewRegionStatusFormatter creates a new formatter for region status table output.
// The colorMode parameter controls whether colors are applied.
// Set isWide to true for extended output with maintenance column.
func NewRegionStatusFormatter(colorMode sdkoutput.ColorMode, isWide bool) *RegionStatusFormatter {
	return &RegionStatusFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatRegionStatuses converts a slice of region statuses into table-displayable rows.
// Statuses are sorted by region name for consistent output.
func (f *RegionStatusFormatter) FormatRegionStatuses(statuses []client.RegionStatus) []RegionStatusTableRow {
	rows := make([]RegionStatusTableRow, 0, len(statuses))

	// Sort by name for consistent output
	sortedStatuses := make([]client.RegionStatus, len(statuses))
	copy(sortedStatuses, statuses)
	sort.Slice(sortedStatuses, func(i, j int) bool {
		return sortedStatuses[i].Name < sortedStatuses[j].Name
	})

	for _, status := range sortedStatuses {
		rows = append(rows, f.formatRegionStatus(status))
	}

	return rows
}

// FormatRegionStatus converts a single SDK RegionStatus into a table-displayable row.
func (f *RegionStatusFormatter) FormatRegionStatus(status client.RegionStatus) RegionStatusTableRow {
	return f.formatRegionStatus(status)
}

// formatRegionStatus is the internal conversion function.
func (f *RegionStatusFormatter) formatRegionStatus(status client.RegionStatus) RegionStatusTableRow {
	return RegionStatusTableRow{
		Code:        status.ID,
		Name:        status.Name,
		Status:      f.formatOperationalStatus(status.Status),
		Health:      f.formatHealthStatus(status.HealthStatus),
		Maintenance: f.formatMaintenance(status.MaintenanceReason, status.MaintenanceEndsAt),
	}
}

// formatOperationalStatus applies coloring to operational status.
func (f *RegionStatusFormatter) formatOperationalStatus(status string) string {
	switch status {
	case "active":
		return f.colorMgr.StatusUp(status)
	case "maintenance":
		return f.colorMgr.StatusWarning(status)
	case "disabled":
		return f.colorMgr.StatusDown(status)
	default:
		return status
	}
}

// formatHealthStatus applies coloring to health status.
func (f *RegionStatusFormatter) formatHealthStatus(health string) string {
	switch health {
	case "healthy":
		return f.colorMgr.StatusUp(health)
	case "warning":
		return f.colorMgr.StatusWarning(health)
	case "degraded":
		return f.colorMgr.StatusDown(health)
	case "unknown":
		return f.colorMgr.StatusInfo(health)
	default:
		return health
	}
}

// formatMaintenance formats maintenance information for display.
func (f *RegionStatusFormatter) formatMaintenance(reason *string, endsAt *time.Time) string {
	if reason == nil && endsAt == nil {
		return "-"
	}

	result := ""
	if reason != nil && *reason != "" {
		result = *reason
	}
	if endsAt != nil {
		if result != "" {
			result += " (until " + endsAt.Format("Jan 2 15:04 MST") + ")"
		} else {
			result = "Until " + endsAt.Format("Jan 2 15:04 MST")
		}
	}
	if result == "" {
		return "-"
	}
	return result
}

// PrintRegionStatuses is a convenience function that formats and prints region statuses
// using the CLI's configured output format. It handles wide mode automatically.
func PrintRegionStatuses(statuses []client.RegionStatus) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	// For JSON/YAML output, print the raw statuses
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(statuses)
	}

	formatter := NewRegionStatusFormatter(colorMode, isWide)
	rows := formatter.FormatRegionStatuses(statuses)

	return printer.Print(rows)
}

// PrintRegionStatus is a convenience function for printing a single region status.
func PrintRegionStatus(status client.RegionStatus) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	// For JSON/YAML output, print the raw status
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(status)
	}

	formatter := NewRegionStatusFormatter(colorMode, isWide)
	row := formatter.FormatRegionStatus(status)

	return printer.Print([]RegionStatusTableRow{row})
}
