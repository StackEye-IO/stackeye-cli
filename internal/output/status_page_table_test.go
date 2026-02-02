package output

import (
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStatusPageTableFormatter_FormatStatusPages(t *testing.T) {
	now := time.Now()

	customDomain := "status.example.com"
	pages := []client.StatusPage{
		{
			ID:                   1,
			Name:                 "Acme Status",
			Slug:                 "acme-status",
			Theme:                "dark",
			IsPublic:             true,
			Enabled:              true,
			ShowUptimePercentage: true,
			CustomDomain:         &customDomain,
			Probes:               []client.StatusPageProbe{{}, {}}, // 2 probes
			CreatedAt:            now,
		},
		{
			ID:                   2,
			Name:                 "Internal Dashboard",
			Slug:                 "internal-dash",
			Theme:                "light",
			IsPublic:             false,
			Enabled:              false,
			ShowUptimePercentage: false,
			Probes:               []client.StatusPageProbe{}, // No probes
			CreatedAt:            now,
		},
	}

	// Test with colors disabled for predictable output
	formatter := NewStatusPageTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatStatusPages(pages)

	assert.Len(t, rows, 2)

	// First page - public, enabled
	assert.Equal(t, "Acme Status", rows[0].Name)
	assert.Equal(t, "acme-status", rows[0].Slug)
	assert.Equal(t, "Dark", rows[0].Theme)
	assert.Equal(t, "Yes", rows[0].Public)
	assert.Equal(t, "Yes", rows[0].Enabled)
	assert.Equal(t, "2", rows[0].Probes)
	assert.Equal(t, "status.example.com", rows[0].Domain)
	assert.Equal(t, "Yes", rows[0].Uptime)
	assert.Equal(t, "1", rows[0].ID)

	// Second page - private, disabled
	assert.Equal(t, "Internal Dashboard", rows[1].Name)
	assert.Equal(t, "internal-dash", rows[1].Slug)
	assert.Equal(t, "Light", rows[1].Theme)
	assert.Equal(t, "No", rows[1].Public)
	assert.Equal(t, "No", rows[1].Enabled)
	assert.Equal(t, "0", rows[1].Probes)
	assert.Equal(t, "-", rows[1].Domain)
	assert.Equal(t, "No", rows[1].Uptime)
}

func TestStatusPageTableFormatter_FormatStatusPage(t *testing.T) {
	formatter := NewStatusPageTableFormatter(sdkoutput.ColorNever, false)

	page := client.StatusPage{
		ID:       42,
		Name:     "Test Page",
		Slug:     "test",
		Theme:    "system",
		IsPublic: true,
		Enabled:  true,
		Probes:   []client.StatusPageProbe{{}},
	}

	row := formatter.FormatStatusPage(page)

	assert.Equal(t, "Test Page", row.Name)
	assert.Equal(t, "test", row.Slug)
	assert.Equal(t, "System", row.Theme)
	assert.Equal(t, "1", row.Probes)
	assert.Equal(t, "42", row.ID)
}

func TestTruncateName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short name", "API Status", 30, "API Status"},
		{"exact length", "12345", 5, "12345"},
		{"needs truncation", "This is a very long status page name", 20, "This is a very lo..."},
		{"empty string", "", 10, ""},
		{"unicode name", "日本語のステータスページ", 10, "日本語のステー..."},
		{"unicode short", "日本語", 10, "日本語"},
		{"emoji name", "🚀 Status Page", 10, "🚀 Statu..."},
		{"small maxLen normalizes to 4", "Long name here", 2, "L..."},
		{"exact min maxLen", "Long name here", 4, "L..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateName(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTheme(t *testing.T) {
	tests := []struct {
		theme    string
		expected string
	}{
		{"light", "Light"},
		{"dark", "Dark"},
		{"system", "System"},
		{"", "System"},
		{"custom", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.theme, func(t *testing.T) {
			result := formatTheme(tt.theme)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatCustomDomain(t *testing.T) {
	short := "status.acme.com"
	long := "status.really-long-domain-name.example.com"
	empty := ""

	tests := []struct {
		name     string
		domain   *string
		expected string
	}{
		{"nil domain", nil, "-"},
		{"empty domain", &empty, "-"},
		{"short domain", &short, "status.acme.com"},
		{"long domain", &long, "status.really-long-dom..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCustomDomain(tt.domain)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatUptimeDisplay(t *testing.T) {
	tests := []struct {
		showUptime bool
		expected   string
	}{
		{true, "Yes"},
		{false, "No"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatUptimeDisplay(tt.showUptime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusPageTableFormatter_PublicColoring(t *testing.T) {
	formatter := NewStatusPageTableFormatter(sdkoutput.ColorAlways, false)

	// Test public page - should have success color
	publicPage := client.StatusPage{
		ID:       1,
		IsPublic: true,
		Enabled:  true,
	}

	row := formatter.FormatStatusPage(publicPage)

	// When colors are enabled, the public field should contain ANSI escape codes
	assert.Contains(t, row.Public, "Yes")
	assert.Contains(t, row.Public, "\x1b[") // ANSI escape code

	// Test private page - should be dim
	privatePage := client.StatusPage{
		ID:       2,
		IsPublic: false,
		Enabled:  true,
	}

	row2 := formatter.FormatStatusPage(privatePage)
	assert.Contains(t, row2.Public, "No")
}

func TestStatusPageTableFormatter_EnabledColoring(t *testing.T) {
	formatter := NewStatusPageTableFormatter(sdkoutput.ColorAlways, false)

	// Test enabled page - should have success color
	enabledPage := client.StatusPage{
		ID:      1,
		Enabled: true,
	}

	row := formatter.FormatStatusPage(enabledPage)
	assert.Contains(t, row.Enabled, "Yes")
	assert.Contains(t, row.Enabled, "\x1b[")

	// Test disabled page - should have warning color
	disabledPage := client.StatusPage{
		ID:      2,
		Enabled: false,
	}

	row2 := formatter.FormatStatusPage(disabledPage)
	assert.Contains(t, row2.Enabled, "No")
	assert.Contains(t, row2.Enabled, "\x1b[")
}

func TestStatusPageTableFormatter_NoColor(t *testing.T) {
	formatter := NewStatusPageTableFormatter(sdkoutput.ColorNever, false)

	page := client.StatusPage{
		ID:       1,
		IsPublic: true,
		Enabled:  false,
	}

	row := formatter.FormatStatusPage(page)

	// When colors are disabled, should be plain text without ANSI codes
	assert.Equal(t, "Yes", row.Public)
	assert.Equal(t, "No", row.Enabled)
	assert.NotContains(t, row.Public, "\x1b[")
	assert.NotContains(t, row.Enabled, "\x1b[")
}

func TestNewStatusPageTableFormatter(t *testing.T) {
	// Test constructor creates valid formatter
	formatter := NewStatusPageTableFormatter(sdkoutput.ColorAuto, true)

	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.colorMgr)
	assert.True(t, formatter.isWide)

	formatterNoWide := NewStatusPageTableFormatter(sdkoutput.ColorNever, false)
	assert.False(t, formatterNoWide.isWide)
}

func TestFormatStatusPageCount(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		page     int
		limit    int
		expected string
	}{
		{"zero total", 0, 1, 20, ""},
		{"first page", 45, 1, 20, "Showing 1-20 of 45 status pages"},
		{"second page", 45, 2, 20, "Showing 21-40 of 45 status pages"},
		{"last page partial", 45, 3, 20, "Showing 41-45 of 45 status pages"},
		{"single page", 5, 1, 20, "Showing 1-5 of 5 status pages"},
		{"zero page", 100, 0, 20, "Showing 1-20 of 100 status pages"},
		{"negative page", 100, -1, 20, "Showing 1-20 of 100 status pages"},
		{"zero limit", 100, 1, 0, "Showing 1-20 of 100 status pages"},
		{"negative limit", 100, 1, -5, "Showing 1-20 of 100 status pages"},
		{"page beyond total", 10, 100, 20, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatStatusPageCount(tt.total, tt.page, tt.limit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for aggregated status functionality

func TestAggregatedStatusTableFormatter_FormatProbeStatuses(t *testing.T) {
	probeID1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	probeID2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

	probes := []client.ProbeStatusSummary{
		{
			ProbeID:          probeID1,
			DisplayName:      "API Server",
			Status:           "up",
			UptimePercent:    99.95,
			ResponseTimeMs:   150,
			ShowResponseTime: true,
		},
		{
			ProbeID:          probeID2,
			DisplayName:      "Database",
			Status:           "down",
			UptimePercent:    85.50,
			ResponseTimeMs:   0,
			ShowResponseTime: false,
		},
	}

	formatter := NewAggregatedStatusTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatProbeStatuses(probes)

	assert.Len(t, rows, 2)

	// First probe - up with response time shown
	assert.Equal(t, "API Server", rows[0].Name)
	assert.Equal(t, "Up", rows[0].Status)
	assert.Equal(t, "99.95%", rows[0].Uptime)
	assert.Equal(t, "150ms", rows[0].ResponseTime)
	assert.Equal(t, probeID1.String(), rows[0].ProbeID)

	// Second probe - down, no response time shown
	assert.Equal(t, "Database", rows[1].Name)
	assert.Equal(t, "Down", rows[1].Status)
	assert.Equal(t, "85.50%", rows[1].Uptime)
	assert.Equal(t, "-", rows[1].ResponseTime)
	assert.Equal(t, probeID2.String(), rows[1].ProbeID)
}

func TestAggregatedStatusTableFormatter_FormatProbeStatusValue(t *testing.T) {
	formatter := NewAggregatedStatusTableFormatter(sdkoutput.ColorNever, false)

	tests := []struct {
		status   string
		expected string
	}{
		{"up", "Up"},
		{"down", "Down"},
		{"degraded", "Degraded"},
		{"pending", "Pending"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := formatter.formatProbeStatusValue(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAggregatedStatusTableFormatter_FormatOverallStatus(t *testing.T) {
	formatter := NewAggregatedStatusTableFormatter(sdkoutput.ColorNever, false)

	tests := []struct {
		status   string
		expected string
	}{
		{"operational", "Operational"},
		{"degraded", "Degraded"},
		{"outage", "Outage"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := formatter.formatOverallStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatUptimePercent(t *testing.T) {
	tests := []struct {
		name     string
		percent  float64
		expected string
	}{
		{"zero", 0, "-"},
		{"100 percent", 100.0, "100.00%"},
		{"99.99 percent", 99.99, "99.99%"},
		{"50.5 percent", 50.5, "50.50%"},
		{"low percent", 0.01, "0.01%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUptimePercent(tt.percent)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatStatusResponseTime(t *testing.T) {
	tests := []struct {
		name             string
		ms               int
		showResponseTime bool
		expected         string
	}{
		{"show disabled", 100, false, "-"},
		{"zero ms", 0, true, "-"},
		{"normal ms", 150, true, "150ms"},
		{"high ms", 5000, true, "5000ms"},
		{"both zero and disabled", 0, false, "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatusResponseTime(tt.ms, tt.showResponseTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAggregatedStatusTableFormatter_WithColors(t *testing.T) {
	formatter := NewAggregatedStatusTableFormatter(sdkoutput.ColorAlways, false)

	probeID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	probe := client.ProbeStatusSummary{
		ProbeID:          probeID,
		DisplayName:      "Test Probe",
		Status:           "up",
		UptimePercent:    99.0,
		ResponseTimeMs:   100,
		ShowResponseTime: true,
	}

	rows := formatter.FormatProbeStatuses([]client.ProbeStatusSummary{probe})

	assert.Len(t, rows, 1)
	// When colors are enabled, status should contain ANSI escape codes
	assert.Contains(t, rows[0].Status, "Up")
	assert.Contains(t, rows[0].Status, "\x1b[") // ANSI escape code
}

func TestNewAggregatedStatusTableFormatter(t *testing.T) {
	formatter := NewAggregatedStatusTableFormatter(sdkoutput.ColorAuto, true)

	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.colorMgr)
	assert.True(t, formatter.isWide)

	formatterNoWide := NewAggregatedStatusTableFormatter(sdkoutput.ColorNever, false)
	assert.False(t, formatterNoWide.isWide)
}

// Tests for domain verification functionality

func TestDomainVerificationTableRow_Fields(t *testing.T) {
	// Test that the struct has expected fields and table tags
	row := DomainVerificationTableRow{
		Host:  "_stackeye-verify.status.example.com",
		Value: "stackeye-verify=abc123def456",
	}

	assert.Equal(t, "_stackeye-verify.status.example.com", row.Host)
	assert.Equal(t, "stackeye-verify=abc123def456", row.Value)
}

func TestDomainVerificationTableRow_EmptyValues(t *testing.T) {
	// Test with empty values - should work without panic
	row := DomainVerificationTableRow{
		Host:  "",
		Value: "",
	}

	assert.Empty(t, row.Host)
	assert.Empty(t, row.Value)
}

func TestDomainVerificationTableRow_LongValues(t *testing.T) {
	// Test with long domain verification values
	longHost := "_stackeye-verify.very-long-subdomain.status.example.com"
	longValue := "stackeye-verify=550e8400-e29b-41d4-a716-446655440000-longer-value"

	row := DomainVerificationTableRow{
		Host:  longHost,
		Value: longValue,
	}

	// Values should not be truncated for verification records - users need the full values
	assert.Equal(t, longHost, row.Host)
	assert.Equal(t, longValue, row.Value)
}
