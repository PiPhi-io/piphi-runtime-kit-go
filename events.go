package runtimekit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// BuildCoreEventPayload returns the standard Core-bound event payload.
func BuildCoreEventPayload(values CoreEventPayload) CoreEventPayload {
	return values
}

// BuildEventIngestResponse returns the standard local event-ingest response.
func BuildEventIngestResponse[TEvent any](event TEvent) IntegrationEventIngestResponse[TEvent] {
	return IntegrationEventIngestResponse[TEvent]{
		OK:    true,
		Event: event,
	}
}

// BuildEventListResponse returns the standard local event-list response.
func BuildEventListResponse[TEvent any](events []TEvent) IntegrationEventListResponse[TEvent] {
	return IntegrationEventListResponse[TEvent]{
		Events: events,
	}
}

// BuildLocalEventRecord returns a local runtime event record with a timestamp.
func BuildLocalEventRecord(event map[string]any) map[string]any {
	record := map[string]any{
		"received_at": time.Now().UTC().Format(time.RFC3339),
	}
	for key, value := range event {
		record[key] = value
	}
	return record
}

// EventClient sends runtime events back to PiPhi Core.
type EventClient struct {
	processState *RuntimeProcessState
	coreBaseURL  string
	timeout      time.Duration
}

// NewEventClient creates an outbound event delivery client.
func NewEventClient(processState *RuntimeProcessState, coreBaseURL string, timeout time.Duration) *EventClient {
	if processState == nil {
		processState = NewRuntimeProcessState()
	}
	if strings.TrimSpace(coreBaseURL) == "" {
		coreBaseURL = processState.CoreBaseURL()
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &EventClient{
		processState: processState,
		coreBaseURL:  coreBaseURL,
		timeout:      timeout,
	}
}

// SendEvent sends one Core event payload to PiPhi Core.
func (c *EventClient) SendEvent(authContext *RuntimeAuthContext, event CoreEventPayload) error {
	resolvedContainerID, internalToken := authContext.Resolve(event.ContainerID)
	if strings.TrimSpace(resolvedContainerID) == "" || strings.TrimSpace(internalToken) == "" {
		return fmt.Errorf("missing runtime auth context for event delivery")
	}

	outbound := event
	outbound.ContainerID = resolvedContainerID

	body, err := json.Marshal(outbound)
	if err != nil {
		return err
	}

	client := c.processState.CoreHTTPClient()
	if client == nil {
		client = &http.Client{Timeout: c.timeout}
	}

	eventsURL := normalizeCoreBaseURL(c.coreBaseURL) + "/api/v2/events/ingest"
	request, err := http.NewRequest(http.MethodPost, eventsURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	for key, values := range BuildRuntimeAuthHeaders(resolvedContainerID, internalToken) {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return classifyCoreDeliveryError("event_delivery", eventsURL, c.timeout, nil, err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return classifyCoreDeliveryError("event_delivery", eventsURL, c.timeout, response, nil)
	}
	return nil
}
