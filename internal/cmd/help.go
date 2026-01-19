// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"unicode"

	"github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/spf13/cobra"
)

// CommandGroup represents a logical grouping of commands for help output.
type CommandGroup struct {
	Name     string
	Commands []*cobra.Command
}

// commandGroupDefs maps command names to their group names.
// Commands not listed here will appear in "Other Commands".
var commandGroupDefs = map[string]string{
	"login":      "Authentication",
	"logout":     "Authentication",
	"whoami":     "Authentication",
	"config":     "Configuration",
	"context":    "Configuration",
	"probe":      "Monitoring",
	"alert":      "Monitoring",
	"channel":    "Management",
	"team":       "Management",
	"org":        "Management",
	"status":     "Management",
	"billing":    "Management",
	"version":    "Utilities",
	"completion": "Utilities",
	"help":       "Utilities",
}

// groupOrder defines the display order of command groups.
var groupOrder = []string{
	"Authentication",
	"Configuration",
	"Monitoring",
	"Management",
	"Utilities",
	"Other Commands",
}

// HelpConfig holds configuration for the custom help system.
type HelpConfig struct {
	// ColorManager provides colored output formatting.
	ColorManager *output.ColorManager
	// Writer is the output destination (defaults to os.Stdout).
	Writer io.Writer
}

// helpConfig is the global help configuration.
var helpConfig *HelpConfig

// InitHelp initializes the custom help system for the CLI.
// This should be called after creating the root command.
//
// Example:
//
//	rootCmd := NewRootCmd()
//	InitHelp(rootCmd, &HelpConfig{
//	    ColorManager: output.NewColorManager(output.ColorAuto),
//	})
func InitHelp(cmd *cobra.Command, cfg *HelpConfig) {
	if cfg == nil {
		cfg = &HelpConfig{}
	}
	if cfg.ColorManager == nil {
		cfg.ColorManager = output.NewColorManager(output.ColorAuto)
	}
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}
	helpConfig = cfg

	// Register template functions with Cobra BEFORE setting templates.
	// This is critical - Cobra's internal template execution needs these functions
	// registered globally, not just passed to template.Funcs().
	registerTemplateFuncs()

	// Set custom help and usage templates
	cmd.SetHelpTemplate(helpTemplate)
	cmd.SetUsageTemplate(usageTemplate)

	// Apply to all subcommands recursively
	applyHelpToSubcommands(cmd)
}

// applyHelpToSubcommands applies the custom help template to all subcommands.
func applyHelpToSubcommands(cmd *cobra.Command) {
	for _, subcmd := range cmd.Commands() {
		subcmd.SetHelpTemplate(helpTemplate)
		subcmd.SetUsageTemplate(usageTemplate)
		applyHelpToSubcommands(subcmd)
	}
}

// GetHelpConfig returns the current help configuration.
// Returns a default config if InitHelp hasn't been called.
func GetHelpConfig() *HelpConfig {
	if helpConfig == nil {
		return &HelpConfig{
			ColorManager: output.NewColorManager(output.ColorAuto),
			Writer:       os.Stdout,
		}
	}
	return helpConfig
}

// SetNoColor updates the help system to disable colors.
// This is called when --no-color flag is set.
func SetNoColor(disabled bool) {
	if helpConfig == nil {
		helpConfig = &HelpConfig{
			Writer: os.Stdout,
		}
	}
	if disabled {
		helpConfig.ColorManager = output.NewColorManager(output.ColorNever)
	} else {
		helpConfig.ColorManager = output.NewColorManager(output.ColorAuto)
	}
}

// GroupCommands organizes commands into logical groups for display.
func GroupCommands(commands []*cobra.Command) []CommandGroup {
	// Build a map of group name to commands
	groups := make(map[string][]*cobra.Command)

	for _, cmd := range commands {
		if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
			continue
		}

		groupName, ok := commandGroupDefs[cmd.Name()]
		if !ok {
			groupName = "Other Commands"
		}
		groups[groupName] = append(groups[groupName], cmd)
	}

	// Build ordered result
	var result []CommandGroup
	for _, groupName := range groupOrder {
		if cmds, ok := groups[groupName]; ok && len(cmds) > 0 {
			result = append(result, CommandGroup{
				Name:     groupName,
				Commands: cmds,
			})
		}
	}

	return result
}

