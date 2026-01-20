// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// probeUpdateTimeout is the maximum time to wait for the API response.
const probeUpdateTimeout = 30 * time.Second

// probeUpdateFlags holds the flag values for the probe update command.
// All fields are pointers to support partial updates (nil = not specified).
type probeUpdateFlags struct {
	// Basic configuration
	name            *string
	url             *string
	method          *string
	intervalSeconds *int
	timeoutSeconds  *int
	regions         []string // Empty slice means not specified
	regionsSet      bool     // Track if --regions was explicitly provided

	// HTTP configuration
	headers             *string
	body                *string
	expectedStatusCodes *string
	followRedirects     *bool
	maxRedirects        *int

	// Content validation
	keywordCheck     *string
	keywordCheckType *string
	jsonPathCheck    *string
	jsonPathExpected *string

	// SSL
	sslCheckEnabled        *bool
	sslExpiryThresholdDays *int
}

// NewProbeUpdateCmd creates and returns the probe update subcommand.
func NewProbeUpdateCmd() *cobra.Command {
	flags := &probeUpdateFlags{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing monitoring probe",
		Long: `Update an existing monitoring probe configuration.

Only the specified flags will be updated; all other fields remain unchanged.
This allows for partial updates without needing to specify the entire configuration.

Examples:
  # Update probe name
  stackeye probe update 550e8400-e29b-41d4-a716-446655440000 --name "New Name"

  # Update check interval and timeout
  stackeye probe update 550e8400-e29b-41d4-a716-446655440000 \
    --interval 120 --timeout 15

  # Update monitoring regions
  stackeye probe update 550e8400-e29b-41d4-a716-446655440000 \
    --regions us-east-1,eu-west-1

  # Update URL and add keyword check
  stackeye probe update 550e8400-e29b-41d4-a716-446655440000 \
    --url https://api.example.com/v2/health \
    --keyword-check "healthy" --keyword-check-type contains

  # Disable SSL monitoring
  stackeye probe update 550e8400-e29b-41d4-a716-446655440000 \
    --ssl-check-enabled=false

  # Clear keyword check (set to empty)
  stackeye probe update 550e8400-e29b-41d4-a716-446655440000 \
    --keyword-check ""`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Track which flags were explicitly set
			if cmd.Flags().Changed("regions") {
				flags.regionsSet = true
			}
			return runProbeUpdate(cmd, args[0], flags)
		},
	}

	// All flags are optional for partial updates
	// Using custom functions to handle nil vs zero-value distinction

	// Basic configuration flags
	cmd.Flags().StringVar(stringPtrVar(&flags.name), "name", "", "probe name")
	cmd.Flags().StringVar(stringPtrVar(&flags.url), "url", "", "target URL or host to monitor")
	cmd.Flags().StringVar(stringPtrVar(&flags.method), "method", "", "HTTP method: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
	cmd.Flags().IntVar(intPtrVar(&flags.intervalSeconds), "interval", 0, "check interval in seconds (30-3600)")
	cmd.Flags().IntVar(intPtrVar(&flags.timeoutSeconds), "timeout", 0, "request timeout in seconds (1-60)")
	cmd.Flags().StringSliceVar(&flags.regions, "regions", nil, "monitoring regions (comma-separated)")

	// HTTP configuration flags
	cmd.Flags().StringVar(stringPtrVar(&flags.headers), "headers", "", "custom headers as JSON object")
	cmd.Flags().StringVar(stringPtrVar(&flags.body), "body", "", "request body for POST/PUT methods")
	cmd.Flags().StringVar(stringPtrVar(&flags.expectedStatusCodes), "expected-status-codes", "", "expected HTTP status codes (comma-separated)")
	cmd.Flags().BoolVar(boolPtrVar(&flags.followRedirects), "follow-redirects", false, "follow HTTP redirects")
	cmd.Flags().IntVar(intPtrVar(&flags.maxRedirects), "max-redirects", 0, "maximum redirects to follow (0-20)")

	// Content validation flags
	cmd.Flags().StringVar(stringPtrVar(&flags.keywordCheck), "keyword-check", "", "keyword to search in response body (empty to clear)")
	cmd.Flags().StringVar(stringPtrVar(&flags.keywordCheckType), "keyword-check-type", "", "keyword check type: contains, not_contains")
	cmd.Flags().StringVar(stringPtrVar(&flags.jsonPathCheck), "json-path-check", "", "JSONPath expression to evaluate (empty to clear)")
	cmd.Flags().StringVar(stringPtrVar(&flags.jsonPathExpected), "json-path-expected", "", "expected value from JSONPath")

	// SSL flags
	cmd.Flags().BoolVar(boolPtrVar(&flags.sslCheckEnabled), "ssl-check-enabled", false, "enable SSL certificate monitoring")
	cmd.Flags().IntVar(intPtrVar(&flags.sslExpiryThresholdDays), "ssl-expiry-threshold-days", 0, "alert when SSL expires within N days (1-365)")

	return cmd
}

// Helper functions to work with pointer flags
func stringPtrVar(p **string) *string {
	*p = new(string)
	return *p
}

