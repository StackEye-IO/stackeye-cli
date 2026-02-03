// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// APIKeyTableRow represents a row in the API key table output.
// The struct tags control column headers and wide mode display.
type APIKeyTableRow struct {
	Name        string `table:"NAME"`
	KeyPrefix   string `table:"PREFIX"`
	Permissions string `table:"PERMISSIONS"`
	LastUsed    string `table:"LAST USED"`
	// Wide mode columns
	Expires string `table:"EXPIRES,wide"`
	Created string `table:"CREATED,wide"`
	ID      string `table:"ID,wide"`
}

// APIKeyTableFormatter converts SDK APIKey types to table-displayable rows.
type APIKeyTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
	isWide   bool
}

// NewAPIKeyTableFormatter creates a new formatter for API key table output.
// The colorMode parameter controls whether colors are applied.
// Set isWide to true for extended output with additional columns.
func NewAPIKeyTableFormatter(colorMode sdkoutput.ColorMode, isWide bool) *APIKeyTableFormatter {
	return &APIKeyTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
		isWide:   isWide,
	}
}

// FormatAPIKeys converts a slice of SDK APIKeys into table-displayable rows.
func (f *APIKeyTableFormatter) FormatAPIKeys(keys []client.APIKey) []APIKeyTableRow {
	rows := make([]APIKeyTableRow, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, f.formatAPIKey(k))
	}
	return rows
}

// FormatAPIKey converts a single SDK APIKey into a table-displayable row.
func (f *APIKeyTableFormatter) FormatAPIKey(key client.APIKey) APIKeyTableRow {
	return f.formatAPIKey(key)
}

// formatAPIKey is the internal conversion function.
func (f *APIKeyTableFormatter) formatAPIKey(k client.APIKey) APIKeyTableRow {
	return APIKeyTableRow{
		Name:        k.Name,
		KeyPrefix:   formatKeyPrefix(k.KeyPrefix),
		Permissions: formatPermissions(k.Permissions),
		LastUsed:    formatLastUsed(k.LastUsedAt),
		Expires:     formatExpires(k.ExpiresAt),
		Created:     formatCreatedTime(k.CreatedAt),
		ID:          k.ID.String(),
	}
}

// formatKeyPrefix formats the key prefix for display.
// Example: se_abc1 -> se_abc1...
func formatKeyPrefix(prefix string) string {
	if prefix == "" {
		return "-"
	}
	return prefix + "..."
}

// formatPermissions formats the permissions string for display.
// Truncates long permission strings for readability.
func formatPermissions(perms string) string {
	if perms == "" {
		return "-"
	}
	// Truncate if too long for table display
	if len(perms) > 30 {
		return perms[:27] + "..."
	}
	return perms
}

// formatLastUsed formats the last used timestamp for display.
// Returns "Never" if the key has never been used.
func formatLastUsed(lastUsed *time.Time) string {
	if lastUsed == nil {
		return "Never"
	}
	return lastUsed.Format("2006-01-02")
}

// formatExpires formats the expiration timestamp for display.
// Returns "Never" if the key does not expire.
func formatExpires(expires *time.Time) string {
	if expires == nil {
		return "Never"
	}
	return expires.Format("2006-01-02")
}

// PrintAPIKeys is a convenience function that formats and prints API keys
// using the CLI's configured output format. It handles wide mode
// automatically based on configuration.
func PrintAPIKeys(keys []client.APIKey) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	// Get color mode from config if available
	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewAPIKeyTableFormatter(colorMode, isWide)
	rows := formatter.FormatAPIKeys(keys)

	return printer.Print(rows)
}

// PrintAPIKey is a convenience function that formats and prints a single API key.
func PrintAPIKey(key client.APIKey) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto
	isWide := printer.Format() == sdkoutput.FormatWide

	if getter := loadConfigGetter(); getter != nil {
		if cfg := getter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewAPIKeyTableFormatter(colorMode, isWide)
	row := formatter.FormatAPIKey(key)

	return printer.Print(row)
}
