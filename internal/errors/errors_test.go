package errors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"syscall"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// setupTest redirects errWriter to a buffer for testing output.
func setupTest() *bytes.Buffer {
	buf := &bytes.Buffer{}
	SetErrWriter(buf)
	return buf
}

func TestHandleError_Nil(t *testing.T) {
	buf := setupTest()

	code := HandleError(nil)

	if code != ExitSuccess {
		t.Errorf("HandleError(nil) = %d, want %d", code, ExitSuccess)
	}
	if buf.Len() != 0 {
		t.Errorf("HandleError(nil) wrote output: %q", buf.String())
	}
}

func TestHandleError_APIError_Unauthorized(t *testing.T) {
	buf := setupTest()

	apiErr := &client.APIError{
		StatusCode: 401,
		Code:       client.ErrorCodeUnauthorized,
		Message:    "Invalid credentials",
	}

	code := HandleError(apiErr)

	if code != ExitAuth {
		t.Errorf("HandleError(Unauthorized) = %d, want %d", code, ExitAuth)
	}
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Authentication required")) {
		t.Errorf("Expected 'Authentication required' in output, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("stackeye login")) {
		t.Errorf("Expected login hint in output, got: %s", output)
	}
}

func TestHandleError_APIError_ExpiredToken(t *testing.T) {
	buf := setupTest()

	apiErr := &client.APIError{
		StatusCode: 401,
		Code:       client.ErrorCodeExpiredToken,
		Message:    "Token expired",
	}

	code := HandleError(apiErr)

	if code != ExitAuth {
		t.Errorf("HandleError(ExpiredToken) = %d, want %d", code, ExitAuth)
	}
	if !bytes.Contains(buf.Bytes(), []byte("session has expired")) {
		t.Errorf("Expected 'session has expired' in output, got: %s", buf.String())
	}
}

func TestHandleError_APIError_InvalidAPIKey(t *testing.T) {
	buf := setupTest()

	apiErr := &client.APIError{
		StatusCode: 401,
		Code:       client.ErrorCodeInvalidAPIKey,
		Message:    "Invalid API key",
	}

	code := HandleError(apiErr)

	if code != ExitAuth {
		t.Errorf("HandleError(InvalidAPIKey) = %d, want %d", code, ExitAuth)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Invalid API key")) {
		t.Errorf("Expected 'Invalid API key' in output, got: %s", buf.String())
	}
}

func TestHandleError_APIError_Forbidden(t *testing.T) {
	buf := setupTest()

	apiErr := &client.APIError{
		StatusCode: 403,
		Code:       client.ErrorCodeForbidden,
		Message:    "Insufficient permissions",
	}

	code := HandleError(apiErr)

	if code != ExitForbidden {
		t.Errorf("HandleError(Forbidden) = %d, want %d", code, ExitForbidden)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Permission denied")) {
		t.Errorf("Expected 'Permission denied' in output, got: %s", buf.String())
	}
}

func TestHandleError_APIError_NotFound(t *testing.T) {
	buf := setupTest()

	apiErr := &client.APIError{
		StatusCode: 404,
		Code:       client.ErrorCodeNotFound,
		Message:    "Probe not found",
	}

	code := HandleError(apiErr)

	if code != ExitNotFound {
		t.Errorf("HandleError(NotFound) = %d, want %d", code, ExitNotFound)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Probe not found")) {
		t.Errorf("Expected message in output, got: %s", buf.String())
	}
}

func TestHandleError_APIError_RateLimited(t *testing.T) {
	buf := setupTest()

	apiErr := &client.APIError{
		StatusCode: 429,
		Code:       client.ErrorCodeRateLimited,
		Message:    "Too many requests",
	}

	code := HandleError(apiErr)

	if code != ExitRateLimited {
		t.Errorf("HandleError(RateLimited) = %d, want %d", code, ExitRateLimited)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Rate limit exceeded")) {
		t.Errorf("Expected 'Rate limit exceeded' in output, got: %s", buf.String())
	}
	if !bytes.Contains(buf.Bytes(), []byte("wait")) {
		t.Errorf("Expected retry hint in output, got: %s", buf.String())
	}
}

func TestHandleError_APIError_Validation(t *testing.T) {
	buf := setupTest()

	apiErr := &client.APIError{
		StatusCode: 400,
		Code:       client.ErrorCodeValidation,
		Message:    "Validation failed",
		Details: map[string]any{
			"fields": map[string]any{
				"name": "name is required",
				"url":  "invalid URL format",
			},
		},
	}

	code := HandleError(apiErr)

	if code != ExitMisuse {
		t.Errorf("HandleError(Validation) = %d, want %d", code, ExitMisuse)
	}
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Invalid request")) {
		t.Errorf("Expected 'Invalid request' in output, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("name")) {
		t.Errorf("Expected field 'name' in output, got: %s", output)
	}
}

