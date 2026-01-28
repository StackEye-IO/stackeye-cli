package output

import (
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionTableFormatter_FormatSubscription(t *testing.T) {
	billingEmail := "billing@acme.com"
	nextBilling := "2026-02-01"
	amountCents := 2900

	info := &client.BillingInfo{
		Plan:          "pro",
		Status:        "active",
		BillingEmail:  billingEmail,
		MonitorCount:  15,
		MonitorLimit:  100,
		NextBillingAt: &nextBilling,
		AmountCents:   &amountCents,
		Currency:      "USD",
	}

	formatter := NewSubscriptionTableFormatter(sdkoutput.ColorNever, false)
	row := formatter.FormatSubscription(info)

	assert.Equal(t, "Pro", row.Plan)
	assert.Equal(t, "ACTIVE", row.Status)
	assert.Equal(t, "15 / 100", row.Monitors)
	assert.Equal(t, "billing@acme.com", row.BillingEmail)
	assert.Equal(t, "Feb 1, 2026", row.NextBilling)
	assert.Equal(t, "$29.00", row.Amount)
}

func TestSubscriptionTableFormatter_NilInfo(t *testing.T) {
	formatter := NewSubscriptionTableFormatter(sdkoutput.ColorNever, false)
	row := formatter.FormatSubscription(nil)

	assert.Equal(t, "-", row.Plan)
	assert.Equal(t, "-", row.Status)
	assert.Equal(t, "-", row.Monitors)
}

func TestSubscriptionTableFormatter_StatusColors(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string // Without color codes
	}{
		{"active", "active", "ACTIVE"},
		{"trialing", "trialing", "TRIALING"},
		{"past_due", "past_due", "PAST DUE"},
		{"canceled", "canceled", "CANCELED"},
		{"cancelled", "cancelled", "CANCELED"},
		{"incomplete", "incomplete", "INCOMPLETE"},
		{"incomplete_expired", "incomplete_expired", "EXPIRED"},
		{"unpaid", "unpaid", "UNPAID"},
		{"paused", "paused", "PAUSED"},
		{"empty", "", "NONE"},
		{"unknown", "custom_status", "CUSTOM_STATUS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewSubscriptionTableFormatter(sdkoutput.ColorNever, false)
			info := &client.BillingInfo{
				Status:       tt.status,
				Plan:         "starter",
				MonitorCount: 5,
				MonitorLimit: 25,
			}
			row := formatter.FormatSubscription(info)
			assert.Equal(t, tt.expected, row.Status)
		})
	}
}

func TestSubscriptionTableFormatter_StatusColorsEnabled(t *testing.T) {
	formatter := NewSubscriptionTableFormatter(sdkoutput.ColorAlways, false)
	info := &client.BillingInfo{
		Status:       "active",
		Plan:         "pro",
		MonitorCount: 10,
		MonitorLimit: 100,
	}
	row := formatter.FormatSubscription(info)

	// Should contain ANSI escape codes when colors enabled
	assert.Contains(t, row.Status, "\x1b[")
}

func TestSubscriptionTableFormatter_MissingOptionalFields(t *testing.T) {
	info := &client.BillingInfo{
		Plan:         "free",
		Status:       "active",
		MonitorCount: 5,
		MonitorLimit: 10,
		// No BillingEmail, NextBillingAt, AmountCents
	}

	formatter := NewSubscriptionTableFormatter(sdkoutput.ColorNever, false)
	row := formatter.FormatSubscription(info)

	assert.Equal(t, "-", row.BillingEmail)
	assert.Equal(t, "-", row.NextBilling)
	assert.Equal(t, "-", row.Amount)
}

func TestUsageTableFormatter_FormatUsage(t *testing.T) {
	usage := &client.UsageInfo{
		MonitorsCount:    15,
		MonitorsLimit:    100,
		TeamMembersCount: 3,
		TeamMembersLimit: 10,
		ChecksCount:      1234567,
		PeriodStart:      "2026-01-01",
		PeriodEnd:        "2026-01-31",
	}

	formatter := NewUsageTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatUsage(usage)

	assert.Len(t, rows, 3)

	// Monitors row
	assert.Equal(t, "Monitors", rows[0].Resource)
	assert.Equal(t, "15", rows[0].Used)
	assert.Equal(t, "100", rows[0].Limit)
	assert.Equal(t, "15.0%", rows[0].Percent)

	// Team Members row
	assert.Equal(t, "Team Members", rows[1].Resource)
	assert.Equal(t, "3", rows[1].Used)
	assert.Equal(t, "10", rows[1].Limit)
	assert.Equal(t, "30.0%", rows[1].Percent)

	// Probe Checks row
	assert.Equal(t, "Probe Checks", rows[2].Resource)
	assert.Equal(t, "1,234,567", rows[2].Used)
	assert.Equal(t, "-", rows[2].Limit)
	assert.Equal(t, "-", rows[2].Percent)
}

