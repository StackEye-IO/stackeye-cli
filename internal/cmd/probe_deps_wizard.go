// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// probeDepsWizardTimeout is the maximum time to wait for API responses.
const probeDepsWizardTimeout = 60 * time.Second

// Infrastructure probe patterns
var infrastructurePatterns = regexp.MustCompile(`(?i)(router|firewall|gateway|vpn|dns|network|core|primary|main|load.?balancer|lb|haproxy|nginx|traefik)`)

// Application probe patterns
var applicationPatterns = regexp.MustCompile(`(?i)(api|app|web|service|database|db|cache|redis|postgres|mysql|mongo|backend|frontend|server|endpoint)`)

// NewProbeDepsWizardCmd creates and returns the probe deps wizard subcommand.
// Task #8028: Implements guided dependency setup wizard.
func NewProbeDepsWizardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive dependency setup wizard",
		Long: `Interactive wizard for setting up probe dependencies.

The wizard helps you configure hierarchical alerting by:
  1. Auto-detecting infrastructure probes (routers, firewalls, etc.)
  2. Auto-detecting application probes (APIs, databases, etc.)
  3. Creating dependencies where applications depend on infrastructure

This reduces alert noise - when your router goes down, you don't want
separate alerts for every service behind it.

Infrastructure probes are detected by naming patterns:
  router, firewall, gateway, vpn, dns, network, core, load-balancer

Application probes are detected by naming patterns:
  api, app, web, service, database, cache, backend, frontend

For non-interactive environments, use individual commands:
  stackeye probe deps add <probe-id> --parent <parent-id>

Examples:
  # Launch the interactive dependency wizard
  stackeye probe deps wizard`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeDepsWizard(cmd.Context())
		},
	}

	return cmd
}

