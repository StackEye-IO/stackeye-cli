// Package cmd implements the CLI commands for StackEye.
//
// This package contains the root command and all subcommands for the
// StackEye CLI tool. Commands are organized hierarchically using Cobra.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/debug"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	cliinteractive "github.com/StackEye-IO/stackeye-cli/internal/interactive"
	clioutput "github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-cli/internal/telemetry"
	cliupdate "github.com/StackEye-IO/stackeye-cli/internal/update"
	"github.com/StackEye-IO/stackeye-cli/internal/version"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/spf13/cobra"
)

// Global flag variables
var (
	configFile      string
	contextOverride string
	debugFlag       bool
	verbosity       int // kubectl-style verbosity level (0-10)
	outputFormat    string
	noColor         bool
	noInput         bool
	dryRun          bool
	timeoutSeconds  int  // HTTP request timeout in seconds (0 = use config/default)
	noUpdateCheck   bool // Disable automatic update checking
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

	// Wire up the API client helper to use our verbosity getter
	api.SetVerbosityGetter(GetVerbosity)

	// Wire up the API client helper to use our timeout getter
	api.SetTimeoutGetter(GetTimeout)

	// Wire up the output package to use our config getter
	clioutput.SetConfigGetter(GetConfig)

	// Wire up the output package to check --no-input for spinner suppression
	clioutput.SetNoInputGetter(GetNoInput)

	// Wire up the interactive package to check --no-input for prompt bypass
	cliinteractive.SetNoInputGetter(GetNoInput)

	// Wire up the debug logger to use our verbosity getter
	debug.SetVerbosityGetter(GetVerbosity)

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
	rootCmd.AddCommand(NewChannelCmd())
	rootCmd.AddCommand(NewOrgCmd())
	rootCmd.AddCommand(NewDashboardCmd())
	rootCmd.AddCommand(NewRegionCmd())
	rootCmd.AddCommand(NewAPIKeyCmd())
	rootCmd.AddCommand(NewMuteCmd())
	rootCmd.AddCommand(NewMaintenanceCmd())
	rootCmd.AddCommand(NewStatusPageCmd())
	rootCmd.AddCommand(NewIncidentCmd())
	rootCmd.AddCommand(NewTeamCmd())
	rootCmd.AddCommand(NewBillingCmd())
	rootCmd.AddCommand(NewAdminCmd())
	rootCmd.AddCommand(NewSetupCmd())
	rootCmd.AddCommand(NewLabelCmd())
	rootCmd.AddCommand(NewTelemetryCmd())
	rootCmd.AddCommand(NewEnvCmd())
	rootCmd.AddCommand(NewUpgradeCmd())

	// Register persistent flags available to all commands
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path (default: ~/.config/stackeye/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&contextOverride, "context", "", "override current context from config")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debug output (shorthand for --v=6)")
	rootCmd.PersistentFlags().IntVarP(&verbosity, "v", "v", 0, "verbosity level (0-10): 5=requests, 6=responses, 7+=headers, 9+=bodies")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format: table, json, yaml, wide")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&noInput, "no-input", false, "disable interactive prompts")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without executing")
	rootCmd.PersistentFlags().IntVar(&timeoutSeconds, "timeout", 0, "HTTP request timeout in seconds (default: 30, or config preference)")
	rootCmd.PersistentFlags().BoolVar(&noUpdateCheck, "no-update-check", false, "disable automatic update checking")

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

	// STACKEYE_DEBUG env var enables debug mode (same as --debug flag)
	if os.Getenv("STACKEYE_DEBUG") != "" && !debugFlag {
		debugFlag = true
	}

	// --debug flag is shorthand for --v=6
	// Only apply if --v wasn't explicitly set
	if debugFlag && verbosity == 0 {
		verbosity = 6
	}

	// --debug flag overrides config preference
	if debugFlag || verbosity > 0 {
		cfg.Preferences.Debug = true
	}

	// Log after verbosity is set so the message is actually emitted
	if os.Getenv("STACKEYE_DEBUG") != "" {
		debug.Log(3, "debug enabled via STACKEYE_DEBUG env var")
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
			return clierrors.InvalidValueError("--output", outputFormat, clierrors.ValidOutputFormats)
		}
	}

	// NO_COLOR environment variable disables colors (per https://no-color.org/)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		cfg.Preferences.Color = config.ColorModeNever
		SetNoColor(true)
	}

	// --no-color flag overrides config preference
	if noColor {
		cfg.Preferences.Color = config.ColorModeNever
		SetNoColor(true)
	}

	// STACKEYE_TIMEOUT env var sets timeout (same as --timeout flag)
	if envTimeout := os.Getenv("STACKEYE_TIMEOUT"); envTimeout != "" && timeoutSeconds == 0 {
		if v, err := strconv.Atoi(envTimeout); err == nil && v > 0 {
			timeoutSeconds = v
		}
	}

	// --timeout flag overrides config preference
	if timeoutSeconds > 0 {
		cfg.Preferences.DefaultTimeout = timeoutSeconds
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
	debug.ConfigLoaded(configFile, cfg.CurrentContext)
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
// Deprecated: Use ExecuteWithContext for signal-aware execution.
func ExecuteWithExitCode() int {
	return ExecuteWithContext(context.Background())
}

// ExecuteWithContext runs the root command with the given context and returns
// an appropriate exit code. The context is passed to Cobra so that all
// subcommands can observe cancellation (e.g. from SIGINT/SIGTERM).
func ExecuteWithContext(ctx context.Context) int {
	startTime := time.Now()

	// Start background update check early (before command execution)
	// so it has time to complete while the command runs.
	var notifier *cliupdate.Notifier
	if shouldCheckForUpdates() {
		notifier = cliupdate.NewNotifier(version.Version, cliupdate.WithColor(!noColor))
		notifier.StartCheck(ctx)
	}

	rootCmd.SetContext(ctx)
	err := rootCmd.Execute()
	exitCode := clierrors.HandleError(err)

	// Track telemetry (async, non-blocking).
	// Flush is handled by the signal handler's OnCleanup chain in main.go,
	// so it runs on both normal exit and signal-driven shutdown.
	duration := time.Since(startTime)
	commandName := getExecutedCommandName()
	telemetry.GetClient().Track(context.Background(), commandName, exitCode, duration)

	// Print update notification after command completes (non-blocking wait).
	// Only print on successful execution to avoid cluttering error output.
	if notifier != nil && exitCode == 0 {
		notifier.PrintNotification()
	}

	return exitCode
}

// getExecutedCommandName extracts the command path from the executed command.
// Returns "stackeye" for the root command, or "stackeye <subcommand>" for subcommands.
func getExecutedCommandName() string {
	// Get the args that were passed (excluding flags)
	args := os.Args[1:]
	var commandParts []string

	for _, arg := range args {
		// Stop at first flag
		if strings.HasPrefix(arg, "-") {
			break
		}
		commandParts = append(commandParts, arg)
	}

	if len(commandParts) == 0 {
		return "stackeye"
	}
	return strings.Join(commandParts, " ")
}

// RootCmd returns the root command for adding subcommands.
// This is used by subcommand packages to register themselves.
func RootCmd() *cobra.Command {
	return rootCmd
}

// GetVerbosity returns the kubectl-style verbosity level (0-10).
// Level meanings:
//   - 0: Errors only (default)
//   - 1: Warnings + errors
//   - 2: Info messages
//   - 3: Extended info (config/context details)
//   - 4: Debug messages (internal flow)
//   - 5: HTTP requests (method, URL, duration)
//   - 6: HTTP responses (+ status code, body size)
//   - 7: Request headers (redacted)
//   - 8: Response headers
//   - 9: Full bodies (truncated at 10KB)
//   - 10: Trace (curl equivalent, timing breakdown)
func GetVerbosity() int {
	return verbosity
}

// GetTimeout returns the effective HTTP request timeout in seconds.
// Returns the --timeout flag value, STACKEYE_TIMEOUT env var, config preference,
// or 0 if none is set (SDK uses its own default of 30s).
func GetTimeout() int {
	return timeoutSeconds
}

// shouldCheckForUpdates returns true if update checking should be performed.
// Returns false if:
// - --no-update-check flag is set
// - STACKEYE_NO_UPDATE_CHECK environment variable is set
// - Running a dev build (version == "dev")
func shouldCheckForUpdates() bool {
	// Check environment variable
	if os.Getenv("STACKEYE_NO_UPDATE_CHECK") != "" {
		return false
	}

	// Use the helper function that checks flag and dev build
	return cliupdate.ShouldCheck(version.Version, noUpdateCheck)
}
