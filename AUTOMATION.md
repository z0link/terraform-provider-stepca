# Automation Instructions

This repository uses GitHub Actions to build and publish test releases of the Terraform provider.

The main workflow is defined in `.github/workflows/release.yml`. Every push to `main` builds the provider and creates a prerelease on GitHub.

## Permissions

The workflow relies on the built in `GITHUB_TOKEN`. Ensure the workflow has permission to write repository contents:

```yaml
permissions:
  contents: write
```

With these permissions the release step can create a GitHub release and upload the packaged binary.

## Publishing to the Terraform Registry

Publishing to the Terraform Registry is optional. Set the environment variable `PUBLISH_TO_TERRAFORM_REGISTRY=true` and provide a `TF_API_TOKEN` secret to enable it.

## Dependencies

- Go 1.24 or newer must be available on the runner.
- The `terraform` CLI is required only when pushing to the Terraform Registry.
- `zip` is used by the packaging script (already installed on GitHub hosted runners).
