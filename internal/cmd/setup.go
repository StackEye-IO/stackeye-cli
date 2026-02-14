// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/auth"
	cliinteractive "github.com/StackEye-IO/stackeye-cli/internal/interactive"
	sdkauth "github.com/StackEye-IO/stackeye-go-sdk/auth"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/spf13/cobra"
)

// setupFlags holds the command flags for the setup command.
type setupFlags struct {
	apiURL        string
	skipProbe     bool
	apiKey        string // API key for non-interactive authentication
	probeName     string // Probe name for non-interactive creation
	probeURL      string // Probe URL for non-interactive creation
	probeInterval int    // Probe interval in seconds for non-interactive creation
}

// NewSetupCmd creates and returns the setup command.
func NewSetupCmd() *cobra.Command {
	flags := &setupFlags{}

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard for first-time users",
		Long: `Interactive setup wizard that guides you through initial StackEye CLI configuration.

The setup wizard helps first-time users get started quickly by:
  1. Authenticating with your StackEye account (if not already logged in)
  2. Selecting an organization (if you belong to multiple)
  3. Optionally creating your first monitoring probe

For scripted or non-interactive environments, use the individual commands:
  stackeye login              Authenticate with StackEye
  stackeye org list           List your organizations
  stackeye org switch <name>  Switch to a different organization
  stackeye probe create       Create a monitoring probe

CI/CD and Automation:
  For non-interactive environments, you can provide an API key directly:

  # Via environment variable (recommended for CI/CD)
  export STACKEYE_API_KEY=se_xxx
  stackeye setup --no-input

  # Via flag (single-command setup)
  stackeye setup --no-input --api-key se_xxx

  # Full setup with probe creation
  stackeye setup --no-input --api-key se_xxx --probe-name "API" --probe-url https://api.example.com

Examples:
  # Run the interactive setup wizard
  stackeye setup

  # Run setup with a specific API URL (for development environments)
  stackeye setup --api-url https://api.dev.stackeye.io

  # Skip the probe creation step
  stackeye setup --skip-probe

  # Non-interactive setup with API key from environment
  STACKEYE_API_KEY=se_xxx stackeye setup --no-input

  # Non-interactive setup with API key and probe creation
  stackeye setup --no-input --api-key se_xxx --probe-name "My API" --probe-url https://example.com`,
		// Override PersistentPreRunE to skip config loading.
		// The setup command should work without a valid configuration.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVar(&flags.apiURL, "api-url", auth.DefaultAPIURL, "StackEye API URL")
	cmd.Flags().BoolVar(&flags.skipProbe, "skip-probe", false, "Skip the optional probe creation step")

	// CI/CD and automation flags
	cmd.Flags().StringVar(&flags.apiKey, "api-key", "", "API key for non-interactive authentication (or use STACKEYE_API_KEY env var)")
	cmd.Flags().StringVar(&flags.probeName, "probe-name", "", "Probe name for non-interactive probe creation (requires --probe-url)")
	cmd.Flags().StringVar(&flags.probeURL, "probe-url", "", "Probe URL for non-interactive probe creation (requires --probe-name)")
	cmd.Flags().IntVar(&flags.probeInterval, "probe-interval", 60, "Probe check interval in seconds (default: 60)")

	return cmd
}

// runSetup executes the setup wizard.
func runSetup(ctx context.Context, flags *setupFlags) error {
	// Check if non-interactive mode is enabled
	if GetNoInput() {
		return runSetupNonInteractive(ctx, flags)
	}

	// Create the wizard
	wiz := interactive.NewWizard(&interactive.WizardOptions{
		Title:         "StackEye Setup Wizard",
		Description:   "Let's get your CLI configured for monitoring",
		ShowProgress:  true,
		ConfirmCancel: true,
	})

	// Store flags in wizard data for steps to access
	wiz.SetData("apiURL", flags.apiURL)
	wiz.SetData("skipProbe", flags.skipProbe)

	// Add wizard steps
	wiz.AddSteps(
		&interactive.Step{
			Name:        "authentication",
			Title:       "Authentication",
			Description: "Sign in to your StackEye account",
			Run:         stepAuthentication,
			Skip:        skipIfAuthenticated,
		},
		&interactive.Step{
			Name:        "organization",
			Title:       "Organization Selection",
			Description: "Choose which organization to work with",
			Run:         stepOrganization,
			Skip:        skipIfSingleOrg,
		},
		&interactive.Step{
			Name:        "first-probe",
			Title:       "Create First Probe",
			Description: "Set up your first monitoring probe (optional)",
			Run:         stepFirstProbe,
			Skip:        skipProbeStep,
		},
	)

	// Run the wizard
	if err := wiz.Run(ctx); err != nil {
		if errors.Is(err, interactive.ErrWizardCancelled) {
			fmt.Println("\nSetup cancelled. You can run 'stackeye setup' again at any time.")
			return nil
		}
		return err
	}

	// Print final summary
	printSetupSummary(wiz)

	return nil
}

