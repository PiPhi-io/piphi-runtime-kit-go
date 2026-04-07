package runtimekit

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"
)

// CoreDeliveryError is the base typed error for failed runtime-to-Core delivery.
type CoreDeliveryError struct {
	Operation  string
	URL        string
	Message    string
	Retryable  bool
	StatusCode int
	Timeout    time.Duration
}

func (e *CoreDeliveryError) Error() string {
	parts := []string{
		e.Message,
		fmt.Sprintf("operation=%s", e.Operation),
		fmt.Sprintf("url=%s", e.URL),
	}
	if e.StatusCode > 0 {
		parts = append(parts, fmt.Sprintf("status_code=%d", e.StatusCode))
	}
	if e.Timeout > 0 {
		parts = append(parts, fmt.Sprintf("timeout=%s", e.Timeout))
	}
	parts = append(parts, fmt.Sprintf("retryable=%t", e.Retryable))
	return strings.Join(parts, " ")
}

type CoreUnavailableError struct{ *CoreDeliveryError }
type CoreTimeoutError struct{ *CoreDeliveryError }
type CoreRouteNotFoundError struct{ *CoreDeliveryError }
type CoreAuthError struct{ *CoreDeliveryError }
type CoreServerError struct{ *CoreDeliveryError }
type CoreUnexpectedResponseError struct{ *CoreDeliveryError }

func normalizeCoreBaseURL(baseURL string) string {
	normalized := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return strings.TrimSuffix(normalized, "/api/v2")
}

func classifyCoreDeliveryError(operation, url string, timeout time.Duration, response *http.Response, err error) error {
	if response != nil {
		base := &CoreDeliveryError{
			Operation:  operation,
			URL:        url,
			Retryable:  false,
			StatusCode: response.StatusCode,
			Timeout:    timeout,
		}
		switch {
		case response.StatusCode == http.StatusNotFound:
			base.Message = "PiPhi Core route is not available"
			return &CoreRouteNotFoundError{CoreDeliveryError: base}
		case response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden:
			base.Message = "PiPhi Core rejected the runtime authentication headers"
			return &CoreAuthError{CoreDeliveryError: base}
		case response.StatusCode >= 500:
			base.Message = "PiPhi Core failed while processing the delivery request"
			base.Retryable = true
			return &CoreServerError{CoreDeliveryError: base}
		default:
			base.Message = "PiPhi Core returned an unexpected non-success response"
			return &CoreUnexpectedResponseError{CoreDeliveryError: base}
		}
	}

	if err == nil {
		return &CoreUnexpectedResponseError{CoreDeliveryError: &CoreDeliveryError{
			Operation: operation,
			URL:       url,
			Message:   "Unknown Core delivery failure",
			Retryable: false,
			Timeout:   timeout,
		}}
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return &CoreTimeoutError{CoreDeliveryError: &CoreDeliveryError{
			Operation: operation,
			URL:       url,
			Message:   "PiPhi Core did not respond before the request timeout",
			Retryable: true,
			Timeout:   timeout,
		}}
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) || errors.Is(err, syscall.ECONNREFUSED) {
		return &CoreUnavailableError{CoreDeliveryError: &CoreDeliveryError{
			Operation: operation,
			URL:       url,
			Message:   "PiPhi Core is unreachable",
			Retryable: true,
			Timeout:   timeout,
		}}
	}

	return &CoreUnexpectedResponseError{CoreDeliveryError: &CoreDeliveryError{
		Operation: operation,
		URL:       url,
		Message:   err.Error(),
		Retryable: false,
		Timeout:   timeout,
	}}
}
