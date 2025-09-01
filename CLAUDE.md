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

### Database Schema
The application uses PostgreSQL with the following schema structure:

**Schema: `godin.eris`**
- `guilds`: Guild registration (guild_id)  
- `services`: Service instances (service_id, guild_id, bot_token)
- `channels`: Discord channels (guild_id, channel_id)
- `webhooks`: Discord webhooks (webhook_id, channel_id, webhook_token)
- `sessions`: Upload sessions (session_id, service_id, file_id, file_chunk_size)
- `files`: File metadata (file_id, file_name)

**Session Management**: Sessions expire after 5 minutes and are used for chunked file uploads

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
       File      // CreateFile, PutFile, GetFile, DeleteFile  
       Session   // CreateSession, GetSession, DeleteSession
   }
   ```

3. **Request Flow**:
   - Router parses URL path to extract resource type and service name
   - Dispatches to appropriate handler method based on HTTP method
   - Error handling via middleware wrapper

### API Endpoints

#### Service Operations (`/v1/service/eris/`)
- **POST**: Create service - Initializes Discord guild with channels and webhooks
  - Request: `{"bot_token": "string", "guild_id": int, "guild_channel_count": int, "guild_webhook_count": int}`
  - Response: `{"service_id": "uuid"}`
- **GET**: Get service information (placeholder)
- **DELETE**: Delete service (placeholder)

#### Session Operations (`/v1/session/eris/{service_id}/`)
- **POST**: Create upload session - Creates file metadata and session for chunked uploads
  - Request: `{"file_name": "string"}`
  - Response: `{"session_id": "uuid", "max_upload_size": int, "expires": int}`
- **GET**: Get session information (placeholder)
- **DELETE**: Delete session (placeholder)

#### File Operations (`/v1/file/eris/{session_id}/`)
- **POST**: Upload file chunks - Processes chunked file upload via Discord webhooks
  - Request: Binary file data in request body
  - Response: Success/error status
- **PUT**: Complete file upload (placeholder)
- **GET**: Download file (placeholder)
- **DELETE**: Delete file (placeholder)

### Configuration Management
- **Primary Config**: `config/config.yml` (YAML format)
- **Environment Variables**: Prefixed with `GODIN_` (e.g., `GODIN_POSTGRES_CONNECTION_STRING`)
- **Dotenv Support**: Loads `.env` files automatically via `github.com/subosito/gotenv`
- **Validation**: Built-in validation for required fields and formats
- **Module Path**: `wednesday.wtf/godin` (custom domain)
- **Required Configuration**: `postgres_connection_string` is mandatory
- **Default Server**: `127.0.0.1:60000` (configurable via `server_address`)

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
  - Guild initialization with channel and webhook creation
  - Upload size limits based on Discord guild premium tier (10MB/50MB/100MB)
  - Webhook management for file operations
  - Chunked file upload via Discord webhooks
- **CDN Pattern**: Services implement standardized interfaces for file, session, and service management
- **Modular Services**: New services can be added by implementing the `cdn.Handler` interface
- **URL Parsing**: Uses `pkg/splitutil.SplitURL()` to extract path parameters from versioned URLs

### Key Dependencies
- **HTTP Router**: Standard library `net/http` with `ServeMux`
- **Database**: PostgreSQL with `github.com/lib/pq` driver
- **Configuration**: `github.com/spf13/viper` for config management
- **Logging**: `go.uber.org/zap` for structured logging
- **Discord**: `github.com/bwmarrin/discordgo` for Discord API integration
- **UUID**: `github.com/google/uuid` for service ID generation
- **Environment**: `github.com/subosito/gotenv` for .env file support

### Service Registration
Services are registered in `internal/api/router.go:NewRouter()`:
```go
services := map[string]cdn.Handler{
    "eris": eris.NewHandler(log, db),
    // Add new services here
}
```

### Utility Packages
- **`pkg/responder`**: Standardized JSON API responses with error handling
- **`pkg/splitutil`**: URL parsing for extracting path parameters from versioned routes
- **`pkg/jsonutil`**: JSON utility functions (currently empty, planned for future use)

### Implementation Status
**Completed Features:**
- ✅ Service creation with Discord guild initialization
- ✅ Session management with upload size detection and 5-minute expiration
- ✅ File upload with chunked processing via Discord webhooks
- ✅ Database transaction handling
- ✅ Error middleware and response handling

**Placeholder Implementations:**
- 🚧 File download operations (`eris/file.go:GetFile()`)
- 🚧 File completion via PUT (`eris/file.go:PutFile()`)
- 🚧 Service/session GET and DELETE operations
- 🚧 Authentication middleware (`middleware/auth.go`)

### Development Notes
- No existing test files - tests should be added following Go conventions
- Configuration requires PostgreSQL connection string
- Server starts on `127.0.0.1:60000` by default (configurable via `server_address`)
- Development mode enabled via `development: true` in config.yml
- File uploads are processed in chunks with size limits based on Discord guild tier