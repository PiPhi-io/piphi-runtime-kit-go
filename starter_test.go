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