// runSetupNonInteractive handles setup in non-interactive mode.
// Supports API key authentication via --api-key flag or STACKEYE_API_KEY env var.
func runSetupNonInteractive(ctx context.Context, flags *setupFlags) error {
	// Check for API key from flag or environment variable
	apiKey := getAPIKeyFromSources(flags)

	// If API key provided, perform non-interactive authentication setup
	if apiKey != "" {
		return runNonInteractiveAuth(ctx, flags, apiKey)
	}

	// No API key provided - fall back to status checking behavior
	return runNonInteractiveStatus(ctx, flags)
}

// getAPIKeyFromSources returns the API key from flag or environment variable.
// Precedence: --api-key flag > STACKEYE_API_KEY env var
func getAPIKeyFromSources(flags *setupFlags) string {
	// Check flag first (highest priority)
	if flags.apiKey != "" {
		return flags.apiKey
	}

	// Check environment variable
	return os.Getenv(sdkauth.APIKeyEnvVar)
}

// runNonInteractiveAuth performs non-interactive setup with the provided API key.
func runNonInteractiveAuth(ctx context.Context, flags *setupFlags, apiKey string) error {
	// Validate API key format
	if !sdkauth.ValidateAPIKey(apiKey) {
		return fmt.Errorf("invalid API key format: must be 67 characters starting with 'se_'")
	}

	// Verify credentials with API
	fmt.Print("Verifying API key...")
	c := client.New(apiKey, flags.apiURL)
	verifyCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	verifyResp, err := client.VerifyCLICredentials(verifyCtx, c)
	if err != nil {
		fmt.Println(" failed")
		return fmt.Errorf("API key verification failed: %w", err)
	}
	fmt.Println(" done")

	// Get organization info
	orgName := verifyResp.OrganizationName
	orgID := verifyResp.OrganizationID
	if orgName == "" {
		orgName = "default"
	}

	// Generate context name
	contextName := generateContextName(orgName, flags.apiURL)

	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		// Create new config if none exists
		cfg = config.NewConfig()
	}

	// Create new context
	newCtx := &config.Context{
		APIURL:           flags.apiURL,
		OrganizationID:   orgID,
		OrganizationName: orgName,
		APIKey:           apiKey,
	}

	// Handle naming conflicts
	existingContextName := contextName
	suffix := 1
	for {
		if _, err := cfg.GetContext(contextName); err != nil {
			break
		}
		// Check if existing context has same API URL - update it instead
		existingCtx, _ := cfg.GetContext(contextName)
		if existingCtx != nil && existingCtx.APIURL == flags.apiURL {
			// Update existing context with new credentials
			existingCtx.APIKey = apiKey
			existingCtx.OrganizationID = orgID
			existingCtx.OrganizationName = orgName
			cfg.CurrentContext = contextName
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Println()
			fmt.Printf("Context:      %s (updated)\n", contextName)
			fmt.Printf("Organization: %s\n", orgName)
			fmt.Printf("API URL:      %s\n", flags.apiURL)

			// Create probe if requested
			if flags.probeName != "" && flags.probeURL != "" {
				return createNonInteractiveProbe(ctx, flags, apiKey)
			}
			return nil
		}
		suffix++
		contextName = fmt.Sprintf("%s-%d", existingContextName, suffix)
	}

	// Save new context
	cfg.SetContext(contextName, newCtx)
	cfg.CurrentContext = contextName

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("Context:      %s\n", contextName)
	fmt.Printf("Organization: %s\n", orgName)
	fmt.Printf("API URL:      %s\n", flags.apiURL)

	// Create probe if requested
	if flags.probeName != "" && flags.probeURL != "" {
		return createNonInteractiveProbe(ctx, flags, apiKey)
	}

	// Print next steps if no probe created
	if !flags.skipProbe && flags.probeName == "" {
		fmt.Println()
		fmt.Println("To create a probe:")
		fmt.Println("  stackeye probe create --name <name> --url <url>")
	}

	return nil
}

