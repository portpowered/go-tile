package tile

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-tile/pkg/tilemodels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin_Success(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()
	require.NoError(t, mockTransport.SetResponseFromFixture("PUT", "/api/v1/clients/known-uuid", "login_success_session.json"))

	loginResp, err := client.Login(ctx, tilemodels.LoginRequest{
		Email:    "user@example.com",
		Password: "secret123",
	})

	require.NoError(t, err)
	require.NotNil(t, loginResp)
	assert.Equal(t, "redacted-session-token", loginResp.SessionToken)
	assert.Equal(t, "redacted-client-uuid", loginResp.ClientUUID)
	assert.Equal(t, "redacted-user-uuid", loginResp.UserUUID)
	assert.Equal(t, 0, loginResp.ResultCode)
}

func TestLogin_VerifiesRequest(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()

	_, err := client.Login(ctx, tilemodels.LoginRequest{
		Email:    "user@example.com",
		Password: "secret123",
	})
	require.NoError(t, err)

	req := mockTransport.FindRequest("PUT", "/clients/")
	require.NotNil(t, req, "expected PUT to /clients/{uuid}")
	assert.Equal(t, "application/json", req.Headers.Get("Content-Type"))
	assert.Equal(t, TileAppID, req.Headers.Get("Tile_app_id"))
	assert.Equal(t, TileAppVersion, req.Headers.Get("Tile_app_version"))
	assert.Equal(t, DefaultUserAgent, req.Headers.Get("User-Agent"))
	assert.Equal(t, DefaultLocale, req.Headers.Get("Accept-Language"))
	var body map[string]string
	require.NoError(t, json.Unmarshal(req.Body, &body))
	assert.Equal(t, "user@example.com", body["email"])
	assert.Equal(t, "secret123", body["password"])
	assert.NotContains(t, body, "client_uuid")
}

func TestLogin_UsesExplicitClientUUID(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()

	_, err := client.Login(ctx, tilemodels.LoginRequest{
		Email:      "user@example.com",
		Password:   "secret123",
		ClientUUID: "explicit-client-uuid",
	})
	require.NoError(t, err)

	req := mockTransport.FindRequest("PUT", "/clients/explicit-client-uuid")
	require.NotNil(t, req, "expected PUT to explicit client UUID")
}

func TestLogin_AuthenticationError(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()

	require.NoError(t, mockTransport.SetResponseFromFixture("PUT", "/api/v1/clients/known-uuid", "login_invalid_credentials.json"))

	loginResp, err := client.Login(ctx, tilemodels.LoginRequest{
		Email:    "bad@example.com",
		Password: "wrong",
	})

	assert.Nil(t, loginResp)
	assert.Error(t, err)
	assert.True(t, tilemodels.IsAuthenticationError(err), "expected AuthenticationError, got %T", err)
}

func TestLogin_ErrorStringOmitsRawVendorBody(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()
	sensitiveEmail := "leaked-user@example.com"
	sensitiveToken := "vendor-session-token-123"
	mockTransport.SetResponseWithBody("PUT", "/api/v1/clients/known-uuid", http.StatusUnauthorized, map[string]string{
		"error": sensitiveEmail,
		"token": sensitiveToken,
	})

	loginResp, err := client.Login(ctx, tilemodels.LoginRequest{
		Email:    "user@example.com",
		Password: "wrong",
	})

	assert.Nil(t, loginResp)
	require.Error(t, err)
	assert.True(t, tilemodels.IsAuthenticationError(err), "expected AuthenticationError, got %T", err)
	assert.False(t, strings.Contains(err.Error(), sensitiveEmail))
	assert.False(t, strings.Contains(err.Error(), sensitiveToken))
}

func TestLogin_ServerError(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()

	mockTransport.SetResponseWithBody("PUT", "/api/v1/clients/known-uuid", http.StatusInternalServerError, map[string]string{
		"error": "internal_error",
	})

	loginResp, err := client.Login(ctx, tilemodels.LoginRequest{
		Email:    "user@example.com",
		Password: "secret123",
	})

	assert.Nil(t, loginResp)
	assert.Error(t, err)
	assert.True(t, tilemodels.IsHTTPError(err), "expected HTTPError, got %T", err)
}

func TestLogin_HTTPErrorOmitsRawVendorBody(t *testing.T) {
	client, mockTransport := newTestClient()
	defer client.Close()
	ctx := newTestContext()
	sensitiveBodyValue := "leaked-user@example.com"
	mockTransport.SetResponseWithBody("PUT", "/api/v1/clients/known-uuid", http.StatusInternalServerError, map[string]string{
		"error": sensitiveBodyValue,
	})

	loginResp, err := client.Login(ctx, tilemodels.LoginRequest{
		Email:    "user@example.com",
		Password: "secret123",
	})

	assert.Nil(t, loginResp)
	require.Error(t, err)
	assert.True(t, tilemodels.IsHTTPError(err), "expected HTTPError, got %T", err)
	assert.False(t, strings.Contains(err.Error(), sensitiveBodyValue))
	var httpErr *tilemodels.HTTPError
	require.True(t, errors.As(err, &httpErr))
	assert.Empty(t, httpErr.Body)
}
