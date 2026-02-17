# Set Shell to bash, otherwise some targets fail with dash/zsh etc.
SHELL := /usr/bin/env bash

# Disable built-in rules
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables
.SUFFIXES:
.SECONDARY:
.DEFAULT_GOAL := help

# General variables
include Makefile.vars.mk

# CI automation
-include ci.mk

golangci_bin = $(go_bin)/golangci-lint

.PHONY: help
help: ## Show this help
	@grep -E -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: build-bin ## All-in-one build

.PHONY: build-bin
build-bin: export CGO_ENABLED = 0
build-bin: fmt vet ## Build binary
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -o $(BIN_FILENAME) .

.PHONY: test
test: test-go ## All-in-one test

.PHONY: test-go
test-go: ## Run unit tests against code
	go test -race ./...

.PHONY: fmt
fmt: ## Run 'go fmt' against code
	go fmt ./...

.PHONY: vet
vet: ## Run 'go vet' against code
	go vet ./...

.PHONY: lint
lint: generate fmt golangci-lint git-diff ## All-in-one linting

git-diff:
	@echo 'Check for uncommitted changes ...'
	git diff --exit-code

.PHONY: golangci-lint
golangci-lint: $(golangci_bin) ## Run golangci linters
	$(golangci_bin) run --timeout 5m ./...

.PHONY: generate
generate: generate-go ## All-in-one code generation

.PHONY: generate-go
generate-go: ## Generate Go artifacts
	@go generate ./...

.PHONY: install-crd
install-crd: export KUBECONFIG = $(KIND_KUBECONFIG)
install-crd: generate kind-setup ## Install CRDs into cluster
	kubectl apply -f package/crds

.PHONY: run-operator
run-operator: ## Run in Operator mode against your current kube context
	go run . -v 1 operator

.PHONY: clean
clean: ## Cleans local build artifacts
	rm -rf dist .cache .work

$(golangci_bin): | $(go_bin)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go_bin)"
