package output

import (
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/stretchr/testify/assert"
)

func TestNewRegionStatusFormatter(t *testing.T) {
	formatter := NewRegionStatusFormatter(sdkoutput.ColorAuto, true)

	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.colorMgr)
	assert.True(t, formatter.isWide)

	formatterNoWide := NewRegionStatusFormatter(sdkoutput.ColorNever, false)
	assert.False(t, formatterNoWide.isWide)
}

func TestRegionStatusFormatter_FormatRegionStatus(t *testing.T) {
	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, false)

	status := client.RegionStatus{
		ID:           "nyc3",
		Name:         "New York 3",
		Status:       "active",
		HealthStatus: "healthy",
	}

	row := formatter.FormatRegionStatus(status)

	assert.Equal(t, "nyc3", row.Code)
	assert.Equal(t, "New York 3", row.Name)
	assert.Equal(t, "active", row.Status)
	assert.Equal(t, "healthy", row.Health)
	assert.Equal(t, "-", row.Maintenance)
}

func TestRegionStatusFormatter_FormatRegionStatuses(t *testing.T) {
	statuses := []client.RegionStatus{
		{ID: "sfo1", Name: "San Francisco 1", Status: "active", HealthStatus: "healthy"},
		{ID: "fra1", Name: "Frankfurt 1", Status: "maintenance", HealthStatus: "warning"},
		{ID: "nyc1", Name: "New York 1", Status: "active", HealthStatus: "healthy"},
	}

	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegionStatuses(statuses)

	assert.Len(t, rows, 3)

	// Should be sorted by name
	assert.Equal(t, "fra1", rows[0].Code)
	assert.Equal(t, "Frankfurt 1", rows[0].Name)

	assert.Equal(t, "nyc1", rows[1].Code)
	assert.Equal(t, "New York 1", rows[1].Name)

	assert.Equal(t, "sfo1", rows[2].Code)
	assert.Equal(t, "San Francisco 1", rows[2].Name)
}

