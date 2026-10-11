# go-tile

`go-tile` is the Port OS-owned Go wrapper for the Tile cloud API. It keeps Tile auth, session, and device-listing behavior behind a reusable vendor-shaped client so backend plugin code can depend on explicit request structs, typed errors, and fixture-backed tests instead of live Tile credentials.

## Import Path

```go
import (
    "github.com/portpowered/go-tile/pkg/tile"
    "github.com/portpowered/go-tile/pkg/tilemodels"
)
```

## Supported Surface

The wrapper currently supports:

- email and password login through `Client.Login`
- explicit session-token and user-UUID inputs through `tilemodels.GetTilesRequest`
- Tile device listing through `Client.GetTiles`
- decoding Tile identity, visibility, type, firmware, last location, connection, and battery fields present in the observed response shape
- typed errors for auth, token/session, bad request, network, HTTP, rate limit, not found, malformed response, and closed-client cases

Unsupported behavior is documented in the Tile research matrix:

- premium location history
- ringing or mutating Tile devices
- refresh flows beyond the observed session material
- pagination, regional endpoint selection, and event streams
- marketplace, sharing, or account-management operations

## Auth and Session Model

`Login` is the only auth-producing operation. It accepts `tilemodels.LoginRequest` and returns `tilemodels.LoginResponse` with the `SessionToken`, `UserUUID`, and `ClientUUID` returned by Tile.

Normal data operations do not store credentials on the client. Callers pass session material per request:

```go
tiles, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
    SessionToken: login.SessionToken,
    UserUUID:     login.UserUUID,
})
```

This shape matches the backend plugin requirement that Port OS reconstructs session material from stored secrets at dispatch time.

## Basic Usage

```go
ctx := context.Background()

client, err := tile.NewClient(
    tile.WithUserAgent("PortOS go-tile"),
    tile.WithLocale("en-US"),
)
if err != nil {
    return err
}
defer client.Close()

login, err := client.Login(ctx, tilemodels.LoginRequest{
    Email:    email,
    Password: password,
})
if err != nil {
    return err
}

tiles, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
    SessionToken: login.SessionToken,
    UserUUID:     login.UserUUID,
})
if err != nil {
    return err
}

_ = tiles
```

See the runnable example in `pkg/tile/example_test.go` for a credential-free flow that uses an in-process HTTP server.

## Fixture Strategy

Normal tests use replay envelopes and a mock `http.RoundTripper` at the HTTP seam. Fixtures live under `test/fixtures/` and may include:

- `_meta.statusCode` for the vendor HTTP status
- `_meta.headers` for safe response headers such as `Retry-After`
- `_body` for the redacted Tile response payload

Fixtures must not contain real emails, passwords, session tokens, user UUIDs, client UUIDs, precise home locations, device serials, or other user-identifying data. The test strategy and provenance expectations live in:

- `./doc/api-wrapper-test-strategy.md`
- `./doc/api-wrapper-research.md`

## Local Verification

Run commands from the repository root:

```bash
make build
make test
make fmt
make vet
make lint
```

`make lint` runs the pinned Staticcheck analyzer and fails on findings. `make test-live` records that no live Tile test is retained; normal verification never requires Tile credentials or network access.

## Package Layout

```text
go-tile/
├── doc/
│   └── architecture.md
├── pkg/
│   ├── tile/          # public client, options, auth, and Tile operations
│   └── tilemodels/    # request/response models and typed errors
├── test/
│   ├── fixtures/      # redacted replay envelopes and response payloads
│   └── mocks/         # HTTP transport test seam
├── Makefile
├── go.mod
└── README.md
```

## References

- [Tile wrapper architecture](./doc/architecture.md)
- [Tile API wrapper research](./doc/api-wrapper-research.md)
- [Tile API wrapper test strategy](./doc/api-wrapper-test-strategy.md)
- [Third-party API wrapper development guide](https://github.com/portpowered/portos-backend/blob/main/docs/processes/integrations/third-party-api-wrapper-development-guide.md)
- [Third-party integration standards](https://github.com/portpowered/portos-backend/blob/main/docs/standards/systems/third-party-integration-standards.md)
