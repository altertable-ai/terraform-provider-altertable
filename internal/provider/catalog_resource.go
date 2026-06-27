package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = (*CatalogResource)(nil)
	_ resource.ResourceWithConfigure        = (*CatalogResource)(nil)
	_ resource.ResourceWithImportState      = (*CatalogResource)(nil)
	_ resource.ResourceWithConfigValidators = (*CatalogResource)(nil)
	_ resource.ResourceWithValidateConfig   = (*CatalogResource)(nil)
)

func NewCatalogResource() resource.Resource {
	return &CatalogResource{}
}

type CatalogResource struct {
	client *client.Client
}

func (r *CatalogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalog"
}

// sshTunnelAttribute returns the nested ssh_tunnel block shared by the
// standard/mysql/postgres connection configs.
func sshTunnelAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Optional SSH bastion tunnel used to reach the database.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"bastion_host":     schema.StringAttribute{MarkdownDescription: "Bastion host.", Optional: true},
			"bastion_port":     schema.Int64Attribute{MarkdownDescription: "Bastion SSH port.", Optional: true},
			"bastion_username": schema.StringAttribute{MarkdownDescription: "Bastion SSH username.", Optional: true},
		},
	}
}

// standardConfigAttributes returns the field set shared by standard_config and
// mysql_config.
func standardConfigAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"host":       schema.StringAttribute{MarkdownDescription: "Database host.", Optional: true},
		"port":       schema.Int64Attribute{MarkdownDescription: "Database port.", Optional: true},
		"database":   schema.StringAttribute{MarkdownDescription: "Database name.", Optional: true},
		"username":   schema.StringAttribute{MarkdownDescription: "Login username.", Optional: true},
		"password":   schema.StringAttribute{MarkdownDescription: "Login password (write-only).", Optional: true, Sensitive: true},
		"schema":     schema.StringAttribute{MarkdownDescription: "Default schema.", Optional: true},
		"ssh_tunnel": sshTunnelAttribute(),
	}
}

