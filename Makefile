BINARY := terraform-provider-stepca
VERSION ?= $(shell git rev-parse --short HEAD)
SEMVER ?= 0.0.0-$(VERSION)

DIST_DIR := dist

.PHONY: build binary package test release install-local

build: package

binary:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/z0link/terraform-provider-stepca/internal/version.Version=$(SEMVER)" -o $(BINARY)

package: binary
	mkdir -p $(DIST_DIR)
	zip $(DIST_DIR)/$(BINARY)_$(VERSION)_linux_amd64.zip $(BINARY)

test:
	go test ./...

release:
	scripts/release.sh $(VERSION)

install-local: package
	PLUGIN_DIR=$(HOME)/.terraform.d/plugins/registry.terraform.io/local/stepca/$(SEMVER)/linux_amd64; \
	mkdir -p $$PLUGIN_DIR; \
	unzip -o $(DIST_DIR)/$(BINARY)_$(VERSION)_linux_amd64.zip -d $$PLUGIN_DIR; \
	mv $$PLUGIN_DIR/$(BINARY) $$PLUGIN_DIR/$(BINARY)_v$(SEMVER)
