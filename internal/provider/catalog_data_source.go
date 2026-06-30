package provider

import (
	"context"
	"fmt"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*CatalogDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*CatalogDataSource)(nil)
)

func NewCatalogDataSource() datasource.DataSource {
	return &CatalogDataSource{}
}

type CatalogDataSource struct {
	client *client.Client
}

type catalogDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	EnvironmentID         types.String `tfsdk:"environment_id"`
	Engine                types.String `tfsdk:"engine"`
	Name                  types.String `tfsdk:"name"`
	Slug                  types.String `tfsdk:"slug"`
	ReadOnly              types.Bool   `tfsdk:"read_only"`
	Description           types.String `tfsdk:"description"`
	Tags                  types.List   `tfsdk:"tags"`
	Catalog               types.String `tfsdk:"catalog"`
	BuiltIn               types.Bool   `tfsdk:"built_in"`
	BucketID              types.String `tfsdk:"bucket_id"`
	SnapshotRetentionDays types.Int64  `tfsdk:"snapshot_retention_days"`
}

func (d *CatalogDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalog"
}

func (d *CatalogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a catalog by environment and ID (or slug), probing databases then connections.",
		Attributes: map[string]schema.Attribute{
			"id":                      schema.StringAttribute{MarkdownDescription: "Catalog ID (or slug) to look up.", Required: true},
			"environment_id":          schema.StringAttribute{MarkdownDescription: "Parent environment ID.", Required: true},
			"engine":                  schema.StringAttribute{MarkdownDescription: "Catalog engine (`altertable` for native databases).", Computed: true},
			"name":                    schema.StringAttribute{MarkdownDescription: "Human-readable catalog name.", Computed: true},
			"slug":                    schema.StringAttribute{MarkdownDescription: "URL-safe catalog slug.", Computed: true},
			"read_only":               schema.BoolAttribute{MarkdownDescription: "Whether the catalog is read-only.", Computed: true},
			"description":             schema.StringAttribute{MarkdownDescription: "Optional description.", Computed: true},
			"tags":                    schema.ListAttribute{MarkdownDescription: "List of tags.", ElementType: types.StringType, Computed: true},
			"catalog":                 schema.StringAttribute{MarkdownDescription: "Underlying catalog identifier.", Computed: true},
			"built_in":                schema.BoolAttribute{MarkdownDescription: "Whether this is a built-in database.", Computed: true},
			"bucket_id":               schema.StringAttribute{MarkdownDescription: "Storage bucket ID (databases only).", Computed: true},
			"snapshot_retention_days": schema.Int64Attribute{MarkdownDescription: "Snapshot retention in days (databases only).", Computed: true},
		},
	}
}

func (d *CatalogDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CatalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data catalogDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	env, id := data.EnvironmentID.ValueString(), data.ID.ValueString()
	if db, err := d.client.GetDatabase(ctx, env, id); err == nil {
		data.ID = types.StringValue(db.ID)
		data.Engine = types.StringValue(engineAltertable)
		data.Name = types.StringValue(db.Name)
		data.Slug = types.StringValue(db.Slug)
		data.ReadOnly = types.BoolValue(db.ReadOnly)
		data.Description = optString(db.Description)
		data.Tags = tagList(db.Tags)
		data.Catalog = types.StringValue(db.Catalog)
		data.BuiltIn = types.BoolValue(db.BuiltIn)
		data.BucketID = optString(db.BucketID)
		data.SnapshotRetentionDays = types.Int64Value(int64(db.SnapshotRetentionDays))
		resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
		return
	} else if !isNotFound(err) {
		resp.Diagnostics.AddError("Error reading catalog", err.Error())
		return
	}
	con, err := d.client.GetConnection(ctx, env, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading catalog", err.Error())
		return
	}
	data.ID = types.StringValue(con.ID)
	data.Engine = types.StringValue(con.Engine)
	data.Name = types.StringValue(con.Name)
	data.Slug = types.StringValue(con.Slug)
	data.ReadOnly = types.BoolValue(con.ReadOnly)
	data.Description = optString(con.Description)
	data.Tags = tagList(con.Tags)
	data.Catalog = types.StringValue(con.Catalog)
	data.BuiltIn = types.BoolValue(false)
	data.BucketID = types.StringNull()
	data.SnapshotRetentionDays = types.Int64Null()
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