// createNonInteractiveProbe creates a probe without interactive prompts.
func createNonInteractiveProbe(ctx context.Context, flags *setupFlags, apiKey string) error {
	// Validate required flags
	if flags.probeName == "" {
		return fmt.Errorf("--probe-name is required for non-interactive probe creation")
	}
	if flags.probeURL == "" {
		return fmt.Errorf("--probe-url is required for non-interactive probe creation")
	}

	// Validate probe URL
	if err := validateProbeURL(flags.probeURL); err != nil {
		return fmt.Errorf("invalid probe URL: %w", err)
	}

	// Validate interval
	intervalSeconds := flags.probeInterval
	if intervalSeconds < 15 {
		fmt.Printf("Note: Interval adjusted from %d to minimum 15 seconds\n", flags.probeInterval)
		intervalSeconds = 15
	}
	if intervalSeconds > 3600 {
		fmt.Printf("Note: Interval adjusted from %d to maximum 3600 seconds\n", flags.probeInterval)
		intervalSeconds = 3600
	}

	fmt.Print("\nCreating probe...")

	c := client.New(apiKey, flags.apiURL)
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req := &client.CreateProbeRequest{
		Name:                   flags.probeName,
		URL:                    flags.probeURL,
		CheckType:              client.CheckTypeHTTP,
		Method:                 "GET",
		IntervalSeconds:        intervalSeconds,
		TimeoutMs:              10000, // 10 seconds
		ExpectedStatusCodes:    []int{200},
		SSLCheckEnabled:        true,
		SSLExpiryThresholdDays: 14,
	}

	probe, err := client.CreateProbe(createCtx, c, req)
	if err != nil {
		fmt.Println(" failed")
		return fmt.Errorf("failed to create probe: %w", err)
	}

	fmt.Println(" done")
	fmt.Printf("Probe created: %s\n", probe.Name)
	fmt.Printf("  ID:       %s\n", probe.ID)
	fmt.Printf("  URL:      %s\n", flags.probeURL)
	fmt.Printf("  Interval: %d seconds\n", intervalSeconds)

	return nil
}

// runNonInteractiveStatus checks and reports the current configuration status.
// This is the original behavior when no API key is provided.
func runNonInteractiveStatus(ctx context.Context, flags *setupFlags) error {
	// Check if already authenticated
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Status: Not configured")
		fmt.Println()
		fmt.Println("To set up in non-interactive mode:")
		fmt.Println("  # Option 1: Use environment variable (recommended for CI/CD)")
		fmt.Println("  export STACKEYE_API_KEY=se_xxx")
		fmt.Println("  stackeye setup --no-input")
		fmt.Println()
		fmt.Println("  # Option 2: Use command flag")
		fmt.Println("  stackeye setup --no-input --api-key se_xxx")
		fmt.Println()
		fmt.Println("  # Option 3: Interactive browser login")
		fmt.Println("  stackeye login")
		return nil
	}

	// Check for valid context
	currentCtx, err := cfg.GetCurrentContext()
	if err != nil || currentCtx.APIKey == "" {
		fmt.Println("Status: Not authenticated")
		fmt.Println()
		fmt.Println("To set up in non-interactive mode:")
		fmt.Println("  # Option 1: Use environment variable (recommended for CI/CD)")
		fmt.Println("  export STACKEYE_API_KEY=se_xxx")
		fmt.Println("  stackeye setup --no-input")
		fmt.Println()
		fmt.Println("  # Option 2: Use command flag")
		fmt.Println("  stackeye setup --no-input --api-key se_xxx")
		fmt.Println()
		fmt.Println("  # Option 3: Interactive browser login")
		fmt.Println("  stackeye login")
		return nil
	}

	// Verify credentials
	c := client.New(currentCtx.APIKey, currentCtx.APIURL)
	verifyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err = client.VerifyCLICredentials(verifyCtx, c)
	if err != nil {
		fmt.Println("Status: Authentication invalid or expired")
		fmt.Println()
		fmt.Println("Re-authenticate with: stackeye login")
		fmt.Println("Or provide a new API key: stackeye setup --no-input --api-key se_xxx")
		return nil
	}

	// Configuration is valid
	fmt.Println("Status: Configured and ready")
	fmt.Println()
	fmt.Printf("Context:      %s\n", cfg.CurrentContext)
	fmt.Printf("Organization: %s\n", currentCtx.OrganizationName)
	fmt.Printf("API URL:      %s\n", currentCtx.APIURL)
	fmt.Println()

	if !flags.skipProbe {
		fmt.Println("To create a probe:")
		fmt.Println("  stackeye probe create --name <name> --url <url>")
	}

	return nil
}

