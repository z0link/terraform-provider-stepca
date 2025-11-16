---
page_title: "stepca_provisioners Data Source"
subcategory: "Provisioners"
description: |-
  Retrieve the provisioners configured in a step-ca instance via the admin API.
---

# stepca_provisioners (Data Source)

Use this data source to list every provisioner that the authenticated admin can access. The resulting list exposes the provisioner name, type, and whether the provisioner has admin privileges.

## Example Usage

```hcl
data "stepca_provisioners" "all" {}

output "admin_provisioners" {
  value = [for p in data.stepca_provisioners.all.provisioners : p if p.admin]
}
```

## Argument Reference

This data source does not take any arguments.

## Attributes Reference

* `provisioners` - List of provisioners returned by the admin API. Each entry exports the following attributes:
  * `name` - Provisioner name.
  * `type` - Provisioner type (for example `JWK`, `ACME`, or `OIDC`).
  * `admin` - Boolean indicating whether the provisioner is flagged as an admin provisioner.
