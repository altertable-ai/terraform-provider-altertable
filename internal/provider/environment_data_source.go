package provider

import (
	"context"
	"fmt"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*EnvironmentDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*EnvironmentDataSource)(nil)
)

func NewEnvironmentDataSource() datasource.DataSource {
	return &EnvironmentDataSource{}
}

type EnvironmentDataSource struct {
	client *client.Client
}

type environmentDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Slug                types.String `tfsdk:"slug"`
	Name                types.String `tfsdk:"name"`
	CloudProvider       types.String `tfsdk:"cloud_provider"`
	CloudProviderRegion types.String `tfsdk:"cloud_provider_region"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func (d *EnvironmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *EnvironmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up an Altertable environment by slug (or ID).",
		Attributes: map[string]schema.Attribute{
			"id":                    schema.StringAttribute{MarkdownDescription: "Environment identifier.", Computed: true},
			"slug":                  schema.StringAttribute{MarkdownDescription: "Environment slug (or UUID) to look up.", Required: true},
			"name":                  schema.StringAttribute{MarkdownDescription: "Human-readable environment name.", Computed: true},
			"cloud_provider":        schema.StringAttribute{MarkdownDescription: "Cloud provider.", Computed: true},
			"cloud_provider_region": schema.StringAttribute{MarkdownDescription: "Cloud provider region.", Computed: true},
			"created_at":            schema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
			"updated_at":            schema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true},
		},
	}
}

func (d *EnvironmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *EnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data environmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	env, err := d.client.GetEnvironment(ctx, data.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading environment", err.Error())
		return
	}
	data.ID = types.StringValue(env.ID)
	data.Slug = types.StringValue(env.Slug)
	data.Name = types.StringValue(env.Name)
	data.CloudProvider = types.StringValue(env.CloudProvider)
	data.CloudProviderRegion = types.StringValue(env.CloudProviderRegion)
	data.CreatedAt = types.StringValue(env.CreatedAt)
	data.UpdatedAt = types.StringValue(env.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