// stepAuthentication handles the authentication step.
func stepAuthentication(ctx context.Context, wiz *interactive.Wizard) error {
	apiURL := wiz.GetDataString("apiURL")
	if apiURL == "" {
		apiURL = auth.DefaultAPIURL
	}

	fmt.Println("Opening your browser for authentication...")
	fmt.Println()

	// Perform browser-based login
	result, err := auth.BrowserLogin(auth.Options{
		APIURL:  apiURL,
		Timeout: auth.DefaultTimeout,
		OnBrowserOpen: func(url string) {
			fmt.Printf("  URL: %s\n", url)
		},
		OnWaiting: func() {
			fmt.Println("  Waiting for authentication...")
			fmt.Println("  (If the browser doesn't open, visit the URL manually)")
			fmt.Println()
		},
	})
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Verify credentials
	fmt.Print("Verifying credentials...")
	c := client.New(result.APIKey, apiURL)
	verifyCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	verifyResp, err := client.VerifyCLICredentials(verifyCtx, c)
	if err != nil {
		fmt.Println(" failed")
		return fmt.Errorf("failed to verify credentials: %w", err)
	}
	fmt.Println(" done")

	// Use org info from verify response if not in callback
	orgName := result.OrgName
	orgID := result.OrgID
	if orgName == "" && verifyResp.OrganizationName != "" {
		orgName = verifyResp.OrganizationName
	}
	if orgID == "" && verifyResp.OrganizationID != "" {
		orgID = verifyResp.OrganizationID
	}
	if orgName == "" {
		orgName = "default"
	}

	// Generate context name
	contextName := generateContextName(orgName, apiURL)

	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create new context
	newCtx := &config.Context{
		APIURL:           apiURL,
		OrganizationID:   orgID,
		OrganizationName: orgName,
		APIKey:           result.APIKey,
	}

	// Handle naming conflicts
	existingContextName := contextName
	suffix := 1
	for {
		if _, err := cfg.GetContext(contextName); err != nil {
			break
		}
		suffix++
		contextName = fmt.Sprintf("%s-%d", existingContextName, suffix)
	}

	// Save context
	cfg.SetContext(contextName, newCtx)
	cfg.CurrentContext = contextName

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Store auth info in wizard for later steps
	wiz.SetData("authenticated", true)
	wiz.SetData("apiKey", result.APIKey)
	wiz.SetData("orgID", orgID)
	wiz.SetData("orgName", orgName)
	wiz.SetData("contextName", contextName)

	fmt.Printf("\n  Logged in as: %s\n", orgName)
	fmt.Printf("  Context: %s\n", contextName)

	return nil
}

// skipIfAuthenticated checks if user is already authenticated.
func skipIfAuthenticated(wiz *interactive.Wizard) bool {
	apiURL := wiz.GetDataString("apiURL")
	if apiURL == "" {
		apiURL = auth.DefaultAPIURL
	}

	cfg, err := config.Load()
	if err != nil {
		return false // Not authenticated, don't skip
	}

	// Check if any context uses this API URL and has valid credentials
	for name, ctx := range cfg.Contexts {
		if ctx.APIURL == apiURL && ctx.APIKey != "" {
			// Store existing auth info
			wiz.SetData("authenticated", true)
			wiz.SetData("apiKey", ctx.APIKey)
			wiz.SetData("orgID", ctx.OrganizationID)
			wiz.SetData("orgName", ctx.OrganizationName)
			wiz.SetData("contextName", name)

			fmt.Printf("Already authenticated as: %s (context: %s)\n\n", ctx.OrganizationName, name)
			return true
		}
	}

	return false
}

