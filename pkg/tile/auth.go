package tile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/portpowered/go-tile/pkg/tilemodels"
)

// Login authenticates with the Tile API using email and password.
// Returns a LoginResponse containing the session token and client UUID
// needed for subsequent API calls.
//
// The Tile API login endpoint is PUT /clients/{client_uuid} with email
// and password in the request body. A new client UUID is generated for
// each login if one is not already set on the client.
func (c *Client) Login(ctx context.Context, req tilemodels.LoginRequest) (*tilemodels.LoginResponse, error) {
	if req.Email == "" {
		return nil, tilemodels.NewBadRequestError("email is required", nil)
	}
	if req.Password == "" {
		return nil, tilemodels.NewBadRequestError("password is required", nil)
	}

	clientUUID := req.ClientUUID
	if clientUUID == "" {
		clientUUID = c.newUUID()
	}

	reqBody := tilemodels.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, tilemodels.NewBadRequestError("failed to marshal login request", err)
	}

	path := fmt.Sprintf(DefaultLoginPath, clientUUID)
	resp, err := c.doRequest(ctx, "PUT", path, bytes.NewReader(bodyBytes), "")
	if err != nil {
		return nil, err
	}

	result, err := decodeResponse[tilemodels.LoginResponse](resp)
	if err != nil {
		return nil, err
	}

	return result, nil
}
