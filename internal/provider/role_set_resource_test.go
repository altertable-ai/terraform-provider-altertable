package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToFromClientGrantsRoundTrip(t *testing.T) {
	in := []roleGrantModel{
		{Role: types.StringValue("organization:member"), ResourceID: types.StringNull()},
		{Role: types.StringValue("environment:writer"), ResourceID: types.StringValue("env_1")},
	}
	got := fromClientGrants(toClientGrants(in))
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Role.ValueString() != "organization:member" || !got[0].ResourceID.IsNull() {
		t.Errorf("grant[0] = %+v, want org:member with null resource_id", got[0])
	}
	if got[1].Role.ValueString() != "environment:writer" || got[1].ResourceID.ValueString() != "env_1" {
		t.Errorf("grant[1] = %+v, want environment:writer/env_1", got[1])
	}
}

func TestPrincipalRefSelectsSetField(t *testing.T) {
	m := roleSetResourceModel{
		UserID:           types.StringNull(),
		ServiceAccountID: types.StringValue("sa_1"),
	}
	ref := m.principalRef()
	if ref.ServiceAccountID != "sa_1" || ref.UserID != "" {
		t.Fatalf("ref = %+v", ref)
	}
}
