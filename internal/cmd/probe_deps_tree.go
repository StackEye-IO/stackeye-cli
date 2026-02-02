// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/config"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// probeDepsTreeTimeout is the maximum time to wait for the API response.
const probeDepsTreeTimeout = 60 * time.Second

// NewProbeDepsTreeCmd creates and returns the probe deps tree subcommand.
// Task #8027: Implements ASCII tree visualization of probe dependencies.
func NewProbeDepsTreeCmd() *cobra.Command {
	var orgID string
	var probeID string
	var useASCII bool

	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Display dependency tree visualization",
		Long: `Display an ASCII tree visualization of probe dependencies for your organization.

The tree shows parent-child relationships between probes, where parent probes
represent infrastructure that child probes depend on.

Status colors (if terminal supports colors):
  UP          Green    - Probe is healthy
  DOWN        Red      - Probe is failing
  UNREACHABLE Yellow   - Probe is unreachable due to parent being down

Tree Structure:
  Root probes (no parents) appear at the top level
  Child probes are indented under their parents
  Orphan probes (no parents or children) are listed separately

Examples:
  # Show full dependency tree for current organization
  stackeye probe deps tree

  # Show tree for a specific organization
  stackeye probe deps tree --org 550e8400-e29b-41d4-a716-446655440000

  # Show subtree starting from a specific probe
  stackeye probe deps tree --probe 550e8400-e29b-41d4-a716-446655440000

  # Use ASCII characters instead of Unicode box-drawing
  stackeye probe deps tree --ascii

  # Output as JSON for scripting
  stackeye probe deps tree -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeDepsTreeCmd(cmd.Context(), orgID, probeID, useASCII)
		},
	}

	cmd.Flags().StringVar(&orgID, "org", "", "Organization ID (uses current organization if not specified)")
	cmd.Flags().StringVar(&probeID, "probe", "", "Start tree from this probe ID (optional)")
	cmd.Flags().BoolVar(&useASCII, "ascii", false, "Use ASCII characters instead of Unicode box-drawing")

	return cmd
}

// runProbeDepsTreeCmd executes the probe deps tree command logic.
func runProbeDepsTreeCmd(ctx context.Context, orgIDArg, probeIDArg string, useASCII bool) error {
	// Validate command line arguments first (before API client initialization)
	var parsedOrgID uuid.UUID
	var parsedProbeID uuid.UUID
	var err error

	// Parse optional organization ID
	if orgIDArg != "" {
		parsedOrgID, err = uuid.Parse(orgIDArg)
		if err != nil {
			return fmt.Errorf("invalid organization ID %q: must be a valid UUID", orgIDArg)
		}
	}

	// Parse optional probe ID for subtree filtering
	if probeIDArg != "" {
		parsedProbeID, err = uuid.Parse(probeIDArg)
		if err != nil {
			return fmt.Errorf("invalid probe ID %q: must be a valid UUID", probeIDArg)
		}
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// If no org ID provided, get it from the CLI config context
	if parsedOrgID == uuid.Nil {
		cfg, cfgErr := config.Load()
		if cfgErr != nil {
			return fmt.Errorf("failed to load config: %w", cfgErr)
		}
		currentCtx, ctxErr := cfg.GetCurrentContext()
		if ctxErr != nil {
			return fmt.Errorf("no current context: %w (run 'stackeye login' first)", ctxErr)
		}
		if currentCtx.OrganizationID == "" {
			return fmt.Errorf("no organization selected. Use 'stackeye org switch <org>' to select one")
		}
		parsedOrgID, err = uuid.Parse(currentCtx.OrganizationID)
		if err != nil {
			return fmt.Errorf("invalid organization ID in config: %w", err)
		}
	}

	// Fetch the dependency tree
	reqCtx, cancel := context.WithTimeout(ctx, probeDepsTreeTimeout)
	defer cancel()

	tree, err := client.GetOrganizationDependencyTree(reqCtx, apiClient, parsedOrgID)
	if err != nil {
		return fmt.Errorf("failed to get dependency tree: %w", err)
	}

	// Check output format
	printer := output.NewPrinter(nil)
	format := printer.Format()
	if format == sdkoutput.FormatJSON || format == sdkoutput.FormatYAML {
		return output.Print(tree)
	}

	// Build and print the ASCII tree
	return printDependencyTree(tree, parsedProbeID, useASCII)
}

// printDependencyTree renders the dependency tree in ASCII format.
func printDependencyTree(tree *client.DependencyTree, startProbeID uuid.UUID, useASCII bool) error {
	if len(tree.Nodes) == 0 {
		fmt.Println("No probes found in the organization.")
		return nil
	}

	// Build lookup maps
	nodeMap := make(map[uuid.UUID]*client.DependencyTreeNode)
	for i := range tree.Nodes {
		node := &tree.Nodes[i]
		nodeMap[node.ProbeID] = node
	}

	// Build parent -> children map
	childrenMap := make(map[uuid.UUID][]uuid.UUID)
	for _, edge := range tree.Edges {
		childrenMap[edge.FromProbeID] = append(childrenMap[edge.FromProbeID], edge.ToProbeID)
	}

	// Create tree printer
	treePrinter := output.NewTreePrinter(sdkoutput.ColorAuto, useASCII)

	// Determine which roots to display
	var rootsToDisplay []uuid.UUID
	if startProbeID != uuid.Nil {
		// Start from specific probe
		if _, exists := nodeMap[startProbeID]; !exists {
			return fmt.Errorf("probe %s not found in organization", startProbeID)
		}
		rootsToDisplay = []uuid.UUID{startProbeID}
	} else {
		// Use all root probes
		rootsToDisplay = tree.RootProbes
	}

	// Build tree nodes from root probes
	var treeNodes []*output.TreeNode
	visited := make(map[uuid.UUID]bool)

	for _, rootID := range rootsToDisplay {
		if visited[rootID] {
			continue
		}
		treeNode := buildTreeNode(rootID, nodeMap, childrenMap, visited)
		if treeNode != nil {
			treeNodes = append(treeNodes, treeNode)
		}
	}

	// Print the tree
	if len(treeNodes) == 0 {
		fmt.Println("No dependency tree to display (all probes are orphans or no dependencies configured).")
	} else {
		fmt.Println("Dependency Tree")
		fmt.Println("===============")
		fmt.Println()
		treePrinter.PrintTree(treeNodes)
	}

	// Print orphan probes (only if not filtering by specific probe)
	if startProbeID == uuid.Nil && len(tree.OrphanProbes) > 0 {
		var orphanNodes []*output.TreeNode
		for _, orphanID := range tree.OrphanProbes {
			if node, exists := nodeMap[orphanID]; exists {
				status := node.Status
				if node.IsUnreachable {
					status = "unreachable"
				}
				orphanNodes = append(orphanNodes, &output.TreeNode{
					Name:   node.Name,
					Status: status,
				})
			}
		}
		treePrinter.PrintOrphanSection("Orphan Probes (no dependencies)", orphanNodes)
	}

	return nil
}

// buildTreeNode recursively builds a tree node from the dependency data.
func buildTreeNode(probeID uuid.UUID, nodeMap map[uuid.UUID]*client.DependencyTreeNode,
	childrenMap map[uuid.UUID][]uuid.UUID, visited map[uuid.UUID]bool) *output.TreeNode {

	if visited[probeID] {
		// Avoid cycles (shouldn't happen with valid DAG but be safe)
		return nil
	}
	visited[probeID] = true

	node, exists := nodeMap[probeID]
	if !exists {
		return nil
	}

	status := node.Status
	if node.IsUnreachable {
		status = "unreachable"
	}

	treeNode := &output.TreeNode{
		Name:   node.Name,
		Status: status,
	}

	// Add children
	children := childrenMap[probeID]
	for _, childID := range children {
		childNode := buildTreeNode(childID, nodeMap, childrenMap, visited)
		if childNode != nil {
			treeNode.Children = append(treeNode.Children, childNode)
		}
	}

	return treeNode
}
