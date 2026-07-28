SHELL := /usr/bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

GO ?= go
BINARY ?= sst
COMMAND ?= ./cmd/sst
PACKAGES ?= ./...

BIN_DIR ?= $(CURDIR)/bin
ARTIFACT_DIR ?= $(CURDIR)/.artifacts
TOOLS_DIR ?= $(CURDIR)/.tools/bin

GOLANGCI_LINT_VERSION ?= v2.12.2
GOVULNCHECK_VERSION ?= v1.6.0
GOLANGCI_LINT ?= $(TOOLS_DIR)/golangci-lint
GOVULNCHECK ?= $(TOOLS_DIR)/govulncheck

.PHONY: help tools doctor format format-check lint vet mod-check test test-race \
	test-cover vuln check audit build install clean data oracle eval test-fuzz \
	stage2-delta require-module require-lint require-vulncheck

help: ## Show the useful targets.
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\n"} \
		/^[a-zA-Z0-9_-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

tools: ## Install pinned analysis tools into .tools/bin.
	@mkdir -p "$(TOOLS_DIR)"
	@curl -sSfL https://golangci-lint.run/install.sh | \
		sh -s -- -b "$(TOOLS_DIR)" "$(GOLANGCI_LINT_VERSION)"
	@GOBIN="$(TOOLS_DIR)" "$(GO)" install \
		golang.org/x/vuln/cmd/govulncheck@"$(GOVULNCHECK_VERSION)"

doctor: ## Print the active Go and analysis tool versions.
	@"$(GO)" version
	@{ "$(GOLANGCI_LINT)" version || printf '%s\n' "golangci-lint: missing (run make tools)"; }
	@{ "$(GOVULNCHECK)" -version || printf '%s\n' "govulncheck: missing (run make tools)"; }

format: require-module require-lint ## Apply canonical formatting and import organization.
	@"$(GOLANGCI_LINT)" fmt

format-check: require-module require-lint ## Report formatting drift without changing files.
	@"$(GOLANGCI_LINT)" fmt --diff

lint: require-module require-lint ## Run the curated static-analysis suite.
	@"$(GOLANGCI_LINT)" config verify
	@"$(GOLANGCI_LINT)" run $(PACKAGES)

vet: require-module ## Run Go's built-in semantic checks.
	@"$(GO)" vet $(PACKAGES)

mod-check: require-module ## Verify dependencies and show required go.mod/go.sum changes.
	@"$(GO)" mod tidy -diff
	@"$(GO)" mod verify

test: require-module ## Run the test suite in randomized order.
	@"$(GO)" test -shuffle=on $(PACKAGES)

test-race: require-module ## Run tests with the race detector.
	@"$(GO)" test -race -shuffle=on $(PACKAGES)

test-cover: require-module ## Write package coverage data under .artifacts.
	@mkdir -p "$(ARTIFACT_DIR)"
	@"$(GO)" test -coverprofile="$(ARTIFACT_DIR)/coverage.out" $(PACKAGES)
	@"$(GO)" tool cover -func="$(ARTIFACT_DIR)/coverage.out"

vuln: require-module require-vulncheck ## Check reachable known vulnerabilities.
	@"$(GOVULNCHECK)" $(PACKAGES)

check: format-check mod-check lint test ## Run the fast local quality gate.

audit: check test-race vuln ## Run the thorough local quality gate.

test-fuzz: require-module ## Fuzz the analyzer and tokenizer briefly.
	@"$(GO)" test -run xxx -fuzz FuzzAnalyzeInvariants -fuzztime 30s ./slop/
	@"$(GO)" test -run xxx -fuzz FuzzTokenizerSpans -fuzztime 30s ./slop/

data: require-module ## Regenerate embedded lists/calibration from the pinned upstream commit.
	@"$(GO)" run ./tools/datagen

oracle: require-module ## Regenerate golden expectations with the vendored Node oracle.
	@node testdata/upstream/runner.mjs

stage2-delta: require-module ## Measure the POS-tagged stage-2 score delta on the behavior corpus.
	@test -d "$(CURDIR)/testdata/upstream/node_modules/wink-pos-tagger" || { \
		printf '%s\n' "wink-pos-tagger is missing; run npm ci in testdata/upstream."; \
		exit 2; \
	}
	@cd testdata/upstream && node --import ./register-loader.mjs runner.mjs --stage2 \
		../corpus/*.txt ../fixtures/*.txt

eval: require-module ## Run the behavior-corpus evaluation (diagnostic firing rates).
	@"$(GO)" run ./tools/eval

build: require-module ## Build the command into bin/.
	@mkdir -p "$(BIN_DIR)"
	@"$(GO)" build -trimpath -o "$(BIN_DIR)/$(BINARY)" "$(COMMAND)"

install: require-module ## Install the command with the active Go toolchain.
	@"$(GO)" install "$(COMMAND)"

clean: ## Remove repository-local build, tool, and test artifacts.
	@rm -rf -- "$(BIN_DIR)" "$(ARTIFACT_DIR)" "$(CURDIR)/.tools"

require-module:
	@test -f go.mod || { \
		printf '%s\n' "go.mod is not present yet; initialize the implementation first."; \
		exit 2; \
	}

require-lint:
	@test -x "$(GOLANGCI_LINT)" || command -v "$(GOLANGCI_LINT)" >/dev/null 2>&1 || { \
		printf '%s\n' "golangci-lint is missing; run make tools."; \
		exit 2; \
	}

require-vulncheck:
	@test -x "$(GOVULNCHECK)" || command -v "$(GOVULNCHECK)" >/dev/null 2>&1 || { \
		printf '%s\n' "govulncheck is missing; run make tools."; \
		exit 2; \
	}
