package output

import (
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/stretchr/testify/assert"
)

func TestRegionTableFormatter_FormatRegions(t *testing.T) {
	regionsByContinent := map[string][]client.Region{
		"north_america": {
			{ID: "nyc1", Name: "New York 1", DisplayName: "New York", CountryCode: "US"},
			{ID: "sfo1", Name: "San Francisco 1", DisplayName: "San Francisco", CountryCode: "US"},
		},
		"europe": {
			{ID: "fra1", Name: "Frankfurt 1", DisplayName: "Frankfurt", CountryCode: "DE"},
			{ID: "lon1", Name: "London 1", DisplayName: "London", CountryCode: "GB"},
		},
	}

	formatter := NewRegionTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegions(regionsByContinent)

	// Should have 4 rows total
	assert.Len(t, rows, 4)

	// Regions should be sorted by continent (europe before north_america alphabetically)
	// Then by name within each continent
	assert.Equal(t, "fra1", rows[0].Code)
	assert.Equal(t, "Frankfurt 1", rows[0].Name)
	assert.Equal(t, "Frankfurt", rows[0].Display)
	assert.Equal(t, "DE", rows[0].Country)
	assert.Equal(t, "Europe", rows[0].Continent)

	assert.Equal(t, "lon1", rows[1].Code)
	assert.Equal(t, "London 1", rows[1].Name)
	assert.Equal(t, "London", rows[1].Display)
	assert.Equal(t, "GB", rows[1].Country)
	assert.Equal(t, "Europe", rows[1].Continent)

	assert.Equal(t, "nyc1", rows[2].Code)
	assert.Equal(t, "New York 1", rows[2].Name)
	assert.Equal(t, "New York", rows[2].Display)
	assert.Equal(t, "US", rows[2].Country)
	assert.Equal(t, "North America", rows[2].Continent)

	assert.Equal(t, "sfo1", rows[3].Code)
	assert.Equal(t, "San Francisco 1", rows[3].Name)
	assert.Equal(t, "San Francisco", rows[3].Display)
	assert.Equal(t, "US", rows[3].Country)
	assert.Equal(t, "North America", rows[3].Continent)
}

func TestRegionTableFormatter_FormatRegionsFlat(t *testing.T) {
	regions := []client.Region{
		{ID: "sfo1", Name: "San Francisco 1", DisplayName: "San Francisco", CountryCode: "US"},
		{ID: "fra1", Name: "Frankfurt 1", DisplayName: "Frankfurt", CountryCode: "DE"},
		{ID: "nyc1", Name: "New York 1", DisplayName: "New York", CountryCode: "US"},
	}

	formatter := NewRegionTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegionsFlat(regions)

	assert.Len(t, rows, 3)

	// Should be sorted by name
	assert.Equal(t, "fra1", rows[0].Code)
	assert.Equal(t, "Frankfurt 1", rows[0].Name)
	assert.Equal(t, "-", rows[0].Continent) // Flat mode shows "-" for continent

	assert.Equal(t, "nyc1", rows[1].Code)
	assert.Equal(t, "New York 1", rows[1].Name)
	assert.Equal(t, "-", rows[1].Continent)

	assert.Equal(t, "sfo1", rows[2].Code)
	assert.Equal(t, "San Francisco 1", rows[2].Name)
	assert.Equal(t, "-", rows[2].Continent)
}

