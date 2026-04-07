package runtimekit

// ScheduleTelemetryDelivery dispatches one telemetry delivery goroutine.
func ScheduleTelemetryDelivery(
	processState *RuntimeProcessState,
	telemetryClient *TelemetryClient,
	authContext *RuntimeAuthContext,
	payload TelemetryPayload,
) {
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
	CreateTrackedTask(processState, func() {
		_ = eventClient.SendEvent(authContext, payload)
	})
}
