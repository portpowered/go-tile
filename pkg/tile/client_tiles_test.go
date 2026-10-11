package tile

import (
	"errors"
	"net/http"
	"testing"

	"github.com/portpowered/go-tile/pkg/tilemodels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTiles_Success(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()
	require.NoError(t, mockTransport.SetResponseFromFixture("GET", "/api/v1/users/test-user-uuid/user_tiles", "tiles_many_variants.json"))

	resp, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
		SessionToken: "test-session-token",
		UserUUID:     "test-user-uuid",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 4, resp.TileCount)
	require.Len(t, resp.Tiles, 4)

	// First tile
	assert.Equal(t, "tile-mate-uuid", resp.Tiles[0].UUID)
	assert.Equal(t, "Keys", resp.Tiles[0].Name)
	assert.Equal(t, "TILE_MATE", resp.Tiles[0].TileType)
	assert.True(t, resp.Tiles[0].Visible)
	assert.InDelta(t, 37.7749, resp.Tiles[0].LastTileState.Latitude, 0.0001)
	assert.InDelta(t, -122.4194, resp.Tiles[0].LastTileState.Longitude, 0.0001)
	assert.False(t, resp.Tiles[0].LastTileState.IsLost)

	// Second tile
	assert.Equal(t, "tile-slim-uuid", resp.Tiles[1].UUID)
	assert.Equal(t, "Wallet", resp.Tiles[1].Name)
	assert.Equal(t, "TILE_SLIM", resp.Tiles[1].TileType)
	assert.True(t, resp.Tiles[1].LastTileState.IsLost)
	assert.Equal(t, "TILE_STICKER", resp.Tiles[2].TileType)
	assert.Equal(t, "TILE_PRO", resp.Tiles[3].TileType)
	assert.False(t, resp.Tiles[3].Visible)
}

func TestGetTiles_VerifiesRequest(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()

	_, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
		SessionToken: "my-session-token",
		UserUUID:     "my-user-uuid",
	})
	require.NoError(t, err)

	req := mockTransport.FindRequest("GET", "/users/my-user-uuid/user_tiles")
	require.NotNil(t, req, "expected GET to /users/{uuid}/user_tiles")
	assert.Equal(t, "my-session-token", req.Headers.Get("Tile_session_token"))
	assert.Equal(t, TileAppID, req.Headers.Get("Tile_app_id"))
	assert.Equal(t, DefaultUserAgent, req.Headers.Get("User-Agent"))
	assert.Equal(t, DefaultLocale, req.Headers.Get("Accept-Language"))
}

func TestGetTiles_EmptyList(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()
	require.NoError(t, mockTransport.SetResponseFromFixture("GET", "/api/v1/users/user-1/user_tiles", "tiles_zero.json"))

	resp, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
		SessionToken: "token",
		UserUUID:     "user-1",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.TileCount)
	assert.Empty(t, resp.Tiles)
}

func TestGetTiles_OneTile(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()
	require.NoError(t, mockTransport.SetResponseFromFixture("GET", "/api/v1/users/user-1/user_tiles", "tiles_one.json"))

	resp, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
		SessionToken: "token",
		UserUUID:     "user-1",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 1, resp.TileCount)
	require.Len(t, resp.Tiles, 1)
	assert.Equal(t, "tile-one-uuid", resp.Tiles[0].UUID)
	assert.Equal(t, "TILE_MATE", resp.Tiles[0].TileType)
	assert.True(t, resp.Tiles[0].Visible)
}

func TestGetTiles_MissingLocationFields_DecodesZeroValues(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()
	require.NoError(t, mockTransport.SetResponseFromFixture("GET", "/api/v1/users/user-1/user_tiles", "tiles_missing_location_fields.json"))

	resp, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
		SessionToken: "token",
		UserUUID:     "user-1",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Tiles, 1)
	state := resp.Tiles[0].LastTileState
	assert.Zero(t, state.Latitude)
	assert.Zero(t, state.Longitude)
	assert.Zero(t, state.Altitude)
}

func TestGetTiles_MissingSessionToken_ReturnsTokenError(t *testing.T) {
	client, _ := newTestClient()
	defer client.Close()
	ctx := newTestContext()

	resp, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
		UserUUID: "user-1",
	})

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.True(t, tilemodels.IsTokenError(err), "expected TokenError, got %T", err)
}

func TestGetTiles_RateLimit_ReturnsTypedError(t *testing.T) {
	assertGetTilesFixtureError(t, "tiles_throttled.json", tilemodels.IsRateLimitError)
}

func TestGetTiles_MalformedJSON_ReturnsTypedError(t *testing.T) {
	assertGetTilesFixtureError(t, "tiles_malformed_json.json", tilemodels.IsMalformedResponseError)
}

func TestGetTiles_NotFound(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()

	mockTransport.SetResponseWithBody("GET", "/api/v1/users/bad-user/user_tiles", http.StatusNotFound, map[string]string{
		"error": "user not found",
	})

	resp, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
		SessionToken: "token",
		UserUUID:     "bad-user",
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.True(t, tilemodels.IsNotFoundError(err), "expected NotFoundError, got %T", err)
}

func TestGetTiles_FixtureBackedVendorFailures_ReturnTypedErrors(t *testing.T) {
	tests := []struct {
		name      string
		fixture   string
		assertErr func(error) bool
	}{
		{name: "expired session", fixture: "tiles_expired_session.json", assertErr: tilemodels.IsAuthenticationError},
		{name: "forbidden session", fixture: "tiles_forbidden.json", assertErr: tilemodels.IsAuthenticationError},
		{name: "vendor server failure", fixture: "tiles_vendor_server_failure.json", assertErr: tilemodels.IsHTTPError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGetTilesFixtureError(t, tt.fixture, tt.assertErr)
		})
	}
}

func TestGetTiles_NetworkFailure_ReturnsTypedError(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()
	mockTransport.SetError("GET", "/api/v1/users/user-1/user_tiles", errors.New("redacted dial failure"))

	tiles, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
		SessionToken: "token",
		UserUUID:     "user-1",
	})

	assert.Nil(t, tiles)
	require.Error(t, err)
	assert.True(t, tilemodels.IsNetworkError(err), "expected NetworkError, got %T", err)
}

func assertGetTilesFixtureError(t *testing.T, fixture string, assertErr func(error) bool) {
	t.Helper()
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()
	require.NoError(t, mockTransport.SetResponseFromFixture("GET", "/api/v1/users/user-1/user_tiles", fixture))

	tiles, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
		SessionToken: "token",
		UserUUID:     "user-1",
	})

	assert.Nil(t, tiles)
	require.Error(t, err)
	assert.True(t, assertErr(err), "expected typed error for fixture %s, got %T", fixture, err)
}
