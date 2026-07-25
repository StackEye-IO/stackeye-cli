// Package output provides CLI output helpers for StackEye commands.
// Task stackeye-5859.
package output

import "github.com/StackEye-IO/stackeye-go-sdk/client"

// DeviceTagTableRow represents a row in the device tag table output.
// Task stackeye-5859.
type DeviceTagTableRow struct {
	Key   string `table:"KEY"`
	Value string `table:"VALUE"`
}

// FormatDeviceTags converts a slice of device tags into table-displayable rows.
// Task stackeye-5859.
func FormatDeviceTags(tags []client.DeviceTag) []DeviceTagTableRow {
	rows := make([]DeviceTagTableRow, 0, len(tags))
	for _, t := range tags {
		value := "(none)"
		if t.TagValue != nil && *t.TagValue != "" {
			value = *t.TagValue
		}
		rows = append(rows, DeviceTagTableRow{
			Key:   t.TagKey,
			Value: value,
		})
	}
	return rows
}

// PrintDeviceTags is a convenience function that formats and prints device tags.
// Task stackeye-5859.
func PrintDeviceTags(tags []client.DeviceTag) error {
	if len(tags) == 0 {
		return PrintEmpty("No tags assigned to this device.")
	}

	printer := getPrinter()
	rows := FormatDeviceTags(tags)
	return printer.Print(rows)
}

// DeviceRegionTableRow represents a row in the device region table output.
// Task stackeye-5859.
type DeviceRegionTableRow struct {
	RegionID string `table:"REGION"`
}

// FormatDeviceRegions converts a slice of device region assignments into
// table-displayable rows. Task stackeye-5859.
func FormatDeviceRegions(regions []client.DeviceRegion) []DeviceRegionTableRow {
	rows := make([]DeviceRegionTableRow, 0, len(regions))
	for _, r := range regions {
		rows = append(rows, DeviceRegionTableRow{RegionID: r.RegionID})
	}
	return rows
}

// PrintDeviceRegions is a convenience function that formats and prints
// device region assignments. Task stackeye-5859.
func PrintDeviceRegions(regions []client.DeviceRegion) error {
	if len(regions) == 0 {
		return PrintEmpty("No regions assigned to this device.")
	}

	printer := getPrinter()
	rows := FormatDeviceRegions(regions)
	return printer.Print(rows)
}
