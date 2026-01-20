// Package auth provides browser-based authentication for the StackEye CLI.
//
// This package handles the OAuth-like flow for CLI authentication:
//  1. Start a local callback server on a random port
//  2. Open the user's browser to the StackEye web UI /cli-auth page
//  3. Wait for the callback with the API key
//  4. Return the authentication result
//
// The actual API key storage is handled by the stackeye-go-sdk/config package.
// This package only handles the browser flow mechanics.
//
// Usage:
//
//	result, err := auth.BrowserLogin(auth.Options{
//	    APIURL:  "https://api.stackeye.io",
//	    Timeout: 5 * time.Minute,
//	})
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Logged in as org: %s\n", result.OrgName)
package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	sdkauth "github.com/StackEye-IO/stackeye-go-sdk/auth"
)

// debugLogger is the logger for debug output. Nil when debug is disabled.
var debugLogger *log.Logger

// SetDebug enables or disables debug logging for the auth package.
func SetDebug(enabled bool) {
	if enabled {
		debugLogger = log.New(os.Stderr, "[auth-debug] ", log.Ltime|log.Lmicroseconds)
	} else {
		debugLogger = nil
	}
}

// debugf logs a debug message if debug logging is enabled.
func debugf(format string, args ...interface{}) {
	if debugLogger != nil {
		debugLogger.Printf(format, args...)
	}
}

// Default configuration values.
const (
	// DefaultTimeout is the maximum time to wait for the browser callback.
	DefaultTimeout = 5 * time.Minute

	// DefaultAPIURL is the production API endpoint.
	DefaultAPIURL = "https://api.stackeye.io"

	// callbackPath is the path for the callback handler.
	callbackPath = "/callback"
)

// Common errors returned by this package.
var (
	// ErrTimeout is returned when the browser callback times out.
	ErrTimeout = errors.New("auth: login timed out waiting for browser callback")

	// ErrCanceled is returned when the login is canceled.
	ErrCanceled = errors.New("auth: login canceled")

	// ErrInvalidAPIKey is returned when the received API key is invalid.
	ErrInvalidAPIKey = errors.New("auth: received invalid API key format")

	// ErrMissingAPIKey is returned when the callback is missing the API key.
	ErrMissingAPIKey = errors.New("auth: callback missing api_key parameter")

	// ErrForbidden is returned when a non-localhost request is received.
	ErrForbidden = errors.New("auth: request from non-localhost IP rejected")
)

// Options configures the browser login flow.
type Options struct {
	// APIURL is the StackEye API URL to authenticate against.
	// If empty, defaults to DefaultAPIURL.
	APIURL string

	// Timeout is the maximum time to wait for the browser callback.
	// If zero, defaults to DefaultTimeout.
	Timeout time.Duration

	// OnBrowserOpen is called when the browser is about to be opened.
	// If nil, a default message is printed to stdout.
	OnBrowserOpen func(url string)

	// OnWaiting is called while waiting for the browser callback.
	// If nil, a default message is printed to stdout.
	OnWaiting func()
}

// Result holds the authentication result from a successful browser login.
type Result struct {
	// APIKey is the generated API key for CLI authentication.
	APIKey string

	// OrgID is the organization ID associated with the API key.
	OrgID string

	// OrgName is the organization name associated with the API key.
	OrgName string
}

// BrowserLogin performs browser-based authentication.
//
// This function:
//  1. Starts a local HTTP server on a random port
//  2. Opens the browser to the StackEye web UI /cli-auth page
//  3. Waits for the callback with the API key
//  4. Returns the authentication result
//
// The caller is responsible for storing the API key using the config package.
func BrowserLogin(opts Options) (*Result, error) {
	return BrowserLoginWithContext(context.Background(), opts)
}

