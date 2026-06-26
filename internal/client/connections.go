package client

import (
	"context"
	"net/http"
)

func connectionsPath(envID string) string { return "/environments/" + envID + "/connections" }

// ListConnections: GET /environments/{envID}/connections
func (c *Client) ListConnections(ctx context.Context, envID string) ([]Connection, error) {
	resp, err := doJSON[ConnectionsListResponse](ctx, c, http.MethodGet, connectionsPath(envID), nil)
	if err != nil {
		return nil, err
	}
	return resp.Connections, nil
}

// GetConnection: GET /environments/{envID}/connections/{idOrSlug}
func (c *Client) GetConnection(ctx context.Context, envID, idOrSlug string) (*Connection, error) {
	resp, err := doJSON[ConnectionResponse](ctx, c, http.MethodGet, connectionsPath(envID)+"/"+idOrSlug, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Connection, nil
}

// CreateConnection: POST /environments/{envID}/connections
func (c *Client) CreateConnection(ctx context.Context, envID string, in CreateConnectionRequest) (*Connection, error) {
	resp, err := doJSON[ConnectionResponse](ctx, c, http.MethodPost, connectionsPath(envID), in)
	if err != nil {
		return nil, err
	}
	return &resp.Connection, nil
}

// UpdateConnection: PATCH /environments/{envID}/connections/{id}
func (c *Client) UpdateConnection(ctx context.Context, envID, id string, in UpdateConnectionRequest) (*Connection, error) {
	resp, err := doJSON[ConnectionResponse](ctx, c, http.MethodPatch, connectionsPath(envID)+"/"+id, in)
	if err != nil {
		return nil, err
	}
	return &resp.Connection, nil
}

// DeleteConnection: DELETE /environments/{envID}/connections/{id}
func (c *Client) DeleteConnection(ctx context.Context, envID, id string) error {
	return c.doRequest(ctx, http.MethodDelete, connectionsPath(envID)+"/"+id, nil, nil)
}
