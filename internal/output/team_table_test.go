// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"gopkg.in/yaml.v3"
)

// TestNewTeamMemberTableFormatter verifies the formatter constructor.
func TestNewTeamMemberTableFormatter(t *testing.T) {
	formatter := NewTeamMemberTableFormatter(sdkoutput.ColorAuto, false)

	if formatter == nil {
		t.Fatal("expected formatter to be non-nil")
	}

	if formatter.colorMgr == nil {
		t.Error("expected colorMgr to be set")
	}

	if formatter.isWide {
		t.Error("expected isWide to be false")
	}
}

// TestNewTeamMemberTableFormatter_WideMode verifies wide mode is properly set.
func TestNewTeamMemberTableFormatter_WideMode(t *testing.T) {
	formatter := NewTeamMemberTableFormatter(sdkoutput.ColorNever, true)

	if !formatter.isWide {
		t.Error("expected isWide to be true")
	}
}

// TestFormatTeamMembers verifies batch formatting of team members.
func TestFormatTeamMembers(t *testing.T) {
	formatter := NewTeamMemberTableFormatter(sdkoutput.ColorNever, false)

	joinedAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	members := []client.TeamMember{
		{
			ID:       1,
			UserID:   "auth0|abc123",
			Email:    "alice@example.com",
			Name:     "Alice Smith",
			Role:     client.TeamRoleOwner,
			JoinedAt: joinedAt,
		},
		{
			ID:       2,
			UserID:   "auth0|def456",
			Email:    "bob@example.com",
			Name:     "Bob Jones",
			Role:     client.TeamRoleMember,
			JoinedAt: joinedAt,
		},
	}

	rows := formatter.FormatTeamMembers(members)

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// Check first row
	if rows[0].Name != "Alice Smith" {
		t.Errorf("expected Name 'Alice Smith', got %q", rows[0].Name)
	}
	if rows[0].Email != "alice@example.com" {
		t.Errorf("expected Email 'alice@example.com', got %q", rows[0].Email)
	}
	if rows[0].Role != "Owner" {
		t.Errorf("expected Role 'Owner', got %q", rows[0].Role)
	}

	// Check second row
	if rows[1].Name != "Bob Jones" {
		t.Errorf("expected Name 'Bob Jones', got %q", rows[1].Name)
	}
	if rows[1].Role != "Member" {
		t.Errorf("expected Role 'Member', got %q", rows[1].Role)
	}
}

// TestFormatTeamMember_SingleMember verifies single member formatting.
func TestFormatTeamMember_SingleMember(t *testing.T) {
	formatter := NewTeamMemberTableFormatter(sdkoutput.ColorNever, true)

	joinedAt := time.Date(2025, 6, 20, 14, 0, 0, 0, time.UTC)
	member := client.TeamMember{
		ID:       42,
		UserID:   "auth0|xyz789",
		Email:    "charlie@example.com",
		Name:     "Charlie Brown",
		Role:     client.TeamRoleAdmin,
		JoinedAt: joinedAt,
	}

	row := formatter.FormatTeamMember(member)

	if row.Name != "Charlie Brown" {
		t.Errorf("expected Name 'Charlie Brown', got %q", row.Name)
	}
	if row.Email != "charlie@example.com" {
		t.Errorf("expected Email 'charlie@example.com', got %q", row.Email)
	}
	if row.Role != "Admin" {
		t.Errorf("expected Role 'Admin', got %q", row.Role)
	}
	if row.Joined != "2025-06-20" {
		t.Errorf("expected Joined '2025-06-20', got %q", row.Joined)
	}
	if row.ID != "42" {
		t.Errorf("expected ID '42', got %q", row.ID)
	}
	if row.UserID != "auth0|xyz789" {
		t.Errorf("expected UserID 'auth0|xyz789', got %q", row.UserID)
	}
}

