// Package cmd implements the CLI commands for StackEye.
// Task #8068
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// probeLabelTimeout is the maximum time to wait for the API response.
const probeLabelTimeout = 30 * time.Second

// NewProbeLabelCmd creates and returns the probe label subcommand.
// Task #8068
func NewProbeLabelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "label <probe-id> <labels...>",
		Short:             "Add labels to a probe",
		ValidArgsFunction: ProbeCompletion(),
		Long: `Add or update labels on a probe.

Labels are added/merged with existing labels. If a label key already exists on
the probe, its value is updated. Label keys are auto-created if they don't
exist in your organization.

Labels can be specified in two formats:
  key=value    Key-value label (e.g., env=production)
  key          Key-only label with no value (e.g., pci)

The probe can be specified by UUID or by name. If the name matches multiple
probes, you'll be prompted to use the UUID instead.

Examples:
  # Add environment and tier labels
  stackeye probe label api-health env=production tier=backend

  # Add a key-only label (no value)
  stackeye probe label api-health pci compliant

  # Mix key-value and key-only labels
  stackeye probe label api-health env=staging internal

  # Use probe UUID
  stackeye probe label 550e8400-e29b-41d4-a716-446655440000 env=dev

  # Update an existing label value
  stackeye probe label api-health env=production  # changes env from staging to production`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeLabel(cmd.Context(), args[0], args[1:])
		},
	}

	return cmd
}

// runProbeLabel executes the probe label command logic.
// Task #8068
func runProbeLabel(ctx context.Context, probeIDArg string, labelArgs []string) error {
	// Parse label arguments
	labels, err := parseLabelArgs(labelArgs)
	if err != nil {
		return err
	}

	// Dry-run check: print what would happen and exit without making API calls
	if GetDryRun() {
		labelStrs := make([]string, len(labels))
		for i, l := range labels {
			if l.Value != nil {
				labelStrs[i] = fmt.Sprintf("%s=%s", l.Key, *l.Value)
			} else {
				labelStrs[i] = l.Key
			}
		}
		dryrun.PrintAction("add labels to", "probe",
			"Probe", probeIDArg,
			"Labels", strings.Join(labelStrs, ", "),
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Resolve probe ID (accepts UUID or name)
	probeID, err := ResolveProbeID(ctx, apiClient, probeIDArg)
	if err != nil {
		return err
	}

	// Call SDK to add labels with timeout
	reqCtx, cancel := context.WithTimeout(ctx, probeLabelTimeout)
	defer cancel()

	result, err := client.AddProbeLabels(reqCtx, apiClient, probeID, labels)
	if err != nil {
		return fmt.Errorf("failed to add labels: %w", err)
	}

	// Print the updated labels
	return output.PrintProbeLabels(result.Labels)
}

// parseLabelArgs parses command-line label arguments into ProbeLabelInput structs.
// Accepts formats: "key=value" and "key" (key-only with no value).
// Task #8068
func parseLabelArgs(args []string) ([]client.ProbeLabelInput, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("at least one label is required")
	}

	labels := make([]client.ProbeLabelInput, 0, len(args))

	for _, arg := range args {
		label, err := parseSingleLabel(arg)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}

	return labels, nil
}

// parseSingleLabel parses a single label argument.
// Returns a ProbeLabelInput with Key always set, and Value set only for key=value format.
// Task #8068
func parseSingleLabel(arg string) (client.ProbeLabelInput, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return client.ProbeLabelInput{}, fmt.Errorf("empty label argument")
	}

	// Check for key=value format
	if idx := strings.Index(arg, "="); idx > 0 {
		key := arg[:idx]
		value := arg[idx+1:]

		if err := validateLabelKey(key); err != nil {
			return client.ProbeLabelInput{}, err
		}

		if value != "" {
			if err := validateLabelValue(value); err != nil {
				return client.ProbeLabelInput{}, err
			}
		}

		return client.ProbeLabelInput{
			Key:   key,
			Value: &value,
		}, nil
	}

	// Key-only format (no value)
	if err := validateLabelKey(arg); err != nil {
		return client.ProbeLabelInput{}, err
	}

	return client.ProbeLabelInput{
		Key: arg,
	}, nil
}

// validateLabelValue validates a label value follows Kubernetes conventions.
// Task #8068
func validateLabelValue(value string) error {
	if len(value) > 63 {
		return fmt.Errorf("label value %q exceeds maximum length of 63 characters", value)
	}

	// Must be alphanumeric with hyphens, underscores, and dots
	for i, r := range value {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '-' || r == '_' || r == '.' {
			continue
		}
		return fmt.Errorf("label value %q contains invalid character %q at position %d; must be alphanumeric, hyphens, underscores, or dots", value, string(r), i)
	}

	return nil
}
