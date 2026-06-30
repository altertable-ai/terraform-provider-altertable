package client

import (
	"context"
	"net/http"
)

// GET /environments/{idOrSlug} (the API accepts a UUID or a slug).
func (c *Client) GetEnvironment(ctx context.Context, idOrSlug string) (*Environment, error) {
	resp, err := request[EnvironmentResponse](ctx, c, http.MethodGet, "/environments/"+idOrSlug, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Environment, nil
}

// POST /environments
func (c *Client) CreateEnvironment(ctx context.Context, params CreateEnvironmentRequest) (*Environment, error) {
	resp, err := request[EnvironmentResponse](ctx, c, http.MethodPost, "/environments", params)
	if err != nil {
		return nil, err
	}
	return &resp.Environment, nil
}

// DELETE /environments/{id}
func (c *Client) DeleteEnvironment(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, "/environments/"+id, nil, nil)
}
