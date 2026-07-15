package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/altertable-ai/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = (*CatalogResource)(nil)
	_ resource.ResourceWithConfigure        = (*CatalogResource)(nil)
	_ resource.ResourceWithImportState      = (*CatalogResource)(nil)
	_ resource.ResourceWithIdentity         = (*CatalogResource)(nil)
	_ resource.ResourceWithConfigValidators = (*CatalogResource)(nil)
	_ resource.ResourceWithValidateConfig   = (*CatalogResource)(nil)
)

type catalogIdentityModel struct {
	EnvironmentID types.String `tfsdk:"environment_id"`
	ID            types.String `tfsdk:"id"`
}

func catalogIdentity(m *catalogResourceModel) catalogIdentityModel {
	return catalogIdentityModel{EnvironmentID: m.EnvironmentID, ID: m.ID}
}

func (r *CatalogResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"environment_id": identityschema.StringAttribute{RequiredForImport: true},
			"id":             identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func NewCatalogResource() resource.Resource {
	return &CatalogResource{}
}

type CatalogResource struct {
	client *client.Client
}

func (r *CatalogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalog"
}

var portValidators = []validator.Int64{int64validator.Between(1, 65535)}

func sshTunnelAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Optional SSH bastion tunnel used to reach the database.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"bastion_host":     schema.StringAttribute{MarkdownDescription: "Bastion host.", Required: true},
			"bastion_port":     schema.Int64Attribute{MarkdownDescription: "Bastion SSH port.", Optional: true, Validators: portValidators},
			"bastion_username": schema.StringAttribute{MarkdownDescription: "Bastion SSH username.", Required: true},
		},
	}
}

// sqlConfigAttributes returns the fields shared by postgres_config and mysql_config. The caller
// adds sslmode for postgres_config. Secret fields are write-only and never stored in state or
// shown in the diff.
func sqlConfigAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"host":       schema.StringAttribute{MarkdownDescription: "Database host.", Required: true},
		"port":       schema.Int64Attribute{MarkdownDescription: "Database port.", Required: true, Validators: portValidators},
		"database":   schema.StringAttribute{MarkdownDescription: "Database name.", Required: true},
		"username":   schema.StringAttribute{MarkdownDescription: "Login username.", Optional: true},
		"password":   schema.StringAttribute{MarkdownDescription: "Login password (write-only; never stored in state).", Optional: true, Sensitive: true, WriteOnly: true},
		"schema":     schema.StringAttribute{MarkdownDescription: "Default schema.", Optional: true},
		"ssh_tunnel": sshTunnelAttribute(),
	}
}

