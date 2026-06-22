package client

import "context"

// GetRoleSet: GET /v1/{users|service-accounts}/{principal_id}/roles
func (c *Client) GetRoleSet(ctx context.Context, p PrincipalRef) (*RoleSet, error) {
	return nil, ErrNotImplemented
}

// PutRoleSet replaces the full set of grants for a principal (idempotent):
// PUT /v1/{users|service-accounts}/{principal_id}/roles
func (c *Client) PutRoleSet(ctx context.Context, p PrincipalRef, grants []RoleGrant) (*RoleSet, error) {
	return nil, ErrNotImplemented
}

// DeleteRoleSet clears all grants for a principal:
// DELETE /v1/{users|service-accounts}/{principal_id}/roles
func (c *Client) DeleteRoleSet(ctx context.Context, p PrincipalRef) error {
	return ErrNotImplemented
}
