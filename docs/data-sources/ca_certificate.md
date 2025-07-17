# stepca_ca_certificate

Fetches the root certificate from the step-ca `/root` endpoint.

## Example Usage

```hcl
data "stepca_ca_certificate" "root" {}
```

## Attributes Reference

* `certificate` - PEM encoded root certificate returned by the CA.
