package client

import (
	"context"
	"net/http"
)

// POST /service_accounts
func (c *Client) CreateServiceAccount(ctx context.Context, params CreateServiceAccountRequest) (*ServiceAccount, error) {
	resp, err := request[ServiceAccountResponse](ctx, c, http.MethodPost, "/service_accounts", params)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceAccount, nil
}

// GET /service_accounts/{id}
func (c *Client) GetServiceAccount(ctx context.Context, id string) (*ServiceAccount, error) {
	resp, err := request[ServiceAccountResponse](ctx, c, http.MethodGet, "/service_accounts/"+id, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceAccount, nil
}

// PATCH /service_accounts/{id}
func (c *Client) UpdateServiceAccount(ctx context.Context, id string, params UpdateServiceAccountRequest) (*ServiceAccount, error) {
	resp, err := request[ServiceAccountResponse](ctx, c, http.MethodPatch, "/service_accounts/"+id, params)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceAccount, nil
}

// DELETE /service_accounts/{id}
func (c *Client) DeleteServiceAccount(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, "/service_accounts/"+id, nil, nil)
}
