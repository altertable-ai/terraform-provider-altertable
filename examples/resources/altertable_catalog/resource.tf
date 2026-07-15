# A native Altertable database catalog
resource "altertable_catalog" "warehouse" {
  environment_id = altertable_environment.example.id
  engine         = "altertable"
  name           = "Warehouse"
}

# An external Postgres connection catalog.
# `password` is write-only: it is sent to the API but never stored in state or shown in the diff.
resource "altertable_catalog" "analytics" {
  environment_id = altertable_environment.example.id
  engine         = "postgres" # also accepts "redshift" and "supabase"
  name           = "Analytics"

  postgres_config = {
    host     = "db.example.com"
    port     = 5432
    database = "analytics"
    username = "altertable"
    password = var.pg_password # write-only
  }
}

# An AWS S3 Tables connection catalog.
resource "altertable_catalog" "lake" {
  environment_id = altertable_environment.example.id
  engine         = "s3tables"
  name           = "Lake"

  s3_tables_config = {
    warehouse             = "arn:aws:s3tables:eu-west-1:123456789012:bucket/my-warehouse"
    default_region        = "eu-west-1"
    aws_access_key_id     = "AKIAEXAMPLE"
    aws_secret_access_key = var.aws_secret_access_key # write-only
  }
}
