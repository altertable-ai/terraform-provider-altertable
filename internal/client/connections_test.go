package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConnectionCRUDPaths(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"connection":{"id":"con_1","name":"PG","slug":"pg","engine":"postgres"}}`)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			_, _ = io.WriteString(w, `{"connection":{"id":"con_1","name":"PG","slug":"pg","engine":"postgres"}}`)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	if _, err := c.CreateConnection(context.Background(), "env_1", CreateConnectionRequest{Name: "PG", Engine: "postgres"}); err != nil {
		t.Fatalf("create: %s", err)
	}
	if _, err := c.GetConnection(context.Background(), "env_1", "pg"); err != nil {
		t.Fatalf("get: %s", err)
	}
	if _, err := c.UpdateConnection(context.Background(), "env_1", "con_1", UpdateConnectionRequest{Name: "PG2"}); err != nil {
		t.Fatalf("update: %s", err)
	}
	if err := c.DeleteConnection(context.Background(), "env_1", "con_1"); err != nil {
		t.Fatalf("delete: %s", err)
	}
	want := []string{
		"POST /environments/env_1/connections",
		"GET /environments/env_1/connections/pg",
		"PATCH /environments/env_1/connections/con_1",
		"DELETE /environments/env_1/connections/con_1",
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
}
