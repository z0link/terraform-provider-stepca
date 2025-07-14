BINARY := terraform-provider-stepca
VERSION ?= $(shell git rev-parse --short HEAD)
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
