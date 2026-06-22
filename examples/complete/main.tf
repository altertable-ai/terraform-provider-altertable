provider "altertable" {
  api_key = var.altertable_management_api_key
}

# ── Data sources ─────────────────────────────

data "altertable_environment" "production" {
  slug = "production"
}

data "altertable_environment" "staging" {
  slug = "staging"
}

data "altertable_catalog" "analytics" {
  environment_id = data.altertable_environment.production.id
  slug           = "analytics"
}

data "altertable_user" "alice" {
  email = "alice@acme.com"
}

# ── Service account ─────────────────────────

resource "altertable_service_account" "dbt" {
  name = "dbt Cloud"
}

# ── Roles ───────────────────────────────────

resource "altertable_role_set" "dbt" {
  service_account_id = altertable_service_account.dbt.id

  roles = [
    { role = "organization:member" },
    { role = "environment:writer", resource_id = data.altertable_environment.production.id },
    { role = "environment:reader", resource_id = data.altertable_environment.staging.id },
  ]
}

resource "altertable_role_set" "alice" {
  user_id = data.altertable_user.alice.id

  roles = [
    { role = "organization:member" },
    { role = "environment:member", resource_id = data.altertable_environment.production.id },
    { role = "catalog:writer", resource_id = data.altertable_catalog.analytics.id },
  ]
}

# ── Credentials ─────────────────────────────

resource "altertable_credential" "dbt_prod" {
  service_account_id = altertable_service_account.dbt.id
  environment_id     = data.altertable_environment.production.id
  label              = "dbt Postgres"
}

output "dbt_password" {
  value     = altertable_credential.dbt_prod.password
  sensitive = true
}
