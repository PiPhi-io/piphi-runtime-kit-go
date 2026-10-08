package adapters

import (
	"strings"

	"github.com/gin-gonic/gin"
	runtimekit "github.com/PiPhi-io/piphi-runtime-kit-go"
)

// SyncRuntimeAuthFromGinContext syncs runtime auth from an incoming Gin request.
func SyncRuntimeAuthFromGinContext(
	runtime *runtimekit.RuntimeContext,
	context *gin.Context,
	payloadContainerID string,
) runtimekit.RuntimeAuthHeaders {
	return runtime.Auth.SyncFromHeaders(context.Request.Header, payloadContainerID)
}

// FormatRuntimeAuthSyncLogFromGinContext formats a standard request-auth log line.
func FormatRuntimeAuthSyncLogFromGinContext(
	context *gin.Context,
	payloadContainerID string,
) string {
	return runtimekit.FormatRuntimeAuthSyncLog(
		runtimekit.ExtractRuntimeAuthHeaders(context.Request.Header),
		payloadContainerID,
	)
}

// GetPayloadContainerIDFromGinMap extracts a container id from a JSON-like Gin body map.
func GetPayloadContainerIDFromGinMap(payload map[string]any) string {
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
