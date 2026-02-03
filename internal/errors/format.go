// Package errors provides CLI error handling and formatting utilities.
package errors

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/StackEye-IO/stackeye-cli/internal/output"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// ErrorFormatter provides color-coded error message formatting for CLI output.
// It uses the SDK's ColorManager to handle color mode preferences (auto/always/never)
// and respects NO_COLOR, TERM=dumb, and piped output detection.
//
// Usage:
//
//	f := NewErrorFormatter()
//	f.PrintError("probe not found")
//	f.PrintContext("Probe ID", "abc123")
//	f.PrintSuggestion("Run 'stackeye probe list' to see available probes")
//	f.PrintRequestID("req_xyz789")
type ErrorFormatter struct {
	cm     *sdkoutput.ColorManager
	writer io.Writer
}

// NewErrorFormatter creates an ErrorFormatter using the CLI's configured
// color mode and writing to stderr.
func NewErrorFormatter() *ErrorFormatter {
	return &ErrorFormatter{
		cm:     output.NewColorManager(),
		writer: os.Stderr,
	}
}

// NewErrorFormatterWithWriter creates an ErrorFormatter with a custom writer.
// This is useful for testing or redirecting error output.
func NewErrorFormatterWithWriter(w io.Writer) *ErrorFormatter {
	return &ErrorFormatter{
		cm:     output.NewColorManager(),
		writer: w,
	}
}

// NewErrorFormatterWithColorManager creates an ErrorFormatter with a custom
// ColorManager and writer. This is useful for testing with specific color modes.
func NewErrorFormatterWithColorManager(cm *sdkoutput.ColorManager, w io.Writer) *ErrorFormatter {
	return &ErrorFormatter{
		cm:     cm,
		writer: w,
	}
}

// PrintError prints an error message with red "Error:" prefix.
// Format: "Error: <message>\n"
func (f *ErrorFormatter) PrintError(msg string) {
	prefix := f.cm.Error("Error:")
	fmt.Fprintf(f.writer, "%s %s\n", prefix, msg)
}

// PrintWarning prints a warning message with yellow "Warning:" prefix.
// Format: "Warning: <message>\n"
func (f *ErrorFormatter) PrintWarning(msg string) {
	prefix := f.cm.Warning("Warning:")
	fmt.Fprintf(f.writer, "%s %s\n", prefix, msg)
}

// PrintHint prints a hint message indented under the error.
// Format: "  <hint>\n"
func (f *ErrorFormatter) PrintHint(hint string) {
	fmt.Fprintf(f.writer, "  %s\n", hint)
}

// PrintContext prints a key-value context pair indented under the error.
// Format: "  <key>: <value>\n"
func (f *ErrorFormatter) PrintContext(key, value string) {
	dimKey := f.cm.Dim(key + ":")
	fmt.Fprintf(f.writer, "  %s %s\n", dimKey, value)
}

// PrintSuggestion prints a suggestion with a dimmed "Suggestion:" prefix.
// Format: "  Suggestion: <text>\n"
func (f *ErrorFormatter) PrintSuggestion(text string) {
	prefix := f.cm.Dim("Suggestion:")
	fmt.Fprintf(f.writer, "  %s %s\n", prefix, text)
}

// PrintRequestID prints a request ID for support ticket reference.
// Format: "  Request ID: <id> (include this when contacting support)\n"
func (f *ErrorFormatter) PrintRequestID(requestID string) {
	if requestID == "" {
		return
	}
	label := f.cm.Dim("Request ID:")
	hint := f.cm.Dim("(include this when contacting support)")
	fmt.Fprintf(f.writer, "  %s %s %s\n", label, requestID, hint)
}

// PrintValidationErrors prints a list of field validation errors.
// Format:
//
//	<field>: <message>
//	<field>: <message>
func (f *ErrorFormatter) PrintValidationErrors(fields map[string]string) {
	for field, msg := range fields {
		fmt.Fprintf(f.writer, "  %s: %s\n", field, msg)
	}
}

// FormatError returns a formatted error string with red "Error:" prefix.
// Use PrintError for direct output; use FormatError when building strings.
func (f *ErrorFormatter) FormatError(msg string) string {
	prefix := f.cm.Error("Error:")
	return fmt.Sprintf("%s %s", prefix, msg)
}

