package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMockServer(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	assert.NotNil(t, ms.Server)
	assert.NotEmpty(t, ms.BaseURL)
	assert.Empty(t, ms.GetCalls())
}

func TestMockServer_RegisterAndHandle(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/test", func(w http.ResponseWriter, r *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{"hello": "world"})
	})

	resp, err := http.Get(ms.BaseURL + "/v1/test")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, "world", body["hello"])
}

func TestMockServer_404ForUnregisteredRoutes(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	resp, err := http.Get(ms.BaseURL + "/v1/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMockServer_CallRecording(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})

	_, err := http.Get(ms.BaseURL + "/v1/probes")
	require.NoError(t, err)

	calls := ms.GetCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, "GET", calls[0].Method)
	assert.Equal(t, "/v1/probes", calls[0].Path)
}

func TestMockServer_BodyRecording(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("POST", "/v1/probes", func(w http.ResponseWriter, r *http.Request, _ []string) {
		// Handler should be able to read body too
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "test-probe")
		RespondWithJSON(w, http.StatusCreated, map[string]string{})
	})

	payload := `{"name":"test-probe"}`
	_, err := http.Post(ms.BaseURL+"/v1/probes", "application/json", strings.NewReader(payload))
	require.NoError(t, err)

	calls := ms.GetCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, "POST", calls[0].Method)
	assert.Contains(t, calls[0].Body, "test-probe")
}

func TestMockServer_QueryRecording(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})

	_, err := http.Get(ms.BaseURL + "/v1/probes?status=up&page=2")
	require.NoError(t, err)

	calls := ms.GetCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, "status=up&page=2", calls[0].Query)
}

func TestMockServer_HasCall(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})

	assert.False(t, ms.HasCall("GET", "/v1/probes"))

	_, err := http.Get(ms.BaseURL + "/v1/probes")
	require.NoError(t, err)

	assert.True(t, ms.HasCall("GET", "/v1/probes"))
	assert.False(t, ms.HasCall("POST", "/v1/probes"))
}

func TestMockServer_CallCount(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})

	assert.Equal(t, 0, ms.CallCount("GET", "/v1/probes"))

	_, _ = http.Get(ms.BaseURL + "/v1/probes")
	_, _ = http.Get(ms.BaseURL + "/v1/probes")
	_, _ = http.Get(ms.BaseURL + "/v1/probes")

	assert.Equal(t, 3, ms.CallCount("GET", "/v1/probes"))
}

func TestMockServer_GetCallsForPath(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})
	ms.Register("GET", "/v1/alerts", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})

	_, _ = http.Get(ms.BaseURL + "/v1/probes")
	_, _ = http.Get(ms.BaseURL + "/v1/alerts")
	_, _ = http.Get(ms.BaseURL + "/v1/probes")

	probeCalls := ms.GetCallsForPath("GET", "/v1/probes")
	assert.Len(t, probeCalls, 2)

	alertCalls := ms.GetCallsForPath("GET", "/v1/alerts")
	assert.Len(t, alertCalls, 1)
}

func TestMockServer_ClearCalls(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})

	_, _ = http.Get(ms.BaseURL + "/v1/probes")
	assert.Len(t, ms.GetCalls(), 1)

	ms.ClearCalls()
	assert.Empty(t, ms.GetCalls())
}

func TestMockServer_WithError(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Force error
	ms.WithError("GET", "/v1/probes", http.StatusInternalServerError, "server error")

	resp, err := http.Get(ms.BaseURL + "/v1/probes")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var body APIResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, "server error", body.Error)
}

