package tilemodels

// LoginRequest represents the credentials sent to the Tile auth endpoint.
type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	ClientUUID string `json:"-"`
}

// LoginResponse contains the session data returned after successful authentication.
type LoginResponse struct {
	Version        int    `json:"version"`
	Revision       int    `json:"revision"`
	Timestamp      string `json:"timestamp"`
	TimestampMs    int64  `json:"timestamp_ms"`
	ResultCode     int    `json:"result_code"`
	Result         string `json:"result"`
	ClientUUID     string `json:"client_uuid"`
	SessionToken   string `json:"session_token"`
	UserUUID       string `json:"user_uuid"`
	SessionStartTs string `json:"session_start_timestamp"`
}

// TileState represents the current state of a Tile device.
type TileState struct {
	UUID          string  `json:"uuid"`
	Timestamp     int64   `json:"timestamp"`
	TimestampMs   int64   `json:"timestamp_ms"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Altitude      float64 `json:"altitude"`
	IsLost        bool    `json:"is_lost"`
	Ring          string  `json:"ring_state_set_timestamp"`
	ConnectionSt  string  `json:"connection_state_timestamp"`
	Voip          string  `json:"voip_state"`
	BatteryLevel  float64 `json:"battery_level"`
	IsBatteryDead bool    `json:"is_battery_dead"`
}

// TileResult represents a single Tile device from the API.
type TileResult struct {
	UUID            string    `json:"uuid"`
	Name            string    `json:"name"`
	TileType        string    `json:"tile_type"`
	FirmwareVersion string    `json:"firmware_version"`
	Visible         bool      `json:"visible"`
	LastTileState   TileState `json:"last_tile_state"`
}

// GetTilesResponse contains the list of tiles returned by the API.
type GetTilesResponse struct {
	Version     int          `json:"version"`
	Revision    int          `json:"revision"`
	Timestamp   string       `json:"timestamp"`
	TimestampMs int64        `json:"timestamp_ms"`
	ResultCode  int          `json:"result_code"`
	Result      string       `json:"result"`
	TileCount   int          `json:"num_tiles"`
	Tiles       []TileResult `json:"result_list"`
}

// GetTilesRequest contains the explicit session inputs needed to list a user's
// Tile devices.
type GetTilesRequest struct {
	SessionToken string
	UserUUID     string
}