func TestUsageTableFormatter_NilUsage(t *testing.T) {
	formatter := NewUsageTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatUsage(nil)

	assert.Len(t, rows, 0)
	assert.NotNil(t, rows) // Should return empty slice, not nil
}

func TestUsageTableFormatter_HighUsageWarning(t *testing.T) {
	usage := &client.UsageInfo{
		MonitorsCount:    95,
		MonitorsLimit:    100,
		TeamMembersCount: 8,
		TeamMembersLimit: 10,
		ChecksCount:      1000,
	}

	// With colors disabled, check the percentage values
	formatter := NewUsageTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatUsage(usage)

	// 95% usage
	assert.Equal(t, "95.0%", rows[0].Percent)

	// 80% usage
	assert.Equal(t, "80.0%", rows[1].Percent)
}

func TestUsageTableFormatter_HighUsageColoring(t *testing.T) {
	usage := &client.UsageInfo{
		MonitorsCount:    95,
		MonitorsLimit:    100,
		TeamMembersCount: 1,
		TeamMembersLimit: 10,
		ChecksCount:      100,
	}

	// With colors enabled, high usage should have color codes
	formatter := NewUsageTableFormatter(sdkoutput.ColorAlways, true)
	rows := formatter.FormatUsage(usage)

	// 95% should be red (StatusDown)
	assert.Contains(t, rows[0].Percent, "\x1b[")
	assert.Contains(t, rows[0].Bar, "\x1b[")

	// 10% should not have color codes in percent (below threshold)
	assert.NotContains(t, rows[1].Percent, "\x1b[")
}

func TestUsageTableFormatter_BarFormat(t *testing.T) {
	formatter := NewUsageTableFormatter(sdkoutput.ColorNever, true)

	tests := []struct {
		name    string
		percent float64
	}{
		{"0%", 0},
		{"50%", 50},
		{"100%", 100},
		{"overflow", 150}, // Should cap at 100%
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := formatter.formatBar(tt.percent)
			// Bar should always have brackets
			assert.Contains(t, bar, "[")
			assert.Contains(t, bar, "]")
			// Bar should have consistent visual width (20 chars + brackets)
			// Note: █ and ░ are multi-byte Unicode chars, so we check rune count
			assert.Equal(t, 22, len([]rune(bar))) // "[" + 20 runes + "]"
		})
	}
}

func TestInvoiceTableFormatter_FormatInvoices(t *testing.T) {
	paidAt := "2026-01-15T10:30:00Z"
	periodStart := "2026-01-01"
	periodEnd := "2026-01-31"

	invoices := []client.Invoice{
		{
			ID:            1,
			InvoiceNumber: "INV-2026-001",
			Status:        "paid",
			Total:         2900,
			Currency:      "USD",
			CreatedAt:     "2026-01-01T00:00:00Z",
			PaidAt:        &paidAt,
			PeriodStart:   &periodStart,
			PeriodEnd:     &periodEnd,
		},
		{
			ID:            2,
			InvoiceNumber: "INV-2026-002",
			Status:        "open",
			Total:         2900,
			Currency:      "USD",
			CreatedAt:     "2026-02-01T00:00:00Z",
		},
	}

	formatter := NewInvoiceTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatInvoices(invoices)

	assert.Len(t, rows, 2)

	// First invoice - paid
	assert.Equal(t, "INV-2026-001", rows[0].Number)
	assert.Equal(t, "Jan 1, 2026", rows[0].Date)
	assert.Equal(t, "PAID", rows[0].Status)
	assert.Equal(t, "$29.00", rows[0].Amount)
	assert.Equal(t, "Jan 15, 2026", rows[0].PaidAt)
	assert.Equal(t, "Jan 1, 2026 - Jan 31, 2026", rows[0].Period)

	// Second invoice - open
	assert.Equal(t, "INV-2026-002", rows[1].Number)
	assert.Equal(t, "Feb 1, 2026", rows[1].Date)
	assert.Equal(t, "OPEN", rows[1].Status)
	assert.Equal(t, "$29.00", rows[1].Amount)
	assert.Equal(t, "-", rows[1].PaidAt)
	assert.Equal(t, "-", rows[1].Period)
}

func TestInvoiceTableFormatter_FormatSingleInvoice(t *testing.T) {
	invoice := client.Invoice{
		ID:            1,
		InvoiceNumber: "INV-2026-001",
		Status:        "paid",
		Total:         5000,
		Currency:      "EUR",
		CreatedAt:     "2026-01-15",
	}

	formatter := NewInvoiceTableFormatter(sdkoutput.ColorNever, false)
	row := formatter.FormatInvoice(invoice)

	assert.Equal(t, "INV-2026-001", row.Number)
	assert.Equal(t, "€50.00", row.Amount)
}

