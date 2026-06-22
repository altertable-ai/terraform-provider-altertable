resource "altertable_role_set" "dbt" {
  service_account_id = altertable_service_account.dbt.id

  roles = [
    { role = "organization:member" },
    { role = "environment:writer", resource_id = altertable_environment.production.id },
  ]
}
