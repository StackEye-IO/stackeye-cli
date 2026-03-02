// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// ProbeGroupTableRow represents a row in probe group table output.
type ProbeGroupTableRow struct {
	ID          string `table:"ID"`
	Name        string `table:"NAME"`
	Description string `table:"DESCRIPTION"`
	Probes      int64  `table:"PROBES"`
	Updated     string `table:"UPDATED"`
	// Wide mode columns
	Type    string `table:"TYPE,wide"`
	Created string `table:"CREATED,wide"`
}

// GroupProbeTableRow represents a row in group probe membership table output.
type GroupProbeTableRow struct {
	ID     string `table:"ID"`
	Name   string `table:"NAME"`
	URL    string `table:"URL"`
	Status string `table:"STATUS"`
	// Wide mode columns
	Type    string `table:"TYPE,wide"`
	Updated string `table:"UPDATED,wide"`
}

// ProbeGroupTableFormatter formats probe groups for table output.
type ProbeGroupTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewProbeGroupTableFormatter creates a new probe group formatter.
func NewProbeGroupTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *ProbeGroupTableFormatter {
	return &ProbeGroupTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatProbeGroups converts probe groups to table rows.
func (f *ProbeGroupTableFormatter) FormatProbeGroups(groups []client.ProbeGroup) []ProbeGroupTableRow {
	rows := make([]ProbeGroupTableRow, 0, len(groups))
	for _, g := range groups {
		rows = append(rows, ProbeGroupTableRow{
			ID:          g.ID.String(),
			Name:        truncateGroupText(g.Name, 36),
			Description: formatGroupDescription(g.Description),
			Probes:      g.MemberCount,
			Updated:     formatGroupTimestamp(g.UpdatedAt),
			Type:        string(g.Type),
			Created:     g.CreatedAt.Format("2006-01-02"),
		})
	}
	return rows
}

// FormatGroupProbes converts probes to table rows for group membership views.
func (f *ProbeGroupTableFormatter) FormatGroupProbes(probes []client.Probe) []GroupProbeTableRow {
	rows := make([]GroupProbeTableRow, 0, len(probes))
	for _, p := range probes {
		rows = append(rows, GroupProbeTableRow{
			ID:      p.ID.String(),
			Name:    truncateGroupText(p.Name, 32),
			URL:     truncateGroupText(p.URL, 44),
			Status:  f.formatProbeStatus(p.Status),
			Type:    string(p.CheckType),
			Updated: p.UpdatedAt.Format("2006-01-02"),
		})
	}
	return rows
}

func (f *ProbeGroupTableFormatter) formatProbeStatus(status string) string {
	s := strings.ToLower(status)
	switch s {
	case "up":
		return f.colorMgr.StatusUp("UP")
	case "down":
		return f.colorMgr.StatusDown("DOWN")
	case "degraded":
		return f.colorMgr.Warning("DEGRADED")
	case "paused":
		return f.colorMgr.Dim("PAUSED")
	case "pending":
		return f.colorMgr.Dim("PENDING")
	default:
		return strings.ToUpper(status)
	}
}

func formatGroupDescription(desc *string) string {
	if desc == nil || strings.TrimSpace(*desc) == "" {
		return "-"
	}
	return truncateGroupText(*desc, 40)
}

func truncateGroupText(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 4 {
		maxLen = 4
	}
	return string(runes[:maxLen-3]) + "..."
}

func formatGroupTimestamp(ts time.Time) string {
	if ts.IsZero() {
		return "-"
	}

	elapsed := time.Since(ts)
	switch {
	case elapsed < time.Minute:
		return "just now"
	case elapsed < time.Hour:
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	case elapsed < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	default:
		return ts.Format("2006-01-02")
	}
}

// PrintProbeGroups prints probe groups respecting current output format.
func PrintProbeGroups(groups []client.ProbeGroup) error {
	printer := getPrinter()
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(groups)
	}

	colorMode := sdkoutput.ColorAuto
	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewProbeGroupTableFormatter(colorMode, format == sdkoutput.FormatWide)
	return printer.Print(formatter.FormatProbeGroups(groups))
}

// PrintGroupProbes prints group member probes respecting current output format.
func PrintGroupProbes(probes []client.Probe) error {
	printer := getPrinter()
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(probes)
	}

	colorMode := sdkoutput.ColorAuto
	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewProbeGroupTableFormatter(colorMode, format == sdkoutput.FormatWide)
	return printer.Print(formatter.FormatGroupProbes(probes))
}
