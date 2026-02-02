// Package output provides CLI output helpers for StackEye commands.
// Task #8065
package output

import (
	"strings"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// LabelKeyTableRow represents a row in the label key table output.
// The struct tags control column headers.
type LabelKeyTableRow struct {
	Key         string `table:"KEY"`
	DisplayName string `table:"DISPLAY NAME"`
	Color       string `table:"COLOR"`
	ValuesInUse string `table:"VALUES IN USE"`
	Probes      string `table:"PROBES"`
}

// LabelKeyTableFormatter converts SDK LabelKey types to table-displayable rows.
type LabelKeyTableFormatter struct {
	colorMgr *sdkoutput.ColorManager
}

// NewLabelKeyTableFormatter creates a new formatter for label key table output.
func NewLabelKeyTableFormatter(colorMode sdkoutput.ColorMode) *LabelKeyTableFormatter {
	return &LabelKeyTableFormatter{
		colorMgr: sdkoutput.NewColorManager(colorMode),
	}
}

// FormatLabelKeys converts a slice of SDK LabelKeys into table-displayable rows.
func (f *LabelKeyTableFormatter) FormatLabelKeys(labelKeys []client.LabelKey) []LabelKeyTableRow {
	rows := make([]LabelKeyTableRow, 0, len(labelKeys))
	for _, lk := range labelKeys {
		rows = append(rows, f.formatLabelKey(lk))
	}
	return rows
}

// FormatLabelKey converts a single SDK LabelKey into a table-displayable row.
func (f *LabelKeyTableFormatter) FormatLabelKey(labelKey client.LabelKey) LabelKeyTableRow {
	return f.formatLabelKey(labelKey)
}

// formatLabelKey is the internal conversion function.
func (f *LabelKeyTableFormatter) formatLabelKey(lk client.LabelKey) LabelKeyTableRow {
	// Format display name (use key if not set)
	displayName := lk.Key
	if lk.DisplayName != nil && *lk.DisplayName != "" {
		displayName = *lk.DisplayName
	}

	// Format values in use
	valuesInUse := "(key-only)"
	if len(lk.ValuesInUse) > 0 {
		valuesInUse = formatValues(lk.ValuesInUse)
	}

	// Format probe count
	probeCount := "0"
	if lk.ProbeCount > 0 {
		probeCount = formatInt64(lk.ProbeCount)
	}

	return LabelKeyTableRow{
		Key:         lk.Key,
		DisplayName: displayName,
		Color:       formatColorDot(lk.Color),
		ValuesInUse: valuesInUse,
		Probes:      probeCount,
	}
}

// formatValues formats a slice of values into a comma-separated string.
// Truncates if there are too many values.
func formatValues(values []string) string {
	if len(values) == 0 {
		return "(key-only)"
	}

	const maxDisplayValues = 3
	const maxTotalLength = 30

	if len(values) <= maxDisplayValues {
		result := strings.Join(values, ", ")
		if len(result) > maxTotalLength {
			return result[:maxTotalLength-3] + "..."
		}
		return result
	}

	// Show first 3 values with "+N more" indicator
	displayed := strings.Join(values[:maxDisplayValues], ", ")
	remaining := len(values) - maxDisplayValues
	return displayed + formatMoreCount(remaining)
}

// formatMoreCount formats the "+N more" suffix.
func formatMoreCount(count int) string {
	if count <= 0 {
		return ""
	}
	return " +" + formatInt64(int64(count))
}

// formatInt64 converts an int64 to a string.
func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	// Simple integer to string conversion
	result := ""
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	if negative {
		return "-" + result
	}
	return result
}

// formatColorDot returns a colored circle indicator for the hex color.
// In terminals that support color, this shows the actual color.
// In plain text, just shows a bullet point.
func formatColorDot(hexColor string) string {
	// For now, just use a bullet point - color rendering would need ANSI support
	return "\u25CF" // Unicode filled circle
}

// PrintLabelKeys is a convenience function that formats and prints label keys
// using the CLI's configured output format.
func PrintLabelKeys(labelKeys []client.LabelKey) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto

	// Get color mode from config if available
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewLabelKeyTableFormatter(colorMode)
	rows := formatter.FormatLabelKeys(labelKeys)

	return printer.Print(rows)
}

// PrintLabelKey is a convenience function that formats and prints a single label key.
func PrintLabelKey(labelKey client.LabelKey) error {
	printer := getPrinter()
	colorMode := sdkoutput.ColorAuto

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			colorMode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	formatter := NewLabelKeyTableFormatter(colorMode)
	row := formatter.FormatLabelKey(labelKey)

	return printer.Print(row)
}

// ProbeLabelTableRow represents a row in the probe label table output.
// Task #8068
type ProbeLabelTableRow struct {
	Key   string `table:"KEY"`
	Value string `table:"VALUE"`
}

// FormatProbeLabels converts a slice of probe labels into table-displayable rows.
// Task #8068
func FormatProbeLabels(labels []client.ProbeLabel) []ProbeLabelTableRow {
	rows := make([]ProbeLabelTableRow, 0, len(labels))
	for _, l := range labels {
		value := "(none)"
		if l.Value != nil && *l.Value != "" {
			value = *l.Value
		}
		rows = append(rows, ProbeLabelTableRow{
			Key:   l.Key,
			Value: value,
		})
	}
	return rows
}

// PrintProbeLabels is a convenience function that formats and prints probe labels.
// Task #8068
func PrintProbeLabels(labels []client.ProbeLabel) error {
	if len(labels) == 0 {
		return PrintEmpty("No labels assigned to this probe.")
	}

	printer := getPrinter()
	rows := FormatProbeLabels(labels)
	return printer.Print(rows)
}
