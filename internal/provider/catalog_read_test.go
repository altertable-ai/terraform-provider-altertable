package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// catalogProbeServer routes the database and connection GETs to fixed status codes so
// readCatalog's empty-engine probing can be exercised without a live API.
func catalogProbeServer(t *testing.T, dbExists, conExists bool) *client.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/databases/"):
			if dbExists {
				_, _ = io.WriteString(w, `{"database":{"id":"cat_1","name":"Main","slug":"main"}}`)
				return
			}
		case strings.Contains(r.URL.Path, "/connections/"):
			if conExists {
				_, _ = io.WriteString(w, `{"connection":{"id":"cat_1","name":"PG","slug":"pg","engine":"postgres"}}`)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found"}}`)
	}))
	t.Cleanup(srv.Close)
	return client.NewClient(srv.URL, "k", "test")
}

func emptyEngineState() *catalogResourceModel {
	return &catalogResourceModel{
		EnvironmentID: types.StringValue("env_1"),
		ID:            types.StringValue("cat_1"),
		Engine:        types.StringValue(""),
	}
}

func TestReadCatalogEmptyEngineErrorsWhenAmbiguous(t *testing.T) {
	r := &CatalogResource{client: catalogProbeServer(t, true, true)}
	found, err := r.readCatalog(context.Background(), emptyEngineState())
	if err == nil {
		t.Fatalf("expected an ambiguity error, got found=%v err=nil", found)
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("error = %q, want it to mention ambiguity", err)
	}
}

func TestReadCatalogEmptyEngineResolvesDatabase(t *testing.T) {
	r := &CatalogResource{client: catalogProbeServer(t, true, false)}
	state := emptyEngineState()
	found, err := r.readCatalog(context.Background(), state)
	if err != nil || !found {
		t.Fatalf("found=%v err=%v, want true/nil", found, err)
	}
	if state.Engine.ValueString() != engineAltertable {
		t.Errorf("engine = %q, want %q", state.Engine.ValueString(), engineAltertable)
	}
}

func TestReadCatalogEmptyEngineResolvesConnection(t *testing.T) {
	r := &CatalogResource{client: catalogProbeServer(t, false, true)}
	state := emptyEngineState()
	found, err := r.readCatalog(context.Background(), state)
	if err != nil || !found {
		t.Fatalf("found=%v err=%v, want true/nil", found, err)
	}
	if state.Engine.ValueString() != "postgres" {
		t.Errorf("engine = %q, want postgres", state.Engine.ValueString())
	}
}

func TestReadCatalogEmptyEngineReturnsGoneWhenNeitherExists(t *testing.T) {
	r := &CatalogResource{client: catalogProbeServer(t, false, false)}
	found, err := r.readCatalog(context.Background(), emptyEngineState())
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if found {
		t.Error("found = true, want false (dropped from state)")
	}
}
