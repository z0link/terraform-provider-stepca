# Codex Agent Instructions

These instructions help ensure Terraform and step-ca are available when running local tests.

## Dependencies
- Install `terraform`, `step-cli`, and `step-ca`. On Debian-based environments you can run:
  ```bash
  apt-get update
  apt-get install -y terraform
  curl -L https://dl.smallstep.com/gh-release/cli/docs-cli-install/latest/step-cli_amd64.deb -o step-cli.deb
  apt-get install -y ./step-cli.deb
  curl -L https://dl.smallstep.com/gh-release/certificates/docs-ca-install/latest/step-ca_amd64.deb -o step-ca.deb
  apt-get install -y ./step-ca.deb
  ```
  Adjust versions as needed if the latest release changes.

## Local CA Setup
- Initialize and start a local `step-ca` before running tests:
  ```bash
  step ca init --name local-ca --dns localhost --address :9000 --provisioner admin@example.com --password-file password.txt
  step-ca $(pwd)/config/ca.json &
  export STEP_CA_PID=$!
  ```
- Run tests with:
  ```bash
  make test
  ```
- Stop the CA after tests:
  ```bash
  kill $STEP_CA_PID
  ```
