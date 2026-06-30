package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateEnvironmentPostsAndDecodes(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"environment":{"id":"env_1","slug":"prod","name":"Production","cloud_provider":"aws","cloud_provider_region":"eu-west-1"}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	env, err := c.CreateEnvironment(context.Background(), CreateEnvironmentRequest{
		Name: "Production", CloudProvider: "aws", CloudProviderAWSRegion: "eu-west-1",
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/environments" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
	if !contains(gotBody, `"cloud_provider_aws_region":"eu-west-1"`) {
		t.Errorf("body = %s", gotBody)
	}
	if env.ID != "env_1" || env.CloudProvider != "aws" {
		t.Errorf("env = %+v", env)
	}
}

func TestGetEnvironmentUsesPathAndDeleteSends(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_, _ = io.WriteString(w, `{"environment":{"id":"env_1","slug":"prod","name":"P"}}`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", "test")
	if _, err := c.GetEnvironment(context.Background(), "prod"); err != nil {
		t.Fatalf("get: %s", err)
	}
	if err := c.DeleteEnvironment(context.Background(), "env_1"); err != nil {
		t.Fatalf("delete: %s", err)
	}
	if paths[0] != "GET /environments/prod" || paths[1] != "DELETE /environments/env_1" {
		t.Errorf("paths = %v", paths)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (indexOf(s, sub) >= 0) }
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
