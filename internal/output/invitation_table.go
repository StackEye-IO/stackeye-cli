// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// InvitationTableRow represents a row in the invitation table output.
// The struct tags control column headers and wide mode display.
type InvitationTableRow struct {
	Email      string `table:"EMAIL"`
	Role       string `table:"ROLE"`
	InviteCode string `table:"INVITE_CODE"`
	Expires    string `table:"EXPIRES"`
	// Wide mode columns
	ID        string `table:"ID,wide"`
	InvitedBy string `table:"INVITED_BY,wide"`
}

// InvitationTableFormatter converts SDK Invitation types to table-displayable rows.
type InvitationTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewInvitationTableFormatter creates a new formatter for invitation table output.
// The colorMode parameter controls whether colors are applied.
// Set isWide to true for extended output with additional columns.
func NewInvitationTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *InvitationTableFormatter {
	return &InvitationTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatInvitations converts a slice of SDK Invitations into table-displayable rows.
func (f *InvitationTableFormatter) FormatInvitations(invitations []client.Invitation) []InvitationTableRow {
	rows := make([]InvitationTableRow, 0, len(invitations))
	for _, inv := range invitations {
		rows = append(rows, f.formatInvitation(inv))
	}
	return rows
}

// FormatInvitation converts a single SDK Invitation into a table-displayable row.
func (f *InvitationTableFormatter) FormatInvitation(invitation client.Invitation) InvitationTableRow {
	return f.formatInvitation(invitation)
}

// formatInvitation is the internal conversion function.
func (f *InvitationTableFormatter) formatInvitation(invitation client.Invitation) InvitationTableRow {
	return InvitationTableRow{
		Email:      invitation.Email,
		Role:       formatInvitationRole(string(invitation.Role)),
		InviteCode: invitation.InviteCode,
		Expires:    formatExpirationDate(invitation.ExpiresAt),
		ID:         formatInvitationID(invitation.ID),
		InvitedBy:  formatInvitedBy(invitation.InvitedBy),
	}
}

// formatInvitationRole formats the role string for display with proper capitalization.
func formatInvitationRole(role string) string {
	if role == "" {
		return "-"
	}
	return strings.ToUpper(role[:1]) + strings.ToLower(role[1:])
}

// formatExpirationDate formats the expiration date as a date string.
func formatExpirationDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}

// formatInvitationID formats the invitation ID for display.
func formatInvitationID(id string) string {
	if id == "" {
		return "-"
	}
	return id
}

// formatInvitedBy formats the inviter information.
func formatInvitedBy(invitedBy string) string {
	if invitedBy == "" {
		return "-"
	}
	return invitedBy
}

// PrintInvitations is a convenience function that formats and prints invitations
// using the CLI's configured output format. It handles coloring
// and wide mode automatically based on configuration.
func PrintInvitations(invitations []client.Invitation) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewInvitationTableFormatter(colorMode, isWide)
	rows := formatter.FormatInvitations(invitations)

	return printer.Print(rows)
}

// PrintInvitation is a convenience function that formats and prints a single invitation.
func PrintInvitation(invitation *client.Invitation) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewInvitationTableFormatter(colorMode, isWide)
	row := formatter.FormatInvitation(*invitation)

	return printer.Print(row)
}

// InvitationRevokedResponse is returned for JSON/YAML output when revoking an invitation.
type InvitationRevokedResponse struct {
	InvitationID string `json:"invitation_id"`
	Email        string `json:"email"`
	Revoked      bool   `json:"revoked"`
}

// PrintInvitationRevoked prints a success message after revoking an invitation.
// For JSON/YAML formats, it outputs a structured response object.
// For table format, it prints a human-friendly success message.
func PrintInvitationRevoked(invitationID, email string) error {
	printer := getPrinter()

	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		result := &InvitationRevokedResponse{
			InvitationID: invitationID,
			Email:        email,
			Revoked:      true,
		}
		return printer.Print(result)
	}

	// For table output, print a success message
	fmt.Println()
	fmt.Println("Invitation revoked successfully!")
	fmt.Println()
	fmt.Printf("  Invitation ID: %s\n", invitationID)
	if email != "" {
		fmt.Printf("  Email:         %s\n", email)
	}
	fmt.Println()
	fmt.Println("The invitation can no longer be used to join your organization.")
	fmt.Println()

	return nil
}

// PrintInvitationCreated prints a success message after creating an invitation.
// For JSON/YAML formats, it outputs the full invitation object.
// For table format, it prints a human-friendly success message.
func PrintInvitationCreated(invitation *client.Invitation) error {
	printer := getPrinter()

	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return printer.Print(invitation)
	}

	// For table output, print a success message with invitation details
	fmt.Println()
	fmt.Println("Invitation sent successfully!")
	fmt.Println()
	fmt.Printf("  Email:       %s\n", invitation.Email)
	fmt.Printf("  Role:        %s\n", formatInvitationRole(string(invitation.Role)))
	fmt.Printf("  Invite Code: %s\n", invitation.InviteCode)
	fmt.Printf("  Expires:     %s\n", formatExpirationDate(invitation.ExpiresAt))
	fmt.Println()
	fmt.Println("The invite code can be shared with the recipient to join your organization.")
	fmt.Println()

	return nil
}
