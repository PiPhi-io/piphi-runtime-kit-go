# Minimal net/http Runtime Example

This example shows the intended shape of a small PiPhi runtime integration
using the Go runtime kit with the standard library `net/http`.

It demonstrates:

- syncing request auth with `adapters.SyncRuntimeAuthFromRequest(...)`
- using `RuntimeContext` and `RuntimeRegistry` for runtime state
- standard config apply and remove responses
- standard health and diagnostics responses
- standard discovery and local event responses
- queued telemetry delivery back to PiPhi Core

Key routes in the example:

- `GET /health`
- `GET /diagnostics`
- `POST /discover`
- `POST /config`
- `POST /deconfigure`
- `GET /state`
- `POST /events/example`
- `GET /events`
- `POST /telemetry/example`

This example is intentionally illustrative and small. It is meant to show the
SDK shape, not to be a production-ready integration by itself.
