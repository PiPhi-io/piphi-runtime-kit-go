# piphi-runtime-kit-go

Small Go helpers for building PiPhi runtime integrations.

This package is intentionally thin. It removes repetitive PiPhi runtime
plumbing without hiding the HTTP contract behind a large framework.

## Initial scope

Version `0.1.x` focuses on:

- runtime auth context
- request/header auth helpers
- process state
- background goroutine tracking
- telemetry delivery to PiPhi Core
- event delivery to PiPhi Core
- config lifecycle helpers
- config sync helpers
- discovery normalization and response helpers
- runtime health and diagnostics helpers
- in-memory runtime registry for active entries and recent events

## Design goals

- keep the API explicit and easy to read
- avoid hidden global state
- stay close to the PiPhi HTTP contract
- be usable from any Go HTTP framework
- rely on the standard library first

## Package shape

```text
.
  adapters/
    nethttp.go
  auth.go
  configuration.go
  context.go
  config_sync.go
  discovery.go
  dispatch.go
  events.go
  health.go
  registry.go
  state.go
  tasks.go
  telemetry.go
  types.go
  examples/
    minimal_nethttp_runtime/
      main.go
      README.md
```

## Example apps

- [`examples/minimal_nethttp_runtime/main.go`](./examples/minimal_nethttp_runtime/main.go)
  shows a small runtime built with the standard library `net/http`.
- [`examples/minimal_nethttp_runtime/README.md`](./examples/minimal_nethttp_runtime/README.md)
  explains the intended route shape and usage patterns.

## net/http adapter

The Go kit includes a tiny `net/http` adapter layer for the most repetitive
request-auth flow.

```go
parsed := adapters.SyncRuntimeAuthFromRequest(runtime, r, "")
log.Println(adapters.FormatRuntimeAuthSyncLogFromRequest(r, ""))
_ = parsed
```

## Gin adapter

The Go kit also includes a thin Gin adapter for integrations that already use
Gin for route wiring.

```go
parsed := adapters.SyncRuntimeAuthFromGinContext(runtime, c, payloadContainerID)
log.Println(adapters.FormatRuntimeAuthSyncLogFromGinContext(c, payloadContainerID))
_ = parsed
```

## What stays outside the SDK

The SDK should cover PiPhi runtime plumbing, not vendor logic.

The following should stay inside each integration:

- vendor discovery logic
- polling behavior
- command execution
- vendor entity transformation
- integration-specific event semantics

## Next steps

- add a small chi example or adapter if real integrations need it
- add a conformance CLI that validates runtime endpoints against the PiPhi contract
