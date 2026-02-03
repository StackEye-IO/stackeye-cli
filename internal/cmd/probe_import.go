// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// probeImportTimeout is the maximum time to wait for all import API calls.
const probeImportTimeout = 120 * time.Second

// probeImportFlags holds the flag values for the probe import command.
type probeImportFlags struct {
	file   string
	format string
	dryRun bool
}

// probeImportResult tracks the outcome of an import operation.
type probeImportResult struct {
	Created []string `json:"created"`
	Skipped []string `json:"skipped"`
	Failed  []string `json:"failed"`
	Total   int      `json:"total"`
	Errors  []string `json:"errors,omitempty"`
}

// NewProbeImportCmd creates and returns the probe import subcommand.
func NewProbeImportCmd() *cobra.Command {
	flags := &probeImportFlags{}

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import probe configurations from file",
		Long: `Import probe configurations from a JSON or YAML file.

Reads probe configurations from a file (as produced by "probe export") and
creates them via the API. The file format is auto-detected from the file
extension, or can be specified with --format.

Duplicate Detection:
  Before creating each probe, the import checks for an existing probe with
  the same name. If a duplicate is found, the probe is skipped and reported.

Supported Formats:
  yaml    YAML format (.yaml, .yml extensions)
  json    JSON format (.json extension)

Examples:
  # Import probes from a YAML file
  stackeye probe import --file probes.yaml

  # Import from JSON
  stackeye probe import --file probes.json

  # Preview what would be imported without creating
  stackeye probe import --file probes.yaml --dry-run

  # Specify format explicitly
  stackeye probe import --file probes.txt --format yaml`,
		Aliases: []string{"imp"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeImport(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVarP(&flags.file, "file", "f", "", "input file path (required)")
	cmd.Flags().StringVar(&flags.format, "format", "", "input format: yaml, json (auto-detected from extension if omitted)")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "preview import without creating probes")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

// runProbeImport executes the probe import command logic.
func runProbeImport(ctx context.Context, flags *probeImportFlags) error {
	// Detect format from file extension if not specified
	format, err := resolveImportFormat(flags.file, flags.format)
	if err != nil {
		return err
	}

	// Read and parse the file
	configs, err := readProbeConfigs(flags.file, format)
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return fmt.Errorf("no probe configurations found in %q", flags.file)
	}

	// Validate all configs before making any API calls
	if err := validateProbeConfigs(configs); err != nil {
		return err
	}

	// In dry-run mode, show what would be imported and exit
	if flags.dryRun {
		return printDryRunSummary(configs)
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeImportTimeout)
	defer cancel()

	// Fetch existing probe names for duplicate detection
	existingNames, err := fetchExistingProbeNames(reqCtx, apiClient)
	if err != nil {
		return err
	}

	// Import each probe
	result := &probeImportResult{
		Total: len(configs),
	}

	for i := range configs {
		cfg := &configs[i]

		// Check for duplicate
		if _, exists := existingNames[cfg.Name]; exists {
			result.Skipped = append(result.Skipped, cfg.Name)
			fmt.Fprintf(os.Stderr, "Skipped %q: probe with this name already exists\n", cfg.Name)
			continue
		}

		// Convert and create
		req := convertExportConfigToCreateRequest(cfg)
		probe, err := client.CreateProbe(reqCtx, apiClient, req)
		if err != nil {
			result.Failed = append(result.Failed, cfg.Name)
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", cfg.Name, err))
			fmt.Fprintf(os.Stderr, "Failed %q: %v\n", cfg.Name, err)
			continue
		}

		result.Created = append(result.Created, probe.Name)
		// Track the newly created name to prevent duplicates within the same import
		existingNames[probe.Name] = true
	}

	return output.Print(result)
}

// resolveImportFormat determines the file format from the flag or file extension.
func resolveImportFormat(filePath, flagFormat string) (string, error) {
	if flagFormat != "" {
		format := strings.ToLower(flagFormat)
		if format != "yaml" && format != "json" {
			return "", clierrors.InvalidValueError("--format", flagFormat, clierrors.ValidExportFormats)
		}
		return format, nil
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".yaml", ".yml":
		return "yaml", nil
	case ".json":
		return "json", nil
	default:
		return "", fmt.Errorf("cannot detect format from extension %q; use --format to specify (yaml or json)", ext)
	}
}

// readProbeConfigs reads and unmarshals probe configurations from a file.
func readProbeConfigs(filePath, format string) ([]probeExportConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("file %q is empty", filePath)
	}

	var configs []probeExportConfig
	switch format {
	case "json":
		if err := json.Unmarshal(data, &configs); err != nil {
			return nil, fmt.Errorf("failed to parse JSON from %q: %w", filePath, err)
		}
	case "yaml":
		if err := yaml.Unmarshal(data, &configs); err != nil {
			return nil, fmt.Errorf("failed to parse YAML from %q: %w", filePath, err)
		}
	}

	return configs, nil
}