// BrowserLoginWithContext performs browser-based authentication with context support.
func BrowserLoginWithContext(ctx context.Context, opts Options) (*Result, error) {
	debugf("BrowserLoginWithContext started")
	debugf("Input API URL: %s", opts.APIURL)
	debugf("Input Timeout: %v", opts.Timeout)

	// Apply defaults
	if opts.APIURL == "" {
		opts.APIURL = DefaultAPIURL
		debugf("Using default API URL: %s", opts.APIURL)
	}
	if opts.Timeout == 0 {
		opts.Timeout = DefaultTimeout
		debugf("Using default timeout: %v", opts.Timeout)
	}

	// Start local callback server
	debugf("Starting local callback server on 127.0.0.1:0")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		debugf("Failed to start local server: %v", err)
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	defer listener.Close()

	// Get the assigned port
	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d%s", port, callbackPath)
	debugf("Callback server started on port %d", port)
	debugf("Callback URL: %s", callbackURL)

	// Build the web UI URL
	debugf("Building web UI URL from API URL: %s", opts.APIURL)
	webUIURL, err := BuildWebUIURL(opts.APIURL, callbackURL)
	if err != nil {
		debugf("Failed to build web UI URL: %v", err)
		return nil, fmt.Errorf("failed to build web UI URL: %w", err)
	}
	debugf("Final web UI URL: %s", webUIURL)

	// Create channel to receive result
	resultCh := make(chan callbackResult, 1)

	// Create HTTP server with callback handler
	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, makeCallbackHandler(resultCh))

	server := &http.Server{
		Handler: mux,
	}

	// Start server in goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Server.Serve returns when the server is closed; we ignore the error
		// since we're intentionally shutting it down via server.Shutdown()
		_ = server.Serve(listener)
	}()

	// Notify about browser opening
	if opts.OnBrowserOpen != nil {
		opts.OnBrowserOpen(webUIURL)
	} else {
		fmt.Printf("Opening browser to: %s\n", webUIURL)
	}

	// Open browser
	if err := OpenBrowser(webUIURL); err != nil {
		fmt.Printf("Warning: could not open browser: %v\n", err)
		fmt.Printf("Please visit: %s\n", webUIURL)
	}

	// Notify about waiting
	if opts.OnWaiting != nil {
		opts.OnWaiting()
	} else {
		fmt.Println("Waiting for authentication...")
		fmt.Println("(If the browser doesn't open, visit the URL manually)")
		fmt.Println()
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Wait for callback
	var result callbackResult
	select {
	case result = <-resultCh:
		// Received callback
	case <-timeoutCtx.Done():
		// Shutdown server
		shutdownServer(server, &wg)
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("%w: waited %v", ErrTimeout, opts.Timeout)
		}
		return nil, fmt.Errorf("%w: %v", ErrCanceled, timeoutCtx.Err())
	case <-ctx.Done():
		// Parent context canceled
		shutdownServer(server, &wg)
		return nil, fmt.Errorf("%w: %v", ErrCanceled, ctx.Err())
	}

	// Shutdown server
	shutdownServer(server, &wg)

	// Check for callback errors
	if result.err != nil {
		return nil, result.err
	}

	// Validate API key format
	if !sdkauth.ValidateAPIKey(result.apiKey) {
		return nil, ErrInvalidAPIKey
	}

	return &Result{
		APIKey:  result.apiKey,
		OrgID:   result.orgID,
		OrgName: result.orgName,
	}, nil
}

// callbackResult holds the result from the callback handler.
type callbackResult struct {
	apiKey  string
	orgID   string
	orgName string
	err     error
}

// makeCallbackHandler creates an HTTP handler for the OAuth-like callback.
func makeCallbackHandler(resultCh chan<- callbackResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		debugf("Callback received: %s %s", r.Method, r.URL.String())
		debugf("Callback RemoteAddr: %s", r.RemoteAddr)
		debugf("Callback Headers: %v", r.Header)

		// Security: only accept requests from localhost
		remoteIP := extractIP(r.RemoteAddr)
		debugf("Extracted remote IP: %s", remoteIP)

		if !IsLocalhost(remoteIP) {
			debugf("Rejected non-localhost request from: %s", remoteIP)
			http.Error(w, "Forbidden: requests must come from localhost", http.StatusForbidden)
			resultCh <- callbackResult{err: fmt.Errorf("%w: %s", ErrForbidden, remoteIP)}
			return
		}

		// Extract API key from query parameters
		apiKey := r.URL.Query().Get("api_key")
		debugf("Received api_key: %s (length: %d)", maskAPIKey(apiKey), len(apiKey))

		if apiKey == "" {
			debugf("Missing api_key parameter")
			debugf("All query params: %v", r.URL.Query())
			http.Error(w, "Missing api_key parameter", http.StatusBadRequest)
			resultCh <- callbackResult{err: ErrMissingAPIKey}
			return
		}

		// Extract optional org info
		orgID := r.URL.Query().Get("org_id")
		orgName := r.URL.Query().Get("org_name")
		debugf("Received org_id: %s", orgID)
		debugf("Received org_name: %s", orgName)

		// Send success response to browser
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, successHTML)

		debugf("Callback successful, sending result to channel")

		// Send result to channel
		resultCh <- callbackResult{
			apiKey:  apiKey,
			orgID:   orgID,
			orgName: orgName,
		}
	}
}

// maskAPIKey masks an API key for safe logging, showing only prefix and suffix.
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// successHTML is the HTML response shown in the browser after successful login.
const successHTML = `<!DOCTYPE html>
<html>
<head>
    <title>StackEye CLI - Login Successful</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
               display: flex; justify-content: center; align-items: center; height: 100vh;
               margin: 0; background: #f5f5f5; }
        .container { text-align: center; padding: 2rem; background: white;
                     border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #10b981; margin-bottom: 0.5rem; }
        p { color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Login Successful</h1>
        <p>You can close this window and return to your terminal.</p>
    </div>
</body>
</html>`

// shutdownServer gracefully shuts down the HTTP server.
func shutdownServer(server *http.Server, wg *sync.WaitGroup) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	wg.Wait()
}

// extractIP extracts the IP address from a RemoteAddr string.
// RemoteAddr may be "IP:port" or just "IP" for certain configurations.
func extractIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// No port, return as-is
		return remoteAddr
	}
	return host
}
