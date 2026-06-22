package client

import "context"

// GetUserByEmail: GET /v1/users?email={email}
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return nil, ErrNotImplemented
}

// GetUser: GET /v1/users/{id}
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	return nil, ErrNotImplemented
}

// CreateUser (invite): POST /v1/users
func (c *Client) CreateUser(ctx context.Context, in UserCreateInput) (*User, error) {
	return nil, ErrNotImplemented
}

// UpdateUser: PATCH /v1/users/{id}
func (c *Client) UpdateUser(ctx context.Context, id string, in UserUpdateInput) (*User, error) {
	return nil, ErrNotImplemented
}

// DeleteUser: DELETE /v1/users/{id}
func (c *Client) DeleteUser(ctx context.Context, id string) error {
	return ErrNotImplemented
}
