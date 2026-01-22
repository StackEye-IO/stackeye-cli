// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// statusPageUpdateTimeout is the maximum time to wait for the API response.
const statusPageUpdateTimeout = 30 * time.Second

// statusPageUpdateFlags holds the flag values for the status-page update command.
// All fields are pointers to support partial updates (nil = not specified).
type statusPageUpdateFlags struct {
	// Basic fields
	name                 *string
	slug                 *string
	customDomain         *string
	logoURL              *string
	faviconURL           *string
	headerText           *string
	footerText           *string
	theme                *string
	isPublic             *bool
	showUptimePercentage *bool
	enabled              *bool

	// YAML file input
	fromFile string
}

// statusPageUpdateYAMLConfig represents the YAML structure for --from-file input on update.
type statusPageUpdateYAMLConfig struct {
	Name                 *string `yaml:"name,omitempty"`
	Slug                 *string `yaml:"slug,omitempty"`
	CustomDomain         *string `yaml:"custom_domain,omitempty"`
	LogoURL              *string `yaml:"logo_url,omitempty"`
	FaviconURL           *string `yaml:"favicon_url,omitempty"`
	HeaderText           *string `yaml:"header_text,omitempty"`
	FooterText           *string `yaml:"footer_text,omitempty"`
	Theme                *string `yaml:"theme,omitempty"`
	IsPublic             *bool   `yaml:"is_public,omitempty"`
	ShowUptimePercentage *bool   `yaml:"show_uptime_percentage,omitempty"`
	Enabled              *bool   `yaml:"enabled,omitempty"`
}

