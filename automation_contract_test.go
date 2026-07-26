package runtimekit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRuntimeDeviceRefRequiresStableIdentity(t *testing.T) {
	_, err := NewRuntimeDeviceRef(RuntimeConfig{ID: ""}, "demo")
	if err == nil || !strings.Contains(err.Error(), "config_id") {
		t.Fatalf("expected actionable identity error, got %v", err)
	}
	identity, err := NewRuntimeDeviceRef(RuntimeConfig{ID: "cfg-1", DeviceID: "device-1"}, "demo")
	if err != nil || identity.ConfigID != "cfg-1" || identity.DeviceID != "device-1" {
		t.Fatalf("unexpected identity: %#v err=%v", identity, err)
	}
}

func TestBuildTelemetryMapsRejectsDuplicates(t *testing.T) {
	_, _, err := BuildTelemetryMaps([]TelemetryReading{
		{Metric: "temperature_c", Value: 21.5},
		{Metric: "temperature_c", Value: 22.0},
	})
	if err == nil {
		t.Fatal("expected duplicate metric error")
	}
}

func TestCoreEventPayloadUsesCanonicalEnvelope(t *testing.T) {
	payload := BuildCoreEventPayload(CoreEventPayload{
		Type: "button.pressed", IntegrationID: "integration-1", ConfigID: "config-1",
		ContainerID: "container-1", DeviceID: "device-1", Data: map[string]any{"button": 1},
	})
	if payload.EventID == "" || payload.TS == "" {
		t.Fatal("event identity and timestamp are required")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, field := range []string{"\"event_id\"", "\"type\"", "\"ts\"", "\"data\""} {
		if !strings.Contains(text, field) {
			t.Fatalf("missing canonical field %s in %s", field, text)
		}
	}
}
