package runtimekit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	// CoreBaseURLEnvName is injected by Core for managed runtimes.
	CoreBaseURLEnvName = "PIPHI_CORE_BASE_URL"
	// RuntimeContainerIDEnvName identifies the current managed runtime container.
	RuntimeContainerIDEnvName = "PIPHI_CONTAINER_ID"
	// RuntimeConfigSnapshotPathEnvName overrides the mounted config snapshot path.
	RuntimeConfigSnapshotPathEnvName = "PIPHI_CONFIG_SNAPSHOT_PATH"
	// DefaultRuntimeVolumeDir is the mount path Core binds into managed runtimes.
	DefaultRuntimeVolumeDir = "/.piphinetwork"
	// DefaultRuntimeConfigSnapshotFilename is used when no container id is available.
	DefaultRuntimeConfigSnapshotFilename = "configs.json"
	// CoreRuntimeConfigFetchPath is Core's internal runtime config rehydrate route.
	CoreRuntimeConfigFetchPath = "/api/v2/integrations/config/fetch/all/by/container/internal"
)

// RuntimeConfigRehydrateResult describes startup config rehydration behavior.
type RuntimeConfigRehydrateResult struct {
	SnapshotFound       bool
	SnapshotApplied     bool
	SnapshotConfigCount int
	SnapshotGeneration  *int
	CoreAttempted       bool
	CoreApplied         bool
	CoreConfigCount     int
	CoreGeneration      *int
	CoreError           string
	MissingRuntimeAuth  bool
}

// RuntimeConfigRehydrateOptions configures SDK-owned startup config rehydration.
type RuntimeConfigRehydrateOptions[TConfig any] struct {
	RuntimeContext  *RuntimeContext
	HTTPClient      *http.Client
	CoreBaseURL     string
	SnapshotPath    string
	ApplySnapshot   func(RuntimeConfigSnapshot[TConfig]) error
	DecodeConfig    func(map[string]any) (TConfig, error)
	SnapshotReason  string
	CoreReason      string
	RaiseCoreErrors bool
}

// ResolveCoreBaseURL resolves Core's base URL from the managed-runtime environment.
func ResolveCoreBaseURL(defaultValue string) string {
	value := strings.TrimRight(strings.TrimSpace(os.Getenv(CoreBaseURLEnvName)), "/")
	if value != "" {
		return value
	}
	return strings.TrimRight(strings.TrimSpace(defaultValue), "/")
}

// ResolveRuntimeConfigSnapshotPath resolves the config snapshot mounted by PiPhi Core.
func ResolveRuntimeConfigSnapshotPath(containerID string, volumeDir string) string {
	explicitPath := strings.TrimSpace(os.Getenv(RuntimeConfigSnapshotPathEnvName))
	if explicitPath != "" {
		return explicitPath
	}

	resolvedContainerID := strings.TrimSpace(containerID)
	if resolvedContainerID == "" {
		resolvedContainerID = strings.TrimSpace(os.Getenv(RuntimeContainerIDEnvName))
	}
	resolvedVolumeDir := strings.TrimSpace(volumeDir)
	if resolvedVolumeDir == "" {
		resolvedVolumeDir = DefaultRuntimeVolumeDir
	}
	if resolvedContainerID != "" {
		return filepath.Join(resolvedVolumeDir, resolvedContainerID+".json")
	}
	return filepath.Join(resolvedVolumeDir, DefaultRuntimeConfigSnapshotFilename)
}

// LoadRuntimeConfigSnapshot loads Core's last known config snapshot from disk.
func LoadRuntimeConfigSnapshot[TConfig any](path string) (*RuntimeConfigSnapshot[TConfig], error) {
	snapshotPath := strings.TrimSpace(path)
	if snapshotPath == "" {
		snapshotPath = ResolveRuntimeConfigSnapshotPath("", "")
	}

	payload, err := os.ReadFile(snapshotPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var snapshot RuntimeConfigSnapshot[TConfig]
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return nil, nil
	}
	return &snapshot, nil
}

func decodeRuntimeConfigPayload[TConfig any](
	payload map[string]any,
	decodeConfig func(map[string]any) (TConfig, error),
) (TConfig, error) {
	if decodeConfig != nil {
		return decodeConfig(payload)
	}

	var config TConfig
	raw, err := json.Marshal(payload)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(raw, &config)
	return config, err
}

func hasNonEmptyString(payload map[string]any, key string) bool {
	value, ok := payload[key]
	if !ok {
		return false
	}
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) != ""
}

