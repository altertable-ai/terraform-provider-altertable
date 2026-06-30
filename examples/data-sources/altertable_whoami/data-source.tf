# Identify the principal and organization behind the configured API key.
data "altertable_whoami" "current" {}

# Use the organization ID to scope an organization-level role grant without
# hardcoding it.
resource "altertable_role_set" "ci" {
  service_account_id = altertable_service_account.ci.id

  roles = [
    {
      role        = "organization:reader"
      resource_id = data.altertable_whoami.current.organization_id
    },
  ]
}
