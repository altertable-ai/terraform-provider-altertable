resource "altertable_credential" "dbt_prod" {
  service_account_id = altertable_service_account.dbt.id
  environment_id     = altertable_environment.production.id
  label              = "dbt Postgres"
}
