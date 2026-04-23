package workflowy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrMissingAPIKey is returned when no API key is provided.
var ErrMissingAPIKey = errors.New("workflowy: API key is required")

// Error represents an error returned by the WorkFlowy API.
type Error struct {
	StatusCode int
	Message    string
	RawBody    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("workflowy: %d - %s", e.StatusCode, e.Message)
}

// IsNotFound reports whether the error is a 404 Not Found.
func IsNotFound(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsRateLimited reports whether the error is a 429 Too Many Requests.
func IsRateLimited(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

func parseAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	msg := http.StatusText(resp.StatusCode)
	var parsed struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &parsed) == nil && parsed.Message != "" {
		msg = parsed.Message
	}
	return &Error{
		StatusCode: resp.StatusCode,
		Message:    msg,
		RawBody:    string(body),
	}
}