// FormatWarning returns a formatted warning string with yellow "Warning:" prefix.
func (f *ErrorFormatter) FormatWarning(msg string) string {
	prefix := f.cm.Warning("Warning:")
	return fmt.Sprintf("%s %s", prefix, msg)
}

// FormatSuccess returns a formatted success string with green "Success:" prefix.
func (f *ErrorFormatter) FormatSuccess(msg string) string {
	prefix := f.cm.Success("Success:")
	return fmt.Sprintf("%s %s", prefix, msg)
}

// FormatInfo returns a formatted info string with cyan "Info:" prefix.
func (f *ErrorFormatter) FormatInfo(msg string) string {
	prefix := f.cm.Info("Info:")
	return fmt.Sprintf("%s %s", prefix, msg)
}

// CommonSuggestions provides standard suggestions for common error scenarios.
// These are used to provide helpful hints to users when errors occur.
var CommonSuggestions = map[string]string{
	// Authentication
	"auth_required":    "Run 'stackeye login' to authenticate.",
	"session_expired":  "Your session has expired. Run 'stackeye login' to re-authenticate.",
	"invalid_api_key":  "Your API key is invalid. Run 'stackeye login' to generate a new one.",
	"token_expired":    "Your token has expired. Run 'stackeye login' to refresh.",
	"no_org_context":   "Run 'stackeye org switch' to select an organization.",
	"no_api_key":       "Run 'stackeye login' or set STACKEYE_API_KEY environment variable.",
	"invalid_token":    "Your authentication token is invalid. Run 'stackeye login' to get a new token.",
	"mfa_required":     "Multi-factor authentication is required. Complete MFA in your browser.",
	"account_locked":   "Your account has been locked. Contact support for assistance.",
	"invalid_password": "The password is incorrect. Try again or reset your password.",

	// Authorization
	"forbidden":          "You may not have access to this resource or organization.",
	"permission_denied":  "You don't have permission to perform this action.",
	"insufficient_scope": "Your API key doesn't have the required scope for this action.",
	"read_only":          "This resource is read-only and cannot be modified.",
	"owner_only":         "Only the organization owner can perform this action.",
	"admin_only":         "This action requires admin privileges.",

	// Resource errors
	"probe_not_found":        "Run 'stackeye probe list' to see available probes.",
	"alert_not_found":        "Run 'stackeye alert list' to see available alerts.",
	"channel_not_found":      "Run 'stackeye channel list' to see available channels.",
	"org_not_found":          "Run 'stackeye org list' to see your organizations.",
	"team_not_found":         "Run 'stackeye team list' to see available teams.",
	"status_page_not_found":  "Run 'stackeye status-page list' to see available status pages.",
	"incident_not_found":     "Run 'stackeye incident list' to see available incidents.",
	"maintenance_not_found":  "Run 'stackeye maintenance list' to see scheduled maintenance windows.",
	"api_key_not_found":      "Run 'stackeye apikey list' to see your API keys.",
	"invitation_not_found":   "Run 'stackeye team invitations' to see pending invitations.",
	"mute_not_found":         "Run 'stackeye mute list' to see active mutes.",
	"label_not_found":        "Run 'stackeye label list' to see available labels.",
	"region_not_found":       "Run 'stackeye region list' to see available monitoring regions.",
	"resource_not_found":     "The requested resource does not exist or has been deleted.",
	"already_exists":         "A resource with this identifier already exists.",
	"conflict":               "The resource was modified by another request. Refresh and try again.",
	"gone":                   "This resource has been permanently deleted.",
	"duplicate":              "A duplicate resource already exists. Use a different identifier.",
	"stale_data":             "The data is outdated. Refresh and try again.",
	"version_conflict":       "Resource version mismatch. Refresh and try again.",
	"dependency_exists":      "Cannot delete because other resources depend on this one.",
	"circular_dependency":    "This would create a circular dependency.",
	"parent_not_found":       "The parent resource does not exist.",
	"child_exists":           "Cannot delete because child resources exist.",
	"referenced_by_other":    "This resource is referenced by other resources.",
	"in_use":                 "This resource is currently in use and cannot be modified.",
	"locked":                 "This resource is locked and cannot be modified.",
	"archived":               "This resource has been archived and cannot be modified.",
	"inactive":               "This resource is inactive. Activate it first.",
	"disabled":               "This resource is disabled. Enable it first.",
	"suspended":              "This resource has been suspended. Contact support.",
	"pending_approval":       "This resource is pending approval.",
	"pending_verification":   "This resource is pending verification.",
	"pending_deletion":       "This resource is pending deletion.",
	"scheduled_for_deletion": "This resource is scheduled for deletion.",

	// Rate limiting
	"rate_limited":  "Please wait a moment and try again.",
	"too_many":      "Too many requests. Wait a few seconds before retrying.",
	"quota_exceed":  "API quota exceeded. Wait until the quota resets.",
	"slow_down":     "Slow down your request rate to avoid rate limiting.",
	"backoff":       "Implement exponential backoff in your automation scripts.",
	"burst_limit":   "Burst limit exceeded. Space out your requests.",
	"daily_limit":   "Daily limit reached. Try again tomorrow.",
	"monthly_limit": "Monthly limit reached. Upgrade your plan for higher limits.",

	// Plan limits
	"plan_limit":          "Upgrade your plan at https://stackeye.io/billing",
	"probe_limit":         "You've reached your plan's probe limit. Upgrade to add more.",
	"team_limit":          "You've reached your plan's team member limit. Upgrade to add more.",
	"channel_limit":       "You've reached your plan's notification channel limit. Upgrade to add more.",
	"status_page_limit":   "You've reached your plan's status page limit. Upgrade to add more.",
	"check_interval":      "Your plan's minimum check interval is limited. Upgrade for faster checks.",
	"retention_limit":     "Your plan's data retention period is limited. Upgrade for longer retention.",
	"region_limit":        "Your plan's monitoring region count is limited. Upgrade for more regions.",
	"integration_limit":   "Your plan's integration count is limited. Upgrade for more integrations.",
	"alert_rule_limit":    "You've reached your plan's alert rule limit. Upgrade to add more.",
	"custom_domain_limit": "Your plan doesn't include custom domains. Upgrade to enable this feature.",
	"api_rate_limit":      "Your plan's API rate limit is restricted. Upgrade for higher limits.",
	"sso_required":        "SSO is not available on your plan. Upgrade to Enterprise.",
	"audit_log_limit":     "Audit logs are not available on your plan. Upgrade to access.",
	"advanced_feature":    "This feature requires a higher plan tier.",
	"enterprise_only":     "This feature is only available on Enterprise plans.",
	"trial_expired":       "Your trial has expired. Subscribe to continue using StackEye.",
	"payment_required":    "Payment is required to continue. Update your billing information.",
	"subscription_issue":  "There's an issue with your subscription. Check your billing settings.",
	"overdue_payment":     "Your payment is overdue. Update your payment method to avoid service interruption.",

	// Server errors
	"server_error":      "If this persists, check https://status.stackeye.io",
	"maintenance":       "StackEye is undergoing maintenance. Check https://status.stackeye.io for updates.",
	"service_down":      "The service is temporarily unavailable. Try again in a few minutes.",
	"database_error":    "A database error occurred. Try again or contact support.",
	"timeout":           "The request timed out. Try again with a smaller request.",
	"upstream_error":    "An upstream service is unavailable. Try again later.",
	"internal_error":    "An internal error occurred. If this persists, contact support.",
	"capacity_exceeded": "Server capacity exceeded. Try again during off-peak hours.",
	"overloaded":        "The service is overloaded. Try again in a few minutes.",
	"circuit_breaker":   "The service is temporarily disabled due to failures. Try again later.",

	// Network errors
	"connection_refused": "The server may be down or unreachable. Check your network connection.",
	"connection_reset":   "Connection was reset. Try again or check https://status.stackeye.io",
	"dns_failure":        "DNS lookup failed. Check your network and DNS settings.",
	"network_error":      "Check your internet connection and try again.",
	"ssl_error":          "SSL/TLS error. Check your system's certificate store.",
	"proxy_error":        "Proxy connection failed. Check your proxy settings.",
	"firewall_blocked":   "Connection may be blocked by a firewall. Check your network settings.",
	"vpn_required":       "This resource may require VPN access.",
	"offline":            "You appear to be offline. Check your internet connection.",
	"unstable_network":   "Network connection is unstable. Move to a more stable connection.",
	"packet_loss":        "High packet loss detected. Check your network quality.",
	"latency_high":       "High latency detected. This may cause timeouts.",

	// Validation errors
	"invalid_url":          "URL must start with http:// or https://",
	"invalid_email":        "Please provide a valid email address.",
	"invalid_uuid":         "Please provide a valid UUID.",
	"invalid_json":         "The provided JSON is malformed.",
	"invalid_yaml":         "The provided YAML is malformed.",
	"invalid_format":       "The input format is invalid.",
	"missing_field":        "Required field is missing.",
	"field_too_long":       "Field value exceeds maximum length.",
	"field_too_short":      "Field value is below minimum length.",
	"invalid_characters":   "Field contains invalid characters.",
	"invalid_range":        "Value is outside the allowed range.",
	"invalid_enum":         "Value must be one of the allowed options.",
	"invalid_pattern":      "Value doesn't match the required pattern.",
	"invalid_date":         "Please provide a valid date.",
	"invalid_time":         "Please provide a valid time.",
	"invalid_timezone":     "Please provide a valid timezone.",
	"invalid_cron":         "Invalid cron expression.",
	"invalid_regex":        "Invalid regular expression.",
	"invalid_ip":           "Please provide a valid IP address.",
	"invalid_port":         "Port must be between 1 and 65535.",
	"invalid_hostname":     "Please provide a valid hostname.",
	"invalid_path":         "Please provide a valid file path.",
	"invalid_interval":     "Please provide a valid time interval.",
	"invalid_threshold":    "Threshold value is invalid.",
	"invalid_percentage":   "Percentage must be between 0 and 100.",
	"invalid_count":        "Count must be a positive integer.",
	"invalid_size":         "Size value is invalid.",
	"invalid_duration":     "Duration value is invalid.",
	"invalid_color":        "Please provide a valid color code.",
	"invalid_currency":     "Please provide a valid currency code.",
	"invalid_country":      "Please provide a valid country code.",
	"invalid_language":     "Please provide a valid language code.",
	"invalid_phone":        "Please provide a valid phone number.",
	"immutable_field":      "This field cannot be changed after creation.",
	"required_together":    "These fields must be provided together.",
	"mutually_exclusive":   "These fields cannot be used together.",
	"conditional_required": "This field is required based on other values.",

	// CLI-specific errors
	"config_not_found":     "Run 'stackeye login' to initialize configuration.",
	"config_parse":         "Config file is corrupted. Run 'stackeye config reset' to fix.",
	"config_write":         "Could not write config file. Check file permissions.",
	"config_read":          "Could not read config file. Check file permissions.",
	"invalid_output":       "Invalid output format. Use 'table', 'json', or 'yaml'.",
	"invalid_flag":         "Invalid flag value. Run 'stackeye <command> --help' for usage.",
	"missing_argument":     "Missing required argument. Run 'stackeye <command> --help' for usage.",
	"too_many_arguments":   "Too many arguments. Run 'stackeye <command> --help' for usage.",
	"invalid_subcommand":   "Unknown subcommand. Run 'stackeye <command> --help' for available subcommands.",
	"shell_not_supported":  "Your shell is not supported for completion. Try bash, zsh, fish, or PowerShell.",
	"interactive_required": "This command requires interactive input. Remove --no-input flag.",
	"stdin_required":       "This command expects input from stdin. Pipe data to the command.",
	"file_not_found":       "The specified file was not found. Check the path and try again.",
	"file_read_error":      "Could not read the specified file. Check file permissions.",
	"file_write_error":     "Could not write to the specified file. Check file permissions.",
	"directory_not_found":  "The specified directory was not found. Check the path and try again.",
	"permission_error":     "Permission denied. Check file/directory permissions.",
	"disk_full":            "Disk is full. Free up space and try again.",
	"temp_file_error":      "Could not create temporary file. Check disk space and permissions.",
}

