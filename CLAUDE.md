# Claude Interaction Guide

This document provides context for the Claude AI assistant to effectively interact with the `Godin-Next` codebase.

## Project Overview

-   **Language:** Go
-   **Primary Framework:** Standard library `net/http`.
-   **Database:** PostgreSQL
-   **Configuration:** `config/config.yml`

## Key Commands

-   **Run application:** `go run ./cmd/godin/main.go`
-   **Run tests:** `go test ./...` (Assumed)
-   **Tidy modules:** `go mod tidy`

## Development Notes

-   Follow existing code conventions.
-   Handlers are located in `internal/api/handler/`.
-   Services contain the core business logic.
-   Use the `pkg/responder` for creating HTTP responses.
-   Ensure all new code is covered by tests.
