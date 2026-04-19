package runtimekit

import (
	"sync"
	"time"
)

// StateSnapshot holds the latest known runtime state for one entry.
type StateSnapshot[TState any] struct {
	DeviceID    string `json:"device_id"`
	State       TState `json:"state"`
	LastUpdated string `json:"last_updated"`
}

// RuntimeRegistry is a small in-memory registry for active runtime working state.
type RuntimeRegistry[TEntry any, TState any, TEvent any] struct {
	mu              sync.RWMutex
	entries         map[string]TEntry
	stateSnapshots  map[string]StateSnapshot[TState]
	recentEvents    []TEvent
	maxRecentEvents int
}

// NewRuntimeRegistry creates a new in-memory registry.
func NewRuntimeRegistry[TEntry any, TState any, TEvent any](maxRecentEvents int) *RuntimeRegistry[TEntry, TState, TEvent] {
	if maxRecentEvents <= 0 {
		maxRecentEvents = 100
	}
	return &RuntimeRegistry[TEntry, TState, TEvent]{
		entries:         map[string]TEntry{},
		stateSnapshots:  map[string]StateSnapshot[TState]{},
		recentEvents:    []TEvent{},
		maxRecentEvents: maxRecentEvents,
	}
}

// Get returns one entry by id.
func (r *RuntimeRegistry[TEntry, TState, TEvent]) Get(entryID string) (TEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.entries[entryID]
	return entry, ok
}

// Set upserts one entry by id.
func (r *RuntimeRegistry[TEntry, TState, TEvent]) Set(entryID string, entry TEntry) TEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[entryID] = entry
	return entry
}

// Remove deletes one entry and its last state snapshot.
func (r *RuntimeRegistry[TEntry, TState, TEvent]) Remove(entryID string) (TEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.stateSnapshots, entryID)
	entry, ok := r.entries[entryID]
	if ok {
		delete(r.entries, entryID)
	}
	return entry, ok
}

// IDs returns the active entry ids.
func (r *RuntimeRegistry[TEntry, TState, TEvent]) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.entries))
	for id := range r.entries {
		ids = append(ids, id)
	}
	return ids
}

// PrimaryEntry returns the first active entry, if any.
func (r *RuntimeRegistry[TEntry, TState, TEvent]) PrimaryEntry() (TEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, entry := range r.entries {
		return entry, true
	}
	var zero TEntry
	return zero, false
}

// UpdateState stores the latest known state snapshot for one entry id.
func (r *RuntimeRegistry[TEntry, TState, TEvent]) UpdateState(entryID string, state TState, deviceID ...string) StateSnapshot[TState] {
	r.mu.Lock()
	defer r.mu.Unlock()
	resolvedDeviceID := entryID
	if len(deviceID) > 0 && deviceID[0] != "" {
		resolvedDeviceID = deviceID[0]
	}
	snapshot := StateSnapshot[TState]{
		DeviceID:    resolvedDeviceID,
		State:       state,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}
	r.stateSnapshots[entryID] = snapshot
	return snapshot
}

// AppendEvent stores one recent local runtime event.
func (r *RuntimeRegistry[TEntry, TState, TEvent]) AppendEvent(event TEvent) TEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recentEvents = append(r.recentEvents, event)
	if len(r.recentEvents) > r.maxRecentEvents {
		r.recentEvents = r.recentEvents[len(r.recentEvents)-r.maxRecentEvents:]
	}
	return event
}

// EntriesSnapshot returns a copy of active entries.
func (r *RuntimeRegistry[TEntry, TState, TEvent]) EntriesSnapshot() map[string]TEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]TEntry, len(r.entries))
	for key, value := range r.entries {
		out[key] = value
	}
	return out
}

// StateSnapshots returns a copy of latest state snapshots.
func (r *RuntimeRegistry[TEntry, TState, TEvent]) StateSnapshots() map[string]StateSnapshot[TState] {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]StateSnapshot[TState], len(r.stateSnapshots))
	for key, value := range r.stateSnapshots {
		out[key] = value
	}
	return out
}

// RecentEvents returns a copy of recent local runtime events.
func (r *RuntimeRegistry[TEntry, TState, TEvent]) RecentEvents() []TEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]TEvent, len(r.recentEvents))
	copy(out, r.recentEvents)
	return out
}
