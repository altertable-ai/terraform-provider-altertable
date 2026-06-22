package client

import "context"

// GetServiceAccountByName: GET /v1/service-accounts?name={name}
func (c *Client) GetServiceAccountByName(ctx context.Context, name string) (*ServiceAccount, error) {
	return nil, ErrNotImplemented
}

// GetServiceAccount: GET /v1/service-accounts/{id}
func (c *Client) GetServiceAccount(ctx context.Context, id string) (*ServiceAccount, error) {
	return nil, ErrNotImplemented
}

// CreateServiceAccount: POST /v1/service-accounts
func (c *Client) CreateServiceAccount(ctx context.Context, in ServiceAccountCreateInput) (*ServiceAccount, error) {
	return nil, ErrNotImplemented
}

// UpdateServiceAccount: PATCH /v1/service-accounts/{id}
func (c *Client) UpdateServiceAccount(ctx context.Context, id string, in ServiceAccountUpdateInput) (*ServiceAccount, error) {
	return nil, ErrNotImplemented
}

// DeleteServiceAccount: DELETE /v1/service-accounts/{id}
func (c *Client) DeleteServiceAccount(ctx context.Context, id string) error {
	return ErrNotImplemented
}
