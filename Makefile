BINARY=terraform-provider-altertable
VERSION?=dev

default: build

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY)

install:
	go install -ldflags "-X main.version=$(VERSION)"

test:
	go test ./... -count=1

testacc:
	TF_ACC=1 go test ./... -count=1 -timeout 120m

lint:
	golangci-lint run

fmt:
	gofmt -s -w .
	terraform fmt -recursive ./examples

docs:
	go tool tfplugindocs generate --provider-name altertable

release:
	./scripts/release.sh

.PHONY: default build install test testacc lint fmt docs release
