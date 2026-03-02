// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	cliinteractive "github.com/StackEye-IO/stackeye-cli/internal/interactive"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const (
	probeGroupAPITimeout      = 30 * time.Second
	probeGroupListProbesLimit = 100
)

// NewProbeGroupCmd creates and returns the probe group command.
func NewProbeGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "group",
		Short:   "Manage probe groups",
		Aliases: []string{"groups"},
		Long: `Manage probe groups for organizing related monitoring probes.

Probe groups help you segment probes by service, environment, or ownership.
Use groups to quickly manage membership and view grouped probe status.

Commands:
  create        Create a new probe group
  list          List probe groups
  get           Get a specific probe group
  update        Update group metadata
  delete        Delete a probe group
  add-probe     Add a probe to a group
  remove-probe  Remove a probe from a group
  list-probes   List probes currently in a group

Examples:
  # Create a probe group
  stackeye probe group create --name "Cluster fruition-infra-doks-sfo3"

  # Add a probe to a group
  stackeye probe group add-probe <group-id> <probe-id>

  # List probes in a group
  stackeye probe group list-probes <group-id>

For more information about a specific command:
  stackeye probe group [command] --help`,
	}

	cmd.AddCommand(NewProbeGroupCreateCmd())
	cmd.AddCommand(NewProbeGroupListCmd())
	cmd.AddCommand(NewProbeGroupGetCmd())
	cmd.AddCommand(NewProbeGroupUpdateCmd())
	cmd.AddCommand(NewProbeGroupDeleteCmd())
	cmd.AddCommand(NewProbeGroupAddProbeCmd())
	cmd.AddCommand(NewProbeGroupRemoveProbeCmd())
	cmd.AddCommand(NewProbeGroupListProbesCmd())

	return cmd
}

type probeGroupCreateFlags struct {
	name        string
	description string
	groupType   string
	labels      []string
}

type groupLabelSelectorCondition struct {
	Key    string   `json:"key"`
	Op     string   `json:"op"`
	Value  string   `json:"value,omitempty"`
	Values []string `json:"values,omitempty"`
}

