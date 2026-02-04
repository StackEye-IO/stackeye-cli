// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// statusPageCreateTimeout is the maximum time to wait for the API response.
const statusPageCreateTimeout = 30 * time.Second

// validThemes lists the valid theme preset values.
var validThemes = []string{"light", "dark", "system"}

// statusPageCreateFlags holds the flag values for the status-page create command.
type statusPageCreateFlags struct {
	// Required
	name string

	// Optional
	slug                 string
	customDomain         string
	logoURL              string
	faviconURL           string
	headerText           string
	footerText           string
	theme                string
	isPublic             bool
	showUptimePercentage bool
	enabled              bool
	fromFile             string
}

// statusPageYAMLConfig represents the YAML structure for --from-file input.
type statusPageYAMLConfig struct {
	Name                 string  `yaml:"name"`
	Slug                 string  `yaml:"slug,omitempty"`
	CustomDomain         *string `yaml:"custom_domain,omitempty"`
	LogoURL              *string `yaml:"logo_url,omitempty"`
	FaviconURL           *string `yaml:"favicon_url,omitempty"`
	HeaderText           *string `yaml:"header_text,omitempty"`
	FooterText           *string `yaml:"footer_text,omitempty"`
	Theme                string  `yaml:"theme,omitempty"`
	IsPublic             *bool   `yaml:"is_public,omitempty"`
	ShowUptimePercentage *bool   `yaml:"show_uptime_percentage,omitempty"`
	Enabled              *bool   `yaml:"enabled,omitempty"`
}

