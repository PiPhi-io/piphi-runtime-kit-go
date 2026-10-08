package main

import (
	"encoding/json"
	"log"
	"net/http"

	runtimekit "github.com/PiPhi-io/piphi-runtime-kit-go"
	"github.com/PiPhi-io/piphi-runtime-kit-go/adapters"
)

type demoConfig struct {
	runtimekit.RuntimeConfig
	Host  string `json:"host"`
	Alias string `json:"alias,omitempty"`
}

type demoEntry struct {
	ConfigID      string         `json:"config_id"`
	DeviceID      string         `json:"device_id"`
	ContainerID   string         `json:"container_id,omitempty"`
	IntegrationID string         `json:"integration_id,omitempty"`
	Host          string         `json:"host"`
	Alias         string         `json:"alias,omitempty"`
	Config        demoConfig     `json:"config"`
	LatestState   map[string]any `json:"latest_state,omitempty"`
}

var (
	starter = runtimekit.NewRuntimeStarter[demoEntry, map[string]any, map[string]any](
		"minimal-nethttp-runtime",
		"Minimal net/http Runtime",
		"0.1.0",
		"",
		100,
	)
	runtime   = starter.Runtime
	registry  = starter.Registry
	telemetry = starter.Telemetry
)

func main() {
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/diagnostics", handleDiagnostics)
	http.HandleFunc("/discover", handleDiscover)
	http.HandleFunc("/config", handleConfig)
	http.HandleFunc("/deconfigure", handleDeconfigure)
	http.HandleFunc("/state", handleState)
	http.HandleFunc("/events/example", handleEventExample)
	http.HandleFunc("/events/device", handleEventForDevice)
	http.HandleFunc("/events", handleEvents)
	http.HandleFunc("/telemetry/example", handleTelemetryExample)
	http.HandleFunc("/telemetry/device", handleTelemetryForDevice)

	log.Println("minimal net/http runtime listening on :8095")
	log.Fatal(http.ListenAndServe(":8095", nil))
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, starter.HealthResponse(nil))
}

func handleDiagnostics(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, starter.DiagnosticsResponse(map[string]any{
		"active_config_ids":  registry.IDs(),
		"recent_event_count": len(registry.RecentEvents()),
		"teaching_mode":      "beginner-and-advanced",
	}))
}

func handleDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload runtimekit.IntegrationDiscoveryRequest
	_ = json.NewDecoder(r.Body).Decode(&payload)
	inputs := runtimekit.NormalizeDiscoveryInputs(payload.Inputs)
	log.Println(runtimekit.FormatDiscoveryAttemptLog(inputs))

	writeJSON(w, http.StatusOK, runtimekit.BuildDiscoveryResponse([]map[string]any{
		{
			"id":                        "demo-device",
			"device_id":                 "demo-device",
			"host":                      valueOrDefault(inputs["host"], "127.0.0.1"),
			"alias":                     "Demo Device",
			"supports_targeted_example": true,
		},
	}))
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload demoConfig
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parsed := adapters.SyncRuntimeAuthFromRequest(runtime, r, payload.ContainerID)
	log.Println(runtimekit.FormatRuntimeAuthSyncLog(parsed, payload.ContainerID))

	logPayload := map[string]any{
		"id":             payload.ID,
		"container_id":   payload.ContainerID,
		"integration_id": payload.IntegrationID,
		"host":           payload.Host,
		"alias":          payload.Alias,
	}
	log.Println(runtimekit.FormatConfigApplyLog(logPayload))
	identity := runtimekit.BuildRuntimeIdentity(payload.RuntimeConfig, "minimal-nethttp-runtime")

	entry := demoEntry{
		ConfigID:      identity.ConfigID,
		DeviceID:      identity.DeviceID,
		ContainerID:   identity.ContainerID,
		IntegrationID: identity.IntegrationID,
		Host:          payload.Host,
		Alias:         payload.Alias,
		Config:        payload,
		LatestState: map[string]any{
			"connected": true,
			"host":      payload.Host,
		},
	}
	registry.Set(payload.ID, entry)
	registry.UpdateState(payload.ID, entry.LatestState, entry.DeviceID)
	appendRuntimeEvent("demo.config.applied", entry, map[string]any{
		"host":  payload.Host,
		"alias": payload.Alias,
	})

	writeJSON(w, http.StatusOK, runtimekit.BuildConfigApplyResponse(
		entry.ConfigID,
		payload.ContainerID,
		map[string]any{
			"host":  payload.Host,
			"alias": payload.Alias,
		},
	))
}

