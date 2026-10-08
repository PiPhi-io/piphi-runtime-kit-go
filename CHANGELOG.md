# Changelog

## Unreleased

- Added `RuntimeStateService` and `RuntimeStarter.State` for concise cached
  state publishing and real on-demand upstream refreshes.
- Added SDK-owned request validation, timeouts, receipts, and explicit
  unsupported responses for push-only runtimes.

- Added validated config/device identity helpers.
- Added typed, device-scoped telemetry readings and required telemetry `config_id`.
- Updated semantic events to Core's idempotent `event_id/type/ts/data` envelope.
- Stabilized telemetry timestamps and event identities before background dispatch.
