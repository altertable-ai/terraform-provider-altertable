package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWhoamiDecodesPrincipalAndOrg(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/whoami" {
			t.Errorf("path = %q, want /whoami", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"principal":{"id":"p_1","type":"ServiceAccount","name":"CI Deploy","slug":"ci-deploy"},"organization":{"id":"o_1","name":"Acme Corp","slug":"acme-corp"}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret", "test")
	got, err := c.Whoami(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got.Principal.ID != "p_1" || got.Principal.Type != "ServiceAccount" {
		t.Errorf("principal = %+v", got.Principal)
	}
	if got.Organization.Slug != "acme-corp" {
		t.Errorf("organization = %+v", got.Organization)
	}
}

func TestWhoamiReturnsAPIErrorOn401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"code":"unauthorized","message":"Invalid or expired token"}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "bad", "test")
	_, err := c.Whoami(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
}
