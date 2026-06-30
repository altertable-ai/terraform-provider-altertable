package provider

import (
	"context"
	"fmt"

	"github.com/altertable-ai/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*ServiceAccountDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ServiceAccountDataSource)(nil)
)

func NewServiceAccountDataSource() datasource.DataSource {
	return &ServiceAccountDataSource{}
}

type ServiceAccountDataSource struct {
	client *client.Client
}

type serviceAccountDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Label types.String `tfsdk:"label"`
	Slug  types.String `tfsdk:"slug"`
}

func (d *ServiceAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *ServiceAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a service account by ID.",
		Attributes: map[string]schema.Attribute{
			"id":    schema.StringAttribute{MarkdownDescription: "Service account ID to look up.", Required: true},
			"label": schema.StringAttribute{MarkdownDescription: "Service account label.", Computed: true},
			"slug":  schema.StringAttribute{MarkdownDescription: "Service account slug.", Computed: true},
		},
	}
}

func (d *ServiceAccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serviceAccountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sa, err := d.client.GetServiceAccount(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading service account", err.Error())
		return
	}
	data.Label = types.StringValue(sa.Label)
	data.Slug = types.StringValue(sa.Slug)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
