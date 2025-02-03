# Name of the linter executable
LINTER = golangci-lint

# Default target
all: lint test build

# Install golangci-lint
install-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run lint checks
lint:
	$(LINTER) run --timeout=5m

# Check for security vulnerabilities
vulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# Format Go code
fmt:
	gofmt -w .

# Run unit tests
test:
	go test -v ./...

# Build the project
build:
	go build -o sso cmd/sso/main.go

# Run database migrations
migrate:
	go run cmd/migrator/main.go --db-url="postgres://test:test@localhost:5432/ssotest?sslmode=disable" --migrations-path="./migrations"

# Start the local service
start:
	go run cmd/sso/main.go --config=./config/local_test.yaml

# Clean generated files
clean:
	rm -rf sso

# Help (displays available commands)
help:
	@echo "Makefile for managing Go project"
	@echo "Available commands:"
	@echo "  install-linter  - install golangci-lint"
	@echo "  lint            - run linter to check code"
	@echo "  vulncheck       - check for security vulnerabilities"
	@echo "  fmt             - format Go code"
	@echo "  test            - run unit tests"
	@echo "  build           - build the project"
	@echo "  migrate         - run database migrations"
	@echo "  start           - start the local service"
	@echo "  clean           - clean the environment"

.PHONY: all lint vulncheck fmt test build migrate start clean help
