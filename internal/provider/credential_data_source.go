package provider

import (
	"context"
	"fmt"

	"github.com/altertable-ai/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*CredentialDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*CredentialDataSource)(nil)
)

func NewCredentialDataSource() datasource.DataSource {
	return &CredentialDataSource{}
}

type CredentialDataSource struct {
	client *client.Client
}

// credentialDataSourceModel deliberately has no password field: the secret is never readable.
type credentialDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	PrincipalType types.String `tfsdk:"principal_type"`
	PrincipalID   types.String `tfsdk:"principal_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Label         types.String `tfsdk:"label"`
	Username      types.String `tfsdk:"username"`
	Default       types.Bool   `tfsdk:"default"`
	Active        types.Bool   `tfsdk:"active"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func (d *CredentialDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (d *CredentialDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up lakehouse credential metadata for a principal in an environment. The password is never returned.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Credential identifier to look up.", Required: true},
			"principal_type": schema.StringAttribute{
				MarkdownDescription: "Principal type: `user` or `service_account`.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf("user", "service_account")},
			},
			"principal_id":   schema.StringAttribute{MarkdownDescription: "ID of the user or service account.", Required: true},
			"environment_id": schema.StringAttribute{MarkdownDescription: "Environment ID.", Required: true},
			"label":          schema.StringAttribute{MarkdownDescription: "Credential label.", Computed: true},
			"username":       schema.StringAttribute{MarkdownDescription: "Generated lakehouse username.", Computed: true},
			"default":        schema.BoolAttribute{MarkdownDescription: "Whether this is the principal's default credential.", Computed: true},
			"active":         schema.BoolAttribute{MarkdownDescription: "Whether the credential is active.", Computed: true},
			"created_at":     schema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
		},
	}
}

func (d *CredentialDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CredentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data credentialDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cred, err := d.client.GetCredential(ctx,
		data.PrincipalType.ValueString(), data.PrincipalID.ValueString(),
		data.EnvironmentID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading credential", err.Error())
		return
	}
	data.Label = types.StringValue(cred.Label)
	data.Username = types.StringValue(cred.Username)
	data.Default = types.BoolValue(cred.Default)
	data.Active = types.BoolValue(cred.Active)
	data.CreatedAt = types.StringValue(cred.CreatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
