package runtimekit

import (
	"encoding/json"
	"slices"
	"strings"
)

var defaultSecretKeys = []string{"password", "token", "secret", "apiKey", "api_key"}

// RedactConfigSecrets returns a shallow copy with likely secret values redacted.
func RedactConfigSecrets(config map[string]any) map[string]any {
	clone := map[string]any{}
	for key, value := range config {
		if slices.Contains(defaultSecretKeys, key) {
			clone[key] = "***redacted***"
			continue
		}
		clone[key] = value
	}
	return clone
}

// FormatConfigApplyLog builds a safe config lifecycle log line.
func FormatConfigApplyLog(config map[string]any) string {
	configID, _ := config["id"].(string)
	containerID, _ := config["container_id"].(string)
	integrationID, _ := config["integration_id"].(string)

	if strings.TrimSpace(configID) == "" {
		configID = "<missing>"
	}
	if strings.TrimSpace(containerID) == "" {
		containerID = "<missing>"
	}
	if strings.TrimSpace(integrationID) == "" {
		integrationID = "<missing>"
	}

	payload, _ := json.Marshal(RedactConfigSecrets(config))
	return "config_apply " +
		"config_id=" + configID + " " +
		"container_id=" + containerID + " " +
		"integration_id=" + integrationID + " " +
		"config=" + string(payload)
}

// BuildConfigApplyResponse returns the standard config apply payload.
func BuildConfigApplyResponse(configID, containerID string, metadata map[string]any) RuntimeConfigApplyResponse {
	return RuntimeConfigApplyResponse{
		OK:          true,
		ConfigID:    configID,
		ContainerID: containerID,
		Metadata:    metadata,
	}
}

// BuildConfigRemoveResponse returns the standard config remove payload.
func BuildConfigRemoveResponse(configID string, removed bool, metadata map[string]any) RuntimeConfigRemoveResponse {
	return RuntimeConfigRemoveResponse{
		OK:       true,
		ConfigID: configID,
		Removed:  removed,
		Metadata: metadata,
	}
}
