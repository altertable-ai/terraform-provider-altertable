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
	_ datasource.DataSource              = (*CatalogDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*CatalogDataSource)(nil)
)

func NewCatalogDataSource() datasource.DataSource {
	return &CatalogDataSource{}
}

type CatalogDataSource struct {
	client *client.Client
}

type catalogDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Slug          types.String `tfsdk:"slug"`
	Name          types.String `tfsdk:"name"`
}

func (d *CatalogDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalog"
}

func (d *CatalogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a catalog by environment and slug.",
		Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{MarkdownDescription: "Catalog identifier.", Computed: true},
			"environment_id": schema.StringAttribute{MarkdownDescription: "Parent environment ID.", Required: true},
			"slug":           schema.StringAttribute{MarkdownDescription: "Catalog slug to look up.", Required: true},
			"name":           schema.StringAttribute{MarkdownDescription: "Human-readable catalog name.", Computed: true},
		},
	}
}

func (d *CatalogDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CatalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data catalogDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cat, err := d.client.GetCatalogBySlug(ctx, data.EnvironmentID.ValueString(), data.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading catalog", err.Error())
		return
	}
	data.ID = types.StringValue(cat.ID)
	data.Name = types.StringValue(cat.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
