package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/altertable-ai/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUserModelApplyUserPending(t *testing.T) {
	var m userResourceModel
	m.applyUser(&client.User{
		IacID:        "invitation_id=inv_1",
		InvitationID: "inv_1", // present on the wire, deliberately not surfaced in state
		Email:        "bob@example.com",
	})
	if m.ID.ValueString() != "invitation_id=inv_1" {
		t.Errorf("id = %q, want the iac_id", m.ID.ValueString())
	}
	if m.Email.ValueString() != "bob@example.com" {
		t.Errorf("email = %q", m.Email.ValueString())
	}
}

func TestUserModelApplyUserAccepted(t *testing.T) {
	var m userResourceModel
	m.applyUser(&client.User{
		ID:    "u_1",                 // the user UUID now exists, but is not surfaced in state
		IacID: "invitation_id=inv_1", // the iac_id stays stable after acceptance
		Email: "bob@example.com",
		Name:  "Bob", // display name is on the wire but deliberately not surfaced in state
	})
	// The id must remain the stable iac_id even though a distinct user UUID now exists.
	if m.ID.ValueString() != "invitation_id=inv_1" {
		t.Errorf("id = %q, want the stable iac_id even post-acceptance", m.ID.ValueString())
	}
	if m.Email.ValueString() != "bob@example.com" {
		t.Errorf("email = %q", m.Email.ValueString())
	}
}

func TestIacIDForImportPassesThroughID(t *testing.T) {
	r := &UserResource{}
	got, err := r.iacIDForImport(context.Background(), userIdentityModel{
		ID:     types.StringValue("invitation_id=inv_1"),
		UserID: types.StringNull(),
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if got != "invitation_id=inv_1" {
		t.Errorf("iac_id = %q, want the id passed through unchanged", got)
	}
}

func TestIacIDForImportResolvesUserID(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		// The user was invited via Terraform: their iac_id keeps its invitation_id= form even
		// though they are now an accepted member addressable by a raw user UUID.
		_, _ = io.WriteString(w, `{"user":{"id":"u_1","invitation_id":null,"iac_id":"invitation_id=inv_1","email":"bob@example.com","name":"Bob"}}`)
	}))
	defer srv.Close()

	r := &UserResource{client: client.NewClient(srv.URL, "k", "test")}
	got, err := r.iacIDForImport(context.Background(), userIdentityModel{
		ID:     types.StringNull(),
		UserID: types.StringValue("u_1"),
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if gotPath != "/users/u_1" {
		t.Errorf("looked up %q, want GET /users/u_1", gotPath)
	}
	if got != "invitation_id=inv_1" {
		t.Errorf("iac_id = %q, want the resolved canonical iac_id", got)
	}
}

func TestIacIDForImportRejectsAmbiguousIdentity(t *testing.T) {
	r := &UserResource{}
	if _, err := r.iacIDForImport(context.Background(), userIdentityModel{
		ID:     types.StringValue("invitation_id=inv_1"),
		UserID: types.StringValue("u_1"),
	}); err == nil {
		t.Error("want error when both id and user_id are set")
	}
	if _, err := r.iacIDForImport(context.Background(), userIdentityModel{
		ID:     types.StringNull(),
		UserID: types.StringNull(),
	}); err == nil {
		t.Error("want error when neither id nor user_id is set")
	}
}
