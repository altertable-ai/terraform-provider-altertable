# Development

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

`build`, `install`, `test`, `testacc`, `lint`, `fmt`, `docs`, `release`.

`test` runs the unit tests. `testacc` runs the acceptance tests, which create and
destroy real resources and therefore need `ALTERTABLE_API_KEY` and a non-production
`ALTERTABLE_API_URL` (they refuse to run against production).

## Releasing

Releases are cut with `make release`, which validates the version, runs the full
test suite (including acceptance tests against staging), then tags the release and
triggers the [Release workflow](.github/workflows/release.yml).

```bash
ALTERTABLE_STAGING_TEST_API_KEY=<staging key> make release
```

Prerequisites:

- `ALTERTABLE_STAGING_TEST_API_KEY` set to a staging API key (used to run the
  acceptance tests against staging).
- The [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated
  (`gh auth login`) — the release workflow is triggered through it.
- A clean working tree (the script refuses to release uncommitted changes).

What it does:

1. Prompts for a version, showing the latest release for reference.
2. Validates the version is `vMAJOR.MINOR.PATCH`, does not already exist, and — if
   there is a previous release — is strictly greater than it. The first release
   accepts any valid version.
3. Runs `go build`, the unit tests, and the acceptance tests against staging.
4. On confirmation, creates an annotated tag, pushes it, and triggers the release
   workflow (goreleaser builds and signs the artifacts from the tag).
