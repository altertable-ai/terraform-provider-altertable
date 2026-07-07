# Look up by slug…
data "altertable_bucket" "landing" {
  environment_id = altertable_environment.production.id
  slug           = "BKT-1"
}

# …or by ID
data "altertable_bucket" "archive" {
  environment_id = altertable_environment.production.id
  id             = "3b9f2e1d-8c7a-4f5b-a6d3-1e0c9b8a7f6e"
}
