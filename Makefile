.PHONY: all fmt fmt-fix vet lint test test-java build clean validate-ci help

# Default target
all: build

# Check Go formatting (fails if any files need formatting)
fmt:
	@echo "Checking formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Go files are not formatted:"; \
		gofmt -d .; \
		exit 1; \
	fi
	@echo "Formatting OK"

# Auto-fix formatting issues
fmt-fix:
	@echo "Fixing formatting..."
	gofmt -w .
	@echo "Done"

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

# Run tests with race detector
test:
	@echo "Running tests..."
	go test -v -race ./...

# Run Java comparison tests (requires Java 21+)
test-java:
	@echo "Running Java comparison tests..."
	go test -v -tags=java -timeout=60m ./compare/...

# Build the CLI binary
build:
	@echo "Building gtfs-merge..."
	go build -o gtfs-merge ./cmd/gtfs-merge

# Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -f gtfs-merge
	go clean

# Run all CI validation checks
validate-ci: fmt vet lint test
	@echo "All CI checks passed!"

# Show help
help:
	@echo "Available targets:"
	@echo "  make validate-ci  - Run all CI checks (fmt, vet, lint, test)"
	@echo "  make fmt          - Check formatting"
	@echo "  make fmt-fix      - Fix formatting issues"
	@echo "  make vet          - Run go vet"
	@echo "  make lint         - Run golangci-lint"
	@echo "  make test         - Run tests with race detector"
	@echo "  make test-java    - Run Java comparison tests (requires Java 21+)"
	@echo "  make build        - Build the CLI binary"
	@echo "  make clean        - Remove build artifacts"
