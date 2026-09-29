# proteng-user-mgmt: auth, users and admins, and notification email.
# Run `make` to list targets. On Windows run it from Git Bash.

SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help
MAKEFLAGS += --no-print-directory

# python3 on Windows is often the Microsoft Store stub, so check that it runs.
PYTHON ?= $(shell python3 -c 'import sys' >/dev/null 2>&1 && echo python3 || echo python)
IMAGE ?= proteng-user-mgmt

.PHONY: help setup hooks env deps run mailhog-up mailhog-down fmt lint test check build docker-build

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "} /^##@/ {printf "\n%s\n", substr($$0, 5)} /^[a-zA-Z0-9_.-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##@ Setup
setup: hooks env deps ## First-time setup: git hooks, .env.local and Go modules (safe to re-run)

hooks: ## Install the git hooks (needs `pip install pre-commit`)
	$(PYTHON) -m pre_commit install --hook-type pre-commit --hook-type pre-push --hook-type commit-msg

env: ## Create .env.local from .env.example (never overwrites)
	@if [ -f .env.local ]; then echo ".env.local exists, left as is"; else cp .env.example .env.local; echo "created .env.local from .env.example"; fi

deps: ## Download the Go modules
	go mod download

##@ Run (needs RabbitMQ and MongoDB: make -C ../manual-guides-2023 infra-up)
run: ## Start Mailhog, then run with ENV=local from the repo root (reads .env.local)
	bash run.sh

mailhog-up: ## Start Mailhog: SMTP on :1025, inbox at http://localhost:8025
	@docker start mailhog >/dev/null 2>&1 || docker run -d --name mailhog -p 1025:1025 -p 8025:8025 mailhog/mailhog >/dev/null
	@echo "Mailhog inbox: http://localhost:8025"

mailhog-down: ## Stop Mailhog (its inbox is kept until the container is removed)
	-docker stop mailhog

##@ Checks
fmt: ## Format every Go file with gofmt
	gofmt -l -w .

lint: ## Fail on files gofmt would change, then go vet
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "not gofmt'd (run: make fmt):"; echo "$$out"; exit 1; fi
	go vet ./...

test: ## Run the unit tests
	go test ./...

check: lint test ## Everything CI checks

build: ## Compile every package
	go build ./...

docker-build: ## Build the image locally
	docker build -t $(IMAGE) .