func TestMockServer_ClearError(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	ms.WithError("GET", "/v1/probes", http.StatusInternalServerError, "server error")

	resp, err := http.Get(ms.BaseURL + "/v1/probes")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// Clear the error
	ms.ClearError("GET", "/v1/probes")

	resp, err = http.Get(ms.BaseURL + "/v1/probes")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMockServer_ClearAllErrors(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})
	ms.Register("GET", "/v1/alerts", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})

	ms.WithError("GET", "/v1/probes", 500, "error1")
	ms.WithError("GET", "/v1/alerts", 503, "error2")

	ms.ClearAllErrors()

	resp, err := http.Get(ms.BaseURL + "/v1/probes")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(ms.BaseURL + "/v1/alerts")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMockServer_SetDynamicResponse(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{"probes": []string{}, "total": 0})
	})

	// Override with dynamic response
	customData := map[string]interface{}{
		"probes": []map[string]string{{"name": "custom-probe"}},
		"total":  1,
	}
	ms.SetDynamicResponse("GET", "/v1/probes", customData)

	resp, err := http.Get(ms.BaseURL + "/v1/probes")
	require.NoError(t, err)
	defer resp.Body.Close()

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, float64(1), body["total"])
}

func TestMockServer_ClearDynamicResponse(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{"total": 0})
	})

	ms.SetDynamicResponse("GET", "/v1/probes", map[string]interface{}{"total": 99})

	// Dynamic response active
	resp, _ := http.Get(ms.BaseURL + "/v1/probes")
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	assert.Equal(t, float64(99), body["total"])

	// Clear dynamic response
	ms.ClearDynamicResponse("GET", "/v1/probes")

	resp, _ = http.Get(ms.BaseURL + "/v1/probes")
	json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	assert.Equal(t, float64(0), body["total"])
}

func TestMockServer_WithLatency(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})

	// Use 100ms latency for better test reliability on CI systems.
	// The 5ms buffer accounts for timing jitter in time.Sleep() across platforms.
	const configuredLatency = 100 * time.Millisecond
	const jitterBuffer = 5 * time.Millisecond

	ms.WithLatency(configuredLatency)

	start := time.Now()
	_, err := http.Get(ms.BaseURL + "/v1/probes")
	elapsed := time.Since(start)
	require.NoError(t, err)

	minExpected := configuredLatency - jitterBuffer
	assert.True(t, elapsed >= minExpected, "expected at least %v latency, got %v", minExpected, elapsed)
}

func TestMockServer_WithRouteLatency(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})
	ms.Register("GET", "/v1/alerts", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})

	// Use 100ms for per-route latency testing - provides clear separation from baseline.
	// Thresholds account for CI system variability:
	// - 5ms jitter buffer for the latency route (time.Sleep() precision varies)
	// - 50ms ceiling for no-latency routes (generous allowance for CI contention)
	const configuredLatency = 100 * time.Millisecond
	const jitterBuffer = 5 * time.Millisecond
	const noLatencyCeiling = 50 * time.Millisecond

	// Only /v1/probes gets latency
	ms.WithRouteLatency("GET", "/v1/probes", configuredLatency)

	start := time.Now()
	_, _ = http.Get(ms.BaseURL + "/v1/probes")
	probeElapsed := time.Since(start)

	start = time.Now()
	_, _ = http.Get(ms.BaseURL + "/v1/alerts")
	alertElapsed := time.Since(start)

	minExpected := configuredLatency - jitterBuffer
	assert.True(t, probeElapsed >= minExpected, "expected probe latency >= %v, got %v", minExpected, probeElapsed)
	assert.True(t, alertElapsed < noLatencyCeiling, "expected alert latency < %v (no latency configured), got %v", noLatencyCeiling, alertElapsed)
}

func TestMockServer_ErrorPrecedenceOverDynamic(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]string{})
	})

	// Set both dynamic response and error - error should take precedence
	ms.SetDynamicResponse("GET", "/v1/probes", map[string]string{"status": "dynamic"})
	ms.WithError("GET", "/v1/probes", http.StatusServiceUnavailable, "forced error")

	resp, err := http.Get(ms.BaseURL + "/v1/probes")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestMockServer_RegexRouteMatching(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	var capturedID string
	ms.Register("GET", `/v1/probes/([a-f0-9-]+)`, func(w http.ResponseWriter, _ *http.Request, matches []string) {
		capturedID = matches[1]
		RespondWithJSON(w, http.StatusOK, map[string]string{"id": matches[1]})
	})

	_, err := http.Get(ms.BaseURL + "/v1/probes/" + ProbeID1)
	require.NoError(t, err)

	assert.Equal(t, ProbeID1, capturedID)
}

