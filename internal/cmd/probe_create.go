// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// probeCreateTimeout is the maximum time to wait for the API response.
const probeCreateTimeout = 30 * time.Second

// probeCreateFlags holds the flag values for the probe create command.
type probeCreateFlags struct {
	// Required
	name string
	url  string

	// Optional - basic
	checkType       string
	method          string
	intervalSeconds int
	timeoutSeconds  int
	regions         []string

	// Optional - HTTP configuration
	headers             string
	body                string
	expectedStatusCodes string
	followRedirects     bool
	maxRedirects        int

	// Optional - content validation
	keywordCheck     string
	keywordCheckType string
	jsonPathCheck    string
	jsonPathExpected string

	// Optional - SSL
	sslCheckEnabled        bool
	sslExpiryThresholdDays int
}

// NewProbeCreateCmd creates and returns the probe create subcommand.
func NewProbeCreateCmd() *cobra.Command {
	flags := &probeCreateFlags{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new monitoring probe",
		Long: `Create a new monitoring probe to check an endpoint.

By default, creates an HTTP probe that monitors the specified URL. The probe
will be executed from all active regions unless specific regions are provided.

Required Flags:
  --name       Human-readable name for the probe
  --url        Target URL or host to monitor

Check Types:
  http         HTTP/HTTPS endpoint (default)
  ping         ICMP ping check
  tcp          TCP port connectivity
  dns_resolve  DNS resolution check

Examples:
  # Create a basic HTTP probe
  stackeye probe create --name "API Health" --url https://api.example.com/health

  # Create with custom interval and timeout
  stackeye probe create --name "Main Site" --url https://example.com \
    --interval 60 --timeout 10

  # Create with specific regions
  stackeye probe create --name "EU Service" --url https://eu.example.com \
    --regions us-east-1,eu-west-1,ap-southeast-1

  # Create with content validation
  stackeye probe create --name "API Status" --url https://api.example.com/status \
    --keyword-check "\"status\":\"healthy\"" --keyword-check-type contains

  # Create with JSON path validation
  stackeye probe create --name "Version Check" --url https://api.example.com/version \
    --json-path-check "$.version" --json-path-expected "1.0.0"

  # Create with custom headers
  stackeye probe create --name "Auth Endpoint" --url https://api.example.com/me \
    --headers '{"Authorization":"Bearer token"}'

  # Create POST probe with body
  stackeye probe create --name "Webhook Test" --url https://api.example.com/webhook \
    --method POST --body '{"test":true}'

  # Create TCP probe
  stackeye probe create --name "Database" --url db.example.com:5432 --check-type tcp

  # Create with SSL monitoring
  stackeye probe create --name "Secure Site" --url https://secure.example.com \
    --ssl-check-enabled --ssl-expiry-threshold-days 30`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeCreate(cmd.Context(), flags)
		},
	}

	// Required flags
	cmd.Flags().StringVar(&flags.name, "name", "", "probe name (required)")
	cmd.Flags().StringVar(&flags.url, "url", "", "target URL or host to monitor (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("url")

	// Basic optional flags
	cmd.Flags().StringVar(&flags.checkType, "check-type", "http", "check type: http, ping, tcp, dns_resolve")
	cmd.Flags().StringVar(&flags.method, "method", "GET", "HTTP method: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
	cmd.Flags().IntVar(&flags.intervalSeconds, "interval", 60, "check interval in seconds (30-3600)")
	cmd.Flags().IntVar(&flags.timeoutSeconds, "timeout", 10, "request timeout in seconds (1-60)")
	cmd.Flags().StringSliceVar(&flags.regions, "regions", nil, "monitoring regions (comma-separated)")

	// HTTP configuration flags
	cmd.Flags().StringVar(&flags.headers, "headers", "", "custom headers as JSON object")
	cmd.Flags().StringVar(&flags.body, "body", "", "request body for POST/PUT methods")
	cmd.Flags().StringVar(&flags.expectedStatusCodes, "expected-status-codes", "200", "expected HTTP status codes (comma-separated)")
	cmd.Flags().BoolVar(&flags.followRedirects, "follow-redirects", true, "follow HTTP redirects")
	cmd.Flags().IntVar(&flags.maxRedirects, "max-redirects", 10, "maximum redirects to follow")

	// Content validation flags
	cmd.Flags().StringVar(&flags.keywordCheck, "keyword-check", "", "keyword to search in response body")
	cmd.Flags().StringVar(&flags.keywordCheckType, "keyword-check-type", "contains", "keyword check type: contains, not_contains")
	cmd.Flags().StringVar(&flags.jsonPathCheck, "json-path-check", "", "JSONPath expression to evaluate")
	cmd.Flags().StringVar(&flags.jsonPathExpected, "json-path-expected", "", "expected value from JSONPath")

	// SSL flags
	cmd.Flags().BoolVar(&flags.sslCheckEnabled, "ssl-check-enabled", true, "enable SSL certificate monitoring")
	cmd.Flags().IntVar(&flags.sslExpiryThresholdDays, "ssl-expiry-threshold-days", 14, "alert when SSL expires within N days")

	return cmd
}

// runProbeCreate executes the probe create command logic.
func runProbeCreate(ctx context.Context, flags *probeCreateFlags) error {
	// Validate required fields
	if flags.name == "" {
		return fmt.Errorf("--name is required")
	}
	if flags.url == "" {
		return fmt.Errorf("--url is required")
	}

	// Validate URL format for HTTP probes
	if flags.checkType == "http" || flags.checkType == "" {
		if err := validateProbeURL(flags.url); err != nil {
			return err
		}
	}

	// Validate check type
	if err := validateCheckType(flags.checkType); err != nil {
		return err
	}

	// Validate HTTP method
	if err := validateHTTPMethod(flags.method); err != nil {
		return err
	}

	// Validate interval
	if flags.intervalSeconds < 30 || flags.intervalSeconds > 3600 {
		return fmt.Errorf("--interval must be between 30 and 3600 seconds, got %d", flags.intervalSeconds)
	}

	// Validate timeout
	if flags.timeoutSeconds < 1 || flags.timeoutSeconds > 60 {
		return fmt.Errorf("--timeout must be between 1 and 60 seconds, got %d", flags.timeoutSeconds)
	}

	// Validate keyword check type
	if flags.keywordCheck != "" {
		if err := validateKeywordCheckType(flags.keywordCheckType); err != nil {
			return err
		}
	}

	// Validate max redirects
	if flags.maxRedirects < 0 || flags.maxRedirects > 20 {
		return fmt.Errorf("--max-redirects must be between 0 and 20, got %d", flags.maxRedirects)
	}

	// Validate SSL expiry threshold
	if flags.sslExpiryThresholdDays < 1 || flags.sslExpiryThresholdDays > 365 {
		return fmt.Errorf("--ssl-expiry-threshold-days must be between 1 and 365, got %d", flags.sslExpiryThresholdDays)
	}

	// Parse expected status codes
	expectedCodes, err := parseStatusCodes(flags.expectedStatusCodes)
	if err != nil {
		return fmt.Errorf("invalid --expected-status-codes: %w", err)
	}

	// Dry-run check: print what would happen and exit without making API calls
	if GetDryRun() {
		dryrun.PrintAction("create", "probe",
			"Name", flags.name,
			"URL", flags.url,
			"Check Type", flags.checkType,
			"Method", flags.method,
			"Interval", fmt.Sprintf("%ds", flags.intervalSeconds),
			"Timeout", fmt.Sprintf("%ds", flags.timeoutSeconds),
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build the create request
	req := buildCreateProbeRequest(flags, expectedCodes)

	// Call SDK to create probe with timeout
	reqCtx, cancel := context.WithTimeout(ctx, probeCreateTimeout)
	defer cancel()

	probe, err := client.CreateProbe(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create probe: %w", err)
	}

	// Print the created probe using the configured output format
	return output.Print(probe)
}

// validateProbeURL validates the URL format for HTTP probes.
func validateProbeURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}

	if parsed.Scheme == "" {
		return fmt.Errorf("URL must include scheme (http:// or https://): %q", rawURL)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got %q", parsed.Scheme)
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must include a host: %q", rawURL)
	}

	return nil
}

// validateCheckType validates the check type value.
func validateCheckType(checkType string) error {
	valid := map[string]bool{
		"http":        true,
		"ping":        true,
		"tcp":         true,
		"dns_resolve": true,
	}
	if !valid[checkType] {
		return clierrors.InvalidValueError("--check-type", checkType, clierrors.ValidCheckTypes)
	}
	return nil
}

// validateHTTPMethod validates the HTTP method value.
func validateHTTPMethod(method string) error {
	valid := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}
	if !valid[strings.ToUpper(method)] {
		return clierrors.InvalidValueError("--method", method, clierrors.ValidHTTPMethods)
	}
	return nil
}

