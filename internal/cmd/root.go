// Package cmd implements the CLI commands for StackEye.
//
// This package contains the root command and all subcommands for the
// StackEye CLI tool. Commands are organized hierarchically using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	clioutput "github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/spf13/cobra"
)

// Global flag variables
var (
	configFile      string
	contextOverride string
	debugFlag       bool
	outputFormat    string
	noColor         bool
	noInput         bool
	dryRun          bool
)

// loadedConfig holds the configuration loaded during PersistentPreRunE.
// Subcommands access this via GetConfig().
var loadedConfig *config.Config

// rootCmd is the base command for the CLI.
var rootCmd = &cobra.Command{
	Use:   "stackeye",
	Short: "StackEye CLI - Eye on your stack",
	Long: `StackEye CLI provides command-line access to the StackEye uptime monitoring platform.

Manage probes, alerts, notification channels, and organizations directly from
your terminal. Integrate monitoring into your CI/CD pipelines and automation
workflows.

Get started:
  stackeye login              Authenticate with your StackEye account
  stackeye probe list         List all monitoring probes
  stackeye alert list         View current alerts

For more information about a command:
  stackeye [command] --help`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return loadConfig()
	},
}

func init() {
	// Wire up the API client helper to use our config getter
	api.SetConfigGetter(GetConfig)

	// Wire up the output package to use our config getter
	clioutput.SetConfigGetter(GetConfig)

	// Register subcommands
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewLoginCmd())
	rootCmd.AddCommand(NewLogoutCmd())
	rootCmd.AddCommand(NewWhoamiCmd())
	rootCmd.AddCommand(NewConfigCmd())
	rootCmd.AddCommand(NewContextCmd())
	rootCmd.AddCommand(NewCompletionCmd())
	rootCmd.AddCommand(NewProbeCmd())
	rootCmd.AddCommand(NewAlertCmd())

	// Register persistent flags available to all commands
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path (default: ~/.config/stackeye/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&contextOverride, "context", "", "override current context from config")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debug output")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format: table, json, yaml, wide")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&noInput, "no-input", false, "disable interactive prompts")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without executing")

	// Initialize custom help system with colored output and grouped commands
	InitHelp(rootCmd, &HelpConfig{
		ColorManager: output.NewColorManager(output.ColorAuto),
		Writer:       os.Stdout,
	})
}

// loadConfig loads the configuration file and applies flag overrides.
// Called by PersistentPreRunE before any subcommand executes.
func loadConfig() error {
	var cfg *config.Config
	var err error

	// Load config from custom path or default location
	if configFile != "" {
		cfg, err = config.LoadFrom(configFile)
	} else {
		cfg, err = config.Load()
	}

	if err != nil {
		// Config read/parse errors are fatal
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Apply flag overrides to preferences
	if cfg.Preferences == nil {
		cfg.Preferences = config.NewPreferences()
	}

	// --debug flag overrides config preference
	if debugFlag {
		cfg.Preferences.Debug = true
	}

	// --output flag overrides config preference
	if outputFormat != "" {
		switch outputFormat {
		case "table":
			cfg.Preferences.OutputFormat = config.OutputFormatTable
		case "json":
			cfg.Preferences.OutputFormat = config.OutputFormatJSON
		case "yaml":
			cfg.Preferences.OutputFormat = config.OutputFormatYAML
		case "wide":
			cfg.Preferences.OutputFormat = config.OutputFormatWide
		default:
			return fmt.Errorf("invalid output format %q: must be table, json, yaml, or wide", outputFormat)
		}
	}

	// --no-color flag overrides config preference
	if noColor {
		cfg.Preferences.Color = config.ColorModeNever
		SetNoColor(true)
	}

	// --context flag overrides current_context from config
	if contextOverride != "" {
		// Validate that the context exists before overriding
		if _, err := cfg.GetContext(contextOverride); err != nil {
			return fmt.Errorf("context %q not found in configuration", contextOverride)
		}
		cfg.CurrentContext = contextOverride
	}

	loadedConfig = cfg
	return nil
}

// GetConfig returns the loaded configuration.
// Returns nil if called before Execute() or if config loading failed.
// Subcommands should call this in their Run/RunE functions.
func GetConfig() *config.Config {
	return loadedConfig
}

// GetConfigOrFail returns the loaded configuration or exits with an error.
// This is a convenience function for commands that require a valid config.
func GetConfigOrFail() *config.Config {
	cfg := GetConfig()
	if cfg == nil {
		fmt.Fprintln(os.Stderr, "Error: configuration not loaded")
		os.Exit(1)
	}
	return cfg
}

// GetNoInput returns true if interactive prompts should be disabled.
// Commands should check this before prompting for user input.
func GetNoInput() bool {
	return noInput
}

// GetDryRun returns true if commands should show what would be done without executing.
// Commands that modify state should check this and print their intended actions instead.
func GetDryRun() bool {
	return dryRun
}

// Execute runs the root command and returns any error.
// This is called by main.main() and handles command execution.
// Deprecated: Use ExecuteWithExitCode() for proper exit code handling.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

// ExecuteWithExitCode runs the root command and returns an appropriate exit code.
// This maps errors to exit codes for proper CLI behavior:
//   - 0: Success
//   - 1: General error
//   - 2: Command misuse (invalid arguments)
//   - 3: Authentication required
//   - 4: Permission denied
//   - 5: Resource not found
//   - 6: Rate limited
//   - 7: Server error
//   - 8: Network error
//   - 9: Timeout
//   - 10: Plan limit exceeded
//
// Example usage in main.go:
//
//	func main() {
//	    os.Exit(cmd.ExecuteWithExitCode())
//	}
func ExecuteWithExitCode() int {
	err := rootCmd.Execute()
	return clierrors.HandleError(err)
}

// RootCmd returns the root command for adding subcommands.
// This is used by subcommand packages to register themselves.
func RootCmd() *cobra.Command {
	return rootCmd
}
