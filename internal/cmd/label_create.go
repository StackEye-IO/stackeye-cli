// Package cmd implements the CLI commands for StackEye.
// Task #8066
package cmd

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// labelCreateTimeout is the maximum time to wait for the API response.
const labelCreateTimeout = 30 * time.Second

// labelKeyMaxLength is the maximum length for a label key (Kubernetes constraint).
const labelKeyMaxLength = 63

// labelKeyPattern matches valid Kubernetes-style label keys.
// Must start and end with alphanumeric, can contain hyphens in between.
// Pattern: ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
var labelKeyPattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// NewLabelCreateCmd creates and returns the label create subcommand.
func NewLabelCreateCmd() *cobra.Command {
	var displayName string
	var description string
	var color string

	cmd := &cobra.Command{
		Use:   "create <key>",
		Short: "Create a new label key",
		Long: `Create a new probe label key for your organization.

Label keys allow you to organize and filter probes using key-value pairs.
Keys must follow Kubernetes naming conventions:
  - Lowercase alphanumeric characters and hyphens only
  - Must start and end with alphanumeric characters
  - Maximum 63 characters

Requires authentication via 'stackeye login' or API key.

Examples:
  # Create a simple label key
  stackeye label create env

  # Create a label key with display name
  stackeye label create env --display-name "Environment"

  # Create a label key with full metadata
  stackeye label create env --display-name "Environment" \
    --description "Deployment environment classification" \
    --color "#10B981"

  # Create label key with custom color
  stackeye label create tier --display-name "Service Tier" --color "#3B82F6"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			return runLabelCreate(cmd.Context(), key, displayName, description, color)
		},
	}

	// Define flags
	cmd.Flags().StringVar(&displayName, "display-name", "", "Human-readable display name for the key")
	cmd.Flags().StringVar(&description, "description", "", "Optional description of the label key")
	cmd.Flags().StringVar(&color, "color", "", "Hex color code for UI badge (e.g., #10B981)")

	return cmd
}

// runLabelCreate executes the label create command logic.
func runLabelCreate(ctx context.Context, key, displayName, description, color string) error {
	// Validate key format locally before API call
	if err := validateLabelKey(key); err != nil {
		return err
	}

	// Validate color format if provided
	if color != "" {
		if err := validateHexColor(color); err != nil {
			return err
		}
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		details := []string{
			"Key", key,
		}
		if displayName != "" {
			details = append(details, "Display Name", displayName)
		}
		if color != "" {
			details = append(details, "Color", color)
		}
		dryrun.PrintAction("create", "label key", details...)
		return nil
	}

	// Get API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build request
	req := client.CreateLabelKeyRequest{
		Key: key,
	}
	if displayName != "" {
		req.DisplayName = &displayName
	}
	if description != "" {
		req.Description = &description
	}
	if color != "" {
		req.Color = &color
	}

	// Call SDK to create label key with timeout
	reqCtx, cancel := context.WithTimeout(ctx, labelCreateTimeout)
	defer cancel()

	result, err := client.CreateLabelKey(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create label key: %w", err)
	}

	// Print the created label key using the configured output format
	return output.PrintLabelKey(result.LabelKey)
}

// validateLabelKey validates the key format follows Kubernetes naming conventions.
func validateLabelKey(key string) error {
	if key == "" {
		return fmt.Errorf("label key is required")
	}

	if len(key) > labelKeyMaxLength {
		return fmt.Errorf("label key must be at most %d characters (got %d)", labelKeyMaxLength, len(key))
	}

	if !labelKeyPattern.MatchString(key) {
		return fmt.Errorf("invalid key format: %q (must be lowercase alphanumeric with hyphens, starting and ending with alphanumeric)", key)
	}

	return nil
}

// validateHexColor validates that the color is a valid 7-character hex code.
func validateHexColor(color string) error {
	if len(color) != 7 {
		return fmt.Errorf("color must be a 7-character hex code (e.g., #RRGGBB), got %q", color)
	}

	if color[0] != '#' {
		return fmt.Errorf("color must start with '#' (e.g., #RRGGBB), got %q", color)
	}

	for i := 1; i < 7; i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("color must be a valid hex code (e.g., #RRGGBB), got %q", color)
		}
	}

	return nil
}
