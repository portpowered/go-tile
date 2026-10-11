package tilemodels

import (
	"errors"
	"fmt"
)

// AuthenticationError indicates the Tile API rejected the provided credentials.
type AuthenticationError struct {
	Message string
	Status  int
}

func NewAuthenticationError(message string, status int) *AuthenticationError {
	return &AuthenticationError{Message: message, Status: status}
}

func (e *AuthenticationError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("authentication error: %s", e.Message)
	}
	return fmt.Sprintf("authentication error (status: %d)", e.Status)
}

func IsAuthenticationError(err error) bool {
	var target *AuthenticationError
	return errors.As(err, &target)
}

// NetworkError indicates a connectivity or transport failure.
type NetworkError struct {
	Message string
	Err     error
}

func NewNetworkError(message string, err error) *NetworkError {
	return &NetworkError{Message: message, Err: err}
}

func (e *NetworkError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("network error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("network error: %s", e.Message)
}

func (e *NetworkError) Unwrap() error { return e.Err }

func IsNetworkError(err error) bool {
	var target *NetworkError
	return errors.As(err, &target)
}

// NotFoundError indicates the requested resource was not found.
type NotFoundError struct {
	Message string
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{Message: message}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.Message)
}

func IsNotFoundError(err error) bool {
	var target *NotFoundError
	return errors.As(err, &target)
}

// HTTPError represents a non-success HTTP response from the Tile API.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func NewHTTPError(statusCode int, status, body string) *HTTPError {
	return &HTTPError{StatusCode: statusCode, Status: status, Body: body}
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error %d: %s", e.StatusCode, e.Status)
}

func IsHTTPError(err error) bool {
	var target *HTTPError
	return errors.As(err, &target)
}

// BadRequestError indicates a malformed or invalid request.
type BadRequestError struct {
	Message string
	Err     error
}

func NewBadRequestError(message string, err error) *BadRequestError {
	return &BadRequestError{Message: message, Err: err}
}

func (e *BadRequestError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("bad request: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("bad request: %s", e.Message)
}

func (e *BadRequestError) Unwrap() error { return e.Err }

func IsBadRequestError(err error) bool {
	var target *BadRequestError
	return errors.As(err, &target)
}

// RateLimitError indicates Tile throttled the request.
type RateLimitError struct {
	Message    string
	StatusCode int
	RetryAfter string
}

func NewRateLimitError(message string, statusCode int, retryAfter string) *RateLimitError {
	return &RateLimitError{Message: message, StatusCode: statusCode, RetryAfter: retryAfter}
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter != "" {
		return fmt.Sprintf("rate limit error (status: %d, retry-after: %s): %s", e.StatusCode, e.RetryAfter, e.Message)
	}
	return fmt.Sprintf("rate limit error (status: %d): %s", e.StatusCode, e.Message)
}

func IsRateLimitError(err error) bool {
	var target *RateLimitError
	return errors.As(err, &target)
}

// TokenError indicates required Tile session material is missing or invalid.
type TokenError struct {
	Message string
}

func NewTokenError(message string) *TokenError {
	return &TokenError{Message: message}
}

func (e *TokenError) Error() string {
	return fmt.Sprintf("token error: %s", e.Message)
}

func IsTokenError(err error) bool {
	var target *TokenError
	return errors.As(err, &target)
}

// MalformedResponseError indicates Tile returned a response body that could not
// be decoded into the expected vendor model.
type MalformedResponseError struct {
	Message string
	Err     error
}

func NewMalformedResponseError(message string, err error) *MalformedResponseError {
	return &MalformedResponseError{Message: message, Err: err}
}

func (e *MalformedResponseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("malformed response: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("malformed response: %s", e.Message)
}

func (e *MalformedResponseError) Unwrap() error { return e.Err }

func IsMalformedResponseError(err error) bool {
	var target *MalformedResponseError
	return errors.As(err, &target)
}

// ClosedError indicates a request was attempted after the client was closed.
type ClosedError struct {
	Message string
}

func NewClosedError(message string) *ClosedError {
	return &ClosedError{Message: message}
}

func (e *ClosedError) Error() string {
	return fmt.Sprintf("closed client: %s", e.Message)
}

func IsClosedError(err error) bool {
	var target *ClosedError
	return errors.As(err, &target)
}
