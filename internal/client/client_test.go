package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoRequestSuccessSetsAuthAndDecodes(t *testing.T) {
	var gotAuth, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"id":"env_1","slug":"prod","name":"Production"}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret")
	var out Environment
	err := c.doRequest(context.Background(), http.MethodGet, "/v1/environments/env_1", nil, &out)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if gotAuth != "Bearer secret" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer secret")
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q, want application/json", gotAccept)
	}
	if out.ID != "env_1" || out.Slug != "prod" || out.Name != "Production" {
		t.Errorf("decoded = %+v", out)
	}
}

func TestDoRequestNon2xxReturnsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"message":"not found"}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret")
	err := c.doRequest(context.Background(), http.MethodGet, "/v1/environments/nope", nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func TestStubReturnsErrNotImplemented(t *testing.T) {
	c := NewClient(DefaultBaseURL, "secret")
	_, err := c.GetEnvironment(context.Background(), "env_1")
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("error = %v, want ErrNotImplemented", err)
	}
}