// NewProbeGroupCreateCmd creates and returns the probe group create command.
func NewProbeGroupCreateCmd() *cobra.Command {
	flags := &probeGroupCreateFlags{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new probe group",
		Long: `Create a new probe group.

Required Flags:
  --name   Group name (must be unique within your organization)

Optional Flags:
  --type         Group type: static or dynamic (default: static)
  --label        Label selector condition (repeat for multiple conditions)
  --description  Human-readable group description

Examples:
  stackeye probe group create --name "Cluster fruition-infra-doks-sfo3"
  stackeye probe group create --name "Core APIs" --description "Production API surfaces"
  stackeye probe group create --name "Prod API Dynamic" --type dynamic --label env=prod --label service in api|edge`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeGroupCreate(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVar(&flags.name, "name", "", "group name (required)")
	cmd.Flags().StringVar(&flags.description, "description", "", "group description")
	cmd.Flags().StringVar(&flags.groupType, "type", string(client.ProbeGroupTypeStatic), "group type: static or dynamic")
	cmd.Flags().StringSliceVar(&flags.labels, "label", nil, "group selector label (repeatable): key=value, key in (a|b), or key (exists)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func runProbeGroupCreate(ctx context.Context, flags *probeGroupCreateFlags) error {
	if strings.TrimSpace(flags.name) == "" {
		return fmt.Errorf("--name is required")
	}
	groupType := strings.ToLower(strings.TrimSpace(flags.groupType))
	if groupType != string(client.ProbeGroupTypeStatic) && groupType != string(client.ProbeGroupTypeDynamic) {
		return fmt.Errorf("--type must be one of: static, dynamic")
	}

	selectorJSON, err := buildGroupLabelSelector(flags.labels)
	if err != nil {
		return err
	}
	if groupType == string(client.ProbeGroupTypeDynamic) && selectorJSON == nil {
		return fmt.Errorf("dynamic groups require at least one --label selector")
	}
	if groupType == string(client.ProbeGroupTypeStatic) && selectorJSON != nil {
		return fmt.Errorf("static groups cannot set --label selectors; use --type dynamic")
	}

	if GetDryRun() {
		details := []string{"Name", flags.name}
		if strings.TrimSpace(flags.description) != "" {
			details = append(details, "Description", flags.description)
		}
		details = append(details, "Type", groupType)
		if len(flags.labels) > 0 {
			details = append(details, "Selector Labels", strings.Join(flags.labels, ","))
		}
		dryrun.PrintAction("create", "probe group", details...)
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	req := &client.CreateProbeGroupRequest{Name: flags.name, Type: groupType}
	if strings.TrimSpace(flags.description) != "" {
		req.Description = &flags.description
	}
	if selectorJSON != nil {
		req.LabelSelector = selectorJSON
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeGroupAPITimeout)
	defer cancel()

	group, err := client.CreateProbeGroup(reqCtx, apiClient, req)
	if err != nil {
		if friendly := probeGroupUniqueNameError(err, flags.name); friendly != nil {
			return friendly
		}
		return fmt.Errorf("failed to create probe group: %w", err)
	}

	return output.Print(group)
}

// NewProbeGroupListCmd creates and returns the probe group list command.
func NewProbeGroupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List probe groups",
		Aliases: []string{"ls"},
		Long: `List probe groups in your organization.

Table Columns:
  ID           Group UUID
  NAME         Group name
  DESCRIPTION  Optional group description
  PROBES       Number of probes in group
  UPDATED      Last update timestamp

Examples:
  stackeye probe group list
  stackeye probe group list -o json
  stackeye probe group list -o wide`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeGroupList(cmd.Context())
		},
	}

	return cmd
}

func runProbeGroupList(ctx context.Context) error {
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeGroupAPITimeout)
	defer cancel()

	result, err := client.ListProbeGroups(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to list probe groups: %w", err)
	}

	if len(result.Groups) == 0 {
		return output.PrintEmpty("No probe groups found. Create one with 'stackeye probe group create'")
	}

	return output.PrintProbeGroups(result.Groups)
}

// NewProbeGroupGetCmd creates and returns the probe group get command.
func NewProbeGroupGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <group-id>",
		Short: "Get probe group details",
		Long: `Get details for a single probe group.

Examples:
  stackeye probe group get <group-id>
  stackeye probe group get <group-id> -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeGroupGet(cmd.Context(), args[0])
		},
	}

	return cmd
}

func runProbeGroupGet(ctx context.Context, groupIDArg string) error {
	groupID, err := parseProbeGroupID(groupIDArg)
	if err != nil {
		return err
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeGroupAPITimeout)
	defer cancel()

	group, err := client.GetProbeGroup(reqCtx, apiClient, groupID)
	if err != nil {
		return fmt.Errorf("failed to get probe group: %w", err)
	}

	return output.Print(group)
}

type probeGroupUpdateFlags struct {
	name        string
	description string
	labels      []string
	clearLabels bool
}

// NewProbeGroupUpdateCmd creates and returns the probe group update command.
func NewProbeGroupUpdateCmd() *cobra.Command {
	flags := &probeGroupUpdateFlags{}
	cmd := &cobra.Command{
		Use:   "update <group-id>",
		Short: "Update probe group metadata",
		Long: `Update one or more fields on a probe group.

At least one update flag is required.

Examples:
  stackeye probe group update <group-id> --name "Platform Core"
  stackeye probe group update <group-id> --description "Services managed by SRE"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeGroupUpdate(cmd, args[0], flags)
		},
	}

	cmd.Flags().StringVar(&flags.name, "name", "", "updated group name")
	cmd.Flags().StringVar(&flags.description, "description", "", "updated group description")
	cmd.Flags().StringSliceVar(&flags.labels, "label", nil, "set selector labels (repeatable): key=value, key in (a|b), or key (exists)")
	cmd.Flags().BoolVar(&flags.clearLabels, "clear-labels", false, "clear dynamic label selector")

	return cmd
}

