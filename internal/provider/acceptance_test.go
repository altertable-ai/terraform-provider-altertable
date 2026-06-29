package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
	const config = `resource "altertable_environment" "test" {
  name           = "Acc Test"
  cloud_provider = "aws"
}

resource "altertable_catalog" "test" {
  environment_id = altertable_environment.test.id
  engine         = "altertable"
  name           = "Acc Test Catalog"
}`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttrSet("altertable_catalog.test", "id"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentityValueMatchesState("altertable_catalog.test", tfjsonpath.New("environment_id")),
					statecheck.ExpectIdentityValueMatchesState("altertable_catalog.test", tfjsonpath.New("id")),
				},
			},
			{
				// Back-compat import via the "environment_id:id" colon string.
				ResourceName:      "altertable_catalog.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["altertable_catalog.test"]
					return rs.Primary.Attributes["environment_id"] + ":" + rs.Primary.Attributes["id"], nil
				},
			},
			{
				// Import via the typed identity block (Terraform 1.12+).
				ResourceName:      "altertable_catalog.test",
				ImportState:       true,
				ImportStateKind:   resource.ImportBlockWithResourceIdentity,
				ImportStateVerify: true,
			},
		},
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
	const config = `resource "altertable_environment" "test" {
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
}`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("altertable_credential.test", "id"),
					resource.TestCheckResourceAttrSet("altertable_credential.test", "password"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentityValueMatchesState("altertable_credential.test", tfjsonpath.New("principal_type")),
					statecheck.ExpectIdentityValueMatchesState("altertable_credential.test", tfjsonpath.New("principal_id")),
					statecheck.ExpectIdentityValueMatchesState("altertable_credential.test", tfjsonpath.New("environment_id")),
					statecheck.ExpectIdentityValueMatchesState("altertable_credential.test", tfjsonpath.New("id")),
				},
			},
			{
				ResourceName:            "altertable_credential.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					a := s.RootModule().Resources["altertable_credential.test"].Primary.Attributes
					return a["principal_type"] + ":" + a["principal_id"] + ":" + a["environment_id"] + ":" + a["id"], nil
				},
			},
			{
				ResourceName:            "altertable_credential.test",
				ImportState:             true,
				ImportStateKind:         resource.ImportBlockWithResourceIdentity,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
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
