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
	var gotAuth, gotAccept, gotUserAgent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotUserAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"id":"env_1","slug":"prod","name":"Production"}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret", "test")
	var out Environment
	if err := c.doRequest(context.Background(), http.MethodGet, "/environments/env_1", nil, &out); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if gotAuth != "Bearer secret" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer secret")
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q, want application/json", gotAccept)
	}
	if gotUserAgent != "terraform-provider-altertable/test" {
		t.Errorf("User-Agent = %q", gotUserAgent)
	}
	if out.ID != "env_1" || out.Slug != "prod" || out.Name != "Production" {
		t.Errorf("decoded = %+v", out)
	}
}

func TestDoRequestDecodesWrappedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"no such environment","details":["x"]}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret", "test")
	err := c.doRequest(context.Background(), http.MethodGet, "/environments/nope", nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if apiErr.Code != "not_found" {
		t.Errorf("Code = %q, want not_found", apiErr.Code)
	}
	if apiErr.Message != "no such environment" {
		t.Errorf("Message = %q", apiErr.Message)
	}
}

func TestDoJSONDecodesEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"environment":{"id":"env_1","slug":"prod","name":"Production"}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "secret", "test")
	got, err := doJSON[EnvironmentResponse](context.Background(), c, http.MethodGet, "/environments/env_1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got.Environment.ID != "env_1" {
		t.Errorf("decoded = %+v", got)
	}
}
