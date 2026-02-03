// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// SubscriptionTableRow represents a row for displaying billing subscription info.
// The struct tags control column headers and wide mode display.
type SubscriptionTableRow struct {
	Plan     string `table:"PLAN"`
	Status   string `table:"STATUS"`
	Monitors string `table:"MONITORS"`
	// Wide mode columns
	BillingEmail string `table:"BILLING EMAIL,wide"`
	NextBilling  string `table:"NEXT BILLING,wide"`
	Amount       string `table:"AMOUNT,wide"`
}

// UsageTableRow represents a row for displaying resource usage metrics.
// Multiple rows are used to show different resource types.
type UsageTableRow struct {
	Resource string `table:"RESOURCE"`
	Used     string `table:"USED"`
	Limit    string `table:"LIMIT"`
	Percent  string `table:"PERCENT"`
	// Wide mode columns
	Bar string `table:"USAGE BAR,wide"`
}

// InvoiceTableRow represents a row for displaying invoice information.
// The struct tags control column headers and wide mode display.
type InvoiceTableRow struct {
	Number string `table:"NUMBER"`
	Date   string `table:"DATE"`
	Status string `table:"STATUS"`
	Amount string `table:"AMOUNT"`
	// Wide mode columns
	PaidAt string `table:"PAID,wide"`
	Period string `table:"PERIOD,wide"`
}

