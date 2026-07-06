package provider

import (
	"context"
	"fmt"

	"github.com/altertable-ai/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*BucketResource)(nil)
	_ resource.ResourceWithConfigure   = (*BucketResource)(nil)
	_ resource.ResourceWithImportState = (*BucketResource)(nil)
	_ resource.ResourceWithIdentity    = (*BucketResource)(nil)
)

type bucketIdentityModel struct {
	EnvironmentID types.String `tfsdk:"environment_id"`
	ID            types.String `tfsdk:"id"`
}

func bucketIdentity(m *bucketResourceModel) bucketIdentityModel {
	return bucketIdentityModel{EnvironmentID: m.EnvironmentID, ID: m.ID}
}

func (r *BucketResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"environment_id": identityschema.StringAttribute{RequiredForImport: true},
			"id":             identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func NewBucketResource() resource.Resource { return &BucketResource{} }

type BucketResource struct{ client *client.Client }

type bucketResourceModel struct {
	ID              types.String `tfsdk:"id"`
	EnvironmentID   types.String `tfsdk:"environment_id"`
	Name            types.String `tfsdk:"name"`
	Slug            types.String `tfsdk:"slug"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Region          types.String `tfsdk:"region"`
	Endpoint        types.String `tfsdk:"endpoint"`
	StorageProvider types.String `tfsdk:"storage_provider"`
	BuiltIn         types.Bool   `tfsdk:"built_in"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (r *BucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (r *BucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computedStr := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	// The API applies only non-blank update fields, so region/endpoint can be changed
	// in place but never cleared: removing one from the config forces a new bucket.
	clearForcesNew := []planmodifier.String{stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.RequiresReplace = req.ConfigValue.IsNull() && !req.StateValue.IsNull()
		},
		"Removing this attribute forces a new bucket (the API cannot clear it in place).",
		"Removing this attribute forces a new bucket (the API cannot clear it in place).",
	)}
	nonEmpty := []validator.String{stringvalidator.LengthAtLeast(1)}
	resp.Schema = schema.Schema{
		MarkdownDescription: "An object-storage bucket connected to an Altertable environment, providing persistence for lakehouse catalogs. Access keys are write-only: the API never returns them.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Bucket identifier.",
				Computed:            true,
				PlanModifiers:       computedStr,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Parent environment ID. Changing this forces a new bucket.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Bucket name in the storage provider (e.g. the S3 bucket name).",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-safe bucket slug (server-assigned).",
				Computed:            true,
				PlanModifiers:       computedStr,
			},
			"access_key_id": schema.StringAttribute{
				MarkdownDescription: "Access key ID used to reach the bucket (write-only; never returned by the API).",
				Required:            true,
				Sensitive:           true,
			},
			"secret_access_key": schema.StringAttribute{
				MarkdownDescription: "Secret access key used to reach the bucket (write-only; never returned by the API).",
				Required:            true,
				Sensitive:           true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Bucket region (e.g. `eu-west-1`). Removing it forces a new bucket.",
				Optional:            true,
				Validators:          nonEmpty,
				PlanModifiers:       clearForcesNew,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Custom S3-compatible endpoint URL; leave unset for AWS S3. Removing it forces a new bucket.",
				Optional:            true,
				Validators:          nonEmpty,
				PlanModifiers:       clearForcesNew,
			},
			"storage_provider": schema.StringAttribute{
				MarkdownDescription: "Storage provider derived from the endpoint: `s3`, `r2`, `gcs` or `custom`.",
				Computed:            true,
			},
			"built_in": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the environment's built-in bucket. Always `false` for buckets managed by Terraform.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true, PlanModifiers: computedStr},
			"updated_at": schema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true},
		},
	}
}

func (r *BucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	bucket, err := r.client.CreateBucket(ctx, plan.EnvironmentID.ValueString(), client.CreateBucketRequest{
		Name:            plan.Name.ValueString(),
		AccessKeyID:     plan.AccessKeyID.ValueString(),
		SecretAccessKey: plan.SecretAccessKey.ValueString(),
		Region:          plan.Region.ValueString(),
		Endpoint:        plan.Endpoint.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating bucket", err.Error())
		return
	}
	applyBucket(&plan, bucket)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, bucketIdentity(&plan))...)
}

func (r *BucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	bucket, err := r.client.GetBucket(ctx, state.EnvironmentID.ValueString(), state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket", err.Error())
		return
	}
	applyBucket(&state, bucket)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, bucketIdentity(&state))...)
}

func (r *BucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	bucket, err := r.client.UpdateBucket(ctx, plan.EnvironmentID.ValueString(), plan.ID.ValueString(), client.UpdateBucketRequest{
		Name:            plan.Name.ValueString(),
		AccessKeyID:     plan.AccessKeyID.ValueString(),
		SecretAccessKey: plan.SecretAccessKey.ValueString(),
		Region:          plan.Region.ValueString(),
		Endpoint:        plan.Endpoint.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating bucket", err.Error())
		return
	}
	applyBucket(&plan, bucket)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, bucketIdentity(&plan))...)
}

func (r *BucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteBucket(ctx, state.EnvironmentID.ValueString(), state.ID.ValueString())
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error deleting bucket", err.Error())
	}
}

func (r *BucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		// Back-compat: "environment_id:id" colon string (CLI / older Terraform).
		env, id, ok := parseEnvScopedImportID(req.ID)
		if !ok {
			resp.Diagnostics.AddError("Invalid import ID", "expected \"environment_id:id\"")
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), env)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
		return
	}
	var ident bucketIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &ident)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), ident.EnvironmentID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), ident.ID)...)
}

// applyBucket maps server fields onto the model, leaving the write-only access keys
// untouched (the API never returns them).
func applyBucket(m *bucketResourceModel, b *client.Bucket) {
	m.ID = types.StringValue(b.ID)
	m.EnvironmentID = types.StringValue(b.EnvironmentID)
	m.Name = types.StringValue(b.Name)
	m.Slug = types.StringValue(b.Slug)
	m.Region = optString(b.Region)
	m.Endpoint = optString(b.Endpoint)
	m.StorageProvider = types.StringValue(b.Provider)
	m.BuiltIn = types.BoolValue(b.BuiltIn)
	m.CreatedAt = types.StringValue(b.CreatedAt)
	m.UpdatedAt = types.StringValue(b.UpdatedAt)
}
