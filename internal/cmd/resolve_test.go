package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient creates a test client pointing to the given server.
func newTestClient(t *testing.T, server *httptest.Server) *client.Client {
	t.Helper()
	c := client.New(
		"test-api-key",
		server.URL,
		client.WithHTTPClient(server.Client()),
	)
	return c
}

func TestResolveProbeID_ValidUUID(t *testing.T) {
	// When input is a valid UUID, it should return immediately without API calls
	expectedID := uuid.New()

	// Server should NOT be called for UUID resolution
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("API should not be called when resolving a valid UUID")
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	result, err := ResolveProbeID(ctx, c, expectedID.String())
	require.NoError(t, err)
	assert.Equal(t, expectedID, result)
}

func TestResolveProbeID_SingleMatch(t *testing.T) {
	// When searching by name with exactly one match, return that probe's ID
	probeID := uuid.New()
	probeName := "Production API"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/probes", r.URL.Path)
		assert.Equal(t, probeName, r.URL.Query().Get("search"))

		response := client.ProbeListResponse{
			Probes: []client.Probe{
				{ID: probeID, Name: probeName, Status: "up"},
			},
			Total: 1,
			Page:  1,
			Limit: 100,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	result, err := ResolveProbeID(ctx, c, probeName)
	require.NoError(t, err)
	assert.Equal(t, probeID, result)
}

func TestResolveProbeID_ExactMatchPreferred(t *testing.T) {
	// When multiple probes match, prefer the exact name match
	exactID := uuid.New()
	otherID := uuid.New()
	searchName := "api"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := client.ProbeListResponse{
			Probes: []client.Probe{
				{ID: otherID, Name: "api-gateway", Status: "up"},
				{ID: exactID, Name: "api", Status: "up"},
				{ID: uuid.New(), Name: "api-server", Status: "up"},
			},
			Total: 3,
			Page:  1,
			Limit: 100,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	result, err := ResolveProbeID(ctx, c, searchName)
	require.NoError(t, err)
	assert.Equal(t, exactID, result)
}

func TestResolveProbeID_NoMatch(t *testing.T) {
	// When no probes match, return a clear error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := client.ProbeListResponse{
			Probes: []client.Probe{},
			Total:  0,
			Page:   1,
			Limit:  100,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	_, err := ResolveProbeID(ctx, c, "nonexistent-probe")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestResolveProbeID_AmbiguousMatch(t *testing.T) {
	// When multiple probes match with no exact match, return ambiguous error with suggestions
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := client.ProbeListResponse{
			Probes: []client.Probe{
				{ID: uuid.New(), Name: "api-gateway", Status: "up"},
				{ID: uuid.New(), Name: "api-server", Status: "up"},
				{ID: uuid.New(), Name: "api-backend", Status: "up"},
			},
			Total: 3,
			Page:  1,
			Limit: 100,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	_, err := ResolveProbeID(ctx, c, "api")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
	assert.Contains(t, err.Error(), "3 matches")
	// Should suggest using UUID
	assert.Contains(t, err.Error(), "UUID")
}

func TestResolveProbeID_CaseInsensitiveExactMatch(t *testing.T) {
	// When multiple probes match case-insensitively, it should be ambiguous
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := client.ProbeListResponse{
			Probes: []client.Probe{
				{ID: uuid.New(), Name: "Production-API", Status: "up"},
				{ID: uuid.New(), Name: "production-api", Status: "up"},
			},
			Total: 2,
			Page:  1,
			Limit: 100,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	// Search for "PRODUCTION-API" matches both names case-insensitively
	// so it should be ambiguous (correct behavior)
	_, err := ResolveProbeID(ctx, c, "PRODUCTION-API")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
}

func TestResolveProbeID_SingleCaseInsensitiveMatch(t *testing.T) {
	// Single case-insensitive match should succeed
	probeID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := client.ProbeListResponse{
			Probes: []client.Probe{
				{ID: probeID, Name: "production-api", Status: "up"},
			},
			Total: 1,
			Page:  1,
			Limit: 100,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	// Search for "PRODUCTION-API" should match "production-api" case-insensitively
	result, err := ResolveProbeID(ctx, c, "PRODUCTION-API")
	require.NoError(t, err)
	assert.Equal(t, probeID, result)
}

func TestResolveProbeIDs_Multiple(t *testing.T) {
	// Resolve multiple identifiers - mix of UUIDs and names
	uuid1 := uuid.New()
	uuid2 := uuid.New()
	probeName := "staging-db"

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Only the name should trigger an API call
		response := client.ProbeListResponse{
			Probes: []client.Probe{
				{ID: uuid2, Name: probeName, Status: "up"},
			},
			Total: 1,
			Page:  1,
			Limit: 100,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	// Resolve one UUID and one name
	results, err := ResolveProbeIDs(ctx, c, []string{uuid1.String(), probeName})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, uuid1, results[0])
	assert.Equal(t, uuid2, results[1])
	// Only one API call should be made (for the name)
	assert.Equal(t, 1, callCount)
}

func TestResolveProbeIDs_ErrorOnFirstFailure(t *testing.T) {
	// If any resolution fails, return error immediately
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty results for name search
		response := client.ProbeListResponse{
			Probes: []client.Probe{},
			Total:  0,
			Page:   1,
			Limit:  100,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx := context.Background()

	_, err := ResolveProbeIDs(ctx, c, []string{uuid.New().String(), "nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve")
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestFormatAmbiguousError_ShowsMatches(t *testing.T) {
	matches := []client.Probe{
		{ID: uuid.New(), Name: "api-1"},
		{ID: uuid.New(), Name: "api-2"},
		{ID: uuid.New(), Name: "api-3"},
	}

	err := formatAmbiguousError("api", matches)
	errStr := err.Error()

	assert.Contains(t, errStr, "ambiguous")
	assert.Contains(t, errStr, "3 matches")
	assert.Contains(t, errStr, "api-1")
	assert.Contains(t, errStr, "api-2")
	assert.Contains(t, errStr, "api-3")
	assert.Contains(t, errStr, "UUID")
}

func TestFormatAmbiguousError_TruncatesLongList(t *testing.T) {
	matches := make([]client.Probe, 10)
	for i := 0; i < 10; i++ {
		matches[i] = client.Probe{ID: uuid.New(), Name: "probe-" + strings.Repeat("x", i)}
	}

	err := formatAmbiguousError("probe", matches)
	errStr := err.Error()

	assert.Contains(t, errStr, "10 matches")
	assert.Contains(t, errStr, "and 5 more")
}
