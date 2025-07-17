BINARY := terraform-provider-stepca
VERSION ?= $(shell git rev-parse --short HEAD)
SEMVER ?= 0.0.0-$(VERSION)
DIST_DIR := dist

.PHONY: build binary package test release

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