// runProbeDepsWizard executes the interactive dependency wizard.
func runProbeDepsWizard(ctx context.Context) error {
	// Check for non-interactive mode
	if GetNoInput() {
		return runProbeDepsWizardNonInteractive()
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	fmt.Println("=== Probe Dependency Setup Wizard ===")
	fmt.Println()
	fmt.Println("This wizard will help you set up hierarchical alerting to reduce")
	fmt.Println("notification noise when infrastructure failures affect multiple services.")
	fmt.Println()

	// Fetch all probes
	reqCtx, cancel := context.WithTimeout(ctx, probeDepsWizardTimeout)
	defer cancel()

	probesResp, err := client.ListProbes(reqCtx, apiClient, nil)
	if err != nil {
		return fmt.Errorf("failed to list probes: %w", err)
	}
	probes := probesResp.Probes

	if len(probes) == 0 {
		fmt.Println("No probes found. Create some probes first:")
		fmt.Println("  stackeye probe create --name \"My Probe\" --url https://example.com")
		return nil
	}

	if len(probes) < 2 {
		fmt.Println("You need at least 2 probes to create dependencies.")
		fmt.Println("Create more probes first:")
		fmt.Println("  stackeye probe create --name \"My Probe\" --url https://example.com")
		return nil
	}

	// Step 1: Identify infrastructure probes
	fmt.Println("Step 1/3: Identify Infrastructure Probes")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Infrastructure probes represent core systems (routers, firewalls,")
	fmt.Println("load balancers) that other services depend on.")
	fmt.Println()

	// Auto-detect potential infrastructure probes
	var suggestedInfra []client.Probe
	for _, p := range probes {
		if infrastructurePatterns.MatchString(p.Name) {
			suggestedInfra = append(suggestedInfra, p)
		}
	}

	// Build options for multi-select
	var infraOptions []string
	var infraDefaults []string
	for _, p := range probes {
		option := fmt.Sprintf("%s (%s)", p.Name, p.ID)
		infraOptions = append(infraOptions, option)
		// Pre-select auto-detected infrastructure probes
		for _, si := range suggestedInfra {
			if si.ID == p.ID {
				infraDefaults = append(infraDefaults, option)
				break
			}
		}
	}

	if len(suggestedInfra) > 0 {
		fmt.Printf("Auto-detected %d potential infrastructure probes (pre-selected):\n", len(suggestedInfra))
		for _, p := range suggestedInfra {
			fmt.Printf("  - %s\n", p.Name)
		}
		fmt.Println()
	}

	selectedInfra, err := interactive.AskMultiSelect(&interactive.MultiSelectPromptOptions{
		Message:  "Select infrastructure probes (parents in the dependency tree):",
		Options:  infraOptions,
		Defaults: infraDefaults,
		PageSize: 10,
	})
	if err != nil {
		if err == interactive.ErrPromptCancelled {
			fmt.Println("\nWizard cancelled.")
			return nil
		}
		return fmt.Errorf("selection failed: %w", err)
	}

	if len(selectedInfra) == 0 {
		fmt.Println("\nNo infrastructure probes selected. No dependencies to create.")
		return nil
	}

	// Parse selected infrastructure probe IDs
	infraProbeIDs := parseSelectedProbeIDs(selectedInfra, probes)

	// Step 2: Identify application probes
	fmt.Println()
	fmt.Println("Step 2/3: Identify Application Probes")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Println("Application probes represent services (APIs, databases, web apps)")
	fmt.Println("that depend on infrastructure.")
	fmt.Println()

	// Build options excluding infrastructure probes
	var appOptions []string
	var appDefaults []string
	for _, p := range probes {
		// Skip already selected infrastructure probes
		isInfra := false
		for _, infraID := range infraProbeIDs {
			if p.ID == infraID {
				isInfra = true
				break
			}
		}
		if isInfra {
			continue
		}

		option := fmt.Sprintf("%s (%s)", p.Name, p.ID)
		appOptions = append(appOptions, option)
		// Pre-select auto-detected application probes
		if applicationPatterns.MatchString(p.Name) {
			appDefaults = append(appDefaults, option)
		}
	}

	if len(appOptions) == 0 {
		fmt.Println("All probes were selected as infrastructure. No application probes available.")
		return nil
	}

	selectedApps, err := interactive.AskMultiSelect(&interactive.MultiSelectPromptOptions{
		Message:  "Select application probes (children in the dependency tree):",
		Options:  appOptions,
		Defaults: appDefaults,
		PageSize: 10,
	})
	if err != nil {
		if err == interactive.ErrPromptCancelled {
			fmt.Println("\nWizard cancelled.")
			return nil
		}
		return fmt.Errorf("selection failed: %w", err)
	}

	if len(selectedApps) == 0 {
		fmt.Println("\nNo application probes selected. No dependencies to create.")
		return nil
	}

	// Parse selected application probe IDs
	appProbeIDs := parseSelectedProbeIDs(selectedApps, probes)

	// Step 3: Create dependencies
	fmt.Println()
	fmt.Println("Step 3/3: Create Dependencies")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Println("Dependencies to create:")
	fmt.Println()

	// Show what will be created
	var depsToCreate []struct{ childID, parentID, childName, parentName string }
	for _, appID := range appProbeIDs {
		appName := getProbeNameByID(appID, probes)
		for _, infraID := range infraProbeIDs {
			infraName := getProbeNameByID(infraID, probes)
			depsToCreate = append(depsToCreate, struct{ childID, parentID, childName, parentName string }{
				childID:    appID.String(),
				parentID:   infraID.String(),
				childName:  appName,
				parentName: infraName,
			})
			fmt.Printf("  %s → %s\n", appName, infraName)
		}
	}

	fmt.Printf("\nTotal: %d dependencies\n\n", len(depsToCreate))

	// Confirm
	confirmed, err := interactive.AskConfirm(&interactive.ConfirmPromptOptions{
		Message: "Create these dependencies?",
		Default: true,
	})
	if err != nil {
		if err == interactive.ErrPromptCancelled {
			fmt.Println("\nWizard cancelled.")
			return nil
		}
		return fmt.Errorf("confirmation failed: %w", err)
	}

	if !confirmed {
		fmt.Println("Operation cancelled.")
		return nil
	}

	// Create the dependencies
	fmt.Println("\nCreating dependencies...")
	var created, skipped, failed int
	for _, dep := range depsToCreate {
		reqCtx, cancel := context.WithTimeout(ctx, probeDepsWizardTimeout)
		_, createErr := client.AddProbeDependency(reqCtx, apiClient,
			parseUUID(dep.childID), parseUUID(dep.parentID))
		cancel()

		if createErr != nil {
			errMsg := createErr.Error()
			if strings.Contains(errMsg, "dependency_exists") {
				fmt.Printf("  [skip] %s → %s (already exists)\n", dep.childName, dep.parentName)
				skipped++
			} else {
				fmt.Printf("  [fail] %s → %s: %v\n", dep.childName, dep.parentName, createErr)
				failed++
			}
		} else {
			fmt.Printf("  [ok] %s → %s\n", dep.childName, dep.parentName)
			created++
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Created: %d\n", created)
	if skipped > 0 {
		fmt.Printf("  Skipped (already exist): %d\n", skipped)
	}
	if failed > 0 {
		fmt.Printf("  Failed: %d\n", failed)
	}

	fmt.Println()
	fmt.Println("Run 'stackeye probe deps tree' to visualize your dependency tree.")

	return nil
}

// runProbeDepsWizardNonInteractive handles non-interactive mode.
func runProbeDepsWizardNonInteractive() error {
	fmt.Println("The dependency wizard requires interactive mode.")
	fmt.Println()
	fmt.Println("For non-interactive dependency management, use:")
	fmt.Println("  stackeye probe deps add <probe-id> --parent <parent-id>")
	fmt.Println("  stackeye probe deps remove <probe-id> --parent <parent-id>")
	fmt.Println("  stackeye probe deps tree")
	fmt.Println()
	fmt.Println("Or enable interactive mode by removing --no-input flag.")
	return nil
}

// parseSelectedProbeIDs extracts probe IDs from selection strings.
func parseSelectedProbeIDs(selected []string, probes []client.Probe) []uuid.UUID {
	var ids []uuid.UUID
	for _, sel := range selected {
		for _, p := range probes {
			option := fmt.Sprintf("%s (%s)", p.Name, p.ID)
			if sel == option {
				ids = append(ids, p.ID)
				break
			}
		}
	}
	return ids
}

// getProbeNameByID finds a probe name by its ID.
func getProbeNameByID(id uuid.UUID, probes []client.Probe) string {
	for _, p := range probes {
		if p.ID == id {
			return p.Name
		}
	}
	return id.String()
}

// parseUUID parses a UUID string, panics on error (should never happen with validated IDs).
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("invalid UUID %q: %v", s, err))
	}
	return id
}
