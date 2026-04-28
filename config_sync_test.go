package runtimekit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCoreBaseURLUsesManagedRuntimeEnv(t *testing.T) {
	t.Setenv(CoreBaseURLEnvName, "http://127.0.0.1:31419/")

	if got := ResolveCoreBaseURL(""); got != "http://127.0.0.1:31419" {
		t.Fatalf("unexpected core base url %q", got)
	}
}

func TestResolveRuntimeConfigSnapshotPathUsesExplicitPathOrContainerID(t *testing.T) {
	t.Setenv(RuntimeConfigSnapshotPathEnvName, "/tmp/runtime-config.json")
	if got := ResolveRuntimeConfigSnapshotPath("container-1", "/ignored"); got != "/tmp/runtime-config.json" {
		t.Fatalf("unexpected explicit snapshot path %q", got)
	}

	t.Setenv(RuntimeConfigSnapshotPathEnvName, "")
	t.Setenv(RuntimeContainerIDEnvName, "container-2")
	if got := ResolveRuntimeConfigSnapshotPath("", "/.piphinetwork"); got != filepath.Join("/.piphinetwork", "container-2.json") {
		t.Fatalf("unexpected container snapshot path %q", got)
	}
}

func TestLoadRuntimeConfigSnapshotReadsCoreVolumeContract(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "container-1.json")
	payload := []byte(`{
		"schema_version": 1,
		"container_id": "container-1",
		"integration_id": "integration-1",
		"driver_pid": 123,
		"reason": "startup",
		"generation": 42,
		"updated_at": "2026-04-27T12:00:00+00:00",
		"configs": [{"id": "device-1", "serial": "abc123"}],
		"deleted_config_ids": ["device-2"],
		"config_hash": "sha256:abc",
		"internal_token": "runtime-token"
	}`)
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("failed to write snapshot: %v", err)
	}

	snapshot, err := LoadRuntimeConfigSnapshot[RuntimeConfigWithID](path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected snapshot")
	}
	if snapshot.SchemaVersion != 1 || snapshot.ContainerID != "container-1" || snapshot.IntegrationID != "integration-1" {
		t.Fatalf("unexpected snapshot identity: %#v", snapshot)
	}
	if snapshot.Generation == nil || *snapshot.Generation != 42 {
		t.Fatalf("unexpected generation: %#v", snapshot.Generation)
	}
	if len(snapshot.Configs) != 1 || snapshot.Configs[0].ID != "device-1" {
		t.Fatalf("unexpected configs: %#v", snapshot.Configs)
	}
	if len(snapshot.DeletedConfigIDs) != 1 || snapshot.DeletedConfigIDs[0] != "device-2" {
		t.Fatalf("unexpected tombstones: %#v", snapshot.DeletedConfigIDs)
	}
	if snapshot.ConfigHash != "sha256:abc" || snapshot.InternalToken != "runtime-token" {
		t.Fatalf("unexpected snapshot metadata: %#v", snapshot)
	}
}

func TestLoadRuntimeConfigSnapshotReturnsNilForMissingOrInvalidSnapshot(t *testing.T) {
	dir := t.TempDir()
	missing, err := LoadRuntimeConfigSnapshot[RuntimeConfigWithID](filepath.Join(dir, "missing.json"))
	if err != nil {
		t.Fatalf("missing snapshot should not error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil missing snapshot, got %#v", missing)
	}

	invalidPath := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("failed to write invalid snapshot: %v", err)
	}
	invalid, err := LoadRuntimeConfigSnapshot[RuntimeConfigWithID](invalidPath)
	if err != nil {
		t.Fatalf("invalid json should not error: %v", err)
	}
	if invalid != nil {
		t.Fatalf("expected nil invalid snapshot, got %#v", invalid)
	}
}

func TestBuildRuntimeConfigSnapshotFromCoreRowsConvertsValidRows(t *testing.T) {
	snapshot, err := BuildRuntimeConfigSnapshotFromCoreRows[RuntimeConfigWithID](
		[]map[string]any{
			{"config_data": map[string]any{"id": "device-1", "name": "Kitchen"}},
			{"config_data": nil},
			{"other": map[string]any{"id": "ignored"}},
		},
		"container-1",
		nil,
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snapshot.ContainerID != "container-1" || snapshot.Reason != "startup_rehydrate" {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	if len(snapshot.Configs) != 1 || snapshot.Configs[0].ID != "device-1" {
		t.Fatalf("unexpected configs: %#v", snapshot.Configs)
	}
	if snapshot.Configs[0].ContainerID != "container-1" {
		t.Fatalf("container id was not injected: %#v", snapshot.Configs[0])
	}
}

func TestFetchCoreRuntimeConfigSnapshotUsesRuntimeAuthHeaders(t *testing.T) {
	runtimeContext := NewRuntimeContext()
	runtimeContext.Auth.Update("container-1", "secret-token")

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != CoreRuntimeConfigFetchPath {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if request.URL.Query().Get("container_id") != "container-1" {
			t.Fatalf("missing container id query: %s", request.URL.RawQuery)
		}
		if request.Header.Get(RuntimeContainerIDHeaderName) != "container-1" {
			t.Fatalf("missing container id header")
		}
		if request.Header.Get(RuntimeInternalTokenHeaderName) != "secret-token" {
			t.Fatalf("missing internal token header")
		}
		_ = json.NewEncoder(response).Encode([]map[string]any{
			{"config_data": map[string]any{"id": "device-1"}},
		})
	}))
	defer server.Close()

	snapshot, err := FetchCoreRuntimeConfigSnapshot[RuntimeConfigWithID](
		context.Background(),
		runtimeContext,
		server.Client(),
		server.URL,
		nil,
		"",
	)
	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}
	if snapshot == nil || len(snapshot.Configs) != 1 || snapshot.Configs[0].ID != "device-1" {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
}