func TestMockServer_RegisterDefaultRoutes(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()
	ms.RegisterDefaultRoutes()

	// Test that probes endpoint returns fixture data
	resp, err := http.Get(ms.BaseURL + "/v1/probes")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, float64(2), body["total"])
}

func TestMockServer_RegisterAllRoutes(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()
	ms.RegisterAllRoutes()

	// Verify all extended endpoints respond correctly
	tests := []struct {
		method     string
		path       string
		wantStatus int
	}{
		// Core endpoints (from RegisterDefaultRoutes)
		{"GET", "/v1/probes", http.StatusOK},
		{"GET", "/v1/alerts", http.StatusOK},
		{"GET", "/v1/channels", http.StatusOK},
		{"GET", "/v1/regions", http.StatusOK},

		// Alert stats/mutes
		{"GET", "/v1/alerts/stats", http.StatusOK},
		{"GET", "/v1/alerts/" + AlertID1 + "/timeline", http.StatusOK},
		{"GET", "/v1/alerts/mutes", http.StatusOK},

		// Probe results/stats/labels/dependencies
		{"GET", "/v1/probes/" + ProbeID1 + "/results", http.StatusOK},
		{"GET", "/v1/probes/" + ProbeID1 + "/stats", http.StatusOK},
		{"GET", "/v1/probes/" + ProbeID1 + "/labels", http.StatusOK},
		{"GET", "/v1/probes/" + ProbeID1 + "/dependencies", http.StatusOK},

		// Label keys
		{"GET", "/v1/label-keys", http.StatusOK},

		// Dependency tree
		{"GET", "/v1/organizations/" + OrganizationID + "/dependency-tree", http.StatusOK},

		// Team
		{"GET", "/v1/team/members", http.StatusOK},
		{"GET", "/v1/team/invitations", http.StatusOK},

		// Status pages
		{"GET", "/v1/status-pages", http.StatusOK},
		{"GET", "/v1/status-pages/" + StatusPageID1, http.StatusOK},
		{"GET", "/v1/status-pages/" + StatusPageID1 + "/incidents", http.StatusOK},

		// Dashboard
		{"GET", "/v1/dashboard", http.StatusOK},

		// Billing
		{"GET", "/v1/billing", http.StatusOK},
		{"GET", "/v1/billing/status", http.StatusOK},
		{"GET", "/v1/billing/usage", http.StatusOK},
		{"GET", "/v1/billing/usage/history", http.StatusOK},

		// API keys
		{"GET", "/v1/api-keys", http.StatusOK},

		// User
		{"GET", "/v1/user/me", http.StatusOK},
		{"GET", "/v1/user/organizations", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ms.BaseURL+tt.path, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode, "unexpected status for %s %s", tt.method, tt.path)
		})
	}
}

