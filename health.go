package runtimekit

// BuildRuntimeHealthResponse returns the standard runtime health payload.
func BuildRuntimeHealthResponse(runtime *RuntimeContext, integration map[string]any, metadata map[string]any) RuntimeHealthResponse {
	containerID, internalToken := runtime.Auth.Resolve("")
	return RuntimeHealthResponse{
		OK:                 true,
		Integration:        integration,
		RuntimeAuthPresent: containerID != "" && internalToken != "",
		CoreClientBound:    runtime.ProcessState.CoreHTTPClient() != nil,
		PendingTaskCount:   runtime.ProcessState.PendingTaskCount(),
		CurrentGeneration:  runtime.ProcessState.CurrentGeneration(),
		ConfigGeneration:   runtime.ProcessState.CurrentGeneration(),
		Metadata:           metadata,
	}
}

// BuildRuntimeDiagnosticsResponse returns the standard diagnostics payload.
func BuildRuntimeDiagnosticsResponse(runtime *RuntimeContext, integration map[string]any, diagnostics map[string]any) RuntimeDiagnosticsResponse {
	containerID, internalToken := runtime.Auth.Resolve("")
	return RuntimeDiagnosticsResponse{
		OK:                 true,
		Integration:        integration,
		RuntimeAuthPresent: containerID != "" && internalToken != "",
		CoreClientBound:    runtime.ProcessState.CoreHTTPClient() != nil,
		PendingTaskCount:   runtime.ProcessState.PendingTaskCount(),
		CurrentGeneration:  runtime.ProcessState.CurrentGeneration(),
		ConfigGeneration:   runtime.ProcessState.CurrentGeneration(),
		Diagnostics:        diagnostics,
	}
}
