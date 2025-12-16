# Contributing to gtfs-merge-go

Thank you for your interest in contributing to gtfs-merge-go!

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/aaronbrethorst/gtfs-merge-go.git
   cd gtfs-merge-go
   ```

2. Ensure you have Go 1.21+ installed:
   ```bash
   go version
   ```

3. Run tests to verify your setup:
   ```bash
   go test ./...
   ```

## Development Approach

This project follows **Test-Driven Development (TDD)**:

1. Write tests that define expected behavior
2. Run tests and observe them fail (red)
3. Write minimal code to make tests pass (green)
4. Refactor while keeping tests passing

## Code Standards

### Formatting

All code must be properly formatted:
```bash
gofmt -w .
```

### Static Analysis

Run static analysis before submitting:
```bash
go vet ./...
golangci-lint run
```

### Testing

Run the full test suite with race detector:
```bash
go test -race ./...
```

For benchmarks:
```bash
go test -bench=. ./...
```

## Project Structure

- **`gtfs/`** - GTFS data model and I/O
- **`merge/`** - Core merge orchestration
- **`strategy/`** - Entity-specific merge strategies
- **`scoring/`** - Duplicate similarity scoring
- **`compare/`** - Java comparison testing
- **`cmd/gtfs-merge/`** - CLI application
- **`testdata/`** - Test fixtures

## Making Changes

1. Create a feature branch from `main`
2. Make your changes following the code standards above
3. Add tests for any new functionality
4. Ensure all tests pass: `go test -race ./...`
5. Submit a pull request

## Integration Tests

If your changes affect merge behavior, run the Java integration tests:

1. Install Java 21+
2. Download the test JAR: `./testdata/java/download.sh`
3. Run integration tests: `go test -v -tags=java ./compare/...`

## Documentation

- Update GoDoc comments for any new or modified public APIs
- Update README.md if adding new features
- Keep CLAUDE.md current for AI-assisted development

## Questions?

Open an issue for any questions or concerns about contributing.
