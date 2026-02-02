// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewLabelCmd creates and returns the label parent command.
// This command provides management of organization label keys.
// Task #8064
func NewLabelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label",
		Short: "Manage probe label keys",
		Long: `Manage probe label keys for your organization.

Labels allow you to organize and filter probes using key-value pairs. Label keys
are defined at the organization level, and each probe can have labels assigned
to it using those keys.

Label keys follow Kubernetes naming conventions:
  - Lowercase alphanumeric characters and hyphens only
  - Must start with a letter
  - Maximum 63 characters

Common label key examples:
  env         Environment classification (production, staging, dev)
  tier        Service tier (frontend, backend, database)
  team        Team ownership (platform, infra, app)
  region      Geographic region (us-east, eu-west, ap-south)
  service     Service or application name

Available Commands:
  list        List all label keys in your organization
  create      Create a new label key
  delete      Delete a label key (removes from all probes)

Examples:
  # List all label keys
  stackeye label list

  # Create a new label key for environment
  stackeye label create env --display-name "Environment" --color "#10B981"

  # Create a label key with description
  stackeye label create tier --display-name "Service Tier" \
    --description "Frontend, backend, or database tier classification"

  # Delete a label key
  stackeye label delete env

  # Delete without confirmation prompt
  stackeye label delete env --yes

For more information about a specific command:
  stackeye label [command] --help

See also:
  stackeye probe label     - Assign labels to a probe
  stackeye probe unlabel   - Remove labels from a probe
  stackeye probe list --labels env=production  - Filter probes by labels`,
		Aliases: []string{"labels", "lbl"},
	}

	// Register subcommands
	cmd.AddCommand(NewLabelListCmd())
	// Note: Additional subcommands will be added in subsequent tasks:
	// - Task #8066: label_create.go
	// - Task #8067: label_delete.go

	return cmd
}
