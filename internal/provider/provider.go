package provider

import (
	"context"
	"os"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	envAPIKey  = "ALTERTABLE_API_KEY"
	envBaseURL = "ALTERTABLE_API_URL"
)

// Ensure AltertableProvider satisfies the provider.Provider interface.
var _ provider.Provider = (*AltertableProvider)(nil)

// AltertableProvider is the provider implementation.
type AltertableProvider struct {
	version string
}

// New returns a provider factory bound to a build version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AltertableProvider{version: version}
	}
}

// AltertableProviderModel maps provider configuration.
type AltertableProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func (p *AltertableProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "altertable"
	resp.Version = p.version
}

func (p *AltertableProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Altertable lakehouse resources.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Altertable management API key. May also be set with the `" + envAPIKey + "` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base URL of the Altertable management API. May also be set with the `" + envBaseURL + "` environment variable. Defaults to `" + client.DefaultBaseURL + "`.",
				Optional:            true,
			},
		},
	}
}

// resolveConfig applies config-over-env-over-default precedence and validates api_key presence.
func resolveConfig(apiKeyCfg, baseURLCfg types.String) (apiKey, baseURL string, diags diag.Diagnostics) {
	apiKey = os.Getenv(envAPIKey)
	if !apiKeyCfg.IsNull() {
		apiKey = apiKeyCfg.ValueString()
	}

	baseURL = os.Getenv(envBaseURL)
	if !baseURLCfg.IsNull() {
		baseURL = baseURLCfg.ValueString()
	}
	if baseURL == "" {
		baseURL = client.DefaultBaseURL
	}

	if apiKey == "" {
		diags.AddAttributeError(
			path.Root("api_key"),
			"Missing Altertable API key",
			"Set the api_key provider argument or the "+envAPIKey+" environment variable.",
		)
	}
	return apiKey, baseURL, diags
}

func (p *AltertableProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config AltertableProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Unknown API key", "api_key cannot be determined at plan time.")
	}
	if config.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("base_url"), "Unknown base URL", "base_url cannot be determined at plan time.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, baseURL, diags := resolveConfig(config.APIKey, config.BaseURL)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := client.NewClient(baseURL, apiKey)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *AltertableProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEnvironmentResource,
	}
}

func (p *AltertableProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
