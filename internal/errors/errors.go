// Package errors provides CLI error handling utilities that map API errors
// to user-friendly messages and appropriate exit codes.
//
// Exit codes follow POSIX conventions with extensions for specific error types:
//
//	0 - Success
//	1 - General error
//	2 - Misuse (invalid arguments, configuration)
//	3 - Authentication required
//	4 - Permission denied
//	5 - Resource not found
//	6 - Rate limited
//	7 - Server error
//
// Usage:
//
//	if err := doSomething(); err != nil {
//	    code := errors.HandleError(err)
//	    os.Exit(code)
//	}
package errors

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// Exit codes for CLI operations.
// These follow POSIX conventions with extensions for StackEye-specific errors.
const (
	// ExitSuccess indicates successful completion.
	ExitSuccess = 0

	// ExitError indicates a general error occurred.
	ExitError = 1

	// ExitMisuse indicates command line misuse (invalid arguments, bad config).
	ExitMisuse = 2

	// ExitAuth indicates authentication is required or credentials are invalid.
	ExitAuth = 3

	// ExitForbidden indicates permission was denied for the operation.
	ExitForbidden = 4

	// ExitNotFound indicates the requested resource was not found.
	ExitNotFound = 5

	// ExitRateLimited indicates the request was rate limited.
	ExitRateLimited = 6

	// ExitServerError indicates a server-side error occurred.
	ExitServerError = 7

	// ExitNetwork indicates a network connectivity error.
	ExitNetwork = 8

	// ExitTimeout indicates the operation timed out.
	ExitTimeout = 9

	// ExitPlanLimit indicates a plan limit was exceeded.
	ExitPlanLimit = 10

	// ExitSIGINT indicates the process was interrupted by Ctrl+C (128 + signal 2).
	// This follows POSIX convention for signal-terminated processes.
	ExitSIGINT = 130

	// ExitSIGTERM indicates the process was terminated by SIGTERM (128 + signal 15).
	// This follows POSIX convention for signal-terminated processes.
	ExitSIGTERM = 143
)

// errWriter is the destination for error messages.
// Can be overridden for testing.
var errWriter io.Writer = os.Stderr

// formatter provides color-coded error output. It is initialized lazily
// and updated when SetErrWriter is called for testing.
var formatter = NewErrorFormatter()

// SetErrWriter sets the writer for error output. Used for testing.
// Also reinitializes the formatter to use the new writer with colors disabled,
// ensuring test output contains plain text for assertion matching.
func SetErrWriter(w io.Writer) {
	errWriter = w
	formatter = NewErrorFormatterWithColorManager(sdkoutput.NewColorManager(sdkoutput.ColorNever), w)
}

// HandleError processes an error, prints a user-friendly message to stderr,
// and returns the appropriate exit code.
//
// If err is nil, returns ExitSuccess (0).
//
// The function handles these error types:
//   - *client.APIError: Maps API errors to appropriate exit codes and messages
//   - Network errors: Connection refused, DNS failures, timeouts
//   - Context errors: Deadline exceeded, canceled
//   - Generic errors: Returns ExitError with the error message
func HandleError(err error) int {
	if err == nil {
		return ExitSuccess
	}

	// Debug: show error type classification
	debug := os.Getenv("STACKEYE_DEBUG") != ""
	if debug {
		fmt.Fprintf(errWriter, "[debug] Error type: %T\n", err)
		fmt.Fprintf(errWriter, "[debug] Error message: %s\n", err.Error())
	}

	// Check for API errors first (most specific)
	if apiErr := client.IsAPIError(err); apiErr != nil {
		if debug {
			fmt.Fprintf(errWriter, "[debug] Detected APIError: status=%d code=%s\n",
				apiErr.StatusCode, apiErr.Code)
		}
		return handleAPIError(apiErr)
	}

	// Check for HTTP status codes in error message before network error fallback.
	// This catches cases where the API returned an error that wasn't properly
	// parsed as an APIError (e.g., wrapped errors with status codes).
	errStr := err.Error()
	if code := detectHTTPStatusCode(errStr); code > 0 {
		if debug {
			fmt.Fprintf(errWriter, "[debug] Detected HTTP status code %d in error message\n", code)
		}
		if exitCode := handleHTTPStatusCode(code); exitCode != ExitError {
			return exitCode
		}
	}

	// Check for network-related errors
	if code := handleNetworkError(err); code != ExitError {
		return code
	}

	// Check for context errors (timeout, cancellation)
	if code := handleContextError(err); code != ExitError {
		return code
	}

	// Generic error fallback
	formatter.PrintError(err.Error())
	return ExitError
}