func TestFormatContinentName(t *testing.T) {
	tests := []struct {
		name     string
		slug     string
		expected string
	}{
		{"north america", "north_america", "North America"},
		{"south america", "south_america", "South America"},
		{"europe", "europe", "Europe"},
		{"asia pacific", "asia_pacific", "Asia Pacific"},
		{"africa", "africa", "Africa"},
		{"oceania", "oceania", "Oceania"},
		{"middle east", "middle_east", "Middle East"},
		{"dash placeholder", "-", "-"},
		{"unknown slug", "unknown_continent", "unknown_continent"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatContinentName(tt.slug)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewRegionTableFormatter(t *testing.T) {
	formatter := NewRegionTableFormatter(sdkoutput.ColorAuto, true)

	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.colorMgr)
	assert.True(t, formatter.isWide)

	formatterNoWide := NewRegionTableFormatter(sdkoutput.ColorNever, false)
	assert.False(t, formatterNoWide.isWide)
}

func TestRegionTableFormatter_FormatRegion(t *testing.T) {
	formatter := NewRegionTableFormatter(sdkoutput.ColorNever, false)

	region := client.Region{
		ID:          "nyc3",
		Name:        "New York 3",
		DisplayName: "New York",
		CountryCode: "US",
	}

	row := formatter.FormatRegion(region, "north_america")

	assert.Equal(t, "nyc3", row.Code)
	assert.Equal(t, "New York 3", row.Name)
	assert.Equal(t, "New York", row.Display)
	assert.Equal(t, "US", row.Country)
	assert.Equal(t, "North America", row.Continent)
}

func TestRegionTableFormatter_EmptyMap(t *testing.T) {
	formatter := NewRegionTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegions(map[string][]client.Region{})

	assert.Len(t, rows, 0)
	assert.NotNil(t, rows) // Should return empty slice, not nil
}

func TestRegionTableFormatter_EmptySlice(t *testing.T) {
	formatter := NewRegionTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegionsFlat([]client.Region{})

	assert.Len(t, rows, 0)
	assert.NotNil(t, rows)
}

func TestRegionTableFormatter_SingleRegion(t *testing.T) {
	regionsByContinent := map[string][]client.Region{
		"europe": {
			{ID: "ams1", Name: "Amsterdam 1", DisplayName: "Amsterdam", CountryCode: "NL"},
		},
	}

	formatter := NewRegionTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegions(regionsByContinent)

	assert.Len(t, rows, 1)
	assert.Equal(t, "ams1", rows[0].Code)
	assert.Equal(t, "Amsterdam 1", rows[0].Name)
	assert.Equal(t, "Amsterdam", rows[0].Display)
	assert.Equal(t, "NL", rows[0].Country)
	assert.Equal(t, "Europe", rows[0].Continent)
}

func TestRegionTableFormatter_AllColorModes(t *testing.T) {
	colorModes := []sdkoutput.ColorMode{
		sdkoutput.ColorAuto,
		sdkoutput.ColorAlways,
		sdkoutput.ColorNever,
	}

	region := client.Region{
		ID:          "test1",
		Name:        "Test Region 1",
		DisplayName: "Test Region",
		CountryCode: "XX",
	}

	for _, mode := range colorModes {
		t.Run(string(mode), func(t *testing.T) {
			formatter := NewRegionTableFormatter(mode, false)
			row := formatter.FormatRegion(region, "europe")

			// All modes should produce valid output
			assert.Equal(t, "test1", row.Code)
			assert.Equal(t, "Test Region 1", row.Name)
		})
	}
}

func TestRegionTableFormatter_WideMode(t *testing.T) {
	region := client.Region{
		ID:          "wide-test",
		Name:        "Wide Test Region",
		DisplayName: "Wide Test",
		CountryCode: "WT",
	}

	formatter := NewRegionTableFormatter(sdkoutput.ColorNever, true)
	row := formatter.FormatRegion(region, "asia_pacific")

	// Wide mode includes continent column (shown by default due to struct tag)
	assert.Equal(t, "wide-test", row.Code)
	assert.Equal(t, "Wide Test Region", row.Name)
	assert.Equal(t, "Wide Test", row.Display)
	assert.Equal(t, "WT", row.Country)
	assert.Equal(t, "Asia Pacific", row.Continent)
}

func TestRegionTableFormatter_SortingWithinContinent(t *testing.T) {
	regionsByContinent := map[string][]client.Region{
		"north_america": {
			{ID: "sfo1", Name: "San Francisco 1", DisplayName: "San Francisco", CountryCode: "US"},
			{ID: "nyc1", Name: "New York 1", DisplayName: "New York", CountryCode: "US"},
			{ID: "tor1", Name: "Toronto 1", DisplayName: "Toronto", CountryCode: "CA"},
		},
	}

	formatter := NewRegionTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegions(regionsByContinent)

	assert.Len(t, rows, 3)
	// Should be sorted by name within continent
	assert.Equal(t, "New York 1", rows[0].Name)
	assert.Equal(t, "San Francisco 1", rows[1].Name)
	assert.Equal(t, "Toronto 1", rows[2].Name)
}

func TestRegionTableFormatter_MultipleContinentsSorted(t *testing.T) {
	regionsByContinent := map[string][]client.Region{
		"oceania": {
			{ID: "syd1", Name: "Sydney 1", DisplayName: "Sydney", CountryCode: "AU"},
		},
		"africa": {
			{ID: "jnb1", Name: "Johannesburg 1", DisplayName: "Johannesburg", CountryCode: "ZA"},
		},
		"europe": {
			{ID: "ams1", Name: "Amsterdam 1", DisplayName: "Amsterdam", CountryCode: "NL"},
		},
	}

	formatter := NewRegionTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatRegions(regionsByContinent)

	assert.Len(t, rows, 3)
	// Continents should be sorted alphabetically
	assert.Equal(t, "Africa", rows[0].Continent)
	assert.Equal(t, "Europe", rows[1].Continent)
	assert.Equal(t, "Oceania", rows[2].Continent)
}
