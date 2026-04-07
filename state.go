package runtimekit

import (
	"net/http"
	"sync"
)

// RuntimeProcessState stores shared runtime process state across routes and jobs.
type RuntimeProcessState struct {
	mu                sync.RWMutex
	coreHTTPClient    *http.Client
	coreBaseURL       string
	currentGeneration *int
	pendingTaskCount  int
	taskWaitGroup     sync.WaitGroup
}

// NewRuntimeProcessState creates a new process state container.
func NewRuntimeProcessState() *RuntimeProcessState {
	return &RuntimeProcessState{
		coreBaseURL: "http://127.0.0.1:31419",
	}
}

// BindCoreHTTPClient stores the Core HTTP client used for outbound runtime calls.
func (s *RuntimeProcessState) BindCoreHTTPClient(client *http.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.coreHTTPClient = client
}

// CoreHTTPClient returns the currently bound Core HTTP client.
func (s *RuntimeProcessState) CoreHTTPClient() *http.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.coreHTTPClient
}

// SetCoreBaseURL updates the outbound Core base URL.
func (s *RuntimeProcessState) SetCoreBaseURL(baseURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if baseURL != "" {
		s.coreBaseURL = baseURL
	}
}

// CoreBaseURL returns the effective Core base URL.
func (s *RuntimeProcessState) CoreBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.coreBaseURL
}

// SetCurrentGeneration stores the most recent config generation.
func (s *RuntimeProcessState) SetCurrentGeneration(generation *int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentGeneration = generation
}

// CurrentGeneration returns the most recent config generation.
func (s *RuntimeProcessState) CurrentGeneration() *int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentGeneration
}

func (s *RuntimeProcessState) incrementPendingTasks() {
	s.mu.Lock()
	s.pendingTaskCount++
	s.mu.Unlock()
	s.taskWaitGroup.Add(1)
}

func (s *RuntimeProcessState) decrementPendingTasks() {
	s.mu.Lock()
	if s.pendingTaskCount > 0 {
		s.pendingTaskCount--
	}
	s.mu.Unlock()
	s.taskWaitGroup.Done()
}

// PendingTaskCount returns the number of tracked background goroutines.
func (s *RuntimeProcessState) PendingTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingTaskCount
}

// WaitPendingTasks blocks until tracked background tasks are finished.
func (s *RuntimeProcessState) WaitPendingTasks() {
	s.taskWaitGroup.Wait()
}
