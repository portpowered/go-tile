package tile

import (
	"net/http"
	"testing"

	"github.com/portpowered/go-tile/pkg/tilemodels"
	"github.com/portpowered/go-tile/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_InvalidOptions_ReturnsTypedBadRequestError(t *testing.T) {
	tests := []struct {
		name string
		opt  Option
	}{
		{name: "nil option", opt: nil},
		{name: "nil HTTP client", opt: WithHTTPClient(nil)},
		{name: "empty base URL", opt: WithBaseURL("")},
		{name: "base URL without host", opt: WithBaseURL("/api/v1")},
		{name: "empty user agent", opt: WithUserAgent(" ")},
		{name: "empty locale", opt: WithLocale("")},
		{name: "nil UUID generator", opt: WithUUIDGenerator(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opt)

			assert.Nil(t, client)
			require.Error(t, err)
			assert.True(t, tilemodels.IsBadRequestError(err), "expected BadRequestError, got %T", err)
		})
	}
}

func TestNewClient_CustomOptions_AreAppliedToRequests(t *testing.T) {
	fixtureDir := getFixtureDir()
	mockTransport := mocks.NewMockTransport(fixtureDir)
	httpClient := &http.Client{Transport: mockTransport}
	client, err := NewClient(
		WithHTTPClient(httpClient),
		WithBaseURL("https://tile.example.test/api/v1/"),
		WithUserAgent("PortOS Tile Test/1.0"),
		WithLocale("fr-FR"),
		WithUUIDGenerator(func() string { return "generated-client-uuid" }),
	)
	require.NoError(t, err)
	defer client.Close()

	_, err = client.Login(newTestContext(), tilemodels.LoginRequest{
		Email:    "user@example.com",
		Password: "secret123",
	})
	require.NoError(t, err)

	req := mockTransport.FindRequest("PUT", "/clients/generated-client-uuid")
	require.NotNil(t, req)
	assert.Contains(t, req.URL, "https://tile.example.test/api/v1/clients/generated-client-uuid")
	assert.Equal(t, "PortOS Tile Test/1.0", req.Headers.Get("User-Agent"))
	assert.Equal(t, "fr-FR", req.Headers.Get("Accept-Language"))
}

func TestClient_AfterClose_ReturnsTypedClosedError(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)
	require.NoError(t, client.Close())

	resp, err := client.GetTiles(newTestContext(), tilemodels.GetTilesRequest{
		SessionToken: "token",
		UserUUID:     "user-uuid",
	})

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.True(t, tilemodels.IsClosedError(err), "expected ClosedError, got %T", err)
}
