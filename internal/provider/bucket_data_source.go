package provider

import (
	"context"
	"fmt"

	"github.com/altertable-ai/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*BucketDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*BucketDataSource)(nil)
)

func NewBucketDataSource() datasource.DataSource {
	return &BucketDataSource{}
}

type BucketDataSource struct {
	client *client.Client
}

type bucketDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	EnvironmentID   types.String `tfsdk:"environment_id"`
	Name            types.String `tfsdk:"name"`
	Slug            types.String `tfsdk:"slug"`
	Region          types.String `tfsdk:"region"`
	Endpoint        types.String `tfsdk:"endpoint"`
	StorageProvider types.String `tfsdk:"storage_provider"`
	BuiltIn         types.Bool   `tfsdk:"built_in"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (d *BucketDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (d *BucketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a bucket by environment and ID or slug. Access keys are write-only and never returned.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Bucket ID to look up. Exactly one of `id` or `slug` must be set.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.ExactlyOneOf(path.MatchRoot("slug"))},
			},
			"environment_id": schema.StringAttribute{MarkdownDescription: "Parent environment ID.", Required: true},
			"name":           schema.StringAttribute{MarkdownDescription: "Bucket name in the storage provider.", Computed: true},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-safe bucket slug (e.g. `BKT-1`) to look up. Exactly one of `id` or `slug` must be set.",
				Optional:            true,
				Computed:            true,
			},
			"region":           schema.StringAttribute{MarkdownDescription: "Bucket region.", Computed: true},
			"endpoint":         schema.StringAttribute{MarkdownDescription: "Custom S3-compatible endpoint URL.", Computed: true},
			"storage_provider": schema.StringAttribute{MarkdownDescription: "Storage provider derived from the endpoint: `s3`, `r2`, `gcs` or `custom`.", Computed: true},
			"built_in":         schema.BoolAttribute{MarkdownDescription: "Whether this is the environment's built-in bucket.", Computed: true},
			"created_at":       schema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true},
			"updated_at":       schema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true},
		},
	}
}

func (d *BucketDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bucketDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	lookup := data.ID.ValueString()
	if lookup == "" {
		lookup = data.Slug.ValueString()
	}
	bucket, err := d.client.GetBucket(ctx, data.EnvironmentID.ValueString(), lookup)
	if err != nil {
		resp.Diagnostics.AddError("Error reading bucket", err.Error())
		return
	}
	data.ID = types.StringValue(bucket.ID)
	data.EnvironmentID = types.StringValue(bucket.EnvironmentID)
	data.Name = types.StringValue(bucket.Name)
	data.Slug = types.StringValue(bucket.Slug)
	data.Region = optString(bucket.Region)
	data.Endpoint = optString(bucket.Endpoint)
	data.StorageProvider = types.StringValue(bucket.Provider)
	data.BuiltIn = types.BoolValue(bucket.BuiltIn)
	data.CreatedAt = types.StringValue(bucket.CreatedAt)
	data.UpdatedAt = types.StringValue(bucket.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