func TestMockServer_RegisterAllRoutes_Mutating(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()
	ms.RegisterAllRoutes()

	tests := []struct {
		method     string
		path       string
		wantStatus int
	}{
		// Alert mutes
		{"POST", "/v1/alerts/mutes", http.StatusCreated},
		{"DELETE", "/v1/alerts/mutes/" + MuteID1, http.StatusNoContent},
		{"POST", "/v1/alerts/mutes/" + MuteID1 + "/expire", http.StatusOK},

		// Probe labels
		{"POST", "/v1/probes/" + ProbeID1 + "/labels", http.StatusCreated},
		{"DELETE", "/v1/probes/" + ProbeID1 + "/labels/" + LabelID1, http.StatusNoContent},

		// Label keys
		{"POST", "/v1/label-keys", http.StatusCreated},
		{"DELETE", "/v1/label-keys/" + LabelKeyID1, http.StatusNoContent},

		// Probe dependencies
		{"POST", "/v1/probes/" + ProbeID1 + "/dependencies", http.StatusCreated},
		{"DELETE", "/v1/probes/" + ProbeID1 + "/dependencies/" + DependencyID1, http.StatusNoContent},

		// Team
		{"POST", "/v1/team/members", http.StatusCreated},
		{"PUT", "/v1/team/members/" + TeamMemberID1, http.StatusOK},
		{"DELETE", "/v1/team/members/" + TeamMemberID1, http.StatusNoContent},
		{"DELETE", "/v1/team/invitations/" + InvitationID1, http.StatusNoContent},

		// Status pages
		{"POST", "/v1/status-pages", http.StatusCreated},
		{"PUT", "/v1/status-pages/" + StatusPageID1, http.StatusOK},
		{"DELETE", "/v1/status-pages/" + StatusPageID1, http.StatusNoContent},
		{"POST", "/v1/status-pages/" + StatusPageID1 + "/incidents", http.StatusCreated},
		{"POST", "/v1/status-pages/" + StatusPageID1 + "/incidents/" + IncidentID1 + "/resolve", http.StatusOK},

		// Billing
		{"POST", "/v1/billing/checkout", http.StatusOK},
		{"POST", "/v1/billing/portal", http.StatusOK},

		// API keys
		{"POST", "/v1/api-keys", http.StatusCreated},
		{"DELETE", "/v1/api-keys/" + APIKeyID1, http.StatusNoContent},

		// User
		{"PUT", "/v1/user/me", http.StatusOK},
		{"POST", "/v1/user/switch-organization", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ms.BaseURL+tt.path, strings.NewReader("{}"))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode, "unexpected status for %s %s", tt.method, tt.path)
		})
	}
}

func TestMockServer_ClearAllDynamicResponses(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.Register("GET", "/v1/probes", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{"total": 0})
	})
	ms.Register("GET", "/v1/alerts", func(w http.ResponseWriter, _ *http.Request, _ []string) {
		RespondWithJSON(w, http.StatusOK, map[string]interface{}{"total": 0})
	})

	ms.SetDynamicResponse("GET", "/v1/probes", map[string]interface{}{"total": 99})
	ms.SetDynamicResponse("GET", "/v1/alerts", map[string]interface{}{"total": 88})

	ms.ClearAllDynamicResponses()

	resp, _ := http.Get(ms.BaseURL + "/v1/probes")
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	assert.Equal(t, float64(0), body["total"])

	resp, _ = http.Get(ms.BaseURL + "/v1/alerts")
	json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	assert.Equal(t, float64(0), body["total"])
}

// Test fixture data consistency
func TestFixtures_ProbeIDs(t *testing.T) {
	probes := ProbeListFixture()
	require.Len(t, probes, 2)
	assert.Equal(t, ProbeID1, probes[0]["id"])
	assert.Equal(t, ProbeID2, probes[1]["id"])
}

func TestFixtures_ProbeFixture(t *testing.T) {
	probe := ProbeFixture(ProbeID1)
	require.NotNil(t, probe)
	assert.Equal(t, "API Health Check", probe["name"])

	missing := ProbeFixture("nonexistent-id")
	assert.Nil(t, missing)
}

func TestFixtures_AlertStatsFixture(t *testing.T) {
	stats := AlertStatsFixture()
	assert.Equal(t, "24h", stats["period"])
	assert.Equal(t, 5, stats["total_alerts"])
}

func TestFixtures_StatusPageFixture(t *testing.T) {
	page := StatusPageFixture(StatusPageID1)
	require.NotNil(t, page)
	assert.Equal(t, "Main Status Page", page["name"])

	missing := StatusPageFixture("nonexistent-id")
	assert.Nil(t, missing)
}

func TestFixtures_TeamMemberList(t *testing.T) {
	members := TeamMemberListFixture()
	require.Len(t, members, 2)
	assert.Equal(t, "owner", members[0]["role"])
	assert.Equal(t, "member", members[1]["role"])
}

func TestFixtures_DashboardStats(t *testing.T) {
	stats := DashboardStatsFixture()
	assert.Equal(t, 10, stats["total_probes"])
	assert.Equal(t, 8, stats["probes_up"])
}
