# piphi-runtime-kit-go

Small Go helpers for building PiPhi runtime integrations.

This package is intentionally thin. It removes repetitive PiPhi runtime
plumbing without hiding the HTTP contract behind a large framework. The goal
is to give Go developers a clear baseline for building runtimes while keeping
vendor logic in the integration.

## Who this is for

This SDK is for developers building PiPhi integrations in Go.

It is a good fit if you are using:

- `net/http`
- Gin
- another Go HTTP framework where you still want shared PiPhi helpers

## What the SDK handles

The SDK is meant to own the common runtime plumbing:

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
- in-memory runtime registry for active entries, state, and recent events
- PiPhi-specific delivery errors

## What stays in your integration

Your integration still owns:

- vendor discovery logic
- polling behavior
- command execution
- vendor protocol handling
- entity transformation
- integration-specific event semantics

The SDK should make runtime plumbing easier, not take over the device logic.

## Install

The module path is:

```text
github.com/piphi-network/piphi-runtime-kit-go
```

Add it to your integration:

```bash
go get github.com/piphi-network/piphi-runtime-kit-go
```

If you are working locally before publishing tags, you can use a `replace`
directive in your integration's `go.mod`.

## The Golden Path

If you are new to PiPhi, start here.

### 1. Create a starter

Use `NewRuntimeStarter(...)` first. It gives you one obvious object with the
most common pieces already wired together:

- shared runtime auth and process state
- an in-memory registry
- a telemetry client
- an event client

```go
starter := runtimekit.NewRuntimeStarter[map[string]any, map[string]any, map[string]any](
	"demo-runtime",
	"Demo Runtime",
	"0.1.0",
	"",
	100,
)

runtime := starter.Runtime
registry := starter.Registry
telemetry := starter.Telemetry
events := starter.Events
```

### 2. Define your runtime entry shape

In Go, the cleanest pattern is to define a small struct for the active runtime
entry you keep in the registry.

```go
type DemoEntry struct {
	DeviceID      string
	ConfigID      string
	IntegrationID string
	Host          string
}
```

### 3. Sync auth from every request

PiPhi runtimes receive auth and scope through headers. Sync that into the
runtime before sending telemetry or events.

With `net/http`:

```go
parsed := adapters.SyncRuntimeAuthFromRequest(runtime, r, payloadContainerID)
log.Println(adapters.FormatRuntimeAuthSyncLogFromRequest(r, payloadContainerID))
_ = parsed
```

With Gin:

```go
parsed := adapters.SyncRuntimeAuthFromGinContext(runtime, c, payloadContainerID)
log.Println(adapters.FormatRuntimeAuthSyncLogFromGinContext(c, payloadContainerID))
_ = parsed
```

### 4. Store active runtime entries in the registry

Use the registry for the runtime's active in-memory working set:

```go
starter.Registry.Set(payload.ID, DemoEntry{
	DeviceID:      payload.ID,
	ConfigID:      payload.ConfigID,
	IntegrationID: payload.IntegrationID,
	Host:          payload.Host,
})
```

PiPhi Core is still the source of truth for configs. The registry is just the
working state inside the runtime process.

### 5. Send telemetry and events

You can call the clients directly:

```go
err := starter.Telemetry.SendMetrics(context.Background(), runtimekit.SendMetricsInput{
	AuthContext: starter.Runtime.Auth,
	DeviceID:    "sensor-1",
	Metrics: map[string]any{
		"temperature_c": 22.4,
		"humidity":      47.0,
	},
})
if err != nil {
	log.Println(err)
}
```

### 6. Expose the common runtime routes

Most runtimes should provide at least:

- `/health`
- `/diagnostics`
- `/discover`
- `/config`
- `/configs/sync` or `/config/sync`
- `/deconfigure`
- `/events`
- `/state`
- `/entities`

Some integrations also expose `/ui` or `/ui-config`.

### 7. Compare against the example app

