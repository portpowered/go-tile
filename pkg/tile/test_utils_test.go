package tile

import (
	"context"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/portpowered/go-tile/test/mocks"
)

// getFixtureDir returns the absolute path to the test fixtures directory.
func getFixtureDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "test", "fixtures")
}

// newTestClient creates a client wired to a MockTransport.
func newTestClient() (*Client, *mocks.MockTransport) {
	fixtureDir := getFixtureDir()
	mockTransport := mocks.NewMockTransport(fixtureDir)
	httpClient := &http.Client{Transport: mockTransport}

	client, err := NewClient(
		WithHTTPClient(httpClient),
		WithUUIDGenerator(func() string { return "known-uuid" }),
	)
	if err != nil {
		panic("failed to create test client: " + err.Error())
	}

	return client, mockTransport
}

// newTestContext returns a background context for use in tests.
func newTestContext() context.Context {
	return context.Background()
}
