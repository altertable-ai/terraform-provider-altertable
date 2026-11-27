package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateRoleBackedBucket(t *testing.T) {
	const roleARN = "arn:aws:iam::123456789012:role/altertable"
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %s", err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"bucket":{"id":"bkt_1","name":"landing","slug":"BKT-1","provider":"s3","region":"eu-west-1","role_arn":"`+roleARN+`"}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	bucket, err := c.CreateBucket(context.Background(), "env_1", CreateBucketRequest{Name: "landing", RoleARN: roleARN, Region: "eu-west-1"})
	if err != nil {
		t.Fatalf("create: %s", err)
	}
	if bucket.RoleARN != roleARN {
		t.Errorf("RoleARN = %q, want %q", bucket.RoleARN, roleARN)
	}
	if body["role_arn"] != roleARN {
		t.Errorf("request role_arn = %v, want %q", body["role_arn"], roleARN)
	}
	for _, k := range []string{"access_key_id", "secret_access_key"} {
		if _, ok := body[k]; ok {
			t.Errorf("request must omit %s for a role-backed bucket, got %v", k, body[k])
		}
	}
}

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
