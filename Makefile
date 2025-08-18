SHELL := /bin/bash

.PHONY: help test test-unit test-integration test-e2e test-all

help:
	@echo "Targets:"
	@echo "  test           - same as test-unit"
	@echo "  test-unit      - fast unit tests only (no integration/e2e)"
	@echo "  test-integration - integration tests (module interactions, no real network)"
	@echo "  test-e2e       - end-to-end tests (<=30s total)"
	@echo "  test-all       - unit, integration, then e2e"

# Common environment safety: clear provider env so tests don't hit real network
define CLEAR_ENV
OPENAI_API_KEY= COPILOT_API_BASE= COPILOT_API_KEY= GEMINI_API_KEY= GEMINI_API_BASE=
endef

# Default test is unit tests
test: test-unit

# Unit tests: quick, high coverage, no integration/e2e (tags excluded by default)
VERBOSE?=0
GOFLAGS?=

test-unit:
    @echo "Running unit tests..."
    $(CLEAR_ENV) GOCACHE=$(PWD)/.gocache go test ./... -cover -short $(GOFLAGS) $(if $(filter 1,$(VERBOSE)),-v,)

# Integration tests: enable tests tagged 'integration'. These use fakes/mocks and
# exercise module interactions (e.g., provider HTTP roundtrips) but avoid real network.
test-integration:
    @echo "Running integration tests..."
    $(CLEAR_ENV) GOCACHE=$(PWD)/.gocache go test -tags=integration ./... -cover -timeout=20s $(GOFLAGS) $(if $(filter 1,$(VERBOSE)),-v,)

# E2E tests: enabled via 'e2e' build tag and capped at 30s by a TestMain guard.
# These live in the root package.
test-e2e:
    @echo "Running e2e tests (budget 30s)..."
    $(CLEAR_ENV) GOCACHE=$(PWD)/.gocache go test -tags=e2e . -timeout=30s -cover $(GOFLAGS) $(if $(filter 1,$(VERBOSE)),-v,)

# Run everything in sequence
test-all: test-unit test-integration test-e2e
