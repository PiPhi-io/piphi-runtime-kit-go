package runtimekit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// StateReader performs one real read from the integration's upstream source.
// It publishes normalized snapshots through RuntimeStateService.Publish.
type StateReader func(context.Context) error

// StateRefreshReceipt proves whether an on-demand upstream read occurred.
type StateRefreshReceipt struct {
	RequestID  string `json:"request_id"`
	Performed  bool   `json:"performed"`
	Status     string `json:"status"`
	ObservedAt string `json:"observed_at,omitempty"`
	Source     string `json:"source,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

// RuntimeStateResponse is the standard /state response payload.
type RuntimeStateResponse[TState any] struct {
	Entries map[string]StateSnapshot[TState] `json:"entries"`
	Refresh *StateRefreshReceipt             `json:"refresh,omitempty"`
}

// RuntimeStateService is the concise state API shared by every runtime.
type RuntimeStateService[TEntry any, TState any, TEvent any] struct {
	registry *RuntimeRegistry[TEntry, TState, TEvent]
	reader   StateReader
	source   string
	timeout  time.Duration
}

// NewRuntimeStateService binds state helpers to one runtime registry.
func NewRuntimeStateService[TEntry any, TState any, TEvent any](
	registry *RuntimeRegistry[TEntry, TState, TEvent],
) *RuntimeStateService[TEntry, TState, TEvent] {
	return &RuntimeStateService[TEntry, TState, TEvent]{
		registry: registry,
		timeout:  10 * time.Second,
	}
}

// Provide registers the one function that performs a real upstream read.
func (s *RuntimeStateService[TEntry, TState, TEvent]) Provide(
	reader StateReader,
	source string,
	timeout ...time.Duration,
) error {
	normalizedSource := strings.TrimSpace(source)
	if normalizedSource == "" {
		return errors.New("source must not be empty")
	}
	resolvedTimeout := 10 * time.Second
	if len(timeout) > 0 {
		resolvedTimeout = timeout[0]
	}
	if resolvedTimeout <= 0 {
		return errors.New("timeout must be positive")
	}
	s.reader = reader
	s.source = normalizedSource
	s.timeout = resolvedTimeout
	return nil
}

// Publish stores one normalized state snapshot.
func (s *RuntimeStateService[TEntry, TState, TEvent]) Publish(
	entryID string,
	state TState,
	deviceID ...string,
) StateSnapshot[TState] {
	return s.registry.UpdateState(entryID, state, deviceID...)
}

// Get returns one cached state snapshot.
func (s *RuntimeStateService[TEntry, TState, TEvent]) Get(entryID string) (StateSnapshot[TState], bool) {
	return s.registry.GetState(entryID)
}

// Response builds the standard state payload and optional refresh receipt.
func (s *RuntimeStateService[TEntry, TState, TEvent]) Response(
	ctx context.Context,
	refresh bool,
	refreshRequestID string,
) (RuntimeStateResponse[TState], error) {
	response := RuntimeStateResponse[TState]{Entries: s.registry.StateSnapshots()}
	if !refresh {
		return response, nil
	}
	requestID := strings.TrimSpace(refreshRequestID)
	if requestID == "" {
		return response, errors.New("refreshRequestID is required when refresh is true")
	}
	receipt := s.refresh(ctx, requestID)
	response.Entries = s.registry.StateSnapshots()
	response.Refresh = &receipt
	return response, nil
}

func (s *RuntimeStateService[TEntry, TState, TEvent]) refresh(
	ctx context.Context,
	requestID string,
) StateRefreshReceipt {
	if s.reader == nil {
		return StateRefreshReceipt{
			RequestID: requestID,
			Performed: false,
			Status: "unsupported",
			Message: "This integration does not support on-demand state refresh.",
		}
	}

	refreshContext, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- s.reader(refreshContext) }()

	select {
	case <-refreshContext.Done():
		return StateRefreshReceipt{
			RequestID: requestID,
			Performed: false,
			Status: "failed",
			Source: s.source,
			Error: "Upstream state refresh timed out.",
		}
	case err := <-result:
		if err != nil {
			return StateRefreshReceipt{
				RequestID: requestID,
				Performed: false,
				Status: "failed",
				Source: s.source,
				Error: fmt.Sprintf("Upstream state refresh failed: %T", err),
			}
		}
	}

	return StateRefreshReceipt{
		RequestID: requestID,
		Performed: true,
		Status: "refreshed",
		ObservedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Source: s.source,
	}
}
