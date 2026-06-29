package provider

import (
	"context"
	"fmt"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*EnvironmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*EnvironmentResource)(nil)
	_ resource.ResourceWithIdentity    = (*EnvironmentResource)(nil)
	_ resource.ResourceWithImportState = (*EnvironmentResource)(nil)
)

func NewEnvironmentResource() resource.Resource { return &EnvironmentResource{} }

type EnvironmentResource struct{ client *client.Client }

type environmentResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	CloudProvider       types.String `tfsdk:"cloud_provider"`
	CloudProviderRegion types.String `tfsdk:"cloud_provider_region"`
	Slug                types.String `tfsdk:"slug"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func (r *EnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	forceNewStr := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	computedStr := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "An Altertable environment. Environments are immutable: every attribute change forces a new environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Environment identifier.",
				Computed:            true,
				PlanModifiers:       computedStr,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable environment name. Changing this forces a new environment.",
				Required:            true,
				PlanModifiers:       forceNewStr,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "Cloud provider: `hetzner` or `aws`. Changing this forces a new environment.",
				Required:            true,
				PlanModifiers:       forceNewStr,
			},
			"cloud_provider_region": schema.StringAttribute{
				MarkdownDescription: "Region within the cloud provider (e.g. `fsn1` for hetzner, `eu-west-1` for aws). Changing this forces a new environment.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       forceNewStr,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-safe environment slug (server-assigned).",
				Computed:            true,
				PlanModifiers:       computedStr,
			},
			"created_at": schema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true, PlanModifiers: computedStr},
			"updated_at": schema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true, PlanModifiers: computedStr},
		},
	}
}

func (r *EnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan environmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := client.CreateEnvironmentRequest{
		Name:          plan.Name.ValueString(),
		CloudProvider: plan.CloudProvider.ValueString(),
	}
	switch plan.CloudProvider.ValueString() {
	case "hetzner":
		in.CloudProviderHetznerRegion = plan.CloudProviderRegion.ValueString()
	case "aws":
		in.CloudProviderAWSRegion = plan.CloudProviderRegion.ValueString()
	}
	env, err := r.client.CreateEnvironment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating environment", err.Error())
		return
	}
	r.apply(&plan, env)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("id"), plan.ID)...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state environmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	env, err := r.client.GetEnvironment(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading environment", err.Error())
		return
	}
	r.apply(&state, env)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("id"), state.ID)...)
}

// Update is unreachable: every attribute is RequiresReplace. Implemented to satisfy
// the resource.Resource interface.
func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan environmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("id"), plan.ID)...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state environmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEnvironment(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error deleting environment", err.Error())
	}
}

func (r *EnvironmentResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *EnvironmentResource) apply(m *environmentResourceModel, env *client.Environment) {
	m.ID = types.StringValue(env.ID)
	m.Name = types.StringValue(env.Name)
	m.CloudProvider = types.StringValue(env.CloudProvider)
	m.CloudProviderRegion = types.StringValue(env.CloudProviderRegion)
	m.Slug = types.StringValue(env.Slug)
	m.CreatedAt = types.StringValue(env.CreatedAt)
	m.UpdatedAt = types.StringValue(env.UpdatedAt)
}
