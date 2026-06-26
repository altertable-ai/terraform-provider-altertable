package provider

import (
	"context"
	"encoding/json"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const engineAltertable = "altertable"

func isDatabaseEngine(engine string) bool { return engine == engineAltertable }

type catalogResourceModel struct {
	ID            types.String `tfsdk:"id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Engine        types.String `tfsdk:"engine"`
	Name          types.String `tfsdk:"name"`
	Slug          types.String `tfsdk:"slug"`
	ReadOnly      types.Bool   `tfsdk:"read_only"`
	Description   types.String `tfsdk:"description"`
	Tags          types.List   `tfsdk:"tags"`
	Catalog       types.String `tfsdk:"catalog"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`

	// Database-only (engine = "altertable")
	BucketID              types.String `tfsdk:"bucket_id"`
	SnapshotRetentionDays types.Int64  `tfsdk:"snapshot_retention_days"`
	BuiltIn               types.Bool   `tfsdk:"built_in"`

	// Connection-only (one per engine family)
	StandardConfig      *standardConfigModel      `tfsdk:"standard_config"`
	MysqlConfig         *standardConfigModel      `tfsdk:"mysql_config"`
	PostgresConfig      *postgresConfigModel      `tfsdk:"postgres_config"`
	BigQueryConfig      *bigQueryConfigModel      `tfsdk:"bigquery_config"`
	SnowflakeConfig     *snowflakeConfigModel     `tfsdk:"snowflake_config"`
	BucketTablesConfig  *bucketTablesConfigModel  `tfsdk:"bucket_tables_config"`
	IcebergTablesConfig *icebergTablesConfigModel `tfsdk:"iceberg_tables_config"`
	DuckDBConfig        *duckDBConfigModel        `tfsdk:"duckdb_config"`
	R2CatalogConfig     *r2CatalogConfigModel     `tfsdk:"r2_catalog_config"`
	S3TablesConfig      *s3TablesConfigModel      `tfsdk:"s3_tables_config"`
	GlueConfig          *glueConfigModel          `tfsdk:"glue_config"`
}

type sshTunnelModel struct {
	BastionHost     types.String `tfsdk:"bastion_host"`
	BastionPort     types.Int64  `tfsdk:"bastion_port"`
	BastionUsername types.String `tfsdk:"bastion_username"`
}

// standardConfigModel is shared by standard_config and mysql_config (identical fields).
type standardConfigModel struct {
	Host      types.String    `tfsdk:"host"`
	Port      types.Int64     `tfsdk:"port"`
	Database  types.String    `tfsdk:"database"`
	Username  types.String    `tfsdk:"username"`
	Password  types.String    `tfsdk:"password"`
	Schema    types.String    `tfsdk:"schema"`
	SshTunnel *sshTunnelModel `tfsdk:"ssh_tunnel"`
}

type postgresConfigModel struct {
	Host      types.String    `tfsdk:"host"`
	Port      types.Int64     `tfsdk:"port"`
	Database  types.String    `tfsdk:"database"`
	Username  types.String    `tfsdk:"username"`
	Password  types.String    `tfsdk:"password"`
	Schema    types.String    `tfsdk:"schema"`
	SshTunnel *sshTunnelModel `tfsdk:"ssh_tunnel"`
	SSLMode   types.String    `tfsdk:"sslmode"`
}

type bigQueryConfigModel struct {
	Dataset           types.String `tfsdk:"dataset"`
	ProjectIDOverride types.String `tfsdk:"project_id_override"`
}

type snowflakeConfigModel struct {
	AccountURL types.String `tfsdk:"account_url"`
	Warehouse  types.String `tfsdk:"warehouse"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	Database   types.String `tfsdk:"database"`
}

type bucketTablesConfigModel struct {
	BucketID        types.String `tfsdk:"bucket_id"`
	FileFormat      types.String `tfsdk:"file_format"`
	AssumeImmutable types.Bool   `tfsdk:"assume_immutable"`
	Tables          types.String `tfsdk:"tables"`
}

type icebergTablesConfigModel struct {
	BucketID types.String `tfsdk:"bucket_id"`
	Tables   types.String `tfsdk:"tables"`
}

type duckDBConfigModel struct {
	BucketID types.String `tfsdk:"bucket_id"`
	Path     types.String `tfsdk:"path"`
}

type r2CatalogConfigModel struct {
	Warehouse types.String `tfsdk:"warehouse"`
	Endpoint  types.String `tfsdk:"endpoint"`
	Token     types.String `tfsdk:"token"`
}

type s3TablesConfigModel struct {
	Warehouse          types.String `tfsdk:"warehouse"`
	DefaultRegion      types.String `tfsdk:"default_region"`
	AWSAccessKeyID     types.String `tfsdk:"aws_access_key_id"`
	AWSSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
}

type glueConfigModel struct {
	Warehouse     types.String `tfsdk:"warehouse"`
	DefaultRegion types.String `tfsdk:"default_region"`
	RoleARN       types.String `tfsdk:"role_arn"`
}

// --- helpers ---

func tagStrings(l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]string, 0, len(l.Elements()))
	for _, e := range l.Elements() {
		if s, ok := e.(types.String); ok {
			out = append(out, s.ValueString())
		}
	}
	return out
}

