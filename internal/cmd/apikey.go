// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewAPIKeyCmd creates and returns the api-key parent command.
// This command provides management of API keys for programmatic access.
func NewAPIKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-key",
		Short: "Manage API keys for programmatic access",
		Long: `Manage API keys for programmatic access to the StackEye API.

API keys provide a secure way to authenticate programmatic access to StackEye
without using interactive login. They are ideal for CI/CD pipelines, automation
scripts, and integrations with other tools.

Key Concepts:
  - API keys are scoped to an organization
  - Each key has a name for identification
  - Keys can be revoked at any time without affecting other keys
  - Keys are shown only once at creation - store them securely
  - Key format: se_<32_random_characters>

Security Best Practices:
  - Use separate keys for different applications or environments
  - Rotate keys periodically
  - Revoke keys immediately if compromised
  - Never commit API keys to version control
  - Use environment variables or secret management tools

Available Commands:
  list        List all API keys for the current organization
  create      Create a new API key
  revoke      Revoke an existing API key

Examples:
  # List all API keys
  stackeye api-key list

  # Create a new API key
  stackeye api-key create --name "CI Pipeline"

  # Create a key with JSON output for scripting
  stackeye api-key create --name "Deploy Script" -o json

  # Revoke an API key by ID
  stackeye api-key revoke <key-id>

  # Revoke without confirmation prompt
  stackeye api-key revoke <key-id> --yes

For more information about a specific command:
  stackeye api-key [command] --help`,
		Aliases: []string{"apikey", "apikeys", "api-keys"},
	}

	// Register subcommands
	cmd.AddCommand(NewAPIKeyListCmd())
	cmd.AddCommand(NewAPIKeyCreateCmd())
	cmd.AddCommand(NewAPIKeyDeleteCmd())

	return cmd
}
