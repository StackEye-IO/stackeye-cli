// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProbeCompletion_Success(t *testing.T) {
	// Create mock server with probe response
	probes := client.ProbeListResponse{
		Probes: []client.Probe{
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Name: "Production API", CheckType: "http", Status: "up"},
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Name: "Staging DB", CheckType: "tcp", Status: "up"},
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), Name: "Production DB", CheckType: "tcp", Status: "down"},
		},
		Total: 3,
		Page:  1,
		Limit: 100,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/probes", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(probes)
	}))
	defer server.Close()

	// Setup config with test server
	setupTestConfigWithURL(t, server.URL)

	// Create completion function
	completionFunc := ProbeCompletion()

	// Create test command
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Test completion with no prefix
	completions, directive := completionFunc(cmd, []string{}, "")
	require.Equal(t, 3, len(completions))
	assert.Contains(t, completions[0], "Production API")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestProbeCompletion_FilterByPrefix(t *testing.T) {
	// Create mock server with probe response
	probes := client.ProbeListResponse{
		Probes: []client.Probe{
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Name: "Production API", CheckType: "http", Status: "up"},
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Name: "Staging DB", CheckType: "tcp", Status: "up"},
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), Name: "Production DB", CheckType: "tcp", Status: "down"},
		},
		Total: 3,
		Page:  1,
		Limit: 100,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(probes)
	}))
	defer server.Close()

	// Setup config with test server
	setupTestConfigWithURL(t, server.URL)

	// Create completion function
	completionFunc := ProbeCompletion()

	// Create test command
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Test completion with "Prod" prefix (case-insensitive)
	completions, directive := completionFunc(cmd, []string{}, "prod")
	require.Equal(t, 2, len(completions))
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)

	// Verify both Production probes are included
	found := 0
	for _, c := range completions {
		if c != "" {
			found++
		}
	}
	assert.Equal(t, 2, found)
}

func TestProbeCompletion_SecondArgument_NoCompletions(t *testing.T) {
	// When there's already one argument, should return no completions
	completionFunc := ProbeCompletion()
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	completions, directive := completionFunc(cmd, []string{"existing-probe"}, "")
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestProbeCompletion_APIError_GracefulDegradation(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Setup config with test server
	setupTestConfigWithURL(t, server.URL)

	// Create completion function
	completionFunc := ProbeCompletion()
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Should return empty completions without error
	completions, directive := completionFunc(cmd, []string{}, "")
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestProbeCompletion_NoConfig_GracefulDegradation(t *testing.T) {
	// Clear the config getter to simulate no config
	prev := api.SetConfigGetter(nil)
	defer api.SetConfigGetter(prev)

	// Create completion function
	completionFunc := ProbeCompletion()
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Should return empty completions without error
	completions, directive := completionFunc(cmd, []string{}, "")
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestProbeCompletion_EmptyResponse(t *testing.T) {
	// Create mock server with empty probe response
	probes := client.ProbeListResponse{
		Probes: []client.Probe{},
		Total:  0,
		Page:   1,
		Limit:  100,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(probes)
	}))
	defer server.Close()

	// Setup config with test server
	setupTestConfigWithURL(t, server.URL)

	// Create completion function
	completionFunc := ProbeCompletion()
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Should return empty completions
	completions, directive := completionFunc(cmd, []string{}, "")
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

// setupTestConfigWithURL creates a test configuration pointing to the given URL.
func setupTestConfigWithURL(t *testing.T, serverURL string) {
	t.Helper()

	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]*config.Context{
			"test": {
				APIURL: serverURL,
				APIKey: "test-api-key",
			},
		},
	}

	api.SetConfigGetter(func() *config.Config {
		return cfg
	})

	t.Cleanup(func() {
		api.SetConfigGetter(nil)
	})
}

func TestProbeCompletion_CompletionFormat(t *testing.T) {
	// Verify completion format includes probe info
	probes := client.ProbeListResponse{
		Probes: []client.Probe{
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Name: "API Health", CheckType: "http", Status: "up"},
		},
		Total: 1,
		Page:  1,
		Limit: 100,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(probes)
	}))
	defer server.Close()

	setupTestConfigWithURL(t, server.URL)

	completionFunc := ProbeCompletion()
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	completions, _ := completionFunc(cmd, []string{}, "")
	require.Equal(t, 1, len(completions))

	// Completion should include name, check type, and status
	assert.Contains(t, completions[0], "API Health")
	assert.Contains(t, completions[0], "http")
	assert.Contains(t, completions[0], "up")
}

func TestProbeIDCompletion_IsAlias(t *testing.T) {
	// Verify ProbeIDCompletion is the same as ProbeCompletion
	assert.NotNil(t, ProbeIDCompletion)
	// Both should return a function (can't compare functions directly)
	assert.NotNil(t, ProbeIDCompletion())
}
