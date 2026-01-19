// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/auth"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/spf13/cobra"
)

// NewConfigCmd creates and returns the config parent command.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long: `Manage CLI configuration including API keys, contexts, and preferences.

The config command provides subcommands for managing your StackEye CLI
configuration. Configuration is stored in ~/.config/stackeye/config.yaml.

Examples:
  # Set API key for current context
  stackeye config set-key

  # Set API key with verification
  stackeye config set-key --verify

  # Set API key for specific context
  stackeye config set-key --context myorg`,
	}

	// Add subcommands
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetKeyCmd())

	return cmd
}

// newConfigGetCmd creates the get subcommand to display current configuration.
func newConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display current configuration",
		Long: `Display the current CLI configuration including context, API URL, and masked API key.

The API key is masked for security, showing only the last 4 characters.
Use this command to verify your configuration is set up correctly.

Examples:
  # Show current configuration
  stackeye config get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigGet()
		},
	}

	return cmd
}

// runConfigGet executes the config get command.
func runConfigGet() error {
	cfg := GetConfig()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Get current context name
	contextName := cfg.CurrentContext
	if contextName == "" {
		contextName = "(not set)"
	}

	// Get context details if available
	var apiURL, maskedKey, orgName, orgID string

	ctx, err := cfg.GetCurrentContext()
	if err == nil && ctx != nil {
		apiURL = ctx.EffectiveAPIURL()
		maskedKey = maskAPIKey(ctx.APIKey)
		orgName = ctx.OrganizationName
		orgID = ctx.OrganizationID
	} else {
		apiURL = config.DefaultAPIURL
		maskedKey = "(not set)"
	}

	// Print configuration
	fmt.Printf("Current Context:    %s\n", contextName)
	fmt.Printf("API URL:            %s\n", apiURL)
	fmt.Printf("API Key:            %s\n", maskedKey)

	if orgName != "" {
		fmt.Printf("Organization:       %s\n", orgName)
	}
	if orgID != "" {
		fmt.Printf("Organization ID:    %s\n", orgID)
	}

	// Show config file path
	fmt.Printf("Config File:        %s\n", config.ConfigPath())

	return nil
}

// maskAPIKey masks an API key, showing only the last 4 characters.
// Returns "(not set)" if the key is empty.
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "(not set)"
	}

	// API keys have format: se_<64 hex chars> (67 chars total)
	// Show prefix and last 4 chars: se_****...xxxx
	if len(apiKey) >= 7 {
		return apiKey[:3] + "****..." + apiKey[len(apiKey)-4:]
	}

	// For malformed keys, just mask most of it
	if len(apiKey) > 4 {
		return "****" + apiKey[len(apiKey)-4:]
	}

	return "****"
}

// configSetKeyFlags holds the command flags for set-key.
type configSetKeyFlags struct {
	verify      bool
	contextName string
}

// newConfigSetKeyCmd creates the set-key subcommand.
func newConfigSetKeyCmd() *cobra.Command {
	flags := &configSetKeyFlags{}

	cmd := &cobra.Command{
		Use:   "set-key [api-key]",
		Short: "Set API key for authentication",
		Long: `Set the API key for authenticating with the StackEye API.

The API key can be provided as an argument or entered interactively
(recommended for security). When entered interactively, input is hidden.

API keys have the format: se_<64 hex characters>

Use --verify to validate the API key by making a test API call before
saving. Use --context to set the key for a specific context instead of
the current one.

Examples:
  # Set API key interactively (recommended)
  stackeye config set-key

  # Set API key with verification
  stackeye config set-key --verify

  # Set API key for specific context
  stackeye config set-key --context production

  # Set API key directly (less secure - visible in history)
  stackeye config set-key se_abc123...`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigSetKey(flags, args)
		},
	}

	cmd.Flags().BoolVar(&flags.verify, "verify", false, "verify API key by calling the API before saving")
	cmd.Flags().StringVar(&flags.contextName, "context", "", "context to set the key for (default: current context)")

	return cmd
}

// runConfigSetKey executes the set-key command.
func runConfigSetKey(flags *configSetKeyFlags, args []string) error {
	// Get the API key from args or prompt
	apiKey, err := getAPIKey(args)
	if err != nil {
		return err
	}

	// Validate format
	if !auth.ValidateAPIKey(apiKey) {
		return fmt.Errorf("invalid API key format: must start with 'se_' followed by 64 hex characters")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine target context
	contextName := flags.contextName
	if contextName == "" {
		contextName = cfg.CurrentContext
	}

	// If no context exists, create default
	if contextName == "" {
		contextName = "default"
		cfg.CurrentContext = contextName
	}

	// Ensure context exists
	ctx, err := cfg.GetContext(contextName)
	if err != nil {
		// Context doesn't exist, create it
		ctx = &config.Context{}
		cfg.SetContext(contextName, ctx)
	}

	// Verify API key if requested
	if flags.verify {
		if err := verifyAPIKey(apiKey, ctx.APIURL); err != nil {
			return err
		}
	}

	// Set the API key
	ctx.APIKey = apiKey
	cfg.SetContext(contextName, ctx)

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("API key set for context '%s'\n", contextName)
	fmt.Printf("Configuration saved to: %s\n", config.ConfigPath())

	return nil
}

// getAPIKey retrieves the API key from args or prompts interactively.
func getAPIKey(args []string) (string, error) {
	// If provided as argument, use it
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}

	// Check if interactive mode is disabled
	if GetNoInput() {
		return "", fmt.Errorf("API key required: provide as argument or use interactive mode (--no-input=false)")
	}

	// Prompt for API key securely
	apiKey, err := interactive.AskPassword(&interactive.PasswordPromptOptions{
		Message: "Enter API key",
		Help:    "Your StackEye API key (format: se_<64 hex characters>)",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("API key cannot be empty")
			}
			if !auth.ValidateAPIKey(s) {
				return fmt.Errorf("invalid format: must start with 'se_' followed by 64 hex characters")
			}
			return nil
		},
	})
	if err != nil {
		if err == interactive.ErrPromptCancelled {
			return "", fmt.Errorf("cancelled")
		}
		return "", fmt.Errorf("failed to read API key: %w", err)
	}

	return apiKey, nil
}

// verifyAPIKey validates the API key by making a test API call.
func verifyAPIKey(apiKey, apiURL string) error {
	fmt.Print("Verifying API key... ")

	// Use default API URL if not set
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	// Create client and call user endpoint
	c := client.New(apiKey, apiURL)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userResp, err := client.GetCurrentUser(ctx, c)
	if err != nil {
		fmt.Println("failed")
		return fmt.Errorf("API key verification failed: %w", err)
	}

	fmt.Println("valid")
	fmt.Printf("  Authenticated as: %s (%s)\n", userResp.User.GetDisplayName(), userResp.User.Email)

	return nil
}