func (r *CatalogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computedStr := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	forceNewStr := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	computedBool := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}

	postgresAttrs := standardConfigAttributes()
	postgresAttrs["sslmode"] = schema.StringAttribute{MarkdownDescription: "Postgres SSL mode (e.g. `require`).", Optional: true}

	resp.Schema = schema.Schema{
		MarkdownDescription: "A catalog within an Altertable environment. This is a facade: `engine = \"altertable\"` manages a native database; any other engine manages an external connection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Catalog identifier.",
				Computed:            true,
				PlanModifiers:       computedStr,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Parent environment ID. Changing this forces a new catalog.",
				Required:            true,
				PlanModifiers:       forceNewStr,
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: "Catalog engine. `altertable` for a native database; otherwise an external connection engine (e.g. `postgres`, `mysql`, `bigquery`). Changing this forces a new catalog.",
				Required:            true,
				PlanModifiers:       forceNewStr,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable catalog name.",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-safe catalog slug (server-assigned).",
				Computed:            true,
				PlanModifiers:       computedStr,
			},
			"read_only": schema.BoolAttribute{
				MarkdownDescription: "Whether the catalog is read-only.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       computedBool,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Optional list of tags.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"catalog": schema.StringAttribute{
				MarkdownDescription: "Underlying catalog identifier (server-assigned).",
				Computed:            true,
				PlanModifiers:       computedStr,
			},
			"created_at": schema.StringAttribute{MarkdownDescription: "Creation timestamp.", Computed: true, PlanModifiers: computedStr},
			"updated_at": schema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true, PlanModifiers: computedStr},

			// Database-only (engine = "altertable")
			"bucket_id": schema.StringAttribute{
				MarkdownDescription: "Storage bucket ID. Only valid when `engine = \"altertable\"`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       computedStr,
			},
			"snapshot_retention_days": schema.Int64Attribute{
				MarkdownDescription: "Snapshot retention in days. Only valid when `engine = \"altertable\"`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"built_in": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a built-in database.",
				Computed:            true,
				PlanModifiers:       computedBool,
			},

			// Connection-only config blocks (one per engine family). Write-only;
			// the API never returns them.
			"standard_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Generic SQL connection config.",
				Optional:            true,
				Attributes:          standardConfigAttributes(),
			},
			"mysql_config": schema.SingleNestedAttribute{
				MarkdownDescription: "MySQL connection config.",
				Optional:            true,
				Attributes:          standardConfigAttributes(),
			},
			"postgres_config": schema.SingleNestedAttribute{
				MarkdownDescription: "PostgreSQL connection config.",
				Optional:            true,
				Attributes:          postgresAttrs,
			},
			"bigquery_config": schema.SingleNestedAttribute{
				MarkdownDescription: "BigQuery connection config.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"dataset":             schema.StringAttribute{MarkdownDescription: "BigQuery dataset.", Optional: true},
					"project_id_override": schema.StringAttribute{MarkdownDescription: "Override the GCP project ID.", Optional: true},
				},
			},
			"snowflake_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Snowflake connection config.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"account_url": schema.StringAttribute{MarkdownDescription: "Snowflake account URL.", Optional: true},
					"warehouse":   schema.StringAttribute{MarkdownDescription: "Warehouse name.", Optional: true},
					"username":    schema.StringAttribute{MarkdownDescription: "Login username.", Optional: true},
					"password":    schema.StringAttribute{MarkdownDescription: "Login password (write-only).", Optional: true, Sensitive: true},
					"database":    schema.StringAttribute{MarkdownDescription: "Database name.", Optional: true},
				},
			},
			"bucket_tables_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Bucket tables connection config.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"bucket_id":        schema.StringAttribute{MarkdownDescription: "Storage bucket ID.", Optional: true},
					"file_format":      schema.StringAttribute{MarkdownDescription: "File format (e.g. `parquet`).", Optional: true},
					"assume_immutable": schema.BoolAttribute{MarkdownDescription: "Treat files as immutable.", Optional: true},
					"tables":           schema.StringAttribute{MarkdownDescription: "Table definitions as a JSON string.", Optional: true},
				},
			},
			"iceberg_tables_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Iceberg tables connection config.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"bucket_id": schema.StringAttribute{MarkdownDescription: "Storage bucket ID.", Optional: true},
					"tables":    schema.StringAttribute{MarkdownDescription: "Table definitions as a JSON string.", Optional: true},
				},
			},
			"duckdb_config": schema.SingleNestedAttribute{
				MarkdownDescription: "DuckDB connection config.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"bucket_id": schema.StringAttribute{MarkdownDescription: "Storage bucket ID.", Optional: true},
					"path":      schema.StringAttribute{MarkdownDescription: "Path to the DuckDB file.", Optional: true},
				},
			},
			"r2_catalog_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Cloudflare R2 catalog connection config.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"warehouse": schema.StringAttribute{MarkdownDescription: "Warehouse name.", Optional: true},
					"endpoint":  schema.StringAttribute{MarkdownDescription: "R2 endpoint.", Optional: true},
					"token":     schema.StringAttribute{MarkdownDescription: "Access token (write-only).", Optional: true, Sensitive: true},
				},
			},
			"s3_tables_config": schema.SingleNestedAttribute{
				MarkdownDescription: "AWS S3 Tables connection config.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"warehouse":             schema.StringAttribute{MarkdownDescription: "Warehouse name.", Optional: true},
					"default_region":        schema.StringAttribute{MarkdownDescription: "AWS region.", Optional: true},
					"aws_access_key_id":     schema.StringAttribute{MarkdownDescription: "AWS access key ID.", Optional: true},
					"aws_secret_access_key": schema.StringAttribute{MarkdownDescription: "AWS secret access key (write-only).", Optional: true, Sensitive: true},
				},
			},
			"glue_config": schema.SingleNestedAttribute{
				MarkdownDescription: "AWS Glue connection config.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"warehouse":      schema.StringAttribute{MarkdownDescription: "Warehouse name.", Optional: true},
					"default_region": schema.StringAttribute{MarkdownDescription: "AWS region.", Optional: true},
					"role_arn":       schema.StringAttribute{MarkdownDescription: "IAM role ARN.", Optional: true},
				},
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

func (r *CatalogResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return nil // structural validation handled in ValidateConfig
}

