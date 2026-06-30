# A native Altertable database catalog
resource "altertable_catalog" "warehouse" {
  environment_id = altertable_environment.example.id
  engine         = "altertable"
  name           = "Warehouse"
}

# An external Postgres connection catalog
resource "altertable_catalog" "analytics" {
  environment_id = altertable_environment.example.id
  engine         = "postgres"
  name           = "Analytics"

  postgres_config = {
    host     = "db.example.com"
    port     = 5432
    database = "analytics"
    username = "altertable"
    password = var.pg_password
  }
}
