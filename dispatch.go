package runtimekit

import "time"

// ScheduleTelemetryDelivery dispatches one telemetry delivery goroutine.
func ScheduleTelemetryDelivery(
	processState *RuntimeProcessState,
	telemetryClient *TelemetryClient,
	authContext *RuntimeAuthContext,
	payload TelemetryPayload,
) {
	if payload.Timestamp == "" {
		payload.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	CreateTrackedTask(processState, func() {
		_ = telemetryClient.SendMetrics(authContext, payload)
	})
}

// ScheduleEventDelivery dispatches one event delivery goroutine.
func ScheduleEventDelivery(
	processState *RuntimeProcessState,
	eventClient *EventClient,
	authContext *RuntimeAuthContext,
	payload CoreEventPayload,
) {
	payload = BuildCoreEventPayload(payload)
	CreateTrackedTask(processState, func() {
		_ = eventClient.SendEvent(authContext, payload)
	})
}
