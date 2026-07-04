# Personal Digital Inventory

Desktop application that inventories files across local storage and Google Drive, helping you understand where each file lives. It scans, compares, and reports — it never deletes, uploads, or syncs files automatically.

## Features

- **Local directory scanning** — recursive walk with ignore patterns
- **Google Drive scanning** — via Google Drive API with OAuth2
- **File comparison** — shows files that are local-only, drive-only, or in both locations
- **Compare view** — tabbed results table with sorting, filtering, search, and infinite scroll
- **Settings** — Google Drive connection management and ignore pattern configuration
- **Cancellation** — scans can be cancelled mid-operation

## Tech Stack

| Layer    | Technology                              |
| -------- | --------------------------------------- |
| Desktop  | [Wails v2](https://wails.io/) (Go + webview) |
| Backend  | Go 1.25, Clean Architecture             |
| Frontend | React 19, TypeScript 5, Vite 6, Tailwind CSS 4 |
| Database | PostgreSQL + GORM + Goose migrations    |
| Auth     | Google OAuth2                           |

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) — `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- [PostgreSQL 15+](https://www.postgresql.org/download/)

## Setup

### 1. Clone and install dependencies

```bash
git clone <repo-url>
cd personal-digital-inventory
cd frontend && npm install && cd ..
```

### 2. Create PostgreSQL database

```sql
CREATE USER inventory WITH PASSWORD 'your_password';
CREATE DATABASE personal_digital_inventory OWNER inventory;
```

### 3. Configure

Edit `config.yaml` or use environment variables:

```yaml
database:
  host: localhost
  port: 5432
  user: inventory
  password: ""       # or set PDI_DB_PASSWORD env var
  name: personal_digital_inventory
  sslmode: disable

google_drive:
  client_id: ""      # or set PDI_GOOGLE_CLIENT_ID env var
  client_secret: ""  # or set PDI_GOOGLE_CLIENT_SECRET env var
  token_file: token.json

ignore:
  - node_modules
  - .git
  - "*.tmp"
  - "*.cache"

logging:
  level: info        # debug | info | warn | error
```

### 4. Google Drive API credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project and enable the **Google Drive API**
3. Create OAuth 2.0 credentials (Desktop app type)
4. Set `PDI_GOOGLE_CLIENT_ID` and `PDI_GOOGLE_CLIENT_SECRET` environment variables

## Development

```bash
# Run in dev mode (hot reload)
wails dev

# Build production binary
wails build
```

### Makefile commands

```bash
make wails-dev        # Dev mode with hot reload
make wails-build      # Production build
make be-test          # Run all Go tests
make be-fmt           # Format Go code
make fe-build         # Build frontend
make fe-lint          # Lint frontend
```

## Testing

```bash
# Run all tests
go test ./... -count=1

# Run with coverage
go test ./internal/... -count=1 -cover
```

## Project Structure

```
├── app.go                          # Wails app + bindings
├── main.go                         # Entry point
├── config.yaml                     # App configuration
├── Makefile
├── internal/
│   ├── domain/                     # Entities, interfaces, comparison engine
│   ├── application/                # Use cases (CompareUseCase)
│   └── infrastructure/
│       ├── config/                 # YAML config loader
│       ├── database/               # PostgreSQL connection + Goose migrations
│       ├── gdrive/                 # Google OAuth2 auth
│       ├── logging/                # Structured logging (slog)
│       ├── repository/             # GORM repository implementations
│       └── scanner/                # Local and Drive file scanners
├── frontend/
│   └── src/
│       ├── components/             # Layout, Sidebar, ResultsTable, ErrorBanner
│       ├── pages/                  # ComparePage, SettingsPage
│       └── lib/                    # Types, utilities
└── docs/
    ├── PRD.md                      # Product Requirements Document
    └── TASKS.md                    # Task breakdown
```

## File Identity

Files are matched by combination of **relative path + filename + file size**. Hash-based identity (SHA256/BLAKE3) is planned for a future version.

## License

Private — personal use only.
