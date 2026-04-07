package runtimekit

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestTelemetryClientReturnsCoreRouteNotFoundError(t *testing.T) {
	processState := NewRuntimeProcessState()
	processState.BindCoreHTTPClient(&http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(`{"detail":"Not Found"}`)),
				Header:     make(http.Header),
				Request:    request,
			}, nil
		}),
	})

	client := NewTelemetryClient(processState, "http://core.test", 5*time.Second)
	auth := &RuntimeAuthContext{}
	auth.Update("container-1", "secret-token")

	err := client.SendMetrics(auth, TelemetryPayload{
		DeviceID: "device-1",
		Metrics:  map[string]any{"temperature": 72.1},
	})
	if err == nil {
		t.Fatal("expected route-not-found error")
	}

	var routeErr *CoreRouteNotFoundError
	if !errors.As(err, &routeErr) {
		t.Fatalf("expected CoreRouteNotFoundError, got %T", err)
	}
	if routeErr.Operation != "telemetry_delivery" {
		t.Fatalf("unexpected operation %q", routeErr.Operation)
	}
}

func TestEventClientReturnsCoreServerError(t *testing.T) {
	processState := NewRuntimeProcessState()
	processState.BindCoreHTTPClient(&http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`{"detail":"boom"}`)),
				Header:     make(http.Header),
				Request:    request,
			}, nil
		}),
	})

	client := NewEventClient(processState, "http://core.test/api/v2", 5*time.Second)
	auth := &RuntimeAuthContext{}
	auth.Update("container-1", "secret-token")

	err := client.SendEvent(auth, CoreEventPayload{
		EventType:     "device.turned_on",
		IntegrationID: "tp-link-kasa",
		ConfigID:      "plug-1",
		ContainerID:   "container-1",
		DeviceID:      "plug-1",
	})
	if err == nil {
		t.Fatal("expected server error")
	}

	var serverErr *CoreServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("expected CoreServerError, got %T", err)
	}
	if !serverErr.Retryable {
		t.Fatal("expected server error to be retryable")
	}
}
