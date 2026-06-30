package client

import (
	"context"
	"fmt"
	"net/http"
)

type roleAssignmentsResponse struct {
	RoleAssignments []RoleGrant `json:"role_assignments"`
}

type updateRoleAssignmentsRequest struct {
	Roles []RoleGrant `json:"roles"`
}

func roleAssignmentsPath(p PrincipalRef) (string, error) {
	switch {
	case p.UserID != "":
		return "/users/" + p.UserID + "/role_assignments", nil
	case p.ServiceAccountID != "":
		return "/service_accounts/" + p.ServiceAccountID + "/role_assignments", nil
	default:
		return "", fmt.Errorf("altertable: role set requires user_id or service_account_id")
	}
}

func (p PrincipalRef) id() string {
	if p.UserID != "" {
		return p.UserID
	}
	return p.ServiceAccountID
}

// GET /{users|service_accounts}/{id}/role_assignments
func (c *Client) GetRoleSet(ctx context.Context, p PrincipalRef) (*RoleSet, error) {
	path, err := roleAssignmentsPath(p)
	if err != nil {
		return nil, err
	}
	resp, err := request[roleAssignmentsResponse](ctx, c, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return &RoleSet{PrincipalID: p.id(), Grants: resp.RoleAssignments}, nil
}

// PUT /{users|service_accounts}/{id}/role_assignments — replaces the full set (idempotent).
func (c *Client) PutRoleSet(ctx context.Context, p PrincipalRef, grants []RoleGrant) (*RoleSet, error) {
	path, err := roleAssignmentsPath(p)
	if err != nil {
		return nil, err
	}
	if grants == nil {
		grants = []RoleGrant{}
	}
	resp, err := request[roleAssignmentsResponse](ctx, c, http.MethodPut, path, updateRoleAssignmentsRequest{Roles: grants})
	if err != nil {
		return nil, err
	}
	return &RoleSet{PrincipalID: p.id(), Grants: resp.RoleAssignments}, nil
}
