package output

import (
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/stretchr/testify/assert"
)

func TestOrgTableFormatter_FormatOrganizations(t *testing.T) {
	orgs := []client.Organization{
		{
			ID:        "org-111-aaa",
			Name:      "Acme Corp",
			Slug:      "acme-corp",
			Role:      "owner",
			IsCurrent: true,
		},
		{
			ID:        "org-222-bbb",
			Name:      "Beta Inc",
			Slug:      "beta-inc",
			Role:      "admin",
			IsCurrent: false,
		},
		{
			ID:        "org-333-ccc",
			Name:      "Gamma LLC",
			Slug:      "gamma-llc",
			Role:      "member",
			IsCurrent: false,
		},
	}

	// Test with colors disabled for predictable output
	formatter := NewOrgTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatOrganizations(orgs)

	assert.Len(t, rows, 3)

	// First org - current org, owner
	assert.Equal(t, "*", rows[0].Status)
	assert.Equal(t, "Acme Corp", rows[0].Name)
	assert.Equal(t, "acme-corp", rows[0].Slug)
	assert.Equal(t, "Owner", rows[0].Role)
	assert.Equal(t, "org-111-aaa", rows[0].ID)

	// Second org - not current, admin
	assert.Equal(t, "", rows[1].Status)
	assert.Equal(t, "Beta Inc", rows[1].Name)
	assert.Equal(t, "beta-inc", rows[1].Slug)
	assert.Equal(t, "Admin", rows[1].Role)
	assert.Equal(t, "org-222-bbb", rows[1].ID)

	// Third org - not current, member
	assert.Equal(t, "", rows[2].Status)
	assert.Equal(t, "Gamma LLC", rows[2].Name)
	assert.Equal(t, "gamma-llc", rows[2].Slug)
	assert.Equal(t, "Member", rows[2].Role)
	assert.Equal(t, "org-333-ccc", rows[2].ID)
}

func TestFormatRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected string
	}{
		{"owner role", "owner", "Owner"},
		{"admin role", "admin", "Admin"},
		{"member role", "member", "Member"},
		{"uppercase role", "OWNER", "Owner"},
		{"mixed case role", "AdMiN", "Admin"},
		{"empty role", "", "-"},
		{"single char role", "a", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRole(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrgTableFormatter_StatusColoring(t *testing.T) {
	formatter := NewOrgTableFormatter(sdkoutput.ColorAlways, false)

	tests := []struct {
		name      string
		isCurrent bool
	}{
		{"current org", true},
		{"not current org", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org := client.Organization{
				ID:        "test-org-id",
				Name:      "Test Org",
				Slug:      "test-org",
				Role:      "member",
				IsCurrent: tt.isCurrent,
			}

			row := formatter.FormatOrganization(org)

			if tt.isCurrent {
				// Status should contain ANSI escape codes when colors enabled
				assert.Contains(t, row.Status, "\x1b[")
			} else {
				// Non-current orgs should have empty status
				assert.Equal(t, "", row.Status)
			}
		})
	}
}

func TestOrgTableFormatter_NoColor(t *testing.T) {
	formatter := NewOrgTableFormatter(sdkoutput.ColorNever, false)

	org := client.Organization{
		ID:        "test-org-id",
		Name:      "Test Org",
		Slug:      "test-org",
		Role:      "owner",
		IsCurrent: true,
	}

	row := formatter.FormatOrganization(org)

	// When colors are disabled, should be plain text without ANSI codes
	assert.Equal(t, "*", row.Status)
	assert.NotContains(t, row.Status, "\x1b[")
}

func TestNewOrgTableFormatter(t *testing.T) {
	// Test constructor creates valid formatter
	formatter := NewOrgTableFormatter(sdkoutput.ColorAuto, true)

	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.colorMgr)
	assert.True(t, formatter.isWide)

	formatterNoWide := NewOrgTableFormatter(sdkoutput.ColorNever, false)
	assert.False(t, formatterNoWide.isWide)
}

func TestOrgTableFormatter_WideMode(t *testing.T) {
	org := client.Organization{
		ID:        "wide-test-org-12345",
		Name:      "Wide Test Org",
		Slug:      "wide-test",
		Role:      "admin",
		IsCurrent: false,
	}

	formatter := NewOrgTableFormatter(sdkoutput.ColorNever, true)
	row := formatter.FormatOrganization(org)

	// Wide mode should include ID field
	assert.Equal(t, "wide-test-org-12345", row.ID)
	assert.Equal(t, "Wide Test Org", row.Name)
	assert.Equal(t, "wide-test", row.Slug)
	assert.Equal(t, "Admin", row.Role)
}

func TestOrgTableFormatter_EmptySlice(t *testing.T) {
	formatter := NewOrgTableFormatter(sdkoutput.ColorNever, false)
	rows := formatter.FormatOrganizations([]client.Organization{})

	assert.Len(t, rows, 0)
	assert.NotNil(t, rows) // Should return empty slice, not nil
}

func TestOrgTableFormatter_SingleOrganization(t *testing.T) {
	formatter := NewOrgTableFormatter(sdkoutput.ColorNever, false)

	org := client.Organization{
		ID:        "single-org-id",
		Name:      "Single Org",
		Slug:      "single",
		Role:      "owner",
		IsCurrent: true,
	}

	row := formatter.FormatOrganization(org)

	assert.Equal(t, "*", row.Status)
	assert.Equal(t, "Single Org", row.Name)
	assert.Equal(t, "single", row.Slug)
	assert.Equal(t, "Owner", row.Role)
	assert.Equal(t, "single-org-id", row.ID)
}

func TestOrgTableFormatter_AllColorModes(t *testing.T) {
	colorModes := []sdkoutput.ColorMode{
		sdkoutput.ColorAuto,
		sdkoutput.ColorAlways,
		sdkoutput.ColorNever,
	}

	org := client.Organization{
		ID:        "color-test-org",
		Name:      "Color Test",
		Slug:      "color-test",
		Role:      "member",
		IsCurrent: true,
	}

	for _, mode := range colorModes {
		t.Run(string(mode), func(t *testing.T) {
			formatter := NewOrgTableFormatter(mode, false)
			row := formatter.FormatOrganization(org)

			// All modes should produce valid output
			assert.NotEmpty(t, row.Status)
			assert.Equal(t, "Color Test", row.Name)
		})
	}
}
