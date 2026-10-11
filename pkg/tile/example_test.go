package tile_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/portpowered/go-tile/pkg/tile"
	"github.com/portpowered/go-tile/pkg/tilemodels"
)

func ExampleClient() {
	mux := http.NewServeMux()
	mux.HandleFunc("/clients/example-client", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(tilemodels.LoginResponse{
			ClientUUID:   "example-client",
			SessionToken: "example-session",
			UserUUID:     "example-user",
		})
	})
	mux.HandleFunc("/users/example-user/user_tiles", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(tilemodels.GetTilesResponse{
			TileCount: 1,
			Tiles: []tilemodels.TileResult{
				{
					UUID:     "tile-1",
					Name:     "Keys",
					TileType: "tile_pro",
					Visible:  true,
				},
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := tile.NewClient(
		tile.WithBaseURL(server.URL),
		tile.WithUUIDGenerator(func() string { return "example-client" }),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()
	login, err := client.Login(ctx, tilemodels.LoginRequest{
		Email:    "user@example.test",
		Password: "redacted",
	})
	if err != nil {
		panic(err)
	}

	tiles, err := client.GetTiles(ctx, tilemodels.GetTilesRequest{
		SessionToken: login.SessionToken,
		UserUUID:     login.UserUUID,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s:%s\n", tiles.Tiles[0].UUID, tiles.Tiles[0].Name)

	// Output:
	// tile-1:Keys
}
