// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	cliinteractive "github.com/StackEye-IO/stackeye-cli/internal/interactive"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/spf13/cobra"
)

// probeWizardTimeout is the maximum time for API operations during the wizard.
const probeWizardTimeout = 30 * time.Second

// NewProbeWizardCmd creates and returns the probe wizard command.
func NewProbeWizardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive wizard for creating a new probe",
		Long: `Interactive wizard that guides you through creating a new monitoring probe.

The wizard helps you configure:
  1. Probe type (HTTP, ping, TCP, DNS)
  2. Target URL or hostname
  3. HTTP method and settings (for HTTP probes)
  4. Monitoring regions
  5. Check interval and timeout
  6. SSL certificate monitoring
  7. Content validation (keyword or JSONPath)
  8. Optional test before saving

For non-interactive environments, use the probe create command instead:
  stackeye probe create --name "My Probe" --url https://example.com

Examples:
  # Launch the interactive probe wizard
  stackeye probe wizard

  # Use non-interactive mode (shows guidance)
  stackeye probe wizard --no-input`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeWizard(cmd.Context())
		},
	}

	return cmd
}

// runProbeWizard executes the interactive probe wizard.
func runProbeWizard(ctx context.Context) error {
	// Check for non-interactive mode
	if GetNoInput() {
		return runProbeWizardNonInteractive()
	}

	// Get authenticated API client early to validate auth
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create the wizard
	wiz := interactive.NewWizard(&interactive.WizardOptions{
		Title:         "Probe Creation Wizard",
		Description:   "Let's set up a new monitoring probe",
		ShowProgress:  true,
		ConfirmCancel: true,
	})

	// Store API client in wizard data for steps to access
	wiz.SetData("apiClient", apiClient)

	// Add wizard steps
	wiz.AddSteps(
		&interactive.Step{
			Name:        "probe-type",
			Title:       "Probe Type",
			Description: "Select the type of check to perform",
			Run:         stepProbeType,
		},
		&interactive.Step{
			Name:        "target",
			Title:       "Target",
			Description: "Enter the URL or hostname to monitor",
			Run:         stepProbeTarget,
		},
		&interactive.Step{
			Name:        "http-settings",
			Title:       "HTTP Settings",
			Description: "Configure HTTP request options",
			Run:         stepHTTPSettings,
			Skip:        skipIfNotHTTP,
		},
		&interactive.Step{
			Name:        "regions",
			Title:       "Regions",
			Description: "Select monitoring regions",
			Run:         stepProbeRegions,
		},
		&interactive.Step{
			Name:        "interval",
			Title:       "Check Interval",
			Description: "How often should we check your endpoint?",
			Run:         stepProbeInterval,
		},
		&interactive.Step{
			Name:        "ssl-monitoring",
			Title:       "SSL Monitoring",
			Description: "Configure SSL certificate monitoring",
			Run:         stepSSLMonitoring,
			Skip:        skipIfNotHTTPS,
		},
		&interactive.Step{
			Name:        "validation",
			Title:       "Content Validation",
			Description: "Optionally validate response content",
			Run:         stepContentValidation,
			Skip:        skipIfNotHTTP,
		},
		&interactive.Step{
			Name:        "probe-name",
			Title:       "Probe Name",
			Description: "Give your probe a name",
			Run:         stepProbeName,
		},
		&interactive.Step{
			Name:        "test-probe",
			Title:       "Test Probe",
			Description: "Test your probe configuration before saving",
			Run:         stepTestProbe,
		},
		&interactive.Step{
			Name:        "create-probe",
			Title:       "Create Probe",
			Description: "Review and create your probe",
			Run:         stepCreateProbe,
		},
	)

	// Run the wizard
	if err := wiz.Run(ctx); err != nil {
		if err == interactive.ErrWizardCancelled {
			fmt.Println("\nWizard cancelled. You can run 'stackeye probe wizard' again at any time.")
			return nil
		}
		return err
	}

	// Print final summary
	printProbeWizardSummary(wiz)

	return nil
}

