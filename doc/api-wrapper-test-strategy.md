# Tile API Wrapper Test Strategy

---
author: Port OS Team
last modified: 2026, april, 12
doc-id: ARCH-066
---

## Overview

Tile wrapper tests must preserve observed Tile cloud behavior without requiring live credentials, real trackers, or the external `pytile` fork during normal development. The primary strategy is to derive redacted HTTP fixtures or mimic envelopes from `pytile` instrumentation and run them through the Go wrapper transport seam.

This document defines the capture path, redaction rules, required cases, and fixture provenance expectations for `integrations/go-tile`.

For reviewer checks, [Tile pytile capture provenance](./pytile-capture-provenance.json) is the canonical machine-checkable source for the `pytile` fork repository, instrumentation branch, instrumentation commit, capture command, committed fixture paths, fixture provenance type, and redaction checklist status. This strategy document explains how those fixtures are used; the manifest owns the structured provenance data.

## System Context

Consumers:

- `integrations/go-tile/pkg/tile` tests use fixture-backed HTTP transports to exercise the real client request and decode path.
- Backend Tile plugin tests use service-boundary mocks and must not replace wrapper transport tests.
- Maintainers with Tile test accounts may run opt-in live probes or `pytile` mimic captures outside normal CI.

Dependencies and boundaries:

