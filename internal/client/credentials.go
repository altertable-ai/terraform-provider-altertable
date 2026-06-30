package client

import (
	"context"
	"fmt"
	"net/http"
)

// credentialBasePath e.g.
// /users/2f1b9c4e-7a3d-4b8e-9c1a-5d6e7f8a9b0c/environments/8c7d6e5f-4a3b-2c1d-0e9f-8a7b6c5d4e3f/credentials or
// /service_accounts/a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d/environments/8c7d6e5f-4a3b-2c1d-0e9f-8a7b6c5d4e3f/credentials.
func credentialBasePath(principalType, principalID, envID string) (string, error) {
	switch principalType {
	case "user":
		return "/users/" + principalID + "/environments/" + envID + "/credentials", nil
	case "service_account":
		return "/service_accounts/" + principalID + "/environments/" + envID + "/credentials", nil
	default:
		return "", fmt.Errorf("altertable: invalid principal type %q (want \"user\" or \"service_account\")", principalType)
	}
}

// POST {base} — returns the credential and the one-time password.
func (c *Client) CreateCredential(ctx context.Context, principalType, principalID, envID string, params CreateCredentialRequest) (*Credential, string, error) {
	base, err := credentialBasePath(principalType, principalID, envID)
	if err != nil {
		return nil, "", err
	}
	resp, err := request[CreateCredentialResponse](ctx, c, http.MethodPost, base, params)
	if err != nil {
		return nil, "", err
	}
	return &resp.Credential, resp.Password, nil
}

// GET {base}/{id} — metadata only, never the password.
func (c *Client) GetCredential(ctx context.Context, principalType, principalID, envID, id string) (*Credential, error) {
	base, err := credentialBasePath(principalType, principalID, envID)
	if err != nil {
		return nil, err
	}
	resp, err := request[CredentialResponse](ctx, c, http.MethodGet, base+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Credential, nil
}

// DELETE {base}/{id}
func (c *Client) RevokeCredential(ctx context.Context, principalType, principalID, envID, id string) error {
	base, err := credentialBasePath(principalType, principalID, envID)
	if err != nil {
		return err
	}
	return c.doRequest(ctx, http.MethodDelete, base+"/"+id, nil, nil)
}
