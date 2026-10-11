// Package tile provides a Go client for the Tile Bluetooth tracker cloud API.
//
// The Tile API uses email/password authentication to obtain a session token.
// That session token is then used for subsequent API calls (listing tiles,
// getting locations). Tokens are injected per-request to keep the client
// stateless for Port OS plugin usage.
package tile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/portpowered/go-tile/pkg/tilemodels"
)

// Client is the top-level Tile API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	locale     string
	newUUID    func() string

	mu     sync.RWMutex
	closed bool
}

// Option configures a Client.
type Option interface {
	Apply(c *Client) error
}

// NewClient creates a new Tile API client with the given options.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
		baseURL:    DefaultBaseURL,
		userAgent:  DefaultUserAgent,
		locale:     DefaultLocale,
	}

	for _, opt := range opts {
		if opt == nil {
			return nil, tilemodels.NewBadRequestError("client option must not be nil", nil)
		}
		if err := opt.Apply(c); err != nil {
			return nil, fmt.Errorf("failed to apply client option: %w", err)
		}
	}

	if c.httpClient == nil {
		return nil, tilemodels.NewBadRequestError("HTTP client must not be nil", nil)
	}
	if c.newUUID == nil {
		c.newUUID = defaultUUID
	}

	return c, nil
}

// Close releases any resources held by the client.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

// --- Option implementations ---

type withHTTPClient struct{ client *http.Client }

func (w withHTTPClient) Apply(c *Client) error {
	if w.client == nil {
		return tilemodels.NewBadRequestError("HTTP client must not be nil", nil)
	}
	c.httpClient = w.client
	return nil
}

// WithHTTPClient sets a custom HTTP client (useful for injecting mock transports).
func WithHTTPClient(client *http.Client) Option {
	return withHTTPClient{client: client}
}

type withBaseURL string

func (w withBaseURL) Apply(c *Client) error {
	rawURL := strings.TrimSpace(string(w))
	if rawURL == "" {
		return tilemodels.NewBadRequestError("base URL must not be empty", nil)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return tilemodels.NewBadRequestError("base URL must be valid", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return tilemodels.NewBadRequestError("base URL must include scheme and host", nil)
	}
	c.baseURL = strings.TrimRight(rawURL, "/")
	return nil
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return withBaseURL(url)
}

type withUserAgent string

func (w withUserAgent) Apply(c *Client) error {
	userAgent := strings.TrimSpace(string(w))
	if userAgent == "" {
		return tilemodels.NewBadRequestError("user agent must not be empty", nil)
	}
	c.userAgent = userAgent
	return nil
}

// WithUserAgent sets the User-Agent header sent to Tile.
func WithUserAgent(userAgent string) Option {
	return withUserAgent(userAgent)
}

type withLocale string

func (w withLocale) Apply(c *Client) error {
	locale := strings.TrimSpace(string(w))
	if locale == "" {
		return tilemodels.NewBadRequestError("locale must not be empty", nil)
	}
	c.locale = locale
	return nil
}

// WithLocale sets the Accept-Language header sent to Tile.
func WithLocale(locale string) Option {
	return withLocale(locale)
}

type withUUIDGenerator struct{ generate func() string }

func (w withUUIDGenerator) Apply(c *Client) error {
	if w.generate == nil {
		return tilemodels.NewBadRequestError("UUID generator must not be nil", nil)
	}
	c.newUUID = w.generate
	return nil
}

// WithUUIDGenerator sets the client UUID generator used when LoginRequest does
// not provide a ClientUUID.
func WithUUIDGenerator(generate func() string) Option {
	return withUUIDGenerator{generate: generate}
}

// --- Internal HTTP helpers ---

// doRequest executes an HTTP request against the Tile API.
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader, sessionToken string) (*http.Response, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, tilemodels.NewClosedError("request attempted after Close")
	}
	baseURL := c.baseURL
	userAgent := c.userAgent
	locale := c.locale
	c.mu.RUnlock()

	url := baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, tilemodels.NewNetworkError("failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Tile_app_id", TileAppID)
	req.Header.Set("Tile_app_version", TileAppVersion)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", locale)

	if sessionToken != "" {
		req.Header.Set("Tile_session_token", sessionToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, tilemodels.NewNetworkError("request failed", err)
	}

	return resp, nil
}

// decodeResponse reads and decodes a JSON response body.
func decodeResponse[T any](resp *http.Response) (*T, error) {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, tilemodels.NewNetworkError("failed to read response body", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, tilemodels.NewAuthenticationError("Tile returned unauthorized response", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, tilemodels.NewAuthenticationError("Tile returned forbidden response", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return nil, tilemodels.NewBadRequestError("Tile returned bad request response", nil)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, tilemodels.NewRateLimitError("Tile returned rate limit response", resp.StatusCode, resp.Header.Get("Retry-After"))
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, tilemodels.NewNotFoundError("Tile returned not found response")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, tilemodels.NewHTTPError(resp.StatusCode, resp.Status, "")
	}

	var result T
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, tilemodels.NewMalformedResponseError("failed to decode response", err)
	}

	return &result, nil
}

func defaultUUID() string {
	return uuid.New().String()
}
