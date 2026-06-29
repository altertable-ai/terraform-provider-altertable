package client

import "encoding/json"

// ---- Environment ----

type Environment struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Slug                string `json:"slug"`
	CloudProvider       string `json:"cloud_provider"`
	CloudProviderRegion string `json:"cloud_provider_region"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type CreateEnvironmentRequest struct {
	Name                       string `json:"name"`
	CloudProvider              string `json:"cloud_provider"`
	CloudProviderHetznerRegion string `json:"cloud_provider_hetzner_region,omitempty"`
	CloudProviderAWSRegion     string `json:"cloud_provider_aws_region,omitempty"`
}

type EnvironmentResponse struct {
	Environment Environment `json:"environment"`
}

// ---- Service Account ----

type ServiceAccount struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

type CreateServiceAccountRequest struct {
	Label string `json:"label"`
}

type UpdateServiceAccountRequest struct {
	Label string `json:"label"`
}

type ServiceAccountResponse struct {
	ServiceAccount ServiceAccount `json:"service_account"`
}

// ---- Connection ----

type Connection struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Slug          string   `json:"slug"`
	Engine        string   `json:"engine"`
	ReadOnly      bool     `json:"read_only"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Catalog       string   `json:"catalog"`
	EnvironmentID string   `json:"environment_id"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type ConnectionSshTunnel struct {
	BastionHost     string `json:"bastion_host,omitempty"`
	BastionPort     int    `json:"bastion_port,omitempty"`
	BastionUsername string `json:"bastion_username,omitempty"`
}

type ConnectionStandardConfig struct {
	Host      string               `json:"host,omitempty"`
	Port      int                  `json:"port,omitempty"`
	Database  string               `json:"database,omitempty"`
	Username  string               `json:"username,omitempty"`
	Password  string               `json:"password,omitempty"`
	Schema    string               `json:"schema,omitempty"`
	SshTunnel *ConnectionSshTunnel `json:"ssh_tunnel,omitempty"`
}

type ConnectionMysqlConfig struct {
	Host      string               `json:"host,omitempty"`
	Port      int                  `json:"port,omitempty"`
	Database  string               `json:"database,omitempty"`
	Username  string               `json:"username,omitempty"`
	Password  string               `json:"password,omitempty"`
	Schema    string               `json:"schema,omitempty"`
	SshTunnel *ConnectionSshTunnel `json:"ssh_tunnel,omitempty"`
}

type ConnectionPostgresConfig struct {
	Host      string               `json:"host,omitempty"`
	Port      int                  `json:"port,omitempty"`
	Database  string               `json:"database,omitempty"`
	Username  string               `json:"username,omitempty"`
	Password  string               `json:"password,omitempty"`
	Schema    string               `json:"schema,omitempty"`
	SshTunnel *ConnectionSshTunnel `json:"ssh_tunnel,omitempty"`
	SSLMode   string               `json:"sslmode,omitempty"`
}

type ConnectionBigQueryConfig struct {
	Dataset           string `json:"dataset,omitempty"`
	ProjectIDOverride string `json:"project_id_override,omitempty"`
}

type ConnectionSnowflakeConfig struct {
	AccountURL string `json:"account_url,omitempty"`
	Warehouse  string `json:"warehouse,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	Database   string `json:"database,omitempty"`
}

type ConnectionBucketTablesConfig struct {
	BucketID        string          `json:"bucket_id,omitempty"`
	FileFormat      string          `json:"file_format,omitempty"`
	AssumeImmutable bool            `json:"assume_immutable,omitempty"`
	Tables          json.RawMessage `json:"tables,omitempty"`
}

type ConnectionIcebergTablesConfig struct {
	BucketID string          `json:"bucket_id,omitempty"`
	Tables   json.RawMessage `json:"tables,omitempty"`
}

type ConnectionDuckDBConfig struct {
	BucketID string `json:"bucket_id,omitempty"`
	Path     string `json:"path,omitempty"`
}

type ConnectionR2CatalogConfig struct {
	Warehouse string `json:"warehouse,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
	Token     string `json:"token,omitempty"`
}

type ConnectionS3TablesConfig struct {
	Warehouse          string `json:"warehouse,omitempty"`
	DefaultRegion      string `json:"default_region,omitempty"`
	AWSAccessKeyID     string `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey string `json:"aws_secret_access_key,omitempty"`
}

type ConnectionGlueConfig struct {
	Warehouse     string `json:"warehouse,omitempty"`
	DefaultRegion string `json:"default_region,omitempty"`
	RoleARN       string `json:"role_arn,omitempty"`
}

