# MCP Go Session Example

This project demonstrates how to build a Model Context Protocol (MCP) server using the [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) with session storage capabilities from the `sessions` branch in the [omgitsads/go-sdk](https://github.com/omgitsads/go-sdk) fork.

> ![IMPORTANT]
> This was mostly done with Claude Code to speed things up and probably shouldn't be used anywhere near production.

## Overview

The main objective of this project is to showcase the enhanced session management features that allow for:

- **Custom Session Storage**: Implement your own session storage backend (e.g., Redis, PostgreSQL, etc.)
- **Multi-Instance Deployment**: Run multiple MCP server instances that share session state
- **Persistent Sessions**: Sessions can survive server restarts and be shared across instances

## Architecture

```
cmd/
├── main.go            # CLI entry point
├── root.go            # Root Cobra command
└── server.go          # Server subcommand

mcp/
└── session_server.go  # MCP server implementation with tools

storage/
└── redis.go           # Redis session storage implementation
```

## Quick Start

### Prerequisites

- Go 1.24.5 or later
- Access to the `sessions` branch from [omgitsads/go-sdk](https://github.com/omgitsads/go-sdk)

### Installation

1. Clone this repository:
   ```bash
   git clone <repository-url>
   cd mcp-go-session-example
   ```

3. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running the Server

**Note: Redis is required for session storage.**

Start the MCP server:
```bash
# Set Redis connection
export REDIS_ADDR=localhost:6379
go run ./cmd server
```

Or with custom host and port using command line flags:
```bash
go run ./cmd server --host 0.0.0.0 --port 3000 --redis-addr localhost:6379
```


## Configuration

The server can be configured using environment variables, command-line flags, or a combination of both. Command-line flags take precedence over environment variables.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MCP_HOST` | Host to bind to | `localhost` |
| `MCP_PORT` | Port to listen on | `8080` |
| `REDIS_ADDR` | Redis server address | _(required)_ |
| `REDIS_PASSWORD` | Redis password | _(empty)_ |
| `REDIS_DB` | Redis database number | `0` |
| `REDIS_PREFIX` | Redis key prefix for sessions | `mcp:session:` |
| `REDIS_TTL` | Redis session TTL | `1h` |

### Example with Environment Variables

```bash
# HTTP server configuration
export MCP_HOST=0.0.0.0
export MCP_PORT=8080

# Redis configuration
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=mypassword
export REDIS_DB=1
export REDIS_PREFIX=myapp:mcp:
export REDIS_TTL=2h

# Start server
go run ./cmd server
```

## Session Storage

This example is built on the `sessions` branch of the go-sdk fork, which introduces a session storage interface. This allows you to:

### Redis Session Storage

This example requires Redis for persistent session storage across multiple server instances.


## Tools

### Hello World Tool

The server includes a simple "hello_world" tool that demonstrates basic tool execution:

- **Name**: `hello_world`
- **Description**: A simple tool that outputs 'Hello world!'
- **Arguments**: None required
- **Response**: Returns "Hello world!" as text content

## Development

This project includes a comprehensive Makefile to streamline development tasks.

### Quick Start with Makefile

```bash
# Show all available commands
make help

# Setup development environment
make setup

# Run the server directly
make run

# Run with Redis
make run-redis

# Build the binary
make build

# Run all quality checks
make check
```

### Building

```bash
# Build the binary
make build
# or manually: go build -o ./bin/mcp-server ./cmd

# Build for multiple platforms
make build-all

# Install to $GOPATH/bin
make install
```

### Development Workflow

```bash
# Start development server with hot reload (requires air)
make dev

# Run code quality checks
make check  # Runs fmt, vet, and test

# Run tests with coverage
make test-cover

# Clean build artifacts
make clean
```

### Redis Development

```bash
# Start Redis container for development
make redis-start

# Run server with Redis
make start-redis

# Stop Redis container
make redis-stop
```

## License

This project follows the same license as the modelcontextprotocol/go-sdk.

## Related Projects

- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) - Official Go SDK for MCP
- [omgitsads/go-sdk](https://github.com/omgitsads/go-sdk) - Fork with session storage enhancements
- [Model Context Protocol Specification](https://modelcontextprotocol.io/) - Official MCP documentation