func (r *CatalogResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg catalogResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() || cfg.Engine.IsUnknown() || cfg.Engine.IsNull() {
		return
	}
	isDB := isDatabaseEngine(cfg.Engine.ValueString())
	anyConfig := cfg.StandardConfig != nil || cfg.MysqlConfig != nil || cfg.PostgresConfig != nil ||
		cfg.BigQueryConfig != nil || cfg.SnowflakeConfig != nil || cfg.BucketTablesConfig != nil ||
		cfg.IcebergTablesConfig != nil || cfg.DuckDBConfig != nil || cfg.R2CatalogConfig != nil ||
		cfg.S3TablesConfig != nil || cfg.GlueConfig != nil
	if isDB && anyConfig {
		resp.Diagnostics.AddError("Invalid catalog config", "connection *_config blocks are not allowed when engine = \"altertable\"")
	}
	if !isDB && (!cfg.BucketID.IsNull() || !cfg.SnapshotRetentionDays.IsNull()) {
		resp.Diagnostics.AddError("Invalid catalog config", "bucket_id and snapshot_retention_days are only valid when engine = \"altertable\"")
	}
}

func (r *CatalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan catalogResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	env := plan.EnvironmentID.ValueString()
	if isDatabaseEngine(plan.Engine.ValueString()) {
		db, err := r.client.CreateDatabase(ctx, env, plan.toCreateDatabaseRequest())
		if err != nil {
			resp.Diagnostics.AddError("Error creating catalog (database)", err.Error())
			return
		}
		plan.applyDatabase(db)
	} else {
		con, err := r.client.CreateConnection(ctx, env, plan.toCreateConnectionRequest())
		if err != nil {
			resp.Diagnostics.AddError("Error creating catalog (connection)", err.Error())
			return
		}
		applyConnectionPreservingConfig(&plan, con)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CatalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state catalogResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, err := r.readCatalog(ctx, &state)
	if err != nil {
		resp.Diagnostics.AddError("Error reading catalog", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CatalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan catalogResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	env, id := plan.EnvironmentID.ValueString(), plan.ID.ValueString()
	if isDatabaseEngine(plan.Engine.ValueString()) {
		db, err := r.client.UpdateDatabase(ctx, env, id, plan.toUpdateDatabaseRequest())
		if err != nil {
			resp.Diagnostics.AddError("Error updating catalog (database)", err.Error())
			return
		}
		plan.applyDatabase(db)
	} else {
		con, err := r.client.UpdateConnection(ctx, env, id, plan.toUpdateConnectionRequest())
		if err != nil {
			resp.Diagnostics.AddError("Error updating catalog (connection)", err.Error())
			return
		}
		applyConnectionPreservingConfig(&plan, con)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CatalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state catalogResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	env, id := state.EnvironmentID.ValueString(), state.ID.ValueString()
	var err error
	if isDatabaseEngine(state.Engine.ValueString()) {
		err = r.client.DeleteDatabase(ctx, env, id)
	} else {
		err = r.client.DeleteConnection(ctx, env, id)
	}
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error deleting catalog", err.Error())
	}
}

// readCatalog refreshes state, routing by engine. With a known engine it hits the
// matching endpoint. With an empty engine (fresh import) it probes databases, then
// connections. Returns (found, error); found=false means 404 (drop from state).
func (r *CatalogResource) readCatalog(ctx context.Context, state *catalogResourceModel) (bool, error) {
	env, id := state.EnvironmentID.ValueString(), state.ID.ValueString()
	engine := state.Engine.ValueString()

	if engine == "" || isDatabaseEngine(engine) {
		db, err := r.client.GetDatabase(ctx, env, id)
		if err == nil {
			state.applyDatabase(db)
			return true, nil
		}
		if !isNotFound(err) {
			return false, err
		}
		if engine != "" { // engine known to be a DB, 404 => gone
			return false, nil
		}
	}
	con, err := r.client.GetConnection(ctx, env, id)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	applyConnectionPreservingConfig(state, con)
	return true, nil
}

// applyConnectionPreservingConfig maps server fields but keeps write-only config
// blocks intact (the API never returns them).
func applyConnectionPreservingConfig(m *catalogResourceModel, con *client.Connection) {
	m.applyConnection(con)
}

// parseCatalogImportID splits the "environment_id:id" back-compat import string.
// Only the first colon splits, so an id containing ':' is preserved.
func parseCatalogImportID(s string) (env, id string, ok bool) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func (r *CatalogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	env, id, ok := parseCatalogImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError("Invalid import ID", "expected \"environment_id:id\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), env)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
