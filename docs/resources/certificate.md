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

## Behavior

`stepca_certificate` is a create-only resource. Terraform stores the issued
certificate in state so it can be referenced elsewhere, but it cannot update or
revoke the certificate in the CA. The provider re-reads the certificate by
serial number when possible and removes it from state if the CA reports it has
been revoked or replaced. Running `terraform destroy` deletes the resource from
state only; you must revoke the certificate manually if necessary.