func TestInvoiceTableFormatter_StatusColors(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"paid", "paid", "PAID"},
		{"open", "open", "OPEN"},
		{"draft", "draft", "DRAFT"},
		{"void", "void", "VOID"},
		{"uncollectible", "uncollectible", "UNCOLLECTIBLE"},
		{"empty", "", "UNKNOWN"},
		{"custom", "custom_status", "CUSTOM_STATUS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewInvoiceTableFormatter(sdkoutput.ColorNever, false)
			invoice := client.Invoice{
				InvoiceNumber: "TEST-001",
				Status:        tt.status,
				Total:         1000,
				Currency:      "USD",
				CreatedAt:     "2026-01-01",
			}
			row := formatter.FormatInvoice(invoice)
			assert.Equal(t, tt.expected, row.Status)
		})
	}
}

func TestInvoiceTableFormatter_EmptySlice(t *testing.T) {
	formatter := NewInvoiceTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatInvoices([]client.Invoice{})

	assert.Len(t, rows, 0)
	assert.NotNil(t, rows)
}

func TestInvoiceTableFormatter_StatusColorsEnabled(t *testing.T) {
	formatter := NewInvoiceTableFormatter(sdkoutput.ColorAlways, false)
	invoice := client.Invoice{
		InvoiceNumber: "TEST-001",
		Status:        "paid",
		Total:         1000,
		Currency:      "USD",
		CreatedAt:     "2026-01-01",
	}
	row := formatter.FormatInvoice(invoice)

	// Should contain ANSI escape codes when colors enabled
	assert.Contains(t, row.Status, "\x1b[")
}

// --- Helper function tests ---

func TestFormatPlan(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"free", "Free"},
		{"STARTER", "Starter"},
		{"pro", "Pro"},
		{"PRO", "Pro"},
		{"team", "Team"},
		{"enterprise", "Enterprise"},
		{"", "None"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatPlan(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"very short max", "hello", 3, "hel"},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateField(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBillingDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"RFC3339", "2026-01-15T10:30:00Z", "Jan 15, 2026"},
		{"RFC3339 with offset", "2026-01-15T10:30:00+05:00", "Jan 15, 2026"},
		{"date only", "2026-01-15", "Jan 15, 2026"},
		{"invalid date", "not-a-date", "not-a-date"},
		{"long invalid", "this is a very long invalid date string", "this is a..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBillingDate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBillingCurrency(t *testing.T) {
	tests := []struct {
		name     string
		cents    int
		currency string
		expected string
	}{
		{"USD", 2900, "USD", "$29.00"},
		{"USD lowercase", 2900, "usd", "$29.00"},
		{"EUR", 4999, "EUR", "€49.99"},
		{"GBP", 1000, "GBP", "£10.00"},
		{"JPY", 1000, "JPY", "¥1000.00"},
		{"empty currency defaults to USD", 5000, "", "$50.00"},
		{"zero amount", 0, "USD", "$0.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBillingCurrency(tt.cents, tt.currency)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculatePercent(t *testing.T) {
	tests := []struct {
		name     string
		used     int
		limit    int
		expected float64
	}{
		{"normal", 50, 100, 50.0},
		{"full", 100, 100, 100.0},
		{"empty", 0, 100, 0.0},
		{"zero limit", 10, 0, 0.0},
		{"negative limit", 10, -5, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePercent(tt.used, tt.limit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatLargeNum(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"small", 123, "123"},
		{"thousands", 1234, "1,234"},
		{"millions", 1234567, "1,234,567"},
		{"zero", 0, "0"},
		{"negative", -1234, "-1,234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatLargeNum(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewFormatterConstructors(t *testing.T) {
	// Test all constructors create valid formatters
	sub := NewSubscriptionTableFormatter(sdkoutput.ColorAuto, true)
	assert.NotNil(t, sub)
	assert.NotNil(t, sub.colorMgr)
	assert.True(t, sub.isWide)

	usage := NewUsageTableFormatter(sdkoutput.ColorNever, false)
	assert.NotNil(t, usage)
	assert.NotNil(t, usage.colorMgr)
	assert.False(t, usage.isWide)

	inv := NewInvoiceTableFormatter(sdkoutput.ColorAlways, true)
	assert.NotNil(t, inv)
	assert.NotNil(t, inv.colorMgr)
	assert.True(t, inv.isWide)
}

func TestAllColorModes(t *testing.T) {
	colorModes := []sdkoutput.ColorMode{
		sdkoutput.ColorAuto,
		sdkoutput.ColorAlways,
		sdkoutput.ColorNever,
	}

	info := &client.BillingInfo{
		Plan:         "pro",
		Status:       "active",
		MonitorCount: 10,
		MonitorLimit: 100,
	}

	for _, mode := range colorModes {
		t.Run(string(mode), func(t *testing.T) {
			formatter := NewSubscriptionTableFormatter(mode, false)
			row := formatter.FormatSubscription(info)
			assert.NotEmpty(t, row.Plan)
			assert.NotEmpty(t, row.Status)
		})
	}
}