// runProbeWizardNonInteractive handles the wizard in non-interactive mode.
func runProbeWizardNonInteractive() error {
	fmt.Println("Probe Creation Wizard (Non-Interactive Mode)")
	fmt.Println()
	fmt.Println("The probe wizard requires interactive input.")
	fmt.Println()
	fmt.Println("To create a probe non-interactively, use:")
	fmt.Println()
	fmt.Println("  stackeye probe create --name <name> --url <url> [options]")
	fmt.Println()
	fmt.Println("Required flags:")
	fmt.Println("  --name       Probe name")
	fmt.Println("  --url        Target URL to monitor")
	fmt.Println()
	fmt.Println("Common options:")
	fmt.Println("  --check-type          Check type: http, ping, tcp, dns_resolve (default: http)")
	fmt.Println("  --method              HTTP method: GET, POST, etc. (default: GET)")
	fmt.Println("  --interval            Check interval in seconds (default: 60)")
	fmt.Println("  --timeout             Request timeout in seconds (default: 10)")
	fmt.Println("  --regions             Comma-separated region IDs")
	fmt.Println("  --ssl-check-enabled   Enable SSL monitoring (default: true)")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  stackeye probe create --name \"My API\" --url https://api.example.com/health")
	fmt.Println()
	fmt.Println("For full options, run: stackeye probe create --help")
	return nil
}

// stepProbeType handles probe type selection.
func stepProbeType(ctx context.Context, wiz *interactive.Wizard) error {
	options := []string{
		"HTTP/HTTPS - Monitor web endpoints",
		"Ping - ICMP connectivity check",
		"TCP - Port connectivity check",
		"DNS Resolve - DNS resolution check",
	}

	selected, err := interactive.AskSelect(&interactive.SelectPromptOptions{
		Message:  "What type of check do you want to perform?",
		Options:  options,
		Default:  options[0],
		PageSize: 5,
	})
	if err != nil {
		return err
	}

	// Map selection to check type
	var checkType string
	switch {
	case strings.HasPrefix(selected, "HTTP"):
		checkType = "http"
	case strings.HasPrefix(selected, "Ping"):
		checkType = "ping"
	case strings.HasPrefix(selected, "TCP"):
		checkType = "tcp"
	case strings.HasPrefix(selected, "DNS"):
		checkType = "dns_resolve"
	}

	wiz.SetData("checkType", checkType)
	fmt.Printf("  Selected: %s\n", checkType)

	return nil
}

// stepProbeTarget handles target URL/host input.
func stepProbeTarget(ctx context.Context, wiz *interactive.Wizard) error {
	checkType := wiz.GetDataString("checkType")

	var message, helpText string
	switch checkType {
	case "http":
		message = "Enter the URL to monitor:"
		helpText = "Full URL including https:// (e.g., https://api.example.com/health)"
	case "ping":
		message = "Enter the hostname or IP to ping:"
		helpText = "Hostname or IP address (e.g., example.com or 192.168.1.1)"
	case "tcp":
		message = "Enter the host:port to check:"
		helpText = "Host and port (e.g., db.example.com:5432)"
	case "dns_resolve":
		message = "Enter the hostname to resolve:"
		helpText = "Hostname to resolve (e.g., example.com)"
	}

	target, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: message,
		Help:    helpText,
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("target cannot be empty")
			}
			if checkType == "http" {
				return validateProbeURL(s)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	wiz.SetData("target", target)
	return nil
}

// stepHTTPSettings handles HTTP-specific configuration.
func stepHTTPSettings(ctx context.Context, wiz *interactive.Wizard) error {
	// HTTP Method selection
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	method, err := interactive.AskSelect(&interactive.SelectPromptOptions{
		Message:  "HTTP method:",
		Options:  methods,
		Default:  "GET",
		PageSize: 7,
	})
	if err != nil {
		return err
	}
	wiz.SetData("method", method)

	// Expected status codes
	statusCodes, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Expected status codes (comma-separated):",
		Default: "200",
		Help:    "Comma-separated list of valid HTTP status codes (e.g., 200,201,204)",
		Validate: func(s string) error {
			_, err := parseStatusCodes(s)
			return err
		},
	})
	if err != nil {
		return err
	}
	wiz.SetData("statusCodes", statusCodes)

	// Follow redirects
	followRedirects, err := cliinteractive.Confirm(
		"Follow HTTP redirects?",
		cliinteractive.WithDefault(true),
	)
	if err != nil {
		return err
	}
	wiz.SetData("followRedirects", followRedirects)

	fmt.Printf("  Method: %s, Status codes: %s, Follow redirects: %v\n", method, statusCodes, followRedirects)
	return nil
}

