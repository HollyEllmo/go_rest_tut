# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

- **Build the application**: `make build` (builds to `bin/go_rest_tut`)
- **Run tests**: `make test` (runs all tests with verbose output)
- **Run the application**: `make run` (builds and runs the binary)
- **Direct Go commands**:
  - Build: `go build -o bin/go_rest_tut cmd/main.go`
  - Test: `go test -v ./...`

## Architecture Overview

This is a Go REST API tutorial project with a clean, layered architecture:

### Project Structure
- `cmd/main.go`: Application entry point, initializes and starts the API server
- `cmd/api/api.go`: Core API server implementation using Gorilla Mux router
- `cmd/service/user/routes.go`: User service handlers and route registration
- `cmd/migrate/`: Database migration directory (currently empty)

### Key Components
1. **APIServer**: Central server struct that manages HTTP routing and database connections
2. **Service Handlers**: Modular handlers organized by domain (e.g., user service)
3. **Router Architecture**: Uses Gorilla Mux with versioned API paths (`/api/v1`)

### Dependencies
- **Gorilla Mux**: HTTP router and URL matcher (v1.8.1)
- **Go 1.24.4**: Language version

### Current Implementation
- Server runs on `localhost:8080`
- User service provides `/login` and `/register` endpoints (POST methods)
- Database integration is set up but not yet implemented (nil db passed to APIServer)
- Migration system structure exists but is not populated

The architecture follows Go's standard project layout with domain-driven service organization.