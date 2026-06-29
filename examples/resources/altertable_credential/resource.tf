# Lakehouse credential a service account uses to query the environment's catalogs.
resource "altertable_credential" "ci" {
  principal_type = "service_account"
  principal_id   = altertable_service_account.example.id
  environment_id = altertable_environment.example.id
  label          = "CI"
}
