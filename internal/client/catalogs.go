package client

import "context"

// GetCatalogBySlug: GET /v1/catalogs?environment_id={env}&slug={slug}
func (c *Client) GetCatalogBySlug(ctx context.Context, environmentID, slug string) (*Catalog, error) {
	return nil, ErrNotImplemented
}

// GetCatalog: GET /v1/catalogs/{id}
func (c *Client) GetCatalog(ctx context.Context, id string) (*Catalog, error) {
	return nil, ErrNotImplemented
}

// CreateCatalog: POST /v1/catalogs
func (c *Client) CreateCatalog(ctx context.Context, in CatalogCreateInput) (*Catalog, error) {
	return nil, ErrNotImplemented
}

// UpdateCatalog: PATCH /v1/catalogs/{id}
func (c *Client) UpdateCatalog(ctx context.Context, id string, in CatalogUpdateInput) (*Catalog, error) {
	return nil, ErrNotImplemented
}

// DeleteCatalog: DELETE /v1/catalogs/{id}
func (c *Client) DeleteCatalog(ctx context.Context, id string) error {
	return ErrNotImplemented
}
