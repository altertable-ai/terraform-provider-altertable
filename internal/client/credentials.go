package client

import (
	"context"
	"fmt"
	"net/http"
)

// credentialBasePath builds the principal-scoped credentials path.
// principalType is "user" or "service_account".
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

// CreateCredential: POST {base} — returns the credential and the one-time password.
func (c *Client) CreateCredential(ctx context.Context, principalType, principalID, envID string, in CreateCredentialRequest) (*Credential, string, error) {
	base, err := credentialBasePath(principalType, principalID, envID)
	if err != nil {
		return nil, "", err
	}
	resp, err := doJSON[CreateCredentialResponse](ctx, c, http.MethodPost, base, in)
	if err != nil {
		return nil, "", err
	}
	return &resp.Credential, resp.Password, nil
}

// GetCredential: GET {base}/{id} — metadata only, never the password.
func (c *Client) GetCredential(ctx context.Context, principalType, principalID, envID, id string) (*Credential, error) {
	base, err := credentialBasePath(principalType, principalID, envID)
	if err != nil {
		return nil, err
	}
	resp, err := doJSON[CredentialResponse](ctx, c, http.MethodGet, base+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Credential, nil
}

// RevokeCredential: DELETE {base}/{id}
func (c *Client) RevokeCredential(ctx context.Context, principalType, principalID, envID, id string) error {
	base, err := credentialBasePath(principalType, principalID, envID)
	if err != nil {
		return err
	}
	return c.doRequest(ctx, http.MethodDelete, base+"/"+id, nil, nil)
}
