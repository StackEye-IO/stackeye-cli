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
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// incidentCreateTimeout is the maximum time to wait for the API response.
const incidentCreateTimeout = 30 * time.Second

// validIncidentStatuses lists the valid incident status values.
var validIncidentStatuses = []string{"investigating", "identified", "monitoring", "resolved"}

// validIncidentImpacts lists the valid incident impact values.
var validIncidentImpacts = []string{"none", "minor", "major", "critical"}

// incidentCreateFlags holds the flag values for the incident create command.
type incidentCreateFlags struct {
	statusPageID uint
	title        string
	message      string
	status       string
	impact       string
	fromFile     string
}

// incidentYAMLConfig represents the YAML structure for --from-file input.
type incidentYAMLConfig struct {
	Title   string `yaml:"title"`
	Message string `yaml:"message"`
	Status  string `yaml:"status,omitempty"`
	Impact  string `yaml:"impact"`
}

// NewIncidentCreateCmd creates and returns the incident create subcommand.
func NewIncidentCreateCmd() *cobra.Command {
	flags := &incidentCreateFlags{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new incident for a status page",
		Long: `Create a new incident to communicate service disruptions via your status page.

Incidents inform customers about ongoing issues and your response. When an issue
occurs, create an incident with an appropriate impact level and status, then
update it as you investigate and resolve the problem.

Required Flags:
  --status-page-id   ID of the status page (required)
  --title            Incident title (required unless using --from-file)
  --message          Detailed incident message/description (required unless using --from-file)
  --impact           Impact level (required unless using --from-file)

Optional Flags:
  --status           Initial status (default: investigating)
  --from-file        Create incident from YAML file

Incident Status Workflow:
  investigating → identified → monitoring → resolved

Impact Levels:
  none     - No impact to services (informational)
  minor    - Minor performance degradation
  major    - Significant service degradation
  critical - Complete service outage

Examples:
  # Create a basic incident
  stackeye incident create --status-page-id 123 \
    --title "API Degradation" \
    --message "We are investigating reports of increased latency" \
    --impact minor

  # Create with scheduled maintenance
  stackeye incident create --status-page-id 123 \
    --title "Database Maintenance" \
    --message "Scheduled maintenance window for database upgrades" \
    --impact none

  # Create with specific status
  stackeye incident create --status-page-id 123 \
    --title "Service Outage" \
    --message "Root cause identified, working on fix" \
    --impact critical \
    --status identified

  # Create from YAML file
  stackeye incident create --status-page-id 123 --from-file incident.yaml

  # Output as JSON for scripting
  stackeye incident create --status-page-id 123 \
    --title "Issue" \
    --message "Investigating reported issues" \
    --impact minor \
    -o json

YAML File Format:
  title: "Database Connectivity Issues"
  message: "Users may experience intermittent connection errors"
  status: "investigating"
  impact: "major"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIncidentCreate(cmd.Context(), flags)
		},
	}

	// Required flags
	cmd.Flags().UintVar(&flags.statusPageID, "status-page-id", 0, "status page ID (required)")
	cmd.Flags().StringVar(&flags.title, "title", "", "incident title (required unless using --from-file)")
	cmd.Flags().StringVar(&flags.impact, "impact", "", "impact level: none, minor, major, critical (required unless using --from-file)")

	// Required unless using --from-file
	cmd.Flags().StringVar(&flags.message, "message", "", "detailed incident message/description (required unless using --from-file)")
	cmd.Flags().StringVar(&flags.status, "status", "investigating", "initial status: investigating, identified, monitoring, resolved")
	cmd.Flags().StringVar(&flags.fromFile, "from-file", "", "create incident from YAML file")

	// Mark required flags
	_ = cmd.MarkFlagRequired("status-page-id")

	return cmd
}

// runIncidentCreate executes the incident create command logic.
func runIncidentCreate(ctx context.Context, flags *incidentCreateFlags) error {
	var req *client.CreateIncidentRequest
	var err error

	// Handle --from-file if provided
	if flags.fromFile != "" {
		req, err = buildIncidentRequestFromYAML(flags.fromFile)
		if err != nil {
			return err
		}
	} else {
		// Build request from flags
		req, err = buildIncidentRequestFromFlags(flags)
		if err != nil {
			return err
		}
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to create incident with timeout
	reqCtx, cancel := context.WithTimeout(ctx, incidentCreateTimeout)
	defer cancel()

	incident, err := client.CreateIncident(reqCtx, apiClient, flags.statusPageID, req)
	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	// Print the created incident using the table formatter
	return output.PrintIncident(*incident)
}

// buildIncidentRequestFromFlags constructs the API request from command flags.
func buildIncidentRequestFromFlags(flags *incidentCreateFlags) (*client.CreateIncidentRequest, error) {
	// Validate required fields
	if flags.title == "" {
		return nil, fmt.Errorf("--title is required")
	}

	if flags.message == "" {
		return nil, fmt.Errorf("--message is required")
	}

	if flags.impact == "" {
		return nil, fmt.Errorf("--impact is required")
	}

	// Validate impact value
	impactLower := strings.ToLower(flags.impact)
	if !slices.Contains(validIncidentImpacts, impactLower) {
		return nil, fmt.Errorf("invalid impact %q: must be one of %v", flags.impact, validIncidentImpacts)
	}

	// Validate status value
	statusLower := strings.ToLower(flags.status)
	if !slices.Contains(validIncidentStatuses, statusLower) {
		return nil, fmt.Errorf("invalid status %q: must be one of %v", flags.status, validIncidentStatuses)
	}

	req := &client.CreateIncidentRequest{
		Title:   flags.title,
		Message: flags.message,
		Status:  statusLower,
		Impact:  impactLower,
	}

	return req, nil
}

// buildIncidentRequestFromYAML constructs the API request from a YAML file.
func buildIncidentRequestFromYAML(filePath string) (*client.CreateIncidentRequest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	var cfg incidentYAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if cfg.Title == "" {
		return nil, fmt.Errorf("YAML file must contain 'title' field")
	}

	if cfg.Message == "" {
		return nil, fmt.Errorf("YAML file must contain 'message' field")
	}

	if cfg.Impact == "" {
		return nil, fmt.Errorf("YAML file must contain 'impact' field")
	}

	// Validate impact value
	impactLower := strings.ToLower(cfg.Impact)
	if !slices.Contains(validIncidentImpacts, impactLower) {
		return nil, fmt.Errorf("invalid impact %q: must be one of %v", cfg.Impact, validIncidentImpacts)
	}

	// Default status to investigating if not provided
	status := cfg.Status
	if status == "" {
		status = "investigating"
	}

	// Validate status value
	statusLower := strings.ToLower(status)
	if !slices.Contains(validIncidentStatuses, statusLower) {
		return nil, fmt.Errorf("invalid status %q: must be one of %v", status, validIncidentStatuses)
	}

	req := &client.CreateIncidentRequest{
		Title:   cfg.Title,
		Message: cfg.Message,
		Status:  statusLower,
		Impact:  impactLower,
	}

	return req, nil
}
