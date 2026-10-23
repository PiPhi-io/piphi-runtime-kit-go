package runtimekit

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const CameraMediaContractVersion = "camera.media.v1"
const CameraMaxSDPBytes = 256 * 1024

type CameraCapabilityEndpoints struct {
	Source      string `json:"source,omitempty"`
	Snapshot    string `json:"snapshot,omitempty"`
	Events      string `json:"events,omitempty"`
	HLS         string `json:"hls,omitempty"`
	WebRTC      string `json:"webrtc,omitempty"`
	WebRTCClose string `json:"webrtc_close,omitempty"`
}

type CameraCapabilityDeclaration struct {
	ContractVersion string                    `json:"contract_version"`
	MessageType     string                    `json:"message_type"`
	CapabilityID    string                    `json:"capability_id"`
	Transports      []string                  `json:"transports"`
	AudioAvailable  bool                      `json:"audio_available,omitempty"`
	RefreshSeconds  int                       `json:"refresh_seconds,omitempty"`
	Endpoints       CameraCapabilityEndpoints `json:"endpoints"`
}

type CameraSourceDescriptor struct {
	ContractVersion       string `json:"contract_version"`
	MessageType           string `json:"message_type"`
	Audience              string `json:"audience"`
	CameraID              string `json:"camera_id"`
	SourceURL             string `json:"source_url"`
	SourceProtocol        string `json:"source_protocol"`
	PreferredTransport    string `json:"preferred_transport,omitempty"`
	ConnectTimeoutSeconds int    `json:"connect_timeout_seconds,omitempty"`
}

type CameraWebRTCOffer struct {
	Type         string `json:"type"`
	SDP          string `json:"sdp"`
	IncludeAudio bool   `json:"include_audio,omitempty"`
}

type CameraSidecarSessionOpen struct {
	ContractVersion    string                 `json:"contract_version"`
	MessageType        string                 `json:"message_type"`
	Audience           string                 `json:"audience"`
	Source             CameraSourceDescriptor `json:"source"`
	Offer              CameraWebRTCOffer      `json:"offer"`
	IdleTimeoutSeconds int                    `json:"idle_timeout_seconds,omitempty"`
}

func (declaration CameraCapabilityDeclaration) Validate() error {
	if declaration.ContractVersion != CameraMediaContractVersion || declaration.MessageType != "camera.capability" {
		return errors.New("camera capability must use camera.media.v1")
	}
	if declaration.CapabilityID == "" || len(declaration.CapabilityID) > 300 || len(declaration.Transports) == 0 {
		return errors.New("camera capability id and transports are required")
	}
	seen := map[string]bool{}
	for _, transport := range declaration.Transports {
		if transport != "snapshot" && transport != "webrtc" && transport != "hls" {
			return fmt.Errorf("unsupported camera transport %q", transport)
		}
		if seen[transport] {
			return errors.New("camera transports must be unique")
		}
		seen[transport] = true
	}
	paths := []string{declaration.Endpoints.Source, declaration.Endpoints.Snapshot, declaration.Endpoints.Events, declaration.Endpoints.HLS, declaration.Endpoints.WebRTC, declaration.Endpoints.WebRTCClose}
	hasEndpoint := false
	for _, path := range paths {
		if path == "" {
			continue
		}
		hasEndpoint = true
		if !safeCameraEndpoint(path) {
			return errors.New("camera endpoints must be safe path-only values")
		}
	}
	if !hasEndpoint {
		return errors.New("at least one camera endpoint is required")
	}
	return nil
}

func (source CameraSourceDescriptor) Validate() error {
	if source.ContractVersion != CameraMediaContractVersion || source.MessageType != "camera.source" || source.Audience != "core-sidecar" {
		return errors.New("invalid private camera source envelope")
	}
	if len(source.SourceURL) == 0 || len(source.SourceURL) > 4096 || strings.ContainsAny(source.SourceURL, "\r\n\x00") {
		return errors.New("invalid camera source URL")
	}
	parsed, err := url.Parse(source.SourceURL)
	if err != nil || parsed.Hostname() == "" || parsed.Scheme != source.SourceProtocol {
		return errors.New("camera source URL and protocol must match")
	}
	switch source.SourceProtocol {
	case "rtsp", "rtsps", "http", "https":
	default:
		return errors.New("unsupported camera source protocol")
	}
	return nil
}

func safeCameraEndpoint(path string) bool {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "\\") {
		return false
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == "." || segment == ".." {
			return false
		}
	}
	return true
}
