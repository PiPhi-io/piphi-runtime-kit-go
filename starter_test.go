package runtimekit

import "testing"

func TestNewRuntimeStarterBundlesCommonRuntimeParts(t *testing.T) {
	starter := NewRuntimeStarter[map[string]any, map[string]any, map[string]any](
		"demo-runtime",
		"Demo Runtime",
		"0.1.0",
		"",
		25,
	)

	if starter.IntegrationID != "demo-runtime" {
		t.Fatalf("unexpected integration id %q", starter.IntegrationID)
	}
	if starter.Registry == nil || starter.Runtime == nil || starter.Telemetry == nil || starter.Events == nil {
		t.Fatal("expected starter to initialize runtime helpers")
	}
	if len(starter.HealthResponse(nil).Integration) == 0 {
		t.Fatal("expected health response to include integration metadata")
	}
}

func TestBuildEntitiesResponseWrapsRuntimeEntities(t *testing.T) {
	response := BuildEntitiesResponse(
		[]RuntimeEntity{
			{
				ID:           "office-plug",
				Name:         "Office Plug",
				ConfigID:     "cfg-1",
				DeviceID:     "office-plug",
				DeviceClass:  "plug",
				EntityType:   "switch",
				Capabilities: []string{"switch", "power"},
				Dashboard: &RuntimeEntityDashboard{
					AllowedWidgets:     []string{"tile", "stat"},
					DefaultWidget:      "tile",
					RecommendedWidgets: []string{"tile"},
				},
			},
		},
		map[string]any{"switch": map[string]any{"kind": "action"}},
		map[string]any{"turn_on": map[string]any{"description": "Turn on"}},
	)

	if len(response.Entities) != 1 {
		t.Fatalf("expected one entity, got %d", len(response.Entities))
	}
	if response.Entities[0].DeviceClass != "plug" {
		t.Fatalf("unexpected device class %q", response.Entities[0].DeviceClass)
	}
	if response.Entities[0].Dashboard == nil || response.Entities[0].Dashboard.DefaultWidget != "tile" {
		t.Fatal("expected dashboard hints to be preserved")
	}
	if response.Commands["turn_on"] == nil {
		t.Fatal("expected commands metadata to be preserved")
	}
	if response.Capabilities["switch"] == nil {
		t.Fatal("expected capabilities metadata to be preserved")
	}
}