// handleAPIError maps SDK APIError types to exit codes and user-friendly messages.
func handleAPIError(apiErr *client.APIError) int {
	switch {
	case apiErr.IsUnauthorized():
		msg := GetUserFriendlyMessage(string(apiErr.Code), "Authentication required.")
		formatter.PrintError(msg)
		formatter.PrintHint(GetSuggestion("auth_required"))
		return ExitAuth

	// Check plan limit BEFORE forbidden since both can have status 403
	case apiErr.IsPlanLimitExceeded():
		msg := "Plan limit exceeded."
		if apiErr.Message != "" {
			msg = apiErr.Message
		}
		formatter.PrintError(msg)
		formatter.PrintHint(GetSuggestion("plan_limit"))
		return ExitPlanLimit

	case apiErr.IsForbidden():
		msg := GetUserFriendlyMessage(string(apiErr.Code), "Permission denied.")
		if apiErr.Message != "" {
			msg = fmt.Sprintf("Permission denied: %s", apiErr.Message)
		}
		formatter.PrintError(msg)
		formatter.PrintHint(GetSuggestion("forbidden"))
		return ExitForbidden

	case apiErr.IsNotFound():
		msg := GetUserFriendlyMessage(string(apiErr.Code), "Resource not found.")
		if apiErr.Message != "" {
			msg = apiErr.Message
		}
		formatter.PrintError(msg)
		return ExitNotFound

	case apiErr.IsRateLimited():
		formatter.PrintError(GetUserFriendlyMessage(string(apiErr.Code), "Rate limit exceeded."))
		formatter.PrintHint(GetSuggestion("rate_limited"))
		return ExitRateLimited

	case apiErr.IsValidationError():
		printValidationError(apiErr)
		return ExitMisuse

	case apiErr.IsServerError():
		formatter.PrintError(GetUserFriendlyMessage(string(apiErr.Code), "Server error occurred."))
		formatter.PrintRequestID(apiErr.RequestID)
		formatter.PrintHint(GetSuggestion("server_error"))
		return ExitServerError

	default:
		// Unknown API error
		msg := GetUserFriendlyMessage(string(apiErr.Code), apiErr.Message)
		formatter.PrintError(msg)
		formatter.PrintRequestID(apiErr.RequestID)
		return ExitError
	}
}

// printValidationError formats and prints validation errors with field details.
func printValidationError(apiErr *client.APIError) {
	formatter.PrintError("Invalid request.")

	// Print field-level validation errors if available
	if fields := apiErr.ValidationErrors(); len(fields) > 0 {
		formatter.PrintValidationErrors(fields)
	} else if apiErr.Message != "" {
		formatter.PrintHint(apiErr.Message)
	}
}

// netTimeout is an interface for checking timeout conditions (part of net.Error).
type netTimeout interface {
	Timeout() bool
}

// handleNetworkError checks for network-related errors and returns appropriate exit codes.
func handleNetworkError(err error) int {
	// Check for timeout interface first (highest priority)
	var te netTimeout
	if errors.As(err, &te) && te.Timeout() {
		formatter.PrintError("Connection timed out.")
		formatter.PrintHint("The server took too long to respond. Try again later.")
		return ExitTimeout
	}

	// Unwrap the error to find the root cause
	unwrapped := errors.Unwrap(err)
	if unwrapped != nil {
		if code := handleNetworkError(unwrapped); code != ExitError {
			return code
		}
	}

	// Check for URL errors (wraps net errors)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return handleNetworkError(urlErr.Err)
	}

	// Check for specific network errors
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return handleNetOpError(netErr)
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		formatter.PrintError(fmt.Sprintf("DNS lookup failed: %s", dnsErr.Name))
		formatter.PrintHint(GetSuggestion("dns_failure"))
		return ExitNetwork
	}

	// Check for connection refused
	if errors.Is(err, syscall.ECONNREFUSED) {
		formatter.PrintError("Connection refused.")
		formatter.PrintHint(GetSuggestion("connection_refused"))
		return ExitNetwork
	}

	// Check for connection reset
	if errors.Is(err, syscall.ECONNRESET) {
		formatter.PrintError("Connection reset by server.")
		formatter.PrintHint(GetSuggestion("connection_reset"))
		return ExitNetwork
	}

	// Check for error message patterns, but filter out HTTP errors
	// to prevent misclassifying authentication failures as network errors.
	errStr := err.Error()
	if containsNetworkError(errStr) && !containsHTTPError(errStr) {
		formatter.PrintError("Network error occurred.")
		formatter.PrintHint(GetSuggestion("network_error"))
		return ExitNetwork
	}

	return ExitError
}

// handleNetOpError handles net.OpError with specific messaging.
func handleNetOpError(netErr *net.OpError) int {
	switch {
	case netErr.Timeout():
		formatter.PrintError("Connection timed out.")
		formatter.PrintHint("The server took too long to respond. Try again later.")
		return ExitTimeout

	case netErr.Temporary():
		formatter.PrintError("Temporary network error.")
		formatter.PrintHint("Try again in a moment.")
		return ExitNetwork

	default:
		// Check the wrapped error
		if netErr.Err != nil {
			if code := handleNetworkError(netErr.Err); code != ExitError {
				return code
			}
		}
		formatter.PrintError("Network error occurred.")
		formatter.PrintHint(GetSuggestion("network_error"))
		return ExitNetwork
	}
}