// GetSuggestion returns a suggestion for the given error type.
// Returns empty string if no suggestion is available.
func GetSuggestion(errorType string) string {
	return CommonSuggestions[strings.ToLower(errorType)]
}

// APIErrorMessages maps API error codes to user-friendly messages.
// These provide clearer context than the raw API error messages.
var APIErrorMessages = map[string]string{
	// Authentication
	"unauthorized":       "Authentication required.",
	"expired_token":      "Your session has expired.",
	"invalid_api_key":    "Invalid API key.",
	"invalid_token":      "Invalid authentication token.",
	"token_revoked":      "Your token has been revoked.",
	"mfa_required":       "Multi-factor authentication required.",
	"account_locked":     "Account locked due to too many failed attempts.",
	"account_disabled":   "Your account has been disabled.",
	"password_expired":   "Your password has expired.",
	"session_invalid":    "Your session is no longer valid.",
	"device_not_trusted": "This device is not trusted.",
	"ip_blocked":         "Your IP address has been blocked.",
	"geo_blocked":        "Access from your location is restricted.",

	// Authorization
	"forbidden":          "Permission denied.",
	"insufficient_scope": "Your API key lacks the required permissions.",
	"read_only":          "This resource is read-only.",
	"owner_only":         "Only the owner can perform this action.",
	"admin_required":     "Admin privileges required.",
	"role_required":      "Insufficient role permissions.",
	"org_mismatch":       "Resource belongs to a different organization.",
	"team_mismatch":      "Resource belongs to a different team.",
	"not_member":         "You are not a member of this organization.",
	"invitation_only":    "Access requires an invitation.",

	// Resource errors
	"not_found":        "Resource not found.",
	"already_exists":   "Resource already exists.",
	"conflict":         "Resource conflict detected.",
	"gone":             "Resource has been deleted.",
	"locked":           "Resource is locked.",
	"archived":         "Resource has been archived.",
	"suspended":        "Resource has been suspended.",
	"version_mismatch": "Resource version mismatch.",
	"stale":            "Resource data is stale.",

	// Rate limiting
	"rate_limited":      "Rate limit exceeded.",
	"quota_exceeded":    "API quota exceeded.",
	"too_many_requests": "Too many requests.",
	"burst_exceeded":    "Request burst limit exceeded.",

	// Plan limits
	"plan_limit_exceeded":    "Plan limit exceeded.",
	"probe_limit_exceeded":   "Probe limit reached.",
	"team_limit_exceeded":    "Team member limit reached.",
	"channel_limit_exceeded": "Notification channel limit reached.",
	"feature_not_available":  "Feature not available on your plan.",
	"upgrade_required":       "Plan upgrade required.",
	"trial_expired":          "Trial period has expired.",
	"subscription_inactive":  "Subscription is inactive.",
	"payment_required":       "Payment required to continue.",
	"payment_failed":         "Payment processing failed.",
	"billing_issue":          "Billing issue detected.",

	// Validation
	"validation":           "Invalid request.",
	"invalid_input":        "Invalid input provided.",
	"malformed_json":       "Malformed JSON in request body.",
	"missing_field":        "Required field is missing.",
	"invalid_field":        "Field value is invalid.",
	"field_too_long":       "Field value exceeds maximum length.",
	"field_too_short":      "Field value is below minimum length.",
	"invalid_format":       "Invalid format.",
	"out_of_range":         "Value is out of allowed range.",
	"constraint_violation": "Constraint violation.",

	// Server errors
	"internal_server":     "Internal server error.",
	"service_unavailable": "Service temporarily unavailable.",
	"bad_gateway":         "Bad gateway error.",
	"gateway_timeout":     "Gateway timeout.",
	"maintenance":         "Service under maintenance.",
	"overloaded":          "Service is overloaded.",
	"database_error":      "Database error occurred.",
	"upstream_error":      "Upstream service error.",
	"configuration_error": "Service configuration error.",
	"dependency_failure":  "Dependent service failure.",
}

// GetUserFriendlyMessage returns a user-friendly message for an API error code.
// Falls back to the provided default if no mapping exists.
func GetUserFriendlyMessage(errorCode string, defaultMsg string) string {
	if msg, ok := APIErrorMessages[strings.ToLower(errorCode)]; ok {
		return msg
	}
	if defaultMsg != "" {
		return defaultMsg
	}
	return "An unexpected error occurred."
}

// ErrorWithContext holds structured error information for formatted output.
type ErrorWithContext struct {
	Message   string
	Context   map[string]string
	Hint      string
	RequestID string
}

// Print outputs the error using the provided formatter.
func (e *ErrorWithContext) Print(f *ErrorFormatter) {
	f.PrintError(e.Message)
	for key, value := range e.Context {
		f.PrintContext(key, value)
	}
	if e.Hint != "" {
		f.PrintSuggestion(e.Hint)
	}
	if e.RequestID != "" {
		f.PrintRequestID(e.RequestID)
	}
}
