package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCredentialUserPathAndPassword(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"credential":{"id":"cred_1","label":"l","username":"u","environment_id":"env_1"},"password":"secret-once"}`)
		case http.MethodGet:
			_, _ = io.WriteString(w, `{"credential":{"id":"cred_1","label":"l","username":"u","environment_id":"env_1"}}`)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	cred, pw, err := c.CreateCredential(context.Background(), "user", "usr_1", "env_1", CreateCredentialRequest{Label: "l"})
	if err != nil {
		t.Fatalf("create: %s", err)
	}
	if pw != "secret-once" || cred.ID != "cred_1" {
		t.Errorf("cred=%+v pw=%q", cred, pw)
	}
	if _, err := c.GetCredential(context.Background(), "user", "usr_1", "env_1", "cred_1"); err != nil {
		t.Fatalf("get: %s", err)
	}
	if err := c.RevokeCredential(context.Background(), "user", "usr_1", "env_1", "cred_1"); err != nil {
		t.Fatalf("revoke: %s", err)
	}
	want := []string{
		"POST /users/usr_1/environments/env_1/credentials",
		"GET /users/usr_1/environments/env_1/credentials/cred_1",
		"DELETE /users/usr_1/environments/env_1/credentials/cred_1",
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
}

func TestCredentialServiceAccountPath(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"credential":{"id":"cred_1"},"password":"p"}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	if _, _, err := c.CreateCredential(context.Background(), "service_account", "sa_1", "env_1", CreateCredentialRequest{}); err != nil {
		t.Fatalf("create: %s", err)
	}
	if path != "/service_accounts/sa_1/environments/env_1/credentials" {
		t.Errorf("path = %q", path)
	}
}

func TestCredentialUnknownPrincipalErrors(t *testing.T) {
	c := NewClient(DefaultBaseURL, "k", "test")
	_, _, err := c.CreateCredential(context.Background(), "bogus", "x", "env_1", CreateCredentialRequest{})
	if err == nil || errors.Is(err, ErrNotImplemented) {
		t.Fatalf("want a validation error, got %v", err)
	}
}
