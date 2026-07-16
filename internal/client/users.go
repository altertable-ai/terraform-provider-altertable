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

// POST /users — invites a user to the organization, or reuses a pending invitation for the
// same email. The returned User carries the stable iac_id to key the resource on.
func (c *Client) CreateUser(ctx context.Context, params CreateUserRequest) (*User, error) {
	resp, err := request[UserResponse](ctx, c, http.MethodPost, "/users", params)
	if err != nil {
		return nil, err
	}
	return &resp.User, nil
}

// GET /users/{iacID} — resolves a user or pending invitation by its stable iac_id.
func (c *Client) GetUser(ctx context.Context, iacID string) (*User, error) {
	resp, err := request[UserResponse](ctx, c, http.MethodGet, "/users/"+iacID, nil)
	if err != nil {
		return nil, err
	}
	return &resp.User, nil
}

// DELETE /users/{iacID} — cancels a pending invitation or removes the member from the
// organization. Absent users return 204, so callers need not special-case a missing target.
func (c *Client) DeleteUser(ctx context.Context, iacID string) error {
	return c.doRequest(ctx, http.MethodDelete, "/users/"+iacID, nil, nil)
}
