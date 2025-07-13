# Terraform Provider StepCA

This is an experimental Terraform provider for interacting with a self-hosted [step-ca](https://github.com/smallstep/certificates) instance. It exposes a simple resource for signing certificates using the `/sign` API endpoint.

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

## Test Releases

Every push to `main` publishes a prerelease on GitHub using the latest commit
hash as the version. The packaged provider binary is uploaded as
`terraform-provider-stepca_<commit>_linux_amd64.zip`.

To use a test build from this repository specify the commit hash as the provider
version:

```hcl
terraform {
  required_providers {
    stepca = {
      source  = "github.com/z0link/terraform-provider-stepca"
      version = "<commit>"
    }
  }
}
```

Replace `<commit>` with the hash shown on the GitHub releases page.