func (r *CatalogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computedStr := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	forceNewStr := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	computedBool := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}

	postgresAttrs := sqlConfigAttributes()
	postgresAttrs["sslmode"] = schema.StringAttribute{
		MarkdownDescription: "Postgres SSL mode (e.g. `require`).",
		Optional:            true,
	}

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
				MarkdownDescription: "Catalog engine. `altertable` for a native database; otherwise an external connection engine (one of `postgres`, `redshift`, `supabase`, `mysql`, `mariadb`, `snowflake`, `bigquery`, `buckettables`, `icebergtables`, `r2catalog`, `s3tables`, `glue`, `duckdb`). Changing this forces a new catalog.",
				Required:            true,
				PlanModifiers:       forceNewStr,
				Validators:          []validator.String{stringvalidator.OneOf(allEngines()...)},
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
			// updated_at changes on every write, so it can't be pinned with UseStateForUnknown.
			"updated_at": schema.StringAttribute{MarkdownDescription: "Last update timestamp.", Computed: true},

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

			// Connection-only config blocks (one per engine family). Exactly the block matching
			// `engine` may be set (enforced in ValidateConfig); each carries only that engine's
			// fields, so a plan never shows settings from an unrelated engine. Secret fields are
			// write-only and never stored in state.
			"postgres_config": schema.SingleNestedAttribute{
				MarkdownDescription: "PostgreSQL connection settings. Engines: `postgres`, `redshift`, `supabase`.",
				Optional:            true,
				Attributes:          postgresAttrs,
			},
			"mysql_config": schema.SingleNestedAttribute{
				MarkdownDescription: "MySQL connection settings. Engines: `mysql`, `mariadb`.",
				Optional:            true,
				Attributes:          sqlConfigAttributes(),
			},
			"snowflake_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Snowflake connection settings. Engine: `snowflake`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"account_url": schema.StringAttribute{MarkdownDescription: "Snowflake account URL.", Required: true},
					"warehouse":   schema.StringAttribute{MarkdownDescription: "Warehouse name.", Required: true},
					"username":    schema.StringAttribute{MarkdownDescription: "Login username.", Optional: true},
					"password":    schema.StringAttribute{MarkdownDescription: "Login password (write-only; never stored in state).", Optional: true, Sensitive: true, WriteOnly: true},
					"database":    schema.StringAttribute{MarkdownDescription: "Database name.", Optional: true},
				},
			},
			"bigquery_config": schema.SingleNestedAttribute{
				MarkdownDescription: "BigQuery connection settings. Engine: `bigquery`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"dataset":             schema.StringAttribute{MarkdownDescription: "BigQuery dataset.", Optional: true},
					"project_id_override": schema.StringAttribute{MarkdownDescription: "Override the GCP project ID.", Optional: true},
				},
			},
			"bucket_tables_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Bucket tables connection settings. Engine: `buckettables`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"bucket_id":        schema.StringAttribute{MarkdownDescription: "Storage bucket ID.", Required: true},
					"file_format":      schema.StringAttribute{MarkdownDescription: "File format. One of `parquet`, `csv`, `json`.", Required: true, Validators: []validator.String{stringvalidator.OneOf("parquet", "csv", "json")}},
					"tables":           schema.StringAttribute{MarkdownDescription: "Table definitions as a JSON object mapping table names to strings.", Required: true},
					"assume_immutable": schema.BoolAttribute{MarkdownDescription: "Treat files as immutable.", Optional: true},
				},
			},
			"iceberg_tables_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Iceberg tables connection settings. Engine: `icebergtables`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"bucket_id": schema.StringAttribute{MarkdownDescription: "Storage bucket ID.", Required: true},
					"tables":    schema.StringAttribute{MarkdownDescription: "Table definitions as a JSON object mapping table names to strings.", Required: true},
				},
			},
			"duckdb_config": schema.SingleNestedAttribute{
				MarkdownDescription: "DuckDB connection settings. Engine: `duckdb`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"bucket_id": schema.StringAttribute{MarkdownDescription: "Storage bucket ID.", Required: true},
					"path":      schema.StringAttribute{MarkdownDescription: "Path to the DuckDB file (must end with `.duckdb`).", Required: true, Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`\.duckdb$`), "path must end with .duckdb")}},
				},
			},
			"r2_catalog_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Cloudflare R2 catalog connection settings. Engine: `r2catalog`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"warehouse": schema.StringAttribute{MarkdownDescription: "Warehouse name.", Required: true},
					"endpoint":  schema.StringAttribute{MarkdownDescription: "R2 endpoint.", Required: true},
					"token":     schema.StringAttribute{MarkdownDescription: "Access token (write-only; never stored in state).", Required: true, Sensitive: true, WriteOnly: true},
				},
			},
			"s3_tables_config": schema.SingleNestedAttribute{
				MarkdownDescription: "AWS S3 Tables connection settings. Engine: `s3tables`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"warehouse":             schema.StringAttribute{MarkdownDescription: "Warehouse name.", Required: true},
					"default_region":        schema.StringAttribute{MarkdownDescription: "AWS region.", Required: true},
					"aws_access_key_id":     schema.StringAttribute{MarkdownDescription: "AWS access key ID.", Required: true},
					"aws_secret_access_key": schema.StringAttribute{MarkdownDescription: "AWS secret access key (write-only; never stored in state).", Required: true, Sensitive: true, WriteOnly: true},
				},
			},
			"glue_config": schema.SingleNestedAttribute{
				MarkdownDescription: "AWS Glue connection settings. Engine: `glue`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"warehouse":      schema.StringAttribute{MarkdownDescription: "Warehouse name.", Required: true},
					"default_region": schema.StringAttribute{MarkdownDescription: "AWS region.", Required: true},
					"role_arn":       schema.StringAttribute{MarkdownDescription: "IAM role ARN.", Required: true},
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
	for _, msg := range validateCatalogConfig(&cfg) {
		resp.Diagnostics.AddError("Invalid catalog config", msg)
	}

	if c := cfg.BucketTablesConfig; c != nil {
		if err := validTablesJSON(c.Tables); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("bucket_tables_config").AtName("tables"), "Invalid tables", err.Error())
		}
	}
	if c := cfg.IcebergTablesConfig; c != nil {
		if err := validTablesJSON(c.Tables); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("iceberg_tables_config").AtName("tables"), "Invalid tables", err.Error())
		}
	}
}

