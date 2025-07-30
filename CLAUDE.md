# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Structure

This is a minimal Go project with the following structure:
- `cmd/` - Contains the main application entry points
  - `main.go` - Main application file (currently contains only package declaration)
  - `root.go` - Root command file (currently empty)
- `go.mod` - Go module definition (module: github.com/omgitsads/mcp-go-session-example, Go 1.24.5)

## Development Commands

This project includes a comprehensive Makefile for common development tasks. Use `make help` to see all available commands.

### Quick Commands
```bash
make help                         # Show all available commands
make run                          # Run the server directly
make build                        # Build the binary
make check                        # Run fmt, vet, and tests
make clean                        # Clean build artifacts
```

### Manual Commands (if not using Makefile)
```bash
go build -o ./bin/mcp-server ./cmd # Build the application
go run ./cmd                       # Run the application directly
go install ./cmd                   # Install the binary
```

### Code Quality
```bash
make fmt                          # Format all Go files
make vet                          # Run Go vet for static analysis
make lint                         # Run golangci-lint
make check                        # Run all quality checks
```

### Testing
```bash
make test                         # Run all tests
make test-race                    # Run tests with race detection
make test-cover                   # Run tests with coverage report
```

## Architecture Notes

The project appears to be in its initial stages with minimal code. The structure suggests it may be intended as a CLI application using the common `cmd/` pattern for Go projects, but the implementation files are currently empty or minimal.