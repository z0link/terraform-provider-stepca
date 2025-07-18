# stepca_provisioner

Manages a provisioner through the step-ca admin API.

## Example Usage

```hcl
resource "stepca_provisioner" "admin" {
  name  = "admin"
  type  = "JWK"
  admin = true
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

## Attributes Reference

This resource has no additional attributes.