// NewStatusPageUpdateCmd creates and returns the status-page update subcommand.
func NewStatusPageUpdateCmd() *cobra.Command {
	flags := &statusPageUpdateFlags{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing status page",
		Long: `Update an existing status page configuration.

Only the specified flags will be updated; all other fields remain unchanged.
This allows for partial updates without needing to specify the entire configuration.

Examples:
  # Update status page name
  stackeye status-page update 123 --name "New Status Page Name"

  # Update theme to dark mode
  stackeye status-page update 123 --theme dark

  # Disable a status page
  stackeye status-page update 123 --enabled=false

  # Re-enable a status page
  stackeye status-page update 123 --enabled=true

  # Make a status page private
  stackeye status-page update 123 --public=false

  # Update custom domain and branding
  stackeye status-page update 123 \
    --custom-domain status.newdomain.com \
    --logo-url {logo-url} \
    --favicon-url {favicon-url}

  # Update header and footer text
  stackeye status-page update 123 \
    --header-text "Scheduled maintenance in progress" \
    --footer-text "Contact {support-email} for help"

  # Hide uptime percentages
  stackeye status-page update 123 --show-uptime=false

  # Update from YAML file
  stackeye status-page update 123 --from-file status-page-updates.yaml

YAML File Format:
  name: "Updated Status Page Name"
  slug: "new-slug"
  theme: "dark"
  is_public: false
  show_uptime_percentage: true
  enabled: true
  header_text: "Service status"
  footer_text: "Contact support for help"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusPageUpdate(cmd, args[0], flags)
		},
	}

	// All flags are optional for partial updates
	cmd.Flags().StringVar(statusPageStringPtrVar(&flags.name), "name", "", "status page name")
	cmd.Flags().StringVar(statusPageStringPtrVar(&flags.slug), "slug", "", "URL-safe slug")
	cmd.Flags().StringVar(statusPageStringPtrVar(&flags.customDomain), "custom-domain", "", "custom domain (e.g., status.example.com)")
	cmd.Flags().StringVar(statusPageStringPtrVar(&flags.logoURL), "logo-url", "", "URL to logo image")
	cmd.Flags().StringVar(statusPageStringPtrVar(&flags.faviconURL), "favicon-url", "", "URL to favicon image")
	cmd.Flags().StringVar(statusPageStringPtrVar(&flags.headerText), "header-text", "", "custom header/announcement text")
	cmd.Flags().StringVar(statusPageStringPtrVar(&flags.footerText), "footer-text", "", "custom footer text")
	cmd.Flags().StringVar(statusPageStringPtrVar(&flags.theme), "theme", "", "page theme: light, dark, system")
	cmd.Flags().BoolVar(statusPageBoolPtrVar(&flags.isPublic), "public", false, "make page publicly accessible")
	cmd.Flags().BoolVar(statusPageBoolPtrVar(&flags.showUptimePercentage), "show-uptime", false, "show uptime percentages")
	cmd.Flags().BoolVar(statusPageBoolPtrVar(&flags.enabled), "enabled", false, "enable the status page")

	// File input
	cmd.Flags().StringVar(&flags.fromFile, "from-file", "", "update status page from YAML file")

	return cmd
}

// statusPageStringPtrVar is a helper to bind a **string to a cobra StringVar.
// This allows detecting whether a flag was explicitly set.
func statusPageStringPtrVar(p **string) *string {
	*p = new(string)
	return *p
}

// statusPageBoolPtrVar is a helper to bind a **bool to a cobra BoolVar.
// This allows detecting whether a flag was explicitly set.
func statusPageBoolPtrVar(p **bool) *bool {
	*p = new(bool)
	return *p
}

// runStatusPageUpdate executes the status-page update command logic.
func runStatusPageUpdate(cmd *cobra.Command, idArg string, flags *statusPageUpdateFlags) error {
	// Parse and validate status page ID (uint)
	id, err := strconv.ParseUint(idArg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid status page ID %q: must be a positive integer", idArg)
	}
	statusPageID := uint(id)

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Handle --from-file if provided
	if flags.fromFile != "" {
		return runStatusPageUpdateFromFile(cmd.Context(), apiClient, statusPageID, flags.fromFile)
	}

	// Build request from flags
	req, err := buildStatusPageUpdateRequest(cmd, flags)
	if err != nil {
		return err
	}

	// Call SDK to update status page with timeout
	ctx, cancel := context.WithTimeout(cmd.Context(), statusPageUpdateTimeout)
	defer cancel()

	statusPage, err := client.UpdateStatusPage(ctx, apiClient, statusPageID, req)
	if err != nil {
		return fmt.Errorf("failed to update status page: %w", err)
	}

	// Print the updated status page using the configured output format
	return output.Print(statusPage)
}

// buildStatusPageUpdateRequest constructs the API request from command flags.
func buildStatusPageUpdateRequest(cmd *cobra.Command, flags *statusPageUpdateFlags) (*client.UpdateStatusPageRequest, error) {
	req := &client.UpdateStatusPageRequest{}
	hasUpdates := false

	// Check name flag
	if cmd.Flags().Changed("name") {
		if *flags.name == "" {
			return nil, fmt.Errorf("--name cannot be empty")
		}
		req.Name = flags.name
		hasUpdates = true
	}

	// Check slug flag
	if cmd.Flags().Changed("slug") {
		if *flags.slug == "" {
			return nil, fmt.Errorf("--slug cannot be empty")
		}
		if err := validateSlug(*flags.slug); err != nil {
			return nil, err
		}
		req.Slug = flags.slug
		hasUpdates = true
	}

	// Check theme flag
	if cmd.Flags().Changed("theme") {
		if err := validateTheme(*flags.theme); err != nil {
			return nil, err
		}
		themeLower := strings.ToLower(*flags.theme)
		req.Theme = &themeLower
		hasUpdates = true
	}

	// Check custom-domain flag
	if cmd.Flags().Changed("custom-domain") {
		req.CustomDomain = flags.customDomain
		hasUpdates = true
	}

	// Check logo-url flag
	if cmd.Flags().Changed("logo-url") {
		req.LogoURL = flags.logoURL
		hasUpdates = true
	}

	// Check favicon-url flag
	if cmd.Flags().Changed("favicon-url") {
		req.FaviconURL = flags.faviconURL
		hasUpdates = true
	}

	// Check header-text flag
	if cmd.Flags().Changed("header-text") {
		req.HeaderText = flags.headerText
		hasUpdates = true
	}

	// Check footer-text flag
	if cmd.Flags().Changed("footer-text") {
		req.FooterText = flags.footerText
		hasUpdates = true
	}

	// Check public flag
	if cmd.Flags().Changed("public") {
		req.IsPublic = flags.isPublic
		hasUpdates = true
	}

	// Check show-uptime flag
	if cmd.Flags().Changed("show-uptime") {
		req.ShowUptimePercentage = flags.showUptimePercentage
		hasUpdates = true
	}

	// Check enabled flag
	if cmd.Flags().Changed("enabled") {
		req.Enabled = flags.enabled
		hasUpdates = true
	}

	// Require at least one update flag
	if !hasUpdates {
		return nil, fmt.Errorf("no update flags specified; use --help to see available options")
	}

	return req, nil
}

// runStatusPageUpdateFromFile executes status page update from a YAML file.
func runStatusPageUpdateFromFile(ctx context.Context, apiClient *client.Client, statusPageID uint, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	var cfg statusPageUpdateYAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	req := &client.UpdateStatusPageRequest{}
	hasUpdates := false

	// Apply name if provided
	if cfg.Name != nil {
		if *cfg.Name == "" {
			return fmt.Errorf("YAML name cannot be empty when specified")
		}
		req.Name = cfg.Name
		hasUpdates = true
	}

	// Apply slug if provided
	if cfg.Slug != nil {
		if *cfg.Slug == "" {
			return fmt.Errorf("YAML slug cannot be empty when specified")
		}
		if err := validateSlug(*cfg.Slug); err != nil {
			return err
		}
		req.Slug = cfg.Slug
		hasUpdates = true
	}

	// Apply theme if provided
	if cfg.Theme != nil {
		if err := validateTheme(*cfg.Theme); err != nil {
			return err
		}
		themeLower := strings.ToLower(*cfg.Theme)
		req.Theme = &themeLower
		hasUpdates = true
	}

	// Apply custom_domain if provided
	if cfg.CustomDomain != nil {
		req.CustomDomain = cfg.CustomDomain
		hasUpdates = true
	}

	// Apply logo_url if provided
	if cfg.LogoURL != nil {
		req.LogoURL = cfg.LogoURL
		hasUpdates = true
	}

	// Apply favicon_url if provided
	if cfg.FaviconURL != nil {
		req.FaviconURL = cfg.FaviconURL
		hasUpdates = true
	}

	// Apply header_text if provided
	if cfg.HeaderText != nil {
		req.HeaderText = cfg.HeaderText
		hasUpdates = true
	}

	// Apply footer_text if provided
	if cfg.FooterText != nil {
		req.FooterText = cfg.FooterText
		hasUpdates = true
	}

	// Apply is_public if provided
	if cfg.IsPublic != nil {
		req.IsPublic = cfg.IsPublic
		hasUpdates = true
	}

	// Apply show_uptime_percentage if provided
	if cfg.ShowUptimePercentage != nil {
		req.ShowUptimePercentage = cfg.ShowUptimePercentage
		hasUpdates = true
	}

	// Apply enabled if provided
	if cfg.Enabled != nil {
		req.Enabled = cfg.Enabled
		hasUpdates = true
	}

	if !hasUpdates {
		return fmt.Errorf("YAML file contains no updates")
	}

	// Call SDK to update status page
	reqCtx, cancel := context.WithTimeout(ctx, statusPageUpdateTimeout)
	defer cancel()

	statusPage, err := client.UpdateStatusPage(reqCtx, apiClient, statusPageID, req)
	if err != nil {
		return fmt.Errorf("failed to update status page: %w", err)
	}

	return output.Print(statusPage)
}
