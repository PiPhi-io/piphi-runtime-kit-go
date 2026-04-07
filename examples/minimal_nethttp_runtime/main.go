package main

import (
	"encoding/json"
	"log"
	"net/http"

	runtimekit "github.com/piphi-network/piphi-runtime-kit-go"
	"github.com/piphi-network/piphi-runtime-kit-go/adapters"
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
	runtime   = runtimekit.NewRuntimeContext()
	registry  = runtimekit.NewRuntimeRegistry[demoEntry, map[string]any, map[string]any](100)
	telemetry = runtimekit.NewTelemetryClient(runtime.ProcessState, "", 0)
)

func main() {
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/diagnostics", handleDiagnostics)
	http.HandleFunc("/discover", handleDiscover)
	http.HandleFunc("/config", handleConfig)
	http.HandleFunc("/deconfigure", handleDeconfigure)
	http.HandleFunc("/state", handleState)
	http.HandleFunc("/events/example", handleEventExample)
	http.HandleFunc("/events", handleEvents)
	http.HandleFunc("/telemetry/example", handleTelemetryExample)

	log.Println("minimal net/http runtime listening on :8095")
	log.Fatal(http.ListenAndServe(":8095", nil))
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, runtimekit.BuildRuntimeHealthResponse(
		runtime,
		map[string]any{
			"id":      "minimal-nethttp-runtime",
			"name":    "Minimal net/http Runtime",
			"version": "0.1.0",
		},
		map[string]any{
			"active_configs": len(registry.IDs()),
		},
	))
}

func handleDiagnostics(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, runtimekit.BuildRuntimeDiagnosticsResponse(
		runtime,
		map[string]any{
			"id":      "minimal-nethttp-runtime",
			"name":    "Minimal net/http Runtime",
			"version": "0.1.0",
		},
		map[string]any{
			"active_config_ids":  registry.IDs(),
			"recent_event_count": len(registry.RecentEvents()),
		},
	))
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
			"id":    "demo-device",
			"host":  valueOrDefault(inputs["host"], "127.0.0.1"),
			"alias": "Demo Device",
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

	entry := demoEntry{
		ConfigID:      payload.ID,
		DeviceID:      firstNonEmpty(payload.DeviceID, payload.ID),
		ContainerID:   payload.ContainerID,
		IntegrationID: payload.IntegrationID,
		Host:          payload.Host,
		Alias:         payload.Alias,
		Config:        payload,
		LatestState: map[string]any{
			"connected": true,
			"host":      payload.Host,
		},
	}
	registry.Set(payload.ID, entry)
	registry.UpdateState(payload.ID, entry.LatestState)

	writeJSON(w, http.StatusOK, runtimekit.BuildConfigApplyResponse(
		payload.ID,
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

	_, removed := registry.Remove(payload.ConfigID)
	writeJSON(w, http.StatusOK, runtimekit.BuildConfigRemoveResponse(payload.ConfigID, removed, map[string]any{
		"remaining_configs": registry.IDs(),
	}))
}

func handleState(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"entries":         registry.EntriesSnapshot(),
		"state_snapshots": registry.StateSnapshots(),
	})
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

	event := registry.AppendEvent(runtimekit.BuildLocalEventRecord(map[string]any{
		"event_type":     "demo.event",
		"source":         "minimal-nethttp-runtime",
		"severity":       "info",
		"device_id":      deviceID,
		"config_id":      configID,
		"container_id":   containerID,
		"integration_id": integrationID,
		"payload": map[string]any{
			"message": "Example local runtime event",
		},
	}))
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

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
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
