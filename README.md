# Personal Digital Inventory

Desktop application that inventories files across local storage and Google Drive, helping you understand where each file lives. It scans, compares, and reports — it never deletes, uploads, or syncs files automatically.

## Features

- **File comparison** — shows files that are local-only, drive-only, or in both locations
- **Duplicate detection** — pattern-based detection (Windows copy, macOS copy, underscore, dash) with same-size validation
- **Export results** — export selected items to clipboard or .txt file (headers or individual files)
- **Local directory scanning** — recursive walk with configurable ignore patterns and .gitignore support
- **Google Drive scanning** — via Google Drive API with OAuth2 and folder picker
- **Selection** — click, Ctrl+Click (toggle), Shift+Click (range) across both tables
- **Compare view** — tabbed results table with sorting, filtering, search, and infinite scroll
- **Duplicates view** — grouped collapsible table with group-based selection, search, filters, sorting, expand/collapse all
- **Settings** — Google Drive connection, ignore patterns, .gitignore toggle

## Tech Stack

| Layer    | Technology                                      |
| -------- | ----------------------------------------------- |
| Desktop  | [Wails v2](https://wails.io/) (Go + webview)    |
| Backend  | Go 1.25, Clean Architecture                     |
| Frontend | React 19, TypeScript 5, Vite 6, Tailwind CSS 4  |
| Auth     | Google OAuth2 (callback server on :8085)         |

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) — `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Setup

### 1. Clone and install dependencies

```bash
git clone <repo-url>
cd personal-digital-inventory
cd frontend && npm install && cd ..
```

### 2. Configure

Edit `config.yaml` or use environment variables:

```yaml
google_drive:
  client_id: ""      # or set PDI_GOOGLE_CLIENT_ID env var
  client_secret: ""  # or set PDI_GOOGLE_CLIENT_SECRET env var
  token_file: token.json

ignore:
  patterns:
    - node_modules
    - .git
    - Thumbs.db
    - desktop.ini
    - "*.tmp"
    - "*.cache"
  respect_gitignore: true

logging:
  level: info  # debug | info | warn | error
```

### 3. Google Drive API credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project and enable the **Google Drive API**
3. Create OAuth 2.0 credentials (Desktop app type)
4. Set `PDI_GOOGLE_CLIENT_ID` and `PDI_GOOGLE_CLIENT_SECRET` environment variables

## Development

```bash
# Run in dev mode (hot reload)
make dev

# Build production binary
make wails-build

# Run built binary
make run
```

### Makefile commands

| Command          | Description                    |
| ---------------- | ------------------------------ |
| `make dev`       | Dev mode with hot reload       |
| `make wails-build` | Production build            |
| `make run`       | Run built binary               |
| `make be-test`   | Run all Go tests               |
| `make be-fmt`    | Format Go code                 |
| `make fe-build`  | Build frontend                 |
| `make fe-lint`   | Lint frontend                  |

## Testing

```bash
# Run all tests
make be-test

# Run with coverage
make be-test-cover
```

## Project Structure

```
├── app.go                          # Wails app + all bindings
├── main.go                         # Entry point
├── config.yaml                     # App configuration
├── Makefile                        # Build/dev/test commands
├── internal/
│   ├── domain/
│   │   ├── entity.go               # FileItem, IgnorePattern
│   │   ├── compare.go              # CompareFiles engine
│   │   ├── duplicate.go            # DuplicatePattern, DetectDuplicates engine
│   │   └── repository.go           # Repository interfaces
│   ├── application/                # Use cases (CompareUseCase)
│   └── infrastructure/
│       ├── config/                 # YAML config loader (IgnoreConfig struct)
│       ├── gdrive/                 # Google OAuth2 with callback server
│       ├── logging/                # Structured logging (slog)
│       └── scanner/                # Local scanner (.gitignore support) and Drive scanner
├── frontend/src/
│   ├── components/
│   │   ├── ResultsTable.tsx        # Compare results with selection + export
│   │   ├── DuplicateResultsTable.tsx # Grouped collapsible table with group selection
│   │   ├── ExportModal.tsx         # Export options (content type, path source, target)
│   │   ├── DriveFolderPicker.tsx   # Google Drive folder browser modal
│   │   ├── ErrorBanner.tsx         # Dismissible error display
│   │   ├── Sidebar.tsx             # Navigation (Compare, Duplicates, Settings)
│   │   └── Layout.tsx              # App shell
│   ├── pages/
│   │   ├── ComparePage.tsx         # Compare local vs Drive
│   │   ├── DuplicatesPage.tsx      # Duplicate detection scan
│   │   └── SettingsPage.tsx        # Drive connection, ignore patterns
│   └── lib/types.ts                # Shared types and utilities
└── docs/
    ├── guides/                     # User guides (e.g. Google Photos to Drive)
    ├── DONE/                       # Completed PRDs and task breakdowns
    └── TODO/                       # Planned features
```

## File Identity

Files are matched by combination of **relative path + filename + file size**. For duplicate detection, files must also match a naming pattern (e.g. `photo.jpg` and `photo (1).jpg`) and have the **same file size** to be grouped.

## Google Photos

Google Photos cannot be scanned directly due to API restrictions (since April 2025). See [docs/guides/google-photos-to-drive.md](docs/guides/google-photos-to-drive.md) for instructions on transferring photos to Google Drive first.

## License

Private — personal use only.
