package client

import (
	"context"
	"net/http"
)

func databasesPath(envID string) string { return "/environments/" + envID + "/databases" }

// ListDatabases: GET /environments/{envID}/databases
func (c *Client) ListDatabases(ctx context.Context, envID string) ([]Database, error) {
	resp, err := doJSON[DatabasesListResponse](ctx, c, http.MethodGet, databasesPath(envID), nil)
	if err != nil {
		return nil, err
	}
	return resp.Databases, nil
}

// GetDatabase: GET /environments/{envID}/databases/{idOrSlug}
func (c *Client) GetDatabase(ctx context.Context, envID, idOrSlug string) (*Database, error) {
	resp, err := doJSON[DatabaseResponse](ctx, c, http.MethodGet, databasesPath(envID)+"/"+idOrSlug, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Database, nil
}

// CreateDatabase: POST /environments/{envID}/databases
func (c *Client) CreateDatabase(ctx context.Context, envID string, in CreateDatabaseRequest) (*Database, error) {
	resp, err := doJSON[DatabaseResponse](ctx, c, http.MethodPost, databasesPath(envID), in)
	if err != nil {
		return nil, err
	}
	return &resp.Database, nil
}

// UpdateDatabase: PATCH /environments/{envID}/databases/{id}
func (c *Client) UpdateDatabase(ctx context.Context, envID, id string, in UpdateDatabaseRequest) (*Database, error) {
	resp, err := doJSON[DatabaseResponse](ctx, c, http.MethodPatch, databasesPath(envID)+"/"+id, in)
	if err != nil {
		return nil, err
	}
	return &resp.Database, nil
}

// DeleteDatabase: DELETE /environments/{envID}/databases/{id}
func (c *Client) DeleteDatabase(ctx context.Context, envID, id string) error {
	return c.doRequest(ctx, http.MethodDelete, databasesPath(envID)+"/"+id, nil, nil)
}
