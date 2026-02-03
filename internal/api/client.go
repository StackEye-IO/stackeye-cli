// Package api provides helpers for initializing and using the StackEye API client.
//
// This package centralizes the pattern of creating an SDK client from the CLI
// configuration, providing consistent error handling across all commands that
// need API access.
package api

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

// verbosityGetter is the interface for getting the verbosity level.
// This allows for testing by providing a mock verbosity getter.
type verbosityGetter func() int

// currentVerbosityGetter returns 0 - subcommands should provide their own.
// This is overridden via SetVerbosityGetter in init or tests.
var currentVerbosityGetter verbosityGetter

// SetVerbosityGetter sets the function used to get the verbosity level.
// This should be called during CLI initialization to wire up the verbosity flag.
//
// Example usage in root.go init():
//
//	api.SetVerbosityGetter(GetVerbosity)
func SetVerbosityGetter(getter func() int) {
	currentVerbosityGetter = getter
}

// getVerbosity returns the current verbosity level, or 0 if not configured.
func getVerbosity() int {
	if currentVerbosityGetter == nil {
		return 0
	}
	return currentVerbosityGetter()
}

// timeoutGetter is the interface for getting the timeout in seconds.
type timeoutGetter func() int

// currentTimeoutGetter returns 0 (use default) - overridden via SetTimeoutGetter.
var currentTimeoutGetter timeoutGetter

// SetTimeoutGetter sets the function used to get the timeout in seconds.
// This should be called during CLI initialization to wire up the timeout flag.
func SetTimeoutGetter(getter func() int) {
	currentTimeoutGetter = getter
}

// getTimeout returns the current timeout in seconds, or 0 if not configured.
func getTimeout() int {
	if currentTimeoutGetter == nil {
		return 0
	}
	return currentTimeoutGetter()
}

// Error types for API client initialization failures.
var (
	// ErrConfigNotLoaded is returned when GetClient is called before config is loaded.
	ErrConfigNotLoaded = errors.New("api: configuration not loaded")

	// ErrNoCurrentContext is returned when no context is set in the configuration.
	ErrNoCurrentContext = errors.New("api: no context configured")

	// ErrContextNotFound is returned when the current context doesn't exist.
	ErrContextNotFound = errors.New("api: context not found")

	// ErrNoAPIKey is returned when the context has no API key configured.
	ErrNoAPIKey = errors.New("api: no API key configured for context")
)

// configGetter is the interface for getting the loaded configuration.
// This allows for testing by providing a mock config getter.
type configGetter func() *config.Config

// defaultConfigGetter returns nil - subcommands should provide their own.
// This is overridden via SetConfigGetter in init or tests.
var currentConfigGetter configGetter

// SetConfigGetter sets the function used to get the loaded configuration.
// This should be called during CLI initialization to wire up the config loader.
// Returns the previous config getter, useful for cleanup in tests.
//
// Example usage in root.go init():
//
//	api.SetConfigGetter(GetConfig)
func SetConfigGetter(getter func() *config.Config) func() *config.Config {
	prev := currentConfigGetter
	currentConfigGetter = getter
	return prev
}

// GetClient creates and returns an SDK client from the current configuration context.
//
// The function:
//  1. Gets the loaded configuration via the configured getter
//  2. Retrieves the current context
//  3. Validates that an API key is present
//  4. Creates and returns an SDK client
//
// Returns an error if:
//   - Configuration is not loaded (ErrConfigNotLoaded)
//   - No current context is set (ErrNoCurrentContext)
//   - Current context doesn't exist (ErrContextNotFound)
//   - No API key is configured (ErrNoAPIKey)
//
// Example:
//
//	client, err := api.GetClient()
//	if err != nil {
//	    return fmt.Errorf("failed to initialize API client: %w", err)
//	}
//	// Use client for API calls
func GetClient() (*client.Client, error) {
	if currentConfigGetter == nil {
		return nil, ErrConfigNotLoaded
	}

	cfg := currentConfigGetter()
	if cfg == nil {
		return nil, ErrConfigNotLoaded
	}

	// Check if there's a current context
	if cfg.CurrentContext == "" {
		return nil, ErrNoCurrentContext
	}

	// Get the current context
	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		if errors.Is(err, config.ErrContextNotFound) {
			return nil, fmt.Errorf("%w: %q", ErrContextNotFound, cfg.CurrentContext)
		}
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	// Check if there's an API key
	if ctx.APIKey == "" {
		return nil, fmt.Errorf("%w: %q", ErrNoAPIKey, cfg.CurrentContext)
	}

	// Create client with optional verbosity logging
	opts := buildClientOptions()
	apiClient := client.New(ctx.APIKey, ctx.EffectiveAPIURL(), opts...)
	return apiClient, nil
}

