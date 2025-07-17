# stepca_certificate

Signs a certificate signing request (CSR) using the step-ca `/sign` API and returns the issued certificate.

## Example Usage

```hcl
resource "stepca_certificate" "example" {
  csr = file("example.csr")
}
```

## Argument Reference

* `csr` - (Required) The PEM encoded certificate signing request.

## Attributes Reference

* `certificate` - The PEM encoded signed certificate returned by the CA.
