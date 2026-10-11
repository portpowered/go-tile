# Tile API Wrapper Research

---
author: Port OS Team
last modified: 2026, april, 12
doc-id: ARCH-065
---

## Overview

The Tile integration wraps the Tile cloud API used by Tile Bluetooth trackers. Tile does not publish a supported public API for this surface, so Port OS treats observed behavior as evidence that must be documented, fixture-backed, and isolated behind a reusable Go wrapper.

This document records the evidence sources for `integrations/go-tile`, the `pytile` fork path, and the current mapping between Tile vendor behavior and Port OS plugin behavior. The reusable Go wrapper must remain the runtime dependency; `pytile` is reference evidence and capture tooling only.

The canonical machine-checkable source for the `pytile` fork, instrumentation branch, instrumentation commit, capture command, fixture paths, provenance type, and redaction status is [Tile pytile capture provenance](./pytile-capture-provenance.json). Keep prose evidence notes in this document for research context, but update the manifest when reviewers need deterministic provenance validation.

## System Context

Consumers:

- `integrations/go-tile` consumes Tile cloud HTTP behavior and exposes a Go client contract.
- `backend/internal/integrations/tile` consumes the Go wrapper through Tile client interfaces and maps Tile devices into Port OS routes and locator reports.
- Future Tile wrapper tests consume redacted fixtures or mimic captures derived from `pytile` or observed Tile traffic.

Dependencies and boundaries:

