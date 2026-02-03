// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// probeCompletionTimeout is the maximum time to wait for the API when fetching completions.
// Keep this short to avoid slowing down tab completion.
const probeCompletionTimeout = 5 * time.Second

// ProbeCompletion returns a cobra.ValidArgsFunction that provides probe name/ID completions.
// It fetches the probe list from the API and filters by the user's input prefix.
// On error (including offline), it returns an empty list to allow graceful degradation.
//
// Example usage:
//
//	cmd := &cobra.Command{
//	    Use:               "get <id>",
//	    ValidArgsFunction: ProbeCompletion(),
//	}
func ProbeCompletion() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first positional argument (probe ID)
		if len(args) >= 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Try to get the API client
		apiClient, err := api.GetClient()
		if err != nil {
			// Not authenticated or config error - return empty completions silently
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Create a context with timeout for the API call
		// Handle nil context (can happen during shell completion)
		parentCtx := cmd.Context()
		if parentCtx == nil {
			parentCtx = context.Background()
		}
		ctx, cancel := context.WithTimeout(parentCtx, probeCompletionTimeout)
		defer cancel()

		// Fetch probes from the API
		opts := &client.ListProbesOptions{
			Limit: 100, // Reasonable limit for completions
		}

		// If user has typed something, use it as a search filter
		if toComplete != "" {
			opts.Search = toComplete
		}

		resp, err := client.ListProbes(ctx, apiClient, opts)
		if err != nil {
			// API error (offline, network issues, etc.) - return empty completions silently
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Build completion list
		var completions []string
		for _, probe := range resp.Probes {
			// Filter by prefix (case-insensitive)
			loweredName := strings.ToLower(probe.Name)
			loweredComplete := strings.ToLower(toComplete)

			if toComplete == "" || strings.HasPrefix(loweredName, loweredComplete) {
				// Add probe name as completion with description
				completion := fmt.Sprintf("%s\t%s (%s)", probe.Name, probe.CheckType, probe.Status)
				completions = append(completions, completion)
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// ProbeIDCompletion is an alias for ProbeCompletion for clarity in command definitions.
// Use this when the command expects a probe ID argument.
var ProbeIDCompletion = ProbeCompletion
