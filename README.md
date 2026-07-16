# Terraform Provider for Altertable

Manage [Altertable](https://altertable.ai) resources with Terraform.

> **Status:** The provider implements environments, service accounts, catalogs
> (databases and connections), credentials, role sets, and users against the
> Altertable Management REST API.

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
