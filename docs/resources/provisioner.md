# stepca_provisioner

Manages a provisioner through the step-ca admin API.

## Example Usage

```hcl
resource "stepca_provisioner" "admin" {
  name  = "admin"
  type  = "JWK"
  admin = true
}

resource "stepca_provisioner" "leaf" {
  name           = "leaf-issuer"
  type           = "JWK"
  x509_template  = "leaf"
  ssh_template   = "ssh-user"
  admin          = false
}
```

The CA created by `step ca init` includes a default JWK admin provisioner. To
create additional provisioners you must supply an admin token from an existing
admin provisioner via the provider's `admin_token` argument. Generate the token
using `step ca admin` or `step ca token --issuer <provisioner>` with the
corresponding admin key.

Provisioners marked as `admin = true` can create additional admins with the
`stepca_admin` resource.

## Argument Reference

* `name` - (Required) Name of the provisioner.
* `type` - (Required) Provisioner type, e.g. `JWK`, `OIDC`, `ACME`, `X5C`.
* `admin` - (Optional) Set to `true` to create an admin provisioner.
* `x509_template` - (Optional) Name of an X.509 template to bind to the provisioner (maps to step-ca's `x509Template`).
* `ssh_template` - (Optional) Name of an SSH template to bind to the provisioner (maps to step-ca's `sshTemplate`).
* `attestation_template` - (Optional) Name of an attestation template to bind to the provisioner (maps to step-ca's `attestationTemplate`).

## Attributes Reference

This resource has no additional attributes.
