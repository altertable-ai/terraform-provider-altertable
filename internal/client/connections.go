package client

import (
	"context"
	"net/http"
)

func connectionsPath(envID string) string { return "/environments/" + envID + "/connections" }

// GET /environments/{envID}/connections
func (c *Client) ListConnections(ctx context.Context, envID string) ([]Connection, error) {
	resp, err := request[ConnectionsListResponse](ctx, c, http.MethodGet, connectionsPath(envID), nil)
	if err != nil {
		return nil, err
	}
	return resp.Connections, nil
}

// GET /environments/{envID}/connections/{idOrSlug}
func (c *Client) GetConnection(ctx context.Context, envID, idOrSlug string) (*Connection, error) {
	resp, err := request[ConnectionResponse](ctx, c, http.MethodGet, connectionsPath(envID)+"/"+idOrSlug, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Connection, nil
}

// POST /environments/{envID}/connections
func (c *Client) CreateConnection(ctx context.Context, envID string, params CreateConnectionRequest) (*Connection, error) {
	resp, err := request[ConnectionResponse](ctx, c, http.MethodPost, connectionsPath(envID), params)
	if err != nil {
		return nil, err
	}
	return &resp.Connection, nil
}

// PATCH /environments/{envID}/connections/{id}
func (c *Client) UpdateConnection(ctx context.Context, envID, id string, params UpdateConnectionRequest) (*Connection, error) {
	resp, err := request[ConnectionResponse](ctx, c, http.MethodPatch, connectionsPath(envID)+"/"+id, params)
	if err != nil {
		return nil, err
	}
	return &resp.Connection, nil
}

// DELETE /environments/{envID}/connections/{id}
func (c *Client) DeleteConnection(ctx context.Context, envID, id string) error {
	return c.doRequest(ctx, http.MethodDelete, connectionsPath(envID)+"/"+id, nil, nil)
}