func validateCatalogConfig(cfg *catalogResourceModel) []string {
	engine := cfg.Engine.ValueString()
	var errs []string
	set := cfg.setConnectionBlocks()

	if isDatabaseEngine(engine) {
		if len(set) > 0 {
			errs = append(errs, fmt.Sprintf("%s not allowed when engine = %q", strings.Join(set, ", "), engineAltertable))
		}
		return errs
	}

	// Connection engine: bucket_id and snapshot_retention_days are database-only.
	if !cfg.BucketID.IsNull() || !cfg.SnapshotRetentionDays.IsNull() {
		errs = append(errs, "bucket_id and snapshot_retention_days are only valid when engine = \"altertable\"")
	}

	block, known := engineBlocks[engine]
	if !known {
		return errs // an unknown engine is already reported by the engine OneOf validator
	}

	// The engine's own block is required; every other block is invalid.
	present := false
	for _, b := range set {
		if b == block {
			present = true
			continue
		}
		errs = append(errs, fmt.Sprintf("%s is not valid for engine %q; use %s", b, engine, block))
	}
	if !present {
		errs = append(errs, fmt.Sprintf("engine %q requires a %s block", engine, block))
	}
	return errs
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
		// Write-only secrets are null in the plan; read them from config and overlay them onto
		// the config blocks (top-level fields still come from the plan).
		var cfg catalogResourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in := plan.toCreateConnectionRequest()
		cfg.applyConfigsToCreate(&in)
		con, err := r.client.CreateConnection(ctx, env, in)
		if err != nil {
			resp.Diagnostics.AddError("Error creating catalog (connection)", err.Error())
			return
		}
		applyConnectionPreservingConfig(&plan, con)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, catalogIdentity(&plan))...)
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
	resp.Diagnostics.Append(resp.Identity.Set(ctx, catalogIdentity(&state))...)
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
		// Write-only secrets are null in the plan; read them from config and overlay them onto
		// the config blocks (top-level fields still come from the plan).
		var cfg catalogResourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in := plan.toUpdateConnectionRequest()
		setUpdateConfigs(&in, cfg.toCreateConnectionRequest())
		con, err := r.client.UpdateConnection(ctx, env, id, in)
		if err != nil {
			resp.Diagnostics.AddError("Error updating catalog (connection)", err.Error())
			return
		}
		applyConnectionPreservingConfig(&plan, con)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, catalogIdentity(&plan))...)
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

	// Known engine: read only the matching endpoint.
	if engine != "" {
		if isDatabaseEngine(engine) {
			db, err := r.client.GetDatabase(ctx, env, id)
			if err != nil {
				if isNotFound(err) {
					return false, nil
				}
				return false, err
			}
			state.applyDatabase(db)
			return true, nil
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

	// Empty engine (fresh import): probe both endpoints and refuse to guess if an id
	// resolves to both a database and a connection.
	db, dbErr := r.client.GetDatabase(ctx, env, id)
	if dbErr != nil && !isNotFound(dbErr) {
		return false, dbErr
	}
	con, conErr := r.client.GetConnection(ctx, env, id)
	if conErr != nil && !isNotFound(conErr) {
		return false, conErr
	}
	dbFound, conFound := dbErr == nil, conErr == nil
	switch {
	case dbFound && conFound:
		return false, fmt.Errorf("ambiguous catalog %q in environment %q: it matches both a database and a connection; set engine to disambiguate the import", id, env)
	case dbFound:
		state.applyDatabase(db)
		return true, nil
	case conFound:
		applyConnectionPreservingConfig(state, con)
		return true, nil
	default:
		return false, nil
	}
}

// applyConnectionPreservingConfig maps server fields but keeps write-only config
// blocks intact (the API never returns them).
func applyConnectionPreservingConfig(m *catalogResourceModel, con *client.Connection) {
	m.applyConnection(con)
}

func (r *CatalogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
	var ident catalogIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &ident)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), ident.EnvironmentID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), ident.ID)...)
}
