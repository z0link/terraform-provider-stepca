# Terraform Provider StepCA

This provider aims to expose [step-ca](https://github.com/smallstep/certificates) CLI operations as declarative Terraform resources. It currently offers a simple resource for signing certificates using the `/sign` API endpoint, but will expand to cover more of step-ca's functionality.

Provider documentation for registry publishing is located in the `docs` directory.

## Project goals

The long-term objective is to manage step-ca configuration statefully through Terraform. Planned capabilities include:

- Creating and managing private keys
- Managing provisioners
- Creating, renewing, and revoking certificates and CSRs
- Generating `defaults.json` and `ca.json` from resource and data references
- Additional step-ca CLI features where it makes sense to manage them declaratively

## Requirements

* [Go](https://go.dev/) 1.24 or newer must be installed and in your `PATH`.
* The `terraform` CLI is only required when publishing to the Terraform registry.
* Install the `step-kms-plugin` so YubiKey-backed keys can be used for the
  provider's `admin_key`.

## Building

```
make build
```

## Testing

`make test` runs all Go unit tests. Terraform acceptance tests for the
certificate, provisioner, and admin resources are implemented with the Terraform
Plugin Framework and run automatically when `TF_ACC=1` is set.

### Running the acceptance tests locally

Follow the local CA instructions from `AGENTS.md` to install Terraform, step-cli,
and step-ca:

```
apt-get update
apt-get install -y terraform
curl -L https://dl.smallstep.com/gh-release/cli/docs-cli-install/latest/step-cli_amd64.deb -o step-cli.deb
apt-get install -y ./step-cli.deb
curl -L https://dl.smallstep.com/gh-release/certificates/docs-ca-install/latest/step-ca_amd64.deb -o step-ca.deb
apt-get install -y ./step-ca.deb
```

Initialize a CA (the password file can contain any passphrase) and start it:

```
step ca init --name local-ca --dns localhost --address :9000 \
  --provisioner admin@example.com --password-file password.txt
step-ca $(pwd)/config/ca.json &
export STEP_CA_PID=$!
```

Trust the root certificate for local HTTPS calls:

```
export SSL_CERT_FILE=$(pwd)/certs/root_ca.crt
```

Generate tokens for the tests and export the environment variables that the
acceptance tests expect:

```
export STEPCA_TEST_CA_URL=https://localhost:9000
export STEPCA_TEST_ADMIN_NAME=admin@example.com
export STEPCA_TEST_ADMIN_KEY=$(pwd)/secrets/provisioner.key
export STEPCA_TEST_ADMIN_PROVISIONER=admin@example.com
export STEPCA_TEST_TOKEN=$(step ca token localhost --issuer "$STEPCA_TEST_ADMIN_PROVISIONER" \
  --key "$STEPCA_TEST_ADMIN_KEY" --password-file password.txt)
export STEPCA_TEST_ADMIN_TOKEN=$(step ca admin token "$STEPCA_TEST_ADMIN_NAME" \
  --key "$STEPCA_TEST_ADMIN_KEY" --password-file password.txt)
```

Run the full suite with:

```
TF_ACC=1 make test
```

Stop the CA after the tests finish:

```
kill $STEP_CA_PID
```

## Releasing

A helper script is provided for publishing the provider to the Terraform registry when it is ready. The actual release is disabled by default and requires setting `PUBLISH_TO_TERRAFORM_REGISTRY=true`.

```
PUBLISH_TO_TERRAFORM_REGISTRY=true make release VERSION=v0.1.0
```

## Example Usage

```
provider "stepca" {
  ca_url = "https://ca.example.com"
  token  = "<one-time-token>"
  admin_name = "admin@example.com"
  admin_key  = "/path/to/admin.key"
  admin_provisioner = "admin"
  # Supply a token generated for the admin API. When using a JWK admin
  # provisioner you can create this token with your `admin_key` using
  # `step ca admin` or `step ca token --issuer <admin_provisioner>`.
  admin_token = "<admin-token>"
}

resource "stepca_certificate" "example" {
  csr = file("example.csr")
}

# Manage a provisioner
resource "stepca_provisioner" "admin" {
  name  = "admin"
  type  = "JWK"
  admin = true
}

# Manage an admin
resource "stepca_admin" "alice" {
  name        = "alice"
  provisioner_name = stepca_provisioner.admin.name
}
```

### Step CA Initialization

After running `step ca init` a single JWK provisioner with admin privileges is
created. Additional provisioners can be managed via the `stepca_provisioner`
resource. Use the provider's optional `admin_token` argument with a token issued
for an existing admin provisioner when creating new ones. Tokens can be
generated using `step ca admin` or `step ca token --issuer <provisioner>` with
the corresponding admin key. Set `admin = true` to create another admin
provisioner if desired.

The resulting certificate will be available as the `certificate` attribute.

### Data Sources

Query the CA version:

```hcl
data "stepca_version" "current" {}
```

The version string will be available as `data.stepca_version.current.version`.

Fetch the root certificate:

```hcl
data "stepca_ca_certificate" "root" {}
```

The certificate is accessible via `data.stepca_ca_certificate.root.certificate`.

## Test Releases


Every push to `main` publishes a prerelease on GitHub using the latest commit
hash as the version. The packaged provider binary is uploaded as
`terraform-provider-stepca_<commit>_linux_amd64.zip`. The provider itself is
built with a pseudo-semver version in the form `0.0.0-<commit>` so Terraform can
process version constraints during testing.

To use a test build from this repository specify the pseudo-semver version as
the provider version:


```hcl
terraform {
  required_providers {
    stepca = {
      source  = "registry.terraform.io/local/stepca"
      version = "0.0.0-<commit>"
    }
  }
}
```

Replace `<commit>` with the commit hash shown on the GitHub releases page.

## Using a Local Build

You can test the provider without publishing it to the Terraform Registry.
Build the provider and place it under Terraform's plugin directory so that
`terraform init` can discover it.

```
# Build the binary with an explicit version
SEMVER=0.1.0 make package

# Install it into Terraform's plugin directory
make install-local SEMVER=0.1.0
```

Terraform configurations then reference the local provider source:

```hcl
terraform {
  required_providers {
    stepca = {
      source  = "registry.terraform.io/local/stepca"
      version = "0.1.0"
    }
  }
}
```

