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
)

// errWriter is the destination for error messages.
// Can be overridden for testing.
var errWriter io.Writer = os.Stderr

// SetErrWriter sets the writer for error output. Used for testing.
func SetErrWriter(w io.Writer) {
	errWriter = w
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
	printError(err.Error())
	return ExitError
}

// handleAPIError maps SDK APIError types to exit codes and user-friendly messages.
func handleAPIError(apiErr *client.APIError) int {
	switch {
	case apiErr.IsUnauthorized():
		msg := "Authentication required."
		if apiErr.Code == client.ErrorCodeExpiredToken {
			msg = "Your session has expired."
		} else if apiErr.Code == client.ErrorCodeInvalidAPIKey {
			msg = "Invalid API key."
		}
		printError(msg)
		printHint("Run 'stackeye login' to authenticate.")
		return ExitAuth

	// Check plan limit BEFORE forbidden since both can have status 403
	case apiErr.IsPlanLimitExceeded():
		msg := "Plan limit exceeded."
		if apiErr.Message != "" {
			msg = apiErr.Message
		}
		printError(msg)
		printHint("Upgrade your plan at https://stackeye.io/billing")
		return ExitPlanLimit

	case apiErr.IsForbidden():
		msg := "Permission denied."
		if apiErr.Message != "" {
			msg = fmt.Sprintf("Permission denied: %s", apiErr.Message)
		}
		printError(msg)
		printHint("You may not have access to this resource or organization.")
		return ExitForbidden

	case apiErr.IsNotFound():
		msg := "Resource not found."
		if apiErr.Message != "" {
			msg = apiErr.Message
		}
		printError(msg)
		return ExitNotFound

	case apiErr.IsRateLimited():
		printError("Rate limit exceeded.")
		printHint("Please wait a moment and try again.")
		return ExitRateLimited

	case apiErr.IsValidationError():
		printValidationError(apiErr)
		return ExitMisuse

	case apiErr.IsServerError():
		printError("Server error occurred.")
		if apiErr.RequestID != "" {
			printHint(fmt.Sprintf("Request ID: %s (include this when contacting support)", apiErr.RequestID))
		}
		printHint("If this persists, check https://status.stackeye.io")
		return ExitServerError

	default:
		// Unknown API error
		msg := apiErr.Message
		if msg == "" {
			msg = "An unexpected error occurred"
		}
		printError(msg)
		if apiErr.RequestID != "" {
			printHint(fmt.Sprintf("Request ID: %s", apiErr.RequestID))
		}
		return ExitError
	}
}

// printValidationError formats and prints validation errors with field details.
func printValidationError(apiErr *client.APIError) {
	printError("Invalid request.")

	// Print field-level validation errors if available
	if fields := apiErr.ValidationErrors(); len(fields) > 0 {
		for field, msg := range fields {
			fmt.Fprintf(errWriter, "  %s: %s\n", field, msg)
		}
	} else if apiErr.Message != "" {
		fmt.Fprintf(errWriter, "  %s\n", apiErr.Message)
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
		printError("Connection timed out.")
		printHint("The server took too long to respond. Try again later.")
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
		printError(fmt.Sprintf("DNS lookup failed: %s", dnsErr.Name))
		printHint("Check your network connection and DNS settings.")
		return ExitNetwork
	}

	// Check for connection refused
	if errors.Is(err, syscall.ECONNREFUSED) {
		printError("Connection refused.")
		printHint("The server may be down or unreachable. Check your network connection.")
		return ExitNetwork
	}

	// Check for connection reset
	if errors.Is(err, syscall.ECONNRESET) {
		printError("Connection reset by server.")
		printHint("Try again. If this persists, check https://status.stackeye.io")
		return ExitNetwork
	}

	// Check for error message patterns, but filter out HTTP errors
	// to prevent misclassifying authentication failures as network errors.
	errStr := err.Error()
	if containsNetworkError(errStr) && !containsHTTPError(errStr) {
		printError("Network error occurred.")
		printHint("Check your internet connection and try again.")
		return ExitNetwork
	}

	return ExitError
}

// handleNetOpError handles net.OpError with specific messaging.
func handleNetOpError(netErr *net.OpError) int {
	switch {
	case netErr.Timeout():
		printError("Connection timed out.")
		printHint("The server took too long to respond. Try again later.")
		return ExitTimeout

	case netErr.Temporary():
		printError("Temporary network error.")
		printHint("Try again in a moment.")
		return ExitNetwork

	default:
		// Check the wrapped error
		if netErr.Err != nil {
			if code := handleNetworkError(netErr.Err); code != ExitError {
				return code
			}
		}
		printError("Network error occurred.")
		printHint("Check your internet connection and try again.")
		return ExitNetwork
	}
}

// handleContextError checks for context-related errors.
func handleContextError(err error) int {
	if errors.Is(err, context.DeadlineExceeded) {
		printError("Operation timed out.")
		printHint("The request took too long. Try again or check your connection.")
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
		printError("Authentication failed.")
		printHint("Your API key may be invalid or expired. Run 'stackeye login' to authenticate.")
		return ExitAuth
	case code == 403:
		printError("Access denied.")
		printHint("You don't have permission to perform this action.")
		return ExitForbidden
	case code == 404:
		printError("Resource not found.")
		return ExitNotFound
	case code >= 500:
		printError("Server error occurred.")
		printHint("The StackEye API is experiencing issues. Try again later.")
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

// printError prints an error message to stderr.
func printError(msg string) {
	fmt.Fprintf(errWriter, "Error: %s\n", msg)
}

// printHint prints a hint message to stderr (indented for clarity).
func printHint(msg string) {
	fmt.Fprintf(errWriter, "  %s\n", msg)
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
	default:
		return fmt.Sprintf("unknown(%d)", code)
	}
}
