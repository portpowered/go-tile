package tile

import (
	"context"

	"github.com/portpowered/go-tile/pkg/tilemodels"
)

// TileClient defines the operations available on the Tile API client.
// Consumers should define their own narrower interfaces; this documents the full API surface.
type TileClient interface {
	// Auth
	Login(ctx context.Context, req tilemodels.LoginRequest) (*tilemodels.LoginResponse, error)

	// Device discovery and location
	GetTiles(ctx context.Context, req tilemodels.GetTilesRequest) (*tilemodels.GetTilesResponse, error)

	// Lifecycle
	Close() error
}

// Verify that Client implements TileClient at compile time.
var _ TileClient = (*Client)(nil)
