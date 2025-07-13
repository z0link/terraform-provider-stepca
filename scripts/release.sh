#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
  echo "Usage: $0 <version>"
  exit 1
fi

if [[ "${PUBLISH_TO_TERRAFORM_REGISTRY:-false}" != "true" ]]; then
  echo "Publishing disabled. Set PUBLISH_TO_TERRAFORM_REGISTRY=true when ready."
  exit 0
fi

# Build provider binary
GOOS=linux GOARCH=amd64 go build -o terraform-provider-stepca

echo "Pushing provider version $VERSION to Terraform Registry"
terraform providers push \
  --os=linux --arch=amd64 \
  --version "$VERSION" \
  terraform-provider-stepca
