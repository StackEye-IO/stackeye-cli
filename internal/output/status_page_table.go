// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"
	"strconv"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// StatusPageTableRow represents a row in the status page table output.
// The struct tags control column headers and wide mode display.
type StatusPageTableRow struct {
	Name    string `table:"NAME"`
	Slug    string `table:"SLUG"`
	Theme   string `table:"THEME"`
	Public  string `table:"PUBLIC"`
	Enabled string `table:"ENABLED"`
	Probes  string `table:"PROBES"`
	// Wide mode columns
	Domain  string `table:"DOMAIN,wide"`
	Uptime  string `table:"UPTIME%,wide"`
	ID      string `table:"ID,wide"`
	Created string `table:"CREATED,wide"`
}

// StatusPageTableFormatter converts SDK StatusPage types to table-displayable rows
// with status coloring support.
type StatusPageTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewStatusPageTableFormatter creates a new formatter for status page table output.
// The colorMode parameter controls whether status colors are applied.
// Set isWide to true for extended output with additional columns.
func NewStatusPageTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *StatusPageTableFormatter {
	return &StatusPageTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatStatusPages converts a slice of SDK StatusPages into table-displayable rows.
func (f *StatusPageTableFormatter) FormatStatusPages(pages []client.StatusPage) []StatusPageTableRow {
	rows := make([]StatusPageTableRow, 0, len(pages))
	for _, p := range pages {
		rows = append(rows, f.formatStatusPage(p))
	}
	return rows
}

// FormatStatusPage converts a single SDK StatusPage into a table-displayable row.
func (f *StatusPageTableFormatter) FormatStatusPage(page client.StatusPage) StatusPageTableRow {
	return f.formatStatusPage(page)
}

// formatStatusPage is the internal conversion function.
func (f *StatusPageTableFormatter) formatStatusPage(p client.StatusPage) StatusPageTableRow {
	return StatusPageTableRow{
		Name:    truncateName(p.Name, 30),
		Slug:    p.Slug,
		Theme:   formatTheme(p.Theme),
		Public:  f.formatPublic(p.IsPublic),
		Enabled: f.formatEnabled(p.Enabled),
		Probes:  strconv.Itoa(len(p.Probes)),
		Domain:  formatCustomDomain(p.CustomDomain),
		Uptime:  formatUptimeDisplay(p.ShowUptimePercentage),
		ID:      strconv.FormatUint(uint64(p.ID), 10),
		Created: p.CreatedAt.Format("2006-01-02"),
	}
}

// formatPublic applies color based on public visibility.
func (f *StatusPageTableFormatter) formatPublic(isPublic bool) string {
	if isPublic {
		return f.colorMgr.StatusUp("Yes")
	}
	return f.colorMgr.Dim("No")
}

// formatEnabled applies color based on enabled status.
func (f *StatusPageTableFormatter) formatEnabled(enabled bool) string {
	if enabled {
		return f.colorMgr.StatusUp("Yes")
	}
	return f.colorMgr.Warning("No")
}

// truncateName truncates a name for display.
func truncateName(name string, maxLen int) string {
	if len(name) > maxLen {
		return name[:maxLen-3] + "..."
	}
	return name
}

// formatTheme converts the theme value to a human-readable string.
func formatTheme(theme string) string {
	switch theme {
	case "light":
		return "Light"
	case "dark":
		return "Dark"
	case "system":
		return "System"
	default:
		if theme == "" {
			return "System"
		}
		return theme
	}
}

// formatCustomDomain formats the custom domain for display.
func formatCustomDomain(domain *string) string {
	if domain == nil || *domain == "" {
		return "-"
	}
	d := *domain
	if len(d) > 25 {
		return d[:22] + "..."
	}
	return d
}

// formatUptimeDisplay formats the uptime percentage display setting.
func formatUptimeDisplay(showUptime bool) string {
	if showUptime {
		return "Yes"
	}
	return "No"
}

// PrintStatusPages is a convenience function that formats and prints status pages
// using the CLI's configured output format. It handles status coloring
// and wide mode automatically based on configuration.
func PrintStatusPages(pages []client.StatusPage) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewStatusPageTableFormatter(colorMode, isWide)
	rows := formatter.FormatStatusPages(pages)

	return printer.Print(rows)
}

// PrintStatusPage is a convenience function that formats and prints a single status page.
func PrintStatusPage(page client.StatusPage) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewStatusPageTableFormatter(colorMode, isWide)
	row := formatter.FormatStatusPage(page)

	return printer.Print(row)
}

// FormatStatusPageCount formats the total count for pagination display.
func FormatStatusPageCount(total int64, page, limit int) string {
	if total == 0 {
		return ""
	}
	start := (page-1)*limit + 1
	end := start + limit - 1
	if int64(end) > total {
		end = int(total)
	}
	return fmt.Sprintf("Showing %d-%d of %d status pages", start, end, total)
}
