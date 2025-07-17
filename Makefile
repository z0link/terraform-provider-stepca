BINARY := terraform-provider-stepca
STEP_VERSION := $(shell step version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -n 1)
COMMIT_TAG := $(shell git tag --points-at HEAD | head -n 1)
ifeq ($(strip $(COMMIT_TAG)),)
COMMIT_TAG := $(shell git rev-parse --short=6 HEAD)
endif
VERSION ?= stepca-$(STEP_VERSION)-$(COMMIT_TAG)
DIST_DIR := dist

.PHONY: build binary package test release

build: package

binary:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)

package: binary
	mkdir -p $(DIST_DIR)
	zip $(DIST_DIR)/$(BINARY)_$(VERSION)_linux_amd64.zip $(BINARY)

test:
	go test ./...

release:
	scripts/release.sh $(VERSION)
