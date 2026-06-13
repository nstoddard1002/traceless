# GEMINI.md

## Project Overview
**traceless.pw** is a secure, privacy-focused secret sharing service. It enables users to share sensitive information—such as passwords, API keys, and recovery codes—through self-destructing, encrypted links. The project prioritizes end-to-end security by encrypting all content in the browser before it reaches the server.

### Core Philosophy
- **Privacy First:** The service operator cannot access shared content.
- **Simplicity:** Secure sharing should be easy for non-technical users.
- **Ephemerality:** Secrets are deleted immediately after their purpose is served (view limit or expiration).

## Status: MVP Completed
The project has reached its MVP milestone. All core features—including client-side encryption, self-destruction, and passcode protection—are implemented and verified.

## Technical Stack
- **Frontend:** Vanilla JS, WebCrypto (AES-256-GCM), Argon2id (WASM).
- **Backend:** Go (Standard Library), `pgx` (PostgreSQL).
- **Database:** PostgreSQL.
- **Orchestration:** Docker & Docker Compose.

## Key Files
- `cmd/server/main.go`: Application entry point.
- `internal/api/`: REST API handlers.
- `internal/db/`: Database persistence layer.
- `internal/worker/`: Background cleanup worker.
- `web/`: Frontend assets (HTML, CSS, JS).
- `docker-compose.yml`: Production and local development setup.
- `docs/SECURITY.md`: Detailed security architecture documentation.

## Development Workflows
### Building and Running
1. Start the stack: `docker compose up -d`
2. Access the UI: `http://localhost:8080`

### Security Notes
- The encryption key is stored in the URL fragment (`#KEY`) and never sent to the server.
- Passcodes enable double-encryption using Argon2id.

