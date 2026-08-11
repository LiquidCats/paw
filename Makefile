SHELL := /bin/sh

.DEFAULT_GOAL := help

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

CONTAINER_RUNTIME ?= docker
GO_VERSION ?= 1.26.4
GOLANGCI_LINT_VERSION ?= 2.11.4
MOCKERY_VERSION ?= 3.7

GO_IMAGE ?= golang:$(GO_VERSION)
GOLANGCI_LINT_IMAGE ?= golangci/golangci-lint:v$(GOLANGCI_LINT_VERSION)-alpine
MOCKERY_IMAGE ?= vektra/mockery:$(MOCKERY_VERSION)
SQLC_IMAGE ?= sqlc/sqlc
BUF_IMAGE ?= bufbuild/buf

GO_TEST_FLAGS ?= -v
GO_BENCH_FLAGS ?= -benchmem
GOLANGCI_LINT_FLAGS ?=

# Set MODULE to a path relative to the repository root to run a target for one
# module. The default runs the target for every module that supports it.
MODULE ?= all

GO_MODULES := $(sort $(patsubst $(ROOT_DIR)/%/go.mod,%,$(wildcard \
	$(ROOT_DIR)/*/go.mod \
	$(ROOT_DIR)/*/*/go.mod)))
SQLC_MODULES := $(sort $(patsubst $(ROOT_DIR)/%/sqlc.yaml,%,$(wildcard \
	$(ROOT_DIR)/*/sqlc.yaml \
	$(ROOT_DIR)/*/*/sqlc.yaml)))
MOCKERY_MODULES := $(sort $(patsubst $(ROOT_DIR)/%/.mockery.yaml,%,$(wildcard \
	$(ROOT_DIR)/*/.mockery.yaml \
	$(ROOT_DIR)/*/*/.mockery.yaml)))

SELECTED_GO_MODULES = $(if $(filter all,$(MODULE)),$(GO_MODULES),$(MODULE))
SELECTED_SQLC_MODULES = $(if $(filter all,$(MODULE)),$(SQLC_MODULES),$(MODULE))
SELECTED_MOCKERY_MODULES = $(if $(filter all,$(MODULE)),$(MOCKERY_MODULES),$(MODULE))

.PHONY: help modules test bench lint lint-fix mock generate-sql \
	proto-lint proto-generate proto-clean run-tests run-bench db-gen

help: ## Show available targets and usage
	@printf '%s\n' \
		'Usage: make <target> [MODULE=path]' \
		'' \
		'Examples:' \
		'  make test' \
		'  make test MODULE=services/rater' \
		'  make lint-fix MODULE=lib/workers' \
		'  make generate-sql MODULE=services/watcher' \
		'' \
		'Targets:' \
		'  modules         List discovered Go, sqlc, and mockery modules' \
		'  test            Run Go tests' \
		'  bench           Run Go benchmarks' \
		'  lint            Run golangci-lint' \
		'  lint-fix        Run golangci-lint with automatic fixes' \
		'  mock            Generate mocks' \
		'  generate-sql    Generate sqlc code' \
		'  proto-lint      Lint protobuf definitions' \
		'  proto-generate  Generate protobuf code' \
		'  proto-clean     Remove generated protobuf code'

modules: ## List modules supported by each target type
	@printf 'Go modules:\n'; printf '  %s\n' $(GO_MODULES)
	@printf '\nsqlc modules:\n'; printf '  %s\n' $(SQLC_MODULES)
	@printf '\nmockery modules:\n'; printf '  %s\n' $(MOCKERY_MODULES)

test: ## Run tests (MODULE=all or a Go module path)
	@set -eu; \
	for module in $(SELECTED_GO_MODULES); do \
		if [ ! -f "$(ROOT_DIR)/$$module/go.mod" ]; then \
			echo "unknown Go module: $$module" >&2; \
			exit 2; \
		fi; \
		echo "==> Testing $$module"; \
		$(CONTAINER_RUNTIME) run --rm -i \
			-e GOWORK=/src/go.work \
			-v "$(ROOT_DIR):/src" \
			-w "/src/$$module" \
			"$(GO_IMAGE)" \
			go test $(GO_TEST_FLAGS) ./...; \
	done

