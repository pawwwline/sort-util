GOLANGCI_LINT ?= golangci-lint

# --- Dev profile ---
lint-dev:
	@echo "Running golangci-lint (Dev profile)..."
	$(GOLANGCI_LINT) run --config .golangci-dev.yml

# --- CI profile ---
lint-ci:
	@echo "Running golangci-lint (CI profile)..."
	$(GOLANGCI_LINT) run --config .golangci.yml

# --- Formatting code ---
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .

# --- All for dev ---
check-dev: fmt lint-dev

# --- All for ci ---
check-ci: fmt lint-ci
