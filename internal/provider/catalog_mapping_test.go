package provider

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/altertable/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func TestApplyConnectionSetsKnownDBOnlyFields(t *testing.T) {
	m := &catalogResourceModel{
		BuiltIn:               types.BoolUnknown(),
		BucketID:              types.StringUnknown(),
		SnapshotRetentionDays: types.Int64Unknown(),
	}
	m.applyConnection(&client.Connection{ID: "con_1", Name: "PG", Slug: "pg", Engine: "postgres"})
	if m.BuiltIn.IsUnknown() || m.BuiltIn.ValueBool() {
		t.Errorf("built_in = %v, want known false", m.BuiltIn)
	}
	if !m.BucketID.IsNull() {
		t.Errorf("bucket_id = %v, want null", m.BucketID)
	}
	if !m.SnapshotRetentionDays.IsNull() {
		t.Errorf("snapshot_retention_days = %v, want null", m.SnapshotRetentionDays)
	}
	if m.Engine.ValueString() != "postgres" {
		t.Errorf("engine = %q", m.Engine.ValueString())
	}
}

func TestCatalogSchemaDBOnlyFieldsAreComputed(t *testing.T) {
	var resp resource.SchemaResponse
	NewCatalogResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	for _, name := range []string{"bucket_id", "snapshot_retention_days"} {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("missing attribute %s", name)
		}
		if !attr.IsComputed() {
			t.Errorf("%s must be Optional+Computed to avoid inconsistent-result-after-apply on the database path", name)
		}
	}
}

// An explicit snapshot_retention_days = 0 must reach the wire; an unset value must be
// omitted so the server applies its own default. A non-pointer omitempty int could not
// tell these apart, which caused "inconsistent result after apply" when the server
// default was non-zero.
func TestToCreateDatabaseRequestTransmitsExplicitZeroRetention(t *testing.T) {
	explicit := &catalogResourceModel{
		Name:                  types.StringValue("Main"),
		Engine:                types.StringValue(engineAltertable),
		SnapshotRetentionDays: types.Int64Value(0),
	}
	if got := explicit.toCreateDatabaseRequest().SnapshotRetentionDays; got == nil || *got != 0 {
		t.Fatalf("SnapshotRetentionDays = %v, want pointer to 0", got)
	}
	if body, _ := json.Marshal(explicit.toCreateDatabaseRequest()); !strings.Contains(string(body), `"snapshot_retention_days":0`) {
		t.Errorf("body = %s, want it to carry snapshot_retention_days:0", body)
	}

	unset := &catalogResourceModel{
		Name:                  types.StringValue("Main"),
		Engine:                types.StringValue(engineAltertable),
		SnapshotRetentionDays: types.Int64Null(),
	}
	if got := unset.toCreateDatabaseRequest().SnapshotRetentionDays; got != nil {
		t.Errorf("SnapshotRetentionDays = %v, want nil (omitted)", *got)
	}
	if body, _ := json.Marshal(unset.toCreateDatabaseRequest()); strings.Contains(string(body), "snapshot_retention_days") {
		t.Errorf("body = %s, want snapshot_retention_days omitted", body)
	}
}

// An empty description from the server must stay "" (not collapse to null) for the
// user-settable Optional+Computed attribute, so an explicit description = "" round-trips.
func TestApplyDatabaseKeepsEmptyDescriptionKnown(t *testing.T) {
	m := &catalogResourceModel{Description: types.StringValue("")}
	m.applyDatabase(&client.Database{ID: "db_1", Description: ""})
	if m.Description.IsNull() || m.Description.ValueString() != "" {
		t.Errorf("description = %v, want known empty string", m.Description)
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
