#!/usr/bin/env bash
set -euo pipefail

BINARY="terraform-provider-stepca"
VERSION="${1:?version required}"
TEST_VERSION="0.0.0"
DIST_DIR="dist"
SEMVER="0.0.0-${VERSION}"

# Build and start dummy CA
cd "$(dirname "$0")"/..

go build -o dummy_ca ./scripts/dummy_ca.go
./dummy_ca &
CA_PID=$!

cleanup() {
  echo "Cleaning up..."
  kill $CA_PID
  rm -rf "$TMP_DIR" "$PLUGIN_DIR" dummy_ca
}
trap cleanup EXIT
sleep 1

PLUGIN_DIR=$(mktemp -d)
unzip "$DIST_DIR/${BINARY}_${VERSION}_linux_amd64.zip" -d "$PLUGIN_DIR"
mkdir -p "$PLUGIN_DIR/registry.terraform.io/local/stepca/${SEMVER}/linux_amd64"
mv "$PLUGIN_DIR/$BINARY" "$PLUGIN_DIR/registry.terraform.io/local/stepca/${SEMVER}/linux_amd64/${BINARY}_v${SEMVER}"


TMP_DIR=$(mktemp -d)
cat <<TF > "$TMP_DIR/main.tf"
terraform {
  required_providers {
    stepca = {
      source  = "registry.terraform.io/local/stepca"
      version = "${SEMVER}"

    }
  }
}

provider "stepca" {
  ca_url = "http://localhost:8080"
  token  = "dummy"
}

resource "stepca_certificate" "test" {
  csr = "dummy"
}

data "stepca_version" "current" {}

output "cert" {
  value = stepca_certificate.test.certificate
}

output "version" {
  value = data.stepca_version.current.version
}
TF

(cd "$TMP_DIR" && terraform init -no-color -input=false -plugin-dir="$PLUGIN_DIR")
(cd "$TMP_DIR" && terraform apply -no-color -input=false -auto-approve)

version=$(cd "$TMP_DIR" && terraform output -raw version)
cert=$(cd "$TMP_DIR" && terraform output -raw cert)

declare -A tests=(
  ["version"]="1.2.3"
)

for key in "${!tests[@]}"; do
  expected="${tests[$key]}"
  value="${!key}"
  if [[ "$value" != "$expected" ]]; then
    echo "Test $key failed: expected $expected got $value"
    exit 1
  fi
done

if [[ "$cert" != *"BEGIN CERTIFICATE"* ]]; then
  echo "certificate output did not contain a PEM block"
  exit 1
fi

echo "Integration tests passed."


