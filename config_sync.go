package runtimekit

import (
	"encoding/json"
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
)

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
