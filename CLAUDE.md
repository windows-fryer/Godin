# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Godin is a Go-based service that provides a Discord CDN API with file and session management capabilities. The application follows a multi-service architecture with daemon management.

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
- Discord API: `github.com/bwmarrin/discordgo`
- AWS SDK v2: DynamoDB integration ready
- Logging: `github.com/golang/glog`
- Environment: `github.com/joho/godotenv`

## Code Structure

### Route Handlers
Route implementations are in `internal/server/routes/` with handlers for:
- Error responses (`error.go`)
- File operations (`file.go`) 
- Session management (`session.go`)
- Service operations (`service.go`)

### Discord Integration
The Discord service (`api/discord/`) includes:
- Bot session management with configurable intents
- Random resource name generation for Discord resources
- File handling capabilities through the CDN interface

## Important Notes

- The application expects log files to be written to the `log/` directory
- Database service is currently a stub and needs implementation
- Discord service implements the CDNService interface for file and session management
- All services use structured logging via glog
- The server uses standard Go HTTP server without additional frameworks