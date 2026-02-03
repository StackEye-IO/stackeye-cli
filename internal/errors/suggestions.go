// Package errors provides CLI error handling utilities.
package errors

import (
	"fmt"
	"sort"
	"strings"
)

// Suggestion represents a suggested correction for a misspelled value.
type Suggestion struct {
	Value    string
	Distance int
}

// SuggestFromOptions returns the closest match from valid options for a given input.
// Returns an empty string if no close match is found (distance > maxDistance) or if
// the input exactly matches one of the options (case-insensitive).
// The maxDistance parameter controls how different the input can be from a valid option.
func SuggestFromOptions(input string, validOptions []string, maxDistance int) string {
	if maxDistance <= 0 {
		maxDistance = 2 // Default: allow up to 2 edits
	}

	inputLower := strings.ToLower(input)

	// Check for exact match (case-insensitive) - no suggestion needed
	for _, opt := range validOptions {
		if inputLower == strings.ToLower(opt) {
			return ""
		}
	}

	var suggestions []Suggestion

	for _, opt := range validOptions {
		dist := levenshteinDistance(inputLower, strings.ToLower(opt))
		if dist <= maxDistance && dist > 0 {
			suggestions = append(suggestions, Suggestion{Value: opt, Distance: dist})
		}
	}

	if len(suggestions) == 0 {
		return ""
	}

	// Sort by distance (closest first), then alphabetically for ties
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Distance != suggestions[j].Distance {
			return suggestions[i].Distance < suggestions[j].Distance
		}
		return suggestions[i].Value < suggestions[j].Value
	})

	return suggestions[0].Value
}

// levenshteinDistance calculates the edit distance between two strings.
// This is the minimum number of single-character edits (insertions, deletions,
// or substitutions) required to change one string into the other.
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a 2D matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}

	// Initialize first row
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// InvalidValueError formats an error message for an invalid value with an optional suggestion.
// This provides a consistent format across all CLI commands.
//
// Example output:
//
//	Invalid value "htpp" for --check-type: must be one of: http, ping, tcp, dns_resolve
//	  Did you mean "http"?
func InvalidValueError(flagName, value string, validOptions []string) error {
	optList := strings.Join(validOptions, ", ")
	suggestion := SuggestFromOptions(value, validOptions, 2)

	if suggestion != "" {
		return fmt.Errorf("invalid value %q for %s: must be one of: %s\n  Did you mean %q?",
			value, flagName, optList, suggestion)
	}
	return fmt.Errorf("invalid value %q for %s: must be one of: %s", value, flagName, optList)
}

// InvalidValueWithHintError formats an error for an invalid value with a custom hint.
// Use this when the valid options are too numerous to list or when a custom message is better.
//
// Example output:
//
//	Invalid value "abc" for --interval: must be a number between 30 and 3600
func InvalidValueWithHintError(flagName, value, hint string) error {
	return fmt.Errorf("invalid value %q for %s: %s", value, flagName, hint)
}

// RequiredFlagError returns a consistent error for missing required flags.
// The format matches Cobra's built-in required flag errors for consistency.
func RequiredFlagError(flagName string) error {
	return fmt.Errorf("required flag %q not set", flagName)
}

// RequiredArgError returns a consistent error for missing required arguments.
func RequiredArgError(argName string) error {
	return fmt.Errorf("required argument %q not provided", argName)
}

// Common valid options for suggestions

// ValidCheckTypes contains the valid probe check types.
var ValidCheckTypes = []string{"http", "ping", "tcp", "dns_resolve"}

// ValidHTTPMethods contains the valid HTTP methods.
var ValidHTTPMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

// ValidOutputFormats contains the valid output formats.
var ValidOutputFormats = []string{"table", "json", "yaml", "wide"}

// ValidKeywordCheckTypes contains the valid keyword check types.
var ValidKeywordCheckTypes = []string{"contains", "not_contains"}

// ValidAlertStatuses contains the valid alert statuses.
var ValidAlertStatuses = []string{"active", "acknowledged", "resolved"}

// ValidProbeStatuses contains the valid probe result statuses.
var ValidProbeStatuses = []string{"success", "failure"}

// ValidPeriods contains the valid time periods for stats.
var ValidPeriods = []string{"24h", "7d", "30d"}

// ValidThemes contains the valid status page themes.
var ValidThemes = []string{"light", "dark", "system"}

// ValidMuteScopes contains the valid mute scopes.
var ValidMuteScopes = []string{"organization", "probe", "channel", "alert_type"}

// ValidMuteAlertTypes contains the valid alert types for mutes.
var ValidMuteAlertTypes = []string{"status_down", "ssl_expiry", "ssl_invalid", "slow_response"}

// ValidChannelTypes contains the valid notification channel types.
var ValidChannelTypes = []string{"email", "slack", "webhook", "pagerduty", "discord", "teams", "sms"}

// ValidIncidentStatuses contains the valid incident statuses.
var ValidIncidentStatuses = []string{"investigating", "identified", "monitoring", "resolved"}

// ValidIncidentImpacts contains the valid incident impact levels.
var ValidIncidentImpacts = []string{"none", "minor", "major", "critical"}

// ValidTeamRoles contains the valid team member roles.
var ValidTeamRoles = []string{"owner", "admin", "member", "viewer"}

// ValidProbeStatusFilters contains the valid probe status filter values.
var ValidProbeStatusFilters = []string{"up", "down", "degraded", "paused", "pending"}

// ValidSeverities contains the valid alert severity levels.
var ValidSeverities = []string{"critical", "warning", "info"}

// ValidDependencyDirections contains the valid probe dependency clear directions.
var ValidDependencyDirections = []string{"parents", "children", "both"}

// ValidBoolStrings contains valid boolean string values.
var ValidBoolStrings = []string{"true", "false"}

// ValidPagerDutySeverities contains valid PagerDuty severity levels.
var ValidPagerDutySeverities = []string{"critical", "error", "warning", "info"}

// ValidWebhookMethods contains valid HTTP methods for webhooks.
var ValidWebhookMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

// ValidExportFormats contains valid export output formats.
var ValidExportFormats = []string{"yaml", "json"}
