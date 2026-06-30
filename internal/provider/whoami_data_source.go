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
	_ datasource.DataSource              = (*WhoamiDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*WhoamiDataSource)(nil)
)

func NewWhoamiDataSource() datasource.DataSource {
	return &WhoamiDataSource{}
}

type WhoamiDataSource struct {
	client *client.Client
}

type whoamiDataSourceModel struct {
	PrincipalID      types.String `tfsdk:"principal_id"`
	PrincipalType    types.String `tfsdk:"principal_type"`
	PrincipalName    types.String `tfsdk:"principal_name"`
	PrincipalEmail   types.String `tfsdk:"principal_email"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	OrganizationName types.String `tfsdk:"organization_name"`
	OrganizationSlug types.String `tfsdk:"organization_slug"`
}

func (d *WhoamiDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_whoami"
}

func (d *WhoamiDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The principal and organization the configured API key authenticates as. Use it to scope `organization:*` role grants to your organization without hardcoding its ID (e.g. `resource_id = data.altertable_whoami.current.organization_id`).",
		Attributes: map[string]schema.Attribute{
			"principal_id":      schema.StringAttribute{MarkdownDescription: "ID of the authenticated principal (user or service account).", Computed: true},
			"principal_type":    schema.StringAttribute{MarkdownDescription: "Principal type: `User` or `ServiceAccount`.", Computed: true},
			"principal_name":    schema.StringAttribute{MarkdownDescription: "Principal display name.", Computed: true},
			"principal_email":   schema.StringAttribute{MarkdownDescription: "Principal email (users only; empty for service accounts).", Computed: true},
			"organization_id":   schema.StringAttribute{MarkdownDescription: "Organization ID.", Computed: true},
			"organization_name": schema.StringAttribute{MarkdownDescription: "Organization name.", Computed: true},
			"organization_slug": schema.StringAttribute{MarkdownDescription: "Organization slug.", Computed: true},
		},
	}
}

func (d *WhoamiDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WhoamiDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	who, err := d.client.Whoami(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading whoami", err.Error())
		return
	}
	data := whoamiDataSourceModel{
		PrincipalID:      types.StringValue(who.Principal.ID),
		PrincipalType:    types.StringValue(who.Principal.Type),
		PrincipalName:    types.StringValue(who.Principal.Name),
		PrincipalEmail:   types.StringValue(who.Principal.Email),
		OrganizationID:   types.StringValue(who.Organization.ID),
		OrganizationName: types.StringValue(who.Organization.Name),
		OrganizationSlug: types.StringValue(who.Organization.Slug),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
