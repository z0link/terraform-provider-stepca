build:
	go build

test:
	go test ./...

release:
	scripts/release.sh $(VERSION)
