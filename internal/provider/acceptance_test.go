package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// These acceptance tests run only when TF_ACC is set and the Altertable API plus the
// client methods in internal/client are implemented. resource.Test skips otherwise.

func TestAccEnvironmentResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `resource "altertable_environment" "test" {
  name           = "Acc Test"
  cloud_provider = "aws"
}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("altertable_environment.test", "id"),
				resource.TestCheckResourceAttr("altertable_environment.test", "cloud_provider", "aws"),
			),
		}},
	})
}

func TestAccCatalogResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `resource "altertable_environment" "test" {
  name           = "Acc Test"
  cloud_provider = "aws"
}

resource "altertable_catalog" "test" {
  environment_id = altertable_environment.test.id
  engine         = "altertable"
  name           = "Acc Test Catalog"
}`,
			Check: resource.TestCheckResourceAttrSet("altertable_catalog.test", "id"),
		}},
	})
}

func TestAccUserResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `resource "altertable_user" "test" {
  email = "acctest@example.com"
  name  = "Acc Test"
}`,
			Check: resource.TestCheckResourceAttrSet("altertable_user.test", "id"),
		}},
	})
}

func TestAccServiceAccountResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `resource "altertable_service_account" "test" {
  label = "acc-test-sa"
}`,
			Check: resource.TestCheckResourceAttrSet("altertable_service_account.test", "id"),
		}},
	})
}

func TestAccRoleSetResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `resource "altertable_service_account" "test" {
  label = "acc-test-sa"
}

resource "altertable_role_set" "test" {
  service_account_id = altertable_service_account.test.id

  roles = [
    { role = "organization:member" },
  ]
}`,
			Check: resource.TestCheckResourceAttrSet("altertable_role_set.test", "id"),
		}},
	})
}

func TestAccCredentialResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `resource "altertable_environment" "test" {
  name           = "Acc Test"
  cloud_provider = "aws"
}

resource "altertable_service_account" "test" {
  label = "acc-test-sa"
}

resource "altertable_credential" "test" {
  principal_type = "service_account"
  principal_id   = altertable_service_account.test.id
  environment_id = altertable_environment.test.id
  label          = "acc-test"
}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("altertable_credential.test", "id"),
				resource.TestCheckResourceAttrSet("altertable_credential.test", "password"),
			),
		}},
	})
}

func TestAccEnvironmentDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "altertable_environment" "test" {
  slug = "production"
}`,
			Check: resource.TestCheckResourceAttrSet("data.altertable_environment.test", "id"),
		}},
	})
}

func TestAccServiceAccountDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `resource "altertable_service_account" "seed" {
  label = "acc-ds-sa"
}

data "altertable_service_account" "test" {
  id = altertable_service_account.seed.id
}`,
			Check: resource.TestCheckResourceAttrSet("data.altertable_service_account.test", "label"),
		}},
	})
}