func runProbeGroupUpdate(cmd *cobra.Command, groupIDArg string, flags *probeGroupUpdateFlags) error {
	groupID, err := parseProbeGroupID(groupIDArg)
	if err != nil {
		return err
	}

	var req client.UpdateProbeGroupRequest
	hasUpdate := false

	if cmd.Flags().Changed("name") {
		if strings.TrimSpace(flags.name) == "" {
			return fmt.Errorf("--name cannot be empty")
		}
		req.Name = &flags.name
		hasUpdate = true
	}
	if cmd.Flags().Changed("description") {
		req.Description = &flags.description
		hasUpdate = true
	}
	if cmd.Flags().Changed("label") && flags.clearLabels {
		return fmt.Errorf("--label and --clear-labels cannot be used together")
	}
	if cmd.Flags().Changed("label") {
		selectorJSON, selectorErr := buildGroupLabelSelector(flags.labels)
		if selectorErr != nil {
			return selectorErr
		}
		if selectorJSON == nil {
			return fmt.Errorf("--label requires at least one value when set")
		}
		req.LabelSelector = selectorJSON
		hasUpdate = true
	}
	if flags.clearLabels {
		emptySelector := json.RawMessage("[]")
		req.LabelSelector = &emptySelector
		hasUpdate = true
	}
	if !hasUpdate {
		return fmt.Errorf("at least one update flag is required (--name, --description, --label, or --clear-labels)")
	}

	if GetDryRun() {
		details := []string{"Group ID", groupID.String()}
		if req.Name != nil {
			details = append(details, "Name", *req.Name)
		}
		if req.Description != nil {
			details = append(details, "Description", *req.Description)
		}
		if cmd.Flags().Changed("label") {
			details = append(details, "Selector Labels", strings.Join(flags.labels, ","))
		}
		if flags.clearLabels {
			details = append(details, "Selector Labels", "(cleared)")
		}
		dryrun.PrintAction("update", "probe group", details...)
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(cmd.Context(), probeGroupAPITimeout)
	defer cancel()

	group, err := client.UpdateProbeGroup(reqCtx, apiClient, groupID, &req)
	if err != nil {
		if req.Name != nil {
			if friendly := probeGroupUniqueNameError(err, *req.Name); friendly != nil {
				return friendly
			}
		}
		return fmt.Errorf("failed to update probe group: %w", err)
	}

	return output.Print(group)
}

type probeGroupDeleteFlags struct {
	yes bool
}

// NewProbeGroupDeleteCmd creates and returns the probe group delete command.
func NewProbeGroupDeleteCmd() *cobra.Command {
	flags := &probeGroupDeleteFlags{}
	cmd := &cobra.Command{
		Use:   "delete <group-id>",
		Short: "Delete a probe group",
		Long: `Delete a probe group.

This removes the group and associated membership links. Probes are not deleted.

Examples:
  stackeye probe group delete <group-id>
  stackeye probe group delete <group-id> --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeGroupDelete(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

func runProbeGroupDelete(ctx context.Context, groupIDArg string, flags *probeGroupDeleteFlags) error {
	groupID, err := parseProbeGroupID(groupIDArg)
	if err != nil {
		return err
	}

	if GetDryRun() {
		dryrun.PrintAction("delete", "probe group", "Group ID", groupID.String())
		return nil
	}

	confirmed, err := cliinteractive.Confirm(
		fmt.Sprintf("Are you sure you want to delete probe group %s?", groupID),
		cliinteractive.WithYesFlag(flags.yes),
	)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Delete cancelled.")
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeGroupAPITimeout)
	defer cancel()

	if err := client.DeleteProbeGroup(reqCtx, apiClient, groupID); err != nil {
		return fmt.Errorf("failed to delete probe group: %w", err)
	}

	fmt.Printf("Deleted probe group %s\n", groupID)
	return nil
}

// NewProbeGroupAddProbeCmd creates and returns the add-probe command.
func NewProbeGroupAddProbeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "add-probe <group-id> <probe-id-or-name>",
		Short:             "Add a probe to a group",
		ValidArgsFunction: ProbeCompletion(),
		Long: `Add a probe to a static group.

The probe can be specified by UUID or by name. Name resolution follows the
same disambiguation behavior as other probe commands.

Examples:
  stackeye probe group add-probe <group-id> <probe-id>
  stackeye probe group add-probe <group-id> "Production API"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeGroupAddProbe(cmd.Context(), args[0], args[1])
		},
	}

	return cmd
}

