BINARY    := cubeapm
MODULE    := github.com/piyush-gambhir/cubeapm-cli/cli-go
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -s -w \
	-X $(MODULE)/cmd.Version=$(VERSION) \
	-X $(MODULE)/cmd.Commit=$(COMMIT) \
	-X $(MODULE)/cmd.BuildDate=$(BUILD_DATE)

# The main package is named cli-go; explicitly name the installed executable.
INSTALL_DIR ?= $(shell go env GOBIN)
ifeq ($(strip $(INSTALL_DIR)),)
INSTALL_DIR := $(shell go env GOPATH)/bin
endif

.PHONY: build install test lint fmt vet clean tidy

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

install:
	mkdir -p "$(INSTALL_DIR)"
	go build -ldflags "$(LDFLAGS)" -o "$(INSTALL_DIR)/$(BINARY)$(shell go env GOEXE)" .

test:
	go test ./... -v -race -count=1

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
	go clean

tidy:
	go mod tidy

.PHONY: llms-check
llms-check: ## verify docs/llms.txt documents every command
	@bash scripts/check-llms.sh
