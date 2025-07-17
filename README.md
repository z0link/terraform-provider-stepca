# Terraform Provider StepCA

This provider aims to expose [step-ca](https://github.com/smallstep/certificates) CLI operations as declarative Terraform resources. It currently offers a simple resource for signing certificates using the `/sign` API endpoint, but will expand to cover more of step-ca's functionality.

## Project goals

The long-term objective is to manage step-ca configuration statefully through Terraform. Planned capabilities include:

- Creating and managing private keys
- Managing provisioners
- Creating, renewing, and revoking certificates and CSRs
- Generating `defaults.json` and `ca.json` from resource and data references
- Additional step-ca CLI features where it makes sense to manage them declaratively

## Requirements

* [Go](https://go.dev/) 1.24 or newer must be installed and in your `PATH`.
* The `terraform` CLI is only required when publishing to the Terraform registry.

## Building

```
make build
```

## Testing

```
make test
```

## Releasing

A helper script is provided for publishing the provider to the Terraform registry when it is ready. The actual release is disabled by default and requires setting `PUBLISH_TO_TERRAFORM_REGISTRY=true`.

```
PUBLISH_TO_TERRAFORM_REGISTRY=true make release VERSION=v0.1.0
```

## Example Usage

```
provider "stepca" {
  ca_url = "https://ca.example.com"
  token  = "<one-time-token>"
}

resource "stepca_certificate" "example" {
  csr = file("example.csr")
}
```

The resulting certificate will be available as the `certificate` attribute.

### Data Sources

Query the CA version:

```hcl
data "stepca_version" "current" {}
```

The version string will be available as `data.stepca_version.current.version`.

## Test Releases

Every push to `main` publishes a prerelease on GitHub using the latest commit
hash as the version. The packaged provider binary is uploaded as
`terraform-provider-stepca_<commit>_linux_amd64.zip`. The provider itself is
built with a pseudo-semver version in the form `0.0.0-<commit>` so Terraform can
process version constraints during testing.

To use a test build from this repository specify the pseudo-semver version as
the provider version:

```hcl
terraform {
  required_providers {
    stepca = {
      source  = "github.com/z0link/terraform-provider-stepca"
      version = "0.0.0-<commit>"
    }
  }
}
```

Replace `<commit>` with the hash shown on the GitHub releases page.