The reference example is:

- [`examples/minimal_nethttp_runtime/main.go`](./examples/minimal_nethttp_runtime/main.go)
- [`examples/minimal_nethttp_runtime/README.md`](./examples/minimal_nethttp_runtime/README.md)

## The IDs You Need To Understand

These ids show up in most integrations:

- `ID`
  The runtime's local config id.
- `ConfigID`
  The real PiPhi Core config UUID.
- `DeviceID`
  The physical or logical device identifier.
- `ContainerID`
  The runtime/container scope used for Core auth.
- `IntegrationID`
  The installed integration id in Core.

The most common mistake is confusing `ID` with `ConfigID`.

If you are sending events back to Core, `ConfigID`, `ContainerID`, and
`IntegrationID` need to be correct.

## Typical Runtime Flow

Most integrations follow this sequence:

1. PiPhi calls your runtime.
2. Your route syncs auth from request headers.
3. You validate or normalize the config payload.
4. You connect to the vendor API or local device.
5. You store the active entry in the registry.
6. You poll, listen, or subscribe for changes.
7. You send telemetry to Core.
8. You emit meaningful events to Core.
9. You expose health and diagnostics for supportability.

The SDK is designed to make steps `2`, `5`, `7`, `8`, and `9` easier.

## Adapters

### net/http adapter

The Go kit includes a small `net/http` adapter for the repetitive request-auth flow.

```go
parsed := adapters.SyncRuntimeAuthFromRequest(runtime, r, "")
log.Println(adapters.FormatRuntimeAuthSyncLogFromRequest(r, ""))
_ = parsed
```

### Gin adapter

The Go kit also includes a thin Gin adapter for integrations that already use Gin.

```go
parsed := adapters.SyncRuntimeAuthFromGinContext(runtime, c, payloadContainerID)
log.Println(adapters.FormatRuntimeAuthSyncLogFromGinContext(c, payloadContainerID))
_ = parsed
```

Both adapters are intentionally narrow. They help with auth extraction and log
formatting, but they do not try to hide the framework itself.

## Clear Error Handling

The SDK classifies common delivery failures into PiPhi-specific errors.

Important examples:

- `CoreUnavailableError`
  PiPhi Core could not be reached.
- `CoreTimeoutError`
  PiPhi Core did not respond before the timeout.
- `CoreRouteNotFoundError`
  The expected Core endpoint is missing or the base URL is wrong.
- `CoreAuthError`
  Core rejected runtime auth.
- `CoreServerError`
  Core failed while processing the request.

This is meant to produce better developer-facing logs than raw transport
failures alone.

## Common Mistakes

- Using `ID` where `ConfigID` should be used.
- Forgetting to sync auth before sending telemetry.
- Sending events without `ConfigID`, `ContainerID`, or `IntegrationID`.
- Treating the registry as the source of truth.
- Expecting the SDK to own polling cadence or vendor protocol logic.

## Troubleshooting

If telemetry or event delivery fails:

- confirm PiPhi Core is reachable
- confirm the Core base URL is correct
- confirm `ContainerID` is present
- confirm `ConfigID` is the real Core config UUID
- read the classified delivery error before chasing transport traces

If config sync behaves incorrectly:

- verify you are storing the right active ids
- verify the sync generation from Core
- verify stale entries are removed correctly

## Package shape

```text
.
  adapters/
    nethttp.go
    gin.go
  auth.go
  configuration.go
  context.go
  config_sync.go
  discovery.go
  dispatch.go
  errors.go
  events.go
  health.go
  registry.go
  starter.go
  state.go
  tasks.go
  telemetry.go
  types.go
  examples/
    minimal_nethttp_runtime/
      main.go
      README.md
```

## Summary

If you are unsure where to start:

1. create a starter
2. sync auth in every route
3. keep config identity straight
4. use the registry for active runtime state
5. send telemetry and events through the SDK
6. compare your runtime to the example app
