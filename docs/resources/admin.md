# stepca_admin

Manages an admin user via the step-ca admin API.

## Example Usage

```hcl
resource "stepca_admin" "alice" {
  name             = "alice"
  provisioner_name = "admin"
}
```

Use the provider's `admin_token` argument with a token issued for the admin
provisioner. When using a JWK admin provisioner you can generate this token
with `step ca admin` or `step ca token` using your `admin_key`.

Admins belong to a specific admin provisioner. Combine this resource with
`stepca_provisioner` to manage both provisioners and their admins.

## Argument Reference

* `name` - (Required) The admin's name/email.
* `provisioner_name` - (Required) Name of the admin provisioner this admin belongs to.

## Attributes Reference

This resource has no additional attributes.
