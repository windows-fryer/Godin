# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Run
```bash
# Build the application
go build -o bin/godin cmd/godin/main.go

# Run the application directly
go run cmd/godin/main.go

# Install dependencies
go mod download

# Update dependencies
go mod tidy
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test -v ./internal/api/...
```

### Database Operations
The application uses PostgreSQL with a custom database abstraction layer. Database operations are wrapped in transactions using the `Database.Transaction()` method.

## High-Level Architecture

### Core Components
- **Entry Point**: `cmd/godin/main.go` - Application bootstrap, dependency injection
- **Server**: `internal/server/server.go` - HTTP server setup and lifecycle management
- **API Router**: `internal/api/router.go` - Route dispatch and handler registration
- **Database Layer**: `internal/database/postgres.go` - PostgreSQL abstraction with transaction support
- **Configuration**: `internal/config/config.go` - Viper-based config management with validation

### API Design Pattern
The API follows a service-oriented architecture with dynamic handler dispatch:

1. **Route Structure**: `/v1/{resourceType}/{serviceName}/...`
   - `resourceType`: service, session, file
   - `serviceName`: Currently supports "eris"

2. **Handler Interface**: All services implement the `cdn.Handler` interface:
   ```go
   type Handler interface {
       Service   // CreateService, GetService, DeleteService
       File      // CreateFile, GetFile, DeleteFile  
       Session   // CreateSession, GetSession, DeleteSession
   }
   ```

3. **Request Flow**:
   - Router parses URL path to extract resource type and service name
   - Dispatches to appropriate handler method based on HTTP method
   - Error handling via middleware wrapper

### Configuration Management
- **Primary Config**: `config/config.yml` (YAML format)
- **Environment Variables**: Prefixed with `GODIN_` (e.g., `GODIN_POSTGRES_CONNECTION_STRING`)
- **Dotenv Support**: Loads `.env` files automatically
- **Validation**: Built-in validation for required fields and formats

### Database Architecture
- **Connection**: Single PostgreSQL connection via `lib/pq` driver
- **Transaction Wrapper**: Custom `TransactionFn` type for safe database operations
- **Schema**: Uses PostgreSQL schemas (e.g., `eris.guilds`, `eris.services`)

### Error Handling
- **Custom Error Type**: `pkg/responder` provides structured error responses
- **Middleware**: `internal/api/middleware/error.go` wraps handlers for consistent error handling
- **Response Format**: JSON with status codes and error messages

### Service Integration
- **Discord Integration**: `internal/discord/discord.go` for Discord API operations
- **CDN Pattern**: Services implement standardized interfaces for file, session, and service management
- **Modular Services**: New services can be added by implementing the `cdn.Handler` interface

### Key Dependencies
- **HTTP Router**: Standard library `net/http` with `ServeMux`
- **Database**: PostgreSQL with `github.com/lib/pq` driver
- **Configuration**: `github.com/spf13/viper` for config management
- **Logging**: `go.uber.org/zap` for structured logging
- **Discord**: `github.com/bwmarrin/discordgo` for Discord API integration
- **UUID**: `github.com/google/uuid` for service ID generation

### Service Registration
Services are registered in `internal/api/router.go:NewRouter()`:
```go
services := map[string]cdn.Handler{
    "eris": eris.NewHandler(log, db),
    // Add new services here
}
```

### Development Notes
- No existing test files - tests should be added following Go conventions
- Configuration requires PostgreSQL connection string
- Server starts on port 60000 by default (configurable via `server_address`)
- Development mode can be enabled via `development: true` in config.yml