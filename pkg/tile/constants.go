package tile

const (
	// DefaultBaseURL is the production Tile API base URL.
	DefaultBaseURL = "https://production.tile-api.com/api/v1"

	// DefaultLoginPath is the authentication endpoint path.
	DefaultLoginPath = "/clients/%s"

	// DefaultTilesPath is the endpoint for listing all tiles.
	DefaultTilesPath = "/users/%s/user_tiles"

	// TileAppID identifies the Tile mobile application.
	TileAppID = "ios-tile-production"

	// TileAppVersion is the version string sent with API requests.
	TileAppVersion = "2.88.1.4450"

	// DefaultUserAgent is the User-Agent sent when callers do not override it.
	DefaultUserAgent = "PortOS go-tile"

	// DefaultLocale is the locale sent when callers do not override it.
	DefaultLocale = "en-US"
)
