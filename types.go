package runtimekit

// RuntimeConfig is the minimal config shape shared across integrations.
type RuntimeConfig struct {
	ID            string `json:"id"`
	ConfigID      string `json:"config_id,omitempty"`
	ContainerID   string `json:"container_id,omitempty"`
	IntegrationID string `json:"integration_id,omitempty"`
	DeviceID      string `json:"device_id,omitempty"`
}

// RuntimeConfigApplyResponse is returned after one config is applied.
type RuntimeConfigApplyResponse struct {
	OK          bool           `json:"ok"`
	ConfigID    string         `json:"config_id"`
	ContainerID string         `json:"container_id,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// RuntimeConfigRemoveResponse is returned after a config removal attempt.
type RuntimeConfigRemoveResponse struct {
	OK       bool           `json:"ok"`
	ConfigID string         `json:"config_id"`
	Removed  bool           `json:"removed"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// RuntimeConfigSnapshot is the runtime-facing config snapshot shape.
type RuntimeConfigSnapshot[TConfig any] struct {
	ContainerID   string    `json:"container_id,omitempty"`
	IntegrationID string    `json:"integration_id,omitempty"`
	Generation    *int      `json:"generation,omitempty"`
	Configs       []TConfig `json:"configs"`
}

// RuntimeConfigSyncResponse is returned after a snapshot reconciliation pass.
type RuntimeConfigSyncResponse struct {
	OK               bool     `json:"ok"`
	AppliedConfigIDs []string `json:"applied_config_ids"`
	RemovedConfigIDs []string `json:"removed_config_ids"`
	SkippedConfigIDs []string `json:"skipped_config_ids"`
	Generation       *int     `json:"generation,omitempty"`
}

// IntegrationDiscoveryRequest carries optional discovery inputs.
type IntegrationDiscoveryRequest struct {
	Inputs map[string]any `json:"inputs,omitempty"`
}

// IntegrationDiscoveryResponse is the standard runtime discovery response.
type IntegrationDiscoveryResponse[TDevice any] struct {
	Devices []TDevice `json:"devices"`
}

// IntegrationEventRequest is the local runtime event shape.
type IntegrationEventRequest struct {
	EventType     string         `json:"event_type"`
	Source        string         `json:"source,omitempty"`
	Severity      string         `json:"severity,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
	DeviceID      string         `json:"device_id,omitempty"`
	ConfigID      string         `json:"config_id,omitempty"`
	ContainerID   string         `json:"container_id,omitempty"`
	IntegrationID string         `json:"integration_id,omitempty"`
}

// IntegrationEventIngestResponse is returned after one local event is recorded.
type IntegrationEventIngestResponse[TEvent any] struct {
	OK    bool   `json:"ok"`
	Event TEvent `json:"event"`
}

// IntegrationEventListResponse returns recent local runtime events.
type IntegrationEventListResponse[TEvent any] struct {
	Events []TEvent `json:"events"`
}

// RuntimeHealthResponse is the standard health payload.
type RuntimeHealthResponse struct {
	OK                 bool           `json:"ok"`
	Integration        map[string]any `json:"integration,omitempty"`
	RuntimeAuthPresent bool           `json:"runtime_auth_present"`
	CoreClientBound    bool           `json:"core_client_bound"`
	PendingTaskCount   int            `json:"pending_task_count"`
	CurrentGeneration  *int           `json:"current_generation,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

// RuntimeDiagnosticsResponse is the standard diagnostics payload.
type RuntimeDiagnosticsResponse struct {
	OK                 bool           `json:"ok"`
	Integration        map[string]any `json:"integration,omitempty"`
	RuntimeAuthPresent bool           `json:"runtime_auth_present"`
	CoreClientBound    bool           `json:"core_client_bound"`
	PendingTaskCount   int            `json:"pending_task_count"`
	CurrentGeneration  *int           `json:"current_generation,omitempty"`
	Diagnostics        map[string]any `json:"diagnostics,omitempty"`
}

// TelemetryPayload is the canonical runtime telemetry shape.
type TelemetryPayload struct {
	DeviceID      string         `json:"device_id"`
	Metrics       map[string]any `json:"metrics"`
	Units         map[string]any `json:"units,omitempty"`
	ContainerID   string         `json:"container_id,omitempty"`
	IntegrationID string         `json:"integration_id,omitempty"`
	Timestamp     string         `json:"timestamp,omitempty"`
}

// CoreEventPayload is the event shape sent back to PiPhi Core.
type CoreEventPayload struct {
	EventType     string         `json:"event_type"`
	Source        string         `json:"source,omitempty"`
	Severity      string         `json:"severity,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
	ConfigID      string         `json:"config_id"`
	ContainerID   string         `json:"container_id"`
	IntegrationID string         `json:"integration_id"`
	DeviceID      string         `json:"device_id,omitempty"`
}
