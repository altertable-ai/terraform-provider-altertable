package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*CredentialResource)(nil)
	_ resource.ResourceWithConfigure   = (*CredentialResource)(nil)
	_ resource.ResourceWithImportState = (*CredentialResource)(nil)
)

func NewCredentialResource() resource.Resource { return &CredentialResource{} }

type CredentialResource struct{ client *client.Client }

type credentialResourceModel struct {
	ID            types.String `tfsdk:"id"`
	PrincipalType types.String `tfsdk:"principal_type"`
	PrincipalID   types.String `tfsdk:"principal_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Label         types.String `tfsdk:"label"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	Default       types.Bool   `tfsdk:"default"`
	Active        types.Bool   `tfsdk:"active"`
	CreatedAt     types.String `tfsdk:"created_at"`
	ExpiresAt     types.String `tfsdk:"expires_at"`
	RevokedAt     types.String `tfsdk:"revoked_at"`
	LastRotatedAt types.String `tfsdk:"last_rotated_at"`
}

func (r *CredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (r *CredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	forceNew := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	useState := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A login credential for a user or service account in an environment. Credentials are immutable; the password is returned only at creation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Credential identifier.", Computed: true, PlanModifiers: useState},
			"principal_type": schema.StringAttribute{
				MarkdownDescription: "Principal type: `user` or `service_account`. Changing this forces a new credential.",
				Required:            true,
				PlanModifiers:       forceNew,
				Validators:          []validator.String{stringvalidator.OneOf("user", "service_account")},
			},
			"principal_id":    schema.StringAttribute{MarkdownDescription: "ID of the user or service account. Changing this forces a new credential.", Required: true, PlanModifiers: forceNew},
			"environment_id":  schema.StringAttribute{MarkdownDescription: "Environment the credential grants access to. Changing this forces a new credential.", Required: true, PlanModifiers: forceNew},
			"label":           schema.StringAttribute{MarkdownDescription: "Human-readable label. Immutable; changing it requires recreating the credential.", Optional: true, Computed: true, PlanModifiers: useState},
			"username":        schema.StringAttribute{MarkdownDescription: "Generated username.", Computed: true, PlanModifiers: useState},
			"password":        schema.StringAttribute{MarkdownDescription: "Generated password. Available only at creation; never re-read from the API.", Computed: true, Sensitive: true, PlanModifiers: useState},
			"default":         schema.BoolAttribute{MarkdownDescription: "Whether this is the principal's default credential.", Computed: true},
			"active":          schema.BoolAttribute{MarkdownDescription: "Whether the credential is active.", Computed: true},
			"created_at":      schema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true, PlanModifiers: useState},
			"expires_at":      schema.StringAttribute{MarkdownDescription: "Expiry timestamp, if any.", Computed: true},
			"revoked_at":      schema.StringAttribute{MarkdownDescription: "Revocation timestamp, if any.", Computed: true},
			"last_rotated_at": schema.StringAttribute{MarkdownDescription: "Last rotation timestamp, if any.", Computed: true},
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
	cred, password, err := r.client.CreateCredential(ctx,
		plan.PrincipalType.ValueString(), plan.PrincipalID.ValueString(), plan.EnvironmentID.ValueString(),
		client.CreateCredentialRequest{Label: plan.Label.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Error creating credential", err.Error())
		return
	}
	applyCredential(&plan, cred)
	plan.Password = types.StringValue(password)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state credentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cred, err := r.client.GetCredential(ctx,
		state.PrincipalType.ValueString(), state.PrincipalID.ValueString(),
		state.EnvironmentID.ValueString(), state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading credential", err.Error())
		return
	}
	// password is write-once and never returned; keep prior state value.
	applyCredential(&state, cred)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is never reached for label-only changes (label uses UseStateForUnknown and
// the API has no update endpoint). It exists to satisfy the interface and to fail
// loudly if Terraform ever routes an in-place change here.
func (r *CredentialResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Credentials are immutable",
		"The Altertable API has no credential update endpoint. Recreate the credential (e.g. taint it) to change it.",
	)
}

func (r *CredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state credentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.RevokeCredential(ctx,
		state.PrincipalType.ValueString(), state.PrincipalID.ValueString(),
		state.EnvironmentID.ValueString(), state.ID.ValueString())
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error revoking credential", err.Error())
	}
}

// ImportState parses "principal_type:principal_id:environment_id:id". The imported
// credential's password will be null (the API never returns it).
func (r *CredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 4 {
		resp.Diagnostics.AddError("Invalid import ID", "expected \"principal_type:principal_id:environment_id:id\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("principal_type"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("principal_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[3])...)
}

func applyCredential(m *credentialResourceModel, cred *client.Credential) {
	m.ID = types.StringValue(cred.ID)
	m.Label = types.StringValue(cred.Label)
	m.Username = types.StringValue(cred.Username)
	m.EnvironmentID = types.StringValue(cred.EnvironmentID)
	m.Default = types.BoolValue(cred.Default)
	m.Active = types.BoolValue(cred.Active)
	m.CreatedAt = types.StringValue(cred.CreatedAt)
	m.ExpiresAt = optString(cred.ExpiresAt)
	m.RevokedAt = optString(cred.RevokedAt)
	m.LastRotatedAt = optString(cred.LastRotatedAt)
}
