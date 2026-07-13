provider "altertable" {
  api_key = var.altertable_management_api_key
}

# ── Environment ───────────────────────────────

resource "altertable_environment" "production" {
  name                  = "Production"
  cloud_provider        = "aws"
  cloud_provider_region = "eu-west-1"
}

# ── Service account ──────────────────────────

resource "altertable_service_account" "ci" {
  label = "CI Deploy"
}

# ── Catalogs ─────────────────────────────────

resource "altertable_catalog" "warehouse" {
  environment_id = altertable_environment.production.id
  engine         = "altertable"
  name           = "Warehouse"
}

resource "altertable_catalog" "analytics" {
  environment_id = altertable_environment.production.id
  engine         = "postgres"
  name           = "Analytics"

  postgres_config = {
    host     = "db.example.com"
    port     = 5432
    database = "analytics"
    username = "altertable"
    password = var.pg_password # write-only
  }
}

# ── Lakehouse credentials ──────────────────────────────

resource "altertable_credential" "ci" {
  principal_type = "service_account"
  principal_id   = altertable_service_account.ci.id
  environment_id = altertable_environment.production.id
  label          = "CI"
}

output "ci_password" {
  value     = altertable_credential.ci.password
  sensitive = true
}