// buildClientOptions creates SDK client options based on current settings.
func buildClientOptions() []client.Option {
	var opts []client.Option

	// Add verbosity logging if enabled
	verbosity := getVerbosity()
	if verbosity > 0 {
		opts = append(opts, client.WithVerbosity(verbosity, os.Stderr))
	}

	// Apply timeout configuration.
	// Precedence: --timeout flag / STACKEYE_TIMEOUT env > config preference > SDK default (30s).
	if timeout := getTimeout(); timeout > 0 {
		opts = append(opts, client.WithTimeout(time.Duration(timeout)*time.Second))
	} else if cfg := getConfigForTimeout(); cfg != nil && cfg.Preferences != nil && cfg.Preferences.DefaultTimeout > 0 {
		opts = append(opts, client.WithTimeout(time.Duration(cfg.Preferences.DefaultTimeout)*time.Second))
	}

	return opts
}

// getConfigForTimeout returns the current config for reading timeout preferences.
func getConfigForTimeout() *config.Config {
	if currentConfigGetter == nil {
		return nil
	}
	return currentConfigGetter()
}

// GetClientOptions returns SDK client options for use by commands that create
// clients directly (e.g., login, config set-api-key). This allows them to
// use verbosity logging without going through GetClient().
//
// Example:
//
//	opts := api.GetClientOptions()
//	c := client.New(apiKey, apiURL, opts...)
func GetClientOptions() []client.Option {
	return buildClientOptions()
}

// RequireClient creates and returns an SDK client, exiting on error.
//
// This is a convenience function for commands that require authentication.
// If the client cannot be created, it prints an error message to stderr
// and exits with code 1.
//
// Example:
//
//	func runMyCommand() error {
//	    client := api.RequireClient()
//	    // Use client for API calls
//	}
func RequireClient() *client.Client {
	apiClient, err := GetClient()
	if err != nil {
		printAuthError(err)
		os.Exit(1)
	}
	return apiClient
}

// printAuthError prints a user-friendly error message for authentication failures.
func printAuthError(err error) {
	fmt.Fprintln(os.Stderr, "Error: Not authenticated")
	fmt.Fprintln(os.Stderr)

	switch {
	case errors.Is(err, ErrConfigNotLoaded):
		fmt.Fprintln(os.Stderr, "Configuration is not loaded.")
	case errors.Is(err, ErrNoCurrentContext):
		fmt.Fprintln(os.Stderr, "No context configured. Run 'stackeye login' to authenticate.")
	case errors.Is(err, ErrContextNotFound):
		fmt.Fprintln(os.Stderr, "Context not found. Run 'stackeye context list' to see available contexts.")
	case errors.Is(err, ErrNoAPIKey):
		fmt.Fprintln(os.Stderr, "No API key configured. Run 'stackeye login' to authenticate.")
	default:
		fmt.Fprintf(os.Stderr, "Failed to initialize API client: %v\n", err)
	}
}

// GetClientWithContext creates an SDK client for a specific named context.
// This is useful for commands that need to operate on a different context
// than the current one.
//
// Returns the same errors as GetClient, but for the specified context name.
func GetClientWithContext(contextName string) (*client.Client, error) {
	if currentConfigGetter == nil {
		return nil, ErrConfigNotLoaded
	}

	cfg := currentConfigGetter()
	if cfg == nil {
		return nil, ErrConfigNotLoaded
	}

	// Get the specified context
	ctx, err := cfg.GetContext(contextName)
	if err != nil {
		if errors.Is(err, config.ErrContextNotFound) {
			return nil, fmt.Errorf("%w: %q", ErrContextNotFound, contextName)
		}
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	// Check if there's an API key
	if ctx.APIKey == "" {
		return nil, fmt.Errorf("%w: %q", ErrNoAPIKey, contextName)
	}

	// Create client with optional verbosity logging
	opts := buildClientOptions()
	apiClient := client.New(ctx.APIKey, ctx.EffectiveAPIURL(), opts...)
	return apiClient, nil
}
