package client

import (
	"context"
	"net/http"
	"net/url"
)

// GetUserByEmail looks up a user by email: GET /users/lookup?email={email}.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	resp, err := doJSON[UserResponse](ctx, c, http.MethodGet, "/users/lookup?email="+url.QueryEscape(email), nil)
	if err != nil {
		return nil, err
	}
	return &resp.User, nil
}
