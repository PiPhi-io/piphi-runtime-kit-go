package runtimekit

// RuntimeContext bundles shared runtime auth and process state.
type RuntimeContext struct {
	Auth         *RuntimeAuthContext
	ProcessState *RuntimeProcessState
}

// NewRuntimeContext creates a fresh runtime context.
func NewRuntimeContext() *RuntimeContext {
	return &RuntimeContext{
		Auth:         &RuntimeAuthContext{},
		ProcessState: NewRuntimeProcessState(),
	}
}
