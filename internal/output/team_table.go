// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// TeamMemberTableRow represents a row in the team member table output.
// The struct tags control column headers and wide mode display.
type TeamMemberTableRow struct {
	Name   string `table:"NAME"`
	Email  string `table:"EMAIL"`
	Role   string `table:"ROLE"`
	Joined string `table:"JOINED"`
	// Wide mode columns
	ID     string `table:"ID,wide"`
	UserID string `table:"USER_ID,wide"`
}

// TeamMemberTableFormatter converts SDK TeamMember types to table-displayable rows.
type TeamMemberTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewTeamMemberTableFormatter creates a new formatter for team member table output.
// The colorMode parameter controls whether colors are applied.
// Set isWide to true for extended output with additional columns.
func NewTeamMemberTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *TeamMemberTableFormatter {
	return &TeamMemberTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatTeamMembers converts a slice of SDK TeamMembers into table-displayable rows.
func (f *TeamMemberTableFormatter) FormatTeamMembers(members []client.TeamMember) []TeamMemberTableRow {
	rows := make([]TeamMemberTableRow, 0, len(members))
	for _, member := range members {
		rows = append(rows, f.formatTeamMember(member))
	}
	return rows
}

// FormatTeamMember converts a single SDK TeamMember into a table-displayable row.
func (f *TeamMemberTableFormatter) FormatTeamMember(member client.TeamMember) TeamMemberTableRow {
	return f.formatTeamMember(member)
}

// formatTeamMember is the internal conversion function.
func (f *TeamMemberTableFormatter) formatTeamMember(member client.TeamMember) TeamMemberTableRow {
	return TeamMemberTableRow{
		Name:   formatMemberName(member.Name),
		Email:  member.Email,
		Role:   formatTeamRole(string(member.Role)),
		Joined: formatJoinedDate(member.JoinedAt),
		ID:     formatMemberID(member.ID),
		UserID: member.UserID,
	}
}

// formatMemberName formats the member name, returning a placeholder if empty.
func formatMemberName(name string) string {
	if name == "" {
		return "-"
	}
	return name
}

// formatTeamRole formats the role string for display with proper capitalization.
func formatTeamRole(role string) string {
	if role == "" {
		return "-"
	}
	// Capitalize first letter
	return strings.ToUpper(role[:1]) + strings.ToLower(role[1:])
}

// formatJoinedDate formats the join date as a relative time or date string.
func formatJoinedDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02")
}

// formatMemberID converts the member ID to a string.
func formatMemberID(id uint) string {
	if id == 0 {
		return "-"
	}
	return strconv.FormatUint(uint64(id), 10)
}

// PrintTeamMembers is a convenience function that formats and prints team members
// using the CLI's configured output format. It handles coloring
// and wide mode automatically based on configuration.
func PrintTeamMembers(members []client.TeamMember) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewTeamMemberTableFormatter(colorMode, isWide)
	rows := formatter.FormatTeamMembers(members)

	return printer.Print(rows)
}

// PrintTeamMember is a convenience function that formats and prints a single team member.
func PrintTeamMember(member client.TeamMember) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewTeamMemberTableFormatter(colorMode, isWide)
	row := formatter.FormatTeamMember(member)

	return printer.Print(row)
}

// PrintRoleUpdated prints a success message after updating a member's role.
// For JSON/YAML formats, it outputs the full response object.
// For table format, it prints a human-friendly success message.
func PrintRoleUpdated(result *client.UpdateMemberRoleResponse) error {
	printer := getPrinter()

	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(result)
	}

	// For table output, print a success message
	fmt.Println()
	fmt.Println("Role updated successfully!")
	fmt.Println()
	fmt.Printf("  Member ID: %d\n", result.MemberID)
	fmt.Printf("  New Role:  %s\n", formatTeamRole(string(result.NewRole)))
	fmt.Println()

	return nil
}