// SubscriptionTableFormatter converts SDK BillingInfo to table-displayable rows.
type SubscriptionTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// UsageTableFormatter converts SDK UsageInfo to table-displayable rows.
type UsageTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// InvoiceTableFormatter converts SDK Invoice types to table-displayable rows.
type InvoiceTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewSubscriptionTableFormatter creates a new formatter for subscription table output.
// The colorMode parameter controls whether status colors are applied.
// Set isWide to true for extended output with additional columns.
func NewSubscriptionTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *SubscriptionTableFormatter {
	return &SubscriptionTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// NewUsageTableFormatter creates a new formatter for usage table output.
// The colorMode parameter controls whether usage bars are colored.
// Set isWide to true for extended output with visual usage bars.
func NewUsageTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *UsageTableFormatter {
	return &UsageTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// NewInvoiceTableFormatter creates a new formatter for invoice table output.
// The colorMode parameter controls whether status colors are applied.
// Set isWide to true for extended output with additional columns.
func NewInvoiceTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *InvoiceTableFormatter {
	return &InvoiceTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatSubscription converts SDK BillingInfo into a table-displayable row.
func (f *SubscriptionTableFormatter) FormatSubscription(info *client.BillingInfo) SubscriptionTableRow {
	if info == nil {
		return SubscriptionTableRow{
			Plan:     "-",
			Status:   "-",
			Monitors: "-",
		}
	}

	// Format plan name with title case
	plan := formatPlan(info.Plan)

	// Format status with color
	status := f.formatSubscriptionStatus(info.Status)

	// Format monitor usage
	monitors := fmt.Sprintf("%d / %d", info.MonitorCount, info.MonitorLimit)

	row := SubscriptionTableRow{
		Plan:     plan,
		Status:   status,
		Monitors: monitors,
	}

	// Wide mode fields
	if info.BillingEmail != "" {
		row.BillingEmail = truncateField(info.BillingEmail, 30)
	} else {
		row.BillingEmail = "-"
	}

	if info.NextBillingAt != nil && *info.NextBillingAt != "" {
		row.NextBilling = formatBillingDate(*info.NextBillingAt)
	} else {
		row.NextBilling = "-"
	}

	if info.AmountCents != nil && *info.AmountCents > 0 {
		row.Amount = formatBillingCurrency(*info.AmountCents, info.Currency)
	} else {
		row.Amount = "-"
	}

	return row
}

// formatSubscriptionStatus applies color based on subscription status.
func (f *SubscriptionTableFormatter) formatSubscriptionStatus(status string) string {
	lower := strings.ToLower(status)
	switch lower {
	case "active":
		return f.colorMgr.StatusUp("ACTIVE")
	case "trialing":
		return f.colorMgr.StatusWarning("TRIALING")
	case "past_due":
		return f.colorMgr.StatusDown("PAST DUE")
	case "canceled", "cancelled":
		return f.colorMgr.StatusInfo("CANCELED")
	case "incomplete":
		return f.colorMgr.StatusWarning("INCOMPLETE")
	case "incomplete_expired":
		return f.colorMgr.StatusDown("EXPIRED")
	case "unpaid":
		return f.colorMgr.StatusDown("UNPAID")
	case "paused":
		return f.colorMgr.StatusInfo("PAUSED")
	default:
		if status == "" {
			return "NONE"
		}
		return strings.ToUpper(status)
	}
}

// FormatUsage converts SDK UsageInfo into table-displayable rows.
// Returns multiple rows, one for each resource type (monitors, team members, checks).
func (f *UsageTableFormatter) FormatUsage(usage *client.UsageInfo) []UsageTableRow {
	if usage == nil {
		return []UsageTableRow{}
	}

	rows := make([]UsageTableRow, 0, 3)

	// Monitors row
	monitorsPercent := calculatePercent(usage.MonitorsCount, usage.MonitorsLimit)
	rows = append(rows, UsageTableRow{
		Resource: "Monitors",
		Used:     fmt.Sprintf("%d", usage.MonitorsCount),
		Limit:    fmt.Sprintf("%d", usage.MonitorsLimit),
		Percent:  f.formatPercent(monitorsPercent),
		Bar:      f.formatBar(monitorsPercent),
	})

	// Team Members row
	teamPercent := calculatePercent(usage.TeamMembersCount, usage.TeamMembersLimit)
	rows = append(rows, UsageTableRow{
		Resource: "Team Members",
		Used:     fmt.Sprintf("%d", usage.TeamMembersCount),
		Limit:    fmt.Sprintf("%d", usage.TeamMembersLimit),
		Percent:  f.formatPercent(teamPercent),
		Bar:      f.formatBar(teamPercent),
	})

	// Probe Checks row (no limit for checks, just count)
	rows = append(rows, UsageTableRow{
		Resource: "Probe Checks",
		Used:     formatLargeNum(usage.ChecksCount),
		Limit:    "-",
		Percent:  "-",
		Bar:      "-",
	})

	return rows
}

// formatPercent formats a percentage with color based on usage level.
func (f *UsageTableFormatter) formatPercent(percent float64) string {
	str := fmt.Sprintf("%.1f%%", percent)
	if percent >= 90 {
		return f.colorMgr.StatusDown(str)
	} else if percent >= 75 {
		return f.colorMgr.StatusWarning(str)
	}
	return str
}

// formatBar creates a visual progress bar for usage.
func (f *UsageTableFormatter) formatBar(percent float64) string {
	width := 20
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}

	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	fillChar := "█"
	emptyChar := "░"

	// Color the fill based on usage level
	bar := strings.Repeat(fillChar, filled) + strings.Repeat(emptyChar, width-filled)
	if percent >= 90 {
		return f.colorMgr.StatusDown("[" + bar + "]")
	} else if percent >= 75 {
		return f.colorMgr.StatusWarning("[" + bar + "]")
	}
	return "[" + bar + "]"
}

// FormatInvoices converts a slice of SDK Invoices into table-displayable rows.
func (f *InvoiceTableFormatter) FormatInvoices(invoices []client.Invoice) []InvoiceTableRow {
	rows := make([]InvoiceTableRow, 0, len(invoices))
	for _, inv := range invoices {
		rows = append(rows, f.formatInvoice(inv))
	}
	return rows
}

// FormatInvoice converts a single SDK Invoice into a table-displayable row.
func (f *InvoiceTableFormatter) FormatInvoice(invoice client.Invoice) InvoiceTableRow {
	return f.formatInvoice(invoice)
}

// formatInvoice is the internal conversion function.
func (f *InvoiceTableFormatter) formatInvoice(inv client.Invoice) InvoiceTableRow {
	row := InvoiceTableRow{
		Number: truncateField(inv.InvoiceNumber, 20),
		Date:   formatBillingDate(inv.CreatedAt),
		Status: f.formatInvoiceStatus(inv.Status),
		Amount: formatBillingCurrency(int(inv.Total), inv.Currency),
	}

	// Wide mode fields
	if inv.PaidAt != nil && *inv.PaidAt != "" {
		row.PaidAt = formatBillingDate(*inv.PaidAt)
	} else {
		row.PaidAt = "-"
	}

	if inv.PeriodStart != nil && inv.PeriodEnd != nil && *inv.PeriodStart != "" && *inv.PeriodEnd != "" {
		start := formatBillingDate(*inv.PeriodStart)
		end := formatBillingDate(*inv.PeriodEnd)
		row.Period = fmt.Sprintf("%s - %s", start, end)
	} else {
		row.Period = "-"
	}

	return row
}

// formatInvoiceStatus applies color based on invoice status.
func (f *InvoiceTableFormatter) formatInvoiceStatus(status string) string {
	lower := strings.ToLower(status)
	switch lower {
	case "paid":
		return f.colorMgr.StatusUp("PAID")
	case "open":
		return f.colorMgr.StatusWarning("OPEN")
	case "draft":
		return f.colorMgr.StatusInfo("DRAFT")
	case "void":
		return f.colorMgr.StatusInfo("VOID")
	case "uncollectible":
		return f.colorMgr.StatusDown("UNCOLLECTIBLE")
	default:
		if status == "" {
			return "UNKNOWN"
		}
		return strings.ToUpper(status)
	}
}

// --- Helper functions ---

// formatPlan formats a plan name for display with proper capitalization.
func formatPlan(plan string) string {
	if plan == "" {
		return "None"
	}
	// Title case the plan name
	return strings.ToUpper(plan[:1]) + strings.ToLower(plan[1:])
}

// truncateField truncates a string to fit in the table display.
func truncateField(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatBillingDate parses an ISO date and returns a formatted date string.
func formatBillingDate(dateStr string) string {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format("Jan 2, 2006")
		}
	}

	// Return truncated original if parsing fails
	return truncateField(dateStr, 12)
}

// formatBillingCurrency formats cents to a currency string.
func formatBillingCurrency(cents int, currency string) string {
	if currency == "" {
		currency = "USD"
	}
	currency = strings.ToUpper(currency)

	dollars := float64(cents) / 100.0

	// Use appropriate symbol
	symbol := "$"
	switch currency {
	case "EUR":
		symbol = "€"
	case "GBP":
		symbol = "£"
	case "JPY":
		symbol = "¥"
		dollars = float64(cents) // JPY doesn't use cents
	}

	return fmt.Sprintf("%s%.2f", symbol, dollars)
}

// calculatePercent calculates usage percentage safely.
func calculatePercent(used, limit int) float64 {
	if limit <= 0 {
		return 0
	}
	return float64(used) / float64(limit) * 100
}

// formatLargeNum formats a large number with thousand separators.
func formatLargeNum(n int64) string {
	if n < 0 {
		return "-" + formatLargeNum(-n)
	}
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	str := fmt.Sprintf("%d", n)
	var result strings.Builder

	length := len(str)
	for i, ch := range str {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(ch)
	}

	return result.String()
}

// --- Convenience print functions ---

// PrintSubscription is a convenience function that formats and prints billing info
// using the CLI's configured output format.
func PrintSubscription(info *client.BillingInfo) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewSubscriptionTableFormatter(colorMode, isWide)
	row := formatter.FormatSubscription(info)

	return printer.Print(row)
}

// PrintUsage is a convenience function that formats and prints usage info
// using the CLI's configured output format.
func PrintUsage(usage *client.UsageInfo) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewUsageTableFormatter(colorMode, isWide)
	rows := formatter.FormatUsage(usage)

	return printer.Print(rows)
}

// PrintInvoices is a convenience function that formats and prints invoices
// using the CLI's configured output format.
func PrintInvoices(invoices []client.Invoice) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewInvoiceTableFormatter(colorMode, isWide)
	rows := formatter.FormatInvoices(invoices)

	return printer.Print(rows)
}

// PrintInvoice is a convenience function that formats and prints a single invoice.
func PrintInvoice(invoice client.Invoice) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewInvoiceTableFormatter(colorMode, isWide)
	row := formatter.FormatInvoice(invoice)

	return printer.Print(row)
}