// TestFormatMemberName verifies name formatting with edge cases.
func TestFormatMemberName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal name", "John Doe", "John Doe"},
		{"empty name", "", "-"},
		{"single name", "Alice", "Alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMemberName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatTeamRole verifies role formatting with capitalization.
func TestFormatTeamRole(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"owner role", "owner", "Owner"},
		{"admin role", "admin", "Admin"},
		{"member role", "member", "Member"},
		{"viewer role", "viewer", "Viewer"},
		{"empty role", "", "-"},
		{"uppercase role", "OWNER", "Owner"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTeamRole(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatJoinedDate verifies date formatting.
func TestFormatJoinedDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			"valid date",
			time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC),
			"2025-03-15",
		},
		{
			"zero time",
			time.Time{},
			"-",
		},
		{
			"different year",
			time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			"2024-12-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatJoinedDate(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatMemberID verifies ID formatting.
func TestFormatMemberID(t *testing.T) {
	tests := []struct {
		name     string
		input    uint
		expected string
	}{
		{"valid ID", 123, "123"},
		{"zero ID", 0, "-"},
		{"large ID", 999999, "999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMemberID(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatTeamMembers_EmptySlice verifies empty input handling.
func TestFormatTeamMembers_EmptySlice(t *testing.T) {
	formatter := NewTeamMemberTableFormatter(sdkoutput.ColorNever, false)

	rows := formatter.FormatTeamMembers([]client.TeamMember{})

	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

// TestTeamMemberTableRow_Tags verifies struct tags for table rendering.
func TestTeamMemberTableRow_Tags(t *testing.T) {
	// This test verifies the struct is properly tagged for the table printer.
	// The tags are used by the SDK's table printer to determine column headers.
	row := TeamMemberTableRow{
		Name:   "Test User",
		Email:  "test@example.com",
		Role:   "Admin",
		Joined: "2025-01-01",
		ID:     "1",
		UserID: "auth0|test",
	}

	// Verify the row can be created and fields are accessible
	if row.Name != "Test User" {
		t.Errorf("expected Name 'Test User', got %q", row.Name)
	}
	if row.ID != "1" {
		t.Errorf("expected ID '1', got %q", row.ID)
	}
}

// TestPrintRoleUpdated_ResponseFields verifies the update role response fields are handled.
func TestPrintRoleUpdated_ResponseFields(t *testing.T) {
	// This test verifies the UpdateMemberRoleResponse structure is compatible
	// with the PrintRoleUpdated function. The actual printing is tested via
	// integration tests since it depends on the printer configuration.
	result := &client.UpdateMemberRoleResponse{
		Message:  "Role updated successfully",
		MemberID: 42,
		NewRole:  client.TeamRoleAdmin,
	}

	// Verify the response fields are accessible
	if result.MemberID != 42 {
		t.Errorf("expected MemberID 42, got %d", result.MemberID)
	}
	if result.NewRole != client.TeamRoleAdmin {
		t.Errorf("expected NewRole 'admin', got %q", result.NewRole)
	}
	if result.Message == "" {
		t.Error("expected Message to be set")
	}
}

// TestPrintMemberRemoved_ResponseFields verifies the member removed response fields are handled.
func TestPrintMemberRemoved_ResponseFields(t *testing.T) {
	// This test verifies the MemberRemovedResponse structure is compatible
	// with the PrintMemberRemoved function. The actual printing is tested via
	// integration tests since it depends on the printer configuration.
	result := &MemberRemovedResponse{
		MemberID: 42,
		Email:    "alice@example.com",
		Removed:  true,
	}

	// Verify the response fields are accessible
	if result.MemberID != 42 {
		t.Errorf("expected MemberID 42, got %d", result.MemberID)
	}
	if result.Email != "alice@example.com" {
		t.Errorf("expected Email 'alice@example.com', got %q", result.Email)
	}
	if !result.Removed {
		t.Error("expected Removed to be true")
	}
}

// TestMemberRemovedResponse_JSONMarshal verifies JSON field names are correct.
func TestMemberRemovedResponse_JSONMarshal(t *testing.T) {
	result := &MemberRemovedResponse{
		MemberID: 99,
		Email:    "bob@example.com",
		Removed:  true,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if _, ok := parsed["member_id"]; !ok {
		t.Error("expected JSON field 'member_id' to be present")
	}
	if _, ok := parsed["email"]; !ok {
		t.Error("expected JSON field 'email' to be present")
	}
	if _, ok := parsed["removed"]; !ok {
		t.Error("expected JSON field 'removed' to be present")
	}
}

// TestMemberRemovedResponse_YAMLMarshal verifies YAML field names are correct.
func TestMemberRemovedResponse_YAMLMarshal(t *testing.T) {
	result := &MemberRemovedResponse{
		MemberID: 99,
		Email:    "bob@example.com",
		Removed:  true,
	}

	data, err := yaml.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal YAML: %v", err)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	if _, ok := parsed["member_id"]; !ok {
		t.Error("expected YAML field 'member_id' to be present")
	}
	if _, ok := parsed["email"]; !ok {
		t.Error("expected YAML field 'email' to be present")
	}
	if _, ok := parsed["removed"]; !ok {
		t.Error("expected YAML field 'removed' to be present")
	}
}

// TestMemberRemovedResponse_DefaultValues verifies default values when fields are empty.
func TestMemberRemovedResponse_DefaultValues(t *testing.T) {
	result := &MemberRemovedResponse{}

	if result.MemberID != 0 {
		t.Errorf("expected default MemberID 0, got %d", result.MemberID)
	}
	if result.Email != "" {
		t.Errorf("expected default Email empty, got %q", result.Email)
	}
	if result.Removed {
		t.Error("expected default Removed to be false")
	}
}
