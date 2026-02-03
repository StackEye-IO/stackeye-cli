// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// probeExportTimeout is the maximum time to wait for the API response.
const probeExportTimeout = 60 * time.Second

// probeExportFlags holds the flag values for the probe export command.
type probeExportFlags struct {
	format   string
	file     string
	probeIDs string
	status   string
	labels   string
}

// probeExportConfig represents a portable probe configuration for export.
// This format is designed for backup and migration between environments.
type probeExportConfig struct {
	Name                   string             `json:"name" yaml:"name"`
	URL                    string             `json:"url" yaml:"url"`
	CheckType              string             `json:"check_type" yaml:"check_type"`
	Method                 string             `json:"method" yaml:"method"`
	Headers                map[string]string  `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body                   *string            `json:"body,omitempty" yaml:"body,omitempty"`
	TimeoutMs              int                `json:"timeout_ms" yaml:"timeout_ms"`
	IntervalSeconds        int                `json:"interval_seconds" yaml:"interval_seconds"`
	Regions                []string           `json:"regions,omitempty" yaml:"regions,omitempty"`
	ExpectedStatusCodes    []int              `json:"expected_status_codes" yaml:"expected_status_codes"`
	KeywordCheck           *string            `json:"keyword_check,omitempty" yaml:"keyword_check,omitempty"`
	KeywordCheckType       *string            `json:"keyword_check_type,omitempty" yaml:"keyword_check_type,omitempty"`
	JSONPathCheck          *string            `json:"json_path_check,omitempty" yaml:"json_path_check,omitempty"`
	JSONPathExpected       *string            `json:"json_path_expected,omitempty" yaml:"json_path_expected,omitempty"`
	SSLCheckEnabled        bool               `json:"ssl_check_enabled" yaml:"ssl_check_enabled"`
	SSLExpiryThresholdDays int                `json:"ssl_expiry_threshold_days" yaml:"ssl_expiry_threshold_days"`
	FollowRedirects        bool               `json:"follow_redirects" yaml:"follow_redirects"`
	MaxRedirects           int                `json:"max_redirects" yaml:"max_redirects"`
	AlertChannelIDs        []string           `json:"alert_channel_ids,omitempty" yaml:"alert_channel_ids,omitempty"`
	Labels                 []probeExportLabel `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// probeExportLabel represents a label in the export format.
type probeExportLabel struct {
	Key   string  `json:"key" yaml:"key"`
	Value *string `json:"value,omitempty" yaml:"value,omitempty"`
}

// NewProbeExportCmd creates and returns the probe export subcommand.
func NewProbeExportCmd() *cobra.Command {
	flags := &probeExportFlags{}

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export probe configurations for backup or migration",
		Long: `Export probe configurations as JSON or YAML for backup or migration.

Exports probe settings including check type, intervals, expected status codes,
content validation rules, SSL settings, monitoring regions, and linked alert
channel IDs. The exported format can be used to recreate probes in another
environment.

By default, exports all probes to stdout in YAML format. Use --probe-ids to
export specific probes, or --status/--labels to filter.

Output Formats:
  yaml    Human-readable YAML (default)
  json    Machine-readable JSON

Examples:
  # Export all probes as YAML to stdout
  stackeye probe export

  # Export all probes as JSON
  stackeye probe export --format json

  # Export to a file
  stackeye probe export --file probes.yaml

  # Export specific probes by ID
  stackeye probe export --probe-ids {probe-uuid-1},{probe-uuid-2}

  # Export only probes that are down
  stackeye probe export --status down

  # Export probes with specific labels
  stackeye probe export --labels "env=production"

  # Export as JSON to a file
  stackeye probe export --format json --file backup.json`,
		Aliases: []string{"exp"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeExport(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVarP(&flags.format, "format", "f", "yaml", "output format: yaml, json")
	cmd.Flags().StringVar(&flags.file, "file", "", "write output to file instead of stdout")
	cmd.Flags().StringVar(&flags.probeIDs, "probe-ids", "", "comma-separated probe IDs to export")
	cmd.Flags().StringVarP(&flags.status, "status", "s", "", "filter by status: up, down, degraded, paused, pending")
	cmd.Flags().StringVarP(&flags.labels, "labels", "l", "", "filter by labels: key=value,key2=value2 (AND logic)")

	return cmd
}

// runProbeExport executes the probe export command logic.
func runProbeExport(ctx context.Context, flags *probeExportFlags) error {
	// Validate format
	format := strings.ToLower(flags.format)
	if format != "yaml" && format != "json" {
		return clierrors.InvalidValueError("--format", flags.format, clierrors.ValidExportFormats)
	}

	// Validate status filter
	var probeStatus client.ProbeStatus
	if flags.status != "" {
		switch flags.status {
		case "up":
			probeStatus = client.ProbeStatusUp
		case "down":
			probeStatus = client.ProbeStatusDown
		case "degraded":
			probeStatus = client.ProbeStatusDegraded
		case "paused":
			probeStatus = client.ProbeStatusPaused
		case "pending":
			probeStatus = client.ProbeStatusPending
		default:
			return clierrors.InvalidValueError("--status", flags.status, clierrors.ValidProbeStatusFilters)
		}
	}

	// Parse label filters
	labelFilters, err := parseLabelFilters(flags.labels)
	if err != nil {
		return err
	}

	// Parse probe IDs if provided
	var probeIDs []uuid.UUID
	if flags.probeIDs != "" {
		parts := strings.Split(flags.probeIDs, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := uuid.Parse(part)
			if err != nil {
				return fmt.Errorf("invalid probe ID %q: %w", part, err)
			}
			probeIDs = append(probeIDs, id)
		}
		if len(probeIDs) == 0 {
			return fmt.Errorf("--probe-ids provided but no valid IDs found")
		}
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Fetch probes
	var probes []client.Probe
	reqCtx, cancel := context.WithTimeout(ctx, probeExportTimeout)
	defer cancel()

	if len(probeIDs) > 0 {
		// Fetch specific probes by ID
		probes, err = fetchProbesByIDs(reqCtx, apiClient, probeIDs)
		if err != nil {
			return err
		}
	} else {
		// Fetch all probes with optional filters
		probes, err = fetchAllProbesForExport(reqCtx, apiClient, probeStatus, labelFilters)
		if err != nil {
			return err
		}
	}

	if len(probes) == 0 {
		return fmt.Errorf("no probes found matching the specified criteria")
	}

	// Convert to export format
	configs := make([]probeExportConfig, 0, len(probes))
	for i := range probes {
		configs = append(configs, convertProbeToExportConfig(&probes[i]))
	}

	// Marshal output
	var data []byte
	switch format {
	case "json":
		data, err = json.MarshalIndent(configs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		data = append(data, '\n')
	case "yaml":
		data, err = yaml.Marshal(configs)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
	}

	// Write output
	if flags.file != "" {
		if err := os.WriteFile(flags.file, data, 0o600); err != nil {
			return fmt.Errorf("failed to write file %q: %w", flags.file, err)
		}
		fmt.Fprintf(os.Stderr, "Exported %d probe(s) to %s\n", len(configs), flags.file)
		return nil
	}

	_, err = os.Stdout.Write(data)
	return err
}

// fetchProbesByIDs fetches specific probes by their UUIDs.
func fetchProbesByIDs(ctx context.Context, apiClient *client.Client, ids []uuid.UUID) ([]client.Probe, error) {
	probes := make([]client.Probe, 0, len(ids))
	for _, id := range ids {
		probe, err := client.GetProbe(ctx, apiClient, id, "")
		if err != nil {
			return nil, fmt.Errorf("failed to get probe %s: %w", id, err)
		}
		probes = append(probes, *probe)
	}
	return probes, nil
}

// fetchAllProbesForExport fetches all probes, paginating through results.
func fetchAllProbesForExport(ctx context.Context, apiClient *client.Client, status client.ProbeStatus, labels map[string]string) ([]client.Probe, error) {
	var allProbes []client.Probe
	page := 1
	limit := 100

	for {
		opts := &client.ListProbesOptions{
			Page:   page,
			Limit:  limit,
			Status: status,
			Labels: labels,
		}

		result, err := client.ListProbes(ctx, apiClient, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list probes: %w", err)
		}

		allProbes = append(allProbes, result.Probes...)

		// If we got fewer results than the limit, we've reached the last page
		if len(result.Probes) < limit {
			break
		}
		page++
	}

	return allProbes, nil
}

// convertProbeToExportConfig converts a Probe to the portable export format.
func convertProbeToExportConfig(p *client.Probe) probeExportConfig {
	cfg := probeExportConfig{
		Name:                   p.Name,
		URL:                    p.URL,
		CheckType:              string(p.CheckType),
		Method:                 p.Method,
		Body:                   p.Body,
		TimeoutMs:              p.TimeoutMs,
		IntervalSeconds:        p.IntervalSeconds,
		Regions:                p.Regions,
		ExpectedStatusCodes:    p.ExpectedStatusCodes,
		KeywordCheck:           p.KeywordCheck,
		KeywordCheckType:       p.KeywordCheckType,
		JSONPathCheck:          p.JSONPathCheck,
		JSONPathExpected:       p.JSONPathExpected,
		SSLCheckEnabled:        p.SSLCheckEnabled,
		SSLExpiryThresholdDays: p.SSLExpiryThresholdDays,
		FollowRedirects:        p.FollowRedirects,
		MaxRedirects:           p.MaxRedirects,
	}

	// Parse headers from JSON string to map
	if p.Headers != "" && p.Headers != "{}" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(p.Headers), &headers); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not parse headers for probe %q: %v\n", p.Name, err)
		} else if len(headers) > 0 {
			cfg.Headers = headers
		}
	}

	// Convert channel IDs to strings
	if len(p.AlertChannelIDs) > 0 {
		channelIDs := make([]string, 0, len(p.AlertChannelIDs))
		for _, id := range p.AlertChannelIDs {
			channelIDs = append(channelIDs, id.String())
		}
		cfg.AlertChannelIDs = channelIDs
	}

	// Convert labels
	if len(p.Labels) > 0 {
		labels := make([]probeExportLabel, 0, len(p.Labels))
		for _, l := range p.Labels {
			labels = append(labels, probeExportLabel{
				Key:   l.Key,
				Value: l.Value,
			})
		}
		cfg.Labels = labels
	}

	return cfg
}