// handleContextError checks for context-related errors.
func handleContextError(err error) int {
	if errors.Is(err, context.DeadlineExceeded) {
		formatter.PrintError("Operation timed out.")
		formatter.PrintHint(GetSuggestion("timeout"))
		return ExitTimeout
	}

	if errors.Is(err, context.Canceled) {
		// User likely canceled (Ctrl+C), no error message needed
		return ExitError
	}

	return ExitError
}

// containsNetworkError checks if an error message indicates a network issue.
// This function has been refined to avoid false positives - patterns like
// "dial tcp" and "dial udp" are too aggressive and can match HTTP errors
// that mention dial operations in their context.
func containsNetworkError(errStr string) bool {
	patterns := []string{
		"connection refused",
		"no such host",
		"network is unreachable",
		"no route to host",
		"connection reset by peer",
		"broken pipe",
		"i/o timeout",
		// REMOVED: "dial tcp", "dial udp" - too aggressive, can match HTTP error contexts
	}

	lower := strings.ToLower(errStr)
	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// detectHTTPStatusCode looks for HTTP status codes in error messages.
// Returns the status code if found, 0 otherwise.
// This is used as a fallback when errors contain HTTP status information
// but weren't properly parsed as APIError (e.g., wrapped errors).
func detectHTTPStatusCode(errStr string) int {
	// Look for common HTTP error patterns in error messages
	// These patterns match how HTTP libraries typically format errors
	lower := strings.ToLower(errStr)

	// Check for explicit status patterns like "status: 401" or "status code: 401"
	statusPatterns := map[string]int{
		"401 unauthorized":        401,
		"403 forbidden":           403,
		"404 not found":           404,
		"500 internal server":     500,
		"502 bad gateway":         502,
		"503 service unavailable": 503,
	}

	for pattern, code := range statusPatterns {
		if strings.Contains(lower, pattern) {
			return code
		}
	}

	// Check for bare status codes with context (e.g., "returned 401", "status 401")
	codePatterns := []struct {
		prefix string
		code   int
	}{
		{"status: 401", 401},
		{"status 401", 401},
		{"returned 401", 401},
		{"status: 403", 403},
		{"status 403", 403},
		{"returned 403", 403},
		{"status: 404", 404},
		{"status 404", 404},
		{"returned 404", 404},
		{"status: 500", 500},
		{"status 500", 500},
		{"returned 500", 500},
	}

	for _, p := range codePatterns {
		if strings.Contains(lower, p.prefix) {
			return p.code
		}
	}

	return 0
}

// handleHTTPStatusCode provides fallback handling for HTTP errors that
// weren't properly parsed as APIError. This improves error messages when
// the SDK's error parsing fails but we can still detect the status code.
func handleHTTPStatusCode(code int) int {
	switch {
	case code == 401:
		formatter.PrintError("Authentication failed.")
		formatter.PrintHint(GetSuggestion("auth_required"))
		return ExitAuth
	case code == 403:
		formatter.PrintError("Access denied.")
		formatter.PrintHint(GetSuggestion("permission_denied"))
		return ExitForbidden
	case code == 404:
		formatter.PrintError("Resource not found.")
		return ExitNotFound
	case code >= 500:
		formatter.PrintError("Server error occurred.")
		formatter.PrintHint(GetSuggestion("server_error"))
		return ExitServerError
	default:
		return ExitError
	}
}

// containsHTTPError checks if an error message looks like an HTTP response error.
// This is used to prevent misclassifying HTTP errors as network errors.
func containsHTTPError(errStr string) bool {
	httpPatterns := []string{
		"status code",
		"status:",
		"401",
		"403",
		"404",
		"500",
		"502",
		"503",
		"unauthorized",
		"forbidden",
		"not found",
	}
	lower := strings.ToLower(errStr)
	for _, pattern := range httpPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// ExitCodeName returns a human-readable name for an exit code.
// Useful for logging and debugging.
func ExitCodeName(code int) string {
	switch code {
	case ExitSuccess:
		return "success"
	case ExitError:
		return "error"
	case ExitMisuse:
		return "misuse"
	case ExitAuth:
		return "auth_required"
	case ExitForbidden:
		return "forbidden"
	case ExitNotFound:
		return "not_found"
	case ExitRateLimited:
		return "rate_limited"
	case ExitServerError:
		return "server_error"
	case ExitNetwork:
		return "network_error"
	case ExitTimeout:
		return "timeout"
	case ExitPlanLimit:
		return "plan_limit"
	case ExitSIGINT:
		return "sigint"
	case ExitSIGTERM:
		return "sigterm"
	default:
		return fmt.Sprintf("unknown(%d)", code)
	}
}
