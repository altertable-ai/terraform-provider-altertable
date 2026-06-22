package provider

import (
	"context"
	"fmt"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*CredentialResource)(nil)
	_ resource.ResourceWithConfigure   = (*CredentialResource)(nil)
	_ resource.ResourceWithImportState = (*CredentialResource)(nil)
)

func NewCredentialResource() resource.Resource {
	return &CredentialResource{}
}

type CredentialResource struct {
	client *client.Client
}

type credentialResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	EnvironmentID    types.String `tfsdk:"environment_id"`
	Label            types.String `tfsdk:"label"`
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
}

func (r *CredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (r *CredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A credential for a service account in an environment. The password is returned only on creation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Credential identifier.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"service_account_id": schema.StringAttribute{
				MarkdownDescription: "Service account the credential belongs to. Changing this forces a new credential.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Environment the credential grants access to. Changing this forces a new credential.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Human-readable label for the credential.",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Generated username.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Generated password. Only available at creation time; never re-read from the API.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *CredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *CredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan credentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cred, err := r.client.CreateCredential(ctx, client.CredentialCreateInput{
		ServiceAccountID: plan.ServiceAccountID.ValueString(),
		EnvironmentID:    plan.EnvironmentID.ValueString(),
		Label:            plan.Label.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating credential", err.Error())
		return
	}
	plan.ID = types.StringValue(cred.ID)
	plan.Label = types.StringValue(cred.Label)
	plan.Username = types.StringValue(cred.Username)
	plan.Password = types.StringValue(cred.Password)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state credentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cred, err := r.client.GetCredential(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading credential", err.Error())
		return
	}
	// password is write-once and never returned by the API; the existing state value is preserved.
	state.ServiceAccountID = types.StringValue(cred.ServiceAccountID)
	state.EnvironmentID = types.StringValue(cred.EnvironmentID)
	state.Label = types.StringValue(cred.Label)
	state.Username = types.StringValue(cred.Username)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan credentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cred, err := r.client.UpdateCredential(ctx, plan.ID.ValueString(), client.CredentialUpdateInput{
		Label: plan.Label.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating credential", err.Error())
		return
	}
	plan.ServiceAccountID = types.StringValue(cred.ServiceAccountID)
	plan.EnvironmentID = types.StringValue(cred.EnvironmentID)
	plan.Label = types.StringValue(cred.Label)
	plan.Username = types.StringValue(cred.Username)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state credentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCredential(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error deleting credential", err.Error())
	}
}

func (r *CredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Note: an imported credential will have a null password, since the API never returns it.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
