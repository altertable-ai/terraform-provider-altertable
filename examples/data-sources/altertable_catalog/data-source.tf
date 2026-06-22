data "altertable_catalog" "analytics" {
  environment_id = data.altertable_environment.production.id
  slug           = "analytics"
}
