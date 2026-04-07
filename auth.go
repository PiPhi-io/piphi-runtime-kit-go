package runtimekit

import (
	"net/http"
	"strings"
	"sync"
)

const (
	// RuntimeContainerIDHeaderName is the shared PiPhi container header.
	RuntimeContainerIDHeaderName = "X-Container-Id"
	// RuntimeInternalTokenHeaderName is the shared PiPhi internal token header.
	RuntimeInternalTokenHeaderName = "X-PiPhi-Integration-Token"
)

// RuntimeAuthHeaders is the normalized request-auth header set.
type RuntimeAuthHeaders struct {
	ContainerID   string
	InternalToken string
}

// MaskToken returns a log-safe representation of a runtime token.
func MaskToken(token string) string {
	normalized := strings.TrimSpace(token)
	if normalized == "" {
		return "missing"
	}
	if len(normalized) <= 10 {
		return "present"
	}
	return normalized[:6] + "..." + normalized[len(normalized)-4:]
}

// ExtractRuntimeAuthHeaders reads PiPhi auth headers from a request header map.
func ExtractRuntimeAuthHeaders(headers http.Header) RuntimeAuthHeaders {
	return RuntimeAuthHeaders{
		ContainerID:   strings.TrimSpace(headers.Get(RuntimeContainerIDHeaderName)),
		InternalToken: strings.TrimSpace(headers.Get(RuntimeInternalTokenHeaderName)),
	}
}

// BuildRuntimeAuthHeaders builds standard outbound Core auth headers.
func BuildRuntimeAuthHeaders(containerID, internalToken string) http.Header {
	headers := make(http.Header)
	if strings.TrimSpace(containerID) != "" {
		headers.Set(RuntimeContainerIDHeaderName, strings.TrimSpace(containerID))
	}
	if strings.TrimSpace(internalToken) != "" {
		headers.Set(RuntimeInternalTokenHeaderName, strings.TrimSpace(internalToken))
	}
	return headers
}

// FormatRuntimeAuthSyncLog builds a consistent auth-sync log line.
func FormatRuntimeAuthSyncLog(parsed RuntimeAuthHeaders, payloadContainerID string) string {
	headerContainerID := parsed.ContainerID
	if headerContainerID == "" {
		headerContainerID = "missing"
	}
	if strings.TrimSpace(payloadContainerID) == "" {
		payloadContainerID = "missing"
	}
	return "runtime_internal_auth " +
		"header_container_id=" + headerContainerID + " " +
		"payload_container_id=" + payloadContainerID + " " +
		"token=" + MaskToken(parsed.InternalToken)
}

// RuntimeAuthContext stores request-derived runtime auth for later background work.
type RuntimeAuthContext struct {
	mu            sync.RWMutex
	containerID   string
	internalToken string
}

// Update stores any non-empty auth fields in the shared runtime auth context.
func (c *RuntimeAuthContext) Update(containerID, internalToken string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if strings.TrimSpace(containerID) != "" {
		c.containerID = strings.TrimSpace(containerID)
	}
	if strings.TrimSpace(internalToken) != "" {
		c.internalToken = strings.TrimSpace(internalToken)
	}
}

// SyncFromHeaders updates the shared context from incoming request headers.
func (c *RuntimeAuthContext) SyncFromHeaders(headers http.Header, payloadContainerID string) RuntimeAuthHeaders {
	parsed := ExtractRuntimeAuthHeaders(headers)
	resolvedContainerID := parsed.ContainerID
	if resolvedContainerID == "" {
		resolvedContainerID = strings.TrimSpace(payloadContainerID)
	}
	c.Update(resolvedContainerID, parsed.InternalToken)
	return parsed
}

// Resolve returns the effective container scope and internal token.
func (c *RuntimeAuthContext) Resolve(containerID string) (string, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resolvedContainerID := strings.TrimSpace(containerID)
	if resolvedContainerID == "" {
		resolvedContainerID = c.containerID
	}
	return resolvedContainerID, c.internalToken
}

// ContainerID returns the current runtime container id.
func (c *RuntimeAuthContext) ContainerID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.containerID
}

// InternalToken returns the current runtime internal token.
func (c *RuntimeAuthContext) InternalToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.internalToken
}
