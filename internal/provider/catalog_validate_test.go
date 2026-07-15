package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// engineBlocks must declare a spec for every connection engine the schema accepts (excluding the
// altertable database facade).
func TestEngineBlocksCoverAllConnectionEngines(t *testing.T) {
	for _, engine := range connectionEngines {
		if _, ok := engineBlocks[engine]; !ok {
			t.Errorf("engine %q has no engineBlocks spec", engine)
		}
	}
	if len(engineBlocks) != len(connectionEngines) {
		t.Errorf("engineBlocks has %d entries, want %d (one per connection engine)", len(engineBlocks), len(connectionEngines))
	}
}

// Every block an engine points at must be a real *_config block on the model, and every block on
// the model must be reachable by some engine (no orphan blocks).
func TestEngineBlocksReferenceRealBlocks(t *testing.T) {
	real := map[string]bool{}
	for _, b := range fullyPopulatedModel().setConnectionBlocks() {
		real[b] = true
	}
	used := map[string]bool{}
	for engine, block := range engineBlocks {
		if !real[block] {
			t.Errorf("engine %q references unknown block %q", engine, block)
		}
		used[block] = true
	}
	for b := range real {
		if !used[b] {
			t.Errorf("block %q is defined on the model but no engine uses it", b)
		}
	}
}

func TestAllEnginesIncludesAltertableAndConnections(t *testing.T) {
	all := allEngines()
	if len(all) != len(connectionEngines)+1 {
		t.Fatalf("allEngines len = %d, want %d", len(all), len(connectionEngines)+1)
	}
	var hasAltertable bool
	for _, e := range all {
		if e == engineAltertable {
			hasAltertable = true
		}
	}
	if !hasAltertable {
		t.Error("allEngines must include the altertable database facade")
	}
}

func TestValidateCatalogConfig(t *testing.T) {
	cases := []struct {
		name    string
		model   *catalogResourceModel
		wantErr string // substring; "" means expect no error
	}{
		{
			name: "postgres ok",
			model: &catalogResourceModel{
				Engine:         types.StringValue("postgres"),
				PostgresConfig: &postgresConfigModel{Host: types.StringValue("h"), Port: types.Int64Value(5432), Database: types.StringValue("db")},
			},
		},
		{
			name: "redshift reuses postgres_config ok",
			model: &catalogResourceModel{
				Engine:         types.StringValue("redshift"),
				PostgresConfig: &postgresConfigModel{Host: types.StringValue("h"), Port: types.Int64Value(5439), Database: types.StringValue("db")},
			},
		},
		{
			name: "wrong block for engine",
			model: &catalogResourceModel{
				Engine:      types.StringValue("postgres"),
				MysqlConfig: &mysqlConfigModel{Host: types.StringValue("h")},
			},
			wantErr: "mysql_config is not valid for engine \"postgres\"; use postgres_config",
		},
		{
			name: "missing required block",
			model: &catalogResourceModel{
				Engine: types.StringValue("snowflake"),
			},
			wantErr: "engine \"snowflake\" requires a snowflake_config block",
		},
		{
			// FIXME: mirrors engineBlocks — bigquery_config is required even though nothing inside
			// it is yet (credentials JSON is not settable over REST). An empty block satisfies it.
			name: "bigquery requires a block",
			model: &catalogResourceModel{
				Engine: types.StringValue("bigquery"),
			},
			wantErr: "engine \"bigquery\" requires a bigquery_config block",
		},
		{
			name: "bigquery ok with empty block",
			model: &catalogResourceModel{
				Engine:         types.StringValue("bigquery"),
				BigQueryConfig: &bigQueryConfigModel{},
			},
		},
		{
			name: "altertable rejects a connection block",
			model: &catalogResourceModel{
				Engine:         types.StringValue("altertable"),
				PostgresConfig: &postgresConfigModel{Host: types.StringValue("h")},
			},
			wantErr: "postgres_config not allowed when engine = \"altertable\"",
		},
		{
			name: "altertable ok with no block",
			model: &catalogResourceModel{
				Engine:   types.StringValue("altertable"),
				BucketID: types.StringValue("b1"),
			},
		},
		{
			name: "connection rejects database-only fields",
			model: &catalogResourceModel{
				Engine:         types.StringValue("postgres"),
				PostgresConfig: &postgresConfigModel{Host: types.StringValue("h"), Port: types.Int64Value(5432), Database: types.StringValue("db")},
				BucketID:       types.StringValue("b1"),
			},
			wantErr: "only valid when engine = \"altertable\"",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Unset pointer blocks default to nil and unset types.String/types.Int64 fields to
			// null, which the block-presence and database-only checks expect.
			errs := validateCatalogConfig(tc.model)
			joined := strings.Join(errs, "\n")
			if tc.wantErr == "" {
				if len(errs) != 0 {
					t.Errorf("expected no errors, got: %s", joined)
				}
				return
			}
			if !strings.Contains(joined, tc.wantErr) {
				t.Errorf("errors = %q, want substring %q", joined, tc.wantErr)
			}
		})
	}
}

