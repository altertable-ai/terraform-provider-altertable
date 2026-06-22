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
	_ resource.Resource                = (*CatalogResource)(nil)
	_ resource.ResourceWithConfigure   = (*CatalogResource)(nil)
	_ resource.ResourceWithImportState = (*CatalogResource)(nil)
)

func NewCatalogResource() resource.Resource {
	return &CatalogResource{}
}

type CatalogResource struct {
	client *client.Client
}

type catalogResourceModel struct {
	ID            types.String `tfsdk:"id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Slug          types.String `tfsdk:"slug"`
	Name          types.String `tfsdk:"name"`
}

func (r *CatalogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalog"
}

func (r *CatalogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A catalog within an Altertable environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Catalog identifier.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Parent environment ID. Changing this forces a new catalog.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-safe catalog slug. Changing this forces a new catalog.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable catalog name.",
				Required:            true,
			},
		},
	}
}

func (r *CatalogResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CatalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan catalogResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cat, err := r.client.CreateCatalog(ctx, client.CatalogCreateInput{
		EnvironmentID: plan.EnvironmentID.ValueString(),
		Slug:          plan.Slug.ValueString(),
		Name:          plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating catalog", err.Error())
		return
	}
	plan.ID = types.StringValue(cat.ID)
	plan.EnvironmentID = types.StringValue(cat.EnvironmentID)
	plan.Slug = types.StringValue(cat.Slug)
	plan.Name = types.StringValue(cat.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CatalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state catalogResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cat, err := r.client.GetCatalog(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading catalog", err.Error())
		return
	}
	state.EnvironmentID = types.StringValue(cat.EnvironmentID)
	state.Slug = types.StringValue(cat.Slug)
	state.Name = types.StringValue(cat.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CatalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan catalogResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cat, err := r.client.UpdateCatalog(ctx, plan.ID.ValueString(), client.CatalogUpdateInput{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating catalog", err.Error())
		return
	}
	plan.EnvironmentID = types.StringValue(cat.EnvironmentID)
	plan.Slug = types.StringValue(cat.Slug)
	plan.Name = types.StringValue(cat.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CatalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state catalogResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCatalog(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error deleting catalog", err.Error())
	}
}

func (r *CatalogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
