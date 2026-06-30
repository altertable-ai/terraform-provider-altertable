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

## Local development (run without a registry)

```bash
make install   # builds and installs the binary into $(go env GOPATH)/bin
```

Add a `~/.terraformrc` (or `$XDG_CONFIG_HOME/terraform.rc`) pointing Terraform at the
local build so `terraform plan` skips `terraform init` downloads:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/altertable-ai/altertable" = "REPLACE_WITH_OUTPUT_OF: go env GOPATH then /bin"
  }
  direct {}
}
```

## Make targets

`build`, `install`, `test`, `testacc`, `lint`, `fmt`, `docs`.
