#!/usr/bin/env bash
set -euo pipefail

BINARY="terraform-provider-stepca"
STEP_VERSION="$(step version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -n 1)"
COMMIT_TAG="$(git tag --points-at HEAD | head -n 1)"
if [[ -z "$COMMIT_TAG" ]]; then
  COMMIT_TAG="$(git rev-parse --short=6 HEAD)"
fi
VERSION="${1:-stepca-${STEP_VERSION}-${COMMIT_TAG}}"
DIST_DIR="dist"

echo "Building provider binary..."
GOOS=linux GOARCH=amd64 go build -o "$BINARY"

mkdir -p "$DIST_DIR"
zip "$DIST_DIR/${BINARY}_${VERSION}_linux_amd64.zip" "$BINARY"

if [[ "${PUBLISH_TO_TERRAFORM_REGISTRY:-false}" == "true" ]]; then
  echo "Pushing provider version $VERSION to Terraform Registry"
  terraform providers push \
    --os=linux --arch=amd64 \
    --version "$VERSION" \
    "$BINARY"
fi
