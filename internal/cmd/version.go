// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"fmt"

	"github.com/StackEye-IO/stackeye-cli/internal/version"
	"github.com/spf13/cobra"
)

// NewVersionCmd creates and returns the version command.
func NewVersionCmd() *cobra.Command {
	var shortFlag bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long: `Display version, git commit SHA, build date, and runtime information.

Use --short to display only the version number.`,
		// Override PersistentPreRunE to skip config loading.
		// The version command should work without a valid configuration.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			info := version.GetInfo()
			if shortFlag {
				fmt.Println(info.Short())
				return
			}
			fmt.Println(info.String())
		},
	}

	cmd.Flags().BoolVar(&shortFlag, "short", false, "print only the version number")

	return cmd
}