func tagList(in []string) types.List {
	l, _ := types.ListValueFrom(context.Background(), types.StringType, in)
	return l
}

func sshToClient(m *sshTunnelModel) *client.ConnectionSshTunnel {
	if m == nil {
		return nil
	}
	return &client.ConnectionSshTunnel{
		BastionHost:     m.BastionHost.ValueString(),
		BastionPort:     int(m.BastionPort.ValueInt64()),
		BastionUsername: m.BastionUsername.ValueString(),
	}
}

func optString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// --- database mapping ---

func (m *catalogResourceModel) toCreateDatabaseRequest() client.CreateDatabaseRequest {
	return client.CreateDatabaseRequest{
		Name:                  m.Name.ValueString(),
		BucketID:              m.BucketID.ValueString(),
		ReadOnly:              m.ReadOnly.ValueBool(),
		Tags:                  tagStrings(m.Tags),
		SnapshotRetentionDays: int(m.SnapshotRetentionDays.ValueInt64()),
		Description:           m.Description.ValueString(),
	}
}

func (m *catalogResourceModel) toUpdateDatabaseRequest() client.UpdateDatabaseRequest {
	ro := m.ReadOnly.ValueBool()
	return client.UpdateDatabaseRequest{
		Name:                  m.Name.ValueString(),
		ReadOnly:              &ro,
		Tags:                  tagStrings(m.Tags),
		SnapshotRetentionDays: int(m.SnapshotRetentionDays.ValueInt64()),
		Description:           m.Description.ValueString(),
	}
}

func (m *catalogResourceModel) applyDatabase(db *client.Database) {
	m.ID = types.StringValue(db.ID)
	m.EnvironmentID = types.StringValue(db.EnvironmentID)
	m.Engine = types.StringValue(engineAltertable)
	m.Name = types.StringValue(db.Name)
	m.Slug = types.StringValue(db.Slug)
	m.ReadOnly = types.BoolValue(db.ReadOnly)
	m.Description = optString(db.Description)
	m.Tags = tagList(db.Tags)
	m.Catalog = types.StringValue(db.Catalog)
	m.CreatedAt = types.StringValue(db.CreatedAt)
	m.UpdatedAt = types.StringValue(db.UpdatedAt)
	m.BucketID = optString(db.BucketID)
	m.SnapshotRetentionDays = types.Int64Value(int64(db.SnapshotRetentionDays))
	m.BuiltIn = types.BoolValue(db.BuiltIn)
}

// --- connection mapping ---

func (m *catalogResourceModel) toCreateConnectionRequest() client.CreateConnectionRequest {
	in := client.CreateConnectionRequest{
		Name:        m.Name.ValueString(),
		Engine:      m.Engine.ValueString(),
		ReadOnly:    m.ReadOnly.ValueBool(),
		Tags:        tagStrings(m.Tags),
		Description: m.Description.ValueString(),
	}
	m.applyConfigsToCreate(&in)
	return in
}

func (m *catalogResourceModel) toUpdateConnectionRequest() client.UpdateConnectionRequest {
	ro := m.ReadOnly.ValueBool()
	in := client.UpdateConnectionRequest{
		Name:        m.Name.ValueString(),
		ReadOnly:    &ro,
		Tags:        tagStrings(m.Tags),
		Description: m.Description.ValueString(),
	}
	c := m.toCreateConnectionRequest()
	in.StandardConfig, in.MysqlConfig, in.PostgresConfig = c.StandardConfig, c.MysqlConfig, c.PostgresConfig
	in.BigQueryConfig, in.SnowflakeConfig = c.BigQueryConfig, c.SnowflakeConfig
	in.BucketTablesConfig, in.IcebergTablesConfig = c.BucketTablesConfig, c.IcebergTablesConfig
	in.DuckDBConfig, in.R2CatalogConfig = c.DuckDBConfig, c.R2CatalogConfig
	in.S3TablesConfig, in.GlueConfig = c.S3TablesConfig, c.GlueConfig
	return in
}

