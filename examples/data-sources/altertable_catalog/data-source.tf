data "altertable_catalog" "warehouse" {
  environment_id = altertable_environment.production.id
  id             = "warehouse"
}
