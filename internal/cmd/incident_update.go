// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// incidentUpdateTimeout is the maximum time to wait for the API response.
const incidentUpdateTimeout = 30 * time.Second

// incidentUpdateFlags holds the flag values for the incident update command.
type incidentUpdateFlags struct {
	statusPageID uint
	incidentID   uint
	title        string
	message      string
	status       string
	impact       string
	fromFile     string
}

// incidentUpdateYAMLConfig represents the YAML structure for --from-file input.
type incidentUpdateYAMLConfig struct {
	Title   string `yaml:"title,omitempty"`
	Message string `yaml:"message,omitempty"`
	Status  string `yaml:"status,omitempty"`
	Impact  string `yaml:"impact,omitempty"`
}

// NewIncidentUpdateCmd creates and returns the incident update subcommand.
func NewIncidentUpdateCmd() *cobra.Command {
	flags := &incidentUpdateFlags{}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing incident on a status page",
		Long: `Update an existing incident to communicate status changes to customers.

Update incidents as you investigate and work to resolve issues. Customers
following your status page will see these updates in real-time.

Required Flags:
  --status-page-id   ID of the status page (required)
  --incident-id      ID of the incident to update (required)

Optional Flags (at least one required):
  --title            New incident title
  --message          Updated incident message/description
  --status           New status: investigating, identified, monitoring, resolved
  --impact           New impact level: none, minor, major, critical
  --from-file        Update incident from YAML file

Incident Status Workflow:
  investigating -> identified -> monitoring -> resolved

Impact Levels:
  none     - No impact to services (informational)
  minor    - Minor performance degradation
  major    - Significant service degradation
  critical - Complete service outage

Examples:
  # Update incident status to identified
  stackeye incident update --status-page-id 123 --incident-id 456 --status identified

  # Update with a new message
  stackeye incident update --status-page-id 123 --incident-id 456 \
    --message "Root cause identified. Database connection pool exhausted."

  # Update multiple fields
  stackeye incident update --status-page-id 123 --incident-id 456 \
    --status monitoring \
    --message "Fix deployed. Monitoring for stability."

  # Update from YAML file
  stackeye incident update --status-page-id 123 --incident-id 456 --from-file update.yaml

  # Output as JSON for scripting
  stackeye incident update --status-page-id 123 --incident-id 456 --status resolved -o json

YAML File Format:
  title: "Updated Incident Title"
  message: "New status update message"
  status: "monitoring"
  impact: "minor"

Note: All fields in the YAML file are optional. Only specified fields will be updated.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIncidentUpdate(cmd.Context(), flags)
		},
	}

	// Required flags
	cmd.Flags().UintVar(&flags.statusPageID, "status-page-id", 0, "status page ID (required)")
	cmd.Flags().UintVar(&flags.incidentID, "incident-id", 0, "incident ID to update (required)")

	// Optional update flags
	cmd.Flags().StringVar(&flags.title, "title", "", "new incident title")
	cmd.Flags().StringVar(&flags.message, "message", "", "updated incident message/description")
	cmd.Flags().StringVar(&flags.status, "status", "", "new status: investigating, identified, monitoring, resolved")
	cmd.Flags().StringVar(&flags.impact, "impact", "", "new impact level: none, minor, major, critical")
	cmd.Flags().StringVar(&flags.fromFile, "from-file", "", "update incident from YAML file")

	// Mark required flags
	_ = cmd.MarkFlagRequired("status-page-id")
	_ = cmd.MarkFlagRequired("incident-id")

	return cmd
}

// runIncidentUpdate executes the incident update command logic.
func runIncidentUpdate(ctx context.Context, flags *incidentUpdateFlags) error {
	var req *client.UpdateIncidentRequest
	var err error

	// Handle --from-file if provided
	if flags.fromFile != "" {
		req, err = buildIncidentUpdateRequestFromYAML(flags.fromFile)
		if err != nil {
			return err
		}
	} else {
		// Build request from flags
		req, err = buildIncidentUpdateRequestFromFlags(flags)
		if err != nil {
			return err
		}
	}

	// Verify at least one field is being updated
	if req.Title == nil && req.Message == nil && req.Status == nil && req.Impact == nil {
		return fmt.Errorf("at least one update field is required: --title, --message, --status, --impact, or --from-file")
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		details := []string{
			"Status Page ID", fmt.Sprintf("%d", flags.statusPageID),
			"Incident ID", fmt.Sprintf("%d", flags.incidentID),
		}
		if req.Title != nil {
			details = append(details, "Title", *req.Title)
		}
		if req.Status != nil {
			details = append(details, "Status", *req.Status)
		}
		if req.Impact != nil {
			details = append(details, "Impact", *req.Impact)
		}
		dryrun.PrintAction("update", "incident", details...)
		return nil
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to update incident with timeout
	reqCtx, cancel := context.WithTimeout(ctx, incidentUpdateTimeout)
	defer cancel()

	incident, err := client.UpdateIncident(reqCtx, apiClient, flags.statusPageID, flags.incidentID, req)
	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	// Print the updated incident using the table formatter
	return output.PrintIncident(*incident)
}

// buildIncidentUpdateRequestFromFlags constructs the API request from command flags.
func buildIncidentUpdateRequestFromFlags(flags *incidentUpdateFlags) (*client.UpdateIncidentRequest, error) {
	req := &client.UpdateIncidentRequest{}

	// Set title if provided
	if flags.title != "" {
		req.Title = &flags.title
	}

	// Set message if provided
	if flags.message != "" {
		req.Message = &flags.message
	}

	// Validate and set status if provided
	if flags.status != "" {
		statusLower := strings.ToLower(flags.status)
		if !slices.Contains(validIncidentStatuses, statusLower) {
			return nil, fmt.Errorf("invalid status %q: must be one of %v", flags.status, validIncidentStatuses)
		}
		req.Status = &statusLower
	}

	// Validate and set impact if provided
	if flags.impact != "" {
		impactLower := strings.ToLower(flags.impact)
		if !slices.Contains(validIncidentImpacts, impactLower) {
			return nil, fmt.Errorf("invalid impact %q: must be one of %v", flags.impact, validIncidentImpacts)
		}
		req.Impact = &impactLower
	}

	return req, nil
}

// buildIncidentUpdateRequestFromYAML constructs the API request from a YAML file.
func buildIncidentUpdateRequestFromYAML(filePath string) (*client.UpdateIncidentRequest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	var cfg incidentUpdateYAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	req := &client.UpdateIncidentRequest{}

	// Set title if provided
	if cfg.Title != "" {
		req.Title = &cfg.Title
	}

	// Set message if provided
	if cfg.Message != "" {
		req.Message = &cfg.Message
	}

	// Validate and set status if provided
	if cfg.Status != "" {
		statusLower := strings.ToLower(cfg.Status)
		if !slices.Contains(validIncidentStatuses, statusLower) {
			return nil, fmt.Errorf("invalid status %q: must be one of %v", cfg.Status, validIncidentStatuses)
		}
		req.Status = &statusLower
	}

	// Validate and set impact if provided
	if cfg.Impact != "" {
		impactLower := strings.ToLower(cfg.Impact)
		if !slices.Contains(validIncidentImpacts, impactLower) {
			return nil, fmt.Errorf("invalid impact %q: must be one of %v", cfg.Impact, validIncidentImpacts)
		}
		req.Impact = &impactLower
	}

	return req, nil
}
