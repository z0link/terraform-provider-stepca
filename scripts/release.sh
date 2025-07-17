#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-$(git rev-parse --short HEAD)}"
BINARY="terraform-provider-stepca"
DIST_DIR="dist"
SEMVER="0.0.0-${VERSION}"

echo "Building provider binary..."
GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/z0link/terraform-provider-stepca/internal/version.Version=${SEMVER}" -o "$BINARY"

mkdir -p "$DIST_DIR"
zip "$DIST_DIR/${BINARY}_${VERSION}_linux_amd64.zip" "$BINARY"

if [[ "${PUBLISH_TO_TERRAFORM_REGISTRY:-false}" == "true" ]]; then
  echo "Pushing provider version $VERSION to Terraform Registry"
  terraform providers push \
    --os=linux --arch=amd64 \
    --version "$VERSION" \
    "$BINARY"
fi