// stepOrganization handles organization selection.
func stepOrganization(ctx context.Context, wiz *interactive.Wizard) error {
	apiKey := wiz.GetDataString("apiKey")
	apiURL := wiz.GetDataString("apiURL")
	if apiURL == "" {
		apiURL = auth.DefaultAPIURL
	}

	// Get API client
	c := client.New(apiKey, apiURL)

	// List organizations
	listCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	orgsResp, err := client.ListOrganizations(listCtx, c)
	if err != nil {
		return fmt.Errorf("failed to list organizations: %w", err)
	}

	if len(orgsResp.Organizations) == 0 {
		fmt.Println("  No organizations found. You may need to create one via the web UI.")
		return nil
	}

	if len(orgsResp.Organizations) == 1 {
		org := orgsResp.Organizations[0]
		wiz.SetData("selectedOrgID", org.ID)
		wiz.SetData("selectedOrgName", org.Name)
		fmt.Printf("  Using organization: %s\n", org.Name)
		return nil
	}

	// Multiple orgs - let user select
	options := make([]string, len(orgsResp.Organizations))
	orgMap := make(map[string]client.Organization)
	var currentOrgName string

	for i, org := range orgsResp.Organizations {
		label := org.Name
		if org.IsCurrent {
			label = fmt.Sprintf("%s (current)", org.Name)
			currentOrgName = org.Name
		}
		options[i] = label
		orgMap[label] = org
	}

	selected, err := interactive.AskSelect(&interactive.SelectPromptOptions{
		Message:  "Select an organization:",
		Options:  options,
		Default:  currentOrgName,
		PageSize: 10,
	})
	if err != nil {
		return err
	}

	// Find the selected org
	selectedOrg := orgMap[selected]
	wiz.SetData("selectedOrgID", selectedOrg.ID)
	wiz.SetData("selectedOrgName", selectedOrg.Name)

	// Update config if different from current
	if !selectedOrg.IsCurrent {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		currentCtx, err := cfg.GetCurrentContext()
		if err == nil {
			currentCtx.OrganizationID = selectedOrg.ID
			currentCtx.OrganizationName = selectedOrg.Name
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
		}
	}

	fmt.Printf("\n  Selected: %s\n", selectedOrg.Name)
	return nil
}

