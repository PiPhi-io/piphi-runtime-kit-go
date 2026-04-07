# Minimal Gin Runtime Example

This example shows the Go SDK's thin Gin adapter path.

It is meant to teach in two layers:

- a beginner path for a first successful PiPhi runtime in Gin
- an advanced path for targeted per-device event and telemetry flows

## What it demonstrates

- starting from `NewRuntimeStarter(...)`
- syncing request auth with `adapters.SyncRuntimeAuthFromGinContext(...)`
- using `RuntimeRegistry` for active runtime state
- standard config apply and remove responses
- standard health and diagnostics responses
- standard discovery and local event responses
- queued telemetry delivery back to PiPhi Core
- targeted event and telemetry examples for a specific configured device

## Beginner route flow

Start with these routes first:

- `GET /health`
- `GET /diagnostics`
- `POST /discover`
- `POST /config`
- `GET /state`
- `POST /events/example`
- `POST /telemetry/example`

## Advanced route flow

These routes are closer to a real integration:

- `POST /deconfigure/:configId`
- `POST /events/device/:configId/example`
- `POST /telemetry/device/:configId/example`

They show how to:

- target a specific configured device
- keep `ConfigID` and `DeviceID` straight
- record config lifecycle events
- queue telemetry for a chosen device instead of always using the primary entry

## Suggested reading order

1. read the starter creation
2. read `handleConfig`
3. read the Gin auth adapter usage
4. read `handleTelemetryExample`
5. read the device-targeted routes
