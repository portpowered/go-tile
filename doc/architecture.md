# go-tile Architecture

This document explains the local wrapper design for `go-tile`: the supported Tile API surface, transport shape, explicit session model, dependency injection seams, typed errors, and known Tile API sharp edges. Broader evidence provenance and the compatibility matrix live in the repository architecture docs.

## Purpose

`go-tile` wraps the unpublished Tile cloud API behavior that Port OS needs for authentication, discovery, and location reporting. The wrapper is intentionally vendor-shaped. It exposes Tile login and Tile device listing without importing backend, route, plugin, auth-link, or generated API packages.

## Public Contract

The public client surface is `tile.TileClient`:

- `Login(ctx, tilemodels.LoginRequest)` creates Tile session material from explicit email/password credentials.
- `GetTiles(ctx, tilemodels.GetTilesRequest)` lists Tile devices for an explicit `SessionToken` and `UserUUID`.
- `Close()` marks the client closed and makes later requests fail with `tilemodels.ClosedError`.

`tile.Client` has a compile-time conformance check against `tile.TileClient`. Backend code should define a narrower service boundary when it needs only a subset of these methods.

## Construction and Dependency Injection

`tile.NewClient(opts ...tile.Option)` validates all configuration before returning a client. Supported options are:

- `tile.WithHTTPClient(*http.Client)` for fixture, replay, proxy, or custom transport injection
- `tile.WithBaseURL(string)` for replay servers, capture harnesses, or future regional endpoints
- `tile.WithUserAgent(string)` for observed Tile client compatibility
- `tile.WithLocale(string)` for the `Accept-Language` request header
- `tile.WithUUIDGenerator(func() string)` for deterministic login request paths in tests

The client owns no long-lived credential state. Optional configuration is process-safe after construction, and `Close()` is idempotent.

## Auth and Session Model

Tile login uses:

- method: `PUT`
- path: `/clients/{client_uuid}`
- body: email and password
- generated request metadata: Tile app ID, app version, user agent, and locale headers

The login response includes the session token and user UUID needed for later data calls. The wrapper returns that material to the caller and does not persist it on the client.

Tile listing uses:

- method: `GET`
- path: `/users/{user_uuid}/user_tiles`
- auth header: `Tile_session_token: <session token>`

Missing session token returns `tilemodels.TokenError`. Missing user UUID returns `tilemodels.BadRequestError`.

## Transport Design

The wrapper uses HTTP/JSON through the configured `*http.Client`. Request construction is centralized in `Client.doRequest`, which attaches these observed Tile headers:

- `Content-Type: application/json`
- `Tile_app_id: ios-tile-production`
- `Tile_app_version: 2.88.1.4450`
- `User-Agent`
- `Accept-Language`
- `Tile_session_token` when the operation requires a session

Response decoding is centralized in `decodeResponse`. Non-success responses are translated before JSON decoding so callers can branch on typed errors.

## Endpoint Shape

The currently supported endpoints are:

| Operation | Method | Path | Session Required |
|---|---|---|---|
| Login | `PUT` | `/clients/{client_uuid}` | No |
| List tiles | `GET` | `/users/{user_uuid}/user_tiles` | Yes |

The default base URL is `https://production.tile-api.com/api/v1`. Tests should override it through `WithBaseURL` or through an injected HTTP transport.

## Typed Error Strategy

Public methods return errors from `tilemodels`:

- `AuthenticationError` for `401` and `403` vendor responses
- `BadRequestError` for invalid caller input or `400` vendor responses
- `NetworkError` for transport and response-read failures
- `HTTPError` for non-success statuses without a narrower mapping
- `RateLimitError` for `429`, preserving safe `Retry-After` metadata
- `TokenError` for missing request-time session material
- `NotFoundError` for `404`
- `MalformedResponseError` for invalid JSON or shape mismatches
- `ClosedError` for use after `Close()`

Use the `tilemodels.Is*Error` helpers with `errors.As` semantics at package boundaries.

## Fixture and Replay Testing

Normal tests run without live Tile credentials. `test/mocks.MockTransport` records outbound HTTP requests and serves redacted fixtures from `test/fixtures/`.

Fixture envelopes may carry `_meta.statusCode`, `_meta.headers`, and `_body`. Tests assert:

- outbound methods and paths
- safe headers
- login body shape
- explicit session header behavior
- decoded Tile models
- typed error mapping for auth, token/session, throttling, server, malformed JSON, and network failures

Live probes or reference-library captures must be promoted into redacted fixtures before becoming normal test coverage.

## Known Tile API Sharp Edges

- Tile's cloud API is unpublished, so `pytile`, captured traffic, existing fixtures, and live probes are evidence sources rather than complete truth.
- `ClientUUID` is part of the login endpoint path and must be deterministic in tests.
- Listing tiles requires both the session token and user UUID returned by login.
- Hidden devices can be present in vendor responses; backend plugin discovery is responsible for filtering them from Port OS routes.
- Location fields may be absent or zero-valued in observed responses. Consumers must tolerate incomplete `last_tile_state` data.
- Premium history, ringing, mutation, sharing, account settings, and refresh behavior are not currently implemented.

## Verification Surface

The wrapper-local `Makefile` exposes the review commands:

```bash
make build
make test
make fmt
make vet
make lint
make test-live
```

`make lint` runs the pinned Staticcheck analyzer and fails on findings. `make test-live` records that no live Tile test is retained. The default test and quality surface must remain fixture-backed.

## References

- [README](../README.md)
- [Tile API wrapper research](./api-wrapper-research.md)
- [Tile API wrapper test strategy](./api-wrapper-test-strategy.md)
- [Third-party API wrapper development guide](https://github.com/portpowered/portos-backend/blob/main/docs/processes/integrations/third-party-api-wrapper-development-guide.md)
- [Third-party integration standards](https://github.com/portpowered/portos-backend/blob/main/docs/standards/systems/third-party-integration-standards.md)
