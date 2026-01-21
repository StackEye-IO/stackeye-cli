package output

import (
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAlertTableFormatter_FormatAlerts(t *testing.T) {
	now := time.Now()
	fiveMinAgo := now.Add(-5 * time.Minute)
	twoHoursAgo := now.Add(-2 * time.Hour)
	oneHourAgo := now.Add(-1 * time.Hour)

	probeInfo := &client.AlertProbe{
		ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:      "API Health",
		URL:       "https://api.example.com/health",
		CheckType: client.CheckTypeHTTP,
	}

	ackBy := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	durationSecs := 3600

	alerts := []client.Alert{
		{
			ID:          uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			Status:      client.AlertStatusActive,
			Severity:    client.AlertSeverityCritical,
			AlertType:   client.AlertTypeStatusDown,
			TriggeredAt: fiveMinAgo,
			Probe:       probeInfo,
		},
		{
			ID:             uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			Status:         client.AlertStatusAcknowledged,
			Severity:       client.AlertSeverityWarning,
			AlertType:      client.AlertTypeSSLExpiry,
			TriggeredAt:    twoHoursAgo,
			AcknowledgedBy: &ackBy,
			Probe:          probeInfo,
		},
		{
			ID:              uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
			Status:          client.AlertStatusResolved,
			Severity:        client.AlertSeverityInfo,
			AlertType:       client.AlertTypeSlowResponse,
			TriggeredAt:     twoHoursAgo,
			ResolvedAt:      &oneHourAgo,
			DurationSeconds: &durationSecs,
			Probe:           probeInfo,
		},
	}

	// Test with colors disabled for predictable output
	formatter := NewAlertTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatAlerts(alerts)

	assert.Len(t, rows, 3)

	// First alert - CRITICAL, ACTIVE
	assert.Equal(t, "CRITICAL", rows[0].Severity)
	assert.Equal(t, "ACTIVE", rows[0].Status)
	assert.Equal(t, "Down", rows[0].Type)
	assert.Equal(t, "API Health", rows[0].Probe)
	assert.Equal(t, "5m ago", rows[0].Triggered)
	assert.Equal(t, "5m", rows[0].Duration)
	assert.Equal(t, "-", rows[0].AckBy)

	// Second alert - WARNING, ACKNOWLEDGED
	assert.Equal(t, "WARNING", rows[1].Severity)
	assert.Equal(t, "ACKNOWLEDGED", rows[1].Status)
	assert.Equal(t, "SSL Expiry", rows[1].Type)
	assert.Equal(t, "2h ago", rows[1].Triggered)
	assert.Equal(t, "99999999...", rows[1].AckBy)

	// Third alert - INFO, RESOLVED with stored duration
	assert.Equal(t, "INFO", rows[2].Severity)
	assert.Equal(t, "RESOLVED", rows[2].Status)
	assert.Equal(t, "Slow Response", rows[2].Type)
	assert.Equal(t, "1h", rows[2].Duration) // Uses stored durationSeconds
}

func TestFormatTriggeredTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"zero time", time.Time{}, "never"},
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5m ago"},
		{"2 hours ago", now.Add(-2 * time.Hour), "2h ago"},
		{"3 days ago", now.Add(-72 * time.Hour), "3d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTriggeredTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatAlertDuration(t *testing.T) {
	now := time.Now()
	fiveMinAgo := now.Add(-5 * time.Minute)
	twoHoursAgo := now.Add(-2 * time.Hour)
	oneHourAgo := now.Add(-1 * time.Hour)

	storedDuration := 7200 // 2 hours in seconds

	tests := []struct {
		name            string
		triggered       time.Time
		resolved        *time.Time
		durationSeconds *int
		expected        string
	}{
		{"zero triggered", time.Time{}, nil, nil, "-"},
		{"active alert (elapsed)", fiveMinAgo, nil, nil, "5m"},
		{"resolved with stored duration", twoHoursAgo, &oneHourAgo, &storedDuration, "2h"},
		{"resolved without stored duration", twoHoursAgo, &oneHourAgo, nil, "1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAlertDuration(tt.triggered, tt.resolved, tt.durationSeconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "0s"},
		{"30 seconds", 30 * time.Second, "30s"},
		{"5 minutes", 5 * time.Minute, "5m"},
		{"1 hour", time.Hour, "1h"},
		{"2 hours 30 min", 2*time.Hour + 30*time.Minute, "2h30m"},
		{"1 day", 24 * time.Hour, "1d"},
		{"2 days 4 hours", 52 * time.Hour, "2d4h"},
		{"3 days even", 72 * time.Hour, "3d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatAlertType(t *testing.T) {
	tests := []struct {
		alertType client.AlertType
		expected  string
	}{
		{client.AlertTypeStatusDown, "Down"},
		{client.AlertTypeSSLExpiry, "SSL Expiry"},
		{client.AlertTypeSSLInvalid, "SSL Invalid"},
		{client.AlertTypeSlowResponse, "Slow Response"},
		{client.AlertType("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.alertType), func(t *testing.T) {
			result := formatAlertType(tt.alertType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatProbeName(t *testing.T) {
	tests := []struct {
		name     string
		probe    *client.AlertProbe
		expected string
	}{
		{"nil probe", nil, "-"},
		{"short name", &client.AlertProbe{Name: "API Health"}, "API Health"},
		{"long name", &client.AlertProbe{Name: "This is a very long probe name that exceeds thirty characters"}, "This is a very long probe n..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatProbeName(tt.probe)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"empty string", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateMessage(t *testing.T) {
	longMsg := "This is a very long message that should be truncated to fit"
	shortMsg := "Short message"

	tests := []struct {
		name     string
		msg      *string
		maxLen   int
		expected string
	}{
		{"nil message", nil, 40, "-"},
		{"empty message", ptr(""), 40, "-"},
		{"short message", &shortMsg, 40, "Short message"},
		{"long message", &longMsg, 30, "This is a very long message..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateMessage(tt.msg, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatAcknowledgedBy(t *testing.T) {
	validUUID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")

	tests := []struct {
		name     string
		ackBy    *uuid.UUID
		expected string
	}{
		{"nil", nil, "-"},
		{"valid UUID", &validUUID, "12345678..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAcknowledgedBy(tt.ackBy)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAlertTableFormatter_SeverityColoring(t *testing.T) {
	formatter := NewAlertTableFormatter(sdkoutput.ColorAlways, false)

	// Test critical severity - should be red
	alert := client.Alert{
		ID:          uuid.New(),
		Status:      client.AlertStatusActive,
		Severity:    client.AlertSeverityCritical,
		AlertType:   client.AlertTypeStatusDown,
		TriggeredAt: time.Now(),
	}

	row := formatter.FormatAlert(alert)

	// When colors are enabled, the severity should contain ANSI escape codes
	assert.Contains(t, row.Severity, "CRITICAL")
	assert.Contains(t, row.Severity, "\x1b[") // ANSI escape code
}

func TestAlertTableFormatter_StatusColoring(t *testing.T) {
	formatter := NewAlertTableFormatter(sdkoutput.ColorAlways, false)

	tests := []struct {
		name   string
		status client.AlertStatus
	}{
		{"active", client.AlertStatusActive},
		{"acknowledged", client.AlertStatusAcknowledged},
		{"resolved", client.AlertStatusResolved},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := client.Alert{
				ID:          uuid.New(),
				Status:      tt.status,
				Severity:    client.AlertSeverityWarning,
				AlertType:   client.AlertTypeStatusDown,
				TriggeredAt: time.Now(),
			}

			row := formatter.FormatAlert(alert)

			// Status should contain ANSI escape codes when colors enabled
			assert.Contains(t, row.Status, "\x1b[")
		})
	}
}

func TestAlertTableFormatter_NoColor(t *testing.T) {
	formatter := NewAlertTableFormatter(sdkoutput.ColorNever, false)

	alert := client.Alert{
		ID:          uuid.New(),
		Status:      client.AlertStatusActive,
		Severity:    client.AlertSeverityCritical,
		AlertType:   client.AlertTypeStatusDown,
		TriggeredAt: time.Now(),
	}

	row := formatter.FormatAlert(alert)

	// When colors are disabled, should be plain text without ANSI codes
	assert.Equal(t, "CRITICAL", row.Severity)
	assert.Equal(t, "ACTIVE", row.Status)
	assert.NotContains(t, row.Severity, "\x1b[")
	assert.NotContains(t, row.Status, "\x1b[")
}

func TestNewAlertTableFormatter(t *testing.T) {
	// Test constructor creates valid formatter
	formatter := NewAlertTableFormatter(sdkoutput.ColorAuto, true)

	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.colorMgr)
	assert.True(t, formatter.isWide)

	formatterNoWide := NewAlertTableFormatter(sdkoutput.ColorNever, false)
	assert.False(t, formatterNoWide.isWide)
}
