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
  admin_name = "admin@example.com"
  admin_key  = "/path/to/admin.key"
  admin_provisioner = "admin"
  # Generate a token for the admin API with `step ca admin` or
  # `step ca token --issuer ${admin_provisioner}` and provide it here.
  admin_token = "<admin-token>"
}
```

## Argument Reference

The following arguments are supported:

* `ca_url` - (Required) The base URL of the step-ca instance.
* `admin_name` - (Required) The admin user name.
* `admin_key` - (Required) Path or KMS URI of the admin private key. Keys on a
  YubiKey can be referenced via the `step-kms-plugin` URI scheme.
* `admin_provisioner` - (Optional) Name of the JWK admin provisioner.
* `token`  - (Required) The one-time bootstrap token used to authenticate.
* `admin_token` - (Optional) Token used for admin API operations. The CA
  initialized by `step ca init` includes a single JWK admin provisioner. Use a
  token issued for that provisioner or another admin to manage resources that
  require admin privileges. Although `admin_key` is mandatory for
  authentication, this provider currently expects a pre-generated token rather
  than generating it automatically.

## Resources

* [`stepca_certificate`](resources/certificate.md) - Sign a CSR and obtain a certificate.
* [`stepca_provisioner`](resources/provisioner.md) - Manage provisioners.
* [`stepca_admin`](resources/admin.md) - Manage admin users.

## Data Sources

* [`stepca_version`](data-sources/version.md) - Retrieve the CA version.
* [`stepca_ca_certificate`](data-sources/ca_certificate.md) - Fetch the root certificate.
* [`stepca_provisioners`](data-sources/provisioners.md) - List provisioners via the admin API.
