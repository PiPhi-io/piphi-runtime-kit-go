package runtimekit

import (
	"encoding/json"
	"sort"
	"strings"
)

// NormalizeDiscoveryInputs trims strings and drops empty values.
func NormalizeDiscoveryInputs(inputs map[string]any) map[string]any {
	normalized := map[string]any{}
	for key, value := range inputs {
		switch typed := value.(type) {
		case string:
			trimmed := strings.TrimSpace(typed)
			if trimmed != "" {
				normalized[key] = trimmed
			}
		case nil:
			continue
		default:
			normalized[key] = value
		}
	}
	return normalized
}

// BuildDiscoveryResponse returns the standard discovery response payload.
func BuildDiscoveryResponse[TDevice any](devices []TDevice) IntegrationDiscoveryResponse[TDevice] {
	return IntegrationDiscoveryResponse[TDevice]{
		Devices: devices,
	}
}

// FormatDiscoveryAttemptLog builds a concise, safe discovery attempt log line.
func FormatDiscoveryAttemptLog(inputs map[string]any) string {
	inputKeys := make([]string, 0, len(inputs))
	for key := range inputs {
		inputKeys = append(inputKeys, key)
	}
	sort.Strings(inputKeys)
	payload, _ := json.Marshal(inputKeys)
	_, usesUsername := inputs["username"]
	_, usesEmail := inputs["email"]
	_, usesPassword := inputs["password"]
	return "discovery_attempt " +
		"input_keys=" + string(payload) + " " +
		"uses_username=" + boolString(usesUsername) + " " +
		"uses_email=" + boolString(usesEmail) + " " +
		"uses_password=" + boolString(usesPassword)
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
