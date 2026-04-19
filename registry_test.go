package runtimekit

import "testing"

func TestRuntimeRegistryUpdateStateUsesExplicitDeviceID(t *testing.T) {
	registry := NewRuntimeRegistry[map[string]any, map[string]any, map[string]any](10)
	snapshot := registry.UpdateState("cfg-1", map[string]any{"temperature_c": 21.5}, "sensor-1")

	if snapshot.DeviceID != "sensor-1" {
		t.Fatalf("unexpected device id %q", snapshot.DeviceID)
	}
	if registry.StateSnapshots()["cfg-1"].DeviceID != "sensor-1" {
		t.Fatalf("unexpected stored device id %q", registry.StateSnapshots()["cfg-1"].DeviceID)
	}
}