bench: ## Run benchmarks (MODULE=all or a Go module path)
	@set -eu; \
	for module in $(SELECTED_GO_MODULES); do \
		if [ ! -f "$(ROOT_DIR)/$$module/go.mod" ]; then \
			echo "unknown Go module: $$module" >&2; \
			exit 2; \
		fi; \
		echo "==> Benchmarking $$module"; \
		$(CONTAINER_RUNTIME) run --rm -i \
			-e GOWORK=/src/go.work \
			-v "$(ROOT_DIR):/src" \
			-w "/src/$$module" \
			"$(GO_IMAGE)" \
			go test -run '^$$' -bench . $(GO_BENCH_FLAGS) ./...; \
	done

lint: ## Run lint checks (MODULE=all or a Go module path)
	@set -eu; \
	for module in $(SELECTED_GO_MODULES); do \
		if [ ! -f "$(ROOT_DIR)/$$module/go.mod" ]; then \
			echo "unknown Go module: $$module" >&2; \
			exit 2; \
		fi; \
		echo "==> Linting $$module"; \
		$(CONTAINER_RUNTIME) run --rm -i \
			-e GOWORK=/src/go.work \
			-v "$(ROOT_DIR):/src" \
			-w "/src/$$module" \
			"$(GOLANGCI_LINT_IMAGE)" \
			golangci-lint run --config=/src/.golangci.yaml $(GOLANGCI_LINT_FLAGS) ./...; \
	done

lint-fix: ## Run lint checks and apply fixes (MODULE=all or a Go module path)
	@set -eu; \
	for module in $(SELECTED_GO_MODULES); do \
		if [ ! -f "$(ROOT_DIR)/$$module/go.mod" ]; then \
			echo "unknown Go module: $$module" >&2; \
			exit 2; \
		fi; \
		echo "==> Fixing lint issues in $$module"; \
		$(CONTAINER_RUNTIME) run --rm -i \
			-e GOWORK=/src/go.work \
			-v "$(ROOT_DIR):/src" \
			-w "/src/$$module" \
			"$(GOLANGCI_LINT_IMAGE)" \
			golangci-lint run --fix --config=/src/.golangci.yaml $(GOLANGCI_LINT_FLAGS) ./...; \
	done

mock: ## Generate mocks (MODULE=all or a module with .mockery.yaml)
	@set -eu; \
	for module in $(SELECTED_MOCKERY_MODULES); do \
		if [ ! -f "$(ROOT_DIR)/$$module/.mockery.yaml" ]; then \
			echo "module has no .mockery.yaml: $$module" >&2; \
			exit 2; \
		fi; \
		echo "==> Generating mocks for $$module"; \
		$(CONTAINER_RUNTIME) run --rm -i \
			-e GOWORK=/src/go.work \
			-v "$(ROOT_DIR):/src" \
			-w "/src/$$module" \
			"$(MOCKERY_IMAGE)"; \
	done

generate-sql: ## Generate sqlc code (MODULE=all or a module with sqlc.yaml)
	@set -eu; \
	for module in $(SELECTED_SQLC_MODULES); do \
		if [ ! -f "$(ROOT_DIR)/$$module/sqlc.yaml" ]; then \
			echo "module has no sqlc.yaml: $$module" >&2; \
			exit 2; \
		fi; \
		echo "==> Generating SQL for $$module"; \
		$(CONTAINER_RUNTIME) run --rm \
			-v "$(ROOT_DIR):/src" \
			-w "/src/$$module" \
			"$(SQLC_IMAGE)" generate; \
	done

proto-lint: ## Lint protobuf definitions
	$(CONTAINER_RUNTIME) run --rm \
		-v "$(ROOT_DIR)/protos:/defs" \
		-w /defs \
		"$(BUF_IMAGE)" lint

proto-generate: proto-clean ## Generate protobuf code
	$(CONTAINER_RUNTIME) run --rm \
		-v "$(ROOT_DIR)/protos:/defs" \
		-w /defs \
		"$(BUF_IMAGE)" generate

proto-clean: ## Remove generated protobuf code
	rm -rf "$(ROOT_DIR)/protos/gen/go"

# Backward-compatible aliases for the old services/kms Makefile.
run-tests: test
run-bench: bench
db-gen: generate-sql
