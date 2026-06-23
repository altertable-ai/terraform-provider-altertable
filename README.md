# Terraform Provider for Altertable

Manage [Altertable](https://altertable.ai) resources with Terraform.

> **Status:** scaffold. The provider builds and validates, but resource/data-source
> operations return "not implemented" until the Altertable management API and the
> client methods in `internal/client/` are completed.

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
    "registry.terraform.io/altertable/altertable" = "REPLACE_WITH_OUTPUT_OF: go env GOPATH then /bin"
  }
  direct {}
}
```

## Make targets

`build`, `install`, `test`, `testacc`, `lint`, `fmt`, `docs`.
