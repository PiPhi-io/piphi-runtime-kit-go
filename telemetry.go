package runtimekit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// TelemetryClient sends runtime telemetry back to PiPhi Core.
type TelemetryClient struct {
	processState *RuntimeProcessState
	coreBaseURL  string
	timeout      time.Duration
}

// NewTelemetryClient creates a telemetry client for outbound Core calls.
func NewTelemetryClient(processState *RuntimeProcessState, coreBaseURL string, timeout time.Duration) *TelemetryClient {
	if processState == nil {
		processState = NewRuntimeProcessState()
	}
	if strings.TrimSpace(coreBaseURL) == "" {
		coreBaseURL = processState.CoreBaseURL()
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &TelemetryClient{
		processState: processState,
		coreBaseURL:  coreBaseURL,
		timeout:      timeout,
	}
}

// SendMetrics sends one telemetry payload to PiPhi Core.
func (c *TelemetryClient) SendMetrics(authContext *RuntimeAuthContext, payload TelemetryPayload) error {
	resolvedContainerID, internalToken := authContext.Resolve(payload.ContainerID)
	if strings.TrimSpace(resolvedContainerID) == "" || strings.TrimSpace(internalToken) == "" {
		return fmt.Errorf("missing runtime auth context for telemetry delivery")
	}

	outbound := payload
	outbound.ContainerID = resolvedContainerID

	body, err := json.Marshal(outbound)
	if err != nil {
		return err
	}

	client := c.processState.CoreHTTPClient()
	if client == nil {
		client = &http.Client{Timeout: c.timeout}
	}

	telemetryURL := normalizeCoreBaseURL(c.coreBaseURL) + "/api/v2/integrations/telemetry"
	request, err := http.NewRequest(http.MethodPost, telemetryURL, bytes.NewReader(body))
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
		return classifyCoreDeliveryError("telemetry_delivery", telemetryURL, c.timeout, nil, err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		return classifyCoreDeliveryError("telemetry_delivery", telemetryURL, c.timeout, response, nil)
	}
	return nil
}
