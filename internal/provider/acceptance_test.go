package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/altertable-ai/terraform-provider-altertable/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// These acceptance tests run only when TF_ACC is set and the Altertable API plus the
// client methods in internal/client are implemented. resource.Test skips otherwise.

func testAccClient() *client.Client {
	return client.NewClient(os.Getenv("ALTERTABLE_API_URL"), os.Getenv("ALTERTABLE_API_KEY"), "test")
}

// checkDestroyed asserts every resource of resourceType is gone from the API after destroy.
// Deletes are synchronous, so one GET suffices: a hit means the resource leaked.
func checkDestroyed(resourceType string, probe func(attrs map[string]string) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			switch err := probe(rs.Primary.Attributes); {
			case err == nil:
				return fmt.Errorf("%s still exists on the API after destroy", name)
			case !isNotFound(err):
				return fmt.Errorf("%s: unexpected error verifying destroy: %w", name, err)
			}
		}
		return nil
	}
}

func TestAccEnvironmentResource_basic(t *testing.T) {
	const config = `resource "altertable_environment" "test" {
  name                  = "Terraform Acceptance Test"
  cloud_provider        = "hetzner"
  cloud_provider_region = "fsn1"
}`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: checkDestroyed("altertable_environment", func(a map[string]string) error {
			_, err := testAccClient().GetEnvironment(context.Background(), a["id"])
			return err
		}),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("altertable_environment.test", "id"),
					resource.TestCheckResourceAttr("altertable_environment.test", "name", "Terraform Acceptance Test"),
					resource.TestCheckResourceAttr("altertable_environment.test", "cloud_provider", "hetzner"),
					resource.TestCheckResourceAttr("altertable_environment.test", "cloud_provider_region", "fsn1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentityValueMatchesState("altertable_environment.test", tfjsonpath.New("id")),
				},
			},
			{
				ResourceName:    "altertable_environment.test",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func TestAccCatalogResource_basic(t *testing.T) {
	const config = `resource "altertable_environment" "test" {
  name                  = "Terraform Acceptance Test"
  cloud_provider        = "hetzner"
  cloud_provider_region = "fsn1"
}

resource "altertable_catalog" "test" {
  environment_id = altertable_environment.test.id
  engine         = "altertable"
  name           = "Terraform Acceptance Test Catalog"
}`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: checkDestroyed("altertable_catalog", func(a map[string]string) error {
			c := testAccClient()
			if isDatabaseEngine(a["engine"]) {
				_, err := c.GetDatabase(context.Background(), a["environment_id"], a["id"])
				return err
			}
			_, err := c.GetConnection(context.Background(), a["environment_id"], a["id"])
			return err
		}),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("altertable_catalog.test", "id"),
					resource.TestCheckResourceAttrPair("altertable_catalog.test", "environment_id", "altertable_environment.test", "id"),
					resource.TestCheckResourceAttr("altertable_catalog.test", "engine", "altertable"),
					resource.TestCheckResourceAttr("altertable_catalog.test", "name", "Terraform Acceptance Test Catalog"),
				),
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
				ResourceName:    "altertable_catalog.test",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func TestAccServiceAccountResource_basic(t *testing.T) {
	const config = `resource "altertable_service_account" "test" {
  label = "Terraform Acceptance Test Service Account"
}`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: checkDestroyed("altertable_service_account", func(a map[string]string) error {
			_, err := testAccClient().GetServiceAccount(context.Background(), a["id"])
			return err
		}),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("altertable_service_account.test", "id"),
					resource.TestCheckResourceAttr("altertable_service_account.test", "label", "Terraform Acceptance Test Service Account"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentityValueMatchesState("altertable_service_account.test", tfjsonpath.New("id")),
				},
			},
			{
				ResourceName:    "altertable_service_account.test",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func TestAccRoleSetResource_basic(t *testing.T) {
	const withRoleSet = `resource "altertable_service_account" "test" {
  label = "Terraform Acceptance Test Service Account"
}

data "altertable_whoami" "current" {}

resource "altertable_role_set" "test" {
  service_account_id = altertable_service_account.test.id

  roles = [
    { role = "organization:member", resource_id = data.altertable_whoami.current.organization_id },
  ]
}`

	// role_set.Delete errors by design (no delete endpoint; erroring keeps stale access
	// in-your-face). The harness auto-destroy would hit that error and leak the service
	// account, so forget role_set from state first with a `removed` block. Destroying the
	// service account then clears its role assignments on the backend — no orphans.
	const forgetRoleSet = `resource "altertable_service_account" "test" {
  label = "Terraform Acceptance Test Service Account"
}

removed {
  from = altertable_role_set.test
  lifecycle {
    destroy = false
  }
}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: checkDestroyed("altertable_service_account", func(a map[string]string) error {
			_, err := testAccClient().GetServiceAccount(context.Background(), a["id"])
			return err
		}),
		Steps: []resource.TestStep{
			{
				Config: withRoleSet,
				Check:  resource.TestCheckResourceAttrSet("altertable_role_set.test", "id"),
			},
			{Config: forgetRoleSet}, // drops role_set from state → clean auto-destroy
		},
	})
}

func TestAccCredentialResource_basic(t *testing.T) {
	const config = `resource "altertable_environment" "test" {
  name                  = "Terraform Acceptance Test"
  cloud_provider        = "hetzner"
  cloud_provider_region = "fsn1"
}

resource "altertable_service_account" "test" {
  label = "Terraform Acceptance Test Service Account"

  # Force the environment to be created before this service account. Creating an environment
  # back-fills a default (unrevokable) credential onto every service account that already exists,
  # and that credential blocks the SA's deletion. Ordering the env first — while this SA does not
  # yet exist — avoids it. Without this, env and SA are created concurrently (no dependency edge).
  depends_on = [altertable_environment.test]
}

resource "altertable_credential" "test" {
  principal_type = "service_account"
  principal_id   = altertable_service_account.test.id
  environment_id = altertable_environment.test.id
  label          = "Terraform Acceptance Test Credential"
}`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: checkDestroyed("altertable_credential", func(a map[string]string) error {
			_, err := testAccClient().GetCredential(context.Background(),
				a["principal_type"], a["principal_id"], a["environment_id"], a["id"])
			return err
		}),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("altertable_credential.test", "id"),
					resource.TestCheckResourceAttrSet("altertable_credential.test", "password"),
					resource.TestCheckResourceAttr("altertable_credential.test", "label", "Terraform Acceptance Test Credential"),
					resource.TestCheckResourceAttr("altertable_credential.test", "principal_type", "service_account"),
					resource.TestCheckResourceAttrPair("altertable_credential.test", "principal_id", "altertable_service_account.test", "id"),
					resource.TestCheckResourceAttrPair("altertable_credential.test", "environment_id", "altertable_environment.test", "id"),
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
				ResourceName:    "altertable_credential.test",
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func TestAccWhoamiDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "altertable_whoami" "current" {}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet("data.altertable_whoami.current", "organization_id"),
				resource.TestCheckResourceAttrSet("data.altertable_whoami.current", "principal_id"),
				resource.TestCheckResourceAttrSet("data.altertable_whoami.current", "principal_type"),
			),
		}},
	})
}

func TestAccEnvironmentDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "altertable_environment" "seed" {
  name                  = "Terraform Acceptance Test Environment Data Source"
  cloud_provider        = "hetzner"
  cloud_provider_region = "fsn1"
}

data "altertable_environment" "test" {
  slug = altertable_environment.seed.slug
}`,
			Check: resource.TestCheckResourceAttrPair(
				"data.altertable_environment.test", "id",
				"altertable_environment.seed", "id"),
		}},
	})
}

func TestAccServiceAccountDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `resource "altertable_service_account" "seed" {
  label = "Terraform Acceptance Test Service Account Data Source"
}

data "altertable_service_account" "test" {
  id = altertable_service_account.seed.id
}`,
			Check: resource.TestCheckResourceAttrSet("data.altertable_service_account.test", "label"),
		}},
	})
}
