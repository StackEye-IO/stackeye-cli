// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// probeTestTimeout is the maximum time to wait for the test API response.
// This includes time to fetch the probe and execute the test check.
const probeTestTimeout = 60 * time.Second

// ProbeTestResult wraps the test response with probe metadata for output formatting.
// This struct is exported to allow JSON/YAML serialization with proper field tags.
type ProbeTestResult struct {
	ProbeID        uuid.UUID `json:"probe_id" yaml:"probe_id"`
	ProbeName      string    `json:"probe_name" yaml:"probe_name"`
	ProbeURL       string    `json:"probe_url" yaml:"probe_url"`
	Status         string    `json:"status" yaml:"status"`
	ResponseTimeMs int       `json:"response_time_ms" yaml:"response_time_ms"`
	StatusCode     *int      `json:"status_code,omitempty" yaml:"status_code,omitempty"`
	ErrorMessage   *string   `json:"error_message,omitempty" yaml:"error_message,omitempty"`
	SSLExpiryDays  *int      `json:"ssl_expiry_days,omitempty" yaml:"ssl_expiry_days,omitempty"`
	CheckedAt      time.Time `json:"checked_at" yaml:"checked_at"`
}

// NewProbeTestCmd creates and returns the probe test subcommand.
func NewProbeTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <id>",
		Short: "Run an immediate test check for a probe",
		Long: `Run an immediate test check for a probe without affecting its history.

This command fetches the probe configuration and executes an ad-hoc test check
using the probe's current settings. The result shows detailed information about
the check including response time, status code, and any errors.

The test check:
  - Uses the probe's current configuration
  - Does NOT affect the probe's history or statistics
  - Does NOT trigger alerts
  - Runs immediately regardless of the probe's check interval

This is useful for:
  - Verifying a probe's configuration is correct
  - Troubleshooting connectivity issues
  - Testing changes before enabling monitoring

Examples:
  # Run a test check for a probe
  stackeye probe test 550e8400-e29b-41d4-a716-446655440000

  # Output as JSON for scripting
  stackeye probe test 550e8400-e29b-41d4-a716-446655440000 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeTest(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runProbeTest executes the probe test command logic.
func runProbeTest(ctx context.Context, idArg string) error {
	// Parse and validate UUID
	probeID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid probe ID %q: must be a valid UUID", idArg)
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create context with timeout for the entire operation
	reqCtx, cancel := context.WithTimeout(ctx, probeTestTimeout)
	defer cancel()

	// First, fetch the probe to get its configuration
	fmt.Printf("Fetching probe %s...\n", probeID)
	probe, err := client.GetProbe(reqCtx, apiClient, probeID, "")
	if err != nil {
		return fmt.Errorf("failed to get probe: %w", err)
	}

	// Convert probe configuration to test request
	testReq := probeToTestRequest(probe)

	// Execute the test check
	fmt.Printf("Running test check for %q (%s)...\n", probe.Name, probe.URL)
	result, err := client.TestProbe(reqCtx, apiClient, testReq)
	if err != nil {
		return fmt.Errorf("failed to execute test check: %w", err)
	}

	// Create combined result with probe metadata
	testResult := &ProbeTestResult{
		ProbeID:        probe.ID,
		ProbeName:      probe.Name,
		ProbeURL:       probe.URL,
		Status:         result.Status,
		ResponseTimeMs: result.ResponseTimeMs,
		StatusCode:     result.StatusCode,
		ErrorMessage:   result.ErrorMessage,
		SSLExpiryDays:  result.SSLExpiryDays,
		CheckedAt:      result.CheckedAt,
	}

	// Print using configured output format (supports json, yaml, table)
	return output.Print(testResult)
}

// probeToTestRequest converts a Probe to a ProbeTestRequest for ad-hoc testing.
func probeToTestRequest(probe *client.Probe) *client.ProbeTestRequest {
	// Convert FollowRedirects bool to *bool
	followRedirects := probe.FollowRedirects

	return &client.ProbeTestRequest{
		URL:                 probe.URL,
		CheckType:           probe.CheckType,
		Method:              probe.Method,
		Headers:             probe.Headers,
		Body:                probe.Body,
		TimeoutMs:           probe.TimeoutMs,
		ExpectedStatusCodes: probe.ExpectedStatusCodes,
		KeywordCheck:        probe.KeywordCheck,
		KeywordCheckType:    probe.KeywordCheckType,
		JSONPathCheck:       probe.JSONPathCheck,
		JSONPathExpected:    probe.JSONPathExpected,
		FollowRedirects:     &followRedirects,
		MaxRedirects:        probe.MaxRedirects,
	}
}
