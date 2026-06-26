package client

import (
	"context"
	"net/http"
)

// CreateServiceAccount: POST /service_accounts
func (c *Client) CreateServiceAccount(ctx context.Context, in CreateServiceAccountRequest) (*ServiceAccount, error) {
	resp, err := doJSON[ServiceAccountResponse](ctx, c, http.MethodPost, "/service_accounts", in)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceAccount, nil
}

// GetServiceAccount: GET /service_accounts/{id}
func (c *Client) GetServiceAccount(ctx context.Context, id string) (*ServiceAccount, error) {
	resp, err := doJSON[ServiceAccountResponse](ctx, c, http.MethodGet, "/service_accounts/"+id, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceAccount, nil
}

// UpdateServiceAccount: PATCH /service_accounts/{id}
func (c *Client) UpdateServiceAccount(ctx context.Context, id string, in UpdateServiceAccountRequest) (*ServiceAccount, error) {
	resp, err := doJSON[ServiceAccountResponse](ctx, c, http.MethodPatch, "/service_accounts/"+id, in)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceAccount, nil
}

// DeleteServiceAccount: DELETE /service_accounts/{id}
func (c *Client) DeleteServiceAccount(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, "/service_accounts/"+id, nil, nil)
}