func TestRehydrateRuntimeConfigsAppliesSnapshotThenCore(t *testing.T) {
	runtimeContext := NewRuntimeContext()
	runtimeContext.Auth.Update("container-1", "secret-token")

	dir := t.TempDir()
	path := filepath.Join(dir, "container-1.json")
	if err := os.WriteFile(path, []byte(`{
		"container_id": "container-1",
		"reason": "startup_snapshot",
		"generation": 1,
		"configs": [{"id": "snapshot-device"}]
	}`), 0o600); err != nil {
		t.Fatalf("failed to write snapshot: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(response).Encode([]map[string]any{
			{"config_data": map[string]any{"id": "core-device"}},
		})
	}))
	defer server.Close()

	applied := [][]string{}
	result, err := RehydrateRuntimeConfigs(
		context.Background(),
		RuntimeConfigRehydrateOptions[RuntimeConfigWithID]{
			RuntimeContext: runtimeContext,
			HTTPClient:     server.Client(),
			CoreBaseURL:    server.URL,
			SnapshotPath:   path,
			ApplySnapshot: func(snapshot RuntimeConfigSnapshot[RuntimeConfigWithID]) error {
				ids := []string{snapshot.Reason}
				for _, config := range snapshot.Configs {
					ids = append(ids, config.ID)
				}
				applied = append(applied, ids)
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("unexpected rehydrate error: %v", err)
	}
	if !result.SnapshotApplied || !result.CoreApplied {
		t.Fatalf("expected snapshot and core applied: %#v", result)
	}
	if len(applied) != 2 || applied[0][1] != "snapshot-device" || applied[1][1] != "core-device" {
		t.Fatalf("unexpected apply order: %#v", applied)
	}
}

func TestRehydrateRuntimeConfigsKeepsSnapshotWhenCoreIsOffline(t *testing.T) {
	runtimeContext := NewRuntimeContext()
	runtimeContext.Auth.Update("container-1", "secret-token")

	dir := t.TempDir()
	path := filepath.Join(dir, "container-1.json")
	if err := os.WriteFile(path, []byte(`{
		"container_id": "container-1",
		"configs": [{"id": "snapshot-device"}]
	}`), 0o600); err != nil {
		t.Fatalf("failed to write snapshot: %v", err)
	}

	applied := []string{}
	result, err := RehydrateRuntimeConfigs(
		context.Background(),
		RuntimeConfigRehydrateOptions[RuntimeConfigWithID]{
			RuntimeContext: runtimeContext,
			HTTPClient:     &http.Client{},
			CoreBaseURL:    "http://127.0.0.1:1",
			SnapshotPath:   path,
			ApplySnapshot: func(snapshot RuntimeConfigSnapshot[RuntimeConfigWithID]) error {
				for _, config := range snapshot.Configs {
					applied = append(applied, config.ID)
				}
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("unexpected rehydrate error: %v", err)
	}
	if !result.SnapshotApplied || !result.CoreAttempted || result.CoreApplied || result.CoreError == "" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if len(applied) != 1 || applied[0] != "snapshot-device" {
		t.Fatalf("unexpected applied configs: %#v", applied)
	}
}

func TestConfigSyncCoordinatorAppliesSnapshotAndTracksGeneration(t *testing.T) {
	state := NewRuntimeProcessState()
	coordinator := NewConfigSyncCoordinator[RuntimeConfigWithID](state)
	generation := 12
	applied := []string{}
	removed := []string{}

	response, err := coordinator.ApplySnapshot(
		RuntimeConfigSnapshot[RuntimeConfigWithID]{
			Generation: &generation,
			Configs: []RuntimeConfigWithID{
				{RuntimeConfig: RuntimeConfig{ID: "cfg-1"}},
				{RuntimeConfig: RuntimeConfig{ID: "cfg-2"}},
			},
		},
		[]string{"cfg-1", "cfg-3"},
		func(config RuntimeConfigWithID) error {
			applied = append(applied, config.ID)
			return nil
		},
		func(configID string) (bool, error) {
			removed = append(removed, configID)
			return true, nil
		},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected sync error: %v", err)
	}
	if len(applied) != 2 || applied[0] != "cfg-1" || applied[1] != "cfg-2" {
		t.Fatalf("unexpected applied ids: %#v", applied)
	}
	if len(removed) != 1 || removed[0] != "cfg-3" {
		t.Fatalf("unexpected removed ids: %#v", removed)
	}
	if response.Generation == nil || *response.Generation != generation {
		t.Fatalf("unexpected response generation: %#v", response.Generation)
	}
	if state.CurrentGeneration() == nil || *state.CurrentGeneration() != generation {
		t.Fatalf("unexpected state generation: %#v", state.CurrentGeneration())
	}
}
