# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Godin is a Go-based service that provides a Discord CDN API with file and session management capabilities. The application follows a multi-service architecture with daemon management and implements a comprehensive Discord service with specialized handlers for files, sessions, and services.

## Architecture

### Core Components

- **Main Entry Point** (`cmd/godin/main.go`): Orchestrates daemon and CDN service startup/shutdown with graceful signal handling
- **Daemon Manager** (`api/daemon/daemon.go`): Registry for core system services (server, database)  
- **HTTP Server** (`internal/server/server.go`): REST API server with file, session, and service endpoints
- **Discord CDN Service** (`api/discord/discord.go`): Discord bot integration for CDN functionality
- **Database Layer** (`internal/database/database.go`): Database service interface (currently minimal)

### Service Architecture

The application uses a dual-service pattern:
1. **Daemons**: Core infrastructure services (HTTP server, database)
2. **CDN Services**: External API integrations implementing the `CDNService` interface

All services follow the Start/Stop lifecycle pattern for graceful management.

### API Endpoints

The REST API provides three main endpoint groups:
- `/v1/file/` - File upload, download, listing, and deletion
- `/v1/session/` - Session creation, retrieval, and management  
- `/v1/service/` - Service configuration and management

## Development Commands

### Running the Application
```bash
go run cmd/godin/main.go
```

### Building
```bash
go build -o main.exe cmd/godin/main.go
```

### Environment Setup
- Requires `.env` file with `DISCORD_TOKEN` and `SERVER_ADDRESS` environment variables
- Uses glog for structured logging with various log levels (INFO, ERROR, FATAL, WARNING)

### Dependencies
- Discord API: `github.com/bwmarrin/discordgo` - Discord bot integration
- Logging: `github.com/golang/glog` - Structured logging
- Environment: `github.com/joho/godotenv` - Environment variable management
- UUID Generation: `github.com/google/uuid` - Unique identifier generation
- PostgreSQL Driver: `github.com/lib/pq` - Database connectivity
- Word Generation: `github.com/wordgen/wordgen` - Random resource naming
- Text Processing: `golang.org/x/text` - Text manipulation utilities

## Code Structure

### Route Handlers
Route implementations are in `internal/server/routes/` with handlers for:
- Error responses (`error.go`)
- File operations (`file.go`) 
- Session management (`session.go`)
- Service operations (`service.go`)

### Discord Integration
The Discord service (`api/discord/`) includes:
- **Core Service** (`discord.go`): Main Discord bot integration with configurable intents
- **File Management** (`file/`): Specialized handlers for file upload, download, retrieval, and deletion
- **Session Management** (`session/`): Session creation, retrieval, and deletion handlers
- **Service Operations** (`service/`): Service creation, management, and deletion functionality
- **Resource Generation** (`internal/resource/`): Random resource name generation utilities

## Important Notes

- The application expects log files to be written to the `log/` directory
- Database service uses PostgreSQL with the `lib/pq` driver
- Discord service implements the CDNService interface for comprehensive file, session, and service management
- All services use structured logging via glog with multiple log levels
- The server uses standard Go HTTP server without additional frameworks
- Module path: `wednesday.wtf/godin`
- Go version: 1.24.3