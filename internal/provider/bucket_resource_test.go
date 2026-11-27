package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// validateBucketConfig runs the provider's full config validation (attribute and
// resource-level validators) for an altertable_bucket with the given string attributes set.
func validateBucketConfig(t *testing.T, attrs map[string]string) []*tfprotov6.Diagnostic {
	t.Helper()
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	NewBucketResource().Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)

	vals := map[string]tftypes.Value{}
	for name, typ := range objType.AttributeTypes {
		if v, ok := attrs[name]; ok {
			vals[name] = tftypes.NewValue(typ, v)
		} else {
			vals[name] = tftypes.NewValue(typ, nil)
		}
	}
	config, err := tfprotov6.NewDynamicValue(objType, tftypes.NewValue(objType, vals))
	if err != nil {
		t.Fatalf("build config: %s", err)
	}
	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatalf("create server: %s", err)
	}
	resp, err := server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{
		TypeName: "altertable_bucket",
		Config:   &config,
	})
	if err != nil {
		t.Fatalf("ValidateResourceConfig: %s", err)
	}
	return resp.Diagnostics
}

func TestBucketResourceValidateConfig(t *testing.T) {
	const roleARN = "arn:aws:iam::123456789012:role/altertable"
	base := func(extra map[string]string) map[string]string {
		attrs := map[string]string{"environment_id": "env_1", "name": "landing"}
		for k, v := range extra {
			attrs[k] = v
		}
		return attrs
	}
	cases := []struct {
		name    string
		attrs   map[string]string
		wantErr string // substring of an error detail; "" means expect no error
	}{
		{name: "access keys", attrs: base(map[string]string{"access_key_id": "ak", "secret_access_key": "sk"})},
		{name: "role", attrs: base(map[string]string{"role_arn": roleARN, "region": "eu-west-1"})},
		{name: "govcloud role", attrs: base(map[string]string{"role_arn": "arn:aws-us-gov:iam::123456789012:role/path/altertable", "region": "us-gov-west-1"})},
		{name: "neither auth method", attrs: base(nil), wantErr: "Exactly one of these attributes must be configured: [access_key_id,role_arn]"},
		{
			name:    "both auth methods",
			attrs:   base(map[string]string{"access_key_id": "ak", "secret_access_key": "sk", "role_arn": roleARN, "region": "eu-west-1"}),
			wantErr: "Exactly one of these attributes must be configured: [access_key_id,role_arn]",
		},
		{name: "access key without secret", attrs: base(map[string]string{"access_key_id": "ak"}), wantErr: "must be configured together: [access_key_id,secret_access_key]"},
		{name: "secret without access key", attrs: base(map[string]string{"secret_access_key": "sk"}), wantErr: "must be configured together: [access_key_id,secret_access_key]"},
		{name: "malformed role ARN", attrs: base(map[string]string{"role_arn": "arn:aws:iam::123:user/bob", "region": "eu-west-1"}), wantErr: "must be an IAM role ARN"},
		{name: "role without region", attrs: base(map[string]string{"role_arn": roleARN}), wantErr: `"region" must be specified when "role_arn" is specified`},
		{
			name:    "role with endpoint",
			attrs:   base(map[string]string{"role_arn": roleARN, "region": "auto", "endpoint": "https://acct.r2.cloudflarestorage.com"}),
			wantErr: `"endpoint" cannot be specified when "role_arn" is specified`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var errs []string
			for _, d := range validateBucketConfig(t, tc.attrs) {
				if d.Severity == tfprotov6.DiagnosticSeverityError {
					errs = append(errs, d.Summary+": "+d.Detail)
				}
			}
			if tc.wantErr == "" {
				if len(errs) > 0 {
					t.Fatalf("unexpected errors: %v", errs)
				}
				return
			}
			for _, e := range errs {
				if strings.Contains(e, tc.wantErr) {
					return
				}
			}
			t.Fatalf("expected an error mentioning %q, got %v", tc.wantErr, errs)
		})
	}
}
