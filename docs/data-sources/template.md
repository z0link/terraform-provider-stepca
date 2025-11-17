---
page_title: "stepca_template Data Source"
subcategory: "Templates"
description: |-
  Fetch an existing X.509 or SSH template from step-ca via the admin API.
---

# stepca_template (Data Source)

Use this data source to load the rendered body and metadata for a template that
was previously stored in step-ca. The returned values let you reuse template
content in multiple Terraform modules without duplicating JSON payloads.

## Example Usage

```hcl
data "stepca_template" "ssh_admin" {
  name = "ssh-admin"
}

output "template_body" {
  value = jsondecode(data.stepca_template.ssh_admin.body)
}
```

You can find a complete example in [`docs/examples/template/data-source.tf`](../examples/template/data-source.tf).

## Argument Reference

* `name` - (Required) The template name to fetch.

## Attributes Reference

* `body` - The template body returned by the admin API.
* `metadata` - Optional metadata map stored alongside the template.
