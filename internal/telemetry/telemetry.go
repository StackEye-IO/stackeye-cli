// Package telemetry provides opt-in anonymous usage analytics for the StackEye CLI.
//
// Telemetry is disabled by default and requires explicit user consent.
// When enabled, it collects anonymous usage data to help improve the CLI:
// - Command names and exit codes
// - Execution duration
// - CLI version and platform info
// - Anonymized organization ID (SHA-256 hash prefix)
//
// NO personal data, API keys, error messages, or identifiable information is collected.
//
// Environment variable STACKEYE_TELEMETRY=0 disables telemetry regardless of config.
package telemetry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/config"
	"github.com/StackEye-IO/stackeye-cli/internal/version"
)

// EnvTelemetry is the environment variable to override telemetry settings.
// Set to "0" or "false" to disable telemetry regardless of config.
const EnvTelemetry = "STACKEYE_TELEMETRY"

// DefaultEndpoint is the telemetry API endpoint.
const DefaultEndpoint = "https://api.stackeye.io/v1/telemetry/cli"

// Event represents a telemetry event to be sent to the backend.
type Event struct {
	// CLIVersion is the version of the CLI (e.g., "1.2.3").
	CLIVersion string `json:"cli_version"`

	// Command is the command name (e.g., "probe create").
	Command string `json:"command"`

	// ExitCode is the command's exit code (0 for success).
	ExitCode int `json:"exit_code"`

	// DurationMs is the execution time in milliseconds.
	DurationMs int64 `json:"duration_ms,omitempty"`

	// OS is the operating system (e.g., "linux", "darwin", "windows").
	OS string `json:"os"`

	// Arch is the architecture (e.g., "amd64", "arm64").
	Arch string `json:"arch"`

	// OrgIDHash is the first 16 characters of the SHA-256 hash of the org ID.
	// Empty if not authenticated.
	OrgIDHash string `json:"org_id_hash,omitempty"`
}

// Client handles telemetry event collection and transmission.
type Client struct {
	mu       sync.Mutex
	endpoint string
	enabled  bool
	http     *http.Client

	// lastEvent stores the most recent event for debugging
	lastEvent *Event

	// wg tracks pending async sends for graceful shutdown
	wg sync.WaitGroup
}

var (
	// globalClient is the singleton telemetry client
	globalClient *Client
	clientOnce   sync.Once
)

// GetClient returns the global telemetry client.
// The client is initialized once and reused across the CLI lifetime.
func GetClient() *Client {
	clientOnce.Do(func() {
		globalClient = &Client{
			endpoint: DefaultEndpoint,
			http: &http.Client{
				Timeout: 5 * time.Second,
			},
		}
		globalClient.loadConfig()
	})
	return globalClient
}

// ResetClient resets the global client for testing.
func ResetClient() {
	clientOnce = sync.Once{}
	globalClient = nil
}

// loadConfig loads telemetry settings from config and environment.
// Must be called with mutex held or during initialization.
func (c *Client) loadConfig() {
	// Environment variable takes precedence
	if envVal := os.Getenv(EnvTelemetry); envVal != "" {
		envVal = strings.ToLower(envVal)
		if envVal == "0" || envVal == "false" || envVal == "no" || envVal == "off" {
			c.enabled = false
			return
		}
		if envVal == "1" || envVal == "true" || envVal == "yes" || envVal == "on" {
			c.enabled = true
			return
		}
	}

	// Load from config file
	cfg, err := config.Load()
	if err != nil {
		c.enabled = false
		return
	}

	if cfg.Preferences != nil {
		c.enabled = cfg.Preferences.TelemetryEnabled
	}
}

// loadConfigLocked loads config while holding the mutex.
func (c *Client) loadConfigLocked() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadConfig()
}

// IsEnabled returns whether telemetry is currently enabled.
func (c *Client) IsEnabled() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.enabled
}

// SetEnabled enables or disables telemetry.
// This does NOT persist the setting - use config.Save() for that.
func (c *Client) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
}

// Track records a telemetry event.
// If telemetry is disabled, this is a no-op.
// Events are sent asynchronously to avoid blocking the CLI.
// Call Flush() before exit to ensure events are sent.
func (c *Client) Track(ctx context.Context, command string, exitCode int, duration time.Duration) {
	c.mu.Lock()
	enabled := c.enabled
	c.mu.Unlock()

	if !enabled {
		return
	}

	event := c.buildEvent(command, exitCode, duration)

	c.mu.Lock()
	c.lastEvent = event
	c.mu.Unlock()

	// Send asynchronously to not block CLI commands
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		sendCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.send(sendCtx, event)
	}()
}

// Flush waits for all pending telemetry events to be sent.
// Call this before exiting the CLI to ensure events are not lost.
// Returns after all pending sends complete or timeout expires.
func (c *Client) Flush(timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All sends completed
	case <-time.After(timeout):
		// Timeout - some events may be lost
	}
}

// buildEvent creates a telemetry event with current context.
func (c *Client) buildEvent(command string, exitCode int, duration time.Duration) *Event {
	event := &Event{
		CLIVersion: version.Version,
		Command:    sanitizeCommand(command),
		ExitCode:   exitCode,
		DurationMs: duration.Milliseconds(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
	}

	// Add anonymized org ID if authenticated
	cfg, err := config.Load()
	if err == nil {
		ctx, err := cfg.GetCurrentContext()
		if err == nil && ctx != nil && ctx.OrganizationID != "" {
			event.OrgIDHash = hashOrgID(ctx.OrganizationID)
		}
	}

	return event
}

// sanitizeCommand extracts just the command name without arguments.
// For example, "probe create --name test" becomes "probe create".
func sanitizeCommand(cmd string) string {
	// Split by common argument indicators
	parts := strings.Split(cmd, " ")
	var sanitized []string
	for _, part := range parts {
		// Stop at flags or values that look like arguments
		if strings.HasPrefix(part, "-") || strings.HasPrefix(part, "/") {
			break
		}
		// Skip empty parts
		if part == "" {
			continue
		}
		sanitized = append(sanitized, part)
	}
	if len(sanitized) == 0 {
		return "unknown"
	}
	return strings.Join(sanitized, " ")
}

// hashOrgID returns the first 16 characters of the SHA-256 hash of the org ID.
func hashOrgID(orgID string) string {
	hash := sha256.Sum256([]byte(orgID))
	return hex.EncodeToString(hash[:])[:16]
}

// send transmits an event to the telemetry endpoint.
func (c *Client) send(ctx context.Context, event *Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "stackeye-cli/"+version.Version)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetLastEvent returns the most recent event for debugging purposes.
// Returns nil if no event has been tracked.
func (c *Client) GetLastEvent() *Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastEvent
}

// SetEndpoint overrides the telemetry endpoint (for testing).
func (c *Client) SetEndpoint(endpoint string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.endpoint = endpoint
}

// Reload refreshes telemetry settings from config.
// Call this after changing telemetry preferences.
func (c *Client) Reload() {
	c.loadConfigLocked()
}
