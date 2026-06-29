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

// DefaultBaseURL is the production Altertable Management REST API root.
const DefaultBaseURL = "https://app.altertable.ai/rest/v1"

// ErrNotImplemented is returned by every entity method until the REST API is wired up.
var ErrNotImplemented = errors.New("altertable: client method not implemented")

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	userAgent  string
}

func NewClient(baseURL, apiKey, version string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		userAgent:  "terraform-provider-altertable/" + version,
	}
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    []string
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("altertable API error: status %d", e.StatusCode)
	if e.Code != "" {
		msg += fmt.Sprintf(" (%s)", e.Code)
	}
	if e.Message != "" {
		msg += ": " + e.Message
	}
	if len(e.Details) > 0 {
		msg += " [" + strings.Join(e.Details, "; ") + "]"
	}
	return msg
}

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
			Error struct {
				Code    string   `json:"code"`
				Message string   `json:"message"`
				Details []string `json:"details"`
			} `json:"error"`
		}
		_ = json.NewDecoder(res.Body).Decode(&payload)
		return &APIError{
			StatusCode: res.StatusCode,
			Code:       payload.Error.Code,
			Message:    payload.Error.Message,
			Details:    payload.Error.Details,
		}
	}

	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// doJSON issues a request and decodes the JSON response into T. It keeps every
// entity method to a single line: doJSON[EnvironmentResponse](ctx, c, "POST", "/environments", in).
func doJSON[T any](ctx context.Context, c *Client, method, path string, body any) (T, error) {
	var out T
	err := c.doRequest(ctx, method, path, body, &out)
	return out, err
}
