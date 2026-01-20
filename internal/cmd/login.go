// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/auth"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/spf13/cobra"
)

// loginFlags holds the command flags for the login command.
type loginFlags struct {
	apiURL string
	debug  bool
}

// NewLoginCmd creates and returns the login command.
func NewLoginCmd() *cobra.Command {
	flags := &loginFlags{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with your StackEye account",
		Long: `Authenticate with your StackEye account via the web browser.

This command opens your default web browser to the StackEye login page.
After you authenticate, an API key is generated and stored locally
for subsequent CLI operations.

The API key is stored in ~/.config/stackeye/config.yaml with secure
permissions (0600). Use 'stackeye logout' to remove the stored credentials.

Examples:
  # Login to production StackEye
  stackeye login

  # Login to a specific environment
  stackeye login --api-url https://api.dev.stackeye.io`,
		// Override PersistentPreRunE to skip config loading.
		// The login command should work without a valid configuration.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(flags)
		},
	}

	cmd.Flags().StringVar(&flags.apiURL, "api-url", auth.DefaultAPIURL, "StackEye API URL")
	cmd.Flags().BoolVar(&flags.debug, "debug", false, "Enable debug logging for troubleshooting")

	return cmd
}

// runLogin executes the login flow.
func runLogin(flags *loginFlags) error {
	// Enable debug logging if requested
	if flags.debug {
		auth.SetDebug(true)
		fmt.Println("Debug logging enabled")
	}

	apiURL := flags.apiURL
	if apiURL == "" {
		apiURL = auth.DefaultAPIURL
	}

	if flags.debug {
		fmt.Printf("[debug] Using API URL: %s\n", apiURL)
	}

	// Check if already authenticated to this API URL
	if err := checkExistingAuth(apiURL); err != nil {
		return err
	}

	// Perform browser-based login using the auth package
	result, err := auth.BrowserLogin(auth.Options{
		APIURL:  apiURL,
		Timeout: auth.DefaultTimeout,
		OnBrowserOpen: func(url string) {
			fmt.Printf("Opening browser to: %s\n", url)
		},
		OnWaiting: func() {
			fmt.Println("Waiting for authentication...")
			fmt.Println("(If the browser doesn't open, visit the URL manually)")
			fmt.Println()
		},
	})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Complete login (verify key, save config, print success)
	return completeLogin(apiURL, result.APIKey, result.OrgID, result.OrgName)
}

// generateContextName creates a context name from the organization name and environment.
func generateContextName(orgName, apiURL string) string {
	// Sanitize org name for use as context name
	name := sanitizeContextName(orgName)

	// Determine environment suffix from API URL
	env := extractEnvironment(apiURL)
	if env != "" && env != "prod" {
		name = name + "-" + env
	}

	return name
}

// sanitizeContextName converts an organization name to a valid context name.
// Converts to lowercase, replaces spaces/special chars with hyphens.
func sanitizeContextName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces and underscores with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove any characters that aren't alphanumeric or hyphens
	re := regexp.MustCompile(`[^a-z0-9-]`)
	name = re.ReplaceAllString(name, "")

	// Remove consecutive hyphens
	re = regexp.MustCompile(`-+`)
	name = re.ReplaceAllString(name, "-")

	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")

	// If empty after sanitization, use default
	if name == "" {
		name = "default"
	}

	return name
}

// extractEnvironment extracts the environment from an API URL.
// Returns "dev", "stg", or "" for production.
func extractEnvironment(apiURL string) string {
	u, err := url.Parse(apiURL)
	if err != nil {
		return ""
	}

	host := u.Host

	// Check for common environment patterns
	if strings.Contains(host, ".dev.") || strings.HasPrefix(host, "dev.") {
		return "dev"
	}
	if strings.Contains(host, ".stg.") || strings.HasPrefix(host, "stg.") {
		return "stg"
	}
	if strings.Contains(host, ".staging.") || strings.HasPrefix(host, "staging.") {
		return "stg"
	}

	// Production has no suffix
	return ""
}

// completeLogin verifies the API key, saves the configuration, and prints success.
func completeLogin(apiURL, apiKey, orgID, orgName string) error {
	// Verify API key works by calling /v1/cli-auth/verify
	// This endpoint works with API keys (unlike /v1/user/me which requires JWT)
	fmt.Print("Verifying credentials...")

	c := client.New(apiKey, apiURL)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	verifyResp, err := client.VerifyCLICredentials(ctx, c)
	if err != nil {
		fmt.Println(" failed")
		return fmt.Errorf("failed to verify API key: %w", err)
	}
	fmt.Println(" done")

	// Use org name from verify response if not provided in callback
	if orgName == "" && verifyResp.OrganizationName != "" {
		orgName = verifyResp.OrganizationName
	}
	// Use org ID from verify response if not provided in callback
	if orgID == "" && verifyResp.OrganizationID != "" {
		orgID = verifyResp.OrganizationID
	}

	// Fallback to "default" if still no org name
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
		APIKey:           apiKey,
	}

	// Check if context already exists and handle naming conflicts
	existingContextName := contextName
	suffix := 1
	for {
		if _, err := cfg.GetContext(contextName); err != nil {
			// Context doesn't exist, we can use this name
			break
		}
		// Context exists, try with suffix
		suffix++
		contextName = fmt.Sprintf("%s-%d", existingContextName, suffix)
	}

	// Save context
	cfg.SetContext(contextName, newCtx)
	cfg.CurrentContext = contextName

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Print success message
	fmt.Println()
	fmt.Printf("Successfully logged in!\n")
	fmt.Printf("  Organization: %s\n", orgName)
	fmt.Printf("  Context:      %s\n", contextName)
	fmt.Printf("  API URL:      %s\n", apiURL)
	fmt.Println()
	fmt.Println("Credentials saved to:", config.ConfigPath())

	return nil
}

// checkExistingAuth checks if user is already authenticated to this API URL.
// If so, prompts for confirmation before proceeding.
func checkExistingAuth(apiURL string) error {
	cfg, err := config.Load()
	if err != nil {
		// No config or load error, proceed with login
		return nil
	}

	// Check if any context uses this API URL
	for name, ctx := range cfg.Contexts {
		if ctx.APIURL == apiURL && ctx.APIKey != "" {
			fmt.Printf("You are already logged in to %s as context '%s'.\n", apiURL, name)
			fmt.Print("Do you want to create a new login? [y/N]: ")

			// Check if non-interactive mode
			if GetNoInput() {
				return fmt.Errorf("already logged in to %s (use --no-input=false to proceed interactively)", apiURL)
			}

			// Read user input
			var response string
			_, _ = fmt.Scanln(&response)
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				return fmt.Errorf("login cancelled")
			}

			fmt.Println()
			break
		}
	}

	return nil
}
