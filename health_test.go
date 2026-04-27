package runtimekit

import "testing"

func TestBuildRuntimeHealthResponseIncludesConfigGeneration(t *testing.T) {
	runtime := NewRuntimeContext()
	generation := 42
	runtime.ProcessState.SetCurrentGeneration(&generation)

	response := BuildRuntimeHealthResponse(runtime, nil, nil)

	if response.CurrentGeneration == nil || *response.CurrentGeneration != generation {
		t.Fatalf("unexpected current generation: %#v", response.CurrentGeneration)
	}
	if response.ConfigGeneration == nil || *response.ConfigGeneration != generation {
		t.Fatalf("unexpected config generation: %#v", response.ConfigGeneration)
	}
}

func TestBuildRuntimeDiagnosticsResponseIncludesConfigGeneration(t *testing.T) {
	runtime := NewRuntimeContext()
	generation := 43
	runtime.ProcessState.SetCurrentGeneration(&generation)

	response := BuildRuntimeDiagnosticsResponse(runtime, nil, nil)

	if response.CurrentGeneration == nil || *response.CurrentGeneration != generation {
		t.Fatalf("unexpected current generation: %#v", response.CurrentGeneration)
	}
	if response.ConfigGeneration == nil || *response.ConfigGeneration != generation {
		t.Fatalf("unexpected config generation: %#v", response.ConfigGeneration)
	}
}
