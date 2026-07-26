package runtimekit

import (
	"fmt"
	"strings"
)

// RuntimeIdentity is the canonical config-backed identity record used by examples.
type RuntimeIdentity struct {
	ConfigID      string `json:"config_id"`
	DeviceID      string `json:"device_id"`
	ContainerID   string `json:"container_id,omitempty"`
	IntegrationID string `json:"integration_id,omitempty"`
}

// RuntimeIdentityError identifies incomplete Core-managed device scope.
type RuntimeIdentityError struct{ Missing []string }

func (e RuntimeIdentityError) Error() string {
	return fmt.Sprintf("runtime device identity is missing %s", strings.Join(e.Missing, ", "))
}

// NewRuntimeDeviceRef validates the stable identity required for telemetry.
func NewRuntimeDeviceRef(config RuntimeConfig, integrationID string) (RuntimeIdentity, error) {
	identity := BuildRuntimeIdentity(config, integrationID)
	missing := []string{}
	if identity.ConfigID == "" {
		missing = append(missing, "config_id")
	}
	if identity.DeviceID == "" {
		missing = append(missing, "device_id")
	}
	if len(missing) > 0 {
		return RuntimeIdentity{}, RuntimeIdentityError{Missing: missing}
	}
	return identity, nil
}

// RequireEventScope validates the additional routing fields semantic events require.
func (identity RuntimeIdentity) RequireEventScope() error {
	missing := []string{}
	if identity.IntegrationID == "" {
		missing = append(missing, "integration_id")
	}
	if identity.ContainerID == "" {
		missing = append(missing, "container_id")
	}
	if len(missing) > 0 {
		return RuntimeIdentityError{Missing: missing}
	}
	return nil
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
