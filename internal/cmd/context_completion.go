// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"

	"github.com/StackEye-IO/stackeye-cli/internal/config"
	"github.com/spf13/cobra"
)

// ContextCompletion returns a cobra.ValidArgsFunction that provides context name completions.
// It reads available contexts from the local config file and filters by the user's input prefix.
// This function makes no network calls, ensuring fast completion times.
//
// Example usage:
//
//	cmd := &cobra.Command{
//	    Use:               "use <name>",
//	    ValidArgsFunction: ContextCompletion(),
//	}
func ContextCompletion() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first positional argument (context name)
		if len(args) >= 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Try to load configuration
		cfg, err := config.Load()
		if err != nil {
			// Config load error - return empty completions silently
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		if cfg == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Get available context names
		names := cfg.ContextNames()
		if len(names) == 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Build completion list
		var completions []string
		loweredComplete := strings.ToLower(toComplete)

		for _, name := range names {
			// Filter by prefix (case-insensitive)
			loweredName := strings.ToLower(name)

			if toComplete == "" || strings.HasPrefix(loweredName, loweredComplete) {
				// Get context details for description
				ctx, err := cfg.GetContext(name)
				if err != nil || ctx == nil {
					// Include without description if context details unavailable
					completions = append(completions, name)
					continue
				}

				// Build description with organization name if available
				var desc string
				if ctx.OrganizationName != "" {
					desc = ctx.OrganizationName
				} else {
					desc = "(no org)"
				}

				// Mark current context
				if cfg.CurrentContext == name {
					desc += " [current]"
				}

				// Add completion with description (format: "value\tdescription")
				completion := name + "\t" + desc
				completions = append(completions, completion)
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// ContextNameCompletion is an alias for ContextCompletion for clarity in command definitions.
// Use this when the command expects a context name argument.
var ContextNameCompletion = ContextCompletion
