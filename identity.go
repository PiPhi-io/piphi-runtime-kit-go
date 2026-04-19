package runtimekit

import "strings"

// RuntimeIdentity is the canonical config-backed identity record used by examples.
type RuntimeIdentity struct {
	ConfigID      string `json:"config_id"`
	DeviceID      string `json:"device_id"`
	ContainerID   string `json:"container_id,omitempty"`
	IntegrationID string `json:"integration_id,omitempty"`
}

func firstNonEmptyIdentity(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// ResolveConfigID returns the Core-owned config id for one runtime config payload.
func ResolveConfigID(config RuntimeConfig) string {
	return firstNonEmptyIdentity(config.ConfigID, config.ID)
}

// ResolveDeviceID returns the integration-owned device id for one runtime config payload.
func ResolveDeviceID(config RuntimeConfig) string {
	return firstNonEmptyIdentity(config.DeviceID, config.ID)
}

// BuildRuntimeIdentity returns the canonical identity record for one runtime config payload.
func BuildRuntimeIdentity(config RuntimeConfig, integrationID string) RuntimeIdentity {
	return RuntimeIdentity{
		ConfigID:      ResolveConfigID(config),
		DeviceID:      ResolveDeviceID(config),
		ContainerID:   strings.TrimSpace(config.ContainerID),
		IntegrationID: firstNonEmptyIdentity(integrationID, config.IntegrationID),
	}
}