// BuildRuntimeConfigSnapshotFromCoreRows converts Core internal config rows into a runtime snapshot.
func BuildRuntimeConfigSnapshotFromCoreRows[TConfig any](
	rows []map[string]any,
	containerID string,
	decodeConfig func(map[string]any) (TConfig, error),
	reason string,
) (RuntimeConfigSnapshot[TConfig], error) {
	configs := []TConfig{}
	if strings.TrimSpace(reason) == "" {
		reason = "startup_rehydrate"
	}

	for _, row := range rows {
		value, ok := row["config_data"]
		if !ok {
			continue
		}
		configData, ok := value.(map[string]any)
		if !ok {
			continue
		}

		payload := make(map[string]any, len(configData)+1)
		for key, item := range configData {
			payload[key] = item
		}
		if !hasNonEmptyString(payload, "container_id") && !hasNonEmptyString(payload, "containerId") {
			payload["container_id"] = containerID
		}

		config, err := decodeRuntimeConfigPayload(payload, decodeConfig)
		if err != nil {
			return RuntimeConfigSnapshot[TConfig]{}, err
		}
		configs = append(configs, config)
	}

	return RuntimeConfigSnapshot[TConfig]{
		ContainerID: containerID,
		Reason:      reason,
		Configs:     configs,
	}, nil
}

