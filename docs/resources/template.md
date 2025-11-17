# stepca_template

Creates or updates stored certificate templates via the step-ca admin API.
Templates allow you to customize X.509 or SSH certificates before attaching
those templates to a provisioner. Remote provisioner management must be
enabled on the CA and the provider must be configured with an `admin_token`
that can call the admin endpoints.

## Example Usage

```hcl
resource "stepca_template" "cicd" {
  name = "cicd-leaf"
  body = jsonencode({
    subject = {
      commonName = "ci.internal"
    }
    sans = ["ci.internal", "10.0.0.5"]
    keyUsage     = ["digitalSignature"]
    extKeyUsage  = ["serverAuth"]
  })
  metadata = {
    type    = "x509"
    purpose = "ci"
  }
}
```

Combine this resource with `stepca_provisioner` by referencing the template
name inside the provisioner options, just like you would do via
`step ca provisioner update --x509-template ...`.

## Argument Reference

* `name` - (Required) Unique name of the template stored in step-ca.
* `body` - (Required) JSON template content that step-ca renders. The body is
treated as a literal string, so wrap structured JSON with `jsonencode`.
* `metadata` - (Optional) Map of metadata that can be used to describe the
template (for example type, owner, or version labels).

## Attributes Reference

This resource exports no additional attributes.
