// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"os"
	"strings"

	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// EnvVarRow represents a row in the environment variable table output.
type EnvVarRow struct {
	Variable    string `json:"variable" yaml:"variable" table:"VARIABLE"`
	Value       string `json:"value" yaml:"value" table:"VALUE"`
	Source      string `json:"source" yaml:"source" table:"SOURCE"`
	Description string `json:"description" yaml:"description" table:"DESCRIPTION,wide"`
}

// EnvVarDefinition defines a supported environment variable.
type EnvVarDefinition struct {
	Name        string
	Description string
	Sensitive   bool
}

// SupportedEnvVars returns all environment variables that affect CLI behavior.
func SupportedEnvVars() []EnvVarDefinition {
	return []EnvVarDefinition{
		{Name: "STACKEYE_API_KEY", Description: "API key for authentication", Sensitive: true},
		{Name: "STACKEYE_API_URL", Description: "API server URL override"},
		{Name: "STACKEYE_CONFIG", Description: "Custom config file path"},
		{Name: "STACKEYE_CONTEXT", Description: "Override current context name"},
		{Name: "STACKEYE_DEBUG", Description: "Enable debug output"},
		{Name: "STACKEYE_TIMEOUT", Description: "HTTP request timeout in seconds"},
		{Name: "STACKEYE_NO_INPUT", Description: "Disable interactive prompts"},
		{Name: "STACKEYE_TELEMETRY", Description: "Override telemetry (0=off, 1=on)"},
		{Name: "NO_COLOR", Description: "Disable colored output (no-color.org)"},
		{Name: "XDG_CONFIG_HOME", Description: "Config directory base path"},
		{Name: "TERM", Description: "Terminal type (dumb disables colors)"},
	}
}

// CollectEnvVars builds table rows for all supported environment variables.
func CollectEnvVars() []EnvVarRow {
	defs := SupportedEnvVars()
	rows := make([]EnvVarRow, 0, len(defs))

	for _, def := range defs {
		val, ok := os.LookupEnv(def.Name)

		var displayVal, source string
		if !ok {
			displayVal = "(not set)"
			source = "-"
		} else if val == "" {
			displayVal = "(empty)"
			source = "env"
		} else if def.Sensitive {
			displayVal = MaskSensitive(val)
			source = "env"
		} else {
			displayVal = val
			source = "env"
		}

		rows = append(rows, EnvVarRow{
			Variable:    def.Name,
			Value:       displayVal,
			Source:      source,
			Description: def.Description,
		})
	}

	return rows
}

// MaskSensitive masks a sensitive value, showing only the prefix and last 4 chars.
// For API keys with format "se_<hex>", shows "se_****...xxxx".
// For short values (<=8 chars), shows "****".
func MaskSensitive(val string) string {
	if len(val) <= 8 {
		return "****"
	}

	if strings.HasPrefix(val, "se_") && len(val) > 7 {
		return "se_****..." + val[len(val)-4:]
	}

	return val[:3] + "****..." + val[len(val)-4:]
}

// PrintEnvVars formats and prints environment variables using the CLI's
// configured output format.
func PrintEnvVars(rows []EnvVarRow) error {
	return getPrinter().Print(rows)
}

// PrintEnvVarsWithFormat formats and prints environment variables using the
// specified output format string. This is used by commands that skip config
// loading and need to respect the --output flag directly.
func PrintEnvVarsWithFormat(rows []EnvVarRow, format string) error {
	opts := sdkoutput.DefaultOptions()

	switch format {
	case "json":
		opts.Format = sdkoutput.FormatJSON
	case "yaml":
		opts.Format = sdkoutput.FormatYAML
	case "wide":
		opts.Format = sdkoutput.FormatWide
	default:
		opts.Format = sdkoutput.FormatTable
	}

	printer := NewPrinterWithOptions(opts)
	return printer.Print(rows)
}
