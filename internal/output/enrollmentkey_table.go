// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// EnrollmentKeyTableRow represents a row in the enrollment key table output.
// The struct tags control column headers and wide mode display.
type EnrollmentKeyTableRow struct {
	Name         string `table:"NAME"`
	KeyPrefix    string `table:"PREFIX"`
	Mode         string `table:"MODE"`
	Capabilities string `table:"CAPABILITIES"`
	Uses         string `table:"USES"`
	Expires      string `table:"EXPIRES"`
	// Wide mode columns
	Environment string `table:"ENVIRONMENT,wide"`
	Created     string `table:"CREATED,wide"`
	ID          string `table:"ID,wide"`
}

// EnrollmentKeyTableFormatter converts SDK EnrollmentKey types to table-displayable rows.
type EnrollmentKeyTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewEnrollmentKeyTableFormatter creates a new formatter for enrollment key table output.
func NewEnrollmentKeyTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *EnrollmentKeyTableFormatter {
	return &EnrollmentKeyTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatEnrollmentKeys converts a slice of SDK EnrollmentKeys into table-displayable rows.
func (f *EnrollmentKeyTableFormatter) FormatEnrollmentKeys(keys []client.EnrollmentKey) []EnrollmentKeyTableRow {
	rows := make([]EnrollmentKeyTableRow, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, f.formatEnrollmentKey(k))
	}
	return rows
}

// FormatEnrollmentKey converts a single SDK EnrollmentKey into a table-displayable row.
func (f *EnrollmentKeyTableFormatter) FormatEnrollmentKey(key client.EnrollmentKey) EnrollmentKeyTableRow {
	return f.formatEnrollmentKey(key)
}

// formatEnrollmentKey is the internal conversion function.
func (f *EnrollmentKeyTableFormatter) formatEnrollmentKey(k client.EnrollmentKey) EnrollmentKeyTableRow {
	return EnrollmentKeyTableRow{
		Name:         formatEnrollmentKeyName(k),
		KeyPrefix:    formatEnrollmentKeyPrefix(k.KeyPrefix),
		Mode:         formatEnrollmentKeyMode(k.Mode),
		Capabilities: formatEnrollmentKeyCapabilities(k.CapabilitySet),
		Uses:         formatEnrollmentKeyUses(k.Uses, k.MaxUses),
		Expires:      formatEnrollmentKeyExpires(k.ExpiresAt, k.RevokedAt),
		Environment:  formatEnrollmentKeyEnvironment(k.Environment),
		Created:      formatEnrollmentKeyTimestamp(k.CreatedAt),
		ID:           k.ID,
	}
}

func formatEnrollmentKeyName(k client.EnrollmentKey) string {
	if k.Name == "" {
		return "(unnamed)"
	}
	return k.Name
}

func formatEnrollmentKeyPrefix(prefix string) string {
	if prefix == "" {
		return "-"
	}
	return prefix + "..."
}

func formatEnrollmentKeyMode(mode client.EnrollmentKeyMode) string {
	if mode == "" {
		return string(client.EnrollmentKeyModeStandard)
	}
	return string(mode)
}

func formatEnrollmentKeyCapabilities(capabilities []string) string {
	if len(capabilities) == 0 {
		return "-"
	}
	return strings.Join(capabilities, ",")
}

func formatEnrollmentKeyUses(uses int, maxUses *int) string {
	if maxUses == nil {
		return fmt.Sprintf("%d/unlimited", uses)
	}
	return fmt.Sprintf("%d/%d", uses, *maxUses)
}

func formatEnrollmentKeyExpires(expiresAt, revokedAt *string) string {
	if revokedAt != nil {
		return "revoked"
	}
	if expiresAt == nil {
		return "never"
	}
	return formatEnrollmentKeyTimestamp(*expiresAt)
}

func formatEnrollmentKeyEnvironment(environment string) string {
	if environment == "" {
		return "-"
	}
	return environment
}

// formatEnrollmentKeyTimestamp renders an ISO 8601 timestamp string as
// YYYY-MM-DD, falling back to the raw string if it doesn't parse (the
// EnrollmentKey SDK type carries timestamps as strings, not time.Time — see
// client/enrollment_keys.go).
func formatEnrollmentKeyTimestamp(ts string) string {
	if ts == "" {
		return "-"
	}
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t.Format("2006-01-02")
	}
	return ts
}

// PrintEnrollmentKeys is a convenience function that formats and prints
// enrollment keys using the CLI's configured output format. It handles wide
// mode automatically based on configuration.
func PrintEnrollmentKeys(keys []client.EnrollmentKey) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewEnrollmentKeyTableFormatter(colorMode, isWide)
	rows := formatter.FormatEnrollmentKeys(keys)

	return printer.Print(rows)
}
