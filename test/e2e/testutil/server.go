// Package testutil provides utilities for end-to-end CLI testing.
package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"time"
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

// ErrorConfig configures an error response for a specific route.
type ErrorConfig struct {
	StatusCode int
	Message    string
}

// MockServer provides a mock HTTP server for E2E testing.
type MockServer struct {
	Server  *httptest.Server
	BaseURL string
	routes  []Route
	mu      sync.RWMutex
	calls   []RecordedCall

	// latency adds a delay before responding to simulate network latency.
	latency time.Duration
	// routeLatencies maps "METHOD /path" to per-route latency overrides.
	routeLatencies map[string]time.Duration
	// routeErrors maps "METHOD /path-pattern" to forced error responses.
	routeErrors map[string]ErrorConfig
	// dynamicResponses maps "METHOD /path-pattern" to custom response data.
	dynamicResponses map[string]interface{}
}

// RecordedCall records details of an API call made to the mock server.
type RecordedCall struct {
	Method string
	Path   string
	Body   string
	Query  string
}

// NewMockServer creates a new mock API server with default routes.
func NewMockServer() *MockServer {
	ms := &MockServer{
		routes:           []Route{},
		calls:            []RecordedCall{},
		routeLatencies:   make(map[string]time.Duration),
		routeErrors:      make(map[string]ErrorConfig),
		dynamicResponses: make(map[string]interface{}),
	}

	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handleRequest))
	ms.BaseURL = ms.Server.URL

	return ms
}

// handleRequest routes incoming requests to registered handlers.
func (ms *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Read request body for recording
	var bodyStr string
	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyStr = string(bodyBytes)
		r.Body.Close()
		// Restore body so handlers can read it
		r.Body = io.NopCloser(strings.NewReader(bodyStr))
	}

	ms.mu.Lock()
	ms.calls = append(ms.calls, RecordedCall{
		Method: r.Method,
		Path:   r.URL.Path,
		Body:   bodyStr,
		Query:  r.URL.RawQuery,
	})
	ms.mu.Unlock()

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	routeKey := r.Method + " " + r.URL.Path

	// Apply global or per-route latency
	if delay, ok := ms.routeLatencies[routeKey]; ok {
		time.Sleep(delay)
	} else if ms.latency > 0 {
		time.Sleep(ms.latency)
	}

	// Check for forced error responses
	if errCfg, ok := ms.routeErrors[routeKey]; ok {
		RespondWithError(w, errCfg.StatusCode, errCfg.Message)
		return
	}

	// Check for dynamic response override
	if resp, ok := ms.dynamicResponses[routeKey]; ok {
		RespondWithJSON(w, http.StatusOK, resp)
		return
	}

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
		Error:  "endpoint not found: " + r.Method + " " + r.URL.Path,
	})
}

// Register adds a new route handler to the mock server.
//
// # Route Pattern Behavior (Task #8216)
//
// Routes use regex patterns to match URL paths. The common pattern `[a-f0-9-]+`
// is used to match UUIDs (e.g., "550e8400-e29b-41d4-a716-446655440000").
//
// WARNING: This pattern can match unintended paths that contain only hex
// characters and hyphens. English words that would falsely match include:
//
//   - "feed", "dead", "cafe", "bead", "beef", "face", "fade", "deaf"
//   - "decade", "facade", "effaced", "defaced"
//
// If you add a route like "/v1/probes/feed" (hypothetical probe aggregation
// feed endpoint), it could be matched by `/v1/probes/([a-f0-9-]+)` instead.
//
// MITIGATION: When adding new routes with potential hex-only names:
//  1. Register specific literal routes BEFORE wildcard UUID routes, as the
//     mock server uses first-match semantics
//  2. Or use more restrictive patterns that require the full UUID structure:
//     `[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`
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