func TestValidTablesJSON(t *testing.T) {
	if err := validTablesJSON(types.StringNull()); err != nil {
		t.Errorf("null should pass, got %v", err)
	}
	if err := validTablesJSON(types.StringValue(`{"orders":"s3://x"}`)); err != nil {
		t.Errorf("valid object should pass, got %v", err)
	}
	if err := validTablesJSON(types.StringValue(`["not","an","object"]`)); err == nil {
		t.Error("a JSON array should fail (must be a string->string object)")
	}
	if err := validTablesJSON(types.StringValue(`not json`)); err == nil {
		t.Error("invalid JSON should fail")
	}
}

// The per-engine blocks encode required fields and write-only secrets directly in the schema:
// required fields are enforced by Terraform when the block is present, and secrets never land in
// state. mysql_config must not carry sslmode (postgres-only).
func TestCatalogSchemaConnectionBlocks(t *testing.T) {
	var resp resource.SchemaResponse
	NewCatalogResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %+v", resp.Diagnostics)
	}

	block := func(name string) map[string]schema.Attribute {
		attr, ok := resp.Schema.Attributes[name].(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("%s is not a SingleNestedAttribute", name)
		}
		return attr.Attributes
	}

	pg := block("postgres_config")
	if !pg["host"].IsRequired() {
		t.Error("postgres_config.host should be required")
	}
	if pw, ok := pg["password"].(schema.StringAttribute); !ok || !pw.WriteOnly {
		t.Error("postgres_config.password should be write-only")
	}
	if _, ok := block("mysql_config")["sslmode"]; ok {
		t.Error("mysql_config must not have sslmode (postgres-only field)")
	}
	if sk, ok := block("s3_tables_config")["aws_secret_access_key"].(schema.StringAttribute); !ok || !sk.WriteOnly {
		t.Error("s3_tables_config.aws_secret_access_key should be write-only")
	}
	if tk, ok := block("r2_catalog_config")["token"].(schema.StringAttribute); !ok || !tk.WriteOnly {
		t.Error("r2_catalog_config.token should be write-only")
	}
}

// fullyPopulatedModel returns a model with every connection block set, for coverage checks.
func fullyPopulatedModel() *catalogResourceModel {
	return &catalogResourceModel{
		PostgresConfig:      &postgresConfigModel{},
		MysqlConfig:         &mysqlConfigModel{},
		SnowflakeConfig:     &snowflakeConfigModel{},
		BigQueryConfig:      &bigQueryConfigModel{},
		BucketTablesConfig:  &bucketTablesConfigModel{},
		IcebergTablesConfig: &icebergTablesConfigModel{},
		DuckDBConfig:        &duckDBConfigModel{},
		R2CatalogConfig:     &r2CatalogConfigModel{},
		S3TablesConfig:      &s3TablesConfigModel{},
		GlueConfig:          &glueConfigModel{},
	}
}
