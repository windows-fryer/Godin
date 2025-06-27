# 🚀 Godin

> A powerful Discord CDN API service built with Go, providing seamless file and session management through Discord's infrastructure.

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Discord](https://img.shields.io/badge/Discord-Integration-7289da.svg)](https://discord.com/)

## ✨ Features

- 🔗 **Discord CDN Integration** - Leverage Discord's robust CDN infrastructure for file storage
- 📁 **File Management** - Upload, download, list, and delete files with ease
- 🔐 **Session Management** - Secure session creation, retrieval, and lifecycle management
- ⚙️ **Service Management** - Dynamic service configuration and control
- 🛡️ **Graceful Shutdown** - Clean service termination with signal handling
- 📊 **Structured Logging** - Comprehensive logging with glog integration
- 🌐 **REST API** - Clean, RESTful endpoints for all operations

## 🏗️ Architecture

Godin is composed of loosely-coupled layers that can be scaled or replaced independently.  At a high level, control flows from the HTTP surface down to the infrastructure adapters:

```
┌────────────────────────────┐
│        cmd/godin           │
│  Application entry point   │
└─────────────┬──────────────┘
              │ starts
┌─────────────▼──────────────┐
│        internal/server     │
│  HTTP router & middleware  │
└─────────────┬──────────────┘
              │ delegates
┌─────────────▼──────────────┐    ┌──────────────────────────┐
│        api/daemon          │    │       api/discord        │
│  Background task manager   │────│   Discord CDN adapter    │
└─────────────┬──────────────┘    └──────────────────────────┘
              │ accesses                              │
┌─────────────▼──────────────┐               ┌────────▼────────┐
│     internal/database      │               │internal/resource│
│  Persistence & migrations  │               │ Name generation │
└────────────────────────────┘               └─────────────────┘
```

The ASCII diagram is intentionally opinionated: *vertical* flow shows synchronous request handling, while *horizontal* arrows represent asynchronous or background interactions.  Each rectangle maps directly to a top-level package in the repository, making it trivial to navigate from docs to code.

## 🚀 Quick Start

### Prerequisites

- Go 1.22 or newer
- Discord Bot Token
- PostgreSQL database (optional, for database features)
- Environment configuration

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd Godin-Next
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Configure your `.env` file**
   ```env
   DISCORD_TOKEN=your_discord_bot_token_here
   SERVER_ADDRESS=:8080
   ```

5. **Run the application**
   
   **Normal logging:**
   ```bash
   go run cmd/godin/main.go -alsologtostderr=true -stderrthreshold=INFO -log_dir=log
   ```
   
   **Verbose logging (for development/debugging):**
   ```bash
   go run cmd/godin/main.go -alsologtostderr=true -stderrthreshold=INFO -log_dir=log -v 2
   ```

### Building for Production

```bash
# Build the binary
go build -o godin cmd/godin/main.go

# Run the binary (normal logging)
./godin -alsologtostderr=true -stderrthreshold=INFO -log_dir=log

# Run with verbose logging
./godin -alsologtostderr=true -stderrthreshold=INFO -log_dir=log -v 2
```

## 📚 API Reference

### File Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/file/` | Upload a new file |
| `GET` | `/v1/file/{id}` | Download a file |
| `GET` | `/v1/file/` | List all files |
| `DELETE` | `/v1/file/{id}` | Delete a file |

### Session Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/session/` | Create a new session |
| `GET` | `/v1/session/{id}` | Get session details |
| `DELETE` | `/v1/session/{id}` | Delete a session |

### Service Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/service/` | Create a service |
| `GET` | `/v1/service/{id}` | Get service details |
| `DELETE` | `/v1/service/{id}` | Delete a service |

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DISCORD_TOKEN` | Discord bot token | Required |
| `SERVER_ADDRESS` | HTTP server bind address | `:8080` |

### Discord Bot Setup

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application
3. Navigate to the "Bot" section
4. Create a bot and copy the token
5. Enable the required intents (Message Content Intent)

## 🛠️ Development

### Project Structure

```
Godin-Next/
├── api/
│   ├── daemon/          # Service registry and management
│   ├── discord/         # Discord CDN integration
│   │   ├── file/        # File operation handlers
│   │   ├── session/     # Session management handlers
│   │   └── service/     # Service operation handlers
│   └── service/         # CDN service interfaces
├── cmd/
│   └── godin/           # Main application entry point
├── internal/
│   ├── database/        # Database layer (PostgreSQL)
│   ├── resource/        # Resource name generation
│   └── server/          # HTTP server and routes
└── log/                 # Application logs
```

### Key Dependencies

- **Discord API**: `github.com/bwmarrin/discordgo` - Discord bot integration
- **Logging**: `github.com/golang/glog` - Structured logging
- **Environment**: `github.com/joho/godotenv` - Environment variable management
- **PostgreSQL**: `github.com/lib/pq` - Database driver
- **UUID Generation**: `github.com/google/uuid` - Unique identifier generation
- **Word Generation**: `github.com/wordgen/wordgen` - Random resource naming
- **Text Processing**: `golang.org/x/text` - Text manipulation utilities

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Code Style

This project follows standard Go conventions:
- Use `gofmt` for formatting
- Follow effective Go guidelines
- Use structured logging with glog
- Implement graceful error handling

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📖 Documentation – Consult the inline Go doc comments and the "Development" section above
- 🐛 Issue Tracker – Create issues for bugs and feature requests
- 💬 Discussions – Join the GitHub Discussions board for questions and ideas

## 🙏 Acknowledgments

- Discord for providing an excellent API and CDN infrastructure
- The Go community for amazing tools and libraries
- All contributors who help make this project better

---

<div align="center">
  <strong>Built with ❤️ and Go</strong>
</div>