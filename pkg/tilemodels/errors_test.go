package tilemodels

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypedErrorHelpers_MatchWrappedErrors(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		check func(error) bool
	}{
		{name: "authentication", err: NewAuthenticationError("unauthorized", 401), check: IsAuthenticationError},
		{name: "bad request", err: NewBadRequestError("invalid", nil), check: IsBadRequestError},
		{name: "network", err: NewNetworkError("request failed", errors.New("dial tcp")), check: IsNetworkError},
		{name: "http", err: NewHTTPError(500, "500 Internal Server Error", "{}"), check: IsHTTPError},
		{name: "rate limit", err: NewRateLimitError("throttled", 429, "30"), check: IsRateLimitError},
		{name: "token", err: NewTokenError("missing session token"), check: IsTokenError},
		{name: "not found", err: NewNotFoundError("missing"), check: IsNotFoundError},
		{name: "malformed response", err: NewMalformedResponseError("decode failed", errors.New("bad json")), check: IsMalformedResponseError},
		{name: "closed", err: NewClosedError("closed"), check: IsClosedError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.check(fmt.Errorf("wrapped: %w", tt.err)))
		})
	}
}
