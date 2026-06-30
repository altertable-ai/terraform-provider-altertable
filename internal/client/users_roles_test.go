package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserByEmailIssuesLookupAndUnwraps(t *testing.T) {
	var gotMethod, gotRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotRawQuery = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"user":{"id":"u_1","email":"alice@example.com","name":"Alice"}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	u, err := c.GetUserByEmail(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotRawQuery != "email=alice%40example.com" {
		t.Errorf("query = %q, want email=alice%%40example.com", gotRawQuery)
	}
	if u.ID != "u_1" || u.Email != "alice@example.com" || u.Name != "Alice" {
		t.Errorf("user = %+v", u)
	}
}

func TestGetRoleSetUsesUserPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		_, _ = io.WriteString(w, `{"role_assignments":[{"role":"organization:member","resource_id":"org_1"}]}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	set, err := c.GetRoleSet(context.Background(), PrincipalRef{UserID: "u_1"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if gotPath != "GET /users/u_1/role_assignments" {
		t.Errorf("path = %q", gotPath)
	}
	if set.PrincipalID != "u_1" || len(set.Grants) != 1 || set.Grants[0].Role != "organization:member" {
		t.Errorf("set = %+v", set)
	}
}

func TestGetRoleSetUsesServiceAccountPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		_, _ = io.WriteString(w, `{"role_assignments":[]}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	set, err := c.GetRoleSet(context.Background(), PrincipalRef{ServiceAccountID: "sa_1"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if gotPath != "GET /service_accounts/sa_1/role_assignments" {
		t.Errorf("path = %q", gotPath)
	}
	if set.PrincipalID != "sa_1" {
		t.Errorf("principal = %q", set.PrincipalID)
	}
}

func TestPutRoleSetIssuesPutWithRolesBody(t *testing.T) {
	var gotMethod string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		_, _ = io.WriteString(w, `{"role_assignments":[{"role":"catalog:writer","resource_id":"cat_1"}]}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	set, err := c.PutRoleSet(context.Background(), PrincipalRef{UserID: "u_1"}, []RoleGrant{
		{Role: "catalog:writer", ResourceID: "cat_1"},
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	roles, ok := gotBody["roles"].([]interface{})
	if !ok || len(roles) != 1 {
		t.Errorf("body roles = %v", gotBody["roles"])
	}
	if set.Grants[0].Role != "catalog:writer" {
		t.Errorf("grants = %+v", set.Grants)
	}
}

func TestPutRoleSetSerializesNilGrantsAsEmptyArray(t *testing.T) {
	var gotBody map[string]json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		_, _ = io.WriteString(w, `{"role_assignments":[]}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	if _, err := c.PutRoleSet(context.Background(), PrincipalRef{UserID: "u_1"}, nil); err != nil {
		t.Fatalf("err: %s", err)
	}
	// roles must serialize as [] not null so an empty grant set clears assignments.
	rolesRaw, ok := gotBody["roles"]
	if !ok {
		t.Fatal("body missing 'roles' key")
	}
	if string(rolesRaw) != "[]" {
		t.Errorf("roles = %s, want []", rolesRaw)
	}
}
