package runtimekit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCameraCapabilityValidation(t *testing.T) {
	declaration := CameraCapabilityDeclaration{ContractVersion: CameraMediaContractVersion, MessageType: "camera.capability", CapabilityID: "camera_stream", Transports: []string{"webrtc"}, Endpoints: CameraCapabilityEndpoints{Source: "/v1/cameras/{config_id}/source"}}
	if err := declaration.Validate(); err != nil {
		t.Fatal(err)
	}
	declaration.Endpoints.Source = "/v1/../secret"
	if err := declaration.Validate(); err == nil {
		t.Fatal("expected traversal rejection")
	}
}

func TestCameraSourceValidationAndJSON(t *testing.T) {
	source := CameraSourceDescriptor{ContractVersion: CameraMediaContractVersion, MessageType: "camera.source", Audience: "core-sidecar", CameraID: "front-door", SourceURL: "rtsps://user:secret@camera.local/live", SourceProtocol: "rtsps"}
	if err := source.Validate(); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"audience":"core-sidecar"`) {
		t.Fatal("missing private audience")
	}
	source.SourceProtocol = "rtsp"
	if err := source.Validate(); err == nil {
		t.Fatal("expected protocol mismatch")
	}
}
