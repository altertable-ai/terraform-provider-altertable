package client

import (
	"context"
	"net/http"
)

// GET /whoami. Identifies the principal and organization behind the API
// key. Used to validate credentials at provider configure time: a 401 means the
// key is invalid/expired, a 403 means it lacks management API access.
func (c *Client) Whoami(ctx context.Context) (*WhoamiResponse, error) {
	resp, err := request[WhoamiResponse](ctx, c, http.MethodGet, "/whoami", nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
