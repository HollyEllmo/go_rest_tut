# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

- **Build the application**: `make build` (builds to `bin/go_rest_tut`)
- **Run tests**: `make test` (runs all tests with verbose output)
- **Run the application**: `make run` (builds and runs the binary)
- **Database migrations**:
  - Create migration: `make migration <migration_name>`
  - Run migrations up: `make migrate-up`
  - Run migrations down: `make migrate-down`
- **Clean build artifacts**: `make clean`
- **Direct Go commands**:
  - Build: `go build -o bin/go_rest_tut cmd/main.go`
  - Test: `go test -v ./...`

## Architecture Overview

This is a Go REST API tutorial project with a clean, layered architecture:

### Project Structure
- `cmd/main.go`: Application entry point, initializes database and starts the API server
- `cmd/api/api.go`: Core API server implementation using Gorilla Mux router
- `cmd/config/env.go`: Environment configuration management
- `cmd/db/db.go`: Database connection and MySQL storage setup
- `cmd/migrate/`: Database migration system with SQL migration files
- `cmd/service/`: Domain-specific service handlers (user, product, cart, order)
- `cmd/types/types.go`: Shared type definitions
- `cmd/utils/utils.go`: Common utility functions

### Key Components
1. **APIServer**: Central server struct that manages HTTP routing and database connections
2. **Service Handlers**: Modular handlers organized by domain (user, product, cart, order)
3. **Store Pattern**: Each service has an associated store for database operations
4. **Router Architecture**: Uses Gorilla Mux with versioned API paths (`/api/v1`)
5. **JWT Authentication**: Token-based authentication system with password hashing
6. **Migration System**: SQL-based database migrations with up/down scripts

### Dependencies
- **Gorilla Mux**: HTTP router and URL matcher (v1.8.1)
- **MySQL Driver**: Database connectivity (go-sql-driver/mysql v1.9.3)
- **JWT**: Token authentication (golang-jwt/jwt/v5 v5.2.3)
- **Validator**: Request validation (go-playground/validator/v10 v10.27.0)
- **Migrate**: Database migration tool (golang-migrate/migrate/v4 v4.18.3)
- **Crypto**: Password hashing (golang.org/x/crypto)
- **Godotenv**: Environment variable loading (joho/godotenv v1.5.1)

### Database Schema
The application includes migrations for:
- Users table (authentication and user data)
- Products table (product catalog)
- Orders table (order management)
- Order items table (order line items)

### Current Implementation
- Server runs on `localhost:8080`
- Full MySQL database integration with connection pooling
- Complete authentication system with JWT tokens and bcrypt password hashing
- RESTful API with user and product services
- Environment-based configuration for database and server settings

The architecture follows Go's standard project layout with domain-driven service organization.