# Terraform Provider StepCA

This is an experimental Terraform provider for interacting with a self-hosted [step-ca](https://github.com/smallstep/certificates) instance. It exposes a simple resource for signing certificates using the `/sign` API endpoint.

## Building

```
go build
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
