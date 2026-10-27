SHELL := /bin/sh

.DEFAULT_GOAL := help

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

CONTAINER_RUNTIME ?= docker
GO_VERSION ?= 1.27
GOLANGCI_LINT_VERSION ?= 2.13.2
MOCKERY_VERSION ?= 3.7

GO_IMAGE ?= golang:$(GO_VERSION)-alpine
GOLANGCI_LINT_IMAGE ?= golangci/golangci-lint:v$(GOLANGCI_LINT_VERSION)-alpine
MOCKERY_IMAGE ?= vektra/mockery:$(MOCKERY_VERSION)
SQLC_IMAGE ?= sqlc/sqlc
BUF_IMAGE ?= bufbuild/buf

.PHONY: help generate-sql generate-mocks proto-lint proto-generate proto-clean

help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk -F '##' '/^[a-zA-Z_-]+:.*?##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

generate-sql: ## Generate sqlc code (MODULE=all or a module with sqlc.yaml)
	$(CONTAINER_RUNTIME) run --rm \
		-v "$(ROOT_DIR):/src" \
		-w "/src" \
		"$(SQLC_IMAGE)" generate; \

generate-mocks: mocks-clean ## Generate mocks
	$(CONTAINER_RUNTIME) run --rm \
		-v "$(ROOT_DIR):/src" \
		-w "/src" \
		"$(MOCKERY_IMAGE)"

proto-lint: ## Lint protobuf definitions
	$(CONTAINER_RUNTIME) run --rm \
		-v "$(ROOT_DIR):/defs" \
		-w /defs \
		"$(BUF_IMAGE)" lint

proto-generate: proto-clean ## Generate protobuf code
	$(CONTAINER_RUNTIME) run --rm \
		-v "$(ROOT_DIR):/defs" \
		-w /defs \
		"$(BUF_IMAGE)" generate

proto-clean: ## Remove generated protobuf code
	rm -rf "$(ROOT_DIR)/pkg/grpc"

mocks-clean: ## Remove generated mocks
	rm -rf "$(ROOT_DIR)/tests/mocks/*" && rm -rf "$(ROOT_DIR)/internal/modules/rater/internal/tests/mocks/*"
