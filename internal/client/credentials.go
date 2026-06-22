package client

import "context"

// CreateCredential: POST /v1/credentials — returns the secret once.
func (c *Client) CreateCredential(ctx context.Context, in CredentialCreateInput) (*Credential, error) {
	return nil, ErrNotImplemented
}

// GetCredential: GET /v1/credentials/{id} — metadata only, never the password.
func (c *Client) GetCredential(ctx context.Context, id string) (*Credential, error) {
	return nil, ErrNotImplemented
}

// UpdateCredential: PATCH /v1/credentials/{id}
func (c *Client) UpdateCredential(ctx context.Context, id string, in CredentialUpdateInput) (*Credential, error) {
	return nil, ErrNotImplemented
}

// DeleteCredential: DELETE /v1/credentials/{id}
func (c *Client) DeleteCredential(ctx context.Context, id string) error {
	return ErrNotImplemented
}
