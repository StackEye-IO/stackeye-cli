package output

import (
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProbeTableFormatter_FormatProbes(t *testing.T) {
	now := time.Now()
	fiveMinAgo := now.Add(-5 * time.Minute)
	twoHoursAgo := now.Add(-2 * time.Hour)

	probes := []client.Probe{
		{
			ID:                uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:              "API Health",
			URL:               "https://api.example.com/health",
			Status:            "up",
			IntervalSeconds:   60,
			CheckType:         client.CheckTypeHTTP,
			LastCheckedAt:     &fiveMinAgo,
			Regions:           []string{"us-east", "eu-west"},
			Uptime:            99.95,
			AvgResponseTimeMs: 245.5,
		},
		{
			ID:              uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:            "Website",
			URL:             "https://www.example.com",
			Status:          "down",
			IntervalSeconds: 300,
			CheckType:       client.CheckTypeHTTP,
			LastCheckedAt:   &twoHoursAgo,
			Regions:         []string{"us-east"},
			Uptime:          85.50,
		},
		{
			ID:              uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Name:            "New Probe",
			URL:             "https://new.example.com",
			Status:          "pending",
			IntervalSeconds: 30,
			CheckType:       client.CheckTypeHTTP,
			LastCheckedAt:   nil,
			Regions:         []string{},
		},
	}

	// Test with colors disabled for predictable output
	formatter := NewProbeTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatProbes(probes)

	assert.Len(t, rows, 3)

	// First probe - UP status
	assert.Equal(t, "UP", rows[0].Status)
	assert.Equal(t, "API Health", rows[0].Name)
	assert.Equal(t, "https://api.example.com/health", rows[0].URL)
	assert.Equal(t, "1m", rows[0].Interval)
	assert.Equal(t, "5m ago", rows[0].LastCheck)
	assert.Equal(t, "http", rows[0].Type)
	assert.Equal(t, "99.95%", rows[0].Uptime)
	assert.Equal(t, "246ms", rows[0].AvgResp)
	assert.Equal(t, "us-east,eu-west", rows[0].Regions)

	// Second probe - DOWN status
	assert.Equal(t, "DOWN", rows[1].Status)
	assert.Equal(t, "Website", rows[1].Name)
	assert.Equal(t, "5m", rows[1].Interval)
	assert.Equal(t, "2h ago", rows[1].LastCheck)
	assert.Equal(t, "85.50%", rows[1].Uptime)

	// Third probe - PENDING, never checked
	assert.Equal(t, "PENDING", rows[2].Status)
	assert.Equal(t, "30s", rows[2].Interval)
	assert.Equal(t, "never", rows[2].LastCheck)
	assert.Equal(t, "-", rows[2].Regions)
}

func TestFormatInterval(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{15, "15s"},
		{30, "30s"},
		{60, "1m"},
		{120, "2m"},
		{300, "5m"},
		{3600, "1h"},
		{7200, "2h"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatInterval(tt.seconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatLastCheck(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     *time.Time
		expected string
	}{
		{"nil time", nil, "never"},
		{"just now", ptr(now.Add(-30 * time.Second)), "just now"},
		{"5 minutes ago", ptr(now.Add(-5 * time.Minute)), "5m ago"},
		{"2 hours ago", ptr(now.Add(-2 * time.Hour)), "2h ago"},
		{"3 days ago", ptr(now.Add(-72 * time.Hour)), "3d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatLastCheck(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateURL(t *testing.T) {
	tests := []struct {
		url      string
		maxLen   int
		expected string
	}{
		{"https://short.com", 50, "https://short.com"},
		{"https://example.com/very/long/path/that/exceeds", 30, "https://example.com/very/lo..."},
		{"https://a.b", 50, "https://a.b"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := truncateURL(tt.url, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		uptime   float64
		expected string
	}{
		{0, "-"},
		{99.99, "99.99%"},
		{100.0, "100.00%"},
		{85.5, "85.50%"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatUptime(tt.uptime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatResponseTime(t *testing.T) {
	tests := []struct {
		ms       float64
		expected string
	}{
		{0, "-"},
		{150, "150ms"},
		{999, "999ms"},
		{1000, "1.00s"},
		{2500, "2.50s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatResponseTime(tt.ms)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatRegions(t *testing.T) {
	tests := []struct {
		name     string
		regions  []string
		expected string
	}{
		{"empty", []string{}, "-"},
		{"single", []string{"us-east"}, "us-east"},
		{"multiple", []string{"us-east", "eu-west"}, "us-east,eu-west"},
		{"three", []string{"us-east", "eu-west", "ap-south"}, "us-east,eu-west,ap-south"},
		{"more than three", []string{"us-east", "eu-west", "ap-south", "au-syd"}, "us-east,eu-west,ap-south +1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRegions(tt.regions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProbeTableFormatter_StatusColoring(t *testing.T) {
	// Test that status coloring is applied when colors are enabled
	formatter := NewProbeTableFormatter(sdkoutput.ColorAlways, false)

	probe := client.Probe{
		ID:              uuid.New(),
		Name:            "Test",
		URL:             "https://test.com",
		Status:          "up",
		IntervalSeconds: 60,
	}

	row := formatter.FormatProbe(probe)

	// When colors are enabled, the status should contain ANSI escape codes
	// The StatusColor method wraps "UP" in green color codes
	assert.Contains(t, row.Status, "UP")
	// ANSI green starts with \x1b[32m
	assert.Contains(t, row.Status, "\x1b[")
}

func TestProbeTableFormatter_NoColor(t *testing.T) {
	formatter := NewProbeTableFormatter(sdkoutput.ColorNever, false)

	probe := client.Probe{
		ID:              uuid.New(),
		Name:            "Test",
		URL:             "https://test.com",
		Status:          "down",
		IntervalSeconds: 60,
	}

	row := formatter.FormatProbe(probe)

	// When colors are disabled, status should be plain text
	assert.Equal(t, "DOWN", row.Status)
	assert.NotContains(t, row.Status, "\x1b[")
}

// ptr is a helper to create pointers to values
func ptr[T any](v T) *T {
	return &v
}
