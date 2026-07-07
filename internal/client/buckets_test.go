package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBucketCRUDPaths(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"bucket":{"id":"bkt_1","name":"Landing","slug":"BKT-1","provider":"s3"}}`)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			_, _ = io.WriteString(w, `{"bucket":{"id":"bkt_1","name":"Landing","slug":"BKT-1","provider":"s3"}}`)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	if _, err := c.CreateBucket(context.Background(), "env_1", CreateBucketRequest{Name: "Landing", AccessKeyID: "ak", SecretAccessKey: "sk"}); err != nil {
		t.Fatalf("create: %s", err)
	}
	if _, err := c.GetBucket(context.Background(), "env_1", "BKT-1"); err != nil {
		t.Fatalf("get: %s", err)
	}
	if _, err := c.UpdateBucket(context.Background(), "env_1", "BKT-1", UpdateBucketRequest{Name: "Landing2"}); err != nil {
		t.Fatalf("update: %s", err)
	}
	if err := c.DeleteBucket(context.Background(), "env_1", "BKT-1"); err != nil {
		t.Fatalf("delete: %s", err)
	}
	want := []string{
		"POST /environments/env_1/buckets",
		"GET /environments/env_1/buckets/BKT-1",
		"PATCH /environments/env_1/buckets/BKT-1",
		"DELETE /environments/env_1/buckets/BKT-1",
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
}
