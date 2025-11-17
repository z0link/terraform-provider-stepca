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
* `force_rotate` - (Optional) Toggle this boolean value to force Terraform to request a fresh certificate without changing the CSR. The value itself is persisted in state so flipping it between `true` and `false` will trigger a new issuance.

## Attributes Reference

* `certificate` - The PEM encoded signed certificate returned by the CA.

## Behavior

`stepca_certificate` stores the issued certificate in state so it can be
referenced elsewhere. When the CSR changes or the `force_rotate` flag is
toggled, the provider sends the CSR to `/sign` again and overwrites the stored
certificate. The provider also re-reads the certificate by serial number when
possible and removes it from state if the CA reports it has been revoked or
replaced. Running `terraform destroy` deletes the resource from state only; you
must revoke the certificate manually if necessary.
