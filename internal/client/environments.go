package client

import "context"

// GetEnvironmentBySlug: GET /v1/environments?slug={slug}
func (c *Client) GetEnvironmentBySlug(ctx context.Context, slug string) (*Environment, error) {
	// TODO: return c.doRequest(ctx, http.MethodGet, "/v1/environments?slug="+url.QueryEscape(slug), nil, &out)
	return nil, ErrNotImplemented
}

// GetEnvironment: GET /v1/environments/{id}
func (c *Client) GetEnvironment(ctx context.Context, id string) (*Environment, error) {
	return nil, ErrNotImplemented
}

// CreateEnvironment: POST /v1/environments
func (c *Client) CreateEnvironment(ctx context.Context, in EnvironmentCreateInput) (*Environment, error) {
	return nil, ErrNotImplemented
}

// UpdateEnvironment: PATCH /v1/environments/{id}
func (c *Client) UpdateEnvironment(ctx context.Context, id string, in EnvironmentUpdateInput) (*Environment, error) {
	return nil, ErrNotImplemented
}

// DeleteEnvironment: DELETE /v1/environments/{id}
func (c *Client) DeleteEnvironment(ctx context.Context, id string) error {
	return ErrNotImplemented
}