// applyConfigsToCreate translates each present config model into its client struct.
func (m *catalogResourceModel) applyConfigsToCreate(in *client.CreateConnectionRequest) {
	if c := m.StandardConfig; c != nil {
		in.StandardConfig = &client.ConnectionStandardConfig{
			Host: c.Host.ValueString(), Port: int(c.Port.ValueInt64()),
			Database: c.Database.ValueString(), Username: c.Username.ValueString(),
			Password: c.Password.ValueString(), Schema: c.Schema.ValueString(),
			SshTunnel: sshToClient(c.SshTunnel),
		}
	}
	if c := m.MysqlConfig; c != nil {
		in.MysqlConfig = &client.ConnectionMysqlConfig{
			Host: c.Host.ValueString(), Port: int(c.Port.ValueInt64()),
			Database: c.Database.ValueString(), Username: c.Username.ValueString(),
			Password: c.Password.ValueString(), Schema: c.Schema.ValueString(),
			SshTunnel: sshToClient(c.SshTunnel),
		}
	}
	if c := m.PostgresConfig; c != nil {
		in.PostgresConfig = &client.ConnectionPostgresConfig{
			Host: c.Host.ValueString(), Port: int(c.Port.ValueInt64()),
			Database: c.Database.ValueString(), Username: c.Username.ValueString(),
			Password: c.Password.ValueString(), Schema: c.Schema.ValueString(),
			SshTunnel: sshToClient(c.SshTunnel), SSLMode: c.SSLMode.ValueString(),
		}
	}
	if c := m.BigQueryConfig; c != nil {
		in.BigQueryConfig = &client.ConnectionBigQueryConfig{
			Dataset:           c.Dataset.ValueString(),
			ProjectIDOverride: c.ProjectIDOverride.ValueString(),
		}
	}
	if c := m.SnowflakeConfig; c != nil {
		in.SnowflakeConfig = &client.ConnectionSnowflakeConfig{
			AccountURL: c.AccountURL.ValueString(), Warehouse: c.Warehouse.ValueString(),
			Username: c.Username.ValueString(), Password: c.Password.ValueString(),
			Database: c.Database.ValueString(),
		}
	}
	if c := m.BucketTablesConfig; c != nil {
		in.BucketTablesConfig = &client.ConnectionBucketTablesConfig{
			BucketID: c.BucketID.ValueString(), FileFormat: c.FileFormat.ValueString(),
			AssumeImmutable: c.AssumeImmutable.ValueBool(),
			Tables:          json.RawMessage(c.Tables.ValueString()),
		}
	}
	if c := m.IcebergTablesConfig; c != nil {
		in.IcebergTablesConfig = &client.ConnectionIcebergTablesConfig{
			BucketID: c.BucketID.ValueString(),
			Tables:   json.RawMessage(c.Tables.ValueString()),
		}
	}
	if c := m.DuckDBConfig; c != nil {
		in.DuckDBConfig = &client.ConnectionDuckDBConfig{
			BucketID: c.BucketID.ValueString(), Path: c.Path.ValueString(),
		}
	}
	if c := m.R2CatalogConfig; c != nil {
		in.R2CatalogConfig = &client.ConnectionR2CatalogConfig{
			Warehouse: c.Warehouse.ValueString(), Endpoint: c.Endpoint.ValueString(),
			Token: c.Token.ValueString(),
		}
	}
	if c := m.S3TablesConfig; c != nil {
		in.S3TablesConfig = &client.ConnectionS3TablesConfig{
			Warehouse: c.Warehouse.ValueString(), DefaultRegion: c.DefaultRegion.ValueString(),
			AWSAccessKeyID: c.AWSAccessKeyID.ValueString(), AWSSecretAccessKey: c.AWSSecretAccessKey.ValueString(),
		}
	}
	if c := m.GlueConfig; c != nil {
		in.GlueConfig = &client.ConnectionGlueConfig{
			Warehouse: c.Warehouse.ValueString(), DefaultRegion: c.DefaultRegion.ValueString(),
			RoleARN: c.RoleARN.ValueString(),
		}
	}
}

func (m *catalogResourceModel) applyConnection(con *client.Connection) {
	m.ID = types.StringValue(con.ID)
	m.EnvironmentID = types.StringValue(con.EnvironmentID)
	m.Engine = types.StringValue(con.Engine)
	m.Name = types.StringValue(con.Name)
	m.Slug = types.StringValue(con.Slug)
	m.ReadOnly = types.BoolValue(con.ReadOnly)
	m.Description = optString(con.Description)
	m.Tags = tagList(con.Tags)
	m.Catalog = types.StringValue(con.Catalog)
	m.CreatedAt = types.StringValue(con.CreatedAt)
	m.UpdatedAt = types.StringValue(con.UpdatedAt)
	// Connection config blocks are write-only; the API never returns them, so they
	// are NOT touched here — the caller preserves them from prior state.
}