func intPtrVar(p **int) *int {
	*p = new(int)
	return *p
}

func boolPtrVar(p **bool) *bool {
	*p = new(bool)
	return *p
}

// runProbeUpdate executes the probe update command logic.
func runProbeUpdate(cmd *cobra.Command, idArg string, flags *probeUpdateFlags) error {
	// Parse and validate UUID
	probeID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid probe ID %q: must be a valid UUID", idArg)
	}

	// Build request with only the changed fields
	req := &client.UpdateProbeRequest{}
	hasUpdates := false

	// Check each flag and add to request only if it was explicitly set
	if cmd.Flags().Changed("name") {
		if *flags.name == "" {
			return fmt.Errorf("--name cannot be empty")
		}
		req.Name = flags.name
		hasUpdates = true
	}

	if cmd.Flags().Changed("url") {
		if *flags.url == "" {
			return fmt.Errorf("--url cannot be empty")
		}
		if err := validateProbeURL(*flags.url); err != nil {
			return err
		}
		req.URL = flags.url
		hasUpdates = true
	}

	if cmd.Flags().Changed("method") {
		upperMethod := strings.ToUpper(*flags.method)
		if err := validateHTTPMethod(upperMethod); err != nil {
			return err
		}
		req.Method = &upperMethod
		hasUpdates = true
	}

	if cmd.Flags().Changed("interval") {
		if *flags.intervalSeconds < 30 || *flags.intervalSeconds > 3600 {
			return fmt.Errorf("--interval must be between 30 and 3600 seconds, got %d", *flags.intervalSeconds)
		}
		req.IntervalSeconds = flags.intervalSeconds
		hasUpdates = true
	}

	if cmd.Flags().Changed("timeout") {
		if *flags.timeoutSeconds < 1 || *flags.timeoutSeconds > 60 {
			return fmt.Errorf("--timeout must be between 1 and 60 seconds, got %d", *flags.timeoutSeconds)
		}
		timeoutMs := *flags.timeoutSeconds * 1000
		req.TimeoutMs = &timeoutMs
		hasUpdates = true
	}

	if flags.regionsSet {
		req.Regions = flags.regions
		hasUpdates = true
	}

	if cmd.Flags().Changed("headers") {
		req.Headers = flags.headers
		hasUpdates = true
	}

	if cmd.Flags().Changed("body") {
		req.Body = flags.body
		hasUpdates = true
	}

	if cmd.Flags().Changed("expected-status-codes") {
		codes, err := parseStatusCodes(*flags.expectedStatusCodes)
		if err != nil {
			return fmt.Errorf("invalid --expected-status-codes: %w", err)
		}
		req.ExpectedStatusCodes = codes
		hasUpdates = true
	}

	if cmd.Flags().Changed("follow-redirects") {
		req.FollowRedirects = flags.followRedirects
		hasUpdates = true
	}

	if cmd.Flags().Changed("max-redirects") {
		if *flags.maxRedirects < 0 || *flags.maxRedirects > 20 {
			return fmt.Errorf("--max-redirects must be between 0 and 20, got %d", *flags.maxRedirects)
		}
		req.MaxRedirects = flags.maxRedirects
		hasUpdates = true
	}

	if cmd.Flags().Changed("keyword-check") {
		req.KeywordCheck = flags.keywordCheck
		hasUpdates = true
	}

	if cmd.Flags().Changed("keyword-check-type") {
		if err := validateKeywordCheckType(*flags.keywordCheckType); err != nil {
			return err
		}
		req.KeywordCheckType = flags.keywordCheckType
		hasUpdates = true
	}

	if cmd.Flags().Changed("json-path-check") {
		req.JSONPathCheck = flags.jsonPathCheck
		hasUpdates = true
	}

	if cmd.Flags().Changed("json-path-expected") {
		req.JSONPathExpected = flags.jsonPathExpected
		hasUpdates = true
	}

	if cmd.Flags().Changed("ssl-check-enabled") {
		req.SSLCheckEnabled = flags.sslCheckEnabled
		hasUpdates = true
	}

	if cmd.Flags().Changed("ssl-expiry-threshold-days") {
		if *flags.sslExpiryThresholdDays < 1 || *flags.sslExpiryThresholdDays > 365 {
			return fmt.Errorf("--ssl-expiry-threshold-days must be between 1 and 365, got %d", *flags.sslExpiryThresholdDays)
		}
		req.SSLExpiryThresholdDays = flags.sslExpiryThresholdDays
		hasUpdates = true
	}

	// Require at least one update flag
	if !hasUpdates {
		return fmt.Errorf("no update flags specified; use --help to see available options")
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to update probe with timeout
	ctx, cancel := context.WithTimeout(cmd.Context(), probeUpdateTimeout)
	defer cancel()

	probe, err := client.UpdateProbe(ctx, apiClient, probeID, req)
	if err != nil {
		return fmt.Errorf("failed to update probe: %w", err)
	}

	// Print the updated probe using the configured output format
	return output.Print(probe)
}
