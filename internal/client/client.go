package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL is used when neither provider config nor ALTERTABLE_API_URL is set.
const DefaultBaseURL = "https://api.altertable.com"

// ErrNotImplemented is returned by every entity method until the REST API is wired up.
var ErrNotImplemented = errors.New("altertable: client method not implemented")

// Client talks to the Altertable management API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	userAgent  string
}

// NewClient builds a Client. baseURL has any trailing slash trimmed.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		userAgent:  "terraform-provider-altertable",
	}
}

// APIError is returned for any non-2xx response.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("altertable API error: status %d", e.StatusCode)
	}
	return fmt.Sprintf("altertable API error: status %d: %s", e.StatusCode, e.Message)
}

// doRequest performs an authenticated JSON request. If body is non-nil it is JSON-encoded;
// if out is non-nil a 2xx response body is decoded into it.
func (c *Client) doRequest(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var payload struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(res.Body).Decode(&payload)
		return &APIError{StatusCode: res.StatusCode, Message: payload.Message}
	}

	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
