resource "altertable_catalog" "analytics" {
  environment_id = altertable_environment.production.id
  slug           = "analytics"
  name           = "Analytics"
}
