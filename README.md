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

Every push to `main` publishes a prerelease on GitHub using a version string
that combines the installed `step` CLI version and the commit tag. If the commit
is not tagged the first six characters of the commit hash are used. The packaged
provider binary is uploaded as
`terraform-provider-stepca_stepca-<step-version>-<commit>_linux_amd64.zip`.

Terraform only accepts semantic version numbers. After extracting a prerelease
binary, rename it to a valid version before running `terraform init`. One option
is to rename the file to `terraform-provider-stepca_v0.0.0` and reference that
version in your configuration:

```bash
unzip terraform-provider-stepca_stepca-<step-version>-<commit>_linux_amd64.zip
mv terraform-provider-stepca terraform-provider-stepca_v0.0.0
```

```hcl
terraform {
  required_providers {
    stepca = {
      source  = "local/stepca" # provider binary declares this address
      version = "0.0.0"
    }
  }
}
```

Replace `<step-version>` with the CLI version embedded in the filename and
`<commit>` with the commit hash shown on the GitHub releases page.
