// Package output provides CLI output helpers for StackEye commands.
// Task stackeye-5859.
package output

import (
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

func TestFormatDeviceTags(t *testing.T) {
	value := "production"
	tags := []client.DeviceTag{
		{TagKey: "env", TagValue: &value},
		{TagKey: "critical"},
	}

	rows := FormatDeviceTags(tags)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].Key != "env" || rows[0].Value != "production" {
		t.Errorf("unexpected row 0: %+v", rows[0])
	}
	if rows[1].Key != "critical" || rows[1].Value != "(none)" {
		t.Errorf("unexpected row 1: %+v", rows[1])
	}
}

func TestPrintDeviceTags_Empty(t *testing.T) {
	if err := PrintDeviceTags(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrintDeviceTags_NonEmpty(t *testing.T) {
	tags := []client.DeviceTag{{TagKey: "env", TagValue: nil}}
	if err := PrintDeviceTags(tags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormatDeviceRegions(t *testing.T) {
	regions := []client.DeviceRegion{
		{RegionID: "us-east"},
		{RegionID: "eu-west"},
	}

	rows := FormatDeviceRegions(regions)
	if len(rows) != 2 || rows[0].RegionID != "us-east" || rows[1].RegionID != "eu-west" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestPrintDeviceRegions_Empty(t *testing.T) {
	if err := PrintDeviceRegions(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrintDeviceRegions_NonEmpty(t *testing.T) {
	regions := []client.DeviceRegion{{RegionID: "us-east"}}
	if err := PrintDeviceRegions(regions); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
