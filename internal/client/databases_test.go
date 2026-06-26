package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDatabaseCRUDPaths(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"database":{"id":"db_1","name":"Main","slug":"main"}}`)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			_, _ = io.WriteString(w, `{"database":{"id":"db_1","name":"Main","slug":"main"}}`)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	if _, err := c.CreateDatabase(context.Background(), "env_1", CreateDatabaseRequest{Name: "Main"}); err != nil {
		t.Fatalf("create: %s", err)
	}
	if _, err := c.GetDatabase(context.Background(), "env_1", "main"); err != nil {
		t.Fatalf("get: %s", err)
	}
	if _, err := c.UpdateDatabase(context.Background(), "env_1", "db_1", UpdateDatabaseRequest{Name: "Main2"}); err != nil {
		t.Fatalf("update: %s", err)
	}
	if err := c.DeleteDatabase(context.Background(), "env_1", "db_1"); err != nil {
		t.Fatalf("delete: %s", err)
	}
	want := []string{
		"POST /environments/env_1/databases",
		"GET /environments/env_1/databases/main",
		"PATCH /environments/env_1/databases/db_1",
		"DELETE /environments/env_1/databases/db_1",
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
}