func TestHandleError_APIError_PlanLimit(t *testing.T) {
	buf := setupTest()

	apiErr := &client.APIError{
		StatusCode: 403,
		Code:       client.ErrorCodePlanLimitExceed,
		Message:    "Maximum probes reached for your plan",
	}

	code := HandleError(apiErr)

	if code != ExitPlanLimit {
		t.Errorf("HandleError(PlanLimit) = %d, want %d", code, ExitPlanLimit)
	}
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Maximum probes")) {
		t.Errorf("Expected plan limit message in output, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("billing")) {
		t.Errorf("Expected billing hint in output, got: %s", output)
	}
}

func TestHandleError_APIError_ServerError(t *testing.T) {
	buf := setupTest()

	apiErr := &client.APIError{
		StatusCode: 500,
		Code:       client.ErrorCodeInternalServer,
		Message:    "Internal server error",
		RequestID:  "req_12345",
	}

	code := HandleError(apiErr)

	if code != ExitServerError {
		t.Errorf("HandleError(ServerError) = %d, want %d", code, ExitServerError)
	}
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Server error")) {
		t.Errorf("Expected 'Server error' in output, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("req_12345")) {
		t.Errorf("Expected request ID in output, got: %s", output)
	}
}

func TestHandleError_ContextDeadlineExceeded(t *testing.T) {
	buf := setupTest()

	code := HandleError(context.DeadlineExceeded)

	if code != ExitTimeout {
		t.Errorf("HandleError(DeadlineExceeded) = %d, want %d", code, ExitTimeout)
	}
	if !bytes.Contains(buf.Bytes(), []byte("timed out")) {
		t.Errorf("Expected timeout message in output, got: %s", buf.String())
	}
}

func TestHandleError_ContextCanceled(t *testing.T) {
	_ = setupTest() // Output is intentionally not checked (user hit Ctrl+C)

	code := HandleError(context.Canceled)

	if code != ExitError {
		t.Errorf("HandleError(Canceled) = %d, want %d", code, ExitError)
	}
	// Canceled should not print an error (user probably hit Ctrl+C)
}

func TestHandleError_ConnectionRefused(t *testing.T) {
	buf := setupTest()

	// Wrap in a url.Error like real HTTP client would
	err := &url.Error{
		Op:  "Get",
		URL: "https://api.stackeye.io/v1/me",
		Err: &net.OpError{
			Op:  "dial",
			Net: "tcp",
			Err: syscall.ECONNREFUSED,
		},
	}

	code := HandleError(err)

	if code != ExitNetwork {
		t.Errorf("HandleError(ConnectionRefused) = %d, want %d", code, ExitNetwork)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Connection refused")) {
		t.Errorf("Expected 'Connection refused' in output, got: %s", buf.String())
	}
}

func TestHandleError_DNSError(t *testing.T) {
	buf := setupTest()

	err := &net.DNSError{
		Err:        "no such host",
		Name:       "api.stackeye.io",
		IsNotFound: true,
	}

	code := HandleError(err)

	if code != ExitNetwork {
		t.Errorf("HandleError(DNSError) = %d, want %d", code, ExitNetwork)
	}
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("DNS lookup failed")) {
		t.Errorf("Expected DNS error message in output, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("api.stackeye.io")) {
		t.Errorf("Expected hostname in output, got: %s", output)
	}
}

func TestHandleError_Timeout(t *testing.T) {
	buf := setupTest()

	// Create a timeout error
	err := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &timeoutError{},
	}

	code := HandleError(err)

	if code != ExitTimeout {
		t.Errorf("HandleError(Timeout) = %d, want %d", code, ExitTimeout)
	}
	if !bytes.Contains(buf.Bytes(), []byte("timed out")) {
		t.Errorf("Expected timeout message in output, got: %s", buf.String())
	}
}

// timeoutError implements net.Error with Timeout() = true
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestHandleError_GenericError(t *testing.T) {
	buf := setupTest()

	err := errors.New("something went wrong")

	code := HandleError(err)

	if code != ExitError {
		t.Errorf("HandleError(generic) = %d, want %d", code, ExitError)
	}
	if !bytes.Contains(buf.Bytes(), []byte("something went wrong")) {
		t.Errorf("Expected error message in output, got: %s", buf.String())
	}
}

func TestHandleError_WrappedError(t *testing.T) {
	_ = setupTest() // Output not checked - we're testing exit code only

	// Error wrapped with fmt.Errorf
	inner := context.DeadlineExceeded
	err := fmt.Errorf("operation failed: %w", inner)

	code := HandleError(err)

	if code != ExitTimeout {
		t.Errorf("HandleError(wrapped DeadlineExceeded) = %d, want %d", code, ExitTimeout)
	}
}