// validateProbeConfigs validates all probe configurations before import.
func validateProbeConfigs(configs []probeExportConfig) error {
	for i, cfg := range configs {
		if cfg.Name == "" {
			return fmt.Errorf("probe at index %d: name is required", i)
		}
		if cfg.URL == "" {
			return fmt.Errorf("probe %q (index %d): url is required", cfg.Name, i)
		}
		if cfg.CheckType == "" {
			return fmt.Errorf("probe %q (index %d): check_type is required", cfg.Name, i)
		}
		if err := validateCheckType(cfg.CheckType); err != nil {
			return fmt.Errorf("probe %q (index %d): %w", cfg.Name, i, err)
		}
		if cfg.CheckType == "http" {
			if err := validateProbeURL(cfg.URL); err != nil {
				return fmt.Errorf("probe %q (index %d): %w", cfg.Name, i, err)
			}
		}
		if cfg.Method != "" {
			if err := validateHTTPMethod(cfg.Method); err != nil {
				return fmt.Errorf("probe %q (index %d): %w", cfg.Name, i, err)
			}
		}
		if cfg.IntervalSeconds != 0 && (cfg.IntervalSeconds < 30 || cfg.IntervalSeconds > 3600) {
			return fmt.Errorf("probe %q (index %d): interval_seconds must be between 30 and 3600, got %d", cfg.Name, i, cfg.IntervalSeconds)
		}
		if cfg.TimeoutMs != 0 {
			timeoutSec := cfg.TimeoutMs / 1000
			if timeoutSec < 1 || timeoutSec > 60 {
				return fmt.Errorf("probe %q (index %d): timeout_ms must be between 1000 and 60000, got %d", cfg.Name, i, cfg.TimeoutMs)
			}
		}
		if cfg.KeywordCheck != nil && *cfg.KeywordCheck != "" {
			if cfg.KeywordCheckType != nil {
				if err := validateKeywordCheckType(*cfg.KeywordCheckType); err != nil {
					return fmt.Errorf("probe %q (index %d): %w", cfg.Name, i, err)
				}
			}
		}
		if cfg.MaxRedirects < 0 || cfg.MaxRedirects > 20 {
			return fmt.Errorf("probe %q (index %d): max_redirects must be between 0 and 20, got %d", cfg.Name, i, cfg.MaxRedirects)
		}
		if cfg.SSLExpiryThresholdDays != 0 && (cfg.SSLExpiryThresholdDays < 1 || cfg.SSLExpiryThresholdDays > 365) {
			return fmt.Errorf("probe %q (index %d): ssl_expiry_threshold_days must be between 1 and 365, got %d", cfg.Name, i, cfg.SSLExpiryThresholdDays)
		}
	}
	return nil
}