// NewStatusPageCreateCmd creates and returns the status-page create subcommand.
func NewStatusPageCreateCmd() *cobra.Command {
	flags := &statusPageCreateFlags{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new status page",
		Long: `Create a new status page for public service status display.

Status pages provide customers and stakeholders with real-time visibility into
your service health. Each page can be customized with branding and configured
to show specific probes.

Required Flags:
  --name       Human-readable name for the status page

Optional Flags:
  --slug                 URL-safe identifier (auto-generated from name if not provided)
  --custom-domain        Custom domain (e.g., status.example.com)
  --logo-url             URL to logo image
  --favicon-url          URL to favicon image
  --header-text          Custom header/announcement text
  --footer-text          Custom footer text
  --theme                Page theme: light, dark, system (default: light)
  --public               Make page publicly accessible (default: true)
  --show-uptime          Show uptime percentages (default: true)
  --enabled              Enable the status page (default: true)
  --from-file            Create status page from YAML file

Plan Limits:
  Free:       1 status page
  Starter:    2 status pages
  Pro:        5 status pages
  Team:       Unlimited
  Enterprise: Unlimited

Examples:
  # Create a basic status page
  stackeye status-page create --name "Acme Status"

  # Create with custom slug and theme
  stackeye status-page create --name "Acme Status" --slug acme-status --theme dark

  # Create with custom domain and branding
  stackeye status-page create --name "Acme Status" \
    --custom-domain status.acme.com \
    --logo-url {logo-url} \
    --favicon-url {favicon-url}

  # Create with custom header and footer
  stackeye status-page create --name "Acme Status" \
    --header-text "Welcome to our status page" \
    --footer-text "Contact {support-email} for help"

  # Create a private status page (not publicly accessible)
  stackeye status-page create --name "Internal Status" --public=false

  # Create from YAML file
  stackeye status-page create --from-file status-page.yaml

YAML File Format:
  name: "Acme Status"
  slug: "acme-status"
  theme: "dark"
  is_public: true
  show_uptime_percentage: true
  enabled: true
  header_text: "Welcome to our status page"
  footer_text: "Contact {support-email} for help"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusPageCreate(cmd.Context(), flags)
		},
	}

	// Required flags (unless using --from-file)
	cmd.Flags().StringVar(&flags.name, "name", "", "status page name (required)")

	// Optional flags
	cmd.Flags().StringVar(&flags.slug, "slug", "", "URL-safe slug (auto-generated from name if not provided)")
	cmd.Flags().StringVar(&flags.customDomain, "custom-domain", "", "custom domain (e.g., status.example.com)")
	cmd.Flags().StringVar(&flags.logoURL, "logo-url", "", "URL to logo image")
	cmd.Flags().StringVar(&flags.faviconURL, "favicon-url", "", "URL to favicon image")
	cmd.Flags().StringVar(&flags.headerText, "header-text", "", "custom header/announcement text")
	cmd.Flags().StringVar(&flags.footerText, "footer-text", "", "custom footer text")
	cmd.Flags().StringVar(&flags.theme, "theme", "light", "page theme: light, dark, system")
	cmd.Flags().BoolVar(&flags.isPublic, "public", true, "make page publicly accessible")
	cmd.Flags().BoolVar(&flags.showUptimePercentage, "show-uptime", true, "show uptime percentages")
	cmd.Flags().BoolVar(&flags.enabled, "enabled", true, "enable the status page")
	cmd.Flags().StringVar(&flags.fromFile, "from-file", "", "create status page from YAML file")

	return cmd
}

// runStatusPageCreate executes the status-page create command logic.
func runStatusPageCreate(ctx context.Context, flags *statusPageCreateFlags) error {
	var req *client.CreateStatusPageRequest
	var err error

	// Handle --from-file if provided
	if flags.fromFile != "" {
		req, err = buildStatusPageRequestFromYAML(flags.fromFile)
		if err != nil {
			return err
		}
	} else {
		// Build request from flags
		req, err = buildStatusPageRequestFromFlags(flags)
		if err != nil {
			return err
		}
	}

	// Dry-run check: after validation/request building, before API calls
	if GetDryRun() {
		dryrun.PrintAction("create", "status page",
			"Name", req.Name,
			"Slug", req.Slug,
			"Theme", req.Theme,
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to create status page with timeout
	reqCtx, cancel := context.WithTimeout(ctx, statusPageCreateTimeout)
	defer cancel()

	statusPage, err := client.CreateStatusPage(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create status page: %w", err)
	}

	// Print the created status page using the configured output format
	return output.Print(statusPage)
}

// buildStatusPageRequestFromFlags constructs the API request from command flags.
func buildStatusPageRequestFromFlags(flags *statusPageCreateFlags) (*client.CreateStatusPageRequest, error) {
	// Validate required fields
	if flags.name == "" {
		return nil, fmt.Errorf("--name is required")
	}

	// Validate theme if provided
	if flags.theme != "" {
		if err := validateTheme(flags.theme); err != nil {
			return nil, err
		}
	}

	// Validate slug format if provided
	if flags.slug != "" {
		if err := validateSlug(flags.slug); err != nil {
			return nil, err
		}
	}

	// Normalize theme to lowercase for API consistency
	theme := strings.ToLower(flags.theme)

	req := &client.CreateStatusPageRequest{
		Name:                 flags.name,
		Slug:                 flags.slug,
		Theme:                theme,
		IsPublic:             &flags.isPublic,
		ShowUptimePercentage: &flags.showUptimePercentage,
		Enabled:              &flags.enabled,
	}

	// Set optional string pointers only if provided
	if flags.customDomain != "" {
		req.CustomDomain = &flags.customDomain
	}
	if flags.logoURL != "" {
		req.LogoURL = &flags.logoURL
	}
	if flags.faviconURL != "" {
		req.FaviconURL = &flags.faviconURL
	}
	if flags.headerText != "" {
		req.HeaderText = &flags.headerText
	}
	if flags.footerText != "" {
		req.FooterText = &flags.footerText
	}

	return req, nil
}

// buildStatusPageRequestFromYAML constructs the API request from a YAML file.
func buildStatusPageRequestFromYAML(filePath string) (*client.CreateStatusPageRequest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	var cfg statusPageYAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if cfg.Name == "" {
		return nil, fmt.Errorf("YAML file must contain 'name' field")
	}

	// Validate theme if provided
	if cfg.Theme != "" {
		if err := validateTheme(cfg.Theme); err != nil {
			return nil, err
		}
	}

	// Validate slug if provided
	if cfg.Slug != "" {
		if err := validateSlug(cfg.Slug); err != nil {
			return nil, err
		}
	}

	// Normalize theme to lowercase for API consistency
	theme := strings.ToLower(cfg.Theme)

	req := &client.CreateStatusPageRequest{
		Name:                 cfg.Name,
		Slug:                 cfg.Slug,
		CustomDomain:         cfg.CustomDomain,
		LogoURL:              cfg.LogoURL,
		FaviconURL:           cfg.FaviconURL,
		HeaderText:           cfg.HeaderText,
		FooterText:           cfg.FooterText,
		Theme:                theme,
		IsPublic:             cfg.IsPublic,
		ShowUptimePercentage: cfg.ShowUptimePercentage,
		Enabled:              cfg.Enabled,
	}

	return req, nil
}

// validateTheme validates the theme value is one of the allowed presets.
func validateTheme(theme string) error {
	themeLower := strings.ToLower(theme)
	if !slices.Contains(validThemes, themeLower) {
		return clierrors.InvalidValueError("--theme", theme, clierrors.ValidThemes)
	}
	return nil
}

// slugRegex validates slug format: lowercase alphanumeric with hyphens, 3-63 chars.
var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

// validateSlug validates the slug format.
func validateSlug(slug string) error {
	if len(slug) < 3 {
		return fmt.Errorf("slug must be at least 3 characters")
	}
	if len(slug) > 63 {
		return fmt.Errorf("slug must be at most 63 characters")
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug must be lowercase alphanumeric with hyphens, start and end with alphanumeric")
	}
	return nil
}
