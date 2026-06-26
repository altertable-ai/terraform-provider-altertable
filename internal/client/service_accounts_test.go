package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServiceAccountCRUDPaths(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"service_account":{"id":"sa_1","label":"CI","slug":"ci"}}`)
		case http.MethodGet, http.MethodPatch:
			_, _ = io.WriteString(w, `{"service_account":{"id":"sa_1","label":"CI","slug":"ci"}}`)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	if _, err := c.CreateServiceAccount(context.Background(), CreateServiceAccountRequest{Label: "CI"}); err != nil {
		t.Fatalf("create: %s", err)
	}
	if _, err := c.GetServiceAccount(context.Background(), "sa_1"); err != nil {
		t.Fatalf("get: %s", err)
	}
	if _, err := c.UpdateServiceAccount(context.Background(), "sa_1", UpdateServiceAccountRequest{Label: "CI2"}); err != nil {
		t.Fatalf("update: %s", err)
	}
	if err := c.DeleteServiceAccount(context.Background(), "sa_1"); err != nil {
		t.Fatalf("delete: %s", err)
	}
	want := []string{"POST /service_accounts", "GET /service_accounts/sa_1", "PATCH /service_accounts/sa_1", "DELETE /service_accounts/sa_1"}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
}
