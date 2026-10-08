package runtimekit

import (
	"context"
	"testing"
)

func TestRuntimeStateServiceRefreshesAndBuildsReceipt(t *testing.T) {
	starter := NewRuntimeStarter[map[string]any, map[string]any, map[string]any](
		"demo", "Demo", "", "", 100,
	)
	err := starter.State.Provide(func(context.Context) error {
		starter.State.Publish("device-1", map[string]any{"is_on": true})
		return nil
	}, "device_api")
	if err != nil {
		t.Fatal(err)
	}

	payload, err := starter.State.Response(context.Background(), true, "request-1")
	if err != nil {
		t.Fatal(err)
	}
	if payload.Refresh == nil || !payload.Refresh.Performed || payload.Refresh.Status != "refreshed" {
		t.Fatalf("unexpected refresh receipt: %#v", payload.Refresh)
	}
	if payload.Entries["device-1"].State["is_on"] != true {
		t.Fatalf("unexpected state: %#v", payload.Entries)
	}
}

func TestRuntimeStateServiceReportsUnsupportedWithoutProvider(t *testing.T) {
	starter := NewRuntimeStarter[map[string]any, map[string]any, map[string]any](
		"push-only", "Push only", "", "", 100,
	)
	payload, err := starter.State.Response(context.Background(), true, "request-2")
	if err != nil {
		t.Fatal(err)
	}
	if payload.Refresh == nil || payload.Refresh.Status != "unsupported" || payload.Refresh.Performed {
		t.Fatalf("unexpected refresh receipt: %#v", payload.Refresh)
	}
}