// RegisterAllRoutes registers all API routes including extended endpoints
// for team, billing, status pages, alert stats/history, probe history/stats,
// API keys, dashboard, alert mutes, labels, and probe dependencies.
func (ms *MockServer) RegisterAllRoutes() {
	ms.RegisterDefaultRoutes()

	// --- CLI Auth Verify ---
	ms.Register("GET", "/v1/cli-auth/verify", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, CLIVerifyFixture())
	})

	// --- Alert Stats ---
	ms.Register("GET", "/v1/alerts/stats", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, AlertStatsFixture())
	})

	// --- Alert Timeline ---
	ms.Register("GET", `/v1/alerts/([a-f0-9-]+)/timeline`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"timeline": AlertHistoryFixture(),
		})
	})

	// --- Alert Mutes ---
	ms.Register("GET", "/v1/alerts/mutes", func(w http.ResponseWriter, r *http.Request, _ []string) {
		mutes := AlertMuteListFixture()
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"mutes": mutes,
			"total": len(mutes),
		})
	})

	ms.Register("POST", "/v1/alerts/mutes", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"mute": NewAlertMuteFixture(),
		})
	})

	ms.Register("DELETE", `/v1/alerts/mutes/([a-f0-9-]+)`, func(w http.ResponseWriter, _ *http.Request, _ []string) {
		w.WriteHeader(http.StatusNoContent)
	})

	ms.Register("POST", `/v1/alerts/mutes/([a-f0-9-]+)/expire`, func(w http.ResponseWriter, _ *http.Request, _ []string) {
		mute := NewAlertMuteFixture()
		mute["status"] = "expired"
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"mute": mute,
		})
	})

	// --- Probe History/Results ---
	ms.Register("GET", `/v1/probes/([a-f0-9-]+)/results`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		results := ProbeHistoryFixture()
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"results": results,
			"total":   len(results),
			"page":    1,
			"limit":   20,
		})
	})

	// --- Probe Stats ---
	ms.Register("GET", `/v1/probes/([a-f0-9-]+)/stats`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		RespondWithJSON(w, http.StatusOK, ProbeStatsFixture())
	})

	// --- Probe Labels ---
	ms.Register("GET", `/v1/probes/([a-f0-9-]+)/labels`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"labels": LabelListFixture(),
		})
	})

	ms.Register("POST", `/v1/probes/([a-f0-9-]+)/labels`, func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"label": NewLabelFixture(),
		})
	})

	ms.Register("DELETE", `/v1/probes/([a-f0-9-]+)/labels/([a-f0-9-]+)`, func(w http.ResponseWriter, _ *http.Request, _ []string) {
		w.WriteHeader(http.StatusNoContent)
	})

	// --- Label Keys ---
	ms.Register("GET", "/v1/label-keys", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"label_keys": LabelKeyListFixture(),
		})
	})

	ms.Register("POST", "/v1/label-keys", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"label_key": NewLabelKeyFixture(),
		})
	})

	ms.Register("DELETE", `/v1/label-keys/([a-f0-9-]+)`, func(w http.ResponseWriter, _ *http.Request, _ []string) {
		w.WriteHeader(http.StatusNoContent)
	})

	// --- Probe Dependencies ---
	ms.Register("GET", `/v1/probes/([a-f0-9-]+)/dependencies`, func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"dependencies": ProbeDependencyListFixture(),
		})
	})

	ms.Register("POST", `/v1/probes/([a-f0-9-]+)/dependencies`, func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"dependency": NewProbeDependencyFixture(),
		})
	})

	ms.Register("DELETE", `/v1/probes/([a-f0-9-]+)/dependencies/([a-f0-9-]+)`, func(w http.ResponseWriter, _ *http.Request, _ []string) {
		w.WriteHeader(http.StatusNoContent)
	})

	// --- Dependency Tree ---
	ms.Register("GET", `/v1/organizations/([a-f0-9-]+)/dependency-tree`, func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"tree": DependencyTreeFixture(),
		})
	})

	// --- Team Members ---
	ms.Register("GET", "/v1/team/members", func(w http.ResponseWriter, r *http.Request, _ []string) {
		members := TeamMemberListFixture()
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"members": members,
			"total":   len(members),
		})
	})

	ms.Register("POST", "/v1/team/members", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"member": NewTeamMemberFixture(),
		})
	})

	ms.Register("PUT", `/v1/team/members/([a-f0-9-]+)`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		member := NewTeamMemberFixture()
		member["id"] = matches[1]
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"member": member,
		})
	})

	ms.Register("DELETE", `/v1/team/members/([a-f0-9-]+)`, func(w http.ResponseWriter, _ *http.Request, _ []string) {
		w.WriteHeader(http.StatusNoContent)
	})

	// --- Team Invitations ---
	ms.Register("GET", "/v1/team/invitations", func(w http.ResponseWriter, r *http.Request, _ []string) {
		invitations := TeamInvitationListFixture()
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"invitations": invitations,
			"total":       len(invitations),
		})
	})

	ms.Register("DELETE", `/v1/team/invitations/([a-f0-9-]+)`, func(w http.ResponseWriter, _ *http.Request, _ []string) {
		w.WriteHeader(http.StatusNoContent)
	})

	// --- Status Pages ---
	ms.Register("GET", "/v1/status-pages", func(w http.ResponseWriter, r *http.Request, _ []string) {
		pages := StatusPageListFixture()
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status_pages": pages,
			"total":        len(pages),
		})
	})

	ms.Register("GET", `/v1/status-pages/([a-f0-9-]+)`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		page := StatusPageFixture(matches[1])
		if page == nil {
			RespondWithError(w, http.StatusNotFound, "status page not found")
			return
		}
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status_page": page,
		})
	})

	ms.Register("POST", "/v1/status-pages", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"status_page": NewStatusPageFixture(),
		})
	})

	ms.Register("PUT", `/v1/status-pages/([a-f0-9-]+)`, func(w http.ResponseWriter, r *http.Request, matches []string) {
		page := StatusPageFixture(matches[1])
		if page == nil {
			RespondWithError(w, http.StatusNotFound, "status page not found")
			return
		}
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status_page": page,
		})
	})

	ms.Register("DELETE", `/v1/status-pages/([a-f0-9-]+)`, func(w http.ResponseWriter, _ *http.Request, _ []string) {
		w.WriteHeader(http.StatusNoContent)
	})

	// --- Status Page Incidents ---
	ms.Register("GET", `/v1/status-pages/([a-f0-9-]+)/incidents`, func(w http.ResponseWriter, r *http.Request, _ []string) {
		incidents := StatusPageIncidentListFixture()
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"incidents": incidents,
			"total":     len(incidents),
		})
	})

	ms.Register("POST", `/v1/status-pages/([a-f0-9-]+)/incidents`, func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"incident": NewStatusPageIncidentFixture(),
		})
	})

	ms.Register("POST", `/v1/status-pages/([a-f0-9-]+)/incidents/([a-f0-9-]+)/resolve`, func(w http.ResponseWriter, _ *http.Request, _ []string) {
		incident := NewStatusPageIncidentFixture()
		incident["status"] = "resolved"
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"incident": incident,
		})
	})

	// --- Dashboard ---
	ms.Register("GET", "/v1/dashboard", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, DashboardStatsFixture())
	})

	// --- Billing ---
	ms.Register("GET", "/v1/billing", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, BillingInfoFixture())
	})

	ms.Register("GET", "/v1/billing/status", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, BillingStatusFixture())
	})

	ms.Register("GET", "/v1/billing/usage", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, BillingUsageFixture())
	})

	ms.Register("GET", "/v1/billing/usage/history", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"usage_history": []map[string]interface{}{BillingUsageFixture()},
		})
	})

	ms.Register("POST", "/v1/billing/checkout", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"checkout_url": "https://checkout.stripe.com/test_session",
		})
	})

	ms.Register("POST", "/v1/billing/portal", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"portal_url": "https://billing.stripe.com/test_portal",
		})
	})

	// --- API Keys ---
	ms.Register("GET", "/v1/api-keys", func(w http.ResponseWriter, r *http.Request, _ []string) {
		keys := APIKeyListFixture()
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"api_keys": keys,
			"total":    len(keys),
		})
	})

	ms.Register("POST", "/v1/api-keys", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"api_key": NewAPIKeyFixture(),
		})
	})

	ms.Register("DELETE", `/v1/api-keys/([a-f0-9-]+)`, func(w http.ResponseWriter, _ *http.Request, _ []string) {
		w.WriteHeader(http.StatusNoContent)
	})

	// --- User ---
	ms.Register("GET", "/v1/user/me", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, UserProfileFixture())
	})

	ms.Register("PUT", "/v1/user/me", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, UserProfileFixture())
	})

	ms.Register("GET", "/v1/user/organizations", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"organizations": UserOrganizationListFixture(),
		})
	})

	ms.Register("POST", "/v1/user/switch-organization", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status": "success",
		})
	})
}

