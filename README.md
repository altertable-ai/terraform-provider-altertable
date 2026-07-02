# Terraform Provider for Altertable

Manage [Altertable](https://altertable.ai) resources with Terraform.

> **Status:** The provider implements environments, service accounts, catalogs
> (databases and connections), credentials, and role sets against the Altertable
> Management REST API. Users are managed outside Terraform — there is no
> `altertable_user` resource, but the `altertable_user` and `altertable_whoami`
> data sources look up a user by email and the configured key's own identity.

## Requirements

- Terraform >= 1.0 (or OpenTofu)
- Go >= 1.25 (to build)

## Provider configuration

```hcl
provider "altertable" {
  api_key = var.altertable_management_api_key # or env ALTERTABLE_API_KEY
}
```

## Development

Building locally, running the tests, and cutting releases are covered in
[DEVELOPMENT.md](DEVELOPMENT.md).
