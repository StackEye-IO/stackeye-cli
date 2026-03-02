package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/google/uuid"
)

func TestNewProbeGroupCmd(t *testing.T) {
	cmd := NewProbeGroupCmd()

	if cmd.Use != "group" {
		t.Fatalf("Use = %q, want group", cmd.Use)
	}
	if !slices.Contains(cmd.Aliases, "groups") {
		t.Fatalf("expected aliases to contain groups, got %v", cmd.Aliases)
	}

	subs := []string{"create", "list", "get", "update", "delete", "add-probe", "remove-probe", "list-probes"}
	for _, sub := range subs {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == sub {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing subcommand %q", sub)
		}
	}
}

func TestNewProbeGroupCreateCmd_Flags(t *testing.T) {
	cmd := NewProbeGroupCreateCmd()
	if cmd.Flags().Lookup("name") == nil {
		t.Fatal("expected --name flag")
	}
	if cmd.Flags().Lookup("description") == nil {
		t.Fatal("expected --description flag")
	}
	if cmd.Flags().Lookup("type") == nil {
		t.Fatal("expected --type flag")
	}
	if cmd.Flags().Lookup("label") == nil {
		t.Fatal("expected --label flag")
	}
}

func TestRunProbeGroupCreate_Validation(t *testing.T) {
	cmd := NewProbeGroupCreateCmd()
	cmd.SetArgs([]string{"--name", "   "})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--name is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunProbeGroupUpdate_Validation(t *testing.T) {
	cmd := NewProbeGroupUpdateCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "at least one update flag") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunProbeGroupCreate_DynamicRequiresLabel(t *testing.T) {
	cmd := NewProbeGroupCreateCmd()
	cmd.SetArgs([]string{"--name", "dynamic-group", "--type", "dynamic"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "dynamic groups require at least one --label selector") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildGroupLabelSelector(t *testing.T) {
	selector, err := buildGroupLabelSelector([]string{"env=prod", "tier in web|api", "pci"})
	if err != nil {
		t.Fatalf("buildGroupLabelSelector() error = %v", err)
	}
	if selector == nil {
		t.Fatal("expected selector")
	}

	var decoded []map[string]any
	if err := json.Unmarshal(*selector, &decoded); err != nil {
		t.Fatalf("unmarshal selector: %v", err)
	}
	if len(decoded) != 3 {
		t.Fatalf("expected 3 conditions, got %d", len(decoded))
	}
	if decoded[0]["key"] != "env" || decoded[0]["op"] != "eq" || decoded[0]["value"] != "prod" {
		t.Fatalf("unexpected first condition: %#v", decoded[0])
	}
	if decoded[1]["key"] != "tier" || decoded[1]["op"] != "in" {
		t.Fatalf("unexpected second condition: %#v", decoded[1])
	}
	if decoded[2]["key"] != "pci" || decoded[2]["op"] != "exists" {
		t.Fatalf("unexpected third condition: %#v", decoded[2])
	}
}

func TestRunProbeGroupListProbes_Validation(t *testing.T) {
	cmd := NewProbeGroupListProbesCmd()
	cmd.SetArgs([]string{"550e8400-e29b-41d4-a716-446655440000", "--page", "0"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid page") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProbeGroupLifecycleIntegration_MockServer(t *testing.T) {
	groupID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	probeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	type groupState struct {
		exists  bool
		members map[uuid.UUID]bool
	}
	state := &groupState{exists: true, members: make(map[uuid.UUID]bool)}
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/groups":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data": map[string]any{
					"group": map[string]any{
						"id":           groupID,
						"name":         "Cluster fruition-infra-doks-sfo3",
						"type":         "static",
						"color":        "#6B7280",
						"member_count": 0,
						"created_at":   "2026-02-01T00:00:00Z",
						"updated_at":   "2026-02-01T00:00:00Z",
					},
				},
			})
			return

		case r.Method == http.MethodPost && r.URL.Path == "/v1/groups/"+groupID.String()+"/members":
			var req struct {
				ProbeIDs []uuid.UUID `json:"probe_ids"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode add members: %v", err)
			}
			for _, id := range req.ProbeIDs {
				state.members[id] = true
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "message": "members added"})
			return

		case r.Method == http.MethodGet && r.URL.Path == "/v1/groups/"+groupID.String()+"/probes":
			ids := make([]uuid.UUID, 0, len(state.members))
			for id := range state.members {
				ids = append(ids, id)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data": map[string]any{
					"probe_ids": ids,
					"total":     len(ids),
				},
			})
			return

		case r.Method == http.MethodGet && r.URL.Path == "/v1/probes/"+probeID.String():
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":               probeID,
				"name":             "Production API",
				"url":              "https://api.example.com/health",
				"status":           "up",
				"check_type":       "http",
				"interval_seconds": 60,
				"timeout_ms":       5000,
				"created_at":       "2026-02-01T00:00:00Z",
				"updated_at":       "2026-02-01T00:00:00Z",
			})
			return

		case r.Method == http.MethodDelete && r.URL.Path == "/v1/groups/"+groupID.String()+"/members/"+probeID.String():
			delete(state.members, probeID)
			w.WriteHeader(http.StatusNoContent)
			return

		case r.Method == http.MethodDelete && r.URL.Path == "/v1/groups/"+groupID.String():
			state.exists = false
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "NOT_FOUND", "message": "not found"})
	}))
	defer server.Close()

	prevGetter := api.SetConfigGetter(func() *config.Config {
		return &config.Config{
			CurrentContext: "test",
			Contexts: map[string]*config.Context{
				"test": {
					APIURL: server.URL,
					APIKey: "test-api-key",
				},
			},
		}
	})
	defer api.SetConfigGetter(prevGetter)

	run := func(args ...string) error {
		cmd := NewProbeGroupCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetContext(context.Background())
		cmd.SetArgs(args)
		return cmd.Execute()
	}

	if err := run("create", "--name", "Cluster fruition-infra-doks-sfo3"); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if err := run("add-probe", groupID.String(), probeID.String()); err != nil {
		t.Fatalf("add-probe failed: %v", err)
	}
	if err := run("list-probes", groupID.String()); err != nil {
		t.Fatalf("list-probes failed: %v", err)
	}

	mu.Lock()
	_, memberExists := state.members[probeID]
	mu.Unlock()
	if !memberExists {
		t.Fatal("expected probe membership after add-probe")
	}

	if err := run("remove-probe", groupID.String(), probeID.String()); err != nil {
		t.Fatalf("remove-probe failed: %v", err)
	}
	if err := run("delete", groupID.String(), "--yes"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	mu.Lock()
	_, memberStillExists := state.members[probeID]
	groupStillExists := state.exists
	mu.Unlock()
	if memberStillExists {
		t.Fatal("expected probe membership removed")
	}
	if groupStillExists {
		t.Fatal("expected group to be deleted")
	}
}
