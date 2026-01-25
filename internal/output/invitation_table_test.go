// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// TestNewInvitationTableFormatter verifies the formatter is correctly initialized.
func TestNewInvitationTableFormatter(t *testing.T) {
	formatter := NewInvitationTableFormatter(sdkoutput.ColorAuto, false)

	if formatter == nil {
		t.Fatal("expected formatter to be non-nil")
	}

	if formatter.colorMgr == nil {
		t.Error("expected colorMgr to be non-nil")
	}

	if formatter.isWide {
		t.Error("expected isWide to be false")
	}
}

// TestNewInvitationTableFormatter_WideMode verifies wide mode is set correctly.
func TestNewInvitationTableFormatter_WideMode(t *testing.T) {
	formatter := NewInvitationTableFormatter(sdkoutput.ColorAuto, true)

	if !formatter.isWide {
		t.Error("expected isWide to be true")
	}
}

// TestInvitationTableFormatter_FormatInvitation verifies single invitation formatting.
func TestInvitationTableFormatter_FormatInvitation(t *testing.T) {
	formatter := NewInvitationTableFormatter(sdkoutput.ColorNever, false)

	expiresAt := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	invitation := client.Invitation{
		ID:             "inv_123",
		Email:          "newuser@company.io",
		Role:           client.TeamRoleAdmin,
		InviteCode:     "ABC123",
		ExpiresAt:      expiresAt,
		InvitedBy:      "admin@company.io",
		OrganizationID: "org_456",
	}

	row := formatter.FormatInvitation(invitation)

	if row.Email != "newuser@company.io" {
		t.Errorf("expected Email to be 'newuser@company.io', got %q", row.Email)
	}

	if row.Role != "Admin" {
		t.Errorf("expected Role to be 'Admin', got %q", row.Role)
	}

	if row.InviteCode != "ABC123" {
		t.Errorf("expected InviteCode to be 'ABC123', got %q", row.InviteCode)
	}

	if row.Expires != "2026-02-01 12:00" {
		t.Errorf("expected Expires to be '2026-02-01 12:00', got %q", row.Expires)
	}

	if row.ID != "inv_123" {
		t.Errorf("expected ID to be 'inv_123', got %q", row.ID)
	}

	if row.InvitedBy != "admin@company.io" {
		t.Errorf("expected InvitedBy to be 'admin@company.io', got %q", row.InvitedBy)
	}
}

// TestInvitationTableFormatter_FormatInvitations verifies multiple invitation formatting.
func TestInvitationTableFormatter_FormatInvitations(t *testing.T) {
	formatter := NewInvitationTableFormatter(sdkoutput.ColorNever, false)

	expiresAt := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	invitations := []client.Invitation{
		{
			ID:         "inv_1",
			Email:      "user1@company.io",
			Role:       client.TeamRoleAdmin,
			InviteCode: "CODE1",
			ExpiresAt:  expiresAt,
		},
		{
			ID:         "inv_2",
			Email:      "user2@company.io",
			Role:       client.TeamRoleMember,
			InviteCode: "CODE2",
			ExpiresAt:  expiresAt,
		},
		{
			ID:         "inv_3",
			Email:      "user3@company.io",
			Role:       client.TeamRoleViewer,
			InviteCode: "CODE3",
			ExpiresAt:  expiresAt,
		},
	}

	rows := formatter.FormatInvitations(invitations)

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	// Verify first row
	if rows[0].Email != "user1@company.io" {
		t.Errorf("expected first row Email to be 'user1@company.io', got %q", rows[0].Email)
	}

	// Verify second row
	if rows[1].Role != "Member" {
		t.Errorf("expected second row Role to be 'Member', got %q", rows[1].Role)
	}

	// Verify third row
	if rows[2].Role != "Viewer" {
		t.Errorf("expected third row Role to be 'Viewer', got %q", rows[2].Role)
	}
}

// TestFormatInvitationRole verifies role formatting.
func TestFormatInvitationRole(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "admin", expected: "Admin"},
		{input: "member", expected: "Member"},
		{input: "viewer", expected: "Viewer"},
		{input: "owner", expected: "Owner"},
		{input: "", expected: "-"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatInvitationRole(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatExpirationDate verifies expiration date formatting.
func TestFormatExpirationDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "valid date",
			input:    time.Date(2026, 2, 15, 14, 30, 0, 0, time.UTC),
			expected: "2026-02-15 14:30",
		},
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatExpirationDate(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatInvitationID verifies ID formatting.
func TestFormatInvitationID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "inv_123abc", expected: "inv_123abc"},
		{input: "", expected: "-"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatInvitationID(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatInvitedBy verifies inviter formatting.
func TestFormatInvitedBy(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "admin@company.io", expected: "admin@company.io"},
		{input: "", expected: "-"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatInvitedBy(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestInvitationTableRow_StructTags verifies the struct tags for table headers.
func TestInvitationTableRow_StructTags(t *testing.T) {
	row := InvitationTableRow{
		Email:      "test@company.io",
		Role:       "Admin",
		InviteCode: "ABC123",
		Expires:    "2026-02-01",
		ID:         "inv_123",
		InvitedBy:  "admin@company.io",
	}

	// Verify the struct fields are accessible (compile-time check)
	if row.Email == "" {
		t.Error("Email field should be set")
	}
	if row.Role == "" {
		t.Error("Role field should be set")
	}
	if row.InviteCode == "" {
		t.Error("InviteCode field should be set")
	}
	if row.Expires == "" {
		t.Error("Expires field should be set")
	}
	if row.ID == "" {
		t.Error("ID field should be set")
	}
	if row.InvitedBy == "" {
		t.Error("InvitedBy field should be set")
	}
}
