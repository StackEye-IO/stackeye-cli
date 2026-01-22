// Package testutil provides utilities for end-to-end CLI testing.
package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
)

// APIResponse represents a standardized API response.
type APIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
	Meta   *APIMeta    `json:"meta,omitempty"`
}

// APIMeta contains pagination metadata.
type APIMeta struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	Pages   int  `json:"pages"`
	HasNext bool `json:"has_next"`
}

// RouteHandler is a function that handles an API route.
type RouteHandler func(w http.ResponseWriter, r *http.Request, matches []string)

// Route represents a registered API route.
type Route struct {
	Method  string
	Pattern *regexp.Regexp
	Handler RouteHandler
}

// MockServer provides a mock HTTP server for E2E testing.
type MockServer struct {
	Server  *httptest.Server
	BaseURL string
	routes  []Route
	mu      sync.RWMutex
	calls   []RecordedCall
}

// RecordedCall records details of an API call made to the mock server.
type RecordedCall struct {
	Method string
	Path   string
	Body   string
}

// NewMockServer creates a new mock API server with default routes.
func NewMockServer() *MockServer {
	ms := &MockServer{
		routes: []Route{},
		calls:  []RecordedCall{},
	}

	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handleRequest))
	ms.BaseURL = ms.Server.URL

	return ms
}

// handleRequest routes incoming requests to registered handlers.
func (ms *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	ms.mu.Lock()
	ms.calls = append(ms.calls, RecordedCall{
		Method: r.Method,
		Path:   r.URL.Path,
	})
	ms.mu.Unlock()

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for _, route := range ms.routes {
		if r.Method != route.Method {
			continue
		}
		if matches := route.Pattern.FindStringSubmatch(r.URL.Path); matches != nil {
			route.Handler(w, r, matches)
			return
		}
	}

	// Default 404 response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Status: "error",
		Error:  "endpoint not found",
	})
}

// Register adds a new route handler to the mock server.
func (ms *MockServer) Register(method, pattern string, handler RouteHandler) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.routes = append(ms.routes, Route{
		Method:  method,
		Pattern: regexp.MustCompile("^" + pattern + "$"),
		Handler: handler,
	})
}