// FetchCoreRuntimeConfigSnapshot fetches the latest runtime configs from Core.
func FetchCoreRuntimeConfigSnapshot[TConfig any](
	ctx context.Context,
	runtimeContext *RuntimeContext,
	httpClient *http.Client,
	coreBaseURL string,
	decodeConfig func(map[string]any) (TConfig, error),
	reason string,
) (*RuntimeConfigSnapshot[TConfig], error) {
	if runtimeContext == nil || runtimeContext.Auth == nil {
		return nil, nil
	}

	containerID, internalToken := runtimeContext.Auth.Resolve("")
	if strings.TrimSpace(containerID) == "" || strings.TrimSpace(internalToken) == "" {
		return nil, nil
	}

	resolvedCoreBaseURL := ResolveCoreBaseURL(coreBaseURL)
	if resolvedCoreBaseURL == "" {
		resolvedCoreBaseURL = ResolveCoreBaseURL("http://127.0.0.1:31419")
	}
	if resolvedCoreBaseURL == "" {
		return nil, nil
	}

	endpoint, err := url.Parse(resolvedCoreBaseURL + CoreRuntimeConfigFetchPath)
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("container_id", containerID)
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	for key, values := range BuildRuntimeAuthHeaders(containerID, internalToken) {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf("core runtime config fetch failed with HTTP %d", response.StatusCode)
	}

	var rows []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	snapshot, err := BuildRuntimeConfigSnapshotFromCoreRows(
		rows,
		containerID,
		decodeConfig,
		reason,
	)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// RehydrateRuntimeConfigs applies the mounted snapshot first, then refreshes from Core.
func RehydrateRuntimeConfigs[TConfig any](
	ctx context.Context,
	options RuntimeConfigRehydrateOptions[TConfig],
) (RuntimeConfigRehydrateResult, error) {
	result := RuntimeConfigRehydrateResult{}
	if options.RuntimeContext == nil {
		options.RuntimeContext = NewRuntimeContext()
	}

	snapshot, err := LoadRuntimeConfigSnapshot[TConfig](options.SnapshotPath)
	if err != nil {
		return result, err
	}
	if snapshot != nil {
		result.SnapshotFound = true
		if strings.TrimSpace(snapshot.Reason) == "" {
			if strings.TrimSpace(options.SnapshotReason) != "" {
				snapshot.Reason = options.SnapshotReason
			} else {
				snapshot.Reason = "startup_snapshot_rehydrate"
			}
		}
		if options.ApplySnapshot != nil {
			if err := options.ApplySnapshot(*snapshot); err != nil {
				return result, err
			}
		}
		result.SnapshotApplied = true
		result.SnapshotConfigCount = len(snapshot.Configs)
		result.SnapshotGeneration = snapshot.Generation
	}

	containerID, internalToken := options.RuntimeContext.Auth.Resolve("")
	if strings.TrimSpace(containerID) == "" || strings.TrimSpace(internalToken) == "" {
		result.MissingRuntimeAuth = true
		return result, nil
	}

	result.CoreAttempted = true
	coreSnapshot, err := FetchCoreRuntimeConfigSnapshot(
		ctx,
		options.RuntimeContext,
		options.HTTPClient,
		options.CoreBaseURL,
		options.DecodeConfig,
		options.CoreReason,
	)
	if err != nil {
		result.CoreError = err.Error()
		if options.RaiseCoreErrors {
			return result, err
		}
		return result, nil
	}
	if coreSnapshot == nil {
		return result, nil
	}

	if options.ApplySnapshot != nil {
		if err := options.ApplySnapshot(*coreSnapshot); err != nil {
			return result, err
		}
	}
	result.CoreApplied = true
	result.CoreConfigCount = len(coreSnapshot.Configs)
	result.CoreGeneration = coreSnapshot.Generation
	return result, nil
}

// ReconcileConfigIDs diffs incoming and active config id sets.
func ReconcileConfigIDs(incomingConfigIDs []string, activeConfigIDs []string) (toKeep []string, toRemove []string) {
	incomingSet := map[string]struct{}{}
	for _, configID := range incomingConfigIDs {
		incomingSet[configID] = struct{}{}
	}

	for _, configID := range activeConfigIDs {
		if _, ok := incomingSet[configID]; ok {
			toKeep = append(toKeep, configID)
			continue
		}
		toRemove = append(toRemove, configID)
	}

	return toKeep, toRemove
}

// BuildSyncResponse returns the standard config sync response payload.
func BuildSyncResponse(
	appliedConfigIDs []string,
	removedConfigIDs []string,
	skippedConfigIDs []string,
	generation *int,
) RuntimeConfigSyncResponse {
	return RuntimeConfigSyncResponse{
		OK:               true,
		AppliedConfigIDs: appliedConfigIDs,
		RemovedConfigIDs: removedConfigIDs,
		SkippedConfigIDs: skippedConfigIDs,
		Generation:       generation,
	}
}

// ConfigSyncCoordinator coordinates a snapshot apply/remove flow.
type ConfigSyncCoordinator[TConfig interface{ GetRuntimeConfigID() string }] struct {
	processState *RuntimeProcessState
}

// NewConfigSyncCoordinator creates a snapshot coordinator.
func NewConfigSyncCoordinator[TConfig interface{ GetRuntimeConfigID() string }](processState *RuntimeProcessState) *ConfigSyncCoordinator[TConfig] {
	if processState == nil {
		processState = NewRuntimeProcessState()
	}
	return &ConfigSyncCoordinator[TConfig]{processState: processState}
}

// ApplySnapshot reconciles one runtime config snapshot.
func (c *ConfigSyncCoordinator[TConfig]) ApplySnapshot(
	snapshot RuntimeConfigSnapshot[TConfig],
	activeConfigIDs []string,
	applyConfig func(TConfig) error,
	removeConfig func(string) (bool, error),
	getActiveConfigIDs func() []string,
) (RuntimeConfigSyncResponse, error) {
	appliedConfigIDs := []string{}
	removedConfigIDs := []string{}
	skippedConfigIDs := []string{}

	for _, config := range snapshot.Configs {
		if err := applyConfig(config); err != nil {
			return RuntimeConfigSyncResponse{}, err
		}
		appliedConfigIDs = append(appliedConfigIDs, config.GetRuntimeConfigID())
	}

	currentActiveIDs := activeConfigIDs
	if getActiveConfigIDs != nil {
		currentActiveIDs = getActiveConfigIDs()
	}

	incomingConfigIDs := make([]string, 0, len(snapshot.Configs))
	for _, config := range snapshot.Configs {
		incomingConfigIDs = append(incomingConfigIDs, config.GetRuntimeConfigID())
	}

	_, toRemove := ReconcileConfigIDs(incomingConfigIDs, currentActiveIDs)
	for _, configID := range toRemove {
		removed, err := removeConfig(configID)
		if err != nil {
			return RuntimeConfigSyncResponse{}, err
		}
		if removed {
			removedConfigIDs = append(removedConfigIDs, configID)
			continue
		}
		skippedConfigIDs = append(skippedConfigIDs, configID)
	}

	c.processState.SetCurrentGeneration(snapshot.Generation)
	return BuildSyncResponse(appliedConfigIDs, removedConfigIDs, skippedConfigIDs, snapshot.Generation), nil
}

// RuntimeConfigWithID is a small helper that lets RuntimeConfig satisfy sync coordination.
type RuntimeConfigWithID struct {
	RuntimeConfig
}

// GetRuntimeConfigID returns the config id for sync coordination.
func (c RuntimeConfigWithID) GetRuntimeConfigID() string {
	return c.ID
}