// skipIfSingleOrg checks if user has only one organization.
func skipIfSingleOrg(wiz *interactive.Wizard) bool {
	// If we haven't authenticated yet, don't skip (we'll handle this in the step)
	if !wiz.GetDataBool("authenticated") {
		return false
	}

	apiKey := wiz.GetDataString("apiKey")
	apiURL := wiz.GetDataString("apiURL")
	if apiURL == "" {
		apiURL = auth.DefaultAPIURL
	}

	c := client.New(apiKey, apiURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orgsResp, err := client.ListOrganizations(ctx, c)
	if err != nil {
		return false // Don't skip on error
	}

	// Skip if only one org (already selected during auth)
	if len(orgsResp.Organizations) <= 1 {
		if len(orgsResp.Organizations) == 1 {
			org := orgsResp.Organizations[0]
			wiz.SetData("selectedOrgID", org.ID)
			wiz.SetData("selectedOrgName", org.Name)
		}
		return true
	}

	return false
}

// stepFirstProbe handles optional first probe creation.
func stepFirstProbe(ctx context.Context, wiz *interactive.Wizard) error {
	// Ask if user wants to create a probe
	createProbe, err := cliinteractive.Confirm(
		"Would you like to create your first monitoring probe?",
		cliinteractive.WithDefault(true),
	)
	if err != nil {
		return err
	}

	if !createProbe {
		fmt.Println("  Skipping probe creation.")
		fmt.Println("  You can create probes later with: stackeye probe create")
		return nil
	}

	// Get probe name
	probeName, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "Probe name:",
		Default: "My First Probe",
		Help:    "A friendly name for your monitoring probe",
		Validate: func(s string) error {
			if len(s) < 1 {
				return fmt.Errorf("probe name cannot be empty")
			}
			if len(s) > 100 {
				return fmt.Errorf("probe name must be less than 100 characters")
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	// Get URL to monitor
	probeURL, err := interactive.AskString(&interactive.StringPromptOptions{
		Message: "URL to monitor:",
		Default: "",
		Help:    "The full URL (including https://) to check, e.g. https://yoursite.com",
		Validate: func(s string) error {
			if err := validateProbeURL(s); err != nil {
				return err
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	// Get interval
	intervalOptions := []string{
		"30 seconds",
		"1 minute",
		"2 minutes",
		"5 minutes",
	}
	intervalSelected, err := interactive.AskSelect(&interactive.SelectPromptOptions{
		Message: "Check interval:",
		Options: intervalOptions,
		Default: "1 minute",
	})
	if err != nil {
		return err
	}

	intervalSeconds := 60 // default
	switch intervalSelected {
	case "30 seconds":
		intervalSeconds = 30
	case "1 minute":
		intervalSeconds = 60
	case "2 minutes":
		intervalSeconds = 120
	case "5 minutes":
		intervalSeconds = 300
	}

	// Create the probe
	apiKey := wiz.GetDataString("apiKey")
	apiURL := wiz.GetDataString("apiURL")
	if apiURL == "" {
		apiURL = auth.DefaultAPIURL
	}

	c := client.New(apiKey, apiURL)
	createCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	fmt.Print("\n  Creating probe...")

	req := &client.CreateProbeRequest{
		Name:                   probeName,
		URL:                    probeURL,
		CheckType:              client.CheckTypeHTTP,
		Method:                 "GET",
		IntervalSeconds:        intervalSeconds,
		TimeoutMs:              10000, // 10 seconds
		ExpectedStatusCodes:    []int{200},
		SSLCheckEnabled:        true,
		SSLExpiryThresholdDays: 14,
	}

	probe, err := client.CreateProbe(createCtx, c, req)
	if err != nil {
		fmt.Println(" failed")
		return fmt.Errorf("failed to create probe: %w", err)
	}

	fmt.Println(" done")
	fmt.Printf("\n  Probe created: %s\n", probe.Name)
	fmt.Printf("  ID: %s\n", probe.ID)
	fmt.Printf("  Monitoring: %s\n", probeURL)
	fmt.Printf("  Interval: %d seconds\n", intervalSeconds)

	wiz.SetData("probeCreated", true)
	wiz.SetData("probeName", probe.Name)
	wiz.SetData("probeID", probe.ID.String())

	return nil
}

// skipProbeStep checks if probe creation should be skipped.
func skipProbeStep(wiz *interactive.Wizard) bool {
	return wiz.GetDataBool("skipProbe")
}

// printSetupSummary prints a summary of what was configured.
func printSetupSummary(wiz *interactive.Wizard) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Setup Complete!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Authentication info
	if contextName := wiz.GetDataString("contextName"); contextName != "" {
		fmt.Printf("Context:      %s\n", contextName)
	}
	if orgName := wiz.GetDataString("orgName"); orgName != "" {
		fmt.Printf("Organization: %s\n", orgName)
	}
	if selectedOrg := wiz.GetDataString("selectedOrgName"); selectedOrg != "" && selectedOrg != wiz.GetDataString("orgName") {
		fmt.Printf("Working Org:  %s\n", selectedOrg)
	}

	// Probe info
	if wiz.GetDataBool("probeCreated") {
		fmt.Println()
		fmt.Println("First Probe:")
		fmt.Printf("  Name: %s\n", wiz.GetDataString("probeName"))
		fmt.Printf("  ID:   %s\n", wiz.GetDataString("probeID"))
	}

	fmt.Println()
	fmt.Println("What's next?")
	fmt.Println("  stackeye probe list      View your probes")
	fmt.Println("  stackeye alert list      View active alerts")
	fmt.Println("  stackeye channel create  Set up notifications")
	fmt.Println("  stackeye --help          See all commands")
	fmt.Println()
}