// printDryRunSummary shows what would be imported without creating probes.
func printDryRunSummary(configs []probeExportConfig) error {
	fmt.Fprintf(os.Stderr, "Dry run: %d probe(s) would be imported:\n\n", len(configs))
	for i, cfg := range configs {
		method := cfg.Method
		if method == "" {
			method = "GET"
		}
		fmt.Fprintf(os.Stderr, "  %d. %s\n     URL: %s\n     Type: %s | Method: %s | Interval: %ds\n",
			i+1, cfg.Name, cfg.URL, cfg.CheckType, method, cfg.IntervalSeconds)
		if len(cfg.Regions) > 0 {
			fmt.Fprintf(os.Stderr, "     Regions: %s\n", strings.Join(cfg.Regions, ", "))
		}
		if len(cfg.Labels) > 0 {
			labelParts := make([]string, 0, len(cfg.Labels))
			for _, l := range cfg.Labels {
				if l.Value != nil {
					labelParts = append(labelParts, fmt.Sprintf("%s=%s", l.Key, *l.Value))
				} else {
					labelParts = append(labelParts, l.Key)
				}
			}
			fmt.Fprintf(os.Stderr, "     Labels: %s\n", strings.Join(labelParts, ", "))
		}
		fmt.Fprintln(os.Stderr)
	}
	fmt.Fprintf(os.Stderr, "No probes were created (dry run).\n")
	return nil
}

// fetchExistingProbeNames fetches all existing probe names for duplicate detection.
func fetchExistingProbeNames(ctx context.Context, apiClient *client.Client) (map[string]bool, error) {
	names := make(map[string]bool)
	page := 1
	limit := 100

	for {
		opts := &client.ListProbesOptions{
			Page:  page,
			Limit: limit,
		}

		result, err := client.ListProbes(ctx, apiClient, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list existing probes: %w", err)
		}

		for _, p := range result.Probes {
			names[p.Name] = true
		}

		if len(result.Probes) < limit {
			break
		}
		page++
	}

	return names, nil
}

// convertExportConfigToCreateRequest converts a probeExportConfig to a CreateProbeRequest.
func convertExportConfigToCreateRequest(cfg *probeExportConfig) *client.CreateProbeRequest {
	method := cfg.Method
	if method == "" {
		method = "GET"
	}

	timeoutMs := cfg.TimeoutMs
	if timeoutMs == 0 {
		timeoutMs = 10000
	}

	intervalSeconds := cfg.IntervalSeconds
	if intervalSeconds == 0 {
		intervalSeconds = 60
	}

	expectedCodes := cfg.ExpectedStatusCodes
	if len(expectedCodes) == 0 {
		expectedCodes = []int{200}
	}

	req := &client.CreateProbeRequest{
		Name:                   cfg.Name,
		URL:                    cfg.URL,
		CheckType:              client.CheckType(cfg.CheckType),
		Method:                 strings.ToUpper(method),
		TimeoutMs:              timeoutMs,
		IntervalSeconds:        intervalSeconds,
		Regions:                cfg.Regions,
		ExpectedStatusCodes:    expectedCodes,
		SSLCheckEnabled:        cfg.SSLCheckEnabled,
		SSLExpiryThresholdDays: cfg.SSLExpiryThresholdDays,
		MaxRedirects:           cfg.MaxRedirects,
	}

	// Convert headers map to JSON string
	if len(cfg.Headers) > 0 {
		headersJSON, err := json.Marshal(cfg.Headers)
		if err == nil {
			req.Headers = string(headersJSON)
		}
	}

	// Set body if provided
	if cfg.Body != nil {
		req.Body = cfg.Body
	}

	// Set keyword check if provided
	if cfg.KeywordCheck != nil && *cfg.KeywordCheck != "" {
		req.KeywordCheck = cfg.KeywordCheck
		req.KeywordCheckType = cfg.KeywordCheckType
	}

	// Set JSON path check if provided
	if cfg.JSONPathCheck != nil && *cfg.JSONPathCheck != "" {
		req.JSONPathCheck = cfg.JSONPathCheck
		req.JSONPathExpected = cfg.JSONPathExpected
	}

	// Set follow redirects
	req.FollowRedirects = &cfg.FollowRedirects

	// Convert alert channel IDs from strings to UUIDs
	if len(cfg.AlertChannelIDs) > 0 {
		channelIDs := make([]uuid.UUID, 0, len(cfg.AlertChannelIDs))
		for _, idStr := range cfg.AlertChannelIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: skipping invalid channel ID %q: %v\n", idStr, err)
				continue
			}
			channelIDs = append(channelIDs, id)
		}
		if len(channelIDs) > 0 {
			req.AlertChannelIDs = channelIDs
		}
	}

	return req
}
