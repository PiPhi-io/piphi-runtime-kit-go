package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
	starter = runtimekit.NewRuntimeStarter[demoEntry, map[string]any, map[string]any](
		"minimal-gin-runtime",
		"Minimal Gin Runtime",
		"0.1.0",
		"",
		100,
	)
	runtime   = starter.Runtime
	registry  = starter.Registry
	telemetry = starter.Telemetry
)

func main() {
	router := gin.Default()

	router.GET("/health", handleHealth)
	router.GET("/diagnostics", handleDiagnostics)
	router.POST("/discover", handleDiscover)
	router.POST("/config", handleConfig)
	router.POST("/deconfigure/:configId", handleDeconfigure)
	router.GET("/state", handleState)
	router.POST("/events/example", handleEventExample)
	router.POST("/events/device/:configId/example", handleEventForDevice)
	router.GET("/events", handleEvents)
	router.POST("/telemetry/example", handleTelemetryExample)
	router.POST("/telemetry/device/:configId/example", handleTelemetryForDevice)

	log.Println("minimal gin runtime listening on :8096")
	log.Fatal(router.Run(":8096"))
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, starter.HealthResponse(map[string]any{
		"active_configs": len(registry.IDs()),
	}))
}

func handleDiagnostics(c *gin.Context) {
	c.JSON(http.StatusOK, starter.DiagnosticsResponse(map[string]any{
		"active_config_ids":  registry.IDs(),
		"recent_event_count": len(registry.RecentEvents()),
		"teaching_mode":      "beginner-and-advanced",
	}))
}

func handleDiscover(c *gin.Context) {
	var payload runtimekit.IntegrationDiscoveryRequest
	_ = c.ShouldBindJSON(&payload)
	inputs := runtimekit.NormalizeDiscoveryInputs(payload.Inputs)
	log.Println(runtimekit.FormatDiscoveryAttemptLog(inputs))

	c.JSON(http.StatusOK, runtimekit.BuildDiscoveryResponse([]map[string]any{
		{
			"id":                        "demo-device",
			"device_id":                 "demo-device",
			"host":                      valueOrDefault(inputs["host"], "127.0.0.1"),
			"alias":                     "Demo Device",
			"supports_targeted_example": true,
		},
	}))
}

func handleConfig(c *gin.Context) {
	var payload demoConfig
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsed := adapters.SyncRuntimeAuthFromGinContext(runtime, c, payload.ContainerID)
	log.Println(runtimekit.FormatRuntimeAuthSyncLog(parsed, payload.ContainerID))
	log.Println(runtimekit.FormatConfigApplyLog(map[string]any{
		"id":             payload.ID,
		"container_id":   payload.ContainerID,
		"integration_id": payload.IntegrationID,
		"host":           payload.Host,
		"alias":          payload.Alias,
	}))
	identity := runtimekit.BuildRuntimeIdentity(payload.RuntimeConfig, "minimal-gin-runtime")

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

	c.JSON(http.StatusOK, runtimekit.BuildConfigApplyResponse(
		entry.ConfigID,
		payload.ContainerID,
		map[string]any{
			"host":  payload.Host,
			"alias": payload.Alias,
		},
	))
}

func handleDeconfigure(c *gin.Context) {
	configID := c.Param("configId")
	entry, removed := registry.Remove(configID)
	if removed {
		appendRuntimeEvent("demo.config.removed", entry, map[string]any{
			"host":  entry.Host,
			"alias": entry.Alias,
		})
	}

	c.JSON(http.StatusOK, runtimekit.BuildConfigRemoveResponse(configID, removed, map[string]any{
		"remaining_configs": registry.IDs(),
	}))
}

func handleState(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]any{
		"summary": map[string]any{
			"active_config_count": len(registry.IDs()),
			"recent_event_count":  len(registry.RecentEvents()),
		},
		"entries":         registry.EntriesSnapshot(),
		"state_snapshots": registry.StateSnapshots(),
	})
}

func handleEventExample(c *gin.Context) {
	entry, ok := registry.PrimaryEntry()
	deviceID := "demo-device"
	configID := "demo-device"
	containerID := runtime.Auth.ContainerID()
	integrationID := "minimal-gin-runtime"
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
	c.JSON(http.StatusOK, runtimekit.BuildEventIngestResponse(event))
}

func handleEventForDevice(c *gin.Context) {
	configID := c.Param("configId")
	entry, ok := registry.Get(configID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "reason": "unknown config_id"})
		return
	}

	event := appendRuntimeEvent("demo.device.checked", entry, map[string]any{
		"message": "Advanced example event for a specific configured device",
		"host":    entry.Host,
	})
	c.JSON(http.StatusOK, runtimekit.BuildEventIngestResponse(event))
}

func handleEvents(c *gin.Context) {
	c.JSON(http.StatusOK, runtimekit.BuildEventListResponse(registry.RecentEvents()))
}

func handleTelemetryExample(c *gin.Context) {
	adapters.SyncRuntimeAuthFromGinContext(runtime, c, "")

	entry, ok := registry.PrimaryEntry()
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"ok": false, "reason": "no configured devices"})
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

	c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
}

func handleTelemetryForDevice(c *gin.Context) {
	adapters.SyncRuntimeAuthFromGinContext(runtime, c, "")

	configID := c.Param("configId")
	entry, ok := registry.Get(configID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "reason": "unknown config_id"})
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

	c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
}

func appendRuntimeEvent(eventType string, entry demoEntry, payload map[string]any) map[string]any {
	return registry.AppendEvent(runtimekit.BuildLocalEventRecord(map[string]any{
		"event_type":     eventType,
		"source":         "minimal-gin-runtime",
		"severity":       "info",
		"device_id":      entry.DeviceID,
		"config_id":      entry.ConfigID,
		"container_id":   entry.ContainerID,
		"integration_id": firstNonEmpty(entry.IntegrationID, "minimal-gin-runtime"),
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
