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

// truncateName truncates a name for display, handling Unicode correctly.
func truncateName(name string, maxLen int) string {
	// Minimum length required to show at least one character plus "..."
	if maxLen < 4 {
		maxLen = 4
	}
	runes := []rune(name)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
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

// formatCustomDomain formats the custom domain for display, handling Unicode correctly.
func formatCustomDomain(domain *string) string {
	if domain == nil || *domain == "" {
		return "-"
	}
	d := *domain
	runes := []rune(d)
	if len(runes) > 25 {
		return string(runes[:22]) + "..."
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
// For JSON/YAML output, it prints the raw StatusPage objects including all probe details.
// For table output, it formats the data into human-readable rows.
func PrintStatusPages(pages []client.StatusPage) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	// For JSON/YAML output, print the raw status pages with full probe details
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(pages)
	}

	// For table output, format as human-readable rows
	formatter := NewStatusPageTableFormatter(colorMode, isWide)
	rows := formatter.FormatStatusPages(pages)

	return printer.Print(rows)
}

// PrintStatusPage is a convenience function that formats and prints a single status page.
// For JSON/YAML output, it prints the raw StatusPage object including all probe details.
// For table output, it formats the data into a human-readable row.
func PrintStatusPage(page client.StatusPage) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	// For JSON/YAML output, print the raw status page object with full probe details
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(page)
	}

	// For table output, format as human-readable row
	formatter := NewStatusPageTableFormatter(colorMode, isWide)
	row := formatter.FormatStatusPage(page)

	return printer.Print(row)
}

// FormatStatusPageCount formats the total count for pagination display.
func FormatStatusPageCount(total int64, page, limit int) string {
	if total == 0 {
		return ""
	}
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20 // Default limit
	}
	start := (page-1)*limit + 1
	// Handle case where start is beyond total
	if int64(start) > total {
		return ""
	}
	end := start + limit - 1
	if int64(end) > total {
		end = int(total)
	}
	return fmt.Sprintf("Showing %d-%d of %d status pages", start, end, total)
}

// ProbeStatusTableRow represents a row in the aggregated status table output.
// The struct tags control column headers and wide mode display.
type ProbeStatusTableRow struct {
	Name         string `table:"NAME"`
	Status       string `table:"STATUS"`
	Uptime       string `table:"UPTIME"`
	ResponseTime string `table:"RESPONSE"`
	// Wide mode columns
	ProbeID string `table:"PROBE ID,wide"`
}

// AggregatedStatusTableFormatter converts SDK AggregatedStatusResponse to table rows
// with status coloring support.
type AggregatedStatusTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewAggregatedStatusTableFormatter creates a new formatter for aggregated status output.
func NewAggregatedStatusTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *AggregatedStatusTableFormatter {
	return &AggregatedStatusTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatProbeStatuses converts a slice of SDK ProbeStatusSummary into table rows.
func (f *AggregatedStatusTableFormatter) FormatProbeStatuses(probes []client.ProbeStatusSummary) []ProbeStatusTableRow {
	rows := make([]ProbeStatusTableRow, 0, len(probes))
	for _, p := range probes {
		rows = append(rows, f.formatProbeStatus(p))
	}
	return rows
}

// formatProbeStatus converts a single probe status into a table row.
func (f *AggregatedStatusTableFormatter) formatProbeStatus(p client.ProbeStatusSummary) ProbeStatusTableRow {
	return ProbeStatusTableRow{
		Name:         truncateName(p.DisplayName, 30),
		Status:       f.formatProbeStatusValue(p.Status),
		Uptime:       formatUptimePercent(p.UptimePercent),
		ResponseTime: formatStatusResponseTime(p.ResponseTimeMs, p.ShowResponseTime),
		ProbeID:      p.ProbeID.String(),
	}
}

// formatProbeStatusValue applies color based on probe status.
func (f *AggregatedStatusTableFormatter) formatProbeStatusValue(status string) string {
	switch status {
	case "up":
		return f.colorMgr.StatusUp("Up")
	case "down":
		return f.colorMgr.StatusDown("Down")
	case "degraded":
		return f.colorMgr.Warning("Degraded")
	case "pending":
		return f.colorMgr.Dim("Pending")
	default:
		return f.colorMgr.Dim(status)
	}
}

// formatOverallStatus applies color based on overall page status.
func (f *AggregatedStatusTableFormatter) formatOverallStatus(status string) string {
	switch status {
	case "operational":
		return f.colorMgr.StatusUp("Operational")
	case "degraded":
		return f.colorMgr.Warning("Degraded")
	case "outage":
		return f.colorMgr.StatusDown("Outage")
	default:
		return f.colorMgr.Dim(status)
	}
}

// formatUptimePercent formats uptime percentage for display.
func formatUptimePercent(percent float64) string {
	if percent == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f%%", percent)
}

// formatStatusResponseTime formats response time for status page display.
func formatStatusResponseTime(ms int, showResponseTime bool) string {
	if !showResponseTime || ms == 0 {
		return "-"
	}
	return fmt.Sprintf("%dms", ms)
}

// PrintAggregatedStatus prints the aggregated status of a status page.
// It displays the overall status header and a table of probe statuses.
func PrintAggregatedStatus(status client.AggregatedStatusResponse) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	// For JSON/YAML output, print the raw response
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(status)
	}

	formatter := NewAggregatedStatusTableFormatter(colorMode, isWide)

	// Print overall status header
	fmt.Printf("Overall Status: %s\n", formatter.formatOverallStatus(status.OverallStatus))
	fmt.Printf("Last Updated:   %s\n\n", status.LastUpdated.Format("2006-01-02 15:04:05 MST"))

	// Print probe statuses table
	if len(status.Probes) == 0 {
		fmt.Println("No probes configured on this status page.")
		return nil
	}

	rows := formatter.FormatProbeStatuses(status.Probes)
	return printer.Print(rows)
}

// DomainVerificationTableRow represents a row in the domain verification table output.
// The struct tags control column headers.
type DomainVerificationTableRow struct {
	Host  string `table:"HOST"`
	Value string `table:"VALUE"`
}

// PrintDomainVerification prints the DNS verification record for a custom domain.
// It displays the host (TXT record name) and value to configure in DNS.
func PrintDomainVerification(verification client.DomainVerificationResponse) error {
	printer := getPrinter()

	// For JSON/YAML output, print the raw response
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(verification)
	}

	// Print instructional header for table output
	fmt.Println("DNS Verification Record")
	fmt.Println("=======================")
	fmt.Println()
	fmt.Println("Add the following TXT record to your DNS configuration:")
	fmt.Println()

	row := DomainVerificationTableRow{
		Host:  verification.Host,
		Value: verification.Value,
	}

	if err := printer.Print(row); err != nil {
		return err
	}

	// Print additional instructions
	fmt.Println()
	fmt.Println("Instructions:")
	fmt.Println("1. Log in to your DNS provider's management console")
	fmt.Println("2. Create a new TXT record with the HOST and VALUE shown above")
	fmt.Println("3. Wait for DNS propagation (typically 5-60 minutes)")
	fmt.Println("4. StackEye will automatically verify the record once it propagates")

	return nil
}