func runProbeGroupAddProbe(ctx context.Context, groupIDArg, probeIDOrName string) error {
	groupID, err := parseProbeGroupID(groupIDArg)
	if err != nil {
		return err
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	probeID, err := ResolveProbeID(ctx, apiClient, probeIDOrName)
	if err != nil {
		return fmt.Errorf("failed to resolve probe: %w", err)
	}

	if GetDryRun() {
		dryrun.PrintAction("add probe to", "probe group", "Group ID", groupID.String(), "Probe ID", probeID.String())
		return nil
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeGroupAPITimeout)
	defer cancel()
	if err := client.AddProbeGroupMembers(reqCtx, apiClient, groupID, &client.AddProbeGroupMembersRequest{ProbeIDs: []uuid.UUID{probeID}}); err != nil {
		return fmt.Errorf("failed to add probe to group: %w", err)
	}

	fmt.Printf("Added probe %s to group %s\n", probeID, groupID)
	return nil
}

// NewProbeGroupRemoveProbeCmd creates and returns the remove-probe command.
func NewProbeGroupRemoveProbeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "remove-probe <group-id> <probe-id-or-name>",
		Short:             "Remove a probe from a group",
		ValidArgsFunction: ProbeCompletion(),
		Long: `Remove a probe from a static group.

The probe can be specified by UUID or by name. Name resolution follows the
same disambiguation behavior as other probe commands.

Examples:
  stackeye probe group remove-probe <group-id> <probe-id>
  stackeye probe group remove-probe <group-id> "Production API"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeGroupRemoveProbe(cmd.Context(), args[0], args[1])
		},
	}

	return cmd
}

func runProbeGroupRemoveProbe(ctx context.Context, groupIDArg, probeIDOrName string) error {
	groupID, err := parseProbeGroupID(groupIDArg)
	if err != nil {
		return err
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	probeID, err := ResolveProbeID(ctx, apiClient, probeIDOrName)
	if err != nil {
		return fmt.Errorf("failed to resolve probe: %w", err)
	}

	if GetDryRun() {
		dryrun.PrintAction("remove probe from", "probe group", "Group ID", groupID.String(), "Probe ID", probeID.String())
		return nil
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeGroupAPITimeout)
	defer cancel()
	if err := client.RemoveProbeGroupMember(reqCtx, apiClient, groupID, probeID); err != nil {
		return fmt.Errorf("failed to remove probe from group: %w", err)
	}

	fmt.Printf("Removed probe %s from group %s\n", probeID, groupID)
	return nil
}

type probeGroupListProbesFlags struct {
	page  int
	limit int
}

// NewProbeGroupListProbesCmd creates and returns the list-probes command.
func NewProbeGroupListProbesCmd() *cobra.Command {
	flags := &probeGroupListProbesFlags{}

	cmd := &cobra.Command{
		Use:   "list-probes <group-id>",
		Short: "List probes in a group",
		Long: `List probes currently belonging to a group.

Table Columns:
  ID      Probe UUID
  NAME    Probe name
  URL     Probe target URL
  STATUS  Current probe status

Pagination Flags:
  --page   Page number (1-indexed)
  --limit  Results per page (max 100)

Examples:
  stackeye probe group list-probes <group-id>
  stackeye probe group list-probes <group-id> --page 2 --limit 25
  stackeye probe group list-probes <group-id> -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeGroupListProbes(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")

	return cmd
}

func runProbeGroupListProbes(ctx context.Context, groupIDArg string, flags *probeGroupListProbesFlags) error {
	if flags.page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.page)
	}
	if flags.limit < 1 || flags.limit > probeGroupListProbesLimit {
		return fmt.Errorf("invalid limit %d: must be between 1 and %d", flags.limit, probeGroupListProbesLimit)
	}

	groupID, err := parseProbeGroupID(groupIDArg)
	if err != nil {
		return err
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, probeGroupAPITimeout)
	defer cancel()

	probeIDsResp, err := client.GetProbeGroupProbes(reqCtx, apiClient, groupID)
	if err != nil {
		return fmt.Errorf("failed to list group probes: %w", err)
	}

	if len(probeIDsResp.ProbeIDs) == 0 {
		return output.PrintEmpty("No probes found in this group")
	}

	start := (flags.page - 1) * flags.limit
	if start >= len(probeIDsResp.ProbeIDs) {
		return output.PrintEmpty("No probes found for the requested page")
	}
	end := start + flags.limit
	if end > len(probeIDsResp.ProbeIDs) {
		end = len(probeIDsResp.ProbeIDs)
	}
	pageProbeIDs := probeIDsResp.ProbeIDs[start:end]

	probes := make([]client.Probe, 0, len(pageProbeIDs))
	for _, probeID := range pageProbeIDs {
		probe, getErr := client.GetProbe(reqCtx, apiClient, probeID, "")
		if getErr != nil {
			return fmt.Errorf("failed to fetch probe %s: %w", probeID, getErr)
		}
		probes = append(probes, *probe)
	}

	return output.PrintGroupProbes(probes)
}

