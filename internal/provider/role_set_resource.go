package provider

import (
	"context"
	"fmt"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = (*RoleSetResource)(nil)
	_ resource.ResourceWithConfigure        = (*RoleSetResource)(nil)
	_ resource.ResourceWithConfigValidators = (*RoleSetResource)(nil)
)

func NewRoleSetResource() resource.Resource {
	return &RoleSetResource{}
}

type RoleSetResource struct {
	client *client.Client
}

type roleSetResourceModel struct {
	ID               types.String     `tfsdk:"id"`
	UserID           types.String     `tfsdk:"user_id"`
	ServiceAccountID types.String     `tfsdk:"service_account_id"`
	Roles            []roleGrantModel `tfsdk:"roles"`
}

type roleGrantModel struct {
	Role       types.String `tfsdk:"role"`
	ResourceID types.String `tfsdk:"resource_id"`
}

func (m roleSetResourceModel) principalRef() client.PrincipalRef {
	return client.PrincipalRef{
		UserID:           m.UserID.ValueString(),
		ServiceAccountID: m.ServiceAccountID.ValueString(),
	}
}

func toClientGrants(in []roleGrantModel) []client.RoleGrant {
	out := make([]client.RoleGrant, 0, len(in))
	for _, g := range in {
		out = append(out, client.RoleGrant{
			Role:       g.Role.ValueString(),
			ResourceID: g.ResourceID.ValueString(),
		})
	}
	return out
}

func fromClientGrants(in []client.RoleGrant) []roleGrantModel {
	out := make([]roleGrantModel, 0, len(in))
	for _, g := range in {
		m := roleGrantModel{Role: types.StringValue(g.Role)}
		if g.ResourceID == "" {
			m.ResourceID = types.StringNull()
		} else {
			m.ResourceID = types.StringValue(g.ResourceID)
		}
		out = append(out, m)
	}
	return out
}

func (r *RoleSetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_set"
}

func (r *RoleSetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The complete set of role grants for a single principal (user or service account).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the principal these roles belong to.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "User this role set applies to. Exactly one of `user_id` or `service_account_id` must be set. Changing this forces a new role set.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"service_account_id": schema.StringAttribute{
				MarkdownDescription: "Service account this role set applies to. Exactly one of `user_id` or `service_account_id` must be set. Changing this forces a new role set.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"roles": schema.SetNestedAttribute{
				MarkdownDescription: "The full set of role grants for the principal.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.StringAttribute{
							MarkdownDescription: "Role identifier, e.g. `organization:member` or `environment:writer`.",
							Required:            true,
						},
						"resource_id": schema.StringAttribute{
							MarkdownDescription: "Resource the role is scoped to. Omit for organization-wide roles.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *RoleSetResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("user_id"),
			path.MatchRoot("service_account_id"),
		),
	}
}

func (r *RoleSetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RoleSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	set, err := r.client.PutRoleSet(ctx, plan.principalRef(), toClientGrants(plan.Roles))
	if err != nil {
		resp.Diagnostics.AddError("Error creating role set", err.Error())
		return
	}
	plan.ID = types.StringValue(set.PrincipalID)
	plan.Roles = fromClientGrants(set.Grants)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *RoleSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	set, err := r.client.GetRoleSet(ctx, state.principalRef())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading role set", err.Error())
		return
	}
	state.Roles = fromClientGrants(set.Grants)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *RoleSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan roleSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	set, err := r.client.PutRoleSet(ctx, plan.principalRef(), toClientGrants(plan.Roles))
	if err != nil {
		resp.Diagnostics.AddError("Error updating role set", err.Error())
		return
	}
	plan.ID = types.StringValue(set.PrincipalID)
	plan.Roles = fromClientGrants(set.Grants)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete always errors: the API has no endpoint to delete role assignments, and a
// principal in an organization always has at least its membership role — there is no
// "no roles" state. Removing a principal's access means removing the principal from the
// organization, not emptying this role set.
func (r *RoleSetResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Role sets cannot be deleted",
		"The Altertable API has no endpoint to delete role assignments, and a principal always "+
			"retains its organization membership role — there is no \"no roles\" state. To remove a "+
			"principal's access, remove the principal from the organization (delete the user or "+
			"service account). To stop managing this role set in Terraform without changing the "+
			"server, remove it from state with `terraform state rm`.",
	)
}
