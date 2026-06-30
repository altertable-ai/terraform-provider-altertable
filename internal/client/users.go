package client

import (
	"context"
	"net/http"
	"net/url"
)

// GET /users/lookup?email={email} — looks up a user by email.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	resp, err := request[UserResponse](ctx, c, http.MethodGet, "/users/lookup?email="+url.QueryEscape(email), nil)
	if err != nil {
		return nil, err
	}
	return &resp.User, nil
}
