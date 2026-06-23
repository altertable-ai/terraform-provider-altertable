package provider

import (
	"context"
	"os"
	"testing"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"altertable": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("ALTERTABLE_API_KEY") == "" {
		t.Fatal("ALTERTABLE_API_KEY must be set for acceptance tests")
	}
	apiURL := os.Getenv("ALTERTABLE_API_URL")
	if apiURL == "" {
		t.Fatal("ALTERTABLE_API_URL must be set for acceptance tests; leaving it unset would target the production default")
	}
	if apiURL == client.DefaultBaseURL {
		t.Fatalf("ALTERTABLE_API_URL must not be the production endpoint %s; point acceptance tests at a non-production instance", client.DefaultBaseURL)
	}
}

func TestProviderSchemaValid(t *testing.T) {
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatalf("create server: %s", err)
	}
	resp, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("GetProviderSchema: %s", err)
	}
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Errorf("schema diagnostic: %s: %s", d.Summary, d.Detail)
		}
	}
}

func TestResolveConfig(t *testing.T) {
	t.Setenv("ALTERTABLE_API_KEY", "")
	t.Setenv("ALTERTABLE_API_URL", "")

	// Config values win over env.
	apiKey, baseURL, diags := resolveConfig(types.StringValue("cfgkey"), types.StringValue("https://cfg"))
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if apiKey != "cfgkey" || baseURL != "https://cfg" {
		t.Fatalf("got apiKey=%q baseURL=%q", apiKey, baseURL)
	}

	// Env fallback + default base URL.
	t.Setenv("ALTERTABLE_API_KEY", "envkey")
	apiKey, baseURL, diags = resolveConfig(types.StringNull(), types.StringNull())
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if apiKey != "envkey" || baseURL != client.DefaultBaseURL {
		t.Fatalf("got apiKey=%q baseURL=%q", apiKey, baseURL)
	}

	// Missing api_key is an error.
	t.Setenv("ALTERTABLE_API_KEY", "")
	if _, _, diags = resolveConfig(types.StringNull(), types.StringNull()); !diags.HasError() {
		t.Fatal("expected error diagnostic for missing api_key")
	}
}