func parseProbeGroupID(value string) (uuid.UUID, error) {
	groupID, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid group ID %q: must be a valid UUID", value)
	}
	return groupID, nil
}

func probeGroupUniqueNameError(err error, name string) error {
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		return nil
	}
	if !apiErr.IsConflict() {
		return nil
	}
	if strings.Contains(strings.ToLower(apiErr.Message), "name") || strings.Contains(strings.ToLower(apiErr.Message), "exists") {
		return fmt.Errorf("group name %q already exists", name)
	}
	return nil
}

func buildGroupLabelSelector(labels []string) (*json.RawMessage, error) {
	if len(labels) == 0 {
		return nil, nil
	}

	conditions := make([]groupLabelSelectorCondition, 0, len(labels))
	for _, raw := range labels {
		token := strings.TrimSpace(raw)
		if token == "" {
			return nil, fmt.Errorf("invalid --label value: empty label is not allowed")
		}

		if strings.Contains(token, "=") {
			parts := strings.SplitN(token, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key == "" {
				return nil, fmt.Errorf("invalid --label %q: key cannot be empty", raw)
			}
			if value == "" {
				return nil, fmt.Errorf("invalid --label %q: value cannot be empty for key=value selector", raw)
			}
			conditions = append(conditions, groupLabelSelectorCondition{Key: key, Op: "eq", Value: value})
			continue
		}

		if strings.Contains(token, "|") {
			parts := strings.SplitN(token, " in ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid --label %q: use \"key in value1|value2\"", raw)
			}
			key := strings.TrimSpace(parts[0])
			valueExpr := strings.TrimSpace(parts[1])
			if key == "" || valueExpr == "" {
				return nil, fmt.Errorf("invalid --label %q: key and values are required", raw)
			}
			valueParts := strings.Split(valueExpr, "|")
			values := make([]string, 0, len(valueParts))
			for _, v := range valueParts {
				trimmed := strings.TrimSpace(v)
				if trimmed == "" {
					return nil, fmt.Errorf("invalid --label %q: empty value in in-selector", raw)
				}
				values = append(values, trimmed)
			}
			conditions = append(conditions, groupLabelSelectorCondition{Key: key, Op: "in", Values: values})
			continue
		}

		conditions = append(conditions, groupLabelSelectorCondition{Key: token, Op: "exists"})
	}

	selectorBytes, err := json.Marshal(conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal label selector: %w", err)
	}
	selector := json.RawMessage(selectorBytes)
	return &selector, nil
}