func TestRegionStatusFormatter_FormatOperationalStatus(t *testing.T) {
	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, false)

	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"active status", "active", "active"},
		{"maintenance status", "maintenance", "maintenance"},
		{"disabled status", "disabled", "disabled"},
		{"unknown status", "custom", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatOperationalStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegionStatusFormatter_FormatHealthStatus(t *testing.T) {
	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, false)

	tests := []struct {
		name     string
		health   string
		expected string
	}{
		{"healthy status", "healthy", "healthy"},
		{"warning status", "warning", "warning"},
		{"degraded status", "degraded", "degraded"},
		{"unknown status", "unknown", "unknown"},
		{"custom status", "custom", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatHealthStatus(tt.health)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegionStatusFormatter_FormatMaintenance(t *testing.T) {
	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, false)

	reason := "Scheduled upgrade"
	endsAt := time.Date(2026, 1, 21, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		reason   *string
		endsAt   *time.Time
		expected string
	}{
		{"no maintenance info", nil, nil, "-"},
		{"reason only", &reason, nil, "Scheduled upgrade"},
		{"end time only", nil, &endsAt, "Until Jan 21 14:30 UTC"},
		{"both reason and end time", &reason, &endsAt, "Scheduled upgrade (until Jan 21 14:30 UTC)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatMaintenance(tt.reason, tt.endsAt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegionStatusFormatter_FormatMaintenanceEmptyReason(t *testing.T) {
	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, false)

	emptyReason := ""
	result := formatter.formatMaintenance(&emptyReason, nil)
	assert.Equal(t, "-", result)
}

func TestRegionStatusFormatter_EmptySlice(t *testing.T) {
	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegionStatuses([]client.RegionStatus{})

	assert.Len(t, rows, 0)
	assert.NotNil(t, rows)
}

func TestRegionStatusFormatter_SingleStatus(t *testing.T) {
	statuses := []client.RegionStatus{
		{ID: "ams1", Name: "Amsterdam 1", Status: "active", HealthStatus: "healthy"},
	}

	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegionStatuses(statuses)

	assert.Len(t, rows, 1)
	assert.Equal(t, "ams1", rows[0].Code)
	assert.Equal(t, "Amsterdam 1", rows[0].Name)
	assert.Equal(t, "active", rows[0].Status)
	assert.Equal(t, "healthy", rows[0].Health)
}

func TestRegionStatusFormatter_AllColorModes(t *testing.T) {
	colorModes := []sdkoutput.ColorMode{
		sdkoutput.ColorAuto,
		sdkoutput.ColorAlways,
		sdkoutput.ColorNever,
	}

	status := client.RegionStatus{
		ID:           "test1",
		Name:         "Test Region 1",
		Status:       "active",
		HealthStatus: "healthy",
	}

	for _, mode := range colorModes {
		t.Run(string(mode), func(t *testing.T) {
			formatter := NewRegionStatusFormatter(mode, false)
			row := formatter.FormatRegionStatus(status)

			// All modes should produce valid output with at least the base text
			assert.Contains(t, row.Status, "active")
			assert.Contains(t, row.Health, "healthy")
		})
	}
}

func TestRegionStatusFormatter_WideMode(t *testing.T) {
	reason := "Database migration"
	endsAt := time.Date(2026, 1, 22, 10, 0, 0, 0, time.UTC)

	status := client.RegionStatus{
		ID:                "wide-test",
		Name:              "Wide Test Region",
		Status:            "maintenance",
		HealthStatus:      "warning",
		MaintenanceReason: &reason,
		MaintenanceEndsAt: &endsAt,
	}

	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, true)
	row := formatter.FormatRegionStatus(status)

	assert.Equal(t, "wide-test", row.Code)
	assert.Equal(t, "Wide Test Region", row.Name)
	assert.Equal(t, "maintenance", row.Status)
	assert.Equal(t, "warning", row.Health)
	assert.Contains(t, row.Maintenance, "Database migration")
	assert.Contains(t, row.Maintenance, "Jan 22")
}

func TestRegionStatusFormatter_AllStatuses(t *testing.T) {
	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, false)

	statuses := []client.RegionStatus{
		{ID: "r1", Name: "Region 1", Status: "active", HealthStatus: "healthy"},
		{ID: "r2", Name: "Region 2", Status: "maintenance", HealthStatus: "warning"},
		{ID: "r3", Name: "Region 3", Status: "disabled", HealthStatus: "degraded"},
		{ID: "r4", Name: "Region 4", Status: "active", HealthStatus: "unknown"},
	}

	rows := formatter.FormatRegionStatuses(statuses)

	assert.Len(t, rows, 4)

	// Verify each status type is properly formatted
	statusMap := make(map[string]RegionStatusTableRow)
	for _, row := range rows {
		statusMap[row.Code] = row
	}

	assert.Equal(t, "active", statusMap["r1"].Status)
	assert.Equal(t, "healthy", statusMap["r1"].Health)

	assert.Equal(t, "maintenance", statusMap["r2"].Status)
	assert.Equal(t, "warning", statusMap["r2"].Health)

	assert.Equal(t, "disabled", statusMap["r3"].Status)
	assert.Equal(t, "degraded", statusMap["r3"].Health)

	assert.Equal(t, "active", statusMap["r4"].Status)
	assert.Equal(t, "unknown", statusMap["r4"].Health)
}

func TestRegionStatusFormatter_SortingIsCaseInsensitive(t *testing.T) {
	statuses := []client.RegionStatus{
		{ID: "z1", Name: "Zurich 1", Status: "active", HealthStatus: "healthy"},
		{ID: "a1", Name: "Amsterdam 1", Status: "active", HealthStatus: "healthy"},
		{ID: "b1", Name: "Berlin 1", Status: "active", HealthStatus: "healthy"},
	}

	formatter := NewRegionStatusFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegionStatuses(statuses)

	assert.Len(t, rows, 3)
	assert.Equal(t, "Amsterdam 1", rows[0].Name)
	assert.Equal(t, "Berlin 1", rows[1].Name)
	assert.Equal(t, "Zurich 1", rows[2].Name)
}