// stepProbeRegions handles region selection.
func stepProbeRegions(ctx context.Context, wiz *interactive.Wizard) error {
	apiClient, ok := wiz.GetData("apiClient").(*client.Client)
	if !ok || apiClient == nil {
		return fmt.Errorf("internal error: API client not available")
	}

	// Fetch available regions
	fmt.Print("  Fetching available regions...")
	reqCtx, cancel := context.WithTimeout(ctx, probeWizardTimeout)
	defer cancel()

	regions, err := client.GetAllRegionsFlat(reqCtx, apiClient)
	if err != nil {
		fmt.Println(" failed")
		return fmt.Errorf("failed to fetch regions: %w", err)
	}
	fmt.Println(" done")

	if len(regions) == 0 {
		fmt.Println("  No regions available. Using default (all regions).")
		wiz.SetData("regions", []string{})
		return nil
	}

	// Build options list
	options := make([]string, len(regions))
	regionMap := make(map[string]string) // display -> ID
	for i, r := range regions {
		display := fmt.Sprintf("%s (%s)", r.DisplayName, r.ID)
		options[i] = display
		regionMap[display] = r.ID
	}

	// Ask for selection
	useAll, err := cliinteractive.Confirm(
		"Monitor from all available regions?",
		cliinteractive.WithDefault(true),
		cliinteractive.WithHelp("Monitoring from multiple regions provides better coverage and detects regional outages"),
	)
	if err != nil {
		return err
	}

	if useAll {
		wiz.SetData("regions", []string{})
		fmt.Printf("  Using all %d regions\n", len(regions))
		return nil
	}

	// Multi-select specific regions
	selected, err := interactive.AskMultiSelect(&interactive.MultiSelectPromptOptions{
		Message:  "Select regions to monitor from:",
		Options:  options,
		Defaults: options[:min(3, len(options))], // Default to first 3
		PageSize: 10,
		Validate: func(selections []string) error {
			if len(selections) == 0 {
				return fmt.Errorf("select at least one region")
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	// Convert to region IDs
	regionIDs := make([]string, len(selected))
	for i, s := range selected {
		regionIDs[i] = regionMap[s]
	}

	wiz.SetData("regions", regionIDs)
	fmt.Printf("  Selected %d regions: %s\n", len(regionIDs), strings.Join(regionIDs, ", "))
	return nil
}

// stepProbeInterval handles interval selection.
func stepProbeInterval(ctx context.Context, wiz *interactive.Wizard) error {
	intervals := []string{
		"30 seconds",
		"1 minute",
		"2 minutes",
		"5 minutes",
		"10 minutes",
	}

	selected, err := interactive.AskSelect(&interactive.SelectPromptOptions{
		Message:  "How often should we check your endpoint?",
		Options:  intervals,
		Default:  "1 minute",
		PageSize: 5,
	})
	if err != nil {
		return err
	}

	// Parse interval to seconds
	var intervalSeconds int
	switch selected {
	case "30 seconds":
		intervalSeconds = 30
	case "1 minute":
		intervalSeconds = 60
	case "2 minutes":
		intervalSeconds = 120
	case "5 minutes":
		intervalSeconds = 300
	case "10 minutes":
		intervalSeconds = 600
	}

	wiz.SetData("intervalSeconds", intervalSeconds)

	// Timeout selection
	timeout, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Request timeout in seconds:",
		Default: "10",
		Help:    "Maximum time to wait for a response (1-60 seconds)",
		Validate: func(s string) error {
			var t int
			if _, err := fmt.Sscanf(s, "%d", &t); err != nil {
				return fmt.Errorf("must be a number")
			}
			if t < 1 || t > 60 {
				return fmt.Errorf("timeout must be between 1 and 60 seconds")
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	var timeoutSeconds int
	_, _ = fmt.Sscanf(timeout, "%d", &timeoutSeconds)
	wiz.SetData("timeoutSeconds", timeoutSeconds)

	fmt.Printf("  Interval: %s, Timeout: %d seconds\n", selected, timeoutSeconds)
	return nil
}

// stepSSLMonitoring handles SSL certificate monitoring configuration.
func stepSSLMonitoring(ctx context.Context, wiz *interactive.Wizard) error {
	enableSSL, err := cliinteractive.Confirm(
		"Enable SSL certificate monitoring?",
		cliinteractive.WithDefault(true),
		cliinteractive.WithHelp("Monitor SSL certificate expiration and get alerts before it expires"),
	)
	if err != nil {
		return err
	}

	wiz.SetData("sslEnabled", enableSSL)

	if !enableSSL {
		wiz.SetData("sslThresholdDays", 14)
		return nil
	}

	// SSL expiry threshold
	thresholds := []string{
		"7 days",
		"14 days",
		"30 days",
		"60 days",
		"90 days",
	}

	selected, err := interactive.AskSelect(&interactive.SelectPromptOptions{
		Message: "Alert when certificate expires within:",
		Options: thresholds,
		Default: "14 days",
	})
	if err != nil {
		return err
	}

	var thresholdDays int
	switch selected {
	case "7 days":
		thresholdDays = 7
	case "14 days":
		thresholdDays = 14
	case "30 days":
		thresholdDays = 30
	case "60 days":
		thresholdDays = 60
	case "90 days":
		thresholdDays = 90
	}

	wiz.SetData("sslThresholdDays", thresholdDays)
	fmt.Printf("  SSL monitoring enabled, alert threshold: %d days\n", thresholdDays)
	return nil
}

// stepContentValidation handles optional content validation.
func stepContentValidation(ctx context.Context, wiz *interactive.Wizard) error {
	enableValidation, err := cliinteractive.Confirm(
		"Add content validation?",
		cliinteractive.WithHelp("Verify the response contains specific content or matches a JSONPath expression"),
	)
	if err != nil {
		return err
	}

	if !enableValidation {
		return nil
	}

	// Validation type
	validationTypes := []string{
		"Keyword - Check if response contains text",
		"JSONPath - Validate JSON response value",
	}

	selected, err := interactive.AskSelect(&interactive.SelectPromptOptions{
		Message: "Validation type:",
		Options: validationTypes,
		Default: validationTypes[0],
	})
	if err != nil {
		return err
	}

	if strings.HasPrefix(selected, "Keyword") {
		// Keyword validation
		keyword, err := interactive.AskString(&interactive.StringPromptOptions{
			Message: "Keyword to search for:",
			Help:    "Text that must appear in the response body",
			Validate: func(s string) error {
				if s == "" {
					return fmt.Errorf("keyword cannot be empty")
				}
				return nil
			},
		})
		if err != nil {
			return err
		}

		checkTypes := []string{"contains", "not_contains"}
		checkType, err := interactive.AskSelect(&interactive.SelectPromptOptions{
			Message: "Check type:",
			Options: checkTypes,
			Default: "contains",
		})
		if err != nil {
			return err
		}

		wiz.SetData("keywordCheck", keyword)
		wiz.SetData("keywordCheckType", checkType)
		fmt.Printf("  Keyword validation: response %s \"%s\"\n", checkType, keyword)
	} else {
		// JSONPath validation
		jsonPath, err := interactive.AskString(&interactive.StringPromptOptions{
			Message: "JSONPath expression:",
			Default: "$.status",
			Help:    "JSONPath expression to evaluate (e.g., $.status, $.data.count)",
			Validate: func(s string) error {
				if s == "" {
					return fmt.Errorf("JSONPath cannot be empty")
				}
				if !strings.HasPrefix(s, "$") {
					return fmt.Errorf("JSONPath must start with $")
				}
				return nil
			},
		})
		if err != nil {
			return err
		}

		expected, err := interactive.AskString(&interactive.StringPromptOptions{
			Message: "Expected value:",
			Help:    "Value that the JSONPath expression should return",
		})
		if err != nil {
			return err
		}

		wiz.SetData("jsonPathCheck", jsonPath)
		wiz.SetData("jsonPathExpected", expected)
		fmt.Printf("  JSONPath validation: %s = \"%s\"\n", jsonPath, expected)
	}

	return nil
}

// stepProbeName handles probe naming.
func stepProbeName(ctx context.Context, wiz *interactive.Wizard) error {
	target := wiz.GetDataString("target")

	// Generate default name from target
	defaultName := generateProbeName(target)

	name, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Probe name:",
		Default: defaultName,
		Help:    "A friendly name to identify this probe",
		Validate: func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("name cannot be empty")
			}
			if len(s) > 100 {
				return fmt.Errorf("name must be less than 100 characters")
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	wiz.SetData("probeName", name)
	return nil
}

// stepTestProbe offers to test the probe configuration before creating.
func stepTestProbe(ctx context.Context, wiz *interactive.Wizard) error {
	runTest, err := cliinteractive.Confirm(
		"Test the probe configuration before creating?",
		cliinteractive.WithDefault(true),
		cliinteractive.WithHelp("Run an ad-hoc test to verify the configuration works"),
	)
	if err != nil {
		return err
	}

	if !runTest {
		fmt.Println("  Skipping test.")
		return nil
	}

	// Build test request from wizard data
	testReq := buildProbeTestRequestFromWizard(wiz)

	apiClient, ok := wiz.GetData("apiClient").(*client.Client)
	if !ok || apiClient == nil {
		return fmt.Errorf("internal error: API client not available")
	}

	fmt.Print("  Running test check...")
	reqCtx, cancel := context.WithTimeout(ctx, probeWizardTimeout)
	defer cancel()

	result, err := client.TestProbe(reqCtx, apiClient, testReq)
	if err != nil {
		fmt.Println(" failed")
		fmt.Printf("  Error: %v\n", err)

		// Ask if user wants to continue anyway
		continueAnyway, err := cliinteractive.Confirm("Test failed. Continue creating the probe anyway?")
		if err != nil {
			return err
		}
		if !continueAnyway {
			return fmt.Errorf("test failed and user chose not to continue")
		}
		wiz.SetData("testPassed", false)
		return nil
	}

	fmt.Println(" done")
	fmt.Printf("  Status: %s\n", result.Status)
	fmt.Printf("  Response time: %d ms\n", result.ResponseTimeMs)
	if result.StatusCode != nil {
		fmt.Printf("  Status code: %d\n", *result.StatusCode)
	}
	if result.SSLExpiryDays != nil {
		fmt.Printf("  SSL expires in: %d days\n", *result.SSLExpiryDays)
	}

	wiz.SetData("testPassed", result.Status == "up")
	wiz.SetData("testResult", result)

	if result.Status != "up" {
		fmt.Printf("  Warning: Test returned status '%s'\n", result.Status)
		if result.ErrorMessage != nil {
			fmt.Printf("  Error: %s\n", *result.ErrorMessage)
		}

		continueAnyway, err := cliinteractive.Confirm("Test did not return 'up' status. Continue creating the probe?")
		if err != nil {
			return err
		}
		if !continueAnyway {
			return fmt.Errorf("test did not pass and user chose not to continue")
		}
	}

	return nil
}

// stepCreateProbe creates the probe.
func stepCreateProbe(ctx context.Context, wiz *interactive.Wizard) error {
	// Print summary
	fmt.Println("\n  Probe Configuration Summary:")
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Name:     %s\n", wiz.GetDataString("probeName"))
	fmt.Printf("  Type:     %s\n", wiz.GetDataString("checkType"))
	fmt.Printf("  Target:   %s\n", wiz.GetDataString("target"))
	fmt.Printf("  Interval: %d seconds\n", wiz.GetDataInt("intervalSeconds"))
	fmt.Printf("  Timeout:  %d seconds\n", wiz.GetDataInt("timeoutSeconds"))

	if wiz.GetDataString("checkType") == "http" {
		fmt.Printf("  Method:   %s\n", wiz.GetDataString("method"))
	}

	if regions := wiz.GetData("regions"); regions != nil {
		if r, ok := regions.([]string); ok && len(r) > 0 {
			fmt.Printf("  Regions:  %s\n", strings.Join(r, ", "))
		} else {
			fmt.Println("  Regions:  All available")
		}
	}

	if wiz.GetDataBool("sslEnabled") {
		fmt.Printf("  SSL:      Enabled (alert at %d days)\n", wiz.GetDataInt("sslThresholdDays"))
	}

	fmt.Println()

	// Confirm creation
	confirm, err := cliinteractive.Confirm(
		"Create this probe?",
		cliinteractive.WithDefault(true),
	)
	if err != nil {
		return err
	}

	if !confirm {
		return fmt.Errorf("probe creation cancelled by user")
	}

	// Build create request
	req := buildCreateProbeRequestFromWizard(wiz)

	apiClient, ok := wiz.GetData("apiClient").(*client.Client)
	if !ok || apiClient == nil {
		return fmt.Errorf("internal error: API client not available")
	}

	fmt.Print("\n  Creating probe...")
	reqCtx, cancel := context.WithTimeout(ctx, probeWizardTimeout)
	defer cancel()

	probe, err := client.CreateProbe(reqCtx, apiClient, req)
	if err != nil {
		fmt.Println(" failed")
		return fmt.Errorf("failed to create probe: %w", err)
	}

	fmt.Println(" done")

	wiz.SetData("createdProbe", probe)
	wiz.SetData("probeCreated", true)

	return nil
}

// Helper functions

// skipIfNotHTTP returns true if the check type is not HTTP.
func skipIfNotHTTP(wiz *interactive.Wizard) bool {
	return wiz.GetDataString("checkType") != "http"
}

// skipIfNotHTTPS returns true if the target doesn't use HTTPS.
func skipIfNotHTTPS(wiz *interactive.Wizard) bool {
	target := wiz.GetDataString("target")
	return !strings.HasPrefix(strings.ToLower(target), "https://")
}

// generateProbeName generates a default probe name from the target.
func generateProbeName(target string) string {
	// Remove scheme
	name := strings.TrimPrefix(target, "https://")
	name = strings.TrimPrefix(name, "http://")

	// Remove port
	if idx := strings.Index(name, ":"); idx > 0 {
		name = name[:idx]
	}

	// Remove path
	if idx := strings.Index(name, "/"); idx > 0 {
		name = name[:idx]
	}

	// Capitalize first letter
	if len(name) > 0 {
		name = strings.ToUpper(string(name[0])) + name[1:]
	}

	return name
}

// buildProbeTestRequestFromWizard creates a test request from wizard data.
func buildProbeTestRequestFromWizard(wiz *interactive.Wizard) *client.ProbeTestRequest {
	checkType := wiz.GetDataString("checkType")
	target := wiz.GetDataString("target")
	timeoutSeconds := wiz.GetDataInt("timeoutSeconds")
	if timeoutSeconds == 0 {
		timeoutSeconds = 10
	}

	req := &client.ProbeTestRequest{
		URL:       target,
		CheckType: client.CheckType(checkType),
		TimeoutMs: timeoutSeconds * 1000,
	}

	if checkType == "http" {
		method := wiz.GetDataString("method")
		if method == "" {
			method = "GET"
		}
		req.Method = method

		statusCodesStr := wiz.GetDataString("statusCodes")
		if statusCodesStr != "" {
			codes, _ := parseStatusCodes(statusCodesStr)
			req.ExpectedStatusCodes = codes
		} else {
			req.ExpectedStatusCodes = []int{200}
		}

		followRedirects := wiz.GetDataBool("followRedirects")
		req.FollowRedirects = &followRedirects

		if keyword := wiz.GetDataString("keywordCheck"); keyword != "" {
			req.KeywordCheck = &keyword
			kwCheckType := wiz.GetDataString("keywordCheckType")
			req.KeywordCheckType = &kwCheckType
		}

		if jsonPath := wiz.GetDataString("jsonPathCheck"); jsonPath != "" {
			req.JSONPathCheck = &jsonPath
			if expected := wiz.GetDataString("jsonPathExpected"); expected != "" {
				req.JSONPathExpected = &expected
			}
		}
	}

	return req
}

// buildCreateProbeRequestFromWizard creates a probe create request from wizard data.
func buildCreateProbeRequestFromWizard(wiz *interactive.Wizard) *client.CreateProbeRequest {
	checkType := wiz.GetDataString("checkType")
	target := wiz.GetDataString("target")
	name := wiz.GetDataString("probeName")
	intervalSeconds := wiz.GetDataInt("intervalSeconds")
	timeoutSeconds := wiz.GetDataInt("timeoutSeconds")

	if intervalSeconds == 0 {
		intervalSeconds = 60
	}
	if timeoutSeconds == 0 {
		timeoutSeconds = 10
	}

	req := &client.CreateProbeRequest{
		Name:            name,
		URL:             target,
		CheckType:       client.CheckType(checkType),
		IntervalSeconds: intervalSeconds,
		TimeoutMs:       timeoutSeconds * 1000,
	}

	// Regions
	if regions := wiz.GetData("regions"); regions != nil {
		if r, ok := regions.([]string); ok && len(r) > 0 {
			req.Regions = r
		}
	}

	// HTTP-specific settings
	if checkType == "http" {
		method := wiz.GetDataString("method")
		if method == "" {
			method = "GET"
		}
		req.Method = method

		statusCodesStr := wiz.GetDataString("statusCodes")
		if statusCodesStr != "" {
			codes, _ := parseStatusCodes(statusCodesStr)
			req.ExpectedStatusCodes = codes
		} else {
			req.ExpectedStatusCodes = []int{200}
		}

		followRedirects := wiz.GetDataBool("followRedirects")
		req.FollowRedirects = &followRedirects
		req.MaxRedirects = 10

		// Content validation
		if keyword := wiz.GetDataString("keywordCheck"); keyword != "" {
			req.KeywordCheck = &keyword
			kwCheckType := wiz.GetDataString("keywordCheckType")
			req.KeywordCheckType = &kwCheckType
		}

		if jsonPath := wiz.GetDataString("jsonPathCheck"); jsonPath != "" {
			req.JSONPathCheck = &jsonPath
			if expected := wiz.GetDataString("jsonPathExpected"); expected != "" {
				req.JSONPathExpected = &expected
			}
		}
	}

	// SSL settings
	if strings.HasPrefix(strings.ToLower(target), "https://") {
		req.SSLCheckEnabled = wiz.GetDataBool("sslEnabled")
		if req.SSLCheckEnabled {
			threshold := wiz.GetDataInt("sslThresholdDays")
			if threshold == 0 {
				threshold = 14
			}
			req.SSLExpiryThresholdDays = threshold
		}
	}

	return req
}

// printProbeWizardSummary prints the final summary after probe creation.
func printProbeWizardSummary(wiz *interactive.Wizard) {
	if !wiz.GetDataBool("probeCreated") {
		return
	}

	probe := wiz.GetData("createdProbe")
	if probe == nil {
		return
	}

	p, ok := probe.(*client.Probe)
	if !ok {
		return
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Probe Created Successfully!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("Name:     %s\n", p.Name)
	fmt.Printf("ID:       %s\n", p.ID)
	fmt.Printf("URL:      %s\n", p.URL)
	fmt.Printf("Status:   %s\n", p.Status)
	fmt.Println()
	fmt.Println("What's next?")
	fmt.Println("  stackeye probe list        View all your probes")
	fmt.Println("  stackeye probe get <id>    View probe details")
	fmt.Println("  stackeye channel create    Set up notifications")
	fmt.Println("  stackeye alert list        View active alerts")
	fmt.Println()

	// Also print using configured output format if not table
	output.Print(p)
}
