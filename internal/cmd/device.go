// Package cmd implements the CLI commands for StackEye.
// Task stackeye-5859.
package cmd

import "github.com/spf13/cobra"

// NewDeviceCmd creates and returns the device parent command.
// This command provides management of device tag and region associations.
// Task stackeye-5859.
func NewDeviceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device",
		Short: "Manage device tag and region associations",
		Long: `Manage device tag and region associations.

Devices are monitored hosts (servers, appliances, etc.) that can be tagged
and assigned to monitoring regions, same as the region assignments used for
probes. These associations are the same scope mechanism the metric-alert
evaluator resolves when an alert's scope is set to a tag or region.

The device can be specified by UUID or by exact name match.

Available Commands:
  tag              Attach a tag to a device
  tag list         List a device's tags
  untag            Remove a tag from a device
  region assign    Assign a device to a region
  region list      List a device's region assignments
  region unassign   Remove a device's region assignment

Examples:
  # Attach tags to a device
  stackeye device tag web-01 env=production tier=backend

  # List a device's tags
  stackeye device tag list web-01

  # Remove a tag from a device
  stackeye device untag web-01 env

  # Assign a device to a region
  stackeye device region assign web-01 us-east

  # List a device's region assignments
  stackeye device region list web-01

  # Remove a device's region assignment
  stackeye device region unassign web-01 us-east

For more information about a specific command:
  stackeye device [command] --help`,
	}

	// Register subcommands
	cmd.AddCommand(NewDeviceTagCmd())
	cmd.AddCommand(NewDeviceUntagCmd())
	cmd.AddCommand(NewDeviceRegionCmd())

	return cmd
}

// NewDeviceRegionCmd creates and returns the device region parent command.
// Task stackeye-5859.
func NewDeviceRegionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "region",
		Short: "Manage a device's region assignments",
		Long: `Manage a device's explicit region assignments.

A device can be assigned to one or more monitoring regions. The region must
be a public region or one owned by your organization.

Available Commands:
  assign      Assign a device to a region
  list        List a device's region assignments
  unassign    Remove a device's region assignment

For more information about a specific command:
  stackeye device region [command] --help`,
	}

	cmd.AddCommand(NewDeviceRegionAssignCmd())
	cmd.AddCommand(NewDeviceRegionListCmd())
	cmd.AddCommand(NewDeviceRegionUnassignCmd())

	return cmd
}