- Upstream reference: [`bachya/pytile`](https://github.com/bachya/pytile), default branch `dev`, MIT licensed.
- Fork: [`portpowered/pytile`](https://github.com/portpowered/pytile), created from upstream `bachya/pytile` on 2026-04-12.
- Mimic instrumentation branch: [`portos/tile-mimic-capture`](https://github.com/portpowered/pytile/tree/portos/tile-mimic-capture).
- Mimic instrumentation commit: [`d789ae7faa8cfd3ca0e2164ee9757c6c9cd0f53c`](https://github.com/portpowered/pytile/commit/d789ae7faa8cfd3ca0e2164ee9757c6c9cd0f53c).
- Runtime dependency boundary: Port OS backend and `integrations/go-tile` must not import Python code or execute `pytile` in production.
- Owned data: Port OS owns committed redacted fixtures, Go wrapper models, error mappings, and backend route/locator mappings. Tile owns the cloud API behavior and can change it without notice.

## Evidence Sources

| Evidence Source | What It Proves | Current Status | Notes |
|---|---|---|---|
| Upstream `bachya/pytile` README | Tile API is unpublished and unofficial; login starts with `async_login`; tile listing uses `async_get_tiles`; Tile object fields include location, visibility, battery, lost state, and firmware/hardware-style metadata; premium history is separate behavior. | Verified on GitHub on 2026-04-12. | Use as reference evidence, not runtime code. |
| Upstream `bachya/pytile` source and tests | Endpoint paths, headers, auth/session details, Tile state/detail object decoding, hidden or missing state variants, and exception mapping. | Verified from the forked source on 2026-04-12. | The Go fixtures use this as reference evidence for shape and field names. |
| [Tile pytile capture provenance](./pytile-capture-provenance.json) | Machine-checkable `pytile` fork, branch, commit, capture command, fixture paths, fixture provenance type, and redaction status. | Committed manifest validated by `make validate-capture-provenance`. | This manifest is the canonical source for reviewer validation; prose rows provide context only. |
| PortPowered fork `portos/tile-mimic-capture` at `d789ae7faa8cfd3ca0e2164ee9757c6c9cd0f53c` | Environment-gated HTTP exchange capture path that emits redacted replay envelopes from `pytile` transport calls. | Pushed on 2026-04-12. | Use `PYTILE_PORTOS_CAPTURE_DIR` with a Tile test account to capture future production evidence; update the provenance manifest before promoting new fixtures. |
| Existing Go fixture `integrations/go-tile/test/fixtures/tile_clients.json` | Login/session response shape with `client_uuid`, `session_token`, `user_uuid`, and `session_start_timestamp`. | Committed fixture exists. | Synthetic values derived from the existing Go wrapper contract and upstream `pytile` login/session behavior. |
| Existing Go fixture `integrations/go-tile/test/fixtures/tile_user_tiles.json` | User tile list response shape with `num_tiles`, `result_list`, visible devices, location fields, lost state, battery fields, and Tile type. | Committed fixture exists. | Synthetic and redacted values derived from the existing Go wrapper contract and upstream `pytile` Tile state/detail fixtures. |
| Go replay fixtures under `integrations/go-tile/test/fixtures/login_*.json` and `tiles_*.json` | Success, zero/one/many, hidden, optional-location, auth, forbidden, throttled, server failure, and malformed response behavior at the wrapper HTTP seam. | Committed fixture suite exists. | Synthetic redacted replay envelopes derived from upstream `pytile` endpoint/model evidence plus existing Go fixture shapes; not captured from a real account. |
| Live Tile probes | Current production behavior for auth failures, throttling, malformed responses, optional fields, hidden devices, and locator-related behavior. | Not used in this story. | Live probes must be opt-in and redacted before fixture promotion. |

## Vendor Model

Tile models a user account with authenticated sessions and a set of Tile devices. The current Go wrapper evidence covers:

- Login/session creation using email and password to create a session token.
- Client UUID and user UUID values returned with the session.
- Listing user tiles from a user-scoped endpoint.
- Tile device identity through UUID, name, `tile_type`, firmware version, and visibility.
- Last-known location through latitude, longitude, altitude, timestamp, lost state, connectivity timestamp, ring timestamp, VOIP state, battery level, and dead-battery flag.

`pytile` also documents premium Tile history. Port OS does not currently map premium history to a plugin behavior in this Tile path.

## Compatibility Matrix

| Vendor Behavior | Evidence Source | Wrapper Support | Port OS Plugin Support | Gap or Backlog |
|---|---|---|---|---|
| Login/session creation | Upstream `pytile` README and existing `tile_clients.json` fixture | Partially supported by current `Login` behavior | Used by Tile auth handler to obtain session token and user UUID | Follow-up stories must align request structs, client UUID, locale, and typed auth errors. |
| Session token on later requests | Existing Go code and fixtures | Supported through explicit `sessionToken` method argument | Stored as auth metadata for later discovery and locator calls | Follow-up story should ensure no ambient mutable credentials remain. |
| Tile listing | Upstream `pytile` `async_get_tiles` documentation and existing `tile_user_tiles.json` fixture | Supported by current `GetTiles` behavior | Used by route discovery and locator lookup | Add zero, one, many, hidden, and malformed fixture coverage. |
| Tile location fields | Upstream `pytile` Tile object documentation and existing fixture fields | Partially supported by current models | Locator path publishes last-known latitude, longitude, altitude, and timestamp-style data | Optional and missing location fields need fixture coverage and model review. |
| Visible Tile filtering | Existing backend mapping code and fixture `visible` field | Model field exists | Discovery should only publish visible devices | Needs backend tests for hidden Tile filtering. |
| Locator-related behavior | Existing Go fixture fields for location, lost, ring timestamp, and connectivity timestamp | Read-only last-known location is partially supported | Locator dispatch looks up a Tile by UUID and publishes a location-related report | Active ringing or live proximity search is unsupported until evidence and protocol mapping exist. |
| Premium/history behavior | Upstream `pytile` README marks history as Tile Premium behavior | Unsupported | Unsupported | Backlog only if Port OS adds a history-capable protocol. |
| Auth failures | Existing wrapper tests and upstream `pytile` error behavior to inspect | Partially supported as authentication errors for HTTP 401 | Should map to re-authentication/actionable auth failure | Add invalid credentials and expired-session mimic fixtures. |
| Throttling | Expected vendor HTTP behavior; no committed Tile fixture yet | Not specifically typed in current research scope | Should map to retryable vendor failure | Add 429 fixture and typed rate-limit mapping in follow-up story. |
| Malformed responses | Existing wrapper decoder behavior | Partially handled as decode failure | Should map to retryable or vendor-malformed backend error depending path | Add malformed JSON fixture and typed malformed response error in follow-up story. |
| Backend route discovery | Existing backend Tile plugin mapping | Wrapper supplies device list | Maps Tile devices to route discovery output | Service-boundary tests needed. |
| Backend locator reporting | Existing backend Tile plugin locator path | Wrapper supplies device list and last-known location | Publishes location-related report for known Tile route | Service-boundary tests must preserve correlation token when present. |

## Design Decisions

### Use `pytile` as Evidence, Not Runtime Code

`pytile` is useful because it already tracks the unpublished Tile API and documents the risk that the API can change without notice. Port OS should not add a production Python runtime dependency for Tile because the reusable wrapper must be a Go module under `integrations/go-tile`, independently testable, injectable, and free of backend host assumptions.

Rejected alternatives:

- Runtime dependency on `pytile`: rejected because it couples the Go backend to Python packaging, event-loop semantics, and a non-Port OS client contract.
- Copying `pytile` models directly into Go without captures: rejected because reference code is evidence, not proof that current Tile production behavior still matches.

### Preserve Vendor Behavior at the Network Seam

The wrapper should validate behavior through recorded or mimic HTTP traffic, not by mocking wrapper methods. This keeps request headers, endpoint paths, response envelopes, and error classification reviewable.

Rejected alternatives:

- Unit tests that only mock `TileClient`: rejected for wrapper behavior because those tests prove only the mock contract.
- Live Tile tests in normal CI: rejected because credentials, devices, rate limits, and account state are not deterministic.

### Keep Capture Tooling Separate From Fixture Promotion

The PRD asks for a forked and pushed `pytile` instrumentation path. The fork now exists at `https://github.com/portpowered/pytile`, and branch `portos/tile-mimic-capture` commit `d789ae7faa8cfd3ca0e2164ee9757c6c9cd0f53c` adds environment-gated capture instrumentation. That commit is the reproducible path for future live Tile mimic evidence, but the currently committed Go fixtures remain synthetic replay envelopes unless the fixture table states otherwise.

Rejected alternatives:

- Claiming a pushed branch without repository evidence: rejected because it would make fixture provenance unreviewable.
- Requiring live Tile captures for normal test execution: rejected because normal development must stay credential-free and deterministic.

## Constraints And Tradeoffs

- Tile API behavior is unpublished and may change without notice.
- Fixture timestamps and IDs must be synthetic or redacted before commit.
- Tile passwords, session tokens, client UUIDs tied to real users, and user UUIDs tied to real users must not be committed.
- `pytile` evidence can lag production behavior. Future fixture promotions should prefer captures from the instrumented fork when behavior matters.
- Current Go fixtures prove wrapper behavior against redacted replay envelopes derived from `pytile` evidence and existing Go fixture shapes. They are not a substitute for periodic live capture refreshes when Tile changes production behavior.

## Related

- [Tile API wrapper test strategy](./api-wrapper-test-strategy.md) explains capture, mimic, fixture, and redaction rules for Tile tests.
- [Tile pytile capture provenance](./pytile-capture-provenance.json) is the canonical structured manifest for Tile `pytile` fork and fixture provenance.
- [Third-Party API Wrapper Development Guide](../../../processes/integrations/third-party-api-wrapper-development-guide.md) governs wrapper research, public API shape, fixture strategy, and documentation.
- [Third-Party Integration Standards](../../../standards/systems/third-party-integration-standards.md) defines the required separation between external wrappers and Port OS plugin glue.
- [Integrations Expectations](../../../processes/AGENTS.md) defines local expectations for reusable clients under `integrations/`.