// RegisterDefaultRoutes registers common API routes with fixture responses.
// Note: SDK uses paths like "/v1/probes" (without /api prefix).
// SDK directly unmarshals responses into typed structs, not wrapped APIResponse.
func (ms *MockServer) RegisterDefaultRoutes() {
	// Probes - SDK expects ProbeListResponse: {probes: [], total, page, limit}
	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"probes": ProbeListFixture(),
			"total":  2,
			"page":   1,
			"limit":  20,
		})
	})

	// Individual probe GET - SDK expects Probe struct directly
	ms.Register("GET", `/v1/probes/([a-f0-9-]+)`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		probeID := matches[1]
		probe := ProbeFixture(probeID)
		if probe == nil {
			RespondWithError(w, http.StatusNotFound, "probe not found")
			return
		}
		RespondWithJSON(w, http.StatusOK, probe)
	})

	// Probe create - SDK expects Probe struct directly
	ms.Register("POST", "/v1/probes", func(w http.ResponseWriter, r *http.Request, _ []string) {
		probe := NewProbeFixture()
		RespondWithJSON(w, http.StatusCreated, probe)
	})

	// Probe update - SDK expects Probe struct directly
	ms.Register("PUT", `/v1/probes/([a-f0-9-]+)`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		probeID := matches[1]
		probe := ProbeFixture(probeID)
		if probe == nil {
			RespondWithError(w, http.StatusNotFound, "probe not found")
			return
		}
		RespondWithJSON(w, http.StatusOK, probe)
	})

	ms.Register("DELETE", `/v1/probes/([a-f0-9-]+)`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		probeID := matches[1]
		probe := ProbeFixture(probeID)
		if probe == nil {
			RespondWithError(w, http.StatusNotFound, "probe not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Probe pause - SDK expects Probe struct directly
	ms.Register("POST", `/v1/probes/([a-f0-9-]+)/pause`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		probeID := matches[1]
		probe := ProbeFixture(probeID)
		if probe == nil {
			RespondWithError(w, http.StatusNotFound, "probe not found")
			return
		}
		probe["status"] = "paused"
		probe["is_paused"] = true
		RespondWithJSON(w, http.StatusOK, probe)
	})

	// Probe resume - SDK expects Probe struct directly
	ms.Register("POST", `/v1/probes/([a-f0-9-]+)/resume`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		probeID := matches[1]
		probe := ProbeFixture(probeID)
		if probe == nil {
			RespondWithError(w, http.StatusNotFound, "probe not found")
			return
		}
		probe["status"] = "up"
		probe["is_paused"] = false
		RespondWithJSON(w, http.StatusOK, probe)
	})

	// Probe test (by ID) - SDK expects ProbeTestResponse directly
	ms.Register("POST", `/v1/probes/([a-f0-9-]+)/test`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		RespondWithJSON(w, http.StatusOK, ProbeTestResultFixture())
	})

	// Probe test (ad-hoc) - Tests probe config without requiring a saved probe
	// CLI uses this: fetches probe, converts to test request, calls /v1/probes/test
	ms.Register("POST", `/v1/probes/test`, func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, ProbeTestResultFixture())
	})

	// Alerts - SDK expects AlertListResponse: {alerts: [], total}
	ms.Register("GET", "/v1/alerts", func(w http.ResponseWriter, r *http.Request, _ []string) {
		alerts := AlertListFixture()
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"alerts": alerts,
			"total":  len(alerts),
		})
	})

	// Alert GET - SDK expects {alert: Alert}
	ms.Register("GET", `/v1/alerts/([a-f0-9-]+)`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		alertID := matches[1]
		alert := AlertFixture(alertID)
		if alert == nil {
			RespondWithError(w, http.StatusNotFound, "alert not found")
			return
		}
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"alert": alert,
		})
	})

	// Alert acknowledge - SDK uses PUT and expects {status: string, alert: Alert}
	ms.Register("PUT", `/v1/alerts/([a-f0-9-]+)/acknowledge`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		alertID := matches[1]
		alert := AlertFixture(alertID)
		if alert == nil {
			RespondWithError(w, http.StatusNotFound, "alert not found")
			return
		}
		alert["status"] = "acknowledged"
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status": "acknowledged",
			"alert":  alert,
		})
	})

	// Alert resolve - SDK expects {status: string, alert: Alert}
	ms.Register("POST", `/v1/alerts/([a-f0-9-]+)/resolve`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		alertID := matches[1]
		alert := AlertFixture(alertID)
		if alert == nil {
			RespondWithError(w, http.StatusNotFound, "alert not found")
			return
		}
		alert["status"] = "resolved"
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status": "resolved",
			"alert":  alert,
		})
	})

	// Channels - SDK expects ChannelListResponse: {channels: [], total}
	ms.Register("GET", "/v1/channels", func(w http.ResponseWriter, r *http.Request, _ []string) {
		channels := ChannelListFixture()
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"channels": channels,
			"total":    len(channels),
		})
	})

	// Channel GET - SDK expects {channel: Channel}
	ms.Register("GET", `/v1/channels/([a-f0-9-]+)`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		channelID := matches[1]
		channel := ChannelFixture(channelID)
		if channel == nil {
			RespondWithError(w, http.StatusNotFound, "channel not found")
			return
		}
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"channel": channel,
		})
	})

	// Channel create - SDK expects {channel: Channel}
	ms.Register("POST", "/v1/channels", func(w http.ResponseWriter, r *http.Request, _ []string) {
		channel := NewChannelFixture()
		RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"channel": channel,
		})
	})

	// Channel delete
	ms.Register("DELETE", `/v1/channels/([a-f0-9-]+)`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		channelID := matches[1]
		channel := ChannelFixture(channelID)
		if channel == nil {
			RespondWithError(w, http.StatusNotFound, "channel not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Channel test - SDK expects ChannelTestResponse directly
	ms.Register("POST", `/v1/channels/([a-f0-9-]+)/test`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"success":          true,
			"message":          "Test notification sent successfully",
			"response_time_ms": 150,
		})
	})

	// Regions - SDK expects RegionListResponse: {status: "success", data: map[continent][]Region}
	ms.Register("GET", "/v1/regions", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   RegionListByContinent(),
		})
	})
}

// Close shuts down the mock server.
func (ms *MockServer) Close() {
	ms.Server.Close()
}

// GetCalls returns all recorded API calls.
func (ms *MockServer) GetCalls() []RecordedCall {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return append([]RecordedCall{}, ms.calls...)
}

// ClearCalls clears the recorded API calls.
func (ms *MockServer) ClearCalls() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.calls = []RecordedCall{}
}

// HasCall checks if a specific API call was made.
func (ms *MockServer) HasCall(method, pathPrefix string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	for _, call := range ms.calls {
		if call.Method == method && strings.HasPrefix(call.Path, pathPrefix) {
			return true
		}
	}
	return false
}

// RespondWithJSON writes a JSON response.
func RespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// RespondWithError writes an error response.
func RespondWithError(w http.ResponseWriter, status int, message string) {
	RespondWithJSON(w, status, APIResponse{
		Status: "error",
		Error:  message,
	})
}
