package runtimekit

// RuntimeStarter is the beginner-friendly "golden path" bundle for a runtime.
//
// It keeps the most common SDK pieces together so new integration authors can
// start with one object instead of wiring context, registry, telemetry, and
// event clients separately.
type RuntimeStarter[TEntry any, TState any, TEvent any] struct {
	IntegrationID   string
	IntegrationName string
	Version         string
	Runtime         *RuntimeContext
	Registry        *RuntimeRegistry[TEntry, TState, TEvent]
	Telemetry       *TelemetryClient
	Events          *EventClient
	State           *RuntimeStateService[TEntry, TState, TEvent]
}

// NewRuntimeStarter returns the recommended beginner-friendly runtime bundle.
func NewRuntimeStarter[TEntry any, TState any, TEvent any](
	integrationID string,
	integrationName string,
	version string,
	coreBaseURL string,
	maxRecentEvents int,
) *RuntimeStarter[TEntry, TState, TEvent] {
	runtime := NewRuntimeContext()
	if version == "" {
		version = "0.1.0"
	}
	if coreBaseURL != "" {
		runtime.ProcessState.SetCoreBaseURL(coreBaseURL)
	}

	registry := NewRuntimeRegistry[TEntry, TState, TEvent](maxRecentEvents)
	return &RuntimeStarter[TEntry, TState, TEvent]{
		IntegrationID:   integrationID,
		IntegrationName: integrationName,
		Version:         version,
		Runtime:         runtime,
		Registry:        registry,
		Telemetry:       NewTelemetryClient(runtime.ProcessState, coreBaseURL, 0),
		Events:          NewEventClient(runtime.ProcessState, coreBaseURL, 0),
		State:           NewRuntimeStateService(registry),
	}
}

// BuildEntitiesResponse wraps runtime entities in the standard /entities response shape.
func BuildEntitiesResponse[T any](entities []T, capabilities map[string]any, commands map[string]any) RuntimeEntitiesResponse[T] {
	if capabilities == nil {
		capabilities = map[string]any{}
	}
	if commands == nil {
		commands = map[string]any{}
	}
	return RuntimeEntitiesResponse[T]{
		Entities:     entities,
		Capabilities: capabilities,
		Commands:     commands,
	}
}

// IntegrationMetadata returns the standard metadata map used in responses.
func (s *RuntimeStarter[TEntry, TState, TEvent]) IntegrationMetadata() map[string]any {
	return map[string]any{
		"id":      s.IntegrationID,
		"name":    s.IntegrationName,
		"version": s.Version,
	}
}

// HealthResponse builds a standard health response from the starter state.
func (s *RuntimeStarter[TEntry, TState, TEvent]) HealthResponse(metadata map[string]any) RuntimeHealthResponse {
	if metadata == nil {
		metadata = map[string]any{
			"active_configs": len(s.Registry.IDs()),
		}
	}
	return BuildRuntimeHealthResponse(s.Runtime, s.IntegrationMetadata(), metadata)
}

// DiagnosticsResponse builds a standard diagnostics response from the starter state.
func (s *RuntimeStarter[TEntry, TState, TEvent]) DiagnosticsResponse(diagnostics map[string]any) RuntimeDiagnosticsResponse {
	if diagnostics == nil {
		diagnostics = map[string]any{
			"active_config_ids":  s.Registry.IDs(),
			"recent_event_count": len(s.Registry.RecentEvents()),
		}
	}
	return BuildRuntimeDiagnosticsResponse(s.Runtime, s.IntegrationMetadata(), diagnostics)
}