func TestExitCodeName(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{ExitSuccess, "success"},
		{ExitError, "error"},
		{ExitMisuse, "misuse"},
		{ExitAuth, "auth_required"},
		{ExitForbidden, "forbidden"},
		{ExitNotFound, "not_found"},
		{ExitRateLimited, "rate_limited"},
		{ExitServerError, "server_error"},
		{ExitNetwork, "network_error"},
		{ExitTimeout, "timeout"},
		{ExitPlanLimit, "plan_limit"},
		{ExitSIGINT, "sigint"},
		{ExitSIGTERM, "sigterm"},
		{99, "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := ExitCodeName(tt.code)
			if got != tt.want {
				t.Errorf("ExitCodeName(%d) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestContainsNetworkError(t *testing.T) {
	tests := []struct {
		errStr string
		want   bool
	}{
		// Real network errors - should match
		{"connection refused", true},
		{"no such host", true},
		{"network is unreachable", true},
		{"i/o timeout", true},
		{"broken pipe", true},
		{"connection reset by peer", true},
		// Generic errors - should not match
		{"some other error", false},
		{"validation failed", false},
		// HTTP errors should not match (dial tcp pattern removed)
		{"dial tcp: i/o timeout", true}, // Still matches via i/o timeout
		{"failed to dial", false},       // dial alone should not match
	}

	for _, tt := range tests {
		t.Run(tt.errStr, func(t *testing.T) {
			got := containsNetworkError(tt.errStr)
			if got != tt.want {
				t.Errorf("containsNetworkError(%q) = %v, want %v", tt.errStr, got, tt.want)
			}
		})
	}
}

func TestHandleError_WrappedAPIError(t *testing.T) {
	buf := setupTest()

	// Create an APIError and wrap it like commands do
	apiErr := &client.APIError{
		StatusCode: 401,
		Code:       client.ErrorCodeUnauthorized,
		Message:    "Invalid API key",
	}
	wrapped := fmt.Errorf("failed to list organizations: %w", apiErr)

	code := HandleError(wrapped)

	if code != ExitAuth {
		t.Errorf("HandleError(wrapped 401) = %d, want %d", code, ExitAuth)
	}
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Authentication required")) {
		t.Errorf("Expected 'Authentication required' in output, got: %s", output)
	}
}

func TestHandleError_DoesNotMisclassifyAuthAsNetwork(t *testing.T) {
	// Error message that might look like a network error but is actually auth failure
	tests := []struct {
		name   string
		errStr string
	}{
		{"401 Unauthorized", "GET https://api.stackeye.io: 401 Unauthorized"},
		{"status 401", "request failed: status 401"},
		{"forbidden", "operation forbidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// containsHTTPError should return true for these
			if !containsHTTPError(tt.errStr) {
				t.Errorf("containsHTTPError(%q) = false, want true", tt.errStr)
			}
			// The combination should NOT classify as network error
			isNetwork := containsNetworkError(tt.errStr) && !containsHTTPError(tt.errStr)
			if isNetwork {
				t.Errorf("Error %q misclassified as network error", tt.errStr)
			}
		})
	}
}

func TestDetectHTTPStatusCode(t *testing.T) {
	tests := []struct {
		errStr string
		want   int
	}{
		{"401 Unauthorized", 401},
		{"403 Forbidden", 403},
		{"404 Not Found", 404},
		{"500 Internal Server", 500},
		{"502 Bad Gateway", 502},
		{"503 Service Unavailable", 503},
		{"status: 401", 401},
		{"status 403", 403},
		{"returned 404", 404},
		{"connection refused", 0},
		{"some random error", 0},
	}

	for _, tt := range tests {
		t.Run(tt.errStr, func(t *testing.T) {
			got := detectHTTPStatusCode(tt.errStr)
			if got != tt.want {
				t.Errorf("detectHTTPStatusCode(%q) = %d, want %d", tt.errStr, got, tt.want)
			}
		})
	}
}

func TestHandleHTTPStatusCode(t *testing.T) {
	tests := []struct {
		code     int
		wantExit int
	}{
		{401, ExitAuth},
		{403, ExitForbidden},
		{404, ExitNotFound},
		{500, ExitServerError},
		{502, ExitServerError},
		{503, ExitServerError},
		{200, ExitError}, // Not an error code
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("code_%d", tt.code), func(t *testing.T) {
			_ = setupTest() // Clear output buffer
			got := handleHTTPStatusCode(tt.code)
			if got != tt.wantExit {
				t.Errorf("handleHTTPStatusCode(%d) = %d, want %d", tt.code, got, tt.wantExit)
			}
		})
	}
}

func TestContainsHTTPError(t *testing.T) {
	tests := []struct {
		errStr string
		want   bool
	}{
		{"status code: 401", true},
		{"401 Unauthorized", true},
		{"403 Forbidden", true},
		{"resource not found", true},
		{"unauthorized access", true},
		{"connection refused", false},
		{"network is unreachable", false},
		{"i/o timeout", false},
	}

	for _, tt := range tests {
		t.Run(tt.errStr, func(t *testing.T) {
			got := containsHTTPError(tt.errStr)
			if got != tt.want {
				t.Errorf("containsHTTPError(%q) = %v, want %v", tt.errStr, got, tt.want)
			}
		})
	}
}

func TestHandleError_HTTPStatusInErrorMessage(t *testing.T) {
	buf := setupTest()

	// Error that contains HTTP status but isn't an APIError
	err := errors.New("GET https://api.stackeye.io/v1/orgs: 401 Unauthorized")

	code := HandleError(err)

	if code != ExitAuth {
		t.Errorf("HandleError(401 in message) = %d, want %d", code, ExitAuth)
	}
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Authentication failed")) {
		t.Errorf("Expected 'Authentication failed' in output, got: %s", output)
	}
}
