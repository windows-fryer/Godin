# 🚀 Godin

> A powerful Go backend service with service-oriented architecture, featuring dynamic handler dispatch and PostgreSQL integration for scalable CDN-like resource management.

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ Features

- 🏗️ **Service-Oriented Architecture** - Dynamic handler dispatch with pluggable service modules
- 🔐 **Robust API Design** - RESTful endpoints with comprehensive error handling middleware
- 📁 **CDN Pattern** - Standardized interfaces for file, session, and service management
- 🗄️ **PostgreSQL Integration** - Custom database abstraction with transaction support
- ⚙️ **Configuration Management** - Viper-based config with environment variable support and validation
- 📊 **Structured Logging** - Comprehensive logging with `go.uber.org/zap`
- 🎮 **Discord Integration** - Built-in Discord API integration for service interactions
- 🔧 **Modular Services** - Easy service registration following the `cdn.Handler` interface

## 🏗️ Architecture

Godin follows a service-oriented architecture with dynamic handler dispatch. The system is designed around the CDN pattern where services implement standardized interfaces for file, session, and service management operations.

### Route Structure
```
/v1/{resourceType}/{serviceName}/...
```
- `resourceType`: service, session, file  
- `serviceName`: Currently supports "eris" (extensible for additional services)

### Handler Interface
All services implement the `cdn.Handler` interface combining three core interfaces:
```go
type Handler interface {
    Service   // CreateService, GetService, DeleteService
    File      // CreateFile, GetFile, DeleteFile  
    Session   // CreateSession, GetSession, DeleteSession
}
```

### Request Flow

1. Router parses URL path to extract resource type and service name
2. Dispatches to appropriate handler method based on HTTP method  
3. Handler executes business logic with database transactions
4. Error handling via middleware wrapper provides consistent responses

```mermaid
graph TD
    A[Client Request] --> B[main.go]
    B --> C[server.go]
    C --> D[router.go]
    D --> E[Error Middleware]
    E --> F[Dynamic Handler Dispatch]
    F --> G[Service Handler]
    G --> H[Database Transaction]
    H --> I[Response via Responder]
    I --> J[Client Response]

    subgraph "Configuration"
        K[config.yml]
        L[Environment Variables]
    end

    subgraph "Infrastructure"
        M[PostgreSQL Database]
        N[Structured Logger]
        O[Discord Integration]
    end

    B -.-> K
    B -.-> L
    G -.-> M
    G -.-> N
    G -.-> O
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24.3 or newer
- PostgreSQL database
- Optional: Discord bot token for Discord integration

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd Godin
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up configuration**
   Create a `config/config.yml` file with your PostgreSQL connection details and other settings.

4. **Environment variables (optional)**
   Set environment variables prefixed with `GODIN_` (e.g., `GODIN_POSTGRES_CONNECTION_STRING`)

5. **Run the application**
   ```bash
   go run cmd/godin/main.go
   # or build first
   go build -o bin/godin cmd/godin/main.go
   ./bin/godin
   ```

## 📚 API Reference

The API follows the pattern `/v1/{resourceType}/{serviceName}/...` where:
- `resourceType`: `service`, `session`, or `file`
- `serviceName`: Currently supports `eris`

### Service Operations
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/service/eris/` | Create a new service instance |
| `GET` | `/v1/service/eris/` | Retrieve service information |
| `DELETE` | `/v1/service/eris/` | Delete service instance |

### Session Operations  
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/session/eris/` | Create a new session |
| `GET` | `/v1/session/eris/` | Retrieve session information |
| `DELETE` | `/v1/session/eris/` | Delete session |

### File Operations
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/file/eris/` | Upload/create a file |
| `GET` | `/v1/file/eris/` | Retrieve file content |
| `DELETE` | `/v1/file/eris/` | Delete file |

## 🔧 Development

### Project Structure

```
Godin/
├── cmd/
│   └── godin/           # Main application entry point
├── config/              # Application configuration files
├── internal/
│   ├── api/             # API handlers, router, and middleware
│   │   ├── handler/
│   │   │   └── eris/    # Eris service handler implementation
│   │   ├── middleware/  # Auth and error handling middleware
│   │   └── router.go    # Route dispatch and handler registration
│   ├── cdn/             # CDN service interface definitions
│   ├── config/          # Configuration loading and validation
│   ├── database/        # PostgreSQL abstraction with transactions
│   ├── discord/         # Discord API integration
│   ├── logger/          # Structured logging setup
│   └── server/          # HTTP server lifecycle management
└── pkg/
    ├── jsonutil/        # JSON utility functions
    ├── responder/       # Standardized API responses
    └── splitutil/       # URL parsing utilities
```

### Key Dependencies

- **HTTP Router**: Standard library `net/http` with `ServeMux`
- **Database**: PostgreSQL with `github.com/lib/pq` driver  
- **Configuration**: `github.com/spf13/viper` for config management
- **Logging**: `go.uber.org/zap` for structured logging
- **Discord API**: `github.com/bwmarrin/discordgo` for Discord integration
- **UUID**: `github.com/google/uuid` for service ID generation
- **Environment**: `github.com/subosito/gotenv` for .env file support

### Development Commands

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

### Running Tests

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

### Configuration

- **Primary Config**: `config/config.yml` (YAML format)
- **Environment Variables**: Prefixed with `GODIN_` (e.g., `GODIN_POSTGRES_CONNECTION_STRING`)
- **Dotenv Support**: Loads `.env` files automatically
- **Validation**: Built-in validation for required fields and formats
- **Default Port**: Server starts on port 60000 (configurable via `server_address`)

### Adding New Services

To add a new service, implement the `cdn.Handler` interface and register it in `internal/api/router.go:NewRouter()`:

```go
services := map[string]cdn.Handler{
    "eris": eris.NewHandler(log, db),
    "newservice": newservice.NewHandler(log, db), // Add here
}
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🆘 Support

- 🐛 **Issue Tracker** – Create issues for bugs and feature requests.
- 💬 **Discussions** – Join the GitHub Discussions board for questions and ideas.

---

<div align="center">
  <strong>Built with ❤️ and Go</strong>
</div>