func handleDeconfigure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		ConfigID string `json:"config_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entry, removed := registry.Remove(payload.ConfigID)
	if removed {
		appendRuntimeEvent("demo.config.removed", entry, map[string]any{
			"host":  entry.Host,
			"alias": entry.Alias,
		})
	}
	writeJSON(w, http.StatusOK, runtimekit.BuildConfigRemoveResponse(payload.ConfigID, removed, map[string]any{
		"remaining_configs": registry.IDs(),
	}))
}

func handleState(w http.ResponseWriter, r *http.Request) {
	response, err := starter.State.Response(
		r.Context(),
		r.URL.Query().Get("refresh") == "true",
		r.URL.Query().Get("refresh_request_id"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": err.Error()})
		return
	}
	payload := map[string]any{
		"summary": map[string]any{
			"active_config_count": len(registry.IDs()),
			"recent_event_count":  len(registry.RecentEvents()),
		},
		"entries":         response.Entries,
		"state_snapshots": registry.StateSnapshots(),
	}
	if response.Refresh != nil {
		payload["refresh"] = response.Refresh
	}
	writeJSON(w, http.StatusOK, payload)
}

func handleEventExample(w http.ResponseWriter, _ *http.Request) {
	entry, ok := registry.PrimaryEntry()
	deviceID := "demo-device"
	configID := "demo-device"
	containerID := runtime.Auth.ContainerID()
	integrationID := "minimal-nethttp-runtime"
	if ok {
		deviceID = entry.DeviceID
		configID = entry.ConfigID
		containerID = entry.ContainerID
		integrationID = entry.IntegrationID
	}

	event := appendRuntimeEvent("demo.event", demoEntry{
		DeviceID:      deviceID,
		ConfigID:      configID,
		ContainerID:   containerID,
		IntegrationID: integrationID,
	}, map[string]any{
		"message": "Example local runtime event",
	})
	writeJSON(w, http.StatusOK, runtimekit.BuildEventIngestResponse(event))
}

func handleEventForDevice(w http.ResponseWriter, r *http.Request) {
	configID := r.URL.Query().Get("config_id")
	if configID == "" {
		http.Error(w, "missing config_id query parameter", http.StatusBadRequest)
		return
	}

	entry, ok := registry.Get(configID)
	if !ok {
		http.Error(w, "unknown config_id", http.StatusNotFound)
		return
	}

	event := appendRuntimeEvent("demo.device.checked", entry, map[string]any{
		"message": "Advanced example event for a specific configured device",
		"host":    entry.Host,
	})
	writeJSON(w, http.StatusOK, runtimekit.BuildEventIngestResponse(event))
}

func handleEvents(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, runtimekit.BuildEventListResponse(registry.RecentEvents()))
}

func handleTelemetryExample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	adapters.SyncRuntimeAuthFromRequest(runtime, r, "")

	entry, ok := registry.PrimaryEntry()
	if !ok {
		writeJSON(w, http.StatusConflict, map[string]any{
			"ok":     false,
			"reason": "no configured devices",
		})
		return
	}

	runtimekit.ScheduleTelemetryDelivery(
		runtime.ProcessState,
		telemetry,
		runtime.Auth,
		runtimekit.TelemetryPayload{
			DeviceID:    entry.DeviceID,
			ContainerID: entry.ContainerID,
			Metrics: map[string]any{
				"connected":     true,
				"temperature_c": 21.4,
			},
			Units: map[string]any{
				"temperature_c": "C",
			},
		},
	)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status": "queued",
	})
}

func handleTelemetryForDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	adapters.SyncRuntimeAuthFromRequest(runtime, r, "")

	configID := r.URL.Query().Get("config_id")
	if configID == "" {
		http.Error(w, "missing config_id query parameter", http.StatusBadRequest)
		return
	}

	entry, ok := registry.Get(configID)
	if !ok {
		http.Error(w, "unknown config_id", http.StatusNotFound)
		return
	}

	runtimekit.ScheduleTelemetryDelivery(
		runtime.ProcessState,
		telemetry,
		runtime.Auth,
		runtimekit.TelemetryPayload{
			DeviceID:    entry.DeviceID,
			ContainerID: entry.ContainerID,
			Metrics: map[string]any{
				"connected":     true,
				"temperature_c": 21.4,
				"humidity":      46.0,
			},
			Units: map[string]any{
				"temperature_c": "C",
				"humidity":      "%",
			},
		},
	)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status": "queued",
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func appendRuntimeEvent(eventType string, entry demoEntry, payload map[string]any) map[string]any {
	return registry.AppendEvent(runtimekit.BuildLocalEventRecord(map[string]any{
		"event_type":     eventType,
		"source":         "minimal-nethttp-runtime",
		"severity":       "info",
		"device_id":      entry.DeviceID,
		"config_id":      entry.ConfigID,
		"container_id":   entry.ContainerID,
		"integration_id": firstNonEmpty(entry.IntegrationID, "minimal-nethttp-runtime"),
		"payload":        payload,
	}))
}

func valueOrDefault(value any, fallback string) string {
	if typed, ok := value.(string); ok && typed != "" {
		return typed
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
