# StepCA Provider

Use the **stepca** provider to interact with a [step-ca](https://github.com/smallstep/certificates) Certificate Authority.

## Example Usage

```hcl
terraform {
  required_providers {
    stepca = {
      source = "github.com/z0link/stepca"
      version = "0.1.0"
    }
  }
}

provider "stepca" {
  ca_url = "https://ca.example.com"
  token  = "<one-time-token>"
  admin_provisioner = "admin"

  # Provide either an admin token or the key pair.
  # admin_token = "<admin-token>"
  # admin_name  = "admin@example.com"
  # admin_key   = "/path/to/admin.key"
}
```

## Argument Reference

The following arguments are supported:

* `ca_url` - (Required) The base URL of the step-ca instance.
* `admin_name` - (Optional) The admin user name. Required when setting
  `admin_key`.
* `admin_key` - (Optional) Path or KMS URI of the admin private key. Keys on a
  YubiKey can be referenced via the `step-kms-plugin` URI scheme. Required when
  setting `admin_name`.
* `admin_provisioner` - (Optional) Name of the JWK admin provisioner.
* `token`  - (Required) The one-time bootstrap token used to authenticate.
* `admin_token` - (Optional) Token used for admin API operations. The CA
  initialized by `step ca init` includes a single JWK admin provisioner. Use a
  token issued for that provisioner or another admin to manage resources that
  require admin privileges, or provide `admin_name`/`admin_key` so Terraform can
  mint its own tokens. The provider will emit a configuration error if neither
  an admin token nor the key pair is supplied.

## Resources

* [`stepca_certificate`](resources/certificate.md) - Sign a CSR and obtain a certificate.
* [`stepca_provisioner`](resources/provisioner.md) - Manage provisioners.
* [`stepca_admin`](resources/admin.md) - Manage admin users.

## Data Sources

* [`stepca_version`](data-sources/version.md) - Retrieve the CA version.
* [`stepca_ca_certificate`](data-sources/ca_certificate.md) - Fetch the root certificate.
* [`stepca_provisioners`](data-sources/provisioners.md) - List provisioners via the admin API.
