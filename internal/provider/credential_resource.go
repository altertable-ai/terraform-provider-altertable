package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
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
	_ resource.ResourceWithIdentity    = (*CredentialResource)(nil)
)

type credentialIdentityModel struct {
	PrincipalType types.String `tfsdk:"principal_type"`
	PrincipalID   types.String `tfsdk:"principal_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ID            types.String `tfsdk:"id"`
}

func credentialIdentity(m *credentialResourceModel) credentialIdentityModel {
	return credentialIdentityModel{
		PrincipalType: m.PrincipalType,
		PrincipalID:   m.PrincipalID,
		EnvironmentID: m.EnvironmentID,
		ID:            m.ID,
	}
}

func (r *CredentialResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"principal_type": identityschema.StringAttribute{RequiredForImport: true},
			"principal_id":   identityschema.StringAttribute{RequiredForImport: true},
			"environment_id": identityschema.StringAttribute{RequiredForImport: true},
			"id":             identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

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
	// label is user-settable, but the API has no credential update endpoint, so changing
	// it forces replacement; UseStateForUnknown keeps the server-assigned value stable.
	labelMods := []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A lakehouse credential (username/password) a user or service account uses to query the catalogs in an environment. Credentials are immutable; the password is returned only at creation.",
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
			"label":           schema.StringAttribute{MarkdownDescription: "Human-readable label. The API has no credential update endpoint, so changing it forces replacement — which mints a new password and revokes the old one.", Optional: true, Computed: true, PlanModifiers: labelMods},
			"username":        schema.StringAttribute{MarkdownDescription: "Generated lakehouse username.", Computed: true, PlanModifiers: useState},
			"password":        schema.StringAttribute{MarkdownDescription: "Generated lakehouse password. Available only at creation; never re-read from the API.", Computed: true, Sensitive: true, PlanModifiers: useState},
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
	resp.Diagnostics.Append(resp.Identity.Set(ctx, credentialIdentity(&plan))...)
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
	// The API returns every field except the password (write-once at creation), so
	// applyCredential refreshes all of them and leaves the prior password value in place.
	applyCredential(&state, cred)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, credentialIdentity(&state))...)
}

// Update is unreachable: every settable attribute (principal_type, principal_id,
// environment_id, label) forces replacement, so Terraform never plans an in-place
// update. It hard-errors as a defensive backstop and to satisfy the interface.
func (r *CredentialResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Credentials are immutable",
		"The Altertable API has no credential update endpoint. Change a credential by recreating it.",
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

// parseCredentialImportID splits the back-compat
// "principal_type:principal_id:environment_id:id" import string. All four
// segments must be non-empty.
func parseCredentialImportID(importID string) (principalType, principalID, env, id string, ok bool) {
	parts := strings.Split(importID, ":")
	if len(parts) != 4 {
		return "", "", "", "", false
	}
	for _, p := range parts {
		if p == "" {
			return "", "", "", "", false
		}
	}
	return parts[0], parts[1], parts[2], parts[3], true
}

func (r *CredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		// Back-compat: "principal_type:principal_id:environment_id:id" colon string.
		pt, pid, env, id, ok := parseCredentialImportID(req.ID)
		if !ok {
			resp.Diagnostics.AddError("Invalid import ID", "expected \"principal_type:principal_id:environment_id:id\"")
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("principal_type"), pt)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("principal_id"), pid)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), env)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
		return
	}
	var ident credentialIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &ident)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("principal_type"), ident.PrincipalType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("principal_id"), ident.PrincipalID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), ident.EnvironmentID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), ident.ID)...)
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
