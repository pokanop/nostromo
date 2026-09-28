.PHONY: all build test check snapshot release help

BIN ?= $(HOME)/bin/nostromo

define success
	echo "\033[0;32m✔\033[0m $(1)"
endef

define failure
	echo "\033[0;31m✘\033[0m $(1)"
endef

all: build ## Standard make call should simply build

build: ## Builds the project
	@go build -o $(BIN) . && \
		$(call success,Build completed) || $(call failure,Build failed)

test: ## Runs vet and the test suite
	@go vet ./... && go test -race ./... && \
		$(call success,Tests passed) || $(call failure,Tests failed)

check: ## Validates the goreleaser configuration
	@goreleaser check

snapshot: ## Builds all release artifacts and package manager manifests locally under dist/ (nothing is published)
	@goreleaser release --snapshot --clean --skip=publish && \
		$(call success,Snapshot completed) || $(call failure,Snapshot failed)

release: test ## Tags VERSION (e.g. make release VERSION=v1.2.3) and pushes it; CI runs goreleaser and publishes to all package managers
	@test -n "$(VERSION)" || { $(call failure,VERSION is required (make release VERSION=vX.Y.Z)); exit 1; }
	@git tag -a $(VERSION) -m "Release $(VERSION)" && git push origin $(VERSION) && \
		$(call success,Release $(VERSION) tagged and pushed) || $(call failure,Release failed)

help: ## Print help documentation
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
