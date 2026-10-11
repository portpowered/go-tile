package tile

import (
	"context"
	"fmt"

	"github.com/portpowered/go-tile/pkg/tilemodels"
)

// GetTiles retrieves all Tile devices for the authenticated user.
// Requires a valid session token (from Login) and the user UUID.
func (c *Client) GetTiles(ctx context.Context, req tilemodels.GetTilesRequest) (*tilemodels.GetTilesResponse, error) {
	if req.SessionToken == "" {
		return nil, tilemodels.NewTokenError("session token is required")
	}
	if req.UserUUID == "" {
		return nil, tilemodels.NewBadRequestError("user UUID is required", nil)
	}

	path := fmt.Sprintf(DefaultTilesPath, req.UserUUID)

	resp, err := c.doRequest(ctx, "GET", path, nil, req.SessionToken)
	if err != nil {
		return nil, err
	}

	return decodeResponse[tilemodels.GetTilesResponse](resp)
}