// Close shuts down the mock server.
func (ms *MockServer) Close() {
	ms.Server.Close()
}

// WithLatency sets a global latency for all responses.
func (ms *MockServer) WithLatency(d time.Duration) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.latency = d
}

// WithRouteLatency sets latency for a specific route (e.g., "GET /v1/probes").
func (ms *MockServer) WithRouteLatency(method, path string, d time.Duration) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.routeLatencies[method+" "+path] = d
}

// WithError forces a specific route to return an error response.
func (ms *MockServer) WithError(method, path string, statusCode int, message string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.routeErrors[method+" "+path] = ErrorConfig{
		StatusCode: statusCode,
		Message:    message,
	}
}

// ClearError removes a forced error for a specific route.
func (ms *MockServer) ClearError(method, path string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.routeErrors, method+" "+path)
}

// ClearAllErrors removes all forced errors.
func (ms *MockServer) ClearAllErrors() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.routeErrors = make(map[string]ErrorConfig)
}

// SetDynamicResponse overrides the response for a specific route with custom data.
func (ms *MockServer) SetDynamicResponse(method, path string, data interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.dynamicResponses[method+" "+path] = data
}

// ClearDynamicResponse removes a dynamic response override.
func (ms *MockServer) ClearDynamicResponse(method, path string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.dynamicResponses, method+" "+path)
}

// ClearAllDynamicResponses removes all dynamic response overrides.
func (ms *MockServer) ClearAllDynamicResponses() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.dynamicResponses = make(map[string]interface{})
}

// GetCalls returns all recorded API calls.
func (ms *MockServer) GetCalls() []RecordedCall {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return append([]RecordedCall{}, ms.calls...)
}

// GetCallsForPath returns recorded calls matching a path prefix.
func (ms *MockServer) GetCallsForPath(method, pathPrefix string) []RecordedCall {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	var matched []RecordedCall
	for _, call := range ms.calls {
		if call.Method == method && strings.HasPrefix(call.Path, pathPrefix) {
			matched = append(matched, call)
		}
	}
	return matched
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

// CallCount returns the number of calls matching the given method and path prefix.
func (ms *MockServer) CallCount(method, pathPrefix string) int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	count := 0
	for _, call := range ms.calls {
		if call.Method == method && strings.HasPrefix(call.Path, pathPrefix) {
			count++
		}
	}
	return count
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
