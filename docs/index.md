# StepCA Provider

Use the **stepca** provider to interact with a [step-ca](https://github.com/smallstep/certificates) Certificate Authority.

## Example Usage

```hcl
terraform {
  required_providers {
    stepca = {
      source = "github.com/z0link/stepca"
      version = "0.1.0"
    }
  }
}

provider "stepca" {
  ca_url = "https://ca.example.com"
  token  = "<one-time-token>"
}
```

## Argument Reference

The following arguments are supported:

* `ca_url` - (Required) The base URL of the step-ca instance.
* `token`  - (Required) The one-time bootstrap token used to authenticate.

## Resources

* [`stepca_certificate`](resources/certificate.md) - Sign a CSR and obtain a certificate.

## Data Sources

* [`stepca_version`](data-sources/version.md) - Retrieve the CA version.
* [`stepca_ca_certificate`](data-sources/ca_certificate.md) - Fetch the root certificate.
