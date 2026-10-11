// Package mocks provides test doubles for HTTP interactions with the Tile API.
package mocks

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// RequestRecord captures the details of an HTTP request made through the mock transport.
type RequestRecord struct {
	Method      string
	URL         string
	Headers     http.Header
	Body        []byte
	BodyString  string
	QueryParams map[string]string
}

// MockTransport implements http.RoundTripper to intercept HTTP requests in tests.
type MockTransport struct {
	mu              sync.RWMutex
	requests        []*RequestRecord
	responses       map[string]*http.Response
	errors          map[string]error
	fixtureDir      string
	defaultResponse *http.Response
}

// NewMockTransport creates a transport that loads fixture responses from the given directory.
func NewMockTransport(fixtureDir string) *MockTransport {
	return &MockTransport{
		requests:   make([]*RequestRecord, 0),
		responses:  make(map[string]*http.Response),
		errors:     make(map[string]error),
		fixtureDir: fixtureDir,
	}
}

// RoundTrip implements http.RoundTripper.
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	_ = m.recordRequest(req)

	key := m.getRequestKey(req)
	m.mu.RLock()
	transportErr, hasErr := m.errors[key]
	response, exists := m.responses[key]
	m.mu.RUnlock()
	if hasErr {
		return nil, transportErr
	}
	if exists {
		return cloneResponse(response), nil
	}

	if m.fixtureDir != "" {
		if resp := m.loadFixtureResponse(req); resp != nil {
			return resp, nil
		}
	}

	if m.defaultResponse != nil {
		return cloneResponse(m.defaultResponse), nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"error":"not found"}`)),
		Header:     http.Header{"Content-Type": {"application/json"}},
	}, nil
}

// SetResponse registers a canned response for a given method+path combination.
func (m *MockTransport) SetResponse(method, path string, resp *http.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[method+" "+path] = resp
}

// SetError registers a transport error for a given method+path combination.
func (m *MockTransport) SetError(method, path string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method+" "+path] = err
}

// SetResponseWithBody is a convenience method that creates a JSON response.
func (m *MockTransport) SetResponseWithBody(method, path string, statusCode int, body interface{}) {
	bodyBytes, _ := json.Marshal(body)
	resp := &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBuffer(bodyBytes)),
		Header:     http.Header{"Content-Type": {"application/json"}},
	}
	m.SetResponse(method, path, resp)
}

// SetResponseFromFixture registers a canned response loaded from a fixture file.
func (m *MockTransport) SetResponseFromFixture(method, path, fixtureName string) error {
	resp, err := m.responseFromFixture(fixtureName)
	if err != nil {
		return err
	}
	m.SetResponse(method, path, resp)
	return nil
}

// SetDefaultResponse sets the fallback response when no match is found.
func (m *MockTransport) SetDefaultResponse(resp *http.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultResponse = resp
}

// GetRequests returns a copy of all recorded requests.
func (m *MockTransport) GetRequests() []*RequestRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*RequestRecord, len(m.requests))
	copy(out, m.requests)
	return out
}

// FindRequest finds the first request matching the given method and path substring.
func (m *MockTransport) FindRequest(method, pathContains string) *RequestRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, r := range m.requests {
		if r.Method == method && strings.Contains(r.URL, pathContains) {
			return r
		}
	}
	return nil
}

// ClearRequests resets the recorded request history.
func (m *MockTransport) ClearRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = make([]*RequestRecord, 0)
}

func (m *MockTransport) recordRequest(req *http.Request) *RequestRecord {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	queryParams := make(map[string]string)
	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	record := &RequestRecord{
		Method:      req.Method,
		URL:         req.URL.String(),
		Headers:     req.Header.Clone(),
		Body:        bodyBytes,
		BodyString:  string(bodyBytes),
		QueryParams: queryParams,
	}

	m.mu.Lock()
	m.requests = append(m.requests, record)
	m.mu.Unlock()

	return record
}

func (m *MockTransport) getRequestKey(req *http.Request) string {
	return req.Method + " " + req.URL.Path
}

// loadFixtureResponse loads a fixture file matching the request.
// Convention: /api/v1/clients/{uuid} → tile_clients.json, /api/v1/users/{uuid}/user_tiles → tile_user_tiles.json
func (m *MockTransport) loadFixtureResponse(req *http.Request) *http.Response {
	path := strings.TrimPrefix(req.URL.Path, "/")
	// Remove api/v1 prefix
	path = strings.TrimPrefix(path, "api/v1/")
	// Split by / and take the first meaningful segment
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil
	}

	// Try "tile_{first_segment}.json" then "tile_{last_segment}.json"
	candidates := []string{
		filepath.Join(m.fixtureDir, "tile_"+parts[0]+".json"),
	}
	if len(parts) > 1 {
		candidates = append(candidates, filepath.Join(m.fixtureDir, "tile_"+parts[len(parts)-1]+".json"))
	}

	for _, fixturePath := range candidates {
		resp, err := m.responseFromFixturePath(fixturePath)
		if err == nil {
			return resp
		}
	}

	return nil
}

func (m *MockTransport) responseFromFixture(fixtureName string) (*http.Response, error) {
	return m.responseFromFixturePath(filepath.Join(m.fixtureDir, fixtureName))
}

func (m *MockTransport) responseFromFixturePath(fixturePath string) (*http.Response, error) {
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		return nil, err
	}

	statusCode := http.StatusOK
	headers := http.Header{"Content-Type": {"application/json"}}
	body := data

	var envelope struct {
		Meta struct {
			StatusCode int                 `json:"statusCode"`
			Headers    map[string][]string `json:"headers"`
		} `json:"_meta"`
		Body json.RawMessage `json:"_body"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil && envelope.Body != nil {
		if envelope.Meta.StatusCode != 0 {
			statusCode = envelope.Meta.StatusCode
		}
		for key, values := range envelope.Meta.Headers {
			headers[key] = values
		}
		body = envelope.Body
	}

	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     headers,
	}, nil
}

func cloneResponse(resp *http.Response) *http.Response {
	var body []byte
	if resp.Body != nil {
		body, _ = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	return &http.Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     resp.Header.Clone(),
	}
}