// validateKeywordCheckType validates the keyword check type value.
func validateKeywordCheckType(checkType string) error {
	valid := map[string]bool{
		"contains":     true,
		"not_contains": true,
	}
	if !valid[checkType] {
		return clierrors.InvalidValueError("--keyword-check-type", checkType, clierrors.ValidKeywordCheckTypes)
	}
	return nil
}

// parseStatusCodes parses a comma-separated string of status codes.
func parseStatusCodes(s string) ([]int, error) {
	if s == "" {
		return []int{200}, nil
	}

	parts := strings.Split(s, ",")
	codes := make([]int, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		code, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid status code %q", p)
		}
		if code < 100 || code > 599 {
			return nil, fmt.Errorf("status code %d out of range (100-599)", code)
		}
		codes = append(codes, code)
	}

	if len(codes) == 0 {
		return []int{200}, nil
	}

	return codes, nil
}

// buildCreateProbeRequest constructs the API request from flags.
func buildCreateProbeRequest(flags *probeCreateFlags, expectedCodes []int) *client.CreateProbeRequest {
	req := &client.CreateProbeRequest{
		Name:                   flags.name,
		URL:                    flags.url,
		CheckType:              client.CheckType(flags.checkType),
		Method:                 strings.ToUpper(flags.method),
		Headers:                flags.headers,
		TimeoutMs:              flags.timeoutSeconds * 1000, // Convert seconds to milliseconds
		IntervalSeconds:        flags.intervalSeconds,
		Regions:                flags.regions,
		ExpectedStatusCodes:    expectedCodes,
		SSLCheckEnabled:        flags.sslCheckEnabled,
		SSLExpiryThresholdDays: flags.sslExpiryThresholdDays,
		MaxRedirects:           flags.maxRedirects,
	}

	// Set body if provided
	if flags.body != "" {
		req.Body = &flags.body
	}

	// Set keyword check if provided
	if flags.keywordCheck != "" {
		req.KeywordCheck = &flags.keywordCheck
		req.KeywordCheckType = &flags.keywordCheckType
	}

	// Set JSON path check if provided
	if flags.jsonPathCheck != "" {
		req.JSONPathCheck = &flags.jsonPathCheck
		if flags.jsonPathExpected != "" {
			req.JSONPathExpected = &flags.jsonPathExpected
		}
	}

	// Set follow redirects
	req.FollowRedirects = &flags.followRedirects

	return req
}
