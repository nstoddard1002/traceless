# traceless.pw

**Share secrets. Leave nothing behind.**

traceless.pw is a minimalist, secure, and privacy-focused secret sharing service. It allows you to share sensitive information (passwords, API keys, notes) through self-destructing links that are encrypted end-to-end.

## Features

- **End-to-End Encryption:** Secrets are encrypted in your browser using AES-256-GCM. The encryption key is never sent to the server.
- **Self-Destruction:** Secrets are automatically deleted after a set number of views or a specific expiration time.
- **Passcode Protection:** Optionally protect secrets with a passcode derived using Argon2id.
- **Zero-Knowledge Architecture:** The service operator cannot read your secrets.
- **Minimalist UI:** A clean, light-mode interface focused on speed and simplicity.
- **Docker Ready:** Easily deployable anywhere with Docker and Docker Compose.

## Tech Stack

- **Backend:** Go (Standard Library + `pgx` for PostgreSQL)
- **Frontend:** Vanilla HTML/CSS/JS (WebCrypto API)
- **Database:** PostgreSQL
- **Security:** AES-256-GCM, Argon2id (WASM)

## Getting Started

### Prerequisites

- Docker and Docker Compose

### Running Locally

1. Clone the repository:
   ```bash
   git clone https://github.com/nstoddard1002/traceless.git
   cd traceless
   ```

2. Start the application:
   ```bash
   docker compose up -d
   ```

3. Open your browser and navigate to `http://localhost:8080`.

## Deployment

The application is designed to be deployed behind a reverse proxy (like Nginx) with SSL/TLS enabled.

1. Set your `DATABASE_URL` and other environment variables in `docker-compose.yml`.
2. Ensure you are using HTTPS, as the WebCrypto API requires a secure context.

## Security Model

For a detailed explanation of our security architecture, see [docs/SECURITY.md](docs/SECURITY.md).

## License

Open source under the [MIT License](LICENSE).