// registerTemplateFuncs registers all custom template functions with Cobra.
// This must be called BEFORE SetHelpTemplate/SetUsageTemplate.
// Cobra uses its own global template function registry, so we must register
// functions via cobra.AddTemplateFunc() for them to be available during
// help rendering.
func registerTemplateFuncs() {
	cfg := GetHelpConfig()

	// Color functions - these use closures to capture the current config
	cobra.AddTemplateFunc("bold", func(s string) string {
		return cfg.ColorManager.Bold(s)
	})
	cobra.AddTemplateFunc("dim", func(s string) string {
		return cfg.ColorManager.Dim(s)
	})
	cobra.AddTemplateFunc("cyan", func(s string) string {
		return cfg.ColorManager.Info(s)
	})
	cobra.AddTemplateFunc("green", func(s string) string {
		return cfg.ColorManager.Success(s)
	})
	cobra.AddTemplateFunc("yellow", func(s string) string {
		return cfg.ColorManager.Warning(s)
	})

	// String helpers
	cobra.AddTemplateFunc("trimTrailingWhitespaces", func(s string) string {
		return strings.TrimRightFunc(s, unicode.IsSpace)
	})
	cobra.AddTemplateFunc("rpad", func(s string, padding int) string {
		f := fmt.Sprintf("%%-%ds", padding)
		return fmt.Sprintf(f, s)
	})

	// Command grouping
	cobra.AddTemplateFunc("groupCommands", GroupCommands)

	// Command helpers
	cobra.AddTemplateFunc("hasSubCommands", func(cmd *cobra.Command) bool {
		return cmd.HasAvailableSubCommands()
	})
	cobra.AddTemplateFunc("hasFlags", func(cmd *cobra.Command) bool {
		return cmd.Flags().HasAvailableFlags()
	})
	cobra.AddTemplateFunc("hasLocalFlags", func(cmd *cobra.Command) bool {
		return cmd.LocalFlags().HasAvailableFlags()
	})
	cobra.AddTemplateFunc("hasInheritedFlags", func(cmd *cobra.Command) bool {
		return cmd.InheritedFlags().HasAvailableFlags()
	})
	cobra.AddTemplateFunc("hasExample", func(cmd *cobra.Command) bool {
		return len(cmd.Example) > 0
	})
	cobra.AddTemplateFunc("hasAliases", func(cmd *cobra.Command) bool {
		return len(cmd.Aliases) > 0
	})
}

// templateFuncs returns template functions for help rendering.
// This is used by RenderHelp() for manual template execution in tests.
func templateFuncs() template.FuncMap {
	cfg := GetHelpConfig()
	return template.FuncMap{
		// Color functions
		"bold": func(s string) string {
			return cfg.ColorManager.Bold(s)
		},
		"dim": func(s string) string {
			return cfg.ColorManager.Dim(s)
		},
		"cyan": func(s string) string {
			return cfg.ColorManager.Info(s)
		},
		"green": func(s string) string {
			return cfg.ColorManager.Success(s)
		},
		"yellow": func(s string) string {
			return cfg.ColorManager.Warning(s)
		},

		// String helpers
		"trimTrailingWhitespaces": func(s string) string {
			return strings.TrimRightFunc(s, unicode.IsSpace)
		},
		"rpad": func(s string, padding int) string {
			f := fmt.Sprintf("%%-%ds", padding)
			return fmt.Sprintf(f, s)
		},

		// Command grouping
		"groupCommands": GroupCommands,

		// Command helpers
		"hasSubCommands": func(cmd *cobra.Command) bool {
			return cmd.HasAvailableSubCommands()
		},
		"hasFlags": func(cmd *cobra.Command) bool {
			return cmd.Flags().HasAvailableFlags()
		},
		"hasLocalFlags": func(cmd *cobra.Command) bool {
			return cmd.LocalFlags().HasAvailableFlags()
		},
		"hasInheritedFlags": func(cmd *cobra.Command) bool {
			return cmd.InheritedFlags().HasAvailableFlags()
		},
		"hasExample": func(cmd *cobra.Command) bool {
			return len(cmd.Example) > 0
		},
		"hasAliases": func(cmd *cobra.Command) bool {
			return len(cmd.Aliases) > 0
		},
	}
}

// RenderHelp renders the help output for a command.
// This can be used to test help rendering.
func RenderHelp(cmd *cobra.Command) (string, error) {
	t, err := template.New("help").Funcs(templateFuncs()).Parse(helpTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse help template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, cmd); err != nil {
		return "", fmt.Errorf("failed to execute help template: %w", err)
	}

	return buf.String(), nil
}

// helpTemplate is the custom help template for all commands.
// It provides colored output, grouped commands, and prominent examples.
const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{bold "USAGE"}}
  {{cyan .UseLine}}{{if .HasAvailableSubCommands}} [command]{{end}}{{if .HasAvailableFlags}} [flags]{{end}}
{{end}}{{if hasAliases .}}
{{bold "ALIASES"}}
  {{.NameAndAliases}}
{{end}}{{if hasExample .}}
{{bold "EXAMPLES"}}
{{.Example}}
{{end}}{{if .HasAvailableSubCommands}}
{{bold "COMMANDS"}}
{{range $group := groupCommands .Commands}}
  {{bold $group.Name}}:
{{range $group.Commands}}    {{cyan (rpad .Name .NamePadding)}} {{.Short}}
{{end}}{{end}}{{end}}{{if hasLocalFlags .}}
{{bold "FLAGS"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if hasInheritedFlags .}}
{{bold "GLOBAL FLAGS"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .HasHelpSubCommands}}
{{bold "ADDITIONAL HELP TOPICS"}}
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}  {{cyan (rpad .CommandPath .CommandPathPadding)}} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.
{{end}}`

// usageTemplate is the custom usage template (shown on errors).
const usageTemplate = `{{bold "Usage"}}:{{if .Runnable}}
  {{cyan .UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{cyan .CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{bold "Aliases"}}:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{bold "Examples"}}:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{bold "Available Commands"}}:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{cyan (rpad .Name .NamePadding)}} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{bold .Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{cyan (rpad .Name .NamePadding)}} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{bold "Additional Commands"}}:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{cyan (rpad .Name .NamePadding)}} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{bold "Flags"}}:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{bold "Global Flags"}}:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{bold "Additional help topics"}}:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{cyan (rpad .CommandPath .CommandPathPadding)}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
