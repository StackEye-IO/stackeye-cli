// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/auth"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/spf13/cobra"
)

const (
	// loginTimeout is the maximum time to wait for the callback.
	loginTimeout = 5 * time.Minute

	// defaultAPIURL is the production API endpoint.
	defaultAPIURL = "https://api.stackeye.io"

	// callbackPath is the path for the OAuth callback handler.
	callbackPath = "/callback"
)

// loginFlags holds the command flags for the login command.
type loginFlags struct {
	apiURL string
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

	cmd.Flags().StringVar(&flags.apiURL, "api-url", defaultAPIURL, "StackEye API URL")

	return cmd
}

// runLogin executes the login flow.
func runLogin(flags *loginFlags) error {
	apiURL := flags.apiURL
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	// Check if already authenticated to this API URL
	if err := checkExistingAuth(apiURL); err != nil {
		return err
	}

	// Start local HTTP server for callback
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start local server: %w", err)
	}
	defer listener.Close()

	// Get the assigned port
	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d%s", port, callbackPath)

	// Build the web UI URL
	webUIURL, err := buildWebUIURL(apiURL, callbackURL)
	if err != nil {
		return fmt.Errorf("failed to build web UI URL: %w", err)
	}

	// Create a channel to receive the API key
	resultCh := make(chan loginResult, 1)

	// Create HTTP server with callback handler
	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, makeCallbackHandler(resultCh))

	server := &http.Server{
		Handler: mux,
	}

	// Start server in goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Serve(listener); err != http.ErrServerClosed {
			// Log error but don't fail - the server might be shut down intentionally
		}
	}()

	// Open browser
	fmt.Printf("Opening browser to: %s\n", webUIURL)
	fmt.Println("Waiting for authentication...")
	fmt.Println("(If the browser doesn't open, visit the URL manually)")
	fmt.Println()

	if err := openBrowser(webUIURL); err != nil {
		fmt.Printf("Warning: could not open browser: %v\n", err)
		fmt.Printf("Please visit: %s\n", webUIURL)
	}

	// Wait for callback with timeout
	ctx, cancel := context.WithTimeout(context.Background(), loginTimeout)
	defer cancel()

	var result loginResult
	select {
	case result = <-resultCh:
		// Received callback
	case <-ctx.Done():
		// Shutdown server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
		wg.Wait()
		return fmt.Errorf("login timed out after %v - please try again", loginTimeout)
	}

	// Shutdown server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
	wg.Wait()

	// Check for callback errors
	if result.err != nil {
		return fmt.Errorf("login failed: %w", result.err)
	}

	// Validate API key format
	if !auth.ValidateAPIKey(result.apiKey) {
		return fmt.Errorf("received invalid API key format")
	}

	// Complete login (verify key, save config, print success)
	return completeLogin(apiURL, result.apiKey, result.orgID, result.orgName)
}

// loginResult holds the result of the callback.
type loginResult struct {
	apiKey  string
	orgID   string
	orgName string
	err     error
}

// makeCallbackHandler creates an HTTP handler for the OAuth callback.
func makeCallbackHandler(resultCh chan<- loginResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Security: only accept requests from localhost
		remoteIP := extractIP(r.RemoteAddr)
		if !isLocalhost(remoteIP) {
			http.Error(w, "Forbidden: requests must come from localhost", http.StatusForbidden)
			resultCh <- loginResult{err: fmt.Errorf("request from non-localhost IP: %s", remoteIP)}
			return
		}

		// Extract API key from query parameters
		apiKey := r.URL.Query().Get("api_key")
		if apiKey == "" {
			http.Error(w, "Missing api_key parameter", http.StatusBadRequest)
			resultCh <- loginResult{err: fmt.Errorf("missing api_key parameter")}
			return
		}

		// Extract optional org info
		orgID := r.URL.Query().Get("org_id")
		orgName := r.URL.Query().Get("org_name")

		// Send success response to browser
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>StackEye CLI - Login Successful</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
               display: flex; justify-content: center; align-items: center; height: 100vh;
               margin: 0; background: #f5f5f5; }
        .container { text-align: center; padding: 2rem; background: white;
                     border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #10b981; margin-bottom: 0.5rem; }
        p { color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Login Successful</h1>
        <p>You can close this window and return to your terminal.</p>
    </div>
</body>
</html>`)

		// Send result to channel
		resultCh <- loginResult{
			apiKey:  apiKey,
			orgID:   orgID,
			orgName: orgName,
		}
	}
}

// buildWebUIURL constructs the web UI authentication URL.
func buildWebUIURL(apiURL, callbackURL string) (string, error) {
	webURL, err := apiURLToWebURL(apiURL)
	if err != nil {
		return "", err
	}

	// Build full URL with callback parameter
	u, err := url.Parse(webURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse web URL: %w", err)
	}

	u.Path = "/cli-auth"
	q := u.Query()
	q.Set("callback", callbackURL)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// apiURLToWebURL converts an API URL to the corresponding web UI URL.
// It transforms api.stackeye.io to app.stackeye.io (or api.X.stackeye.io to app.X.stackeye.io).
func apiURLToWebURL(apiURL string) (string, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse API URL: %w", err)
	}

	// Transform the host: api.X -> app.X
	host := u.Host

	// Handle standard patterns:
	// api.stackeye.io -> app.stackeye.io
	// api.dev.stackeye.io -> app.dev.stackeye.io
	// api.stg.stackeye.io -> app.stg.stackeye.io
	if strings.HasPrefix(host, "api.") {
		host = "app." + strings.TrimPrefix(host, "api.")
	} else {
		// For non-standard URLs, try to construct a reasonable web URL
		// This handles cases like localhost or custom domains
		return apiURL, nil
	}

	webURL := &url.URL{
		Scheme: u.Scheme,
		Host:   host,
	}

	return webURL.String(), nil
}

// openBrowser opens the specified URL in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
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
	// Verify API key works by calling /v1/user/me
	fmt.Print("Verifying credentials...")

	c := client.New(apiKey, apiURL)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userResp, err := client.GetCurrentUser(ctx, c)
	if err != nil {
		fmt.Println(" failed")
		return fmt.Errorf("failed to verify API key: %w", err)
	}
	fmt.Println(" done")

	// Use org name from callback or user's email domain as fallback
	if orgName == "" {
		orgName = extractOrgFromEmail(userResp.User.Email)
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
	fmt.Printf("  User:         %s (%s)\n", userResp.User.GetDisplayName(), userResp.User.Email)
	if orgName != "" {
		fmt.Printf("  Organization: %s\n", orgName)
	}
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
			fmt.Scanln(&response)
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

// extractOrgFromEmail extracts a default organization name from an email address.
// Uses the domain name without TLD as the org name.
func extractOrgFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "default"
	}

	domain := parts[1]
	domainParts := strings.Split(domain, ".")

	if len(domainParts) >= 2 {
		// Use the main domain part (e.g., "acme" from "acme.com")
		return domainParts[0]
	}

	return domain
}

// extractIP extracts the IP address from a RemoteAddr string.
// RemoteAddr may be "IP:port" or just "IP" for certain configurations.
func extractIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// No port, return as-is
		return remoteAddr
	}
	return host
}

// isLocalhost checks if an IP address is a localhost address.
func isLocalhost(ip string) bool {
	// Check for IPv4 localhost
	if ip == "127.0.0.1" || strings.HasPrefix(ip, "127.") {
		return true
	}
	// Check for IPv6 localhost
	if ip == "::1" || ip == "[::1]" {
		return true
	}
	return false
}
