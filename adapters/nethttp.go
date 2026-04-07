package adapters

import (
	"net/http"
	"strings"

	runtimekit "github.com/piphi-network/piphi-runtime-kit-go"
)

// GetPayloadContainerID extracts a container id from a JSON-like payload map.
func GetPayloadContainerID(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	if value, ok := payload["container_id"].(string); ok {
		return strings.TrimSpace(value)
	}
	if value, ok := payload["containerId"].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

// SyncRuntimeAuthFromRequest syncs runtime auth from an incoming net/http request.
func SyncRuntimeAuthFromRequest(
	runtime *runtimekit.RuntimeContext,
	request *http.Request,
	payloadContainerID string,
) runtimekit.RuntimeAuthHeaders {
	return runtime.Auth.SyncFromHeaders(request.Header, payloadContainerID)
}

// FormatRuntimeAuthSyncLogFromRequest formats a standard request-auth log line.
func FormatRuntimeAuthSyncLogFromRequest(
	request *http.Request,
	payloadContainerID string,
) string {
	return runtimekit.FormatRuntimeAuthSyncLog(
		runtimekit.ExtractRuntimeAuthHeaders(request.Header),
		payloadContainerID,
	)
}
