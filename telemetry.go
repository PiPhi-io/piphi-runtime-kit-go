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

// TelemetryReading is one validated native metric observation.
type TelemetryReading struct {
	Metric string
	Value  any
	Unit   string
}

// BuildTelemetryMaps validates readings and converts them to the Core wire shape.
func BuildTelemetryMaps(readings []TelemetryReading) (map[string]any, map[string]any, error) {
	metrics := map[string]any{}
	units := map[string]any{}
	for _, reading := range readings {
		metric := strings.TrimSpace(reading.Metric)
		if metric == "" {
			return nil, nil, fmt.Errorf("telemetry reading metric cannot be empty")
		}
		if _, exists := metrics[metric]; exists {
			return nil, nil, fmt.Errorf("duplicate telemetry metric: %s", metric)
		}
		switch reading.Value.(type) {
		case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string:
		default:
			return nil, nil, fmt.Errorf("telemetry reading %s has unsupported value type %T", metric, reading.Value)
		}
		metrics[metric] = reading.Value
		if strings.TrimSpace(reading.Unit) != "" {
			units[metric] = reading.Unit
		}
	}
	if len(metrics) == 0 {
		return nil, nil, fmt.Errorf("at least one telemetry reading is required")
	}
	return metrics, units, nil
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
	if strings.TrimSpace(outbound.Timestamp) == "" {
		outbound.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

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

// SendDeviceReadings sends typed readings using one validated device identity.
func (c *TelemetryClient) SendDeviceReadings(authContext *RuntimeAuthContext, device RuntimeIdentity, readings []TelemetryReading, timestamp string) error {
	if strings.TrimSpace(device.ConfigID) == "" || strings.TrimSpace(device.DeviceID) == "" {
		return RuntimeIdentityError{Missing: []string{"config_id", "device_id"}}
	}
	metrics, units, err := BuildTelemetryMaps(readings)
	if err != nil {
		return err
	}
	return c.SendMetrics(authContext, TelemetryPayload{
		DeviceID: device.DeviceID, ConfigID: device.ConfigID, ContainerID: device.ContainerID,
		IntegrationID: device.IntegrationID, Metrics: metrics, Units: units, Timestamp: timestamp,
	})
}
