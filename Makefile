.DEFAULT_GOAL := help

SHELL := bash
PATH := $(CURDIR)/.dev/go-tools/bin:$(PATH)

# Load .env file if it exists.
ifneq (,$(wildcard ./.env))
  include .env
  export
endif

.PHONY: help
help: ## Show help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[/0-9a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2}'


# --------------------------------------------------------------------------------------
# Development environment
# --------------------------------------------------------------------------------------
.PHONY: setup
setup: ## Setup development environment
	@echo "==> Setting up development environment..."
	@mkdir -p $(CURDIR)/.dev/go-tools
	@export GOPATH=$(CURDIR)/.dev/go-tools && \
		go install github.com/axw/gocov/gocov@latest && \
		go install github.com/matm/gocov-html/cmd/gocov-html@latest
	@export GOPATH=$(CURDIR)/.dev/go-tools && go clean -modcache && rm -rf $(CURDIR)/.dev/go-tools/pkg

.PHONY: clean
clean: ## Clean up development environment
	@rm -rf .dev


# --------------------------------------------------------------------------------------
# Testing, Formatting and etc.
# --------------------------------------------------------------------------------------
.PHONY: format
format: ## Format source code
	@go fmt ./...

.PHONY: test
test: ## Run tests
	@go test -race -timeout 30m ./...

.PHONY: test-short
test-short: ## Run short tests
	@go test -short -race -timeout 30m ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose outputting
	@go test -race -timeout 30m -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	@mkdir -p $(CURDIR)/.dev/test
	@go test -race -coverpkg=./... -coverprofile=$(CURDIR)/.dev/test/coverage.out ./...
	@gocov convert $(CURDIR)/.dev/test/coverage.out | gocov-html > $(CURDIR)/.dev/test/coverage.html

.PHONY: test-cover-open
test-cover-open: ## Open coverage report in browser
	@open $(CURDIR)/.dev/test/coverage.html


# --------------------------------------------------------------------------------------
# Go commands
# --------------------------------------------------------------------------------------
.PHONY: go-generate
go-generate: ## Run go generate
	@go generate ./...

.PHONY: go-mod-tidy
go-mod-tidy: ## Run go mod tidy
	@go mod tidy

# --------------------------------------------------------------------------------------
# Website
# --------------------------------------------------------------------------------------
.PHONY: start
start: ## Start echo-viewkit-website server in development mode
	@if [[ ! -d "website/node_modules" ]]; then \
		cd website && npm install; \
	fi
	@cd website && go run . -debug

.PHONY: build
build: ## Build echo-viewkit-website binary
	@if [[ ! -d "website/node_modules" ]]; then \
		cd website && npm install; \
	fi
	@cd website && npm run build
	@mkdir -p .dev/build/dev
	@cd website && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../.dev/build/dev/echo-viewkit-website .

