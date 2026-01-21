// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"strings"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// OrgTableRow represents a row in the organization table output.
// The struct tags control column headers and wide mode display.
type OrgTableRow struct {
	Status string `table:"STATUS"`
	Name   string `table:"NAME"`
	Slug   string `table:"SLUG"`
	Role   string `table:"ROLE"`
	// Wide mode columns
	ID string `table:"ID,wide"`
}

// OrgTableFormatter converts SDK Organization types to table-displayable rows
// with current organization indicator coloring support.
type OrgTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewOrgTableFormatter creates a new formatter for organization table output.
// The colorMode parameter controls whether status colors are applied.
// Set isWide to true for extended output with additional columns.
func NewOrgTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *OrgTableFormatter {
	return &OrgTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatOrganizations converts a slice of SDK Organizations into table-displayable rows.
// The current organization is indicated with a green checkmark in the STATUS column.
func (f *OrgTableFormatter) FormatOrganizations(orgs []client.Organization) []OrgTableRow {
	rows := make([]OrgTableRow, 0, len(orgs))
	for _, org := range orgs {
		rows = append(rows, f.formatOrganization(org))
	}
	return rows
}

// FormatOrganization converts a single SDK Organization into a table-displayable row.
func (f *OrgTableFormatter) FormatOrganization(org client.Organization) OrgTableRow {
	return f.formatOrganization(org)
}

// formatOrganization is the internal conversion function.
func (f *OrgTableFormatter) formatOrganization(org client.Organization) OrgTableRow {
	return OrgTableRow{
		Status: f.formatCurrentStatus(org.IsCurrent),
		Name:   org.Name,
		Slug:   org.Slug,
		Role:   formatRole(org.Role),
		ID:     org.ID,
	}
}

// formatCurrentStatus applies color based on whether this is the current organization.
// Returns a colored checkmark for the current org, empty string otherwise.
func (f *OrgTableFormatter) formatCurrentStatus(isCurrent bool) string {
	if isCurrent {
		return f.colorMgr.StatusUp("*")
	}
	return ""
}

// formatRole formats the role string for display with proper capitalization.
func formatRole(role string) string {
	if role == "" {
		return "-"
	}
	// Capitalize first letter
	return strings.ToUpper(role[:1]) + strings.ToLower(role[1:])
}

// PrintOrganizations is a convenience function that formats and prints organizations
// using the CLI's configured output format. It handles status coloring
// and wide mode automatically based on configuration.
func PrintOrganizations(orgs []client.Organization) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewOrgTableFormatter(colorMode, isWide)
	rows := formatter.FormatOrganizations(orgs)

	return printer.Print(rows)
}

// PrintOrganization is a convenience function that formats and prints a single organization.
func PrintOrganization(org client.Organization) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewOrgTableFormatter(colorMode, isWide)
	row := formatter.FormatOrganization(org)

	return printer.Print(row)
}
