// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// newPowerShellCompletionCmd creates the PowerShell completion subcommand.
// This generates a PowerShell completion script that can be sourced to enable
// tab-completion for the StackEye CLI on Windows.
func newPowerShellCompletionCmd() *cobra.Command {
	var noDescriptions bool

	cmd := &cobra.Command{
		Use:   "powershell",
		Short: "Generate PowerShell completion script",
		Long: `Generate PowerShell completion script for StackEye CLI.

To enable PowerShell completion, add the generated script to your PowerShell profile:

  # Find your profile path
  $PROFILE

  # Create profile if it doesn't exist
  if (!(Test-Path -Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force
  }

  # Add completion to your profile
  stackeye completion powershell >> $PROFILE

  # Or save to a separate file and source it from your profile
  stackeye completion powershell > "$env:USERPROFILE\.stackeye-completion.ps1"
  # Then add to $PROFILE: . "$env:USERPROFILE\.stackeye-completion.ps1"

After modifying your profile, restart PowerShell or reload the profile:

  . $PROFILE

Requirements:
  - PowerShell 5.0+ (Windows PowerShell) or PowerShell Core 6.0+

The --no-descriptions flag disables completion descriptions for a cleaner
but less informative completion experience.
`,
		Example: `  # Generate completion script to stdout
  stackeye completion powershell

  # Append to PowerShell profile
  stackeye completion powershell >> $PROFILE

  # Save to separate file
  stackeye completion powershell > "$env:USERPROFILE\.stackeye-completion.ps1"

  # Generate without descriptions (shorter output)
  stackeye completion powershell --no-descriptions`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate completion script to stdout
			// GenPowerShellCompletionWithDesc includes descriptions for completions
			// GenPowerShellCompletion omits descriptions
			if noDescriptions {
				return rootCmd.GenPowerShellCompletion(os.Stdout)
			}
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		},
	}

	cmd.Flags().BoolVar(&noDescriptions, "no-descriptions", false, "disable completion descriptions")

	return cmd
}
