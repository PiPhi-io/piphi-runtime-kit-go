package runtimekit

import "testing"

func TestBuildRuntimeIdentityPrefersExplicitConfigAndDeviceIDs(t *testing.T) {
	config := RuntimeConfig{
		ID:            "core-config-uuid",
		ConfigID:      "core-config-uuid",
		DeviceID:      "vendor-device-42",
		ContainerID:   "container-1",
		IntegrationID: "demo-runtime",
	}

	identity := BuildRuntimeIdentity(config, "")
	if identity.ConfigID != "core-config-uuid" {
		t.Fatalf("unexpected config id %q", identity.ConfigID)
	}
	if identity.DeviceID != "vendor-device-42" {
		t.Fatalf("unexpected device id %q", identity.DeviceID)
	}
	if identity.ContainerID != "container-1" {
		t.Fatalf("unexpected container id %q", identity.ContainerID)
	}
	if identity.IntegrationID != "demo-runtime" {
		t.Fatalf("unexpected integration id %q", identity.IntegrationID)
	}
}

func TestBuildRuntimeIdentityFallsBackToID(t *testing.T) {
	config := RuntimeConfig{ID: "core-config-uuid"}
	identity := BuildRuntimeIdentity(config, "demo-runtime")
	if identity.ConfigID != "core-config-uuid" {
		t.Fatalf("unexpected config id %q", identity.ConfigID)
	}
	if identity.DeviceID != "core-config-uuid" {
		t.Fatalf("unexpected device id %q", identity.DeviceID)
	}
	if identity.IntegrationID != "demo-runtime" {
		t.Fatalf("unexpected integration id %q", identity.IntegrationID)
	}
}
