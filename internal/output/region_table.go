// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"sort"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// RegionTableRow represents a row in the region table output.
// The struct tags control column headers and wide mode display.
type RegionTableRow struct {
	Code      string `table:"CODE"`
	Name      string `table:"NAME"`
	Display   string `table:"DISPLAY"`
	Country   string `table:"COUNTRY"`
	Continent string `table:"CONTINENT,wide"`
}

// RegionTableFormatter converts SDK Region types to table-displayable rows.
type RegionTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewRegionTableFormatter creates a new formatter for region table output.
// The colorMode parameter controls whether colors are applied.
// Set isWide to true for extended output with continent column.
func NewRegionTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *RegionTableFormatter {
	return &RegionTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatRegions converts a map of regions grouped by continent into table-displayable rows.
// Regions are sorted by continent then by name for consistent output.
func (f *RegionTableFormatter) FormatRegions(regionsByContinent map[string][]client.Region) []RegionTableRow {
	// Count total regions for pre-allocation
	total := 0
	for _, regions := range regionsByContinent {
		total += len(regions)
	}

	rows := make([]RegionTableRow, 0, total)

	// Sort continent keys for consistent output
	continents := make([]string, 0, len(regionsByContinent))
	for continent := range regionsByContinent {
		continents = append(continents, continent)
	}
	sort.Strings(continents)

	for _, continent := range continents {
		regions := regionsByContinent[continent]
		// Sort regions by name within each continent
		sortedRegions := make([]client.Region, len(regions))
		copy(sortedRegions, regions)
		sort.Slice(sortedRegions, func(i, j int) bool {
			return sortedRegions[i].Name < sortedRegions[j].Name
		})

		for _, region := range sortedRegions {
			rows = append(rows, f.formatRegion(region, continent))
		}
	}

	return rows
}

// FormatRegionsFlat converts a flat slice of regions into table-displayable rows.
// Regions are sorted by name for consistent output. Continent column shows "-".
func (f *RegionTableFormatter) FormatRegionsFlat(regions []client.Region) []RegionTableRow {
	rows := make([]RegionTableRow, 0, len(regions))

	// Sort by name for consistent output
	sortedRegions := make([]client.Region, len(regions))
	copy(sortedRegions, regions)
	sort.Slice(sortedRegions, func(i, j int) bool {
		return sortedRegions[i].Name < sortedRegions[j].Name
	})

	for _, region := range sortedRegions {
		rows = append(rows, f.formatRegion(region, "-"))
	}

	return rows
}

// FormatRegion converts a single SDK Region into a table-displayable row.
func (f *RegionTableFormatter) FormatRegion(region client.Region, continent string) RegionTableRow {
	return f.formatRegion(region, continent)
}

// formatRegion is the internal conversion function.
func (f *RegionTableFormatter) formatRegion(region client.Region, continent string) RegionTableRow {
	return RegionTableRow{
		Code:      region.ID,
		Name:      region.Name,
		Display:   region.DisplayName,
		Country:   region.CountryCode,
		Continent: formatContinentName(continent),
	}
}

// formatContinentName converts a continent slug to a display-friendly name.
func formatContinentName(slug string) string {
	switch slug {
	case "north_america":
		return "North America"
	case "south_america":
		return "South America"
	case "europe":
		return "Europe"
	case "asia_pacific":
		return "Asia Pacific"
	case "africa":
		return "Africa"
	case "oceania":
		return "Oceania"
	case "middle_east":
		return "Middle East"
	case "-":
		return "-"
	default:
		return slug
	}
}

// PrintRegions is a convenience function that formats and prints regions
// using the CLI's configured output format. It handles wide mode automatically.
func PrintRegions(regionsByContinent map[string][]client.Region) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewRegionTableFormatter(colorMode, isWide)
	rows := formatter.FormatRegions(regionsByContinent)

	return printer.Print(rows)
}

// PrintRegionsFlat is a convenience function for printing a flat list of regions.
func PrintRegionsFlat(regions []client.Region) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewRegionTableFormatter(colorMode, isWide)
	rows := formatter.FormatRegionsFlat(regions)

	return printer.Print(rows)
}
