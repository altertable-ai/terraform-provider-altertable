package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateUserPostsEmailAndUnwraps(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"user":{"id":null,"invitation_id":"inv_1","iac_id":"invitation_id=inv_1","email":"bob@example.com","name":null}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	u, err := c.CreateUser(context.Background(), CreateUserRequest{Email: "bob@example.com"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/users" {
		t.Errorf("request = %s %s, want POST /users", gotMethod, gotPath)
	}
	if gotBody["email"] != "bob@example.com" {
		t.Errorf("body email = %q", gotBody["email"])
	}
	// A pending invitation: iac_id and invitation_id are set; the user UUID and name are null.
	if u.IacID != "invitation_id=inv_1" || u.InvitationID != "inv_1" || u.Email != "bob@example.com" {
		t.Errorf("user = %+v", u)
	}
	if u.ID != "" || u.Name != "" {
		t.Errorf("null fields decoded to non-empty: id=%q name=%q", u.ID, u.Name)
	}
}

func TestGetUserResolvesByIacID(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = io.WriteString(w, `{"user":{"id":"u_1","invitation_id":null,"iac_id":"invitation_id=inv_1","email":"carol@example.com","name":"Carol"}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	// The iac_id keeps its invitation_id=<uuid> form even after acceptance; it must reach the
	// server verbatim (the "=" is a literal path character, not a query separator).
	u, err := c.GetUser(context.Background(), "invitation_id=inv_1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/users/invitation_id=inv_1" {
		t.Errorf("request = %s %s, want GET /users/invitation_id=inv_1", gotMethod, gotPath)
	}
	if u.ID != "u_1" || u.InvitationID != "" || u.Name != "Carol" {
		t.Errorf("user = %+v", u)
	}
}

func TestDeleteUserIssuesDelete(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	if err := c.DeleteUser(context.Background(), "u_1"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/users/u_1" {
		t.Errorf("request = %s %s, want DELETE /users/u_1", gotMethod, gotPath)
	}
}
