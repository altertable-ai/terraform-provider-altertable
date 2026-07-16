package provider

import (
	"context"
	"fmt"

	"github.com/altertable-ai/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*UserResource)(nil)
	_ resource.ResourceWithConfigure   = (*UserResource)(nil)
	_ resource.ResourceWithIdentity    = (*UserResource)(nil)
	_ resource.ResourceWithImportState = (*UserResource)(nil)
)

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *client.Client
}

type userResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
}

// applyUser copies an API user onto the model. It keys the resource on the server's iac_id:
// the API guarantees that value is stable across the invitation → acceptance transition, so
// the same import-once id keeps resolving whether the user is still pending or now a member.
// The API's user_id, invitation_id, and display name are deliberately not surfaced in state —
// nothing in Terraform consumes them, and role sets/credentials key on the iac_id.
func (m *userResourceModel) applyUser(u *client.User) {
	m.ID = types.StringValue(u.IacID)
	m.Email = types.StringValue(u.Email)
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A member of your Altertable organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Opaque identifier for the membership. Remains stable across the invitation → acceptance transition.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address to invite to the organization. Changing it cancels the existing invitation and creates a new one.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	u, err := r.client.CreateUser(ctx, client.CreateUserRequest{Email: plan.Email.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Error inviting user", err.Error())
		return
	}
	plan.applyUser(u)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("id"), plan.ID)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	u, err := r.client.GetUser(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}
	state.applyUser(u)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("id"), state.ID)...)
}

// Update is unreachable: email is the only configurable attribute and it forces replacement,
// so Terraform never plans an in-place update. The resource.Resource interface requires the
// method, so it asserts the invariant loudly — reaching it means a schema change added an
// updatable attribute without wiring up a real Update.
func (r *UserResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	panic("altertable_user.Update called, but the resource has no in-place-updatable attributes (email forces replacement): this is a provider bug")
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteUser(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error removing user", err.Error())
	}
}

type userIdentityModel struct {
	ID     types.String `tfsdk:"id"`
	UserID types.String `tfsdk:"user_id"`
}

// The persisted identity is always the stable iac_id in `id`. `user_id` exists only as an
// alternate import key: it is the raw user UUID (present once a member has accepted), which
// the provider resolves to the iac_id at import time. Both are OptionalForImport because a
// caller supplies exactly one.
func (r *UserResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id":      identityschema.StringAttribute{OptionalForImport: true},
			"user_id": identityschema.StringAttribute{OptionalForImport: true},
		},
	}
}

// iacIDForImport turns an import identity into the stable iac_id to key the resource on.
// Exactly one of id (the iac_id itself) or user_id (a raw user UUID, resolved via the API)
// must be set.
func (r *UserResource) iacIDForImport(ctx context.Context, ident userIdentityModel) (string, error) {
	hasID, hasUser := !ident.ID.IsNull(), !ident.UserID.IsNull()
	switch {
	case hasID && hasUser:
		return "", fmt.Errorf("set only one of id or user_id, not both")
	case hasID:
		return ident.ID.ValueString(), nil
	case hasUser:
		u, err := r.client.GetUser(ctx, ident.UserID.ValueString())
		if err != nil {
			return "", err
		}
		return u.IacID, nil
	default:
		return "", fmt.Errorf("set one of id or user_id")
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		// Identity-only by design: a user is imported through an identity block, never the legacy
		// string form. That keeps the alternate user_id key unambiguous and leaves one flow to learn.
		resp.Diagnostics.AddError(
			"String import not supported",
			"Import altertable_user with an identity block (Terraform 1.12+): "+
				"`identity = { id = <id> }`, or `identity = { user_id = <user UUID> }` "+
				"to resolve an accepted member's UUID to their id.",
		)
		return
	}
	var ident userIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &ident)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Resolve to the iac_id here, not in Read: the persisted identity must be canonical from
	// the first moment, or it would change during the post-import refresh — which Terraform rejects.
	iacID, err := r.iacIDForImport(ctx, ident)
	if err != nil {
		resp.Diagnostics.AddError("Invalid user import identity", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), iacID)...)
	resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("id"), iacID)...)
}
