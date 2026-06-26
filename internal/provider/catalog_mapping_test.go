package provider

import (
	"testing"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIsDatabaseEngine(t *testing.T) {
	if !isDatabaseEngine("altertable") {
		t.Error("altertable should be a database engine")
	}
	if isDatabaseEngine("postgres") {
		t.Error("postgres should not be a database engine")
	}
}

func TestApplyDatabasePopulatesCommonAndDBFields(t *testing.T) {
	m := &catalogResourceModel{}
	m.applyDatabase(&client.Database{
		ID: "db_1", Name: "Main", Slug: "main", ReadOnly: true, BuiltIn: true,
		Catalog: "cat", BucketID: "b1", SnapshotRetentionDays: 7, EnvironmentID: "env_1",
	})
	if m.ID.ValueString() != "db_1" || m.Slug.ValueString() != "main" {
		t.Errorf("common = %+v", m)
	}
	if m.Engine.ValueString() != "altertable" {
		t.Errorf("engine = %q, want altertable", m.Engine.ValueString())
	}
	if m.BucketID.ValueString() != "b1" || m.SnapshotRetentionDays.ValueInt64() != 7 {
		t.Errorf("db fields = %+v", m)
	}
	if !m.BuiltIn.ValueBool() {
		t.Errorf("built_in = %v, want true", m.BuiltIn.ValueBool())
	}
}

func TestToCreateConnectionRequestMapsPostgresConfig(t *testing.T) {
	m := &catalogResourceModel{
		Name:   types.StringValue("PG"),
		Engine: types.StringValue("postgres"),
		PostgresConfig: &postgresConfigModel{
			Host:     types.StringValue("h"),
			Port:     types.Int64Value(5432),
			Password: types.StringValue("secret"),
		},
	}
	in := m.toCreateConnectionRequest()
	if in.Engine != "postgres" || in.PostgresConfig == nil {
		t.Fatalf("in = %+v", in)
	}
	if in.PostgresConfig.Host != "h" || in.PostgresConfig.Port != 5432 || in.PostgresConfig.Password != "secret" {
		t.Errorf("pg config = %+v", in.PostgresConfig)
	}
}