type CreateConnectionRequest struct {
	Name                string                         `json:"name"`
	Engine              string                         `json:"engine"`
	ReadOnly            bool                           `json:"read_only,omitempty"`
	Tags                []string                       `json:"tags,omitempty"`
	Description         string                         `json:"description,omitempty"`
	StandardConfig      *ConnectionStandardConfig      `json:"standard_config,omitempty"`
	MysqlConfig         *ConnectionMysqlConfig         `json:"mysql_config,omitempty"`
	PostgresConfig      *ConnectionPostgresConfig      `json:"postgres_config,omitempty"`
	BigQueryConfig      *ConnectionBigQueryConfig      `json:"bigquery_config,omitempty"`
	SnowflakeConfig     *ConnectionSnowflakeConfig     `json:"snowflake_config,omitempty"`
	BucketTablesConfig  *ConnectionBucketTablesConfig  `json:"bucket_tables_config,omitempty"`
	IcebergTablesConfig *ConnectionIcebergTablesConfig `json:"iceberg_tables_config,omitempty"`
	DuckDBConfig        *ConnectionDuckDBConfig        `json:"duckdb_config,omitempty"`
	R2CatalogConfig     *ConnectionR2CatalogConfig     `json:"r2_catalog_config,omitempty"`
	S3TablesConfig      *ConnectionS3TablesConfig      `json:"s3_tables_config,omitempty"`
	GlueConfig          *ConnectionGlueConfig          `json:"glue_config,omitempty"`
}

type UpdateConnectionRequest struct {
	Name                string                         `json:"name,omitempty"`
	ReadOnly            *bool                          `json:"read_only,omitempty"`
	Tags                []string                       `json:"tags,omitempty"`
	Description         *string                        `json:"description,omitempty"`
	StandardConfig      *ConnectionStandardConfig      `json:"standard_config,omitempty"`
	MysqlConfig         *ConnectionMysqlConfig         `json:"mysql_config,omitempty"`
	PostgresConfig      *ConnectionPostgresConfig      `json:"postgres_config,omitempty"`
	BigQueryConfig      *ConnectionBigQueryConfig      `json:"bigquery_config,omitempty"`
	SnowflakeConfig     *ConnectionSnowflakeConfig     `json:"snowflake_config,omitempty"`
	BucketTablesConfig  *ConnectionBucketTablesConfig  `json:"bucket_tables_config,omitempty"`
	IcebergTablesConfig *ConnectionIcebergTablesConfig `json:"iceberg_tables_config,omitempty"`
	DuckDBConfig        *ConnectionDuckDBConfig        `json:"duckdb_config,omitempty"`
	R2CatalogConfig     *ConnectionR2CatalogConfig     `json:"r2_catalog_config,omitempty"`
	S3TablesConfig      *ConnectionS3TablesConfig      `json:"s3_tables_config,omitempty"`
	GlueConfig          *ConnectionGlueConfig          `json:"glue_config,omitempty"`
}

type ConnectionResponse struct {
	Connection Connection `json:"connection"`
}

type ConnectionsListResponse struct {
	Connections []Connection `json:"connections"`
}

// ---- Database ----

type Database struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Slug                  string   `json:"slug"`
	ReadOnly              bool     `json:"read_only"`
	BuiltIn               bool     `json:"built_in"`
	Description           string   `json:"description"`
	Tags                  []string `json:"tags"`
	Catalog               string   `json:"catalog"`
	BucketID              string   `json:"bucket_id"`
	SnapshotRetentionDays int      `json:"snapshot_retention_days"`
	EnvironmentID         string   `json:"environment_id"`
	CreatedAt             string   `json:"created_at"`
	UpdatedAt             string   `json:"updated_at"`
}

type CreateDatabaseRequest struct {
	Name                  string   `json:"name"`
	BucketID              string   `json:"bucket_id,omitempty"`
	ReadOnly              bool     `json:"read_only,omitempty"`
	Tags                  []string `json:"tags,omitempty"`
	SnapshotRetentionDays int      `json:"snapshot_retention_days,omitempty"`
	Description           string   `json:"description,omitempty"`
}

type UpdateDatabaseRequest struct {
	Name                  string   `json:"name,omitempty"`
	ReadOnly              *bool    `json:"read_only,omitempty"`
	Tags                  []string `json:"tags,omitempty"`
	SnapshotRetentionDays *int     `json:"snapshot_retention_days,omitempty"`
	Description           *string  `json:"description,omitempty"`
}

type DatabaseResponse struct {
	Database Database `json:"database"`
}

type DatabasesListResponse struct {
	Databases []Database `json:"databases"`
}

// ---- Credential ----

type Credential struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Username      string `json:"username"`
	Default       bool   `json:"default"`
	Active        bool   `json:"active"`
	EnvironmentID string `json:"environment_id"`
	CreatedAt     string `json:"created_at"`
	ExpiresAt     string `json:"expires_at"`
	RevokedAt     string `json:"revoked_at"`
	LastRotatedAt string `json:"last_rotated_at"`
}

type CreateCredentialRequest struct {
	Label string `json:"label,omitempty"`
}

type CreateCredentialResponse struct {
	Credential Credential `json:"credential"`
	Password   string     `json:"password"`
}

type CredentialResponse struct {
	Credential Credential `json:"credential"`
}

// ---- User ----

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UserResponse struct {
	User User `json:"user"`
}

// ---- Role set ----

type RoleGrant struct {
	Role       string `json:"role"`
	ResourceID string `json:"resource_id,omitempty"`
}

// Exactly one of UserID or ServiceAccountID is set.
type PrincipalRef struct {
	UserID           string `json:"user_id,omitempty"`
	ServiceAccountID string `json:"service_account_id,omitempty"`
}

type RoleSet struct {
	PrincipalID string      `json:"principal_id"`
	Grants      []RoleGrant `json:"grants"`
}