- Capture source: upstream [`bachya/pytile`](https://github.com/bachya/pytile) plus the PortPowered fork [`portpowered/pytile`](https://github.com/portpowered/pytile).
- Instrumentation branch: [`portos/tile-mimic-capture`](https://github.com/portpowered/pytile/tree/portos/tile-mimic-capture).
- Instrumentation commit: [`d789ae7faa8cfd3ca0e2164ee9757c6c9cd0f53c`](https://github.com/portpowered/pytile/commit/d789ae7faa8cfd3ca0e2164ee9757c6c9cd0f53c).
- Normal test boundary: `go test ./...` inside `integrations/go-tile` must not require network access, Tile credentials, real devices, Python, or the external fork.
- Fixture boundary: committed fixtures must be redacted, deterministic, reviewable JSON.

## Capture And Mimic Approach

1. Instrument `pytile` transport calls in the PortPowered fork so each Tile HTTP exchange can be written as a replayable envelope.
2. Capture login, tile listing, and representative failure responses with a dedicated Tile test account and test trackers.
3. Normalize captured payloads into fixture envelopes or plain fixture bodies used by the Go wrapper tests.
4. Replace credentials, session tokens, client UUIDs, user UUIDs, device UUIDs, names, timestamps, precise coordinates, and account-specific metadata with synthetic values.
5. Commit only redacted fixtures under `integrations/go-tile/test/fixtures/` or a future replay fixture directory.
6. Record fixture provenance in this document or fixture-local metadata before promoting a capture into normal tests.

The committed Go fixtures are redacted synthetic replay envelopes derived from upstream `pytile` endpoint/model evidence and existing Go fixture shapes. They satisfy normal wrapper tests because they preserve observable Tile HTTP behavior at the Go transport seam without live credentials. Future live captures from the PortPowered fork should replace or annotate these fixtures when maintainers need current production evidence.

## Redaction Rules

Always redact or synthesize:

- Email addresses and passwords.
- Session tokens, refresh tokens, client UUIDs, and user UUIDs.
- Real Tile UUIDs and device names.
- Precise home, work, or travel coordinates.
- Raw timestamps that identify a real user routine.
- Headers that contain secrets or stable account identifiers.
- Error payloads that echo account data.

Use deterministic replacements:

- `test-session-token-*` for session values.
- `test-client-uuid-*` and `test-user-uuid-*` for account identifiers.
- `tile-uuid-*` for Tile devices.
- Coarse city-level coordinates only when location behavior needs latitude and longitude.
- Fixed timestamps that cannot be tied to a real account.

## Required Success Cases

| Case | Evidence Source | Fixture Expectation | Assertions |
|---|---|---|---|
| Successful login/session creation | `pytile` capture and existing `tile_clients.json` | HTTP 2xx response with client UUID, session token, user UUID, and session start timestamp | Method, path, safe headers, request body shape, decoded session fields, no credential logging. |
| Zero tiles | `pytile` capture or synthetic fixture based on observed envelope | `num_tiles: 0` and empty `result_list` | Empty decoded slice and stable response metadata. |
| One tile | `pytile` capture or redacted fixture | One visible Tile with full state | Decoded identity, type, visibility, location, battery, and lost state. |
| Many tiles | Existing `tile_user_tiles.json` plus future capture | Multiple devices with different `tile_type` values | Stable ordering expectations only when vendor order is observed; otherwise assert set membership. |
| Hidden tiles | Captured or mimic fixture with `visible: false` | Hidden device included in wrapper response | Wrapper decodes visibility; backend plugin filters later. |
| Optional or missing location fields | Captured or synthetic malformed/partial fixture based on observed optionality | Missing latitude, longitude, altitude, timestamp, or state object where Tile allows it | Wrapper behavior is explicit: zero value, pointer omission, or typed malformed response according to model contract. |
| Representative device variants | `pytile` Tile object fields and captures from different trackers | Tile Mate, Tile Slim, phone-like or other observed variants | Decoded `tile_type`, firmware/hardware-style metadata, and visibility. |

## Required Failure Cases

| Case | Evidence Source | Fixture Expectation | Assertions |
|---|---|---|---|
| Invalid credentials | `pytile` capture or live probe | Login failure response | Typed authentication error. |
| Expired or missing session | `pytile` capture or live probe | Tile listing with HTTP 401 or equivalent Tile failure body | Typed authentication or token/session error. |
| Unauthorized or forbidden response | Captured HTTP 401/403 | Error envelope with no secrets | Typed auth or HTTP error according to status. |
| Throttling | Captured HTTP 429 or synthetic envelope from observed status | Retry-after header when present | Typed rate-limit error with retry context if available. |
| Vendor server failure | Captured or synthetic HTTP 5xx | Server failure body | Typed HTTP or retryable vendor error. |
| Malformed JSON | Synthetic fixture | Invalid JSON body with HTTP 2xx | Typed malformed response error. |
| Network failure | Mock transport error | No HTTP response | Typed network error wrapping the transport failure. |

## Request Assertions

Wrapper transport tests should assert safe, stable request details:

- HTTP method for login and tile listing.
- Path shape such as `/clients/{client_uuid}` and `/users/{user_uuid}/user_tiles`.
- Required Tile app headers that do not contain secrets.
- Session-token header presence without snapshotting a real token.
- JSON request body shape for login without committing real credentials.
- Query parameters when Tile behavior uses them.

Tests should avoid asserting unstable values such as generated UUIDs unless injected through explicit options.

## Fixture Provenance

Every promoted fixture should record:

- Capture source: upstream `pytile`, PortPowered fork branch or commit, live probe, existing Go fixture, or synthetic fixture derived from documented behavior.
- Capture date.
- Tile account type and device type in broad terms only.
- Redaction checklist result.
- The wrapper behavior the fixture is meant to preserve.

Record those values in [Tile pytile capture provenance](./pytile-capture-provenance.json) when a fixture depends on `pytile` evidence or another capture tool whose fork, branch, commit, command, fixture path, provenance type, or redaction status must be validated. Keep this section as a readable summary; do not make it a parallel source of truth for machine-checkable provenance.

Current committed fixtures:

| Fixture | Provenance | Preserved Behavior |
|---|---|---|
| `integrations/go-tile/test/fixtures/tile_clients.json` | Existing Go fixture with synthetic identifiers | Login/session response shape. |
| `integrations/go-tile/test/fixtures/tile_user_tiles.json` | Existing Go fixture with synthetic identifiers and coarse coordinates | Multi-tile listing, location fields, visibility, lost state, and battery fields. |
| `integrations/go-tile/test/fixtures/login_success_session.json` | Synthetic replay envelope derived from upstream `pytile` login/session behavior and existing Go login fixture shape | Successful session creation, returned session token, client UUID, user UUID, and session timestamp. |
| `integrations/go-tile/test/fixtures/login_invalid_credentials.json` | Synthetic replay envelope derived from upstream `pytile` invalid-auth error mapping | Invalid credential response mapped to typed authentication behavior. |
| `integrations/go-tile/test/fixtures/tiles_zero.json` | Synthetic replay envelope derived from upstream `pytile` tile state list behavior | Empty Tile listing with `num_tiles: 0` and empty `result_list`. |
| `integrations/go-tile/test/fixtures/tiles_one.json` | Synthetic replay envelope derived from upstream `pytile` Tile object fields and existing Go tile fixture shape | Single visible Tile with identity, type, location, lost state, and battery fields. |
| `integrations/go-tile/test/fixtures/tiles_many_variants.json` | Synthetic replay envelope derived from upstream `pytile` Tile object fields and existing Go multi-tile fixture shape | Multiple Tile variants, hidden Tile visibility, coarse coordinates, and battery/lost fields. |
| `integrations/go-tile/test/fixtures/tiles_missing_location_fields.json` | Synthetic replay envelope derived from upstream `pytile` missing `last_tile_state` fixture behavior | Optional or absent location fields remain deterministic in decoding tests. |
| `integrations/go-tile/test/fixtures/tiles_expired_session.json` | Synthetic replay envelope derived from upstream `pytile` invalid-auth error mapping | Expired or missing session maps to typed auth/session behavior. |
| `integrations/go-tile/test/fixtures/tiles_forbidden.json` | Synthetic replay envelope derived from HTTP 403 vendor failure semantics | Forbidden Tile listing maps to actionable wrapper error behavior. |
| `integrations/go-tile/test/fixtures/tiles_throttled.json` | Synthetic replay envelope derived from HTTP 429 vendor throttling semantics | Throttling maps to typed rate-limit behavior and preserves safe retry headers. |
| `integrations/go-tile/test/fixtures/tiles_vendor_server_failure.json` | Synthetic replay envelope derived from HTTP 5xx vendor failure semantics | Server failures map to typed HTTP/vendor failure behavior. |
| `integrations/go-tile/test/fixtures/tiles_malformed_json.json` | Synthetic malformed fixture | Malformed successful response maps to typed malformed response behavior. |

## Design Decisions

### Test The Wrapper At The Transport Seam

Tile wrapper tests should inject an HTTP client or round tripper and exercise the real request construction, response decoding, and typed error classification. This gives reviewers evidence that the Go wrapper still speaks the Tile-shaped HTTP contract.

Rejected alternatives:

- Mocking `TileClient` for wrapper tests: rejected because it bypasses the request and decode behavior the wrapper owns.
- Requiring live Tile credentials for normal tests: rejected because normal CI must be deterministic and secret-free.

### Keep Backend Tests At The Service Boundary

Backend Tile plugin tests should mock the Tile service or client interface. Backend tests prove Port OS route discovery, auth metadata, and locator message mapping. They should not duplicate every Tile HTTP fixture.

Rejected alternatives:

- Backend tests using live Tile API: rejected because backend plugin behavior must remain deterministic.
- Backend tests importing `pytile`: rejected because Python reference tooling stays outside the production runtime and backend test contract.

## Verification Commands

Use these focused commands when changing Tile wrapper or backend plugin behavior:

```bash
cd integrations/go-tile && go test ./...
cd backend && go test ./internal/integrations/tile
cd backend && go test ./internal/integrations/tile -run 'TestTilePluginSmoke_'
```

The smoke selector runs the backend Tile auth, discovery, locator, and expired-session checks through the real `go-tile` client with redacted replay fixtures. It complements the wrapper transport tests and backend service-boundary mock tests without requiring live Tile credentials or network access.

For capture provenance changes, run this reviewer smoke before accepting new Tile fixture provenance:

```bash
make validate-capture-provenance
```

A passing run must report that `docs/architecture/integrations/tile/pytile-capture-provenance.json` passed validation. That proves the committed manifest, fixture paths, completed redaction checklist, PortPowered fork repository, instrumentation branch, and instrumentation commit are machine-checkable from repository state plus public git metadata. The command does not execute `pytile`, require Python, require Tile credentials, require real Tile devices, or call the Tile API.

The negative-case coverage for the validator lives in `scripts/validateCaptureProvenance.node.mjs` and runs through `make root-workflow-tests`. It includes failures for missing required fields, missing fixture paths, invalid provenance types, incomplete redaction status or checklist items, unreachable branch metadata, and instrumentation commit drift.

## Constraints And Tradeoffs

- Captures from the external fork are available through branch `portos/tile-mimic-capture`, but no live-account capture is committed unless a fixture's provenance row says so.
- Current fixtures are synthetic replay envelopes. They are acceptable for normal tests because they are derived from documented `pytile` evidence and existing Go fixture shapes, but production drift should be checked with fresh fork captures when Tile behavior changes.
- Coordinates in fixtures can prove location mapping but must not identify real user locations.
- Tile throttling behavior may depend on account state and rate limits; capture metadata should record only enough context to make retry behavior reviewable.

## Related

- [Tile API wrapper research](./api-wrapper-research.md) records the evidence sources, fork status, vendor model, and compatibility matrix.
- [Tile pytile capture provenance](./pytile-capture-provenance.json) is the canonical structured manifest for Tile `pytile` fork and fixture provenance.
- [Third-Party API Wrapper Development Guide](../../../processes/integrations/third-party-api-wrapper-development-guide.md) defines the wrapper capture and fixture strategy.
- [Third-Party Integration Standards](../../../standards/systems/third-party-integration-standards.md) requires fixture-backed behavior and typed external failure mapping.
- [Integrations Expectations](../../../processes/AGENTS.md) defines local expectations for tests, docs, examples, and command surfaces under `integrations/`.
