package client

import (
	"context"
	"net/http"
)

// GetEnvironment: GET /environments/{idOrSlug} (the API accepts a UUID or a slug).
func (c *Client) GetEnvironment(ctx context.Context, idOrSlug string) (*Environment, error) {
	resp, err := doJSON[EnvironmentResponse](ctx, c, http.MethodGet, "/environments/"+idOrSlug, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Environment, nil
}

// CreateEnvironment: POST /environments
func (c *Client) CreateEnvironment(ctx context.Context, in CreateEnvironmentRequest) (*Environment, error) {
	resp, err := doJSON[EnvironmentResponse](ctx, c, http.MethodPost, "/environments", in)
	if err != nil {
		return nil, err
	}
	return &resp.Environment, nil
}

// DeleteEnvironment: DELETE /environments/{id}
func (c *Client) DeleteEnvironment(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, "/environments/"+id, nil, nil)
